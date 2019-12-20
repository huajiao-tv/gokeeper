package client

import (
	"os"
	"os/signal"
	"reflect"
	"syscall"
)

//
var (
	SignalHandlerCallback = newSignalHandler()
)

// SignalHandler ...
type SignalHandler struct {
	call []*callFunc
}

type callFunc struct {
	function reflect.Value
	params   []reflect.Value
}

func newSignalHandler() *SignalHandler {
	return &SignalHandler{}
}

func newCallFunc() *callFunc {
	return &callFunc{}
}

func (sign *SignalHandler) registerCallFunc(function interface{}, params ...interface{}) {
	call := newCallFunc()
	call.function = reflect.ValueOf(function)
	if len(params) != call.function.Type().NumIn() {
		return
	}
	if len(params) > 0 {
		for _, value := range params {
			call.params = append(call.params, reflect.ValueOf(value))
		}
	}
	sign.call = append(sign.call, call)
}

func (sign *SignalHandler) callBack() {
	if len(sign.call) > 0 {
		for _, f := range sign.call {
			f.function.Call(f.params)
		}
	}
}

// SignalNotifyDeamon listen signal
// @todo registerCallFunc registe Client or Client implement callback interface
func SignalNotifyDeamon(c *Client) {
	SignalHandlerCallback.registerCallFunc(nodeExit, c)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGUSR1)
	select {
	case <-sigChan:
		SignalHandlerCallback.callBack()
		os.Exit(0)
	}
}
