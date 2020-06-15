package storage

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/huajiao-tv/gokeeper/server/storage/etcd"

	"github.com/huajiao-tv/gokeeper/server/storage/operate"

	lo "github.com/huajiao-tv/gokeeper/server/logger"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/integration"
)

var (
	testKey         = "test_key"
	testKey2        = "test_key2"
	testValue       = "test_value"
	testValue2      = "test_value2"
	testSection     = "test_section"
	testSectionData = map[string]string{testKey: testValue}
	testFileData    = map[string]map[string]string{testSection: testSectionData}
	testFile        = "/test_file.conf"
	testFile2       = "/test_file2.conf"
	testFileData2   = map[string]map[string]string{testSection: {testKey: testValue2, testKey2: testValue}}
	testDomain      = "test_domain"
	testDomainData  = map[string]map[string]map[string]string{testFile: testFileData}
	testStorage     Storage
	testNodeId      = "test_node_id"
	testAddr        = "keeper.admin.com"
	testKeeperId    = int64(1)
)

var endpoints []string

func TestMain(m *testing.M) {
	cfg := integration.ClusterConfig{Size: 1}
	clus := integration.NewClusterV3(nil, &cfg)
	endpoints = []string{clus.Client(0).Endpoints()[0]}
	var err error
	os.Mkdir("log", os.ModePerm)
	lo.InitLogger("./log/", "log")
	defer os.RemoveAll("./log")
	testStorage, err = etcd.NewEtcdStorage(endpoints, "", "", lo.Logex)
	if err != nil {
		panic("init test storage etcd error:" + err.Error())
	}
	defer clus.Terminate(nil)
	m.Run()
}

func TestStorageEtcd_SetDomain(t *testing.T) {
	err := testStorage.SetDomain(testDomain, testDomainData, "test set domain")
	if err != nil {
		t.Fatal("set domain error:", err)
	}
}

func TestStorageEtcd_GetDomain(t *testing.T) {
	domain, err := testStorage.GetDomain(testDomain, true)
	if err != nil {
		t.Fatal("get domain error:", err)
	}
	if len(domain) == 0 {
		t.Fatal("set or get domain failed!")
	}
}

func TestStorageEtcd_GetDomainNames(t *testing.T) {
	domainNames, err := testStorage.GetDomainNames(true)
	if err != nil {
		t.Fatal("get domain names error:", err)
	}
	if len(domainNames) != 1 || domainNames[0] != testDomain {
		t.Fatal("set domain or get domain failed!")
	}
}

//create
func TestStorageEtcd_AddFile(t *testing.T) {
	err := testStorage.AddFile(testDomain, testFile2, testFileData, "test add file")
	if err != nil {
		t.Fatal("add file error:", err)
	}
	domain, err := testStorage.GetDomain(testDomain, true)
	if err != nil {
		t.Fatal("get domain error:", err)
	}
	if _, ok := domain[testFile2]; !ok {
		t.Fatal("add file failed!", domain)
	}
}

//update
func TestStorageEtcd_AddFile2(t *testing.T) {
	err := testStorage.AddFile(testDomain, testFile2, testFileData2, "test add file")
	if err != nil {
		t.Fatal("add file error:", err)
	}
	domain, err := testStorage.GetDomain(testDomain, true)
	if err != nil {
		t.Fatal("get domain error:", err)
	}
	if d, ok := domain[testFile2]; !ok || d[testSection][testKey] != testValue2 {
		t.Fatal("add file failed!", domain)
	}
}

func TestStorageEtcd_SetKey(t *testing.T) {
	err := testStorage.SetKey(testDomain, testFile, testSection, testKey2, testValue2, "test set key")
	if err != nil {
		t.Fatal("set key error:", err)
	}
}

func TestStorageEtcd_GetKey(t *testing.T) {
	value, err := testStorage.GetKey(testDomain, testFile, testSection, testKey2, true)
	if err != nil {
		t.Fatal("get key error:", err)
	}
	if value != testValue2 {
		t.Fatal("get wrong key:", testKey2, value)
	}
}

