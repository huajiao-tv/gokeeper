package etcd

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	dm "github.com/huajiao-tv/gokeeper/model/discovery"
	"github.com/huajiao-tv/gokeeper/server/discovery/registry"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/etcdserver/api/v3rpc/rpctypes"
	"go.etcd.io/etcd/mvcc/mvccpb"
)

var (
	ErrLeaseIdNotFound = errors.New("lease id not found")
	ErrInstanceIsNil   = errors.New("instance is nil")
)

type etcdRegistry struct {
	option *registry.Option
	client *clientv3.Client
	//后台属性值，只能通过后台设置 serviceName -> Property
	properties map[string]*dm.Property

	//保存leaseid，id -> leaseId，注意，leaseId为后端存储的概念，不要暴露给外侧
	leaseIds map[string]clientv3.LeaseID
	sync.RWMutex
}

func NewEtcdRegistry() *etcdRegistry {
	option := &registry.Option{
		Addrs:   []string{"127.0.0.1:2379"},
		Timeout: defaultTimeout,
		Logger:  registry.DefaultLogger,
	}
	r := &etcdRegistry{
		option:     option,
		properties: map[string]*dm.Property{},
		leaseIds:   map[string]clientv3.LeaseID{},
	}

	return r
}

//Registry 初始化操作
func (r *etcdRegistry) Init(opts ...registry.OpOption) error {
	for _, opt := range opts {
		opt(r.option)
	}

	cfg := clientv3.Config{
		Endpoints:   r.option.Addrs,
		Username:    r.option.Username,
		Password:    r.option.Password,
		DialTimeout: r.option.Timeout,
	}

	client, err := clientv3.New(cfg)
	if err != nil {
		return err
	}
	r.client = client

	return nil
}

//服务注册,返回元组信息，discovery需要将返回的MD信息添加到Instance的Metadata中
//leaseId从Instance中的metadata中获取
func (r *etcdRegistry) Register(instance *dm.Instance, opts ...registry.OpRegisterOption) error {
	if instance == nil {
		return ErrInstanceIsNil
	}

	var (
		op = &registry.RegisterOption{
			TTL: defaultLeaseTTL,
		}
		leaseId clientv3.LeaseID
	)

	for _, opt := range opts {
		opt(op)
	}

	for {
		//如果强制刷新，则进行Put操作
		if op.Refresh {
			break
		}
		leaseId, ok := r.leaseIds[instance.Id]
		if !ok {
			break
		}
		if _, err := r.client.KeepAliveOnce(context.TODO(), leaseId); err == rpctypes.ErrLeaseNotFound {
			//如果lease not found，需要重新创建lease
			r.option.Logger.Error("Etcd Register client.KeepAliveOnce lease not found:", err, leaseId, instance)
			break
		}
		return nil
	}

	gCtx, gCancel := context.WithTimeout(context.Background(), r.option.Timeout)
	defer gCancel()

	if grantResp, err := r.client.Grant(gCtx, int64(op.TTL/time.Second)); err != nil {
		return err
	} else {
		leaseId = grantResp.ID
	}

	//将leaseId填充到instance metadata中
	//分布式节点情况下，client节点第一次创建leaseId可能在server A，进行keepalive时可能在server B，
	//如果server B没有对应的leaseId，会重新创建。为避免该问题，将leaseId广播到每个节点
	leaseIdStr := strconv.FormatInt(int64(leaseId), 10)
	if instance.Metadata == nil {
		instance.Metadata = dm.MD{}
	}
	instance.Metadata[etcdMetadataLeaseId] = leaseIdStr

	info, err := encodeInstance(instance)
	if err != nil {
		return err
	}
	//删除etcdMetadataLeaseId，防止影响入参
	delete(instance.Metadata, etcdMetadataLeaseId)

	pCtx, pCancel := context.WithTimeout(context.Background(), r.option.Timeout)
	defer pCancel()

	if _, err = r.client.Put(pCtx, getInstancePath(instance.ServiceName, instance.Id), info, clientv3.WithLease(leaseId)); err != nil {
		return err
	}
	return nil
}

