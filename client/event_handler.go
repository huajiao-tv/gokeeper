package client

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/huajiao-tv/gokeeper/model"
)

var (
	ProcInfo *model.ProcInfo
)

func init() {
	ProcInfo = model.NewProcInfo()
	ProcInfo.Init(strconv.Itoa(os.Getpid()))
	EventCallback.RegisterCallFunc(model.EventNodeConfChanged, nodeConfigChanged)
	EventCallback.RegisterCallFunc(model.EventNodeProc, nodeProc)
	EventCallback.RegisterCallFunc(model.EventNodeRegister, nodeRegister)
}

func nodeRegister(c *Client, evt model.Event) error {
	c.node.AddEvent(model.Event{EventType: model.EventNodeRegister, Data: c.node.Info()})
	return nil
}

func nodeConfigChanged(c *Client, evt model.Event) error {
	sdata, ok := (evt.Data).([]model.StructData)
	if !ok {
		return ErrEventDataInvalid
	}
	if len(sdata) == 0 {
		return nil
	}

	rdata := c.data
	structs := map[string]interface{}{}

	for _, sd := range sdata {
		itr, err := fill(rdata, sd)
		if err != nil {
			Stderr.Write([]byte(fmt.Sprintf("%s|gokeeper|nodeConfigChanged|fill|%s \n", time.Now().String(), err.Error())))
			continue
		}
		structs[sd.Name] = itr
	}
	for k, v := range structs {
		c.objContainer.Update(k, v)
	}

	c.node.SetVersion(sdata[0].Version)
	for _, fn := range c.callback {
		fn()
	}

	return nil
}

func nodeProc(c *Client, evt model.Event) error {
	ProcInfo.Cpu.Refresh()
	ProcInfo.Mem.Refresh()
	ProcInfo.StartDate()
	ProcInfo.Cpu.CurrentUsage()

	n := c.node
	n.SetProc(ProcInfo)

	c.node.AddEvent(model.Event{EventType: model.EventNodeProc, Data: *c.node})
	return nil
}

func nodeExit(c *Client) error {
	Stdout.Write([]byte(fmt.Sprintf("%s|gokeeper|node exit|%s \n", time.Now().String(), c.node.GetID())))
	return nil
}
