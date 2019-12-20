package main

import (
	"fmt"
	"log"

	"github.com/huajiao-tv/gokeeper/agent/data"
	gokeeper "github.com/huajiao-tv/gokeeper/client"
	"github.com/huajiao-tv/gokeeper/utility/process"
)

func main() {
	gokeeper.Debug = true
	client := gokeeper.New(KeeperAddr, Domain, NodeID, Component, []string{"agent.conf/" + NodeID}, nil)
	client.LoadData(data.ObjectsContainer)
	err := client.Work()
	if err != nil {
		log.Fatal("| [error] start agent failed | client.Work:", err)
	}
	if err := initConfig(); err != nil {
		log.Fatal("| [error] start agent failed | initConfig:", err)
	}
	if err := savePid(); err != nil {
		log.Fatal("| [error] start agent failed | savePid:", err)
	}
	if err := initLogger(LogPath); err != nil {
		log.Fatal("| [error] start agent failed | initLogger:", err)
	}

	Logex.Trace("agent BasePath", BasePath)
	Logex.Trace("agent LogPath", LogPath)
	Logex.Trace("agent TmpPath", TmpPath)
	Logex.Trace("agent ComponentBinPath", ComponentBinPath)
	Logex.Trace("agent ComponentLogPath", ComponentLogPath)

	select {}
}

func savePid() error {
	pidFile := fmt.Sprintf("%s/%s-%s.pid", TmpPath, Component, NodeID)
	if err := process.SavePid(pidFile); err != nil {
		return err
	}
	return nil
}
