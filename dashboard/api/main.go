package main

import (
	"github.com/gin-gonic/gin"
	"github.com/huajiao-tv/gokeeper/dashboard/api/conf"
	"github.com/huajiao-tv/gokeeper/dashboard/api/controllers"
	"github.com/huajiao-tv/gokeeper/dashboard/api/models"
)

func main() {
	err := conf.Init()
	if err != nil {
		panic("init config fail: " + err.Error())
	}

	err = models.Init(conf.KeeperAdmin)
	if err != nil {
		panic("init model fail: " + err.Error())
	}

	r := gin.Default()
	err = controllers.Init(r)
	if err != nil {
		panic("init controller fail: " + err.Error())
	}

	err = r.Run(conf.Listen)
	if err != nil {
		panic("listen gin fail: " + err.Error())
	}

	select {}
}
