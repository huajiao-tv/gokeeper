package apihttp

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/huajiao-tv/gokeeper/model"
	"github.com/huajiao-tv/gokeeper/server/conf"
	Kdomain "github.com/huajiao-tv/gokeeper/server/domain"
	"github.com/huajiao-tv/gokeeper/server/logger"
	"github.com/huajiao-tv/gokeeper/server/setting"
	"github.com/huajiao-tv/gokeeper/server/storage"
	"github.com/huajiao-tv/gokeeper/server/storage/etcd"
	confSync "github.com/huajiao-tv/gokeeper/server/sync"
)

var (
	confManageLock = new(sync.Mutex)
)

func (s *ServiceController) ConfListAction() {
	qDomain := s.query("domain")
	cf, err := Kdomain.DomainConfs.GetDomain(qDomain)
	if err != nil {
		logger.Logex.Error("ConfList", "GetDomain", err.Error())
		s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
		return
	}
	s.renderJSON(Resp{Data: cf.FileList()})
}

func (s *ServiceController) AddFileAction() {
	confManageLock.Lock()
	defer confManageLock.Unlock()

	if ok := s.required("domain", "file", "conf", "note"); !ok {
		return
	}

	qDomain := s.query("domain")
	qFile := s.query("file")
	qConf := s.query("conf")
	qNote := s.query("note")

	//qFile example: sr-bjcc/global.conf
	if conf.Ignore(qFile, false) {
		s.renderJSON(Resp{ErrorCode: 1, Error: fmt.Sprintf("file name %s is invalid", qConf)})
		return
	}

	if err := conf.AddFile(qDomain, qFile, qConf, qNote); err != nil {
		logger.Logex.Error("AddFileAction", "AddFile", err.Error())
		s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
		return
	}

	s.renderJSON(Resp{})
}

func (s *ServiceController) ConfManageAction() {
	confManageLock.Lock()
	defer confManageLock.Unlock()

	if ok := s.required("domain", "operates"); !ok {
		return
	}
	qDomain := s.query("domain")
	qNote := s.query("note")
	qOperate := s.query("operates")

	var operates []model.Operate
	if err := json.Unmarshal([]byte(qOperate), &operates); err != nil {
		logger.Logex.Error("ConfManageAction", "json.Unmarshal", err.Error())
		s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
		return
	}

	var notes []string
	if err := json.Unmarshal([]byte(qNote), &notes); err != nil {
		//do nothing
	}

	var operate model.Operate
	var evt model.Event
	if len(operates) == 1 {
		operate = operates[0]
		operate.Domain = qDomain
		if len(notes) > 0 {
			operate.Note = notes[0]
		} else if len(qNote) > 0 {
			operate.Note = qNote
		}
		evt.EventType = model.EventOperate
		evt.Data = operate
	} else {
		for k := range operates {
			operates[k].Domain = qDomain
			if k < len(notes) {
				operates[k].Note = notes[k]
			}
		}
		evt.EventType = model.EventOperateBatch
		evt.Data = operates
	}

	if err := confSync.EventOperateProxy(evt); err != nil {
		logger.Logex.Error("ConfManageAction", "EventOperateProxy", err.Error())
		s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
		return
	}

	s.renderJSON(Resp{})
}

func (s *ServiceController) PackageListAction() {
	qDomain := s.query("domain")
	if qDomain == "" {
		s.renderJSON(Resp{ErrorCode: 1, Error: "domain is required"})
		return
	}
	qOffset := s.query("offset")
	qLimit := s.query("limit")
	offset, _ := strconv.ParseInt(qOffset, 10, 64)
	limit, _ := strconv.ParseInt(qLimit, 10, 64)
	if limit == 0 {
		limit = 50
	}

	data, err := storage.KStorage.GetHistoryVersions(qDomain, limit, offset, true)
	if err != nil {
		s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
		return
	}

	//如果更换了底层存储，需要修改这个地方
	recodes, ok := data.([]*etcd.Recode)
	if !ok {
		s.renderJSON(Resp{Data: []*etcd.Recode{}})
		return
	}

	s.renderJSON(Resp{Data: recodes})
}

func (s *ServiceController) ConfRollbackAction() {
	confManageLock.Lock()
	defer confManageLock.Unlock()

	//为兼容老版gokeepr， 将version更改为id
	if ok := s.required("domain", "id"); !ok {
		return
	}
	qDomain := s.query("domain")
	qVersion := s.query("id")

	version, err := strconv.ParseInt(qVersion, 10, 64)
	if err != nil {
		logger.Logex.Error("ConfRollback", "strconv.Atoi", err.Error())
		s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
		return
	}

	qNote := fmt.Sprintf("rollback to version:%s", qVersion)
	evt := model.Event{
		EventType: model.EventOperateRollback,
		Data:      model.Operate{Domain: qDomain, Version: version, Note: qNote},
	}
	if err := confSync.EventOperateProxy(evt); err != nil {
		logger.Logex.Error("ConfRollback", "EventOperateProxy", err.Error())
		s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
		return
	}
	s.renderJSON(Resp{})
}

