package discovery

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"time"

	"github.com/huajiao-tv/gokeeper/client/schedule"
	dm "github.com/huajiao-tv/gokeeper/model/discovery"
	pb "github.com/huajiao-tv/gokeeper/pb/go"
	uuid "github.com/satori/go.uuid"
	"github.com/silenceper/pool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	ConnectTimeout     = time.Duration(5) * time.Second
	ReadTimeout        = time.Duration(10) * time.Second
	WriteTimeout       = time.Duration(10) * time.Second
	DefaultRegistryTTL = time.Duration(60) * time.Second

	poolInitCap     = 0 //如果在discovery service挂掉的情况下重启client节点，此时需要采用持久化中的数据，这是需要确保连接池中初始化数值为0，否则直接panic
	poolMaxCap      = 1 //注意，这里设置为1个连接，如果设置多个，server端可能保存多个session，会出现性能问题
	poolIdleTimeout = 2 * time.Hour

	Stdout io.Writer = os.Stdout
	Stderr io.Writer = os.Stderr

	SchemaHttp = "http"
	SchemaRpc  = "rpc"

	ErrNoService   = errors.New("has no service")
	ErrNoScheduler = errors.New("has no scheduler")
	ErrNoSchema    = errors.New("has no schema")
)

//服务注册及服务发现的相关参数设置
type option struct {
	//租约时间，keepalive每隔registryTTL/3续约一次
	registryTTL time.Duration
	//节点信息,注册时使用,如果不进行服务注册，可以为空
	instance *dm.Instance
	//服务的订阅者，用于标识订阅者
	subscriber string
	//订阅服务的列表
	serviceNames []string
	//scheduler配置, serviceName -> scheduler
	schedulers map[string]schedule.Scheduler
	//是否持久化service列表
	isPersist bool
	//持久化文件路径
	persistFile string
}

//服务发现的client结构体
type Client struct {
	//参数设置
	option *option

	//grpc
	discoveryClient pb.DiscoveryClient
	//discoveryStream pb.Sync_SyncClient
	pool pool.Pool

	//服务信息
	services atomic.Value
	//当前client保存的service版本
	versions map[string]int64
	//是否停止注册
	stopRegisterCh chan struct{}
}

//注意，id每个节点确保唯一，否则会相互覆盖，服务发现列表中会丢失节点！！！类似于gokeepr中的nodeId
func NewInstance(id, serviceName string, Addrs map[string]string) *dm.Instance {
	hostname, zone := getHostInfo()
	return &dm.Instance{
		Id:          id,
		ServiceName: serviceName,
		Zone:        zone,
		Hostname:    hostname,
		Addrs:       Addrs,
	}
}

func getHostInfo() (hostname string, zone string) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", "unknown"
	}

	zone = "unknown"
	hostArr := strings.Split(hostname, ".")
	if len(hostArr) >= 3 {
		zone = hostArr[len(hostArr)-3]
	}
	return hostname, zone
}

func GenRandomId(prefix ...string) string {
	rawId := uuid.NewV4()
	pre := strings.Join(prefix, ":")
	id := fmt.Sprintf("%s:%x", pre, rawId)
	return id
}

type OpOption func(option *option)

//设置续约时长，ttl为租约时长，每隔ttl/3续约一次
func WithRegistryTTL(ttl time.Duration) OpOption {
	//ttl最低3秒
	if ttl/time.Second < 3 {
		ttl = 3 * time.Second
	}
	return func(option *option) {
		option.registryTTL = ttl
	}
}

//设置要注册的服务，如果没有服务需要注册，不用调用该函数
func WithRegistry(instance *dm.Instance) OpOption {
	return func(option *option) {
		option.instance = instance
	}
}

//设置要发现的服务列表，如果不需要服务发现，不用调用该函数
//subscriber为服务的订阅者，理论上subscriber是唯一的，例如live订阅了counter的服务，此时subscriber为live，serviceNames为[]string{"counter"}
func WithDiscovery(subscriber string, serviceNames []string) OpOption {
	return func(option *option) {
		option.subscriber = subscriber
		option.serviceNames = append(option.serviceNames, serviceNames...)
	}
}

