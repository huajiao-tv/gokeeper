package main

import (
	"fmt"
	"path/filepath"

	"github.com/huajiao-tv/gokeeper/utility/logger"
)

var (
	Logex *logger.Logger
)

func initLogger(logPath string) error {
	var err error
	logfile := filepath.Join(logPath, fmt.Sprintf("%s-%s", Component, NodeID))
	logBackupPath := filepath.Join(logPath, "/backup/")
	Logex, err = logger.NewLogger(logfile, "gokeeper|"+Component+"|"+NodeID, logBackupPath)
	if err != nil {
		return err
	}
	return nil
}
