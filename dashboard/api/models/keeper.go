package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sort"
	"time"
)

type BaseResponse struct {
	Code    int    `json:"error_code"`
	Message string `json:"error"`
}

type ClusterInfo struct {
	Name    string `json:"name"`
	Version int64  `json:"version"`
}

type Clusters []*ClusterInfo

func (s Clusters) Len() int           { return len(s) }
func (s Clusters) Less(i, j int) bool { return s[i].Name < s[j].Name }
func (s Clusters) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type QueryClustersResp struct {
	BaseResponse
	Data Clusters `json:"data"`
}

type KeyConfig struct {
	Type     string `json:"type"`
	RawKey   string `json:"raw_key"`
	RawValue string `json:"raw_value"`
	Key      string `json:"key"`
}

type SectionConfig struct {
	Name string                `json:"name"`
	Keys map[string]*KeyConfig `json:"keys"`
}

type FileConfig struct {
	Name     string           `json:"name"`
	Sections []*SectionConfig `json:"sections"`
}

type QueryConfigResp struct {
	BaseResponse
	Data []*FileConfig `json:"data"`
}

type NodeConfig struct {
	Name    string                `json:"name"`
	Version int64                 `json:"version"`
	Data    map[string]*KeyConfig `json:"data"`
}

type ProcessBaseInfo struct {
	Pid       string `json:"Pid"`
	ParentPid string `json:"PPid"`
	Command   string `json:"Command"`
	State     string `json:"State"`
	StartTime string `json:"StartTime"`
}

type ProcessCpuInfo struct {
	UTime     int64  `json:"Utime"`
	STime     int64  `json:"Stime"`
	Cutime    int64  `json:"Cutime"`
	Cstime    int64  `json:"Cstime"`
	StartTime int64  `json:"StartTime"`
	LastUS    int64  `json:"LastUS"`
	LastTimer string `json:"LastTimer"`
	CpuUsage  string `json:"CpuUsage"`
	Pid       string `json:"Pid"`
	PPid      string `json:"PPid"`
	Command   string `json:"Command"`
	State     string `json:"State"`
}

type ProcessMemoryInfo struct {
	VmSize    int64  `json:"VmSize"`
	VmRss     int64  `json:"VmRss"`
	VmData    int64  `json:"VmData"`
	VmStk     int64  `json:"VmStk"`
	VmExe     int64  `json:"VmExe"`
	VmLib     int64  `json:"VmLib"`
	Pid       string `json:"Pid"`
	PPid      string `json:"PPid"`
	Command   string `json:"Command"`
	State     string `json:"State"`
	StartTime string `json:"StartTime"`
}

type ProcessInfo struct {
	Base *ProcessBaseInfo   `json:"Base"`
	Cpu  *ProcessCpuInfo    `json:"Cpu"`
	Mem  *ProcessMemoryInfo `json:"Mem"`
}

type CompileInfo struct {
	Operator  string `json:"operator"`
	TimeStamp string `json:"timestamp"`
	VCS       string `json:"vcs"`
	Version   string `json:"version"`
}

type NodeInfo struct {
	Id              string        `json:"id"`
	KeeperAddr      string        `json:"keeper_addr"`
	Cluster         string        `json:"domain"`
	Component       string        `json:"component"`
	Hostname        string        `json:"hostname"`
	StartTime       int64         `json:"start_time"`
	UpdateTime      int64         `json:"update_time"`
	RawSubscription []string      `json:"raw_subscription"`
	Status          int           `json:"status"`
	Version         int64         `json:"version"`
	CompileInfo     *CompileInfo  `json:"component_tags"`
	Subscription    []string      `json:"subscription"`
	Configs         []*NodeConfig `json:"struct_datas"`
	ProcessInfo     *ProcessInfo  `json:"proc"`
}

type QueryNodeListResp struct {
	BaseResponse
	Data []*NodeInfo `json:"data"`
}

