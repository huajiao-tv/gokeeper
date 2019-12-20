package discovery

import (
	"time"

	dm "github.com/huajiao-tv/gokeeper/model/discovery"
	pb "github.com/huajiao-tv/gokeeper/pb/go"
	"github.com/huajiao-tv/gokeeper/server/discovery/registry"
	"github.com/huajiao-tv/gokeeper/server/logger"
)

var (
	defaultRegistryTTL = 30 * time.Second
)

//注册
func Register(instance *dm.Instance, registryTTL time.Duration) error {
	if registryTTL == 0 {
		registryTTL = defaultRegistryTTL
	}
	err := registryServiceBook.registry.Register(instance, registry.WithRegistryTTL(registryTTL), registry.WithRefresh())
	return err
}

//保活
func KeepAlive(instance *dm.Instance, registryTTL time.Duration) error {
	if registryTTL == 0 {
		registryTTL = defaultRegistryTTL
	}
	err := registryServiceBook.registry.Register(instance, registry.WithRegistryTTL(registryTTL))
	return err
}

//解注册
func Deregister(instance *dm.Instance) error {
	err := registryServiceBook.registry.Deregister(instance)
	return err
}

//轮询获取service配置
func Polls(stream pb.Discovery_PollsServer) error {
	//读取请求
	req, err := stream.Recv()
	if err != nil {
		return err
	}

	var serviceNames []string
	for serviceName := range req.PollServices {
		serviceNames = append(serviceNames, serviceName)
	}
	session := NewSession(req.Subscriber, stream, serviceNames, registryServiceBook)

	//首次接收请求时，需要检测service是否有更新
	session.CheckUpgradedService(registryServiceBook, req, true)

	//将session添加到SessionBook中
	discoverySessionBook.Add(session)

	//如果有错误，直接退出当前stream
	select {
	case err := <-session.ErrCh:
		logger.Logex.Error("Polls error:", session.Id, serviceNames)
		//从SessionBook中删除当前session
		discoverySessionBook.Delete(session)
		//关闭当前session
		session.Close()
		return err
	}
}

//获取某个服务
func GetService(serviceName string) (*dm.Service, error) {
	return registryServiceBook.registry.GetService(serviceName)
}

//获取服务列表
func ListServices() ([]*dm.Service, error) {
	return registryServiceBook.registry.ListServices()
}

//设置属性
func SetProperty(property *dm.Property) error {
	err := registryServiceBook.registry.SetProperty(property)
	return err
}
