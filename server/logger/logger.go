package logger

import (
	"path/filepath"

	"github.com/huajiao-tv/gokeeper/server/setting"
	"github.com/huajiao-tv/gokeeper/utility/logger"
	"github.com/huajiao-tv/gokeeper/utility/process"
)

var (
	Logex *logger.Logger
)

func InitLogger(logPath, logName string) error {
	var err error
	logfile := filepath.Join(logPath, logName)
	logBackupPath := filepath.Join(logPath, "/backup/")
	Logex, err = logger.NewLogger(logfile, "gokeeper|master|", logBackupPath)
	if err != nil {
		return err
	}
	return nil
}

func SavePid() error {
	pidFile := "gokeeper.pid"
	pidFile = filepath.Join(setting.TmpPath, pidFile)
	if err := process.SavePid(pidFile); err != nil {
		return err
	}
	return nil
}
