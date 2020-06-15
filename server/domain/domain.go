package domain

import (
	"fmt"
	"sync"

	"github.com/huajiao-tv/gokeeper/server/setting"

	"github.com/huajiao-tv/gokeeper/model"
	"github.com/huajiao-tv/gokeeper/server/storage"
)

type Domain struct {
	Name    string `json:"name"`
	Version int    `json:"version"` //int type，兼容client老版本

	root  string
	nodes map[string]*model.Node
	sync.RWMutex
}

func NewDomain(name string) *Domain {
	return &Domain{
		Name:  name,
		nodes: map[string]*model.Node{},
	}
}

func (d *Domain) Register(node *model.Node) error {
	d.Lock()
	defer d.Unlock()
	d.nodes[node.ID] = node
	return storage.KStorage.SetKeeperAddr(node.Domain, node.ID, setting.GetEncodedKeeperAddr())
}

func (d *Domain) GetName() string {
	return d.Name
}

func (d *Domain) GetVersion() int {
	return d.Version
}

func (d *Domain) DelNode(ID string) {
	d.Lock()
	if node, exist := d.nodes[ID]; exist {
		if err := storage.KStorage.DelKeeperAddr(node.Domain, node.ID, setting.GetEncodedKeeperAddr()); err != nil {
			//@todo handle error
		}
		delete(d.nodes, ID)
	}
	d.Unlock()
}

func (d *Domain) GetNode(ID string) (*model.Node, error) {
	d.RLock()
	n, ok := d.nodes[ID]
	d.RUnlock()
	if !ok {
		return nil, fmt.Errorf("node %s not found", ID)
	}
	return n, nil
}

func (d *Domain) GetNodes() []*model.Node {
	nodes := []*model.Node{}
	d.RLock()
	for _, node := range d.nodes {
		nodes = append(nodes, node)
	}
	d.RUnlock()
	return nodes
}

func (d *Domain) SetVersion(version int) {
	d.Lock()
	d.Version = version
	d.Unlock()
}
