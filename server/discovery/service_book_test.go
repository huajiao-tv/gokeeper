package discovery

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"go.etcd.io/etcd/integration"

	"github.com/huajiao-tv/gokeeper/server/logger"

	"github.com/huajiao-tv/gokeeper/server/discovery/registry/etcd"

	dr "github.com/huajiao-tv/gokeeper/server/discovery/registry"

	dm "github.com/huajiao-tv/gokeeper/model/discovery"
)

var (
	registryEtcdUrl = []string{}
	testProperty    = "property1"
)

type TestSessionBook struct {
	PushFunc func(upgradeServices map[string]*dm.Service) error
	sync.RWMutex
}

func (s TestSessionBook) Add(session *Session) {

}
func (s TestSessionBook) Delete(session *Session) {

}
func (s TestSessionBook) GetSubscribers(serviceName string) ([]string, error) {
	return nil, nil

}
func (s TestSessionBook) Push(upgradeServices map[string]*dm.Service) error {
	return s.PushFunc(upgradeServices)
}

var reg = etcd.NewEtcdRegistry()

func TestMain(m *testing.M) {
	cfg := integration.ClusterConfig{Size: 1}
	clus := integration.NewClusterV3(nil, &cfg)
	registryEtcdUrl = []string{clus.Client(0).Endpoints()[0]}
	err := reg.Init(dr.WithAddrs(registryEtcdUrl...))
	if err != nil {
		panic("init registry error:" + err.Error())
	}
	os.Mkdir("log", os.ModePerm)
	logger.InitLogger("./log/", "log")
	defer os.RemoveAll("./log")
	m.Run()
	clus.Terminate(nil)
}

func TestWithRegistry(t *testing.T) {
	o := &option{}
	f := WithRegistry(reg)
	f(o)
	if o.registry == nil {
		t.Fatal("with registry failed")
	}
}

func TestInitServiceBook(t *testing.T) {
	discoverySessionBook = &TestSessionBook{PushFunc: func(upgradeServices map[string]*dm.Service) error {
		fmt.Println("push successfully!")
		return nil
	}}
	err := InitServiceBook(WithRegistry(reg))
	if err != nil {
		t.Fatal("init service book error:", err)
	}
	if len(registryServiceBook.services) == 0 {
		t.Fatal("init service book doer not load services")
	}
	if len(registryServiceBook.services) > 1 {
		t.Log("have other services in etcd")
	}
	if registryServiceBook.eventChan == nil {
		t.Fatal("init service book does not init eventChan")
	}
}

func initTestServiceBook(t *testing.T, withTestInstance bool) {
	registryServiceBook = &ServiceBook{
		registry: reg,
		services: map[string]*dm.Service{},
	}

	eventChan, err := reg.Watch()
	if err != nil {
		t.Fatal("watch error:", err)
	}

	err = reg.Register(testInstance)
	if err != nil {
		t.Fatal("register test service error")
	}
	services, err := reg.ListServices()
	if err != nil {
		t.Fatal("list service error:", err)
	}

	registryServiceBook.Lock()

	for _, service := range services {
		if !withTestInstance {
			service.Instances = map[string]map[string]*dm.Instance{}
		} else {
			service.Instances = map[string]map[string]*dm.Instance{testZone: {testInstanceId: testInstance}}
		}
		registryServiceBook.services[service.ServiceName] = service
	}
	registryServiceBook.Unlock()

	registryServiceBook.eventChan = eventChan
}

func TestCreateInstance(t *testing.T) {
	initTestServiceBook(t, false)
	if len(registryServiceBook.services[testServiceName].Instances) != 0 {
		t.Fatal("init test failed:services is not empty")
	}
	err := registryServiceBook.createInstance(testInstance, 0)
	if err != nil {
		t.Fatal("create instance error:", err)
	}
	if _, ok := registryServiceBook.services[testServiceName]; !ok {
		t.Fatal("create instance failed")
	}
}
func TestModifyInstance(t *testing.T) {
	initTestServiceBook(t, true)
	if s, ok := registryServiceBook.services[testServiceName].Instances[testZone][testInstanceId]; !ok {
		t.Fatal("init test service book with instance error:instance is empty")
	} else {
		testInstance2 := *testInstance
		testInstance2.Metadata["property2"] = "202"
		err := registryServiceBook.modifyInstance(&testInstance2, 1)
		if err != nil {
			t.Fatal("modify instance error:", err)
		}
		if v, ok := s.Metadata["property2"]; (!ok) || v != "202" {
			t.Fatal("test property modify failed")
		}
		testInstance2.Id = "test_wrong_instance"
		err = registryServiceBook.modifyInstance(&testInstance2, 1)
		if err == nil {
			t.Fatal("modify wrong instance")
		}
	}

}

func TestFindInstance(t *testing.T) {
	initTestServiceBook(t, true)
	instance, err := registryServiceBook.FindInstance(testServiceName, testZone, testInstanceId)
	if err != nil {
		t.Fatal("find instance error:", err)
	}
	if instance == nil {
		t.Fatal("find no instance!")
	}
	in2, err := registryServiceBook.FindInstance(testServiceName, "", testInstanceId)
	if err != nil {
		t.Fatal("find instance error:", err)
	}
	if in2 == nil {
		t.Fatal("find no instance!")
	}
	in3, err := registryServiceBook.FindInstance(testServiceName, "test_wrong_zone", testInstanceId)
	if err == nil || in3 != nil {
		t.Fatal("find wrong instance!")
	}
	in4, err := registryServiceBook.FindInstance("test_wrong_service", testZone, testInstanceId)
	if err == nil || in4 != nil {
		t.Fatal("find wrong instance!")
	}
	in5, err := registryServiceBook.FindInstance(testServiceName, testZone, "test_wrong_id")
	if err == nil || in5 != nil {
		t.Fatal("find wrong instance!")
	}
}

