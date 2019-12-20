// psync：
// pepper gorpc sync server
// 采用花椒内部的gorpc协议封装的sync接口

package sync

import (
	"errors"
	"fmt"

	"github.com/huajiao-tv/gokeeper/model"
	Kdomain "github.com/huajiao-tv/gokeeper/server/domain"
	"github.com/huajiao-tv/gokeeper/server/logger"
)

var (
	errEventDataInvalid = errors.New("event data invalid")
)

//为兼容先前gorpc，仍命名为Server
type Server struct{}

func (srv *Server) Sync(evtReq *model.Event, evtResp *model.Event) error {
	logger.Logex.Trace("Sync", "evtReq", fmt.Sprintf("%#v", evtReq))
	evt, node, err := eventProxy(evtReq)

	// add prometheus info
	AddPromNodeEvent("req", evtReq, node)

	if err != nil {
		logger.Logex.Trace("Sync", "eventProxy", err.Error())
		return err
	}
	if evt == nil {
		return nil
	}

	*evtResp = *evt
	logger.Logex.Trace("Sync", "evtResp", fmt.Sprintf("%#v", evtResp))
	return nil

}

func (srv *Server) GetNode(nodeInfoReq *model.NodeInfo, nodeResp *model.Node) error {
	logger.Logex.Trace("GetNodeInfo", "nodeInfoReq", fmt.Sprintf("%#v", nodeInfoReq))
	domain, err := Kdomain.DomainBooks.GetDomain(nodeInfoReq.GetDomain())
	if err != nil {
		return err
	}
	node, err := domain.GetNode(nodeInfoReq.GetID())
	if err != nil {
		logger.Logex.Warn("GetNodeInfo", "GetNode", err)
		return err
	}
	*nodeResp = *node
	logger.Logex.Trace("GetNodeInfo", "nodeResp", fmt.Sprintf("%#v", nodeResp))
	return nil
}
