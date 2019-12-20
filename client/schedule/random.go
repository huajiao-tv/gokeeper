package schedule

import (
	"math/rand"
	"sync"

	dm "github.com/huajiao-tv/gokeeper/model/discovery"
)

//负载均衡随机策略
type randomScheduler struct {
	instances []*dm.Instance
	sync.RWMutex
}

func NewRandomScheduler() *randomScheduler {
	return &randomScheduler{}
}

//初始化
func (scheduler *randomScheduler) Init(opts ...OpOption) error {
	return nil
}

//service变更后需要执行的操作,比如Instance根据机房的流量权重进行重新计算
//@todo 按照机房权重进行分配
func (scheduler *randomScheduler) Build(service *dm.Service) error {
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
func (scheduler *randomScheduler) Select() (*dm.Instance, error) {
	scheduler.RLock()
	defer scheduler.RUnlock()

	l := len(scheduler.instances)
	if l == 0 {
		return nil, ErrNoInstance
	}
	instance := scheduler.instances[rand.Intn(l)]
	return instance, nil
}

//scheduler类型
func (scheduler *randomScheduler) Type() string {
	return "random"
}
