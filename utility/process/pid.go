package process

import (
	"io/ioutil"
	"os"
	"strconv"
)

func SavePid(pidFile string) error {
	pid := os.Getpid()
	pidString := strconv.Itoa(pid)
	if err := ioutil.WriteFile(pidFile, []byte(pidString), 0777); err != nil {
		return err
	}
	return nil
}
