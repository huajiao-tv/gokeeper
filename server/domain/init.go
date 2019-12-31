package domain

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/huajiao-tv/gokeeper/server/conf"
)

const (
	DefaultDomain         = "DEFAULT_CLUSTER"
	DefaultDomainConfPath = "/tmp/gokeeper/init/%s"
)

func Init() {
	// Init domain
	defaultDomain := os.Getenv(DefaultDomain)
	if defaultDomain == "" {
		return
	}
	if _, err := DomainBooks.GetDomain(defaultDomain); err == nil {
		return
	}
	// Add domain
	DomainBooks.AddDomain(defaultDomain)

	// Add confs
	filePath := fmt.Sprintf(DefaultDomainConfPath, defaultDomain)
	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		return
	}
	for _, file := range files {
		// TODO
		if file.IsDir() {
			continue
		}
		if conf.Ignore(file.Name(), false) {
			continue
		}
		content, err := ioutil.ReadFile(path.Join(filePath, file.Name()))
		if err != nil {
			continue
		}
		_ = conf.AddFile(defaultDomain, file.Name(), string(content), "gokeeper booting init")
	}
}
