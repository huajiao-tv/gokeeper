package schedule

import (
	"errors"

	dm "github.com/huajiao-tv/gokeeper/model/discovery"
)

var (
	ErrNoInstance = errors.New("service has no instance")
)

type option struct {
}

type OpOption func(*option)

//client 选取服务节点的interface
//TODO 如果调用某个服务的失败次数过多，是否适当降低该节点权重
type Scheduler interface {
	//初始化
	Init(opts ...OpOption) error
	//service变更后需要执行的操作,比如Instance根据机房的流量权重进行重新计算
	Build(service *dm.Service) error
	//根据调度策略返回一个可用的实例
	Select() (*dm.Instance, error)
	//scheduler类型
	Type() string
}
