package storage

import (
	"github.com/huajiao-tv/gokeeper/server/storage/etcd"
	"github.com/huajiao-tv/gokeeper/server/storage/operate"
	"github.com/huajiao-tv/gokeeper/utility/logger"
)

type Storage interface {
	SetKey(domain, file, section, key, value, note string) error
	GetKey(domain, file, section, key string, withLock bool) (string, error)
	DelKey(domain, file, section, key, note string) error

	AddFile(domain, file string, data map[string]map[string]string, note string) error
	DelFile(domain, file, note string) error

	SetDomain(domain string, data map[string]map[string]map[string]string, note string) error
	GetDomain(domain string, withLock bool) (map[string]map[string]map[string]string, error)
	DelDomain(domain string, note string) error

	GetDomainNames(withLock bool) ([]string, error)

	Rollback(domain string, version int64, note string) error

	SetCurrentVersion(domain string, version int64) error
	GetCurrentVersion(domain string, withLock bool) (int64, error)
	GetMaxVersion(domain string, withLock bool) (int64, error)
	GetHistoryVersions(domain string, num, offset int64, withLock bool) (interface{}, error)

	SetKeeperAddr(domain, nodeID, addr string) error
	GetKeeperAddr(domain, nodeID string, withLock bool) (string, error)
	GetKeeperAddrs(domain string, withLock bool) ([]string, error)
	DelKeeperAddr(domain, nodeID, preValue string) error

	//keeper addr 续租
	KeepAlive(id int64, addr string)
	GetAliveKeeperNodes(withLock bool) (map[int64]string, error)

	Watch(operate.EventModeType, chan<- operate.Event)
}

const (
	EventChanSize = 10000
)

var (
	KStorage  Storage
	EventChan chan operate.Event
)

func InitStorage(storgaeUrl []string, storageUsername, storagePassword string, mode operate.EventModeType, logger *logger.Logger) error {
	var err error
	KStorage, err = etcd.NewEtcdStorage(storgaeUrl, storageUsername, storagePassword, logger)
	if err != nil {
		return err
	}
	EventChan = make(chan operate.Event, EventChanSize)
	go KStorage.Watch(mode, EventChan)
	return nil
}
