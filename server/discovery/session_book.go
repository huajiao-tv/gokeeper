package discovery

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	dm "github.com/huajiao-tv/gokeeper/model/discovery"
	pb "github.com/huajiao-tv/gokeeper/pb/go"
	"github.com/huajiao-tv/gokeeper/server/logger"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

var (
	ErrEmptyService = errors.New("service is empty")
)

//会话信息
type Session struct {
	//唯一id，server端接收到请求时生成
	Id string
	//stream流信息
	Stream pb.Discovery_PollsServer
	//订阅者
	Subscriber string
	//订阅的服务列表
	ServiceNames []string
	//变动的服务列表
	Services chan map[string]*dm.Service
	//session.Loop消费
	CloseCh chan struct{}
	//error，主要由stream Recv、Send产生,在rpc接口中消费,一旦出错，则关闭当前session
	ErrCh chan error
}

//创建session，生成唯一的id
func NewSession(subscriber string, stream pb.Discovery_PollsServer, serviceNames []string, serviceBook ServiceBooker) *Session {
	rawId := uuid.NewV4()
	id := fmt.Sprintf("%s:%x", getStreamClientAddr(stream), rawId)
	session := &Session{
		Id:           id,
		Stream:       stream,
		Subscriber:   subscriber,
		ServiceNames: serviceNames,
		Services:     make(chan map[string]*dm.Service, 10),
		CloseCh:      make(chan struct{}),
		ErrCh:        make(chan error, 2),
	}
	go session.readLoop(serviceBook)
	go session.writeLoop()

	return session
}

//获取stream client ip
func getStreamClientAddr(stream grpc.ServerStream) string {
	pr, ok := peer.FromContext(stream.Context())
	if !ok {
		return ""
	}
	if pr.Addr == net.Addr(nil) {
		return ""
	}
	return pr.Addr.String()
}

//根据client的请求参数，检查需要更新的service,如果reconnect时，推全量数据
func (session *Session) CheckUpgradedService(serviceBook ServiceBooker, req *pb.PollsReq, reconnect bool) {
	var (
		serviceVersions = map[string]int64{}
	)
	for serviceName, version := range req.PollServices {
		serviceVersions[serviceName] = version
	}

	//获取已经更新的服务配置项
	upgradeServices := serviceBook.GetUpgradedServices(serviceVersions, reconnect)
	if len(upgradeServices) > 0 {
		//推送更新
		session.Services <- upgradeServices
	}
}

//阻塞读poll请求
func (session *Session) readLoop(serviceBook ServiceBooker) {
	for {
		req, err := session.Stream.Recv()
		if err != nil {
			logger.Logex.Error("session readLoop recv error:", err, session.Id, session.Services)
			session.ErrCh <- err
			return
		}
		session.CheckUpgradedService(serviceBook, req, false)
	}
}

//周期推送service数据
func (session *Session) writeLoop() {
	var (
		upgradedServices map[string]*dm.Service
		ticker           = time.NewTicker(dm.DefaultPollsInterval)
		eventType        pb.DiscoveryEventType
	)
	defer ticker.Stop()

	for {
		select {
		case upgradedServices = <-session.Services:
			eventType = pb.DiscoveryEventType_DISCOVERY_EVENT_UPDATE
		case <-ticker.C:
			upgradedServices = map[string]*dm.Service{}
			eventType = pb.DiscoveryEventType_DISCOVERY_EVENT_NONE
		case <-session.CloseCh:
			return
		}
		err := session.sendServices(upgradedServices, eventType)
		if err != nil {
			logger.Logex.Error("session writeLoop sendServices error:", err, session.Id, session.Services, upgradedServices)
			session.ErrCh <- err
			return
		}
	}
}

//send service配置
func (session *Session) sendServices(upgradedServices map[string]*dm.Service, eventType pb.DiscoveryEventType) error {
	resp := &pb.PollsResp{
		Services:  map[string]*pb.Service{},
		EventType: eventType,
	}
	for serviceName, service := range upgradedServices {
		resp.Services[serviceName] = dm.FormatService(service)
	}
	return session.Stream.Send(resp)
}

//关闭session
func (session *Session) Close() {
	close(session.CloseCh)
}

type SessionBooker interface {
	Add(session *Session)
	Delete(session *Session)
	GetSubscribers(serviceName string) ([]string, error)
	Push(upgradeServices map[string]*dm.Service) error
	sync.Locker
}

//会话列表信息
type SessionBook struct {
	//会话列表，ServiceName -> (session)
	sessions map[string]map[string]*Session

	sync.RWMutex
}

var (
	discoverySessionBook SessionBooker
)

func InitSessionBook() {
	discoverySessionBook = &SessionBook{
		sessions: map[string]map[string]*Session{},
	}
}

//添加session会话
func (book *SessionBook) Add(session *Session) {
	book.Lock()
	defer book.Unlock()

	for _, serviceName := range session.ServiceNames {
		serviceSession, ok := book.sessions[serviceName]
		if !ok {
			serviceSession = map[string]*Session{}
			book.sessions[serviceName] = serviceSession
		}
		serviceSession[session.Id] = session
	}
}

//删除session会话
func (book *SessionBook) Delete(session *Session) {
	book.Lock()
	defer book.Unlock()

	for _, serviceName := range session.ServiceNames {
		serviceSession, ok := book.sessions[serviceName]
		if !ok {
			continue
		}
		delete(serviceSession, session.Id)
	}
}

//推送service变更
/*func (book *SessionBook) Push(serviceName string) error {
	book.RLock()
	defer book.RUnlock()

	sessions, ok := book.sessions[serviceName]
	if !ok {
		return nil
	}

	//获取已经更新的服务配置项
	upgradeServices := registryServiceBook.GetUpgradedServices(map[string]int64{serviceName: 0}, false)
	if len(upgradeServices) == 0 {
		logger.Logex.Error("SessionBook Push GetUpgradedSe"+
			"rvices error:", serviceName)
		return ErrEmptyService
	}

	for _, session := range sessions {
		session.Services <- upgradeServices
	}
	return nil
}*/
//减少外部依赖，便于添加单元测试，同时支持批量push
func (book *SessionBook) Push(upgradeServices map[string]*dm.Service) error {
	book.RLock()
	defer book.RUnlock()
	for serviceName, service := range upgradeServices {

		sessions, ok := book.sessions[serviceName]
		if !ok {
			return nil
		}
		upgradeService := map[string]*dm.Service{serviceName: service}
		for _, session := range sessions {
			session.Services <- upgradeService
		}
	}
	return nil
}

//获取某个服务的订阅列表，如果后台需要改数据，需要考虑分布式多节点情况，最好周期存库，方便处理
func (book *SessionBook) GetSubscribers(serviceName string) ([]string, error) {
	book.RLock()
	defer book.RUnlock()

	sessions, ok := book.sessions[serviceName]
	if !ok {
		return nil, nil
	}

	var (
		exists      = make(map[string]struct{})
		subscribers []string
	)
	for _, session := range sessions {
		if _, ok := exists[session.Subscriber]; ok {
			continue
		}
		exists[session.Subscriber] = struct{}{}
		subscribers = append(subscribers, session.Subscriber)
	}

	return subscribers, nil
}