func (s *ServiceController) ConfStatusAction() {
	qDomain := s.query("domain")
	domain, err := Kdomain.DomainBooks.GetDomain(qDomain)
	if err != nil {
		logger.Logex.Error("ConfStatus", "GetDomain", err.Error(), qDomain)
		s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
		return
	}

	qTransit := s.query("transit")
	var nodes []model.NodeInfo
	if qTransit != "true" {
		hosts, err := getAdminAddrs(qDomain)
		if err != nil {
			logger.Logex.Error("ConfStatus", "GetKeeperAddrs", err.Error(), qDomain)
			s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
			return
		}
		for _, host := range hosts {
			if host != setting.KeeperAdminAddr {
				resp := transitConfStatus(host, qDomain)
				if resp.ErrorCode == 0 {
					nodes = append(nodes, resp.Data.([]model.NodeInfo)...)
				}
			}
		}
	}

	//@todo 版本信息
	for _, node := range domain.GetNodes() {
		nodes = append(nodes, node.Info())
	}
	s.renderJSON(Resp{Data: nodes})
}

func (s *ServiceController) ConfReloadAction() {
	confManageLock.Lock()
	defer confManageLock.Unlock()

	qDomain := s.query("domain")
	qTransit := s.query("transit")

	status := map[string]bool{}
	if qTransit != "true" {
		keeperNodes, err := storage.KStorage.GetAliveKeeperNodes(true)
		if err != nil {
			logger.Logex.Error("ConfStatus", "GetAliveKeeperNodes", err.Error(), qDomain)
			s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
			return
		}
		for _, host := range keeperNodes {
			if host != setting.KeeperAdminAddr {
				resp := transitConfReload(host, qDomain)
				status[host] = false
				if resp.ErrorCode == 0 {
					data := resp.Data.(map[string]bool)
					if st, exist := data[host]; exist {
						status[host] = st
					}
				}
			}
		}
	}

	err := Kdomain.DomainConfs.Reload(qDomain)
	if err != nil {
		logger.Logex.Error("ConfReload", "DomainConfs.Reload", err.Error())
		s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
		return
	}
	domainConf, err := Kdomain.DomainConfs.GetDomain(qDomain)
	if err != nil {
		logger.Logex.Error("ConfReload", "DomainConfs.GetDomain", err.Error())
		s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
		return
	}
	//@todo version信息没有用到
	err = Kdomain.DomainBooks.Reload(qDomain, 0, domainConf)
	if err != nil {
		logger.Logex.Error("ConfReload", "DomainBooks.Reload", err.Error())
		s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
		return
	}
	status[setting.KeeperAdminAddr] = true

	s.renderJSON(Resp{Data: status})
}

func (s *ServiceController) NodeListAction() {
	qDomain := s.query("domain")
	domain, err := Kdomain.DomainBooks.GetDomain(qDomain)
	if err != nil {
		logger.Logex.Error("NodeList", "GetDomain", err.Error(), qDomain)
		s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
		return
	}
	qComponent := s.query("component")
	qTransit := s.query("transit")

	var nodes []model.Node
	if qTransit != "true" {
		hosts, err := getAdminAddrs(qDomain)
		if err != nil {
			logger.Logex.Error("NodeManage", "GetKeeperAddrs", err.Error(), qDomain)
			s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
			return
		}
		for _, host := range hosts {
			if host != setting.KeeperAdminAddr {
				resp := transitNodeList(host, qDomain, qComponent)
				if resp.ErrorCode == 0 {
					nodes = append(nodes, resp.Data.([]model.Node)...)
				}
			}
		}
	}

	for _, node := range domain.GetNodes() {
		if qComponent == "" {
			nodes = append(nodes, *node)
		} else if node.GetComponent() == qComponent {
			nodes = append(nodes, *node)
		}

		node.AddEvent(model.Event{EventType: model.EventNodeProc})
	}

	s.renderJSON(Resp{Data: nodes})
}

func (s *ServiceController) NodeInfoAction() {
	qDomain := s.query("domain")
	domain, err := Kdomain.DomainBooks.GetDomain(qDomain)
	if err != nil {
		logger.Logex.Error("NodeInfo", "GetDomain", err.Error(), qDomain)
		s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
		return
	}
	id := s.query("nodeid")
	qTransit := s.query("transit")

	if qTransit != "true" {
		host, err := getAdminAddr(qDomain, id)
		if err != nil {
			logger.Logex.Error("NodeManage", "GetKeeperAddr", err.Error(), qDomain)
			s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
			return
		}
		if host != setting.KeeperAdminAddr {
			s.renderJSON(transitNodeInfo(host, qDomain, id))
			return
		}
	}

	node, err := domain.GetNode(id)
	if err != nil {
		logger.Logex.Error("NodeInfo", "GetNode", err.Error(), qDomain, id)
		s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
		return
	}
	node.AddEvent(model.Event{EventType: model.EventNodeProc})

	s.renderJSON(Resp{Data: node})
}

