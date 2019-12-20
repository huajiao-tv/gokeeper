package registry

import (
	dm "github.com/huajiao-tv/gokeeper/model/discovery"
)

//watch 事件类型
type WatchEventType string

const (
	WatchEventTypeCreate WatchEventType = "create"
	WatchEventTypeModify WatchEventType = "modify"
	WatchEventTypeDelete WatchEventType = "delete"
)

//watch 信息类型
type WatchInfoType string

const (
	//节点实例信息类型，主要有service provider注册
	WatchInfoTypeInstance = "instance"
	//节点属性信息类型，主要有后台干预设置
	WatchInfoTypeProperty = "property"
)

//watch返回的数据结构
type WatchEvent struct {
	//事件类型,主要用于instance create/modify/delete
	EventType WatchEventType
	//信息类型，主要分为instance和property
	InfoType WatchEventType
	//数据信息
	Data interface{}
	//版本号，用于更新service的版本
	Version int64
}

// 服务注册interface,只在keeper中调用，不暴露给client。
// （client通过grpc接口更keeper通信，然后keeper通过该接口进行注册)
type Registry interface {
	//Registry 初始化操作
	Init(opts ...OpOption) error
	//服务注册,返回元组信息，discovery需要将返回的MD信息添加到Instance的Metadata中 @todo keepalive是否共用此接口
	Register(instance *dm.Instance, opts ...OpRegisterOption) error
	//服务解除注册,client可以采用信号等机制退出时，调用解除注册的接口
	Deregister(instance *dm.Instance) error
	//根据ServiceName获取服务信息（Instance列表）@todo 如果keeper挂了，过一段时间重启后，可能数据已经变了，甚至部分服务为空了，此时应该如何处理?需要加策略，client端只有接受到当前服务个数的x%，才接受
	GetService(serviceName string) (*dm.Service, error)
	//列出已注册的所有service列表(该接口比较重，应该是keeper首次启动的时候调用一次,其余时间不会调用)
	ListServices() ([]*dm.Service, error)
	//订阅相关的service变化，支持监测所有service,这里的watcher需要抽象一下，不仅仅适用于etcd
	Watch() (<-chan *WatchEvent, error)
	//registry类型，例如etcd、consul等
	Type() string
	//属性设置
	SetProperty(property *dm.Property) error
}
