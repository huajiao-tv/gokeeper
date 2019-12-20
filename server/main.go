package server

import (
	"fmt"

	"github.com/huajiao-tv/gokeeper/server/discovery"
	Kdomain "github.com/huajiao-tv/gokeeper/server/domain"
	"github.com/huajiao-tv/gokeeper/server/logger"
	"github.com/huajiao-tv/gokeeper/server/metrics"
	"github.com/huajiao-tv/gokeeper/server/service"
	"github.com/huajiao-tv/gokeeper/server/setting"
	"github.com/huajiao-tv/gokeeper/server/storage"
	"github.com/huajiao-tv/gokeeper/server/sync"
)

func Main() {
	if err := setting.InitConfig(); err != nil {
		panic(err)
	}
	if err := logger.InitLogger(setting.LogPath, setting.Component); err != nil {
		panic(err)
	}

	if err := storage.InitStorage(setting.StorageUrl, setting.StorageUsername, setting.StoragePassword, setting.EventMode, logger.Logex); err != nil {
		panic(err)
	}

	domainNames, err := Kdomain.InitDomainConf()
	if err != nil {
		panic(err)
	}
	//Kdomain.DomainConfs.Debug()

	if err := Kdomain.InitDomainBook(domainNames, setting.EventInterval); err != nil {
		panic(err)
	}

	if err = logger.SavePid(); err != nil {
		panic(err)
	}

	//keeper addr lease
	go storage.KStorage.KeepAlive(setting.KeeperID, setting.KeeperAdminAddr)

	// init metrics
	metrics.Init(setting.PromListen)

	// discovery 初始化
	discovery.InitSessionBook()

	if err := discovery.InitServiceBook(); err != nil {
		panic("InitServiceBook error:" + err.Error())
	}
	//start discovery
	//service.StartDiscoveryServer()

	// start server
	service.StartSyncServer(true)

	go sync.Watch()

	fmt.Printf("start finish")
}
