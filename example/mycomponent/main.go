package main

import (
	"flag"
	"fmt"
	"time"

	gokeeper "github.com/huajiao-tv/gokeeper/client"
	"github.com/huajiao-tv/gokeeper/example/mycomponent/data"
)

var (
	component = "mycomponent"

	keeperAddr string
	domain     string
	nodeID     string
)

func main() {
	flag.StringVar(&domain, "d", "mydomain", "domain name")
	flag.StringVar(&nodeID, "n", "node1", "current node id")
	flag.StringVar(&keeperAddr, "k", "127.0.0.1:7001", "keeper address ip:port")
	flag.Parse()

	sections := []string{"test.conf/127.0.0.2:80"}
	client := gokeeper.New(keeperAddr, domain, nodeID, component, sections, nil, gokeeper.WithGrpc())
	client.LoadData(data.ObjectsContainer).RegisterCallback(run)
	if err := client.Work(); err != nil {
		panic(err)
	}

	select {}
}

func run() {
	fmt.Println(fmt.Sprintf("%#v", data.CurrentTest()))
	time.Sleep(2 * time.Second)
}
