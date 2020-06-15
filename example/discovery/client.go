package main

import (
	"fmt"
	"syscall"
	"time"

	cd "github.com/huajiao-tv/gokeeper/client/discovery"
	"github.com/huajiao-tv/gokeeper/client/schedule"
)

func main() {
	//@todo 重启时通过信号进行解注册，后续需要封装在sdk中
	//注意，ID需要确保每个节点是唯一的
	//注意，最好确保ID在重复启停过程中是不变的，这样才能记忆上次记录的权重（基于权重的负载均衡策略)@todo 后续权重是否要基于ip
	instance := cd.NewInstance(cd.GenRandomId(), "demo.test.com", map[string]string{cd.SchemaHttp: "127.0.0.1:17000", cd.SchemaRpc: "127.0.0.1:17001"})
	instance.Id = "test_id_1"
	fmt.Println("instance:", instance)

	client := cd.New(
		"127.0.0.1:7001",
		cd.WithRegistry(instance),
		cd.WithRegistryTTL(60*time.Second),
		cd.WithDiscovery("example_client1", []string{"demo.test.com"}),
		cd.WithScheduler(map[string]schedule.Scheduler{
			//"demo.test.com": schedule.NewRandomScheduler(),
			"demo.test.com": schedule.NewRandomScheduler(),
			//"demo.test.com": schedule.NewRoundRobinScheduler(),
		}),
		cd.WithPersistence(),
	)
	//收到信息解注册，否则退出后节点在一段时间内仍被发现
	client.SignalDeregister(true, syscall.SIGINT, syscall.SIGTERM)
	client.Work()

	for {
		addr, err := client.GetServiceAddr("demo.test.com", cd.SchemaHttp)
		fmt.Println("addr:", addr, "err:", err)
		time.Sleep(2 * time.Second)
	}
	select {}
}

//根据instance status下掉某个节点
//机房权重
//keeper停止服务一段时间怎么办

/*func main() {
	instance := cd.NewInstance(cd.GenRandomId(), "demo.test.com", map[string]string{cd.SchemaHttp: "127.0.0.1:17001"})
	instance.Id = "test_id_2"
	fmt.Println("instance:", instance)

	client := cd.New(
		"127.0.0.1:7001",
		cd.WithRegistry(instance),
		cd.WithRegistryTTL(60*time.Second),
		//cd.WithDiscovery([]string{"demo.test.com"}),
		//cd.WithScheduler(map[string]schedule.Scheduler{
		//	"demo.test.com": schedule.NewRandomScheduler(),
		//}),
	)
	//收到信息解注册，否则退出后节点在一段时间内仍被发现
	client.SignalDeregister(true, syscall.SIGINT, syscall.SIGTERM)
	client.Work()
	select {}
}*/
