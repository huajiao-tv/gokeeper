package models

import (
	"errors"
	"net/url"

	"github.com/huajiao-tv/gokeeper/model/discovery"
)

type GetServicesResp struct {
	Error     string   `json:"error"`
	ErrorCode int      `json:"error_code"`
	Data      []string `json:"data"`
}

type GetServiceResp struct {
	Error     string            `json:"error"`
	ErrorCode int               `json:"error_code"`
	Data      discovery.Service `json:"data"`
}

func (c *Client) GetServices() ([]string, error) {
	resp := &GetServicesResp{}
	args := url.Values{}
	if err := c.request("/discovery/list/services", args, false, resp); err != nil {
		return nil, err
	} else {
		if resp.ErrorCode != 0 {
			return nil, errors.New("list services error:" + resp.Error)
		}
		cs := resp.Data
		return cs, nil
	}
}

func (c *Client) GetService(serviceName string) (*discovery.Service, error) {
	resp := &GetServiceResp{}
	args := url.Values{"service_name": []string{serviceName}}
	if err := c.request("/discovery/get/service", args, false, resp); err != nil {
		return nil, err
	} else {
		if resp.ErrorCode != 0 {
			return nil, errors.New("list services error:" + resp.Error)
		}
		return &resp.Data, nil
	}
}