//设置负载均衡调度器，如果不需要服务发现，不用调用该函数
//如果需要服务发现而不调用该函数，默认选择随机负载均衡调度器
//@todo 在项目中动态设置负载均衡调度器
func WithScheduler(schedulers map[string]schedule.Scheduler) OpOption {
	return func(option *option) {
		for serviceName, scheduler := range schedulers {
			option.schedulers[serviceName] = scheduler
		}
	}
}

//持久化文件设置
func WithPersistence() OpOption {
	return func(option *option) {
		option.isPersist = true
	}
}

func New(discoveryAddr string, opts ...OpOption) *Client {
	option := &option{
		registryTTL: DefaultRegistryTTL,
		schedulers:  map[string]schedule.Scheduler{},
	}
	for _, opt := range opts {
		opt(option)
	}

	//如果没有提供subscriber，则不能采用持久化方式
	if len(option.subscriber) == 0 {
		option.isPersist = false
	}
	//初始化持久化文件路径
	if option.isPersist {
		option.persistFile = fmt.Sprintf("/tmp/%s.yaml", option.subscriber)
	}

	//服务发现列表中，如果没有指定负载均衡scheduler，则默认指定randomScheduler
	for _, serviceName := range option.serviceNames {
		if _, ok := option.schedulers[serviceName]; !ok {
			option.schedulers[serviceName] = schedule.NewRandomScheduler()
		}
	}

	dialOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		//注意，该值一定要设置，代表tcp conn断连重试的等待最大时长，否则在server挂掉很长时间的情况下，重连等待事件最长可能到120s，
		//等待时长按指数递增，算法可参考 https://github.com/grpc/grpc/blob/master/doc/connection-backoff.md
		grpc.WithBackoffMaxDelay(1 * time.Second),
	}

	dialCtx, dialCancel := context.WithTimeout(context.Background(), ConnectTimeout)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, discoveryAddr, dialOpts...)
	if err != nil {
		if option.isPersist {
			Stderr.Write([]byte(fmt.Sprintf("%s|discovery|client|New|New discovery client grpc.DialContext addr:%s error:%s\n", time.Now().String(), discoveryAddr, err.Error())))
		} else {
			//如果不能从持久化文件中加载，则直接panic
			panic(fmt.Sprintf("New discovery client grpc.DialContext addr:%s error:%s", discoveryAddr, err.Error()))
		}
	}

	discoveryClient := pb.NewDiscoveryClient(conn)
	p, err := NewPollsStreamPool(discoveryClient, poolInitCap, poolMaxCap, poolIdleTimeout)
	if err != nil {
		if option.isPersist {
			Stderr.Write([]byte(fmt.Sprintf("%s|discovery|client|New|NewPollsStreamPool error: error:%s\n", time.Now().String(), err.Error())))
		} else {
			//如果不能从持久化文件中加载，则直接panic
			panic("NewPollsStreamPool error:" + err.Error())
		}
	}

	versions := make(map[string]int64, len(option.serviceNames))
	for _, serviceName := range option.serviceNames {
		versions[serviceName] = 0
	}

	client := &Client{
		option:          option,
		discoveryClient: discoveryClient,
		pool:            p,
		versions:        versions,
		stopRegisterCh:  make(chan struct{}, 1),
	}
	client.services.Store(map[string]*dm.Service{})

	return client
}

func (client *Client) Work() {
	if client.option.instance != nil {
		err := client.register()
		if err != nil {
			panic("discovery register error:" + err.Error())
		}
		go client.loopKeepalive()
	}
	if len(client.option.serviceNames) > 0 {
		//采用持久化的服务列表需要满足如下条件:
		//1、首次加载
		//2、已设置持久化
		//3、poll请求discovery server时报错
		//该种策略主要解决discovery server挂掉的情况下client节点重启问题
		err := client.polls(client.option.isPersist)
		if err != nil {
			panic("discovery polls error:" + err.Error())
		}
		go client.loopPolls()
	}
}

