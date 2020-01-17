package main

import (
	"fmt"
	"time"

	cd "github.com/huajiao-tv/gokeeper/client/discovery"
	"github.com/huajiao-tv/gokeeper/client/schedule"
)

func discoverService(down chan<- int) {

	client := cd.New(
		*discoveryAddr,
		cd.WithDiscovery("example_client1", []string{service}),
		cd.WithScheduler(map[string]schedule.Scheduler{
			//"demo.test.com": schedule.NewRandomScheduler(),
			"demo.test.com": schedule.NewRandomScheduler(),
			//"demo.test.com": schedule.NewRoundRobinScheduler(),
		}),
		cd.WithPersistence(),
	)
	client.Work()

	for {
		addr, err := client.GetServiceAddr(service, cd.SchemaHttp)
		fmt.Println("addr:", addr, "err:", err)
		if err != nil || addr == "" {
			down <- 1
		}
		time.Sleep(1 * time.Second)
	}
	select {}
}
