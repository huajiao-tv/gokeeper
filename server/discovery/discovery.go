package discovery

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	dm "github.com/huajiao-tv/gokeeper/model/discovery"
	pb "github.com/huajiao-tv/gokeeper/pb/go"
	"google.golang.org/grpc/status"
)

const (
	ErrCodeInternal = 500 //内部错误

	//重启n秒时间内polls请求拒绝，主要处理discovery server长时间停掉导致后端存储服务列表为空情况
	//在该时间内，client节点会重新调用keepalive接口进行注册，正常情况下，该时间段后，服务注册成功
	restartPollsUnWorkingDuration = 5 * time.Second
)

var (
	pollsWorking = false
)

func init() {
	time.AfterFunc(restartPollsUnWorkingDuration, func() {
		pollsWorking = true
	})
}

type Server struct{}

func paramsError(params ...string) error {
	pStr := strings.Join(params, ",")
	return status.Errorf(ErrCodeInternal, fmt.Sprintf("%s is invalid", pStr))
}

//服务注册接口,服务就绪后调用
func (s *Server) Register(ctx context.Context, req *pb.RegisterReq) (*empty.Empty, error) {
	if req.Instance == nil {
		return nil, paramsError("instance")
	}

	instance := dm.ParseInstance(req.Instance)
	registryTTL := time.Duration(req.LeaseSecond) * time.Second
	err := Register(instance, registryTTL)
	if err != nil {
		return &empty.Empty{}, status.Errorf(ErrCodeInternal, err.Error())
	}

	return &empty.Empty{}, nil
}

//服务心跳接口
func (s *Server) KeepAlive(ctx context.Context, req *pb.KeepAliveReq) (*empty.Empty, error) {
	if req.Instance == nil {
		return nil, paramsError("instance")
	}

	instance := dm.ParseInstance(req.Instance)
	registryTTL := time.Duration(req.LeaseSecond) * time.Second
	err := KeepAlive(instance, registryTTL)
	if err != nil {
		return &empty.Empty{}, status.Errorf(ErrCodeInternal, err.Error())
	}

	return &empty.Empty{}, nil
}

//服务解除注册接口，服务下线时调用
func (s *Server) Deregister(ctx context.Context, req *pb.DeregisterReq) (*empty.Empty, error) {
	if req.Instance == nil {
		return nil, paramsError("instance")
	}

	instance := dm.ParseInstance(req.Instance)
	err := Deregister(instance)
	if err != nil {
		return &empty.Empty{}, status.Errorf(ErrCodeInternal, err.Error())
	}

	return &empty.Empty{}, nil
}

//批量获取服务的节点信息,如果没有相关数据更新，则阻塞一段时间
//@todo 是否需要client携带业务标识，方便后台查看哪些服务订阅了某个服务？
func (s *Server) Polls(stream pb.Discovery_PollsServer) error {
	//在重启的一段时间内，拒绝poll请求，防止后端存储列表为空导致client获取不到节点（该时间段内，client节点会重新注册)
	if !pollsWorking {
		return status.Errorf(ErrCodeInternal, "server is restarting")
	}

	err := Polls(stream)
	if err != nil {
		return status.Errorf(ErrCodeInternal, err.Error())
	}
	return nil
}
