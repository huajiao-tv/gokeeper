package schedule

import (
	"sync"
	"sync/atomic"

	dm "github.com/huajiao-tv/gokeeper/model/discovery"
)

//负载均衡轮询策略
type roundRobinScheduler struct {
	instances []*dm.Instance
	index     int64
	sync.RWMutex
}

func NewRoundRobinScheduler() *roundRobinScheduler {
	return &roundRobinScheduler{}
}

//初始化
func (scheduler *roundRobinScheduler) Init(opts ...OpOption) error {
	return nil
}

//service变更后需要执行的操作,比如Instance根据机房的流量权重进行重新计算
//@todo 按照机房权重进行分配
func (scheduler *roundRobinScheduler) Build(service *dm.Service) error {
	var instances []*dm.Instance
	for _, zoneInstances := range service.Instances {
		for _, instance := range zoneInstances {
			instances = append(instances, instance)
		}
	}
	scheduler.Lock()
	defer scheduler.Unlock()

	scheduler.instances = instances
	return nil
}

//根据调度策略返回一个可用的实例
//@todo 按照机房权重进行分配
func (scheduler *roundRobinScheduler) Select() (*dm.Instance, error) {
	scheduler.RLock()
	defer scheduler.RUnlock()

	l := len(scheduler.instances)
	if l == 0 {
		return nil, ErrNoInstance
	}

	if scheduler.index >= int64(l) {
		scheduler.index = 0
	}
	instance := scheduler.instances[int(scheduler.index)%l] //从数组中获取时，需要考虑并发的情况，需要采用(index%len)的方式，否则可能会崩溃
	atomic.AddInt64(&scheduler.index, 1)

	return instance, nil
}

//scheduler类型
func (scheduler *roundRobinScheduler) Type() string {
	return "round_robin"
}
