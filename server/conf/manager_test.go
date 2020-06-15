package conf

import (
	"fmt"
	"sync"
	"testing"

	"github.com/huajiao-tv/gokeeper/utility/go-ini/ini"

	"github.com/huajiao-tv/gokeeper/server/logger"
	"github.com/huajiao-tv/gokeeper/server/setting"
	"github.com/huajiao-tv/gokeeper/server/storage"

	"github.com/huajiao-tv/gokeeper/model"
	"github.com/huajiao-tv/gokeeper/server/storage/operate"
)

var (
	testDomainName  = "test_123"
	testFileName    = "/global.conf"
	testSectionName = "DEFAULT"
	testKey         = "listen"
	testValue       = "test_value"
	globalConf      = `
front_listen string = :8081
admin_listen string = :17800
session_rpcs []string = 127.0.0.1:8082,127.0.0.1:8084
session_fronts []string = 127.0.0.1:8081,127.0.0.1:8083

[127.0.0.1:80]
front_listen string = :8088
`
	testEvent    operate.Event
	testConfData *model.ConfData
)

func TestMain(m *testing.M) {
	err := storage.InitStorage([]string{"127.0.0.1:2379"}, "", "", setting.EventMode, logger.Logex)
	if err != nil {
		panic("init storage error:" + err.Error())
	}
	testConfData, err = model.NewConfData(testKey, testValue)
	if err != nil {
		panic("new conf data error:" + err.Error())
	}
	data, err := model.EncodeConfData(*testConfData)
	if err != nil {
		panic("encode conf data error:" + err.Error())
	}
	testEvent = operate.Event{
		Opcode:  operate.OpcodeDeleteKey,
		Domain:  testDomainName,
		File:    testFileName,
		Section: testSectionName,
		Key:     testKey,
		Data:    data,
	}
	m.Run()
}

func TestParseSubscribe(t *testing.T) {
	confManager, err := NewConfManagerFromLocal(testdata, testdataBackup)
	if err != nil {
		t.Fatal(err)
	}
	s := []string{"/node/global.conf"}
	subs := confManager.ParseSubscribe(s)
	err = inSlice(subs, "/global.conf/DEFAULT", "/node/global.conf/DEFAULT")
	if err != nil {
		t.Fatal(err)
	}

	s = []string{"/node/global.conf/test1"}
	subs = confManager.ParseSubscribe(s)

	err = inSlice(subs, "/global.conf/DEFAULT", "/node/global.conf/test1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestSubscribe(t *testing.T) {
	confManager, err := NewConfManagerFromLocal(testdata, testdataBackup)
	if err != nil {
		t.Fatal(err)
	}
	s := []string{"/node/global.conf"}
	subs := confManager.ParseSubscribe(s)
	structDatas := confManager.Subscribe(subs)

	var structData model.StructData
	flag := false
	for _, sd := range structDatas {
		if sd.Name == "Global" {
			flag = true
			structData = sd
		}
	}
	if flag == false {
		t.Fatal("error: struct global not found")
	}

	cfd, ok := structData.Data["timeout"]
	if !ok {
		t.Fatal("error: key timeout not found")
	}
	if cfd.Type != "string" || cfd.Value.(string) != "1" {
		t.Fatal("error: timeout value incorrect", cfd)
	}

}

func TestAddFile(t *testing.T) {
	err := AddFile(testDomainName, testFileName, globalConf, "test")
	if err != nil {
		t.Fatalf("AddFile error: %v %v %v", testDomainName, testFileName, err.Error())
	}
	files, err := InitFiles(testDomainName, false)
	if err != nil {
		t.Fatalf("InitFiles error: %v %v", testDomainName, err.Error())
	}

	_, err = files.GetFile(testFileName)
	if err != nil {
		t.Fatalf("files.Get error:%v %v", testFileName, err.Error())
	}
}

func TestInitFiles(t *testing.T) {
	files, err := InitFiles(testDomainName, false)
	if err != nil {
		t.Fatalf("conf InitFiles error: %v %v", testDomainName, err.Error())
	}
	sectionData := files.SectionKeyList(testFileName, "DEFAULT")
	confData, exist := sectionData["front_listen"]
	if !exist {
		t.Fatalf("section DEFAULT: front_listen does not exist")
	}
	if confData.Value != ":8081" {
		t.Fatal("section DEFAULT: front_list is not :8081")
	}
	sectionData = files.SectionKeyList(testFileName, "127.0.0.1:80")
	confData, exist = sectionData["front_listen"]
	if !exist {
		t.Fatalf("section 127.0.0.1:80: front_listen does not exist")
	}
	if confData.Value != ":8088" {
		t.Fatal("section 127.0.0.1:80: front_list is not :8081")
	}
}

func TestInitConfManager(t *testing.T) {
	cm, err := InitConfManager(testDomainName, false)
	if err != nil {
		t.Fatalf("InitConfManager error: %v %v", testDomainName, err.Error())
	}
	fileList := cm.FileList()

	found := false
	for _, file := range fileList {
		if file.Name == testFileName {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("InitConfManager error: file %v not exist", testFileName)
	}
}

func TestConfManager(t *testing.T) {
	cm, err := NewConfManager(testEvent)
	if err != nil {
		t.Fatalf("NewConfManager error:%v", err.Error())
	}

	file, err := cm.files.GetFile(testFileName)
	if err != nil {
		t.Fatalf("files.Get error:%v", err.Error())
	}
	confData, err := file.GetKeyData(testSectionName, testKey)
	if err != nil {
		t.Fatalf("GetKeyData error:%v", err.Error())
	}
	if confData.Key != testKey || confData.Value != testValue {
		t.Fatal("NewConfData error: key or value key not equal")
	}

	err = cm.Delete(testEvent)
	if err != nil {
		t.Fatal("delete error:", err)
	}
	_, err = file.GetKeyData("DEFAULT", testKey)
	if err == nil {
		t.Fatalf("GetKeyData error:%v", err.Error())
	}
}

func TestNewFile(t *testing.T) {
	cm := &ConfManager{files: &Files{
		root:     "",
		child:    []*File{},
		childMap: map[string]*File{},
		mu:       sync.RWMutex{},
	}}
	file, err := cm.NewFile(testEvent)
	if err != nil {
		t.Fatal("new file error:", err)
	}
	if len(file.Sections) != 1 || file.Sections[0].Keys[testKey].RawValue != testValue {
		t.Fatal("new file failed!")
	}

}

func TestWrapConfManager(t *testing.T) {
	files := []File{{
		Name: testFileName,
		Sections: []*Section{&Section{
			Name: testSectionName,
			Keys: map[string]model.ConfData{
				testKey: *testConfData,
			},
		}},
		path:   "",
		info:   nil,
		config: &ini.File{},
		mu:     sync.RWMutex{},
	}}
	cm, err := WrapConfManager(files)
	if err != nil {
		t.Fatal("wrap conf manage error:", err)
	}
	if len(cm.files.child) != 1 || cm.files.child[0].Name != testFileName {
		t.Fatal("wrap conf manage failed!")
	}
}

func inSlice(ss []string, args ...string) error {
	for _, v := range args {
		flag := false
		for _, s := range ss {
			if s == v {
				flag = true
			}
		}
		if flag == false {
			return fmt.Errorf("%s not found", v)
		}
	}

	return nil
}
