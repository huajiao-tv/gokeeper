package apihttp

import (
	"encoding/json"
	"fmt"

	dm "github.com/huajiao-tv/gokeeper/model/discovery"
	"github.com/huajiao-tv/gokeeper/server/discovery"
	"github.com/huajiao-tv/gokeeper/server/logger"
)

//获取服务信息
func (s *ServiceController) DiscoveryGetServiceAction() {
	if ok := s.required("service_name"); !ok {
		return
	}

	serviceName := s.query("service_name")
	service, err := discovery.GetService(serviceName)
	if err != nil {
		logger.Logex.Error("DiscoveryGetService", "GetService", err.Error())
		s.renderJSON(Resp{ErrorCode: 1, Error: fmt.Sprintf("DiscoveryGetService error:%s", err.Error())})
		return
	}

	s.renderJSON(Resp{Data: service})
}

//获取服务列表
func (s *ServiceController) DiscoveryListServicesAction() {
	services, err := discovery.ListServices()
	if err != nil {
		logger.Logex.Error("DiscoveryListServices", "ListServices", err.Error())
		s.renderJSON(Resp{ErrorCode: 1, Error: fmt.Sprintf("ListServices error:%s", err.Error())})
		return
	}
	s.renderJSON(Resp{Data: services})
}

//后台设置属性
//更新metadata时，key需要以 backend-metadata- 开始，否则不生效
func (s *ServiceController) DiscoverySetPropertyAction() {
	body, err := s.readBody()
	if err != nil {
		logger.Logex.Error("DiscoverySetProperty", "readBody", err.Error())
		s.renderJSON(Resp{ErrorCode: 1, Error: fmt.Sprintf("readBody error:%s", err.Error())})
		return
	}

	var property dm.Property
	if err := json.Unmarshal([]byte(body), &property); err != nil {
		logger.Logex.Error("DiscoverySetProperty", "Unmarshal property", err.Error())
		s.renderJSON(Resp{ErrorCode: 1, Error: fmt.Sprintf("decode property error:%s, body:%s", err.Error(), body)})
		return
	}

	if err := discovery.SetProperty(&property); err != nil {
		logger.Logex.Error("DiscoverySetProperty", "SetProperty", err.Error())
		s.renderJSON(Resp{ErrorCode: 1, Error: fmt.Sprintf("discovery.SetProperty error:%s", err.Error())})
		return
	}

	s.renderJSON(Resp{})
	return
}
