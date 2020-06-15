package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/huajiao-tv/gokeeper/server/tests/dev/dev_client/conf/data"
)

var (
	goRpcAddr  = flag.String("go", "127.0.0.1:7000", "connect keeper with goRpc")
	grpcAddr   = flag.String("g", "127.0.0.1:7001", "connect keeper with grpc")
	adminAddr  = flag.String("a", "127.0.0.1:17000", "add domain and change config")
	withChange = flag.Bool("w", true, "update config with adminAddr,if not,you can update in the backend")

	goRpcChangeNum = 0
	grpcChangeNum  = 0
)

const component = "dev_client"

func init() {
	if err := initConfig(); err != nil {
		panic("init config failed:" + err.Error())
	}
	//wait for server sync config
	time.Sleep(time.Second)
}

func main() {
	success := true
	connectGoRpc()
	connectGrpc()

	//wait for callback
	time.Sleep(time.Second * 1)
	if goRpcChangeNum != 1 || grpcChangeNum != 1 {
		fmt.Println("load config failed!")
		success = false
	}

	if *withChange {
		fmt.Println("change the config,add one zero behind the key timeout")
		err := changeConfig("timeout", "string", data.CurrentTest().Timeout+"0")
		if err != nil {
			fmt.Println("change config failed:", err)
			success = false
		}

		//wait for callback
		time.Sleep(time.Second)
		if goRpcChangeNum < 2 || grpcChangeNum < 2 {
			fmt.Println("do not callback when config change!")
			success = false
		}
	}
	if success {
		fmt.Println("config test success!")
	} else {
		fmt.Println("config test failed!")
	}
}
