package domain

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/huajiao-tv/gokeeper/server/conf"
)

const (
	DefaultDomainConfPath = "/tmp/gokeeper/init/"
)

func Init() {
	infos, err := ioutil.ReadDir(DefaultDomainConfPath)
	if err != nil || len(infos) == 0 {
		fmt.Println("find no init domains")
		return
	}
	successInitDomains := make([]string, 0, len(infos))
	for _, info := range infos {
		if !info.IsDir() {
			continue
		}
		if []byte(info.Name())[0] == '.' {
		}
		Domain := info.Name()

		if _, err := DomainBooks.GetDomain(Domain); err == nil {
			return
		}

		// Add confs
		filePath := DefaultDomainConfPath + Domain
		files, err := ioutil.ReadDir(filePath)
		if err != nil || len(files) == 0 {
			continue
		}
		success := false
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
			err = conf.AddFile(Domain, file.Name(), string(content), "gokeeper booting init")
			if err != nil {
				fmt.Println("init domain", Domain, "file", file.Name(), "failed:", err)
			} else {
				success = true
			}
		}
		if success {
			successInitDomains = append(successInitDomains, Domain)
		}
	}
	fmt.Println("init domains:", successInitDomains)
}
