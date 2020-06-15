package process

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func GetProcessBinaryDir() (string, error) {
	var dir, p string
	var err error
	pid := os.Getpid()
	lnk := "/proc/" + strconv.Itoa(pid) + "/exe"
	p, err = os.Readlink(lnk)
	if err != nil {
		return "", err
	}
	dir = filepath.Dir(p)
	dir = strings.Replace(dir, "\\", "/", -1)
	return dir, nil
}
