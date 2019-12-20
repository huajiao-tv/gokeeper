package etcd

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	PathSeparator    = string(filepath.Separator)
	GokeeperRootPath = "gokeeper"
	ConfRootPath     = "conf"

	VersionRootPath    = "version"
	AllVersionPath     = "all"
	CurrentVersionPath = "current"

	AddrRootPath = "addr"

	//租约path
	LeaseRootPath = "lease"
	NodeLeasePath = "node"

	LockRootPath = "lock"
	lockTimeout  = 5 * time.Second

	dialTimeout        = 5 * time.Second //etcd连接超时时间
	readTimeout        = 5 * time.Second //etcd读超时时间,如果都是读超时，则etcd操作的等待实际为 readTimeout * retryCount
	writeTimeout       = 5 * time.Second //etcd写超时时间,如果都是写超时，则etcd操作的等待实际为 writeTimeout * retryCount
	keeperAddrLeaseTTL = 3600            //keeper租约时间 每个ttl/3自动续约一次
	retryCount         = 3               //失败重试次数

	CronPath     = "cron"
	cronTime     = "0 0 0 * *" // crontab 时间设置
	cronLeaseTTL = 3600        //cron 锁定时长，确保多个几点情况下，只有一个节点执行cron
)

var (
	ErrKeyNotExist = errors.New("key not exist")
)

func getGoKeeperRootPath() string {
	return filepath.Join(PathSeparator, GokeeperRootPath)
}

func getConfDirPath() string {
	return filepath.Join(PathSeparator, GokeeperRootPath, ConfRootPath)
}

//注意，这里防止prefix导致混杂的情况（比如，session会把session_test相关的配置取到），所以拼接path路径时，最后要追加/
func getConfDomainPath(domain string) string {
	return filepath.Join(PathSeparator, GokeeperRootPath, ConfRootPath, domain) + PathSeparator
}

func getConfFilePath(domain, file string) string {
	return filepath.Join(PathSeparator, GokeeperRootPath, ConfRootPath, domain, file)
}

func getConfKeyPath(domain, file, section, key string) string {
	return filepath.Join(PathSeparator, GokeeperRootPath, ConfRootPath, domain, file, section, key)
}

func getVersionKeyPath(domain string, version int64) string {
	return filepath.Join(PathSeparator, GokeeperRootPath, VersionRootPath, AllVersionPath, domain, strconv.FormatInt(version, 10))
}

func getCurrentVersionDirPath() string {
	return filepath.Join(PathSeparator, GokeeperRootPath, VersionRootPath, CurrentVersionPath)
}

func getCurrentVersionKeyPath(domain string) string {
	return filepath.Join(PathSeparator, GokeeperRootPath, VersionRootPath, CurrentVersionPath, domain)
}

func getAddrDirPath(domain string) string {
	return filepath.Join(PathSeparator, GokeeperRootPath, AddrRootPath, domain)
}

func getAddrKeyPath(domain, nodeID string) string {
	return filepath.Join(PathSeparator, GokeeperRootPath, AddrRootPath, domain, nodeID)
}

func getNodeLeaseDirPath() string {
	return filepath.Join(PathSeparator, GokeeperRootPath, LeaseRootPath, NodeLeasePath)
}

func getNodeLeaseKeyPath(id int64) string {
	return filepath.Join(PathSeparator, GokeeperRootPath, LeaseRootPath, NodeLeasePath, fmt.Sprintf("%d", id))
}

func parseConfKeyPath(path string) (domain, file, section, key string, err error) {
	list := strings.Split(strings.Trim(path, PathSeparator), PathSeparator)
	if len(list) < 6 || list[0] != GokeeperRootPath || list[1] != ConfRootPath {
		return "", "", "", "", errors.New("conf path is invalid")
	}
	pathLen := len(list)
	return list[2], formatFilePath(list[3 : pathLen-2]...), list[pathLen-2], list[pathLen-1], nil
}

func parseCurrentVersionKeyPath(path string) (domain string, err error) {
	list := strings.Split(strings.Trim(path, PathSeparator), PathSeparator)
	if len(list) != 4 || list[0] != GokeeperRootPath || list[1] != VersionRootPath || list[2] != CurrentVersionPath {
		return "", errors.New("current version path is invalid")
	}
	return list[3], nil
}

func parseNodeLeaseKeyPath(path string) (id int64, err error) {
	list := strings.Split(strings.Trim(path, PathSeparator), PathSeparator)
	if len(list) != 4 || list[0] != GokeeperRootPath || list[1] != LeaseRootPath || list[2] != NodeLeasePath {
		return -1, errors.New("node lease path is invalid")
	}
	return strconv.ParseInt(list[3], 10, 64)
}

func getLockPath(path string) string {
	return filepath.Join(PathSeparator, GokeeperRootPath, LockRootPath, path)
}

func getCronPath() string {
	return filepath.Join(PathSeparator, GokeeperRootPath, CronPath)
}

func formatFilePath(names ...string) string {
	tmpNames := []string{PathSeparator}
	tmpNames = append(tmpNames, names...)
	return filepath.Join(tmpNames...)
}
