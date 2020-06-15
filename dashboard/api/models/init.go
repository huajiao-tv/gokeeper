package models

import (
	"net"
	"net/http"
	"time"
)

func Init(keeperAddr string) error {
	InitClient(keeperAddr)
	return nil
}

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
	return
}