func TestDeleteInstance(t *testing.T) {
	initTestServiceBook(t, true)
	err := registryServiceBook.deleteInstance(testInstance, 10)
	if err != nil {
		t.Fatal("delete instance error:", err)
	}
	if _, ok := registryServiceBook.services[testServiceName].Instances[testZone][testInstanceId]; ok {
		t.Fatal("delete instance failed")
	}

	testWrongInstance := *testInstance
	testWrongInstance.Id = "test_wrong_id"
	err = registryServiceBook.deleteInstance(&testWrongInstance, 10)
	if err == nil {
		t.Fatal("delete wrong instance")
	}
}

func TestUpdateProperty(t *testing.T) {
	initTestServiceBook(t, true)
	testBackendProperty := dm.BackendMetadataPrefix + testProperty
	err := registryServiceBook.updateProperty(&dm.Property{
		ServiceName: testServiceName,
		ZoneWeights: dm.ZoneWeights{testZone: dm.ZoneWeight{
			Src: testZone,
			Dst: map[string]uint64{testZone: 100},
		}},
		Attrs: map[string]dm.MD{testInstanceId: dm.MD{testBackendProperty: "208"}},
	}, 10)
	if err != nil {
		t.Fatal("update property error:", err)
	}
	if registryServiceBook.services[testServiceName].Instances[testZone][testInstanceId].Metadata[testBackendProperty] != "208" {
		t.Fatal("update property failed!")
	}
	if registryServiceBook.services[testServiceName].Metadata[dm.BackendMetadataZoneWeight] == "" {
		t.Fatal("update zone weight failed!")
	}
}

func TestUpdateMetadata(t *testing.T) {
	in := testInstance
	attr := dm.MD{testProperty: "999"}
	updateMetadata(in, attr, false)
	if in.Metadata[testProperty] != "999" {
		t.Fatal("update metadata failed!")
	}
}

func TestGetUpgradedServices(t *testing.T) {
	initTestServiceBook(t, true)
	service := registryServiceBook.services[testServiceName]
	version := service.Version
	s := registryServiceBook.GetUpgradedServices(map[string]int64{testServiceName: 0}, false)
	if len(s) == 0 {
		t.Fatal("get upgraded service failed!")
	}
	s1 := registryServiceBook.GetUpgradedServices(map[string]int64{testServiceName: version}, false)
	if len(s1) != 0 {
		t.Fatal("get upgraded service failed!")
	}
}

func TestWatch(t *testing.T) {
	initTestServiceBook(t, false)
	resChan := make(chan int, 10)
	testSessionBook := &TestSessionBook{
		PushFunc: func(upgradeServices map[string]*dm.Service) error {
			resChan <- 1
			return nil
		}}

	go registryServiceBook.Watch(testSessionBook)

	monitorChan := func(t *testing.T, ch <-chan int) {
		ti := time.NewTicker(time.Second * 2)
		select {
		case <-ch:
			break
		case <-ti.C:
			t.Fatal("watch does not push update!")
		}
	}

	eventChan := make(chan *dr.WatchEvent, 10)
	registryServiceBook.eventChan = eventChan
	eventChan <- &dr.WatchEvent{
		EventType: dr.WatchEventTypeCreate,
		InfoType:  dr.WatchInfoTypeInstance,
		Data:      testInstance,
		Version:   10,
	}
	monitorChan(t, resChan)
	if _, ok := registryServiceBook.services[testServiceName].Instances[testZone][testInstanceId]; !ok {
		t.Fatal("watch create instance failed!")
	}

	testInstance2 := *testInstance
	testInstance2.Metadata[testProperty] = "502"
	eventChan <- &dr.WatchEvent{
		EventType: dr.WatchEventTypeModify,
		InfoType:  dr.WatchInfoTypeInstance,
		Data:      &testInstance2,
		Version:   10,
	}
	monitorChan(t, resChan)
	if i, ok := registryServiceBook.services[testServiceName].Instances[testZone][testInstanceId]; !ok {
		t.Fatal("have no instance!")
	} else {
		if i.Metadata[testProperty] != "502" {
			t.Fatal("watch modify failed")
		}
	}

	eventChan <- &dr.WatchEvent{
		EventType: dr.WatchEventTypeDelete,
		InfoType:  dr.WatchInfoTypeInstance,
		Data:      testInstance,
		Version:   10,
	}
	monitorChan(t, resChan)
	if _, ok := registryServiceBook.services[testServiceName].Instances[testZone][testInstanceId]; ok {
		t.Fatal("watch delete instance failed!")
	}

	eventChan <- &dr.WatchEvent{
		EventType: "",
		InfoType:  dr.WatchInfoTypeProperty,
		Data: &dm.Property{
			ServiceName: testServiceName,
			ZoneWeights: dm.ZoneWeights{testZone: dm.ZoneWeight{
				Src: testZone,
				Dst: map[string]uint64{testZone: 100},
			}},
			Attrs: nil,
		},
		Version: 10,
	}
	monitorChan(t, resChan)
	if _, ok := registryServiceBook.services[testServiceName].Metadata[dm.BackendMetadataZoneWeight]; !ok {
		t.Fatal("watch update property failed!")
	}
}
