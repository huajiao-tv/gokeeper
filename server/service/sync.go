package service

import (
	"net"

	pb "github.com/huajiao-tv/gokeeper/pb/go"
	"github.com/huajiao-tv/gokeeper/server/setting"
	"github.com/huajiao-tv/gokeeper/server/sync"
	"github.com/johntech-o/gorpc"
	"google.golang.org/grpc"
)

//@todo 完善server初始化部分内容,优化代码
var (
	grpcListener net.Listener //grpc中配置和服务发现的端口暂时共用
)

//启动sync server，默认开启gorpc协议的server
func StartSyncServer(withGrpc bool) {
	go startPSync()

	if withGrpc {
		go startGSync()
	}
}

//启动pepper gorpc协议的server
func startPSync() {
	s := gorpc.NewServer(setting.GoRpcListen)
	err := s.Register(new(sync.Server))
	if err != nil {
		panic("psync register error:" + err.Error())
	}
	s.Serve()
	panic("start gorpc service error")
}

//启动google grpc协议的server @todo 完善server
func startGSync() {
	s := grpc.NewServer([]grpc.ServerOption{}...)
	pb.RegisterSyncServer(s, &sync.GSyncServer{})
	registerDiscovery(s)
	err := s.Serve(setting.GrpcListener)
	if err != nil {
		panic("start grpc service error:" + err.Error())
	}
}
