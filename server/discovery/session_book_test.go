package discovery

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"

	dm "github.com/huajiao-tv/gokeeper/model/discovery"
	"google.golang.org/grpc/peer"

	pb "github.com/huajiao-tv/gokeeper/pb/go"
	"google.golang.org/grpc"
)

//channel阻塞等待时长，降低可以在出错时加快执行时间（未出错时不会有channel的长时间阻塞）
const testChanMonitorTime = 1

var (
	testServiceName = "test_service"
	testSubscriber  = "test_client"
	rpcAddr         = "127.0.0.1:8000"
	httpAddr        = "127.0.0.1:8001"
	testZone        = "test_bjcc"
	testInstanceId  = "test_instance_id"
	testInstance    = &dm.Instance{
		Id:          testInstanceId,
		ServiceName: testServiceName,
		Zone:        testZone,
		Env:         "test",
		Hostname:    "test.com",
		Addrs:       map[string]string{"rpc": rpcAddr, "http": httpAddr},
		Metadata:    dm.MD{"propery1": "100"},
		RegTime:     0,
		UpdateTime:  time.Now().Unix(),
	}
	testService = &dm.Service{
		ServiceName: testServiceName,
		Instances:   map[string]map[string]*dm.Instance{testZone: {testInstanceId: testInstance}},
		Metadata:    nil,
		UpdateTime:  time.Now().Unix(),
		Version:     10,
	}
	testContextFunc = func() context.Context {
		ip, _ := net.ResolveIPAddr("", "127.0.0.1")
		return peer.NewContext(context.Background(), &peer.Peer{Addr: ip})
	}
	testSendFunc = func(r *pb.PollsResp) error {
		return nil
	}
	testRecvFunc = func() (*pb.PollsReq, error) {
		select {}
	}
	testPollsServer             = NewTestPollsServer(testSendFunc, testRecvFunc, testContextFunc)
	testGetUpgradedServicesFunc = func(serviceVersions map[string]int64, reconnect bool) map[string]*dm.Service {
		res := map[string]*dm.Service{}
		for service, version := range serviceVersions {
			if version < 10 {
				res[service] = testService
			}
		}
		return res
	}
	testServiceBook = &TestServiceBook{GetUpgradedServicesFunc: testGetUpgradedServicesFunc}
	testPollsReq    = &pb.PollsReq{
		PollServices: map[string]int64{testServiceName: 0},
		Subscriber:   testSubscriber,
	}
)

type TestPollsServer struct {
	SendFunc    func(r *pb.PollsResp) error
	RecvFunc    func() (*pb.PollsReq, error)
	ContextFunc func() context.Context
	grpc.ServerStream
}

func (p TestPollsServer) Context() context.Context {
	return p.ContextFunc()
}

func NewTestPollsServer(SendFunc func(r *pb.PollsResp) error, RecvFunc func() (*pb.PollsReq, error), ContextFunc func() context.Context) TestPollsServer {
	return TestPollsServer{
		SendFunc:    SendFunc,
		RecvFunc:    RecvFunc,
		ContextFunc: ContextFunc,
	}
}

func (p TestPollsServer) Send(r *pb.PollsResp) error {
	return p.SendFunc(r)
}
func (p TestPollsServer) Recv() (*pb.PollsReq, error) {
	return p.RecvFunc()
}

type TestServiceBook struct {
	GetUpgradedServicesFunc func(serviceVersions map[string]int64, reconnect bool) map[string]*dm.Service
}

func (t TestServiceBook) GetUpgradedServices(serviceVersions map[string]int64, reconnect bool) map[string]*dm.Service {
	return t.GetUpgradedServicesFunc(serviceVersions, reconnect)
}
func (t TestServiceBook) FindInstance(serviceName, zone, id string) (*dm.Instance, error) {
	return nil, nil
}

func (t TestServiceBook) Watch() {

}

func TestGetStreamClientAddr(t *testing.T) {
	s := TestPollsServer{ContextFunc: testContextFunc}
	if addr := getStreamClientAddr(s); addr != "127.0.0.1" {
		t.Fatal("get stream client addr error:", addr)
	}
	s = TestPollsServer{ContextFunc: func() context.Context {
		return peer.NewContext(context.Background(), &peer.Peer{Addr: nil})
	}}
	if addr := getStreamClientAddr(s); addr != "" {
		t.Fatal("get stream client addr error:", addr)
	}
}

func TestNewSession(t *testing.T) {
	session := NewSession(testSubscriber, testPollsServer, []string{testServiceName}, testServiceBook)
	if s := strings.Split(session.Id, ":"); len(s) != 2 || s[0] != "127.0.0.1" {
		t.Fatal("wrong id format:", session.Id)
	}
}

func TestReadLoop(t *testing.T) {
	t1 := testPollsServer
	t1.RecvFunc = func() (*pb.PollsReq, error) {
		e := errors.New("test error!")
		return nil, e
	}
	session := newTestSession(testSubscriber, t1, []string{testServiceName})
	go session.readLoop(testServiceBook)
	tt := time.NewTicker(time.Second * testChanMonitorTime)
	defer tt.Stop()
	select {
	case <-session.ErrCh:
		break
	case <-tt.C:
		t.Fatal("don't receive error")
	}
}

