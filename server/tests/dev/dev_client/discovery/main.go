package main

import (
	"flag"
	"fmt"
	"time"
)

var (
	discoveryAddr = flag.String("d", "127.0.0.1:7001", "use to connect keeper discovery")
	service       = "demo.test.com"
)

func main() {
	success := true
	//use to deregister service
	closeChan := make(chan int, 2)

	//use to observe if consumer can discover the service address
	downChan := make(chan int, 3)
	go registryService(closeChan)
	go discoverService(downChan)
	t := time.NewTicker(time.Second * 3)
	select {
	case <-downChan:
		success = false
		fmt.Println("discover service failed!")
	case <-t.C:
	}
	closeChan <- 1
	t = time.NewTicker(time.Second * 3)
	select {
	case <-downChan:
	case <-t.C:
		success = false
		fmt.Println("discover a deregister service!")
	}
	if success {
		fmt.Println("discovery test success!")
	} else {
		fmt.Println("discovery test failed!")
	}

}
