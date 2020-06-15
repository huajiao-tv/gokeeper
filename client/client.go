package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/huajiao-tv/gokeeper/model"
	pb "github.com/huajiao-tv/gokeeper/pb/go"
	"github.com/johntech-o/gorpc"
	"github.com/silenceper/pool"
	"google.golang.org/grpc"
)

//
var (
	Debug            = false
	Stdout io.Writer = os.Stdout
	Stderr io.Writer = os.Stderr

	errUsage = errors.New("gokeeper usage: ./component/bin -n=nodeID -d=domain -k=keeper_address")
)

const (
	SchemeTypeRpc  = "prpc"
	SchemeTypeGrpc = "grpc"

	poolInitCap     = 2
	poolMaxCap      = 2
	poolIdleTimeout = 2 * time.Hour
)

// Client ...
type Client struct {
	node       *model.Node
	schemeType string
	//gorpc
	rpc *gorpc.Client
	//grpc
	grpc *grpc.ClientConn
	//syncStream pb.Sync_SyncClient
	pool pool.Pool

	objContainer model.ObjContainer
	data         map[string]interface{}
	callback     []func()
}

type Option func(c *Client)

func WithGrpc() Option {
	return func(c *Client) {
		c.schemeType = SchemeTypeGrpc
	}
}

// New return client struct
func New(keeperAddr, domain, nodeID, component string, rawSubscription []string, tags map[string]string, opts ...Option) *Client {
	//SchemeTypeRpc is default scheme type
	c := &Client{
		schemeType: SchemeTypeRpc,
		data:       map[string]interface{}{},
	}

	for _, opt := range opts {
		opt(c)
	}

	for k := range rawSubscription {
		rawSubscription[k] = filepath.Join("/", rawSubscription[k])
	}
	hostname, _ := os.Hostname()
	nodeInfo := model.NewNodeInfo(nodeID, hostname, keeperAddr, domain, component, rawSubscription, tags)

	c.node = model.NewNode(*nodeInfo)
	switch c.schemeType {
	case SchemeTypeRpc:
		c.initRpcClient()
	case SchemeTypeGrpc:
		c.initGrpcClient()
	}

	return c
}

// new rpc client
func (c *Client) initRpcClient() {
	c.rpc = gorpc.NewClient(gorpc.NewNetOptions(ConnectTimeout, ReadTimeout, WriteTimeout))
}

// new grpc client
func (c *Client) initGrpcClient() {
	dialOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBackoffMaxDelay(1 * time.Second),
	}

	dialCtx, dialCancel := context.WithTimeout(context.Background(), ConnectTimeout)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, c.node.KeeperAddr, dialOpts...)
	if err != nil {
		panic(fmt.Sprintf("grpc.DialContext addr:%s error:%s", c.node.KeeperAddr, err.Error()))
	}

	c.grpc = conn
	syncClient := pb.NewSyncClient(conn)
	c.pool, err = NewSyncStreamPool(syncClient, poolInitCap, poolMaxCap, poolIdleTimeout)
	if err != nil {
		panic("NewSyncStreamPool error:" + err.Error())
	}
}

// LoadData register data
func (c *Client) LoadData(objContainer model.ObjContainer) *Client {
	c.objContainer = objContainer
	s := c.objContainer.GetStructs()
	for k, v := range s {
		c.data[k] = v
	}
	return c
}

// RegisterCallback event
func (c *Client) RegisterCallback(args ...func()) *Client {
	for _, v := range args {
		c.callback = append(c.callback, v)
	}
	return c
}

func (c *Client) sync(req model.Event) (model.Event, error) {
	var resp = model.NewEvent()

	switch c.schemeType {
	case SchemeTypeRpc:
		err := c.rpc.CallWithAddress(c.node.GetKeeperAddr(), "Server", "Sync", &req, &resp)
		if err != nil {
			return resp, errors.New(err.Error())
		}
	case SchemeTypeGrpc:
		pbReq, err := model.FormatEvent(&req)
		if err != nil {
			return resp, err
		}

		var pbResp *pb.ConfigEvent
		pbResp, err = Sync(c.pool, pbReq, WithSyncCallRetryTimes(poolMaxCap))
		if err != nil {
			return resp, err
		}
		respPtr, err := model.ParseEvent(pbResp)
		if err != nil {
			return resp, err
		}
		resp = *respPtr
	}
	return resp, nil
}

// Work get data from keeper service and listen data change
func (c *Client) Work() error {
	if len(c.data) == 0 {
		Stdout.Write([]byte("gokeeper did not load any data (forgotten LoadData?) \n"))
	}
	if c.node.GetKeeperAddr() == "" || c.node.GetID() == "" || c.node.GetDomain() == "" {
		return errUsage
	}

	// 第一次必须阻塞式加载数据
	evtReq := model.Event{
		EventType: model.EventNodeRegister,
		Data:      c.node.Info(),
	}

	if Debug {
		Stdout.Write([]byte(fmt.Sprintf("%s|gokeeper|Work|event request|%#v \n", time.Now().String(), evtReq)))
	}

	evtResp, err := c.sync(evtReq)
	if err != nil {
		return err
	}

	if Debug {
		Stdout.Write([]byte(fmt.Sprintf("%s|gokeeper|Work|event response|%#v \n", time.Now().String(), evtResp)))
	}

	if err := c.eventParser(evtResp); err != nil {
		return err
	}

	go c.eventLoop()
	go SignalNotifyDeamon(c)
	return nil
}

func (c *Client) eventParser(evt model.Event) error {
	switch evt.EventType {
	case model.EventNone:
		return nil
	default:
		if err := eventCallback(evt.EventType, c, evt); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) eventLoop() {
	for {
		evtReq := model.Event{EventType: model.EventNone, Data: c.node.Info()}

		select {
		case evtReq = <-c.node.Event:
		default:
		}

		if Debug {
			Stdout.Write([]byte(fmt.Sprintf("%s|gokeeper|eventLoop|event request|%#v \n", time.Now().String(), evtReq)))
		}

		evtResp, err := c.sync(evtReq)
		if err != nil {
			Stderr.Write([]byte(fmt.Sprintf("%s|gokeeper|eventLoop|Server|Sync|%s \n", time.Now().String(), err.Error())))
			time.Sleep(EventInterval)
			continue
		}

		if Debug {
			Stdout.Write([]byte(fmt.Sprintf("%s|gokeeper|eventLoop|event response|%#v \n", time.Now().String(), evtResp)))
		}

		if err = c.eventParser(evtResp); err != nil {
			Stderr.Write([]byte(fmt.Sprintf("%s|gokeeper|eventLoop|eventParser|%s \n", time.Now().String(), err.Error())))
			continue
		}
	}
}
