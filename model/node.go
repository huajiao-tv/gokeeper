package model

import (
	"encoding/gob"
	"errors"
	"fmt"
	"sync"
	"time"
)

func init() {
	gob.Register(NodeInfo{})
	gob.Register(Node{})
}

type NodeStatus int

const (
	PkgPrefix     string     = "github.com/huajiao-tv/gokeeper/model."
	StatusStop    NodeStatus = 0
	StatusRunning NodeStatus = 1

	eventChanSize = 10
)

type NodeInfo struct {
	ID              string            `json:"id"`
	KeeperAddr      string            `json:"keeper_addr"`
	Domain          string            `json:"domain"`
	Component       string            `json:"component"`
	Hostname        string            `json:"hostname"`
	StartTime       int64             `json:"start_time"`
	UpdateTime      int64             `json:"update_time"`
	RawSubscription []string          `json:"raw_subscription"`
	Status          NodeStatus        `json:"status"`
	Version         int               `json:"version"`
	mu              sync.RWMutex      `json:"-"`
	ComponentTags   map[string]string `json:"component_tags"`
}

func NewNodeInfo(ID, hostname, keeperAddr, domain, component string, rawSubscription []string, tag map[string]string) *NodeInfo {
	info := &NodeInfo{
		ID:              ID,
		KeeperAddr:      keeperAddr,
		Domain:          domain,
		Component:       component,
		StartTime:       time.Now().Unix(),
		Status:          StatusRunning,
		Hostname:        hostname,
		RawSubscription: rawSubscription,
		ComponentTags:   tag,
	}
	return info
}

func (n *NodeInfo) GetID() string {
	return n.ID
}

func (n *NodeInfo) GetDomain() string {
	return n.Domain
}

func (n *NodeInfo) GetRawSubscription() []string {
	n.mu.RLock()
	rawSubscription := n.RawSubscription
	n.mu.RUnlock()
	return rawSubscription
}

func (n *NodeInfo) GetVersion() int {
	n.mu.RLock()
	v := n.Version
	n.mu.RUnlock()
	return v
}

func (n *NodeInfo) GetKeeperAddr() string {
	n.mu.RLock()
	s := n.KeeperAddr
	n.mu.RUnlock()
	return s
}

func (n *NodeInfo) GetComponent() string {
	n.mu.RLock()
	s := n.Component
	n.mu.RUnlock()
	return s
}

func (n *NodeInfo) GetUpdateTime() int64 {
	n.mu.RLock()
	t := n.UpdateTime
	n.mu.RUnlock()
	return t
}

func (n *NodeInfo) SetStatus(status NodeStatus) {
	n.mu.Lock()
	n.Status = status
	n.mu.Unlock()
}

func (n *NodeInfo) SetVersion(version int) {
	n.mu.Lock()
	n.Version = version
	n.mu.Unlock()
}

func (n *NodeInfo) SetComponentTag(tag map[string]string) {
	n.mu.Lock()
	n.ComponentTags = tag
	n.mu.Unlock()
}

func (n *NodeInfo) SetStartTime(t int64) {
	n.mu.Lock()
	n.StartTime = t
	n.mu.Unlock()
}

func (n *NodeInfo) SetUpdateTime(t int64) {
	n.mu.Lock()
	n.UpdateTime = t
	n.mu.Unlock()
}

func (n *NodeInfo) SetRawSubscription(rawSubscription []string) {
	n.mu.Lock()
	n.RawSubscription = rawSubscription
	n.mu.Unlock()
}

// Node stores the information about a node
type Node struct {
	*NodeInfo
	Subscription []string     `json:"subscription"`
	StructDatas  []StructData `json:"struct_datas"`
	Proc         *ProcInfo    `json:"proc"`

	Event chan Event `json:"-"`
}

func NewNode(nodeInfo NodeInfo) *Node {
	node := &Node{
		NodeInfo: NewNodeInfo(nodeInfo.ID, nodeInfo.Hostname, nodeInfo.KeeperAddr, nodeInfo.Domain, nodeInfo.Component, nodeInfo.RawSubscription, nodeInfo.ComponentTags),
		Event:    make(chan Event, eventChanSize),
	}
	return node
}

func (n *Node) GetSubscription() []string {
	n.mu.RLock()
	s := n.Subscription
	n.mu.RUnlock()
	return s
}
func (n *Node) GetStructDatas() []StructData {
	n.mu.RLock()
	structDatas := n.StructDatas
	n.mu.RUnlock()
	return structDatas
}

func (n *Node) GetProc() *ProcInfo {
	n.mu.RLock()
	p := n.Proc
	n.mu.RUnlock()
	return p
}

func (n *Node) SetSubscription(subscription []string) {
	n.mu.Lock()
	n.Subscription = subscription
	n.mu.Unlock()
}

func (n *Node) SetStructDatas(structDatas []StructData, version int) {
	for k := range structDatas {
		structDatas[k].Version = version
	}
	n.mu.Lock()
	n.StructDatas = structDatas
	n.mu.Unlock()
}

func (n *Node) SetProc(proc *ProcInfo) {
	n.mu.Lock()
	n.Proc = proc
	n.mu.Unlock()
}

func (n *Node) AddEvent(evt Event) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.Status == StatusStop {
		return errors.New("node has stopped")
	}

	select {
	case n.Event <- evt:
	default:
		return fmt.Errorf("%s event chan full", n.ID)
	}
	return nil
}

func (n *Node) Info() NodeInfo {
	n.mu.RLock()
	info := n.NodeInfo
	n.mu.RUnlock()
	return *info
}

func (n *Node) Exit() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.Status = StatusStop
	for {
		select {
		case <-n.Event:
		default:
			return
		}
	}
}
