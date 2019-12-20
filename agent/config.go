package main

import (
	"errors"
	"flag"
	"os"
	"path/filepath"

	"github.com/huajiao-tv/gokeeper/agent/data"
)

//
const (
	Component = "agent"
)

//
var (
	BasePath string
	LogPath  = "log"
	TmpPath  = "tmp"

	ComponentBinPath string
	ComponentLogPath string
	ComponentPidPath string

	Domain     string
	KeeperAddr string
	NodeID     string
)

func init() {
	flag.StringVar(&Domain, "d", "", "domain name")
	flag.StringVar(&NodeID, "n", "", "current node id")
	flag.StringVar(&KeeperAddr, "k", "", "keeper address ip:port")

	flag.Parse()
}

func initConfig() error {
	if data.CurrentAgent().BasePath == "" {
		return errors.New("base path is empty")
	}
	LogPath = filepath.Join(data.CurrentAgent().BasePath, "log")
	TmpPath = filepath.Join(data.CurrentAgent().BasePath, "tmp")
	ComponentPidPath = filepath.Join(TmpPath, "component")

	ComponentBinPath = data.CurrentAgent().ComponentBinPath
	ComponentLogPath = data.CurrentAgent().ComponentLogPath

	if ComponentBinPath == "" {
		return errors.New("component bin path is empty")
	}
	if ComponentLogPath == "" {
		return errors.New("component log path is empty")
	}

	if err := mkDir(); err != nil {
		return err
	}

	return nil
}

func mkDir() error {
	var err error
	if err = os.MkdirAll(LogPath, 0744); err != nil {
		return err
	}
	if err = os.MkdirAll(TmpPath, 0744); err != nil {
		return err
	}
	if err = os.MkdirAll(ComponentPidPath, 0744); err != nil {
		return err
	}
	return nil
}