//服务解除注册,client可以采用信号等机制退出时，调用解除注册的接口
func (r *etcdRegistry) Deregister(instance *dm.Instance) error {
	if instance == nil {
		return ErrInstanceIsNil
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.option.Timeout)
	defer cancel()

	_, err := r.client.Delete(ctx, getInstancePath(instance.ServiceName, instance.Id))
	return err
}

func (r *etcdRegistry) getServiceAux(serviceName string, kvs []*mvccpb.KeyValue) *dm.Service {
	var (
		maxModRevision int64 = 0
		property       *dm.Property
		serviceMd      = dm.MD{}
		instances      = map[string]map[string]*dm.Instance{}
	)

	for _, ev := range kvs {
		_, infoType, _, err := parseRawPath(string(ev.Key))
		if err != nil {
			r.option.Logger.Error("Etcd GetService parseRawPath error:", err, string(ev.Key), string(ev.Value))
			continue
		}

		switch infoType {
		case infoTypeInstance:
			instance, err := decodeInstance(string(ev.Value))
			if err != nil {
				r.option.Logger.Error("Etcd GetService decodeInstance error:", err, string(ev.Key), string(ev.Value))
				continue
			}
			if instance.Metadata == nil {
				instance.Metadata = dm.MD{}
			}
			zoneInstances, ok := instances[instance.Zone]
			if !ok {
				zoneInstances = map[string]*dm.Instance{}
				instances[instance.Zone] = zoneInstances
			}
			zoneInstances[instance.Id] = instance
			//更新leaseId,主要解决discovery server重启后leaseId为空问题
			_ = r.updateLeaseIds(instance, mvccpb.PUT)
		case infoTypeProperty:
			property, err = decodeProperty(string(ev.Value))
			if err != nil {
				r.option.Logger.Error("Etcd GetService decodeProperty error:", err, string(ev.Key), string(ev.Value))
				continue
			}
			r.properties[property.ServiceName] = property
		}

		if maxModRevision < ev.ModRevision {
			maxModRevision = ev.ModRevision
		}
	}

	//将后台设置的属性添加到instance metadata节点中
	if property != nil {
		//设置节点属性
		for _, zi := range instances {
			for _, i := range zi {
				if md, ok := property.Attrs[i.Id]; ok {
					i.Metadata.Join(md)
				}
			}
		}
		//设置service的机房权重
		zwValue, err := dm.EncodeZoneWeight(property.ZoneWeights)
		if err != nil {
			r.option.Logger.Error("Etcd GetService EncodeZoneWeight error:", err, serviceName, property.ZoneWeights)
		} else {
			serviceMd[dm.BackendMetadataZoneWeight] = zwValue
		}
	}

	service := &dm.Service{
		ServiceName: serviceName,
		Instances:   instances,
		Metadata:    serviceMd,
		Version:     maxModRevision,
	}
	return service
}

//根据ServiceName获取服务信息（Instance列表),service的版本号采用当前service下所有key中最大的ModRevision
//注意，如果service节点列表为空时，版本号是不准确的，但影响不大
func (r *etcdRegistry) GetService(serviceName string) (*dm.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.option.Timeout)
	defer cancel()

	resp, err := r.client.Get(ctx, getServicePath(serviceName), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	service := r.getServiceAux(serviceName, resp.Kvs)
	return service, nil
}

//列出已注册的所有service列表(该接口比较重，应该是keeper首次启动的时候调用一次,其余时间不会调用)
func (r *etcdRegistry) ListServices() ([]*dm.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.option.Timeout)
	defer cancel()

	resp, err := r.client.Get(ctx, getRootPath(), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	var (
		servicesKvs = map[string][]*mvccpb.KeyValue{}
		services    []*dm.Service
	)

	for _, ev := range resp.Kvs {
		service, _, _, err := parseRawPath(string(ev.Key))
		if err != nil {
			r.option.Logger.Error("Etcd GetService parseRawPath error:", err, string(ev.Key), string(ev.Value))
			continue
		}
		servicesKvs[service] = append(servicesKvs[service], ev)
	}

	for serviceName, kvs := range servicesKvs {
		services = append(services, r.getServiceAux(serviceName, kvs))
	}

	return services, nil
}

//订阅相关的service变化，支持监测所有service,需要抽象一下，不仅仅适用于etcd
func (r *etcdRegistry) Watch() (<-chan *registry.WatchEvent, error) {
	chanEvent := make(chan *registry.WatchEvent, 10)
	go r.watch(chanEvent)
	return chanEvent, nil
}

