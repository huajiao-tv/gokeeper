package main

import (
	"fmt"
	"os"

	gokeeper "github.com/huajiao-tv/gokeeper/client"
	"github.com/huajiao-tv/gokeeper/server/tests/dev/dev_client/conf/data2"
)

func connectGrpc() {
	sections := []string{"test.conf/DEFAULT"}
	node, err := os.Hostname()
	if err != nil {
		fmt.Println("get hostname failed:", err)
	}
	client := gokeeper.New(*grpcAddr, domain, node+"grpc", component, sections, nil, gokeeper.WithGrpc())
	client.LoadData(data2.ObjectsContainer).RegisterCallback(GrpcCallBack)
	if err := client.Work(); err != nil {
		panic(err)
	}
}
func GrpcCallBack() {
	grpcChangeNum++
	fmt.Println("gprc config change:", data2.CurrentTest())
}
