package setting

import (
	"net"
	"testing"
)

func TestGetKeeperAddr(t *testing.T) {
	listen := "10.16.59.183:7000"
	keeperAddr, err := getKeeperAddr(listen)
	if err != nil {
		t.Error(err)
	}
	if listen != keeperAddr {
		t.Fatalf("getKeeperAddr error:%v %v", listen, keeperAddr)
	}
	listen = ":7000"
	keeperAddr, err = getKeeperAddr(listen)
	if err != nil {
		t.Error(err)
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp", keeperAddr)
	if len(tcpAddr.IP) == 0 || tcpAddr.Port != 7000 {
		t.Fatalf("getKeeperAddr error:%v %v", listen, keeperAddr)
	}
}

func TestGetAgentNodeID(t *testing.T) {
	nodeID := "127.0.0.1:80"
	ret := GetAgentNodeID(nodeID)
	if ret != "127.0.0.1" {
		t.Fatalf("GetAgentNodeID error: %v %v", nodeID, ret)
	}
}
