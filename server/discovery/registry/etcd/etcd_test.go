package etcd

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/huajiao-tv/gokeeper/model/discovery"
	"github.com/huajiao-tv/gokeeper/server/discovery/registry"
)

//go test . -test.v -test.run=Test*

var instance = &discovery.Instance{
	Id:          "test-id",
	ServiceName: "test.service",
	Zone:        "bjcc",
}

var property = &discovery.Property{
	ServiceName: instance.ServiceName,
	ZoneWeights: map[string]discovery.ZoneWeight{
		"bjcc": discovery.ZoneWeight{
			Src: "bjcc",
			Dst: map[string]uint64{
				"bjyt": 10,
				"bjcc": 20,
			},
		},
	},
	Attrs: map[string]discovery.MD{
		"test-id": discovery.MD{
			discovery.BackendMetadataInstanceWeight: "5",
		},
	},
}

var r = getRegistry()

func getRegistry() *etcdRegistry {
	r := NewEtcdRegistry()
	err := r.Init(registry.WithTimeout(5 * time.Second))
	if err != nil {
		panic(err)
	}
	return r
}

func register(t *testing.T) {
	err := r.Register(instance, registry.WithRegistryTTL(30*time.Second))
	if err != nil {
		t.Fatal("registry Register error:", err)
	}
}

func setProperty(t *testing.T) {
	err := r.SetProperty(property)
	if err != nil {
		t.Log("SetProperty error:", err)
	}
}

func deleteProperty(t *testing.T) {
	err := r.SetProperty(&discovery.Property{})
	if err != nil {
		t.Log("SetProperty error:", err)
	}
}

func deregister(t *testing.T) {
	err := r.Deregister(instance)
	if err != nil {
		t.Fatal("registry Deregister error:", err)
	}
	//调用getservice时会更新leaseid，为避免对后续测试产生影响，需要清除
	if _, ok := r.leaseIds[instance.Id]; ok {
		delete(r.leaseIds, instance.Id)
	}
}

func TestRegister(t *testing.T) {
	register(t)
}

func TestDeregister(t *testing.T) {
	deregister(t)
}

func TestSetProperty(t *testing.T) {
	setProperty(t)
	deleteProperty(t)
}

func TestGetService(t *testing.T) {
	register(t)
	setProperty(t)
	service, err := r.GetService(instance.ServiceName)
	if err != nil {
		t.Fatal("registry GetService error:", err)
	}

	if len(service.Instances[instance.Zone]) == 0 {
		t.Fatal("GetService get zero instance")
	} else {
		t.Log("service:", service)
	}
	if s, ok := service.Metadata[discovery.BackendMetadataZoneWeight]; !ok {
		t.Fatal("GetService get zone weight error")
	} else {
		zw, err := discovery.DecodeZoneWeight(s)
		if err != nil {
			t.Fatal("GetService DecodeZoneWeight error:", err)
		}
		t.Log("zone weight:", zw)
	}

	sin := service.Instances[instance.Zone][instance.Id]
	if s, ok := sin.Metadata[discovery.BackendMetadataInstanceWeight]; !ok {
		t.Fatal("GetService get instance weight error")
	} else {
		if property.Attrs[instance.Id][discovery.BackendMetadataInstanceWeight] != s {
			t.Fatal("GetService saved instance weight is not equal set")
		}
	}

	deregister(t)
	deleteProperty(t)
}

func TestListServices(t *testing.T) {
	register(t)
	services, err := r.ListServices()
	if err != nil {
		t.Fatal("ListServices error:", err)
	}
	if len(services) == 0 {
		t.Fatal("ListServices get zero service")
	}
	found := false
	for _, service := range services {
		if service.ServiceName == instance.ServiceName {
			found = true
			if len(service.Instances[instance.Zone]) == 0 {
				t.Fatal("ListServices test service get zero instance")
			}
		}
	}
	if !found {
		t.Fatal("ListServices not get test service")
	}
	deregister(t)
}

func TestWatch(t *testing.T) {
	watchEvent, err := r.Watch()
	if err != nil {
		t.Fatal("watch error:", err)
	}
	version := int64(0)
	register(t)
	if event := <-watchEvent; event.InfoType == registry.WatchInfoTypeInstance && event.EventType == registry.WatchEventTypeCreate {
		if event.Version < version {
			t.Fatal("version doesn't increase!:", version, "->", event.Version)
		} else {
			version = event.Version
		}
	} else {
		t.Fatal("register event error:", event)
	}
	setProperty(t)
	//由于删除后台属性通过置空来实现，所以首次之后即为modify
	if event := <-watchEvent; event.InfoType == registry.WatchInfoTypeProperty && (event.EventType == registry.WatchEventTypeCreate || event.EventType == registry.WatchEventTypeModify) {
		if event.Version < version {
			t.Fatal("version doesn't increase!:", version, "->", event.Version)
		} else {
			version = event.Version
		}
	} else {
		t.Fatal("set property event error:", event)
	}
	deleteProperty(t)
	if event := <-watchEvent; event.InfoType == registry.WatchInfoTypeProperty && event.EventType == registry.WatchEventTypeModify {
		if event.Version < version {
			t.Fatal("version doesn't increase!:", version, "->", event.Version)
		} else {
			version = event.Version
		}
	} else {
		t.Fatal("update property event error:", event)
	}
	deregister(t)
	if event := <-watchEvent; event.InfoType == registry.WatchInfoTypeInstance && event.EventType == registry.WatchEventTypeDelete {
		if event.Version < version {
			t.Fatal("version doesn't increase!:", version, "->", event.Version)
		} else {
			version = event.Version
		}
	} else {
		t.Fatal("deregister event error:", event)
	}
	return
}

func TestEncodeProperty(t *testing.T) {
	p := &discovery.Property{
		ServiceName: "demo.test.com",
		ZoneWeights: map[string]discovery.ZoneWeight{},
		Attrs: map[string]discovery.MD{
			"test_id_1": discovery.MD{
				discovery.BackendMetadataInstanceWeight: "20",
			},
		},
	}
	b, _ := json.Marshal(p)
	t.Log("test_id_1 property:", string(b))
}
