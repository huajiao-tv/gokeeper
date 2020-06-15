package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/huajiao-tv/gokeeper/server/api/apihttp"
	"github.com/huajiao-tv/gokeeper/server/conf"
)

func writeFiles(files map[string]string, path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	for k, v := range files {
		fname := filepath.Join(path, fmt.Sprintf("%s.go", k))
		err := ioutil.WriteFile(fname, []byte(v), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func getConfManagerFromNetwork(host, domain string) (*conf.ConfManager, error) {
	params := map[string]string{"domain": domain}
	resp, err := apihttp.Call(host, "/conf/list", "GET", params)
	if err != nil {
		return nil, err
	}
	if resp["error_code"].Int() != 0 {
		return nil, errors.New("request error:" + resp["error"].String())
	}

	var files []conf.File
	err = json.Unmarshal([]byte(resp["data"].String()), &files)
	if err != nil {
		return nil, err
	}
	//filter agent.conf
	var filteredFiles []conf.File
	for _, f := range files {
		if strings.Contains(f.Name, "agent.conf") {
			continue
		}
		filteredFiles = append(filteredFiles, f)
	}
	return conf.WrapConfManager(filteredFiles)
}
