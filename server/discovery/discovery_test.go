package discovery

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	dm "github.com/huajiao-tv/gokeeper/model/discovery"
	pb "github.com/huajiao-tv/gokeeper/pb/go"
)

type TestSessionBook2 struct {
	AddFunc    func(session *Session)
	DeleteFunc func(session *Session)
	sync.RWMutex
}

func (s TestSessionBook2) Add(session *Session) {
	s.AddFunc(session)
}
func (s TestSessionBook2) Delete(session *Session) {
	s.DeleteFunc(session)
}
func (s TestSessionBook2) GetSubscribers(serviceName string) ([]string, error) {
	return nil, nil
}
func (s TestSessionBook2) Push(upgradeServices map[string]*dm.Service) error {
	return nil
}

var testPbInstance = &pb.Instance{
	Id:          testInstanceId,
	ServiceName: testServiceName,
	Zone:        testZone,
	Env:         "test",
	Hostname:    "test.com",
	Addrs:       map[string]string{"rpc": rpcAddr, "http": httpAddr},
	Metadata:    map[string]string{testProperty: "101"},
	RegTime:     0,
	UpdateTime:  time.Now().Unix(),
}
var s = Server{}

func TestServer_Register(t *testing.T) {
	initTestServiceBook(t, false)
	testReq := &pb.RegisterReq{
		Instance:    testPbInstance,
		LeaseSecond: 3,
	}
	_, err := s.Register(context.Background(), testReq)
	if err != nil {
		t.Fatal("register error:", err)
	}
	services, err := reg.ListServices()
	if err != nil {
		t.Fatal("list service(storage) error:", err)
	}
	for _, s := range services {
		if _, ok := s.Instances[testZone][testInstanceId]; ok {
			return
		}
	}
	t.Fatal("registry failed!")
}

func TestServer_Deregister(t *testing.T) {
	initTestServiceBook(t, true)
	testReq := &pb.DeregisterReq{
		Instance: testPbInstance,
	}
	_, err := s.Deregister(context.Background(), testReq)
	if err != nil {
		t.Fatal("deregister failed!")
	}
	services, err := reg.ListServices()
	if err != nil {
		t.Fatal("list service(storage) error:", err)
	}
	for _, s := range services {
		if _, ok := s.Instances[testZone][testInstanceId]; ok {
			t.Fatal("registry failed!")
		}
	}
}

func TestServer_Polls(t *testing.T) {
	var session *Session
	discoverySessionBook = &TestSessionBook2{
		AddFunc: func(s *Session) {
			session = s
		},
		DeleteFunc: func(s *Session) {
			session = nil
		},
	}
	testStream := testPollsServer
	testStream.RecvFunc = func() (req *pb.PollsReq, e error) {
		req = nil
		e = errors.New("test error")
		return
	}
	err := s.Polls(testStream)
	if err == nil {
		t.Fatal("polls wrong stream")
	}

	testStream.RecvFunc = func() (req *pb.PollsReq, e error) {
		req = &pb.PollsReq{
			PollServices: map[string]int64{testServiceName: 0},
			Env:          "test",
			Subscriber:   testSubscriber,
		}
		e = nil
		return
	}

	go func() {
		err = s.Polls(testStream)
		if err != nil {
			if pollsWorking {
				t.Fatal("polls error:", err)
			} else {
				pollsWorking = true
				err = s.Polls(testStream)
			}
		}
	}()
	time.Sleep(testChanMonitorTime * time.Second)
	if session == nil {
		t.Fatal("polls does not add session!")
	}
	session.ErrCh <- errors.New("test error")
	time.Sleep(testChanMonitorTime * time.Second)
	if session != nil || err == nil {
		t.Fatal("polls does not delete session!")
	}
}

func TestServer_KeepAlive(t *testing.T) {
	initTestServiceBook(t, false)
	testReq := &pb.KeepAliveReq{
		Instance:    testPbInstance,
		LeaseSecond: 3,
	}
	_, err := s.KeepAlive(context.Background(), testReq)
	if err != nil {
		t.Fatal("keep alive error:", err)
	}
	services, err := reg.ListServices()
	if err != nil {
		t.Fatal("list service(storage) error:", err)
	}
	for _, s := range services {
		if _, ok := s.Instances[testZone][testInstanceId]; ok {
			return
		}
	}
	t.Fatal("registry failed!")
}
