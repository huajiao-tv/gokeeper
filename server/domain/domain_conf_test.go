package domain

import (
	"testing"

	"github.com/huajiao-tv/gokeeper/model"
	"github.com/huajiao-tv/gokeeper/server/storage/operate"
)

var testFileName = "test_file"

func TestInitDomainConf(t *testing.T) {
	_, err := InitDomainConf()
	if err != nil {
		t.Fatalf("InitDomainConf error: %v", err.Error())
	}

	if DomainConfs == nil {
		t.Fatalf("InitDomainConf error: Domainconfs is nil")
	}
}

func TestUpdate(t *testing.T) {
	DomainBooks = NewDomainBook()
	confData, err := model.NewConfData("front_listen string", ":8666")
	if err != nil {
		t.Fatalf("NewConfData error:%v", err.Error())
	}
	data, err := model.EncodeConfData(*confData)
	if err != nil {
		t.Fatalf("EncodeConfData error: %v", err.Error())
	}
	event := operate.Event{
		Opcode:  operate.OpcodeUpdateKey,
		Domain:  testDomainName,
		File:    testFileName,
		Section: "DEFAULT",
		Key:     "front_listen",
		Data:    data,
	}
	err = DomainConfs.UpdateKey(event)
	if err != nil {
		t.Fatalf("DomainConfs.Update error:%v", err.Error())
	}
	cm, err := DomainConfs.GetDomain(testDomainName)
	if err != nil {
		t.Fatalf("DomainConfs.GetDomain error:%v", err.Error())
	}
	file, err := cm.GetFiles().GetFile(testFileName)
	if err != nil {
		t.Fatalf("GetFiles().GetFile error: %v", err.Error())
	}
	confData2, err := file.GetKeyData("DEFAULT", "front_listen")
	if err != nil {
		t.Fatalf("GetKeyData error: %v", err.Error())
	}
	if confData2.Value != confData.Value {
		t.Fatalf("DomainConfs.Update error: not equal with written value")
	}
}
