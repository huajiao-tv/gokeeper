package main

import (
	"syscall"
	"time"

	cd "github.com/huajiao-tv/gokeeper/client/discovery"
)

func registryService(close <-chan int) {
	instance := cd.NewInstance(cd.GenRandomId(), service, map[string]string{cd.SchemaHttp: "127.0.0.1:7000", cd.SchemaRpc: "127.0.0.1:7001"})
	instance.Id = "test_id_1"

	client := cd.New(
		*discoveryAddr,
		cd.WithRegistry(instance),
		cd.WithRegistryTTL(60*time.Second),
	)
	//收到信息解注册，否则退出后节点在一段时间内仍被发现
	client.SignalDeregister(true, syscall.SIGINT, syscall.SIGTERM)
	client.Work()

	select {
	case <-close:
		client.Deregister()
	}
}