//@todo watch坑需要研究一下
func (r *etcdRegistry) watch(chanEvent chan *registry.WatchEvent) {
	//注意，watch机制中，如果watch某个key的特定版本，而该版本被etcd compact压缩掉，此时server会关闭chan，这是需要重新创建watcher
ReWatch:
	watcher := clientv3.NewWatcher(r.client)
	watchChan := watcher.Watch(context.TODO(), getRootPath(), clientv3.WithPrefix())
	for {
		select {
		case resp := <-watchChan:
			if err := resp.Err(); err != nil {
				r.option.Logger.Error("Etcd Watch resp error:", err)
				time.Sleep(100 * time.Millisecond)
				goto ReWatch
			}

			fmt.Println("etcd watch:", resp)
			//@todo 检查resp error
			for _, ev := range resp.Events {
				serviceName, infoType, id, err := parseRawPath(string(ev.Kv.Key))
				if err != nil {
					r.option.Logger.Error("Etcd Watch parseRawPath error:", err, string(ev.Kv.Key), string(ev.Kv.Value))
					continue
				}
				var rw registry.WatchEvent
				//信息类型
				switch infoType {
				case infoTypeInstance:
					var (
						instance *dm.Instance
						err      error
					)
					if ev.Type == mvccpb.PUT {
						instance, err = decodeInstance(string(ev.Kv.Value))
						if err != nil {
							r.option.Logger.Error("Etcd Watch decodeInstance error:", err, string(ev.Kv.Key), string(ev.Kv.Value))
							continue
						}
						//新增节点填充后台属性值
						if property, ok := r.properties[instance.ServiceName]; ok {
							if md, ok := property.Attrs[instance.Id]; ok {
								instance.Metadata.Join(md)
							}
						}
					} else {
						instance = &dm.Instance{Id: id, ServiceName: serviceName}
					}
					rw.InfoType = registry.WatchInfoTypeInstance
					rw.Data = instance

					//更新leaseId，在每个server节点保存leaseId
					if err = r.updateLeaseIds(instance, ev.Type); err != nil {
						r.option.Logger.Error("Etcd Watch updateLeaseIds error:", err, instance)
					}

				case infoTypeProperty:
					property, err := decodeProperty(string(ev.Kv.Value))
					if err != nil {
						r.option.Logger.Error("Etcd Watch decodeProperty error:", err, ev)
						continue
					}
					rw.InfoType = registry.WatchInfoTypeProperty
					rw.Data = property
					r.properties[property.ServiceName] = property
				}

				//事件类型
				switch ev.Type {
				case mvccpb.PUT:
					if ev.IsCreate() {
						rw.EventType = registry.WatchEventTypeCreate
					} else if ev.IsModify() {
						rw.EventType = registry.WatchEventTypeModify
					}
				case mvccpb.DELETE:
					rw.EventType = registry.WatchEventTypeDelete
				}

				//设置版本
				rw.Version = ev.Kv.ModRevision
				chanEvent <- &rw
			}
		}
	}
}

//更新leaseId
func (r *etcdRegistry) updateLeaseIds(instance *dm.Instance, typ mvccpb.Event_EventType) error {
	r.Lock()
	defer r.Unlock()

	switch typ {
	case mvccpb.PUT:
		rawLeaseId, ok := instance.Metadata[etcdMetadataLeaseId]
		if !ok {
			return ErrLeaseIdNotFound
		}
		leaseId, err := strconv.ParseInt(rawLeaseId, 10, 64)
		if err != nil {
			return err
		}
		r.leaseIds[instance.Id] = clientv3.LeaseID(leaseId)
	case mvccpb.DELETE:
		delete(r.leaseIds, instance.Id)
	}

	return nil
}

//registry类型
func (r *etcdRegistry) Type() string {
	return "etcd"
}

//属性设置
func (r *etcdRegistry) SetProperty(property *dm.Property) error {
	propertyStr, err := encodeProperty(property)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.option.Timeout)
	defer cancel()

	_, err = r.client.Put(ctx, getPropertyPath(property.ServiceName, "backend"), propertyStr)
	return err
}
