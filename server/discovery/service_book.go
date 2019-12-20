package discovery

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	dm "github.com/huajiao-tv/gokeeper/model/discovery"
	dr "github.com/huajiao-tv/gokeeper/server/discovery/registry"
	de "github.com/huajiao-tv/gokeeper/server/discovery/registry/etcd"
	"github.com/huajiao-tv/gokeeper/server/logger"
	"github.com/huajiao-tv/gokeeper/server/setting"
)

type ServiceBooker interface {
	FindInstance(serviceName, zone, id string) (*dm.Instance, error)
	GetUpgradedServices(serviceVersions map[string]int64, reconnect bool) map[string]*dm.Service
}

//所有服务列表
type ServiceBook struct {
	registry  dr.Registry
	services  map[string]*dm.Service
	eventChan <-chan *dr.WatchEvent

	sync.RWMutex
}

var (
	ErrServiceNotFound  = errors.New("service not found")
	ErrZoneNotFound     = errors.New("zone not found")
	ErrInstanceNotFound = errors.New("instance not found")
	ErrInvalidEventData = errors.New("event data is invalid")

	registryServiceBook *ServiceBook
)

type option struct {
	registry dr.Registry
}

type OpOption func(option *option)

func WithRegistry(registry dr.Registry) OpOption {
	return func(option *option) {
		option.registry = registry
	}
}

//初始化services
//如果discovery server重启后，service的version字段也需要初始化，目前etcd版本中，读取key的最大modVersion作为service version
func InitServiceBook(opts ...OpOption) error {
	option := &option{}
	for _, opt := range opts {
		opt(option)
	}

	registry := option.registry
	if registry == nil {
		//设置默认的registry
		registry = de.NewEtcdRegistry()
		err := registry.Init(
			dr.WithTimeout(5*time.Second),
			dr.WithAddrs(setting.RegistryUrl...),
			dr.WithAuth(setting.RegistryUsername, setting.RegistryPassword),
		)
		if err != nil {
			return err
		}
	}

	registryServiceBook = &ServiceBook{
		registry: registry,
		services: map[string]*dm.Service{},
	}

	//获取服务列表
	services, err := registry.ListServices()
	if err != nil {
		return err
	}

	eventChan, err := registry.Watch()
	if err != nil {
		return err
	}

	registryServiceBook.Lock()
	defer registryServiceBook.Unlock()

	//初始化服务列表
	for _, service := range services {
		registryServiceBook.services[service.ServiceName] = service
		fmt.Println("InitServiceBook service:", *service)
		for zone, zoneInstance := range service.Instances {
			fmt.Println("zone:", zone, "instance:", zoneInstance)
		}
	}
	registryServiceBook.eventChan = eventChan

	go registryServiceBook.Watch(discoverySessionBook)

	return nil
}

//处理instance更新事件
func (book *ServiceBook) instanceEventProxy(eventType dr.WatchEventType, instance *dm.Instance, version int64) error {
	var err error
	switch eventType {
	//创建节点
	case dr.WatchEventTypeCreate:
		err = book.createInstance(instance, version)
	//修改节点
	case dr.WatchEventTypeModify:
		err = book.modifyInstance(instance, version)
	//删除节点
	case dr.WatchEventTypeDelete:
		err = book.deleteInstance(instance, version)
	}
	return err
}

//???为什么要穿version，容易导致问题,创建是否严格保证不能创建原有instance

//创建instance
func (book *ServiceBook) createInstance(instance *dm.Instance, version int64) error {
	book.Lock()
	defer book.Unlock()

	service, ok := book.services[instance.ServiceName]
	if !ok {
		service = &dm.Service{
			ServiceName: instance.ServiceName,
			Instances:   map[string]map[string]*dm.Instance{},
		}
		book.services[service.ServiceName] = service
	}
	zone, ok := service.Instances[instance.Zone]
	if !ok {
		zone = map[string]*dm.Instance{}
		service.Instances[instance.Zone] = zone
	}

	zone[instance.Id] = instance
	service.UpdateTime = time.Now().Unix()
	service.Version = version
	return nil
}

//更新instance
func (book *ServiceBook) modifyInstance(instance *dm.Instance, version int64) error {
	book.Lock()
	defer book.Unlock()

	savedInstance, err := book.FindInstance(instance.ServiceName, instance.Zone, instance.Id)
	if err != nil {
		return err
	}

	service := book.services[instance.ServiceName]
	//替换instance中的metadata信息，仅替换非后台操作的metadata内容
	updateMetadata(savedInstance, instance.Metadata, false)
	//更新时间及版本号
	service.UpdateTime = time.Now().Unix()
	service.Version = version

	return nil
}

