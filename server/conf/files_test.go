package conf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/huajiao-tv/gokeeper/model"
	"github.com/huajiao-tv/gokeeper/server/storage/operate"
)

var (
	testdata       = "../../testdata/config/app1"
	testdataBackup = "backup/app1"
	testfile       = "/node/global.conf"
	testSection    = "/node/global.conf/fortest"
)

func init() {
	testdata, _ = filepath.Abs(testdata)
	testdataBackup = filepath.Join(filepath.Dir(filepath.Dir(testdata)), testdataBackup)
	os.MkdirAll(testdataBackup, 0755)
}

func TestOpenPath(t *testing.T) {
	files, err := OpenPath(testdata)
	if err != nil {
		t.Fatal(err)
	}
	_ = files
}

func TestWalk(t *testing.T) {
	files, err := OpenPath(testdata)
	if err != nil {
		t.Fatal(err)
	}
	_ = files

	err = files.Walk(func(file *File) error {
		//fmt.Println(file.path, file.info.Name())
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

}

func TestGet(t *testing.T) {
	files, err := OpenPath(testdata)
	if err != nil {
		t.Fatal(err)
	}
	_ = files

	file, err := files.GetFile(testfile)
	if err != nil {
		t.Fatal("Get error")
	}
	_ = file
}

func TestGetSectionParentSubscribe(t *testing.T) {
	files, err := OpenPath(testdata)
	if err != nil {
		t.Fatal(err)
	}
	expect := []string{"/global.conf", "/node/global.conf"}
	fs := files.GetExistPaths(Subscription(testSection).InvolvedFilesPath())
	if err := inSlice(fs, expect...); err != nil {
		t.Error(err)
	}
}

func TestIgnore(t *testing.T) {
	f := "abc.conf"
	expect := false
	actual := Ignore(f, false)
	if expect != actual {
		t.Error(expect, actual, f)
	}
	f = ".abc.conf"
	expect = true
	actual = Ignore(f, false)
	if expect != actual {
		t.Error(expect, actual, f)
	}

	f = "abc"
	expect = true
	actual = Ignore(f, false)
	if expect != actual {
		t.Error(expect, actual, f)
	}

	f = "abc"
	expect = false
	actual = Ignore(f, true)
	if expect != actual {
		t.Error(expect, actual, f)
	}

	f = ".abc"
	expect = true
	actual = Ignore(f, false)
	if expect != actual {
		t.Error(expect, actual, f)
	}

	f = ".abc"
	expect = true
	actual = Ignore(f, true)
	if expect != actual {
		t.Error(expect, actual, f)
	}
}

func TestGetSection(t *testing.T) {
	files, err := OpenPath(testdata)
	if err != nil {
		t.Fatal(err)
	}
	file, err := files.GetFile(testfile)
	if err != nil {
		t.Fatal("Get error")
	}
	seciton, err := file.GetSection("test1")
	if err != nil {
		t.Fatal(err)
	}
	if seciton.Name != "test1" {
		t.Fatal("error section value", seciton.Name)
	}
}

func TestGetKeyList(t *testing.T) {
	files, err := OpenPath(testdata)
	if err != nil {
		t.Fatal(err)
	}
	file, err := files.GetFile(testfile)
	if err != nil {
		t.Fatal("Get error")
	}
	flag := false
	for _, cfd := range file.KeyList() {
		if cfd.Key == "timeout" && cfd.RawValue == "1" {
			flag = true
		}
	}
	if flag == false {
		t.Fatal("test KeyList error")
	}
}

func TestSet(t *testing.T) {
	files, err := OpenPath(testdata)
	if err != nil {
		t.Fatal(err)
	}
	section := "fortest"
	confData, err := model.NewConfData("name string", "zhangsan")
	if err != nil {
		t.Fatal(err)
	}
	value, err := model.EncodeConfData(*confData)
	if err != nil {
		t.Fatal(err)
	}
	err = files.SetKey(testfile, section, confData.Key, value)
	if err != nil {
		t.Fatal(err)
	}

	file, err := files.GetFile(testfile)
	if err != nil {
		t.Fatal("Get error")
	}

	flag := false
	for _, cfd := range file.KeyList() {
		if cfd.Key == confData.Key && cfd.RawValue == confData.RawValue {
			flag = true
		}
	}
	if flag == false {
		t.Fatal("test Set error")
	}

}

func TestSectionKeyList(t *testing.T) {
	files, err := OpenPath(testdata)
	if err != nil {
		t.Fatal(err)
	}

	flag := false
	keys := files.SectionKeyList(testfile, "test1")
	for _, cfd := range keys {
		if cfd.Key == "test" && cfd.RawValue == "1" {
			flag = true
		}
	}

	if flag == false {
		t.Fatal("test SectionKeyList error")
	}
}

func TestFileList(t *testing.T) {
	files, err := OpenPath(testdata)
	if err != nil {
		t.Fatal(err)
	}

	flag := false
	for _, file := range files.FileList() {
		if file.Name == "/node/test.conf" {
			flag = true
		}
	}

	if flag == false {
		t.Fatal("test FileList error")
	}
}

func TestNewFiles(t *testing.T) {
	confData, err := model.NewConfData("listen string", ":8080")
	if err != nil {
		t.Fatalf("NewConfData error:%v", err.Error())
	}
	data, err := model.EncodeConfData(*confData)
	if err != nil {
		t.Fatalf("EncodeConfData error: %v", err.Error())
	}
	event := operate.Event{
		Opcode:  operate.OpcodeUpdateKey,
		Domain:  "test",
		File:    "/global.conf",
		Section: "DEFAULT",
		Key:     "listen",
		Data:    data,
	}
	files, err := NewFiles(event)
	if err != nil {
		t.Fatalf("NewFiles error:%v", err.Error())
	}
	file, err := files.GetFile("/global.conf")
	if err != nil {
		t.Fatalf("files.Get error:%v", err.Error())
	}
	confData2, err := file.GetKeyData("DEFAULT", "listen")
	if err != nil {
		t.Fatalf("GetKeyData error:%v", err.Error())
	}
	if confData.Key != confData2.Key {
		t.Fatalf("NewConfData error: %v %v key not equal", confData.Key, confData2.Key)
	}
}
