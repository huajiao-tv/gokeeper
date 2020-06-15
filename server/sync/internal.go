package sync

import (
	"time"

	"github.com/huajiao-tv/gokeeper/model"
	"github.com/johntech-o/gorpc"
)

const (
	ConnectTimeout = time.Duration(10) * time.Second
	ReadTimeout    = time.Duration(30) * time.Second
	WriteTimeout   = time.Duration(30) * time.Second
)

type Client struct {
	*gorpc.Client
}

var client = &Client{
	gorpc.NewClient(gorpc.NewNetOptions(ConnectTimeout, ReadTimeout, WriteTimeout)),
}

func (c *Client) getNode(address string, nodeInfo *model.NodeInfo) (*model.Node, error) {
	var node model.Node
	if err := c.CallWithAddress(address, "Server", "GetNode", nodeInfo, &node); err != nil {
		return nil, err
	}
	return &node, nil
}