type ConfigOperate struct {
	Action  string `json:"opcode"`
	Cluster string `json:"domain"`
	File    string `json:"file"`
	Section string `json:"section"`
	Key     string `json:"key"`
	Type    string `json:"type"`
	Value   string `json:"value"`
	Comment string `json:"note"`
	ID      int    `json:"id"`
}

const (
	Start   = "start"
	Stop    = "stop"
	Restart = "restart"
)

const (
	GetConfig    = "get"
	AddConfig    = "add"
	UpdateConfig = "update"
	DeleteConfig = "delete"
)

type Client struct {
	keeper     string
	httpClient http.Client
}

var KeeperAdminClient *Client

func InitClient(keeper string) {
	KeeperAdminClient = &Client{
		keeper: keeper,
		httpClient: http.Client{
			Transport: &http.Transport{
				Dial: func(network, addr string) (net.Conn, error) {
					c, err := net.DialTimeout(network, addr, time.Second)
					if err != nil {
						return nil, err
					}
					return c, nil
				},
				MaxIdleConnsPerHost: 5,
			},
			Timeout: time.Second * 5,
		},
	}
	fmt.Println("keeper admin address:", keeper)
	return
}

func (c *Client) Keeper() string {
	return c.keeper
}

// 查询 keeper 下所有集群
func (c *Client) QueryClusters() (Clusters, error) {
	resp := &QueryClustersResp{}
	args := url.Values{}
	if err := c.request("/domain/list", args, false, resp); err != nil {
		return nil, err
	} else {
		cs := resp.Data
		sort.Sort(cs)
		return cs, nil
	}
}

func (c *Client) ManageNode(cluster, node, operation string) error {
	return nil
}

// 查询集群所有结点
func (c *Client) QueryNodeList(cluster, component string) ([]*NodeInfo, error) {
	resp := &QueryNodeListResp{}
	args := url.Values{
		"domain":    []string{cluster},
		"component": []string{component},
	}
	if err := c.request("/node/list", args, false, resp); err != nil {
		return nil, err
	} else {
		return resp.Data, nil
	}
}

// 查询集群结点信息
func (c *Client) QueryNodeInfo(cluster, node string) error {

	return nil
}

// 查询集群配置
func (c *Client) QueryConfig(cluster string) ([]*FileConfig, error) {
	resp := &QueryConfigResp{}
	args := url.Values{
		"domain": []string{cluster},
	}
	if err := c.request("/conf/list", args, false, resp); err != nil {
		return nil, err
	} else {
		return resp.Data, nil
	}
}

func (c *Client) ReloadConfig(cluster string) error {
	path := "/conf/reload"
	_ = path
	return nil
}

func (c *Client) RollbackConfig(cluster string, id int64) error {
	path := "/conf/rollback"
	_ = path
	return nil
}

// 管理集群配置
func (c *Client) ManageConfig(cluster string, ops []*ConfigOperate, note string) error {
	data, err := json.Marshal(ops)
	if err != nil {
		return err
	}
	args := url.Values{
		"domain":   []string{cluster},
		"operates": []string{string(data)},
		"note":     []string{note},
	}
	resp := &BaseResponse{}
	if err := c.request("/conf/manage", args, true, resp); err != nil {
		return err
	} else {
		return nil
	}
}

//
func (c *Client) QueryHistory() error {
	path := "/package/list"
	_ = path
	return nil
}

// node status
func (c *Client) QueryNodeStatus() error {
	path := "/conf/status"
	_ = path
	return nil
}

func (c *Client) request(path string, form url.Values, post bool, data interface{}) (err error) {
	var resp *http.Response
	if post {
		uri := fmt.Sprintf("http://%s%s", c.keeper, path)
		resp, err = c.httpClient.PostForm(uri, form)
	} else {
		uri := fmt.Sprintf("http://%s%s?%s", c.keeper, path, form.Encode())
		resp, err = c.httpClient.Get(uri)
	}
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("error status:" + resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(body, data); err != nil {
		return err
	}
	return nil
}