func TestStorageEtcd_DelKey(t *testing.T) {
	err := testStorage.DelKey(testDomain, testFile, testSection, testKey2, "test del key")
	if err != nil {
		t.Fatal("del key error:", err)
	}
}

func TestStorageEtcd_GetKey2(t *testing.T) {
	_, err := testStorage.GetKey(testDomain, testFile, testSection, testKey2, true)
	if err == nil {
		t.Fatal("get wrong key or del key failed!")
	}
}

func TestStorageEtcd_DelFile(t *testing.T) {
	err := testStorage.DelFile(testDomain, testFile2, "test del file")
	if err != nil {
		t.Fatal("del file error:", err)
	}
	domain, err := testStorage.GetDomain(testDomain, true)
	if err != nil {
		t.Fatal("get domain error:", err)
	}
	if _, ok := domain[testFile2]; ok {
		t.Fatal("del file failed!", domain)
	}
}

func TestStorageEtcd_DelDomain(t *testing.T) {
	err := testStorage.DelDomain(testDomain, "test del domain")
	if err != nil {
		t.Fatal("del domain error:", err)
	}
}

func TestStorageEtcd_GetDomain2(t *testing.T) {
	domain, _ := testStorage.GetDomain(testDomain, true)
	if len(domain) != 0 {
		t.Fatal("get domain failed:", domain)
	}
}

func TestStorageEtcd_SetKeeperAddr(t *testing.T) {
	err := testStorage.SetKeeperAddr(testDomain, testNodeId, testAddr)
	if err != nil {
		t.Fatal("set keeper addr error:", err)
	}
}

func TestStorageEtcd_GetKeeperAddr(t *testing.T) {
	addr, err := testStorage.GetKeeperAddr(testDomain, testNodeId, true)
	if err != nil {
		t.Fatal("get keeper addr error:", err)
	}
	if addr != testAddr {
		t.Fatal("set or get keeper addr failed!")
	}
}

func TestStorageEtcd_GetKeeperAddrs(t *testing.T) {
	addrs, err := testStorage.GetKeeperAddrs(testDomain, true)
	if err != nil {
		t.Fatal("get keeper addrs error:", err)
	}
	if len(addrs) != 1 || addrs[0] != testAddr {
		t.Fatal("set or get keeper addrs failed!")
	}
}

func TestStorageEtcd_DelKeeperAddr(t *testing.T) {
	err := testStorage.DelKeeperAddr(testDomain, testNodeId, testAddr)
	if err != nil {
		t.Fatal("del keeper addr error:", err)
	}
}

func TestStorageEtcd_GetKeeperAddr2(t *testing.T) {
	addr, err := testStorage.GetKeeperAddr(testDomain, testNodeId, true)
	if err != nil && err != etcd.ErrKeyNotExist {
		t.Fatal("get keeper addr error:", err)
	}
	if addr != "" {
		t.Fatal("get wrong keeper addr ordelete failed!")
	}
}

func TestStorageEtcd_GetKeeperAddrs2(t *testing.T) {
	addrs, _ := testStorage.GetKeeperAddrs(testDomain, true)
	if len(addrs) != 0 {
		t.Fatal("get wrong keeper addrs or delete failed!")
	}
}

func TestStorageEtcd_KeepAlive(t *testing.T) {
	go testStorage.KeepAlive(testKeeperId, testAddr)
}

func TestStorageEtcd_GetAliveKeeperNodes(t *testing.T) {
	nodes, err := testStorage.GetAliveKeeperNodes(true)
	if err != nil {
		t.Fatal("get alive keeper nodes error:", err)
	}
	if len(nodes) != 1 || nodes[testKeeperId] != testAddr {
		t.Fatal("keep alive or get alive keeper nodes failed!")
	}
}

