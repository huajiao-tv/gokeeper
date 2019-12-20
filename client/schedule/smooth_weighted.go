package schedule

import (
	"strconv"
	"sync"

	dm "github.com/huajiao-tv/gokeeper/model/discovery"
)

var (
	DefaultWeight = 10 //节点默认权重
)

//平滑的基于权重的负载均衡策略
type smoothWeightedScheduler struct {
	instances []*weightedInstance
	sync.RWMutex
}

func NewSmoothWeightedScheduler() *smoothWeightedScheduler {
	return &smoothWeightedScheduler{}
}

//初始化
func (scheduler *smoothWeightedScheduler) Init(opts ...OpOption) error {
	return nil
}

//service变更后需要执行的操作
func (scheduler *smoothWeightedScheduler) Build(service *dm.Service) error {
	var instances []*weightedInstance
	for _, zoneInstances := range service.Instances {
		for _, instance := range zoneInstances {
			instances = append(instances, &weightedInstance{
				instance:      instance,
				weight:        scheduler.getWeight(instance),
				currentWeight: 0,
			})
		}
	}
	scheduler.Lock()
	defer scheduler.Unlock()

	scheduler.instances = instances
	return nil
}

//获取节点权重
func (scheduler *smoothWeightedScheduler) getWeight(instance *dm.Instance) int {
	wStr, ok := instance.Metadata[dm.BackendMetadataInstanceWeight]
	if !ok {
		return DefaultWeight
	}
	wInt, err := strconv.ParseInt(wStr, 10, 64)
	if err != nil {
		return DefaultWeight
	}
	return int(wInt)
}

//根据调度策略返回一个可用的实例
//@todo 按照机房权重进行分配
func (scheduler *smoothWeightedScheduler) Select() (*dm.Instance, error) {
	scheduler.RLock()
	defer scheduler.RUnlock()

	return scheduler.nextInstance()
}

//带有权重的instance
type weightedInstance struct {
	instance      *dm.Instance
	weight        int
	currentWeight int
}

//根据基于权重的负载均衡算法选取节点，参考 https://studygolang.com/articles/9353 或 https://tenfy.cn/2018/11/12/smooth-weighted-round-robin/
func (scheduler *smoothWeightedScheduler) nextInstance() (*dm.Instance, error) {
	var (
		total = 0
		best  *weightedInstance
	)

	for _, weightedInstance := range scheduler.instances {
		weightedInstance.currentWeight += weightedInstance.weight
		total += weightedInstance.weight
		if best == nil || weightedInstance.currentWeight > best.currentWeight {
			best = weightedInstance
		}
	}

	if best == nil {
		return nil, ErrNoInstance
	}

	best.currentWeight -= total
	return best.instance, nil
}

//scheduler类型
func (scheduler *smoothWeightedScheduler) Type() string {
	return "smooth_weighted"
}
