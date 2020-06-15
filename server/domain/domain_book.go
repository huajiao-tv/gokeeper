package domain

import (
	"fmt"
	"sync"
	"time"

	"github.com/huajiao-tv/gokeeper/server/metrics"

	"github.com/huajiao-tv/gokeeper/model"
	"github.com/huajiao-tv/gokeeper/server/conf"
)

type DomainBook struct {
	domains map[string]*Domain
	sync.RWMutex
}

var (
	DomainBooks *DomainBook
)

func InitDomainBook(domains []string, eventInterval int) error {
	DomainBooks = NewDomainBook()
	for _, domain := range domains {
		DomainBooks.AddDomain(domain)
	}
	go DomainBooks.Monitor(time.Duration(eventInterval)*time.Second, int64(2*eventInterval))
	return nil
}

func NewDomainBook() *DomainBook {
	return &DomainBook{domains: map[string]*Domain{}}
}

func (d *DomainBook) AddDomain(domainName string) {
	domain := NewDomain(domainName)
	d.Lock()
	_, ok := d.domains[domainName]
	if !ok {
		d.domains[domainName] = domain
	}
	d.Unlock()
	return
}

func (d *DomainBook) GetDomain(domainName string) (*Domain, error) {
	d.RLock()
	domain, ok := d.domains[domainName]
	d.RUnlock()
	if !ok {
		return nil, fmt.Errorf("domain %s not found", domainName)
	}
	return domain, nil
}

func (d *DomainBook) GetDomains() []*Domain {
	var domains []*Domain
	d.RLock()
	for _, domain := range d.domains {
		domains = append(domains, domain)
	}
	d.RUnlock()
	return domains
}

func (d *DomainBook) GetDomainsInfo() []Domain {
	var domains []Domain
	d.RLock()
	for _, domain := range d.domains {
		dm := Domain{Name: domain.Name, Version: domain.Version}
		domains = append(domains, dm)
	}
	d.RUnlock()
	return domains
}

func (d *DomainBook) DeleteDomain(domainName string) {
	d.Lock()
	delete(d.domains, domainName)
	d.Unlock()
}

func (d *DomainBook) Reload(domainName string, version int, domainConf *conf.ConfManager) error {
	domain, err := d.GetDomain(domainName)
	if err != nil {
		return err
	}

	domain.SetVersion(version)
	for _, node := range domain.GetNodes() {
		structDatas := domainConf.Subscribe(node.GetSubscription())
		node.SetStructDatas(structDatas, version)
		node.AddEvent(model.Event{EventType: model.EventNodeConfChanged, Data: node.GetStructDatas()})
	}
	return nil
}

func (d *DomainBook) Monitor(interval time.Duration, expire int64) {
	for {
		time.Sleep(interval)
		for _, domain := range d.GetDomains() {
			for _, node := range domain.GetNodes() {
				if node.GetUpdateTime()+expire < time.Now().Unix() {
					node.SetStatus(model.StatusStop)
					//@todo 再考虑考虑，要不要删除节点信息
					domain.DelNode(node.ID)
				} else {
					// add prometheus alive
					metrics.Metrics.AddNodeAlive(node.ID, node.Domain, node.Hostname, 1)
				}
			}
		}
	}
}
