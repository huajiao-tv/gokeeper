package client

import (
	"errors"
	"time"
)

//
var (
	ErrEventDataInvalid = errors.New("event data invalid")
	ErrArgs             = errors.New("gokeeper client useage: ./bin/component -d=domain -n=nodeid -k=keeper_address")

	ConnectTimeout = time.Duration(10) * time.Second
	ReadTimeout    = time.Duration(300) * time.Second
	WriteTimeout   = time.Duration(300) * time.Second
	EventInterval  = time.Duration(5) * time.Second
)
