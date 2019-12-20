//该文件主要如下事件：
//1、后台直接操作产生的配置更新事件
//2、监听etcd等backend storage产生的事件

package sync

import (
	"fmt"

	"github.com/huajiao-tv/gokeeper/model"
	"github.com/huajiao-tv/gokeeper/server/domain"
	"github.com/huajiao-tv/gokeeper/server/logger"
	"github.com/huajiao-tv/gokeeper/server/storage"
	"github.com/huajiao-tv/gokeeper/server/storage/operate"
)

func EventOperateProxy(evt model.Event) error {
	var ok bool
	var op model.Operate
	var ops []model.Operate

	// parser event
	switch evt.EventType {
	case model.EventOperate, model.EventOperateRollback:
		op, ok = (evt.Data).(model.Operate)
		if !ok {
			return fmt.Errorf("event data invalid: %s", fmt.Sprintf("%#v", evt))
		}
	case model.EventOperateBatch:
		ops, ok = (evt.Data).([]model.Operate)
		if !ok || len(ops) == 0 {
			return fmt.Errorf("event data invalid: %s", fmt.Sprintf("%#v", evt))
		}
		op = ops[0]
	default:
		return fmt.Errorf("unsupport event: %s", fmt.Sprintf("%#v", evt))
	}

	switch evt.EventType {
	case model.EventOperate:
		if err := Update(op); err != nil {
			return err
		}
	case model.EventOperateBatch:
		if err := Update(ops...); err != nil {
			return err
		}
	case model.EventOperateRollback:
		if err := Rollback(op); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupport event: %s", fmt.Sprintf("%#v", evt))
	}
	return nil
}

// check operates and commit data
func Update(operates ...model.Operate) error {
	var values []string
	for _, op := range operates {
		if op.Opcode == model.OpcodeDelete {
			continue
		}

		op.Format()
		rawKey := model.GetRawKey(op.Key, op.Type)
		newCfd, err := model.NewConfData(rawKey, op.Value)
		if err != nil {
			return err
		}

		//check type when the key exists
		if s, err := storage.KStorage.GetKey(op.Domain, op.File, op.Section, newCfd.Key, true); err == nil {
			originalCfd, err := model.DecodeConfData(s)
			if err == nil {
				if originalCfd.Type != newCfd.Type {
					return fmt.Errorf("key %s type conflict", newCfd.Key)
				}
			}
		}

		switch op.Opcode {
		case model.OpcodeAdd, model.OpcodeDelete, model.OpcodeUpdate:
		default:
			return fmt.Errorf("operate invalid: %s", fmt.Sprintf("%#v", op))
		}

		value, err := model.EncodeConfData(*newCfd)
		if err != nil {
			return err
		}
		values = append(values, value)
	}

	var err error
	for k, op := range operates {
		switch op.Opcode {
		case model.OpcodeAdd:
			if e := storage.KStorage.SetKey(op.Domain, op.File, op.Section, op.Key, values[k], op.Note); e != nil {
				logger.Logex.Error("Update", "storage.KStorage.SetKey", op, e.Error())
				err = e
			}
		case model.OpcodeUpdate:
			if e := storage.KStorage.SetKey(op.Domain, op.File, op.Section, op.Key, values[k], op.Note); e != nil {
				logger.Logex.Error("Update", "storage.KStorage.SetKey", op, e.Error())
				err = e
			}
		case model.OpcodeDelete:
			if e := storage.KStorage.DelKey(op.Domain, op.File, op.Section, op.Key, op.Note); e != nil {
				logger.Logex.Error("Update", "storage.KStorage.DelKey", op, e.Error())
				err = e
			}
		}
	}

	return err
}

func Rollback(operate model.Operate) error {
	return storage.KStorage.Rollback(operate.Domain, operate.Version, operate.Note)
}

func Watch() {
	for event := range storage.EventChan {
		//fmt.Println("receive event:", event)
		switch event.Opcode {
		case operate.OpcodeUpdateKey:
			if err := domain.DomainConfs.UpdateKey(event); err != nil {
				logger.Logex.Error("Watch", "domain.DomainConfs.UpdateKey", event, err.Error())
				continue
			}
		case operate.OpcodeDeleteKey:
			if err := domain.DomainConfs.DeleteKey(event); err != nil {
				logger.Logex.Error("Watch", "domain.DomainConfs.DeleteKey", event, err.Error())
				continue
			}
		case operate.OpcodeUpdateDomain:
			if err := domain.DomainConfs.UpdateDomain(event); err != nil {
				logger.Logex.Error("Watch", "domain.DomainConfs.UpdateDomain", event, err.Error())
				continue
			}
		}
		// tell client to reload config
		domainConf, err := domain.DomainConfs.GetDomain(event.Domain)
		if err != nil {
			logger.Logex.Error("Watch", "domain.DomainConfs.GetDomain", event.Domain, err.Error())
			continue
		}
		// @todo version ???,优化??? 合并
		if err = domain.DomainBooks.Reload(event.Domain, 0, domainConf); err != nil {
			logger.Logex.Error("Watch", "domain.DomainBooks.Reload", event.Domain, err.Error())
			continue
		}
	}
}
