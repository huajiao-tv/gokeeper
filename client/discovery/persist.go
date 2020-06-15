package discovery

import (
	"io/ioutil"

	dm "github.com/huajiao-tv/gokeeper/model/discovery"
	"gopkg.in/yaml.v2"
)

//持久化服务列表
func writeServices(persistFile string, services map[string]*dm.Service) error {
	data, err := yaml.Marshal(services)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(persistFile, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

//加载服务列表
func readServices(persistFile string) (map[string]*dm.Service, error) {
	data, err := ioutil.ReadFile(persistFile)
	if err != nil {
		return nil, err
	}

	services := make(map[string]*dm.Service)
	err = yaml.Unmarshal(data, services)
	if err != nil {
		return nil, err
	}
	return services, nil
}