func (client *Client) loopKeepalive() {
	ticker := time.NewTicker(client.option.registryTTL / 3)
KEEPALIVE:
	for {
		select {
		case <-ticker.C:
		RETRY:
			err := client.keepalive()
			if err != nil {
				Stderr.Write([]byte(fmt.Sprintf("%s|discovery|client|keepalive|code:%d msg:%s \n", time.Now().String(), status.Code(err), err.Error())))
				time.Sleep(1 * time.Second)
				goto RETRY
			}
		case <-client.stopRegisterCh:
			break KEEPALIVE
		}
	}
}

func (client *Client) loopPolls() {
	for {
		err := client.polls(false)
		if err != nil {
			Stderr.Write([]byte(fmt.Sprintf("%s|discovery|client|polls|code:%d msg:%s \n", time.Now().String(), status.Code(err), err.Error())))
			time.Sleep(1 * time.Second)
		}
	}
}

//注册服务 @todo 填充metadata信息，用于分布式下的会话机制
func (client *Client) register() error {
	//如果instance为nil，则不进行注册
	if client.option.instance == nil {
		return nil
	}

	instance := client.option.instance
	instance.RegTime = time.Now().Unix()
	instance.UpdateTime = time.Now().Unix()

	req := &pb.RegisterReq{
		Instance:    dm.FormatInstance(instance),
		LeaseSecond: int64(client.option.registryTTL / time.Second),
	}

	ctx, cancel := context.WithTimeout(context.Background(), WriteTimeout)
	defer cancel()
	_, err := client.discoveryClient.Register(ctx, req)
	if err != nil {
		Stderr.Write([]byte(fmt.Sprintf("%s|discovery|client|register|code:%d msg:%s \n", time.Now().String(), status.Code(err), err.Error())))
	}
	return nil
}

//解除服务注册
func (client *Client) deregister() error {
	//如果instance为nil，则不进行注册
	if client.option.instance == nil {
		return nil
	}

	instance := client.option.instance
	req := &pb.DeregisterReq{
		Instance: dm.FormatInstance(instance),
	}

	ctx, cancel := context.WithTimeout(context.Background(), WriteTimeout)
	defer cancel()
	_, err := client.discoveryClient.Deregister(ctx, req)
	if err != nil {
		Stderr.Write([]byte(fmt.Sprintf("%s|discovery|client|deregister|code:%d msg:%s \n", time.Now().String(), status.Code(err), err.Error())))
	}
	return nil
}

