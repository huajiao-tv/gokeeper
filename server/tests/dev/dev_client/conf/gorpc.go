package main

import (
	"fmt"
	"os"

	gokeeper "github.com/huajiao-tv/gokeeper/client"
	"github.com/huajiao-tv/gokeeper/server/tests/dev/dev_client/conf/data"
)

func connectGoRpc() {
	sections := []string{"test.conf/DEFAULT"}
	node, err := os.Hostname()
	if err != nil {
		fmt.Println("get hostname failed:", err)
	}
	client := gokeeper.New(*goRpcAddr, domain, node+"gorpc", component, sections, nil)
	client.LoadData(data.ObjectsContainer).RegisterCallback(GoRpcCallBack)
	if err := client.Work(); err != nil {
		panic(err)
	}

}
func GoRpcCallBack() {
	goRpcChangeNum++
	fmt.Println("gorpc config change:", data.CurrentTest())
}