func (s *ServiceController) NodeManageAction() {
	if ok := s.required("domain", "operate", "nodeid", "component"); !ok {
		return
	}
	qDomain := s.query("domain")
	qOperate := s.query("operate")
	qNodeID := s.query("nodeid")
	qComponent := s.query("component")
	qTransit := s.query("transit")

	if qTransit != "true" {
		host, err := getAdminAddr(qDomain, setting.GetAgentNodeID(qNodeID))
		if err != nil {
			logger.Logex.Error("NodeManage", "GetKeeperAddr", err.Error(), qDomain)
			s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
			return
		}
		if host != setting.KeeperAdminAddr {
			s.renderJSON(transitNodeManage(host, qDomain, qOperate, qNodeID, qComponent))
			return
		}
	}

	domain, err := Kdomain.DomainBooks.GetDomain(qDomain)
	if err != nil {
		logger.Logex.Error("NodeManage", "GetDomain", err.Error(), qDomain)
		s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
		return
	}
	agent, err := domain.GetNode(setting.GetAgentNodeID(qNodeID))
	if err != nil {
		logger.Logex.Error("NodeManage", "agent GetNode", err.Error(), qDomain)
		s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
		return
	}

	operateEvent, err := getOperateType(qOperate)
	if err != nil {
		logger.Logex.Error("NodeManageAction", "getOperateType", err.Error())
		s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
		return
	}

	evt := model.NewEvent()
	evt.EventType = operateEvent
	evt.Data = *model.NewNodeInfo(qNodeID, "", agent.KeeperAddr, qDomain, qComponent, []string{}, nil)
	if err = agent.AddEvent(evt); err != nil {
		logger.Logex.Error("NodeManageAction", "AddEvent", err.Error())
		s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
		return
	}
	s.renderJSON(Resp{})
}

func (s *ServiceController) DomainListAction() {
	domains := Kdomain.DomainBooks.GetDomainsInfo()
	s.renderJSON(Resp{Data: domains})
}

func getOperateType(operate string) (int, error) {
	switch operate {
	case "start":
		return model.EventCmdStart, nil
	case "restart":
		return model.EventCmdRestart, nil
	case "stop":
		return model.EventCmdStop, nil
	}
	return 0, fmt.Errorf("operate invalid: %s", operate)
}

// test interface
func (s *ServiceController) GetDomainFromDsAction() {
	if ok := s.required("domain"); !ok {
		return
	}
	qDomain := s.query("domain")
	data, err := storage.KStorage.GetDomain(qDomain, true)
	if err != nil {
		s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
		return
	}
	s.renderJSON(Resp{Data: data})
}

// test interface
func (s *ServiceController) GetDomainFromConfAction() {
	if ok := s.required("domain"); !ok {
		return
	}
	qDomain := s.query("domain")
	cf, err := Kdomain.DomainConfs.GetDomain(qDomain)
	if err != nil {
		s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
		return
	}

	s.renderJSON(Resp{Data: cf.FileList()})
}

// test interface
func (s *ServiceController) DeleteDomainAction() {
	if ok := s.required("domain", "note"); !ok {
		return
	}
	qDomain := s.query("domain")
	qNote := s.query("note")
	err := storage.KStorage.DelDomain(qDomain, qNote)
	if err != nil {
		s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
		return
	}

	s.renderJSON(Resp{})
}

func (s *ServiceController) DeleteFileAction() {
	if ok := s.required("domain", "file", "note"); !ok {
		return
	}
	qDomain := s.query("domain")
	qFile := s.query("file")
	qNote := s.query("note")
	err := storage.KStorage.DelFile(qDomain, qFile, qNote)
	if err != nil {
		s.renderJSON(Resp{ErrorCode: 1, Error: err.Error()})
		return
	}

	s.renderJSON(Resp{})
}

//从storage中获取某个节点的keeper admin注册地址
func getAdminAddr(qDomain, qNodeID string) (string, error) {
	rawAddr, err := storage.KStorage.GetKeeperAddr(qDomain, qNodeID, true)
	if err != nil {
		return "", err
	}
	admin, _ := setting.GetDecodedKeeperAddr(rawAddr)
	return admin, nil
}

//从storage中获取某个域的的keeper admin注册地址
func getAdminAddrs(qDomain string) ([]string, error) {
	rawAddrs, err := storage.KStorage.GetKeeperAddrs(qDomain, false)
	if err != nil {
		return nil, err
	}
	var adminAddrs []string
	for _, rawAddr := range rawAddrs {
		admin, _ := setting.GetDecodedKeeperAddr(rawAddr)
		adminAddrs = append(adminAddrs, admin)
	}
	return adminAddrs, nil
}