//服务保活
func (client *Client) keepalive() error {
	//如果instance为nil，则不进行保活
	if client.option.instance == nil {
		return nil
	}

	instance := client.option.instance
	instance.UpdateTime = time.Now().Unix()
	req := &pb.KeepAliveReq{
		Instance:    dm.FormatInstance(instance),
		LeaseSecond: int64(client.option.registryTTL / time.Second),
	}

	ctx, cancel := context.WithTimeout(context.Background(), WriteTimeout)
	defer cancel()
	_, err := client.discoveryClient.KeepAlive(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

//获取服务列表
func (client *Client) polls(usePersistenceWhenError bool) error {
	if len(client.option.serviceNames) == 0 {
		return nil
	}

	newServices := map[string]*dm.Service{}
	resp, err := Polls(client.pool, &pb.PollsReq{Subscriber: client.option.subscriber, PollServices: client.versions}, WithPollsCallTimeout(dm.DefaultPollsInterval+WriteTimeout))
	if err != nil {
		if !usePersistenceWhenError {
			return err
		}
		Stderr.Write([]byte(fmt.Sprintf("%s|discovery|client|polls|Polls error:%s, try to use persistence services\n", time.Now().String(), err.Error())))
		newServices, err = readServices(client.option.persistFile)
		if err != nil {
			Stderr.Write([]byte(fmt.Sprintf("%s|discovery|client|polls|use persistence services failed\n", time.Now().String())))
			return err
		}
		Stderr.Write([]byte(fmt.Sprintf("%s|discovery|client|polls|use persistence services successfully\n", time.Now().String())))
		//初始化client version
		for serviceName, service := range newServices {
			client.versions[serviceName] = service.Version
		}
	} else {
		//事件类型处理，目前只处理服务列表变更时的情况
		switch resp.EventType {
		case pb.DiscoveryEventType_DISCOVERY_EVENT_UPDATE:
		default:
			return nil
		}

		//@todo 如果resp.Services为空，是否更新服务列表
		//if len(resp.Services) == 0 {
		//	return nil
		//}

		//解析服务列表，只修改变动的服务
		oldServices := client.services.Load().(map[string]*dm.Service)
		for serviceName, service := range oldServices {
			newServices[serviceName] = service
		}
		for serviceName, service := range resp.Services {
			newServices[serviceName] = dm.FilterOfflineInstance(dm.ParseService(service))
			client.versions[serviceName] = service.Version

			dm.PrintService("polls:", newServices[serviceName])
		}
	}

	client.services.Store(newServices)

	//适配负载均衡scheduler的数据
	for serviceName, service := range newServices {
		scheduler, ok := client.option.schedulers[serviceName]
		if !ok {
			Stderr.Write([]byte(fmt.Sprintf("%s|discovery|client|polls|service %s has no scheduler\n", time.Now().String(), serviceName)))
			continue
		}
		err := scheduler.Build(service)
		if err != nil {
			Stderr.Write([]byte(fmt.Sprintf("%s|discovery|client|polls|service:%s %s scheduler build error:%s\n", time.Now().String(), serviceName, scheduler.Type(), err.Error())))
		}
	}

	//如果需要持久化services列表，需要同步写数据，不要开辟协程，防止多次更新时导致go协程执行乱序问题
	if client.option.isPersist {
		//@todo 从文件中加载service列表时无需重写
		if err := writeServices(client.option.persistFile, newServices); err != nil {
			Stderr.Write([]byte(fmt.Sprintf("%s|discovery|client|polls|persist services error:%s\n", time.Now().String(), err.Error())))
		}
	}

	return nil
}

//获取服务列表信息
func (client *Client) GetService(serviceName string) (*dm.Service, error) {
	services := client.services.Load().(map[string]*dm.Service)
	service, ok := services[serviceName]
	if !ok {
		return nil, ErrNoService
	}
	return service, nil
}

//根据负载均衡策略，随机选择服务中的某个instance，返回节点addr
func (client *Client) GetServiceAddr(serviceName, schema string) (string, error) {
	scheduler, ok := client.option.schedulers[serviceName]
	if !ok {
		return "", ErrNoScheduler
	}
	instance, err := scheduler.Select()
	if err != nil {
		return "", err
	}
	if instance.Addrs == nil {
		return "", ErrNoSchema
	}
	addr, ok := instance.Addrs[schema]
	if !ok {
		return "", ErrNoSchema
	}

	return addr, nil
}

//发送信号进行解注册
func (client *Client) SignalDeregister(exit bool, signals ...os.Signal) {
	if client.option.instance == nil {
		return
	}

	ch := make(chan os.Signal, 10)
	signal.Notify(ch, signals...)
	go func() {
		for {
			sig := <-ch
			Stdout.Write([]byte(fmt.Sprintf("%s|discovery|client|SignalDeregister|receive signal:%s,try to degrester\n", time.Now().String(), sig)))
			err := client.deregister()
			if err != nil {
				Stderr.Write([]byte(fmt.Sprintf("%s|discovery|client|SignalDeregister|degrester error:%s\n", time.Now().String(), err.Error())))
			} else {
				Stdout.Write([]byte(fmt.Sprintf("%s|discovery|client|SignalDeregister|receive signal:%s,degrester successfully\n", time.Now().String(), sig)))
				client.stopRegisterCh <- struct{}{}
			}
			if exit {
				os.Exit(0)
			}
		}
	}()
}
