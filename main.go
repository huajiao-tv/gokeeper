package main

import (
	"github.com/huajiao-tv/gokeeper/server"
	"github.com/huajiao-tv/gokeeper/server/api/apihttp"
)

func main() {
	server.Main()
	go apihttp.StartHttpServer()

	select {}
}
