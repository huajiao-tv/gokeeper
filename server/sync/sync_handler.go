//该文件主要如下事件：
//1、keeper client端注册、轮询等产生的事件

package sync

import (
	"fmt"
	"time"

	"github.com/huajiao-tv/gokeeper/model"
	Kdomain "github.com/huajiao-tv/gokeeper/server/domain"
	"github.com/huajiao-tv/gokeeper/server/logger"
	"github.com/huajiao-tv/gokeeper/server/metrics"
	"github.com/huajiao-tv/gokeeper/server/setting"
	"github.com/huajiao-tv/gokeeper/server/storage"
	"github.com/huajiao-tv/gokeeper/server/storage/etcd"
)

func eventProxy(evtReq *model.Event) (*model.Event, *model.Node, error) {
	switch evtReq.EventType {
	case model.EventNodeRegister:
		return eventNodeRegister(evtReq)
	case model.EventNodeProc:
		return eventNodeProc(evtReq)
	case model.EventNone:
		return eventNone(evtReq)
	}

	return nil, nil, fmt.Errorf("event unsupport: eventType=%d", evtReq.EventType)
}

func eventNodeRegister(evtReq *model.Event) (*model.Event, *model.Node, error) {
	nodeReq, ok := (evtReq.Data).(model.NodeInfo)
	if !ok {
		return nil, nil, errEventDataInvalid
	}

	domain, err := Kdomain.DomainBooks.GetDomain(nodeReq.GetDomain())
	if err != nil {
		return nil, nil, err
	}
	if node, err := domain.GetNode(nodeReq.GetID()); err == nil {
		node.Exit()
	}
	node := model.NewNode(nodeReq)

	domainConf, err := Kdomain.DomainConfs.GetDomain(nodeReq.GetDomain())
	if err != nil {
		return nil, nil, err
	}
	subscription := domainConf.ParseSubscribe(nodeReq.GetRawSubscription())
	structDatas := domainConf.Subscribe(subscription)

	node.SetSubscription(subscription)
	node.SetStructDatas(structDatas, domain.GetVersion())
	node.SetUpdateTime(time.Now().Unix())
	node.SetStatus(model.StatusRunning)
	node.SetVersion(nodeReq.GetVersion())
	err = domain.Register(node)

	return &model.Event{EventType: model.EventNodeConfChanged, Data: node.GetStructDatas()}, node, err
}

func eventNodeProc(evtReq *model.Event) (*model.Event, *model.Node, error) {
	nodeReq, ok := (evtReq.Data).(model.Node)
	if !ok {
		return nil, nil, errEventDataInvalid
	}

	domain, err := Kdomain.DomainBooks.GetDomain(nodeReq.GetDomain())
	if err != nil {
		return nil, nil, err
	}
	node, err := domain.GetNode(nodeReq.GetID())
	if err != nil {
		return &model.Event{EventType: model.EventNodeRegister}, nil, nil
	}

	node.SetUpdateTime(time.Now().Unix())
	node.SetStatus(model.StatusRunning)
	node.SetProc(nodeReq.GetProc())

	return nil, node, nil
}

func eventNone(evtReq *model.Event) (*model.Event, *model.Node, error) {
	nodeReq, ok := (evtReq.Data).(model.NodeInfo)
	fmt.Println("process eventNone:", evtReq)
	if !ok {
		return nil, nil, errEventDataInvalid
	}

	domain, err := Kdomain.DomainBooks.GetDomain(nodeReq.GetDomain())
	if err != nil {
		return nil, nil, err
	}

	var node *model.Node
	var rpcAddr string

	//获取节点的keeper注册地址
	rawAddrs, err := storage.KStorage.GetKeeperAddr(nodeReq.GetDomain(), nodeReq.GetID(), false)
	if err == etcd.ErrKeyNotExist {
		return &model.Event{EventType: model.EventNodeRegister}, nil, nil
	}
	if err != nil {
		logger.Logex.Warn("eventNone", "GetKeeperAddr", nodeReq, err)
		return nil, nil, err
	}
	_, rpcAddr = setting.GetDecodedKeeperAddr(rawAddrs)
	if len(rpcAddr) == 0 {
		logger.Logex.Warn("eventNone", "GetDecodedKeeperAddr rcpAddr is empty", nodeReq, err)
		return &model.Event{EventType: model.EventNodeRegister}, nil, nil
	}
	//节点注册的地址与自身地址不一致情况
	if rpcAddr != setting.KeeperRpcAddr {
		//rpc远程调用获取node信息,如果获取不到node信息，则返回EventNodeRegister，重新注册节点
		node, err = client.getNode(rpcAddr, &nodeReq)
		if err != nil {
			logger.Logex.Warn("eventNone", "rpc.getNode", rpcAddr, nodeReq, err)
			return &model.Event{EventType: model.EventNodeRegister}, nil, nil
		}
		//返回EventNone（出现了多条tcp连接）
		return &model.Event{EventType: model.EventNone}, node, nil
	}

	node, err = domain.GetNode(nodeReq.GetID())
	if err != nil {
		return &model.Event{EventType: model.EventNodeRegister}, nil, nil
	}

	node.SetUpdateTime(time.Now().Unix())
	node.SetStatus(model.StatusRunning)

	if domain.GetVersion() != nodeReq.GetVersion() && len(node.GetStructDatas()) > 0 {
		return &model.Event{EventType: model.EventNodeConfChanged, Data: node.GetStructDatas()}, node, nil
	}

	timer := time.NewTimer(time.Duration(setting.EventInterval) * time.Second)
	for {
		select {
		case <-timer.C:
			return nil, node, nil
		case evt := <-node.Event:
			return &evt, node, nil
		}
	}
}

func AddPromNodeEvent(typ string, evt *model.Event, node *model.Node) {
	if evt == nil {
		return
	}

	evtType := "unknown"
	switch evt.EventType {
	case model.EventNodeRegister:
		evtType = "register"
	case model.EventNodeProc:
		evtType = "proc"
	case model.EventNone:
		evtType = "none"
	}

	if node == nil {
		metrics.Metrics.AddNodeEvent(typ, evtType, "unknown", "unknown", "unknown", 1)
	} else {
		metrics.Metrics.AddNodeEvent(typ, evtType, node.ID, node.Domain, node.Hostname, 1)
	}
}
