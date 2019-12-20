package service

import (
	pb "github.com/huajiao-tv/gokeeper/pb/go"
	"github.com/huajiao-tv/gokeeper/server/discovery"
	"github.com/huajiao-tv/gokeeper/server/setting"
	"google.golang.org/grpc"
)

func StartDiscoveryServer() {
	s := grpc.NewServer([]grpc.ServerOption{}...)
	go func() {
		err := s.Serve(setting.GrpcListener)
		if err != nil {
			panic("start grpc service error:" + err.Error())
		}
	}()
}

func registerDiscovery(server *grpc.Server) {
	pb.RegisterDiscoveryServer(server, &discovery.Server{})
}
