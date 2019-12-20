package model

import (
	"encoding/gob"
	"encoding/json"
	"fmt"

	pb "github.com/huajiao-tv/gokeeper/pb/go"
)

// EventType
const (
	EventError = iota - 2
	EventNone
	_
	EventSync
	EventNodeConfChanged
	EventNodeRegister
	EventNodeStatus
	EventNodeProc
	EventNodeExit
	EventCmdStart
	EventCmdStop
	EventCmdRestart

	EventOperate
	EventOperateBatch
	EventOperateRollback
)

func init() {
	gob.Register(Event{})
}

// Event .
type Event struct {
	EventType int
	Data      interface{}
}

// NewEvent return EventTypeNone event
func NewEvent() Event {
	return Event{EventType: EventNone}
}

//将pb格式的event转化为model格式的event
func ParseEvent(pbEvent *pb.ConfigEvent) (*Event, error) {
	if pbEvent == nil {
		return nil, fmt.Errorf("pb event pointer is nil")
	}

	var (
		modelEvent Event
		err        error

		en = func() error {
			data := Node{}
			err := json.Unmarshal([]byte(pbEvent.Data), &data)
			modelEvent.Data = data
			return err
		}
		enf = func() error {
			data := NodeInfo{}
			err := json.Unmarshal([]byte(pbEvent.Data), &data)
			modelEvent.Data = data
			return err
		}
		ed = func() error {
			var data []StructData
			err := json.Unmarshal([]byte(pbEvent.Data), &data)
			modelEvent.Data = data
			return err
		}
	)
	modelEvent.EventType = int(pbEvent.EventType)

	switch pbEvent.EventType {
	case EventNodeRegister:
		err = enf()
	case EventNodeProc:
		err = en()
	case EventNone:
		err = enf()
	case EventNodeConfChanged:
		err = ed()
	default:
		return nil, fmt.Errorf("event unsupport: eventType=%d", pbEvent.EventType)
	}

	if err != nil {
		return nil, err
	}

	return &modelEvent, nil
}

//将model格式的event转化为pb格式的event
func FormatEvent(modelEvent *Event) (*pb.ConfigEvent, error) {
	if modelEvent == nil {
		return nil, fmt.Errorf("model event pointer is nil")
	}
	dataStr, err := json.Marshal(modelEvent.Data)
	if err != nil {
		return nil, err
	}
	return &pb.ConfigEvent{
		EventType: pb.ConfigEventType(modelEvent.EventType),
		Data:      string(dataStr),
	}, nil
}