func TestCheckUpgradedService(t *testing.T) {
	session := newTestSession(testSubscriber, testPollsServer, []string{testServiceName})
	session.CheckUpgradedService(testServiceBook, testPollsReq, false)
	tt := time.NewTicker(time.Second * testChanMonitorTime)
	defer tt.Stop()
	select {
	case s := <-session.Services:
		if len(s) != 1 {
			t.Fatal("get wrong num services!")
		}
		if s[testServiceName] != testService {
			t.Fatal("get error service:", s[testServiceName])
		}
		break
	case <-tt.C:
		t.Fatal("don't receive error")
	}
	pollsReq := &pb.PollsReq{
		PollServices: map[string]int64{testServiceName: 100},
		Subscriber:   testSubscriber,
	}
	session.CheckUpgradedService(testServiceBook, pollsReq, false)
	tt = time.NewTicker(time.Second * testChanMonitorTime)
	defer tt.Stop()
	select {
	case s := <-session.Services:
		t.Fatal("get service when version is bigger", s[testServiceName])
	case <-tt.C:
		break
	}
}

func TestSendServices(t *testing.T) {
	res := false
	t1 := testPollsServer
	t1.SendFunc = func(r *pb.PollsResp) error {
		res = true
		return nil
	}
	session := newTestSession(testSubscriber, t1, []string{testServiceName})
	err := session.sendServices(map[string]*dm.Service{testServiceName: testService}, pb.DiscoveryEventType_DISCOVERY_EVENT_UPDATE)
	if err != nil {
		t.Fatal("send services error:", err)
	}
	if res == false {
		t.Fatal("send failed!")
	}
}

func newTestSession(subscriber string, stream pb.Discovery_PollsServer, serviceNames []string) *Session {
	rawId := uuid.NewV4()
	id := fmt.Sprintf("%s:%x", getStreamClientAddr(stream), rawId)
	return &Session{
		Id:           id,
		Stream:       stream,
		Subscriber:   subscriber,
		ServiceNames: serviceNames,
		Services:     make(chan map[string]*dm.Service, 10),
		CloseCh:      make(chan struct{}),
		ErrCh:        make(chan error, 2),
	}
}

func TestWriteLoop(t *testing.T) {
	t1 := testPollsServer
	trueChan := make(chan int, 1)
	t1.SendFunc = func(r *pb.PollsResp) error {
		if _, ok := r.Services["test_change"]; ok {
			trueChan <- 1
		}
		return nil
	}
	session := newTestSession(testSubscriber, t1, []string{testServiceName})
	go session.writeLoop()
	session.Services <- map[string]*dm.Service{"test_change": testService}
	tt := time.NewTicker(time.Second * testChanMonitorTime)
	defer tt.Stop()
	select {
	case <-trueChan:
		break
	case <-tt.C:
		t.Fatal("don't send service")
	}

	session.CloseCh <- struct{}{}
	session.Services <- map[string]*dm.Service{"test_change": testService}
	tt = time.NewTicker(time.Second * testChanMonitorTime)
	defer tt.Stop()
	select {
	case <-trueChan:
		t.Fatal("don't close")
	case <-tt.C:
		break
	}

	t2 := testPollsServer
	t2.SendFunc = func(r *pb.PollsResp) error {
		return errors.New("test error")
	}
	session2 := newTestSession(testSubscriber, t2, []string{testServiceName})
	go session2.writeLoop()
	session2.Services <- map[string]*dm.Service{"test_change": testService}
	tt = time.NewTicker(time.Second * testChanMonitorTime)
	defer tt.Stop()
	select {
	case <-session2.ErrCh:
		break
	case <-tt.C:
		t.Fatal("don't receive error")
	}

}

func TestInitSessionBook(t *testing.T) {
	InitSessionBook()
	if discoverySessionBook == nil {
		t.Fatal("discoverySessionBook does not init!")
	}
}
func DeleteSessionBook(t *testing.T) {
	discoverySessionBook = nil
}

func TestAddAndDelete(t *testing.T) {
	testSessionBook := &SessionBook{
		sessions: map[string]map[string]*Session{},
	}
	session := newTestSession(testSubscriber, testPollsServer, []string{testServiceName})
	testSessionBook.Add(session)
	if _, ok := testSessionBook.sessions[testServiceName][session.Id]; !ok {
		t.Fatal("add session failed!")
	}
	testSessionBook.Delete(session)
	if _, ok := testSessionBook.sessions[testServiceName][session.Id]; ok {
		t.Fatal("delete session failed!")
	}
}

func TestPush(t *testing.T) {
	session := newTestSession(testSubscriber, testPollsServer, []string{testServiceName})
	testSessionBook := &SessionBook{
		sessions: map[string]map[string]*Session{testServiceName: {session.Id: session}},
	}
	err := testSessionBook.Push(map[string]*dm.Service{testServiceName: testService})
	if err != nil {
		t.Fatal("push error:", err)
	}
	tt := time.NewTicker(time.Second * testChanMonitorTime)
	defer tt.Stop()
	select {
	case <-session.Services:
		break
	case <-tt.C:
		t.Fatal("don't receive services")
	}
}

func TestGetSubscribers(t *testing.T) {
	session := newTestSession(testSubscriber, testPollsServer, []string{testServiceName})
	testSessionBook := &SessionBook{
		sessions: map[string]map[string]*Session{testServiceName: {session.Id: session}},
	}
	subscribers, err := testSessionBook.GetSubscribers(testServiceName)
	if err != nil {
		t.Fatal("get subscribers error:", err)
	}
	if len(subscribers) != 1 {
		t.Fatal("get wrong num subscribes:", len(subscribers))
	}
	if subscribers[0] != testSubscriber {
		t.Fatal("get wrong subscriber:", subscribers[0])
	}
}
