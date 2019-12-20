package client

import (
	"errors"
	"reflect"
	"strconv"
	"sync"
)

type EventCall struct {
	sync.RWMutex
	regFunc map[int]interface{}
}

var EventCallback = newEventCall()

func newEventCall() *EventCall {
	return &EventCall{regFunc: make(map[int]interface{})}
}

func (this *EventCall) RegisterCallFunc(e int, function interface{}) {
	this.Lock()
	defer this.Unlock()
	this.regFunc[e] = function
}

func (this *EventCall) getFunc(e int) (reflect.Value, error) {
	this.RLock()
	defer this.RUnlock()

	var funcs reflect.Value
	f, ok := this.regFunc[e]
	if !ok {
		return funcs, errors.New("func error:" + strconv.Itoa(e))
	}
	funcs = reflect.ValueOf(f)
	return funcs, nil
}

func callback(e int, params ...interface{}) (result []reflect.Value, err error) {
	f, err := EventCallback.getFunc(e)
	if err != nil {
		return nil, err
	}

	if len(params) != f.Type().NumIn() {
		err = errors.New("Number of params dismatches")
		return
	}
	in := make([]reflect.Value, len(params))
	if len(params) > 0 {
		for k, param := range params {
			in[k] = reflect.ValueOf(param)
		}
	}
	result = f.Call(in)
	return
}

func eventCallback(e int, params ...interface{}) error {
	res, err := callback(e, params...)
	if err != nil {
		return err
	}

	if len(res) <= 0 {
		return nil
	}

	itr := res[0].Interface()
	err, ok := itr.(error)
	if ok && err != nil {
		return err
	}

	return nil
}