func TestStorageEtcd_SetCurrentVersion(t *testing.T) {
	err := testStorage.SetDomain(testDomain, testDomainData, "test set domain")
	if err != nil {
		t.Fatal("set domain error:", err)
	}
	err = testStorage.SetCurrentVersion(testDomain, 100)
	if err != nil {
		t.Fatal("set version error:", err)
	}
}

func TestStorageEtcd_GetCurrentVersion(t *testing.T) {
	version, err := testStorage.GetCurrentVersion(testDomain, true)
	if err != nil {
		t.Fatal("get current version error:", err)
	}
	if version < 0 {
		t.Fatal("get wrong version", version)
	}
}

func TestStorageEtcd_GetHistoryVersions(t *testing.T) {
	historyVersion, err := testStorage.GetHistoryVersions(testDomain, 1, 1, true)
	if err != nil {
		t.Fatal("get history version error:", err)
	}
	version, err := testStorage.GetCurrentVersion(testDomain, true)
	if err != nil {
		t.Fatal("get current version error:", err)
	}
	if h, ok := historyVersion.([]*etcd.Recode); !ok || len(h) == 0 || h[0].Version != version-1 {
		t.Fatal("get history version failed!")
	}
}

func TestStorageEtcd_Rollback(t *testing.T) {
	t.Log("rollback init:del and set")
	TestStorageEtcd_DelDomain(t)
	TestStorageEtcd_SetDomain(t)
	TestStorageEtcd_GetDomain(t)
	version, err := testStorage.GetCurrentVersion(testDomain, true)
	if err != nil {
		t.Fatal("get current version error:", err)
	}
	err = testStorage.Rollback(testDomain, version-1, "test rollback")
	if err != nil {
		t.Fatal("rollback error:", err)
	}
	t.Log("rollback verify:")
	TestStorageEtcd_GetDomain2(t)
}

func TestStorageEtcd_GetMaxVersion(t *testing.T) {
	v, err := testStorage.GetMaxVersion(testDomain, true)
	if err != nil {
		t.Fatal("get max version error:", err)
	}
	version, err := testStorage.GetCurrentVersion(testDomain, true)
	if err != nil {
		t.Fatal("get current version error:", err)
	}
	if v != version {
		t.Fatal("get wrong max version!")
	}
}

func TestStorageEtcd_Watch(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: endpoints,
	})
	if err != nil {
		t.Fatal("create etcd")
	}
	testChan := make(chan operate.Event, 10)
	go testStorage.Watch(operate.EventModeConf, testChan)
	_, err = client.Put(context.Background(), fmt.Sprintf("/gokeeper/conf/test_domain/test_file.conf/DEFAULT/test_key"), "test_change_value2")
	if err != nil {
		t.Fatal("change key error:", err)
	}
	tt := time.NewTicker(time.Second * 2)
	defer tt.Stop()
	select {
	case e := <-testChan:
		if e.Opcode != operate.OpcodeUpdateKey || e.Domain != testDomain || e.File != testFile || e.Section != testSection || e.Key != testKey || e.Data.(string) != "test_change_value2" {
			t.Fatal("receive wrong event")
		} else {
			return
		}
	case <-tt.C:
		t.Fatal("do not receive conf change event!")
	}
}
func TestStorageEtcd_Watch2(t *testing.T) {
	testChan := make(chan operate.Event, 10)
	go testStorage.Watch(operate.EventModeVersion, testChan)
	TestStorageEtcd_SetKey(t)
	dv, err := testStorage.GetCurrentVersion(testDomain, true)
	if err != nil {
		t.Fatal("get current version error:", err)
	}
	tt := time.NewTicker(time.Second * 2)
	defer tt.Stop()
	select {
	case e := <-testChan:
		if e.Opcode != operate.OpcodeUpdateDomain || e.Domain != testDomain || e.Data.(int64) != dv {
			t.Fatal("receive wrong event")
		}
	case <-tt.C:
		t.Fatal("do not receive conf change event!")
	}
}
