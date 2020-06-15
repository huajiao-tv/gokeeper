package setting

import (
	"errors"
	"net"
	"strings"
)

var (
	KeeperID        int64  //keeper id
	KeeperRpcAddr   string //keeper server address for node (gorpc listen)
	KeeperAdminAddr string //keeper server address for node (admin listen)
)

func InitKeeperAddr(rpcListen, adminListen string) error {
	var err error
	KeeperRpcAddr, err = getKeeperAddr(rpcListen)
	if err != nil {
		return err
	}
	KeeperAdminAddr, err = getKeeperAddr(adminListen)
	return err
}

func getKeeperAddr(listen string) (string, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", listen)
	if err != nil {
		return "", err
	}
	if len(tcpAddr.IP) != 0 {
		return listen, nil
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	var ips []net.IP
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP)
			}
		}
	}
	if len(ips) == 0 {
		return "", errors.New("can not get listen ip")
	}
	//优先获取以10.为开头的ip
	for _, ip := range ips {
		if ip.To4()[0] == 10 {
			return (&net.TCPAddr{IP: ip, Port: tcpAddr.Port}).String(), nil
		}
	}
	return (&net.TCPAddr{IP: ips[0], Port: tcpAddr.Port}).String(), nil
}

// agent 不需要端口，取ip作为agent nodeid
// 初衷：想通过各个服务的节点 nodeID，直接获取agent的node信息，
// 而仅通过服务的nodeID是无法得知agent的端口号的，所以干脆就直接使用ip作为索引
func GetAgentNodeID(nodeID string) string {
	n := strings.Split(nodeID, ":")
	return n[0]
}

func GetEncodedKeeperAddr() string {
	return strings.Join([]string{KeeperAdminAddr, KeeperRpcAddr}, ",")
}

func GetDecodedKeeperAddr(rawAddrs string) (admin, rpc string) {
	if len(rawAddrs) == 0 {
		return "", ""
	}
	addrs := strings.Split(rawAddrs, ",")
	if len(addrs) == 1 {
		return addrs[0], ""
	}
	return addrs[0], addrs[1]
}
