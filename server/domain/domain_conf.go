package domain

import (
	"errors"
	"fmt"
	"sync"

	"github.com/huajiao-tv/gokeeper/server/conf"
	"github.com/huajiao-tv/gokeeper/server/storage"
	"github.com/huajiao-tv/gokeeper/server/storage/operate"
)

type DomainConf struct {
	domains map[string]*conf.ConfManager
	sync.RWMutex
}

var (
	DomainConfs *DomainConf
)

func InitDomainConf() ([]string, error) {
	DomainConfs = &DomainConf{domains: map[string]*conf.ConfManager{}}
	if err := DomainConfs.init(); err != nil {
		return nil, err
	}
	return DomainConfs.GetDomainNames(), nil
}

func (d *DomainConf) init() error {
	//首次加载，不加锁
	domainNames, err := storage.KStorage.GetDomainNames(false)
	if err != nil {
		return err
	}
	for _, domainName := range domainNames {
		//首次加载，不加锁
		cf, e := conf.InitConfManager(domainName, false)
		if e != nil {
			err = e
			continue
		}
		d.domains[domainName] = cf
	}

	return err
}

func (d *DomainConf) Reload(domainName string) error {
	cf, err := conf.InitConfManager(domainName, true)
	if err != nil {
		return err
	}

	d.Lock()
	defer d.Unlock()
	d.domains[domainName] = cf

	return nil
}

func (d *DomainConf) GetDomainNames() []string {
	var names []string
	d.RLock()
	for name := range d.domains {
		names = append(names, name)
	}
	d.RUnlock()
	return names
}

func (d *DomainConf) GetDomain(domainName string) (*conf.ConfManager, error) {
	d.RLock()
	cf, ok := d.domains[domainName]
	d.RUnlock()
	if !ok {
		return nil, fmt.Errorf("domain %s not found", domainName)
	}
	return cf, nil
}

func (d *DomainConf) UpdateKey(event operate.Event) error {
	d.Lock()
	defer d.Unlock()

	_, ok := event.Data.(string)
	if !ok {
		return errors.New("event data is invalid")
	}
	cm, ok := d.domains[event.Domain]
	if !ok {
		cm, err := conf.NewConfManager(event)
		if err != nil {
			return err
		}
		d.domains[event.Domain] = cm
		DomainBooks.AddDomain(event.Domain)
		return nil
	}

	files := cm.FileList()
	existFile := false
	for _, file := range files {
		if file.Name == event.File {
			existFile = true
			break
		}
	}
	if !existFile {
		_, err := cm.NewFile(event)
		return err
	}
	if err := cm.Update(event); err != nil {
		return err
	}
	fmt.Println(storage.KStorage.GetDomain("mydomain", true))
	fmt.Println(d.domains["mydomain"].GetKeyData())
	/*node := DomainBooks.domains["mydomain"].nodes["node1"]
	node.AddEvent(model.Event{EventType: model.EventNodeConfChanged, Data: node.GetStructDatas()})*/
	return nil
}

func (d *DomainConf) DeleteKey(event operate.Event) error {
	cf, err := d.GetDomain(event.Domain)
	if err != nil {
		return err
	}
	return cf.Delete(event)
}

func (d *DomainConf) UpdateDomain(event operate.Event) error {
	version, ok := event.Data.(int64)
	if !ok {
		return errors.New("event.Data type is not int64")
	}
	domainBook, err := DomainBooks.GetDomain(event.Domain)
	if err != nil {
		err = DomainConfs.Reload(event.Domain)
		if err != nil {
			return err
		}
		//if load domain success, add domain
		DomainBooks.AddDomain(event.Domain)
		domainBook, err := DomainBooks.GetDomain(event.Domain)
		if err != nil {
			return err
		}
		domainBook.SetVersion(int(version))
		return nil
	} else {
		if domainBook.Version == int(version) {
			return nil
		}
		err = DomainConfs.Reload(event.Domain)
		if err != nil {
			return err
		}
		//if load success, update version
		domainBook.Version = int(version)
		return nil
	}
}

func (d *DomainConf) Debug() {
	for domainNames, cf := range d.domains {
		fmt.Println("domainName:", domainNames)
		cf.Debug()
		fmt.Println("----------------------------------------------------------")
	}
}
