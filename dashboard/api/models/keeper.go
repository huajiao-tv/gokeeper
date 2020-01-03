package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"

	km "github.com/huajiao-tv/gokeeper/model"
	"github.com/huajiao-tv/gokeeper/server/conf"
	kd "github.com/huajiao-tv/gokeeper/server/domain"
)

type BaseResponse struct {
	Code    int    `json:"error_code"`
	Message string `json:"error"`
}

type Clusters []*kd.Domain

func (s Clusters) Len() int           { return len(s) }
func (s Clusters) Less(i, j int) bool { return s[i].Name < s[j].Name }
func (s Clusters) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type QueryClustersResp struct {
	BaseResponse
	Data Clusters `json:"data"`
}

type QueryNodeListResp struct {
	BaseResponse
	Data []*km.NodeInfo `json:"data"`
}

type QueryConfigResp struct {
	BaseResponse
	Data []*conf.File `json:"data"`
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
func (c *Client) QueryNodeList(cluster, component string) ([]*km.NodeInfo, error) {
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
func (c *Client) QueryConfig(cluster string) ([]*conf.File, error) {
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
func (c *Client) ManageConfig(cluster string, ops []*km.Operate, note string) error {
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
		if resp.Code != 0 {
			return errors.New(resp.Message)
		}
		return nil
	}
}

func (c *Client) AddFile(cluster string, file string, confData string, note string) error {
	resp := &BaseResponse{}
	args := url.Values{
		"domain": []string{cluster},
		"file":   []string{file},
		"conf":   []string{confData},
		"note":   []string{note},
	}
	if err := c.request("/add/file", args, false, resp); err != nil {
		return err
	} else {
		if resp.Code != 0 {
			return errors.New(resp.Message)
		}
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