//更新instance中的metadata,分为后台和非后台数据，采用直接覆盖的方式(先删除后赋值)
func updateMetadata(instance *dm.Instance, freshMD dm.MD, backend bool) {
	if instance == nil || instance.Metadata == nil || freshMD == nil {
		return
	}

	//如果backend为true(updateProperty),以BackendMetadataPrefix为前缀的需要更新
	//如果backend为false(instanceEventProxy),不以BackendMetadataPrefix为前缀的需要更新
	needUpdate := func(k string) bool {
		b := strings.HasPrefix(k, dm.BackendMetadataPrefix)
		if backend {
			return b
		} else {
			return !b
		}
	}

	//避免删除需要修改的而未修改的属性
	/*for k := range instance.Metadata {
		if !needUpdate(k) {
			continue
		}
		delete(instance.Metadata, k)
	}*/

	for k, v := range freshMD {
		//如果是后台参数，不更新
		if !needUpdate(k) {
			continue
		}
		instance.Metadata[k] = v
	}
}

//从ServiceBook中查找对应的instance
func (book *ServiceBook) FindInstance(serviceName, zone, id string) (*dm.Instance, error) {
	service, ok := book.services[serviceName]
	if !ok {
		return nil, ErrServiceNotFound
	}
	if len(zone) == 0 {
		//如果zone为空，扫描所有zone
		for _, zoneInstance := range service.Instances {
			for savedId, instance := range zoneInstance {
				if savedId == id {
					return instance, nil
				}
			}
		}
		return nil, ErrInstanceNotFound
	} else {
		//zone非空，直接去对应zone的实例信息
		zoneInstances, ok := service.Instances[zone]
		if !ok {
			return nil, ErrZoneNotFound
		}
		savedInstance, ok := zoneInstances[id]
		if !ok {
			return nil, ErrInstanceNotFound
		}
		return savedInstance, nil
	}
}

//删除节点
func (book *ServiceBook) deleteInstance(instance *dm.Instance, version int64) error {
	book.Lock()
	defer book.Unlock()

	//校验对应的节点是否存在(zone可能为空)
	oldInstance, err := book.FindInstance(instance.ServiceName, instance.Zone, instance.Id)
	if err != nil {
		return err
	}

	delete(book.services[instance.ServiceName].Instances[oldInstance.Zone], instance.Id)
	book.services[instance.ServiceName].UpdateTime = time.Now().Unix()
	book.services[instance.ServiceName].Version = version
	return nil
}

//更新属性配置
func (book *ServiceBook) updateProperty(property *dm.Property, version int64) error {
	book.Lock()
	defer book.Unlock()

	service, ok := book.services[property.ServiceName]
	if !ok {
		return ErrServiceNotFound
	}

	zwStr, err := dm.EncodeZoneWeight(property.ZoneWeights)
	if err != nil {
		return err
	}

	//更新机房权重
	service.Metadata[dm.BackendMetadataZoneWeight] = zwStr
	service.Version = version

	//更新后台配置的属性信息
	for id, attr := range property.Attrs {
		instance, err := book.FindInstance(property.ServiceName, "", id)
		if err != nil {
			logger.Logex.Error("discovery updateProperty error:", err, id, property, version)
			continue
		}
		updateMetadata(instance, attr, true)
	}
	return nil
}

//根据版本号获取已更新的service，reconnect为true时，代表重新建立连接，此时推全量数据
//???
func (book *ServiceBook) GetUpgradedServices(serviceVersions map[string]int64, reconnect bool) map[string]*dm.Service {
	book.RLock()
	defer book.RUnlock()

	upgradedServices := map[string]*dm.Service{}
	for serviceName, version := range serviceVersions {
		service, ok := book.services[serviceName]
		if !ok {
			continue
		}
		if service.Version > version {
			upgradedServices[service.ServiceName] = service
		}
	}
	return upgradedServices
}

//监听服务配置
func (book *ServiceBook) Watch(discoverySessionBook SessionBooker) {
	var err error
	for event := range registryServiceBook.eventChan {
		fmt.Println("servicebook:", event)
		var serviceName string
		switch event.InfoType {
		case dr.WatchInfoTypeInstance:
			instance, ok := event.Data.(*dm.Instance)
			if !ok {
				err = ErrInvalidEventData
				logger.Logex.Error("discovery watch error:", err, event)
				continue
			}
			serviceName = instance.ServiceName
			err = book.instanceEventProxy(event.EventType, instance, event.Version)
		case dr.WatchInfoTypeProperty:
			property, ok := event.Data.(*dm.Property)
			if !ok {
				err = ErrInvalidEventData
				logger.Logex.Error("discovery watch error:", err, event)
				continue
			}
			serviceName = property.ServiceName
			err = book.updateProperty(property, event.Version)
		}
		if err != nil {
			logger.Logex.Error("discovery watch process error:", err, event)
			continue
		}
		dm.PrintService("Service Watch:", book.services[serviceName])

		//推送service更新
		book.Lock()
		service := book.services[serviceName]
		book.Unlock()
		upgradeService := map[string]*dm.Service{serviceName: service}
		err := discoverySessionBook.Push(upgradeService)
		if err != nil {
			logger.Logex.Error("discovery discoverySessionBook.Push:", err, serviceName, event)
		}
	}
}
