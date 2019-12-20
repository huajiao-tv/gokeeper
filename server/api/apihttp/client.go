package apihttp

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/huajiao-tv/gokeeper/model"
	"github.com/tidwall/gjson"
)

const (
	MethodGet            = "GET"
	MethodPost           = "POST"
	DefaultClientTimeout = 5 * time.Second
	DefaultPartner       = "server"
)

type ClientParams struct {
	Url       string
	Host      string
	Method    string
	ReqParams map[string]string
	Timeout   time.Duration
}

type Client struct {
	Params ClientParams
}

//调用外部接口
func Call(host, url, method string, params map[string]string) (map[string]gjson.Result, error) {
	client := &Client{}
	client.Params.Host = "http://" + host
	client.Params.Url = url
	client.Params.Method = method
	client.Params.Timeout = DefaultClientTimeout

	//guid校验
	guidParams := &GuidParams{
		Partner: DefaultPartner,
		Rand:    strconv.FormatInt(rand.Int63(), 10),
		Time:    strconv.FormatInt(time.Now().Unix(), 10),
	}
	params["guid"] = GetServerGUID(guidParams)
	params["partner"] = guidParams.Partner
	params["rand"] = guidParams.Rand
	params["time"] = guidParams.Time

	client.Params.ReqParams = params
	return client.Do()
}

func (c *Client) Do() (map[string]gjson.Result, error) {
	Url := c.Params.Host + c.Params.Url
	var err error
	var req *http.Request
	if c.Params.Method == MethodPost {
		postValue := url.Values{}
		for key, value := range c.Params.ReqParams {
			postValue.Set(key, value)
		}
		ctx, cancel := context.WithCancel(context.TODO())
		time.AfterFunc(c.Params.Timeout, func() {
			cancel()
		})
		req, err = http.NewRequest("POST", Url, strings.NewReader(postValue.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(postValue.Encode())))
		req = req.WithContext(ctx)
	} else {
		getValues := url.Values{}
		for key, value := range c.Params.ReqParams {
			getValues.Set(key, value)
		}
		Url += "?" + getValues.Encode()
		ctx, cancel := context.WithCancel(context.TODO())
		time.AfterFunc(c.Params.Timeout, func() {
			cancel()
		})
		req, err = http.NewRequest("GET", Url, nil)
		req = req.WithContext(ctx)
	}

	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyByt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var resMap gjson.Result
	resMap = gjson.ParseBytes(bodyByt)

	return resMap.Map(), nil
}

func transitConfStatus(host, domain string) Resp {
	params := map[string]string{
		"domain":  domain,
		"transit": "true",
	}
	respRaw, err := Call(host, "/conf/status", MethodGet, params)
	if err != nil {
		return Resp{ErrorCode: 1, Error: err.Error()}
	}
	return parseResp(respRaw, []model.NodeInfo{})
}

func transitConfReload(host, domain string) Resp {
	params := map[string]string{
		"domain":  domain,
		"transit": "true",
	}
	respRaw, err := Call(host, "/conf/reload", MethodPost, params)
	if err != nil {
		return Resp{ErrorCode: 1, Error: err.Error()}
	}
	return parseResp(respRaw, map[string]bool{})
}

func transitNodeList(host, domain, component string) Resp {
	params := map[string]string{
		"domain":    domain,
		"component": component,
		"transit":   "true",
	}
	respRaw, err := Call(host, "/node/list", MethodGet, params)
	if err != nil {
		return Resp{ErrorCode: 1, Error: err.Error()}
	}
	return parseResp(respRaw, []model.Node{})
}

func transitNodeInfo(host, domain, nodeID string) Resp {
	params := map[string]string{
		"domain":  domain,
		"nodeid":  nodeID,
		"transit": "true",
	}
	respRaw, err := Call(host, "/node/info", MethodGet, params)
	if err != nil {
		return Resp{ErrorCode: 1, Error: err.Error()}
	}
	return parseResp(respRaw, model.Node{})
}

func transitNodeManage(host, domain, operate, nodeid, component string) Resp {
	params := map[string]string{
		"domain":    domain,
		"operate":   operate,
		"nodeid":    nodeid,
		"component": component,
		"transit":   "true",
	}
	respRaw, err := Call(host, "/node/manage", MethodPost, params)
	if err != nil {
		return Resp{ErrorCode: 1, Error: err.Error()}
	}
	resp := Resp{}
	resp.ErrorCode = int(respRaw["error_code"].Int())
	resp.Error = respRaw["error"].String()
	return resp
}

func parseResp(respRaw map[string]gjson.Result, typ interface{}) Resp {
	resp := Resp{}
	resp.ErrorCode = int(respRaw["error_code"].Int())
	resp.Error = respRaw["error"].String()
	if resp.ErrorCode != 0 {
		return resp
	}

	data := reflect.New(reflect.TypeOf(typ))
	err := json.Unmarshal([]byte(respRaw["data"].String()), data.Interface())
	if err != nil {
		resp.ErrorCode = 1
		resp.Error = err.Error()
		return resp
	}
	resp.Data = data.Elem().Interface()

	return resp
}
