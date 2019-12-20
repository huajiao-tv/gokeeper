package domain

import (
	"os"
	"sync"
	"testing"

	"github.com/huajiao-tv/gokeeper/server/storage"
	"github.com/huajiao-tv/gokeeper/server/storage/etcd"
	"go.etcd.io/etcd/integration"

	"github.com/huajiao-tv/gokeeper/server/conf"
	lo "github.com/huajiao-tv/gokeeper/server/logger"
)

var testDomainBook *DomainBook
var testDomainName = "test_domain"

func TestMain(m *testing.M) {
	cfg := integration.ClusterConfig{Size: 1}
	clus := integration.NewClusterV3(nil, &cfg)
	endpoints := []string{clus.Client(0).Endpoints()[0]}
	var err error
	os.Mkdir("log", os.ModePerm)
	lo.InitLogger("./log/", "log")
	defer os.RemoveAll("./log")
	storage.KStorage, err = etcd.NewEtcdStorage(endpoints, "", "", lo.Logex)
	if err != nil {
		panic("init test storage etcd error:" + err.Error())
	}
	defer clus.Terminate(nil)
	testDomainBook = NewDomainBook()
	if testDomainBook == nil {
		panic("new domain book failed!")
	}
	m.Run()
}

func TestDomainBook_AddDomain(t *testing.T) {
	testDomainBook.AddDomain(testDomainName)

	if domain, ok := testDomainBook.domains[testDomainName]; !ok || domain.Name != testDomainName {
		t.Fatal("add domain error")
	}
}

func TestDomainBook_GetDomain(t *testing.T) {
	domain, err := testDomainBook.GetDomain(testDomainName)
	if err != nil {
		t.Fatal("get domain error:", err)
	}
	if domain.Name != testDomainName {
		t.Fatal("get wrong domain")
	}
}

func TestDomainBook_GetDomain2(t *testing.T) {
	_, err := testDomainBook.GetDomain("test_wrong_domain")
	if err == nil {
		t.Fatal("get wrong domain!")
	}
}

func TestDomainBook_GetDomains(t *testing.T) {
	domains := testDomainBook.GetDomains()
	if len(domains) != 1 || domains[0].Name != testDomainName {
		t.Fatal("get domains failed!")
	}

}

func TestDomainBook_GetDomainsInfo(t *testing.T) {
	domainsInfo := testDomainBook.GetDomainsInfo()
	if len(domainsInfo) != 1 || domainsInfo[0].Name != testDomainName {
		t.Fatal("get domains info failed!")
	}
}

func TestDomainBook_DeleteDomain(t *testing.T) {
	testDomainBook.DeleteDomain(testDomainName)
	domain, _ := testDomainBook.GetDomain(testDomainName)
	if domain != nil {
		t.Fatal("delete domain failed!")
	}

	testDomainBook.AddDomain(testDomainName)
}

//todo 结果检测不完整
func TestDomainBook_Reload(t *testing.T) {
	err := testDomainBook.Reload(testDomainName, 100, &conf.ConfManager{
		RWMutex: sync.RWMutex{},
	})
	if err != nil {
		t.Fatal("reload error:", err)
	}
	if domain, ok := testDomainBook.domains[testDomainName]; !ok || domain.Version != 100 {
		t.Fatal("reload failed")
	}
}
