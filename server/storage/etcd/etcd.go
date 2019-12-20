package etcd

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/huajiao-tv/gokeeper/server/storage/operate"
	"github.com/huajiao-tv/gokeeper/utility/cron"
	"github.com/huajiao-tv/gokeeper/utility/logger"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
	"go.etcd.io/etcd/mvcc/mvccpb"
)

type StorageEtcd struct {
	client  *clientv3.Client
	session *concurrency.Session
	kvApi   clientv3.KV

	cron   *cron.Cron
	logger *logger.Logger
	sync.RWMutex
}

func NewEtcdStorage(endpoints []string, username, password string, logger *logger.Logger) (*StorageEtcd, error) {
	cfg := clientv3.Config{
		Endpoints:   endpoints,
		Username:    username,
		Password:    password,
		DialTimeout: dialTimeout,
	}
	client, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}
	session, err := concurrency.NewSession(client)
	if err != nil {
		return nil, err
	}

	kvApi := clientv3.NewKV(client)

	se := &StorageEtcd{
		client:  client,
		session: session,
		kvApi:   kvApi,
		cron:    cron.New(),
		logger:  logger,
	}

	//监听session，如果done，需要重新生成，注意加锁
	go se.monitorSession()

	//add crontab
	se.cron.AddJob(cronTime, se)
	se.cron.Start()

	return se, nil
}

//监听session，如果keepAlive有报错，则重新初始化session
func (se *StorageEtcd) monitorSession() {
	for {
		select {
		case <-se.session.Done():
			se.logger.Error("monitorSession session done by some reason")
			//重新初始化session
			se.Lock()
			session, err := concurrency.NewSession(se.client)
			if err != nil {
				se.logger.Error("monitorSession NewSession error:", err)
			} else {
				se.session = session
			}
			se.Unlock()
		}
	}
}

func (se *StorageEtcd) set(key, value string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	var err error
	var resp *clientv3.PutResponse
	rc := 0
	for rc < retryCount {
		ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
		resp, err = se.kvApi.Put(ctx, key, value, opts...)
		cancel()
		if err == nil {
			return resp, nil
		}
		rc++
		time.Sleep(1 * time.Second)
	}
	return nil, err
}

func (se *StorageEtcd) get(key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	var err error
	var resp *clientv3.GetResponse
	rc := 0
	for rc < retryCount {
		ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
		resp, err = se.kvApi.Get(ctx, key, opts...)
		cancel()
		if err == nil {
			return resp, nil
		}
		rc++
		time.Sleep(1 * time.Second)
	}
	return nil, err
}

func (se *StorageEtcd) delete(key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	var err error
	var resp *clientv3.DeleteResponse
	rc := 0
	for rc < retryCount {
		ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
		resp, err = se.kvApi.Delete(ctx, key, opts...)
		cancel()
		if err == nil {
			return resp, nil
		}
		rc++
		time.Sleep(1 * time.Second)
	}
	return nil, err
}

//封装的函数只支持Then(),不添加超时重试处理
func (se *StorageEtcd) commitTransaction(ops ...clientv3.Op) (*clientv3.TxnResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()
	return se.kvApi.Txn(ctx).Then(ops...).Commit()
}

func (se *StorageEtcd) SetKey(domain, file, section, key, value, note string) error {
	se.Lock()
	defer se.Unlock()

	// lock domain when write(update domain version)
	locker := NewLocker(se.session, getLockPath(domain))
	if err := locker.Lock(); err != nil {
		return err
	}
	defer locker.Unlock()

	var err error
	var resp *clientv3.PutResponse
	path := getConfKeyPath(domain, file, section, key)
	resp, err = se.set(path, value)
	if err != nil {
		goto fail
	}
	err = se.incrVersion(domain, resp.Header.Revision, note)
	if err != nil {
		goto fail
	}
	return nil

fail:
	se.recoverDomain(domain)
	return err
}

func (se *StorageEtcd) GetKey(domain, file, section, key string, withLock bool) (string, error) {
	se.RLock()
	defer se.RUnlock()

	if withLock {
		locker := NewLocker(se.session, getLockPath(domain))
		if err := locker.Lock(); err != nil {
			return "", err
		}
		defer locker.Unlock()
	}

	dv, err := se.getCurrentVersionAux(domain)
	if err != nil {
		return "", err
	}

	path := getConfKeyPath(domain, file, section, key)
	resp, err := se.get(path, clientv3.WithRev(dv.EtcdVersion))
	if err != nil {
		return "", err
	}

	for _, ev := range resp.Kvs {
		if string(ev.Key) == path {
			return string(ev.Value), nil
		}
	}
	return "", ErrKeyNotExist
}

func (se *StorageEtcd) DelKey(domain, file, section, key, note string) error {
	se.Lock()
	defer se.Unlock()

	// lock domain when write(update domain version)
	locker := NewLocker(se.session, getLockPath(domain))
	if err := locker.Lock(); err != nil {
		return err
	}
	defer locker.Unlock()

	var err error
	var resp *clientv3.DeleteResponse
	path := getConfKeyPath(domain, file, section, key)
	resp, err = se.delete(path)
	if err != nil {
		goto fail
	}
	err = se.incrVersion(domain, resp.Header.Revision, note)
	if err != nil {
		goto fail
	}
	return nil

fail:
	se.recoverDomain(domain)
	return err
}

//如果该文件已经存在，则更新
func (se *StorageEtcd) AddFile(domain, file string, data map[string]map[string]string, note string) error {
	se.Lock()
	defer se.Unlock()

	// lock domain when write(update domain version)
	locker := NewLocker(se.session, getLockPath(domain))
	if err := locker.Lock(); err != nil {
		return err
	}
	defer locker.Unlock()

	var err error
	var etcdVersion int64
	var originData map[string]map[string]string

	originData, err = se.getFileAux(domain, file, -1)
	if err != nil {
		return err
	}
	etcdVersion, err = se.reviseData(domain, map[string]map[string]map[string]string{file: data}, map[string]map[string]map[string]string{file: originData})
	if err != nil {
		goto fail
	}
	if etcdVersion == 0 {
		return nil
	}
	err = se.incrVersion(domain, etcdVersion, note)
	if err != nil {
		goto fail
	}
	return nil

fail:
	se.recoverDomain(domain)
	return err
}

func (se *StorageEtcd) DelFile(domain, file, note string) error {
	se.Lock()
	defer se.Unlock()

	locker := NewLocker(se.session, getLockPath(domain))
	if err := locker.Lock(); err != nil {
		return err
	}
	defer locker.Unlock()

	var err error
	var resp *clientv3.DeleteResponse
	resp, err = se.delete(getConfFilePath(domain, file), clientv3.WithPrefix())
	if err != nil {
		goto fail
	}
	if resp.Deleted == 0 {
		return ErrKeyNotExist
	}
	err = se.incrVersion(domain, resp.Header.Revision, note)
	if err != nil {
		goto fail
	}
	return nil

fail:
	se.recoverDomain(domain)
	return err
}

func (se *StorageEtcd) getFileAux(domain, file string, etcdVersion int64) (map[string]map[string]string, error) {
	var resp *clientv3.GetResponse
	var err error
	if etcdVersion >= 0 {
		resp, err = se.get(getConfFilePath(domain, file), clientv3.WithPrefix(), clientv3.WithRev(etcdVersion))
	} else {
		resp, err = se.get(getConfFilePath(domain, file), clientv3.WithPrefix())
	}

	if err != nil {
		return nil, err
	}

	data := map[string]map[string]string{}
	for _, ev := range resp.Kvs {
		_, _, section, key, err := parseConfKeyPath(string(ev.Key))
		if err != nil {
			se.logger.Error("Etcd getDomainAux", "parseConfKeyPath", string(ev.Key), string(ev.Value), err.Error())
			continue
		}
		_, exist := data[section]
		if !exist {
			data[section] = map[string]string{}
		}
		data[section][key] = string(ev.Value)
	}
	return data, nil
}

func (se *StorageEtcd) SetDomain(domain string, data map[string]map[string]map[string]string, note string) error {
	se.Lock()
	defer se.Unlock()

	// lock domain when write(update domain version)
	locker := NewLocker(se.session, getLockPath(domain))
	if err := locker.Lock(); err != nil {
		return err
	}
	defer locker.Unlock()

	var err error
	var etcdVersion int64
	etcdVersion, err = se.setDomainAux(domain, data)
	if err != nil {
		goto fail
	}
	err = se.incrVersion(domain, etcdVersion, note)
	if err != nil {
		goto fail
	}
	return nil

fail:
	se.recoverDomain(domain)
	return err
}

func (se *StorageEtcd) setDomainAux(domain string, data map[string]map[string]map[string]string) (int64, error) {
	var ops []clientv3.Op
	for file, fileData := range data {
		for section, sectionData := range fileData {
			for key, value := range sectionData {
				ops = append(ops, clientv3.OpPut(getConfKeyPath(domain, file, section, key), value))
			}
		}
	}
	resp, err := se.commitTransaction(ops...)
	if err != nil {
		return 0, err
	}
	return resp.Header.Revision, nil
}

func (se *StorageEtcd) GetDomain(domain string, withLock bool) (map[string]map[string]map[string]string, error) {
	se.RLock()
	defer se.RUnlock()

	if withLock {
		locker := NewLocker(se.session, getLockPath(domain))
		if err := locker.Lock(); err != nil {
			return nil, err
		}
		defer locker.Unlock()
	}

	dv, err := se.getCurrentVersionAux(domain)
	if err != nil {
		return nil, err
	}
	return se.getDomainAux(domain, dv.EtcdVersion)
}

func (se *StorageEtcd) getDomainAux(domain string, etcdVersion int64) (map[string]map[string]map[string]string, error) {
	var resp *clientv3.GetResponse
	var err error
	if etcdVersion >= 0 {
		resp, err = se.get(getConfDomainPath(domain), clientv3.WithPrefix(), clientv3.WithRev(etcdVersion))
	} else {
		resp, err = se.get(getConfDomainPath(domain), clientv3.WithPrefix())
	}

	if err != nil {
		return nil, err
	}

	data := map[string]map[string]map[string]string{}
	for _, ev := range resp.Kvs {
		_, file, section, key, err := parseConfKeyPath(string(ev.Key))
		if err != nil {
			se.logger.Error("Etcd getDomainAux", "parseConfKeyPath", string(ev.Key), string(ev.Value), err.Error())
			continue
		}
		_, exist := data[file]
		if !exist {
			data[file] = map[string]map[string]string{}
		}
		_, exist = data[file][section]
		if !exist {
			data[file][section] = map[string]string{}
		}
		data[file][section][key] = string(ev.Value)
	}
	return data, nil
}

//???是否需要同时删除version
func (se *StorageEtcd) DelDomain(domain string, note string) error {
	se.Lock()
	defer se.Unlock()

	locker := NewLocker(se.session, getLockPath(domain))
	if err := locker.Lock(); err != nil {
		return err
	}
	defer locker.Unlock()

	var err error
	var etcdVersion int64
	etcdVersion, err = se.delDomainAux(domain)
	if err != nil {
		goto fail
	}
	err = se.incrVersion(domain, etcdVersion, note)
	if err != nil {
		goto fail
	}
	return nil

fail:
	se.recoverDomain(domain)
	return err
}

func (se *StorageEtcd) delDomainAux(domain string) (int64, error) {
	resp, err := se.delete(getConfDomainPath(domain), clientv3.WithPrefix())
	if err != nil {
		return 0, err
	}
	if resp.Deleted == 0 {
		return 0, ErrKeyNotExist
	}
	return resp.Header.Revision, nil
}

//??? TODO
func (se *StorageEtcd) SetCurrentVersion(domain string, version int64) error {
	return nil
}

func (se *StorageEtcd) GetCurrentVersion(domain string, withLock bool) (int64, error) {
	se.RLock()
	defer se.RUnlock()

	if withLock {
		locker := NewLocker(se.session, getLockPath(domain))
		if err := locker.Lock(); err != nil {
			return 0, err
		}
		defer locker.Unlock()
	}

	dv, err := se.getCurrentVersionAux(domain)
	if err != nil {
		return 0, err
	}
	return dv.Version, nil
}

func (se *StorageEtcd) getCurrentVersionAux(domain string) (*domainVersion, error) {
	path := getCurrentVersionKeyPath(domain)
	resp, err := se.get(path)
	if err != nil {
		return nil, err
	}
	for _, ev := range resp.Kvs {
		if string(ev.Key) == path {
			return decodeDomainVersion(string(ev.Value))
		}
	}
	return nil, ErrKeyNotExist
}

func (se *StorageEtcd) GetDomainNames(withLock bool) ([]string, error) {
	se.RLock()
	defer se.RUnlock()

	if withLock {
		locker := NewLocker(se.session, getLockPath(getGoKeeperRootPath()))
		if err := locker.Lock(); err != nil {
			return nil, err
		}
		defer locker.Unlock()
	}

	return se.getDomainNamesAux()
}

func (se *StorageEtcd) getDomainNamesAux() ([]string, error) {
	path := getCurrentVersionDirPath()
	resp, err := se.get(path, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	var domainNames []string
	for _, ev := range resp.Kvs {
		domainName, err := parseCurrentVersionKeyPath(string(ev.Key))
		if err != nil {
			continue
		}
		domainNames = append(domainNames, domainName)
	}
	return domainNames, nil
}

func (se *StorageEtcd) getDomainRecode(domain string, version int64) (*Recode, error) {
	path := getVersionKeyPath(domain, version)
	resp, err := se.get(path)
	if err != nil {
		return nil, err
	}
	for _, ev := range resp.Kvs {
		if string(ev.Key) == path {
			return decodeRecode(string(ev.Value))
		}
	}
	return nil, ErrKeyNotExist
}

func (se *StorageEtcd) Rollback(domain string, version int64, note string) error {
	se.Lock()
	defer se.Unlock()

	locker := NewLocker(se.session, getLockPath(domain))
	if err := locker.Lock(); err != nil {
		return err
	}
	defer locker.Unlock()

	var err error
	var recode *Recode
	var etcdVersion int64

	recode, err = se.getDomainRecode(domain, version)
	rollbackData, err := se.getDomainAux(domain, recode.PackageVersion)
	if err != nil {
		return err
	}
	latestVersionData, err := se.getDomainAux(domain, -1)
	if err != nil {
		return err
	}
	etcdVersion, err = se.reviseData(domain, rollbackData, latestVersionData)
	if err != nil {
		goto fail
	}
	if etcdVersion == 0 {
		etcdVersion = recode.PackageVersion
	}
	err = se.incrVersion(domain, etcdVersion, note)
	if err != nil {
		goto fail
	}
	return nil

fail:
	se.recoverDomain(domain)
	return err
}

func (se *StorageEtcd) GetMaxVersion(domain string, withLock bool) (int64, error) {
	se.RLock()
	defer se.RUnlock()

	if withLock {
		locker := NewLocker(se.session, getLockPath(domain))
		if err := locker.Lock(); err != nil {
			return 0, err
		}
		defer locker.Unlock()
	}

	dv, err := se.getCurrentVersionAux(domain)
	if err != nil {
		return 0, err
	}
	return dv.Version, nil
}

func (se *StorageEtcd) GetHistoryVersions(domain string, num, offset int64, withLock bool) (interface{}, error) {
	se.RLock()
	defer se.RUnlock()

	if withLock {
		locker := NewLocker(se.session, getLockPath(domain))
		if err := locker.Lock(); err != nil {
			return 0, err
		}
		defer locker.Unlock()
	}

	dv, err := se.getCurrentVersionAux(domain)
	if err != nil {
		return nil, err
	}
	endVersion := dv.Version - offset
	if endVersion < 0 {
		return nil, errors.New("have no more versions")
	}
	startVersion := dv.Version - offset - num + 1
	if startVersion < 0 {
		startVersion = 0
	}

	var ops []clientv3.Op
	for v := endVersion; v >= startVersion; v-- {
		ops = append(ops, clientv3.OpGet(getVersionKeyPath(domain, v)))
	}
	resp, err := se.commitTransaction(ops...)
	if err != nil {
		return nil, err
	}

	var recodes []*Recode
	for _, r := range resp.Responses {
		if rangeResp := r.GetResponseRange(); rangeResp != nil {
			for _, ev := range rangeResp.Kvs {
				recode, err := decodeRecode(string(ev.Value))
				if err == nil {
					recodes = append(recodes, recode)
				}
			}
		}
	}
	return recodes, err
}

func (se *StorageEtcd) getLatestEtcdVersion() (int64, error) {
	resp, err := se.get(getGoKeeperRootPath())
	if err != nil {
		return 0, err
	}
	return resp.Header.Revision, nil
}

func (se *StorageEtcd) incrVersion(domain string, etcdVersion int64, note string) error {
	dv, err := se.getCurrentVersionAux(domain)

	var currentVersion int64
	if err != nil && err != ErrKeyNotExist {
		return err
	}
	if err == ErrKeyNotExist {
		currentVersion = 0
	} else {
		currentVersion = dv.Version
	}
	currentVersion += 1

	versionData, err := encodeDomainVersion(domainVersion{currentVersion, etcdVersion})
	if err != nil {
		return err
	}

	recode := &Recode{
		ID:             currentVersion,
		Domain:         domain,
		Version:        currentVersion,
		PackageVersion: etcdVersion,
		Note:           note,
		Timestamp:      time.Now().Unix(),
	}
	recodeData, err := encodeRecode(*recode)
	if err != nil {
		return err
	}
	ops := []clientv3.Op{
		clientv3.OpPut(getCurrentVersionKeyPath(domain), versionData),
		clientv3.OpPut(getVersionKeyPath(domain, currentVersion), recodeData),
	}
	_, err = se.commitTransaction(ops...)
	return err
}

func (se *StorageEtcd) updateVersion(domain string, version, etcdVersion int64, recode *Recode) error {
	versionData, err := encodeDomainVersion(domainVersion{version, etcdVersion})
	if err != nil {
		return err
	}
	recode.PackageVersion = etcdVersion
	recodeData, err := encodeRecode(*recode)
	if err != nil {
		return err
	}
	ops := []clientv3.Op{
		clientv3.OpPut(getCurrentVersionKeyPath(domain), versionData),
		clientv3.OpPut(getVersionKeyPath(domain, version), recodeData),
	}
	_, err = se.commitTransaction(ops...)
	return err
}

func (se *StorageEtcd) SetKeeperAddr(domain, nodeID, addr string) error {
	se.Lock()
	defer se.Unlock()

	path := getAddrKeyPath(domain, nodeID)

	locker := NewLocker(se.session, getLockPath(path))
	if err := locker.Lock(); err != nil {
		return err

	}
	defer locker.Unlock()

	_, err := se.set(path, addr)
	return err
}

func (se *StorageEtcd) GetKeeperAddr(domain, nodeID string, withLock bool) (string, error) {
	se.Lock()
	defer se.Unlock()

	path := getAddrKeyPath(domain, nodeID)

	if withLock {
		locker := NewLocker(se.session, getLockPath(path))
		if err := locker.Lock(); err != nil {
			return "", err
		}
		defer locker.Unlock()
	}

	resp, err := se.get(path)
	if err != nil {
		return "", err
	}
	for _, ev := range resp.Kvs {
		if string(ev.Key) == path {
			return string(ev.Value), nil
		}
	}
	return "", ErrKeyNotExist
}

func (se *StorageEtcd) GetKeeperAddrs(domain string, withLock bool) ([]string, error) {
	se.Lock()
	defer se.Unlock()

	path := getAddrDirPath(domain)

	if withLock {
		locker := NewLocker(se.session, getLockPath(path))
		if err := locker.Lock(); err != nil {
			return nil, err
		}
		defer locker.Unlock()
	}

	resp, err := se.get(path, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	var addrs []string
	existAddrs := map[string]bool{}
	for _, ev := range resp.Kvs {
		addr := string(ev.Value)
		if _, exist := existAddrs[addr]; !exist {
			addrs = append(addrs, addr)
			existAddrs[addr] = true
		}
	}
	return addrs, nil
}

func (se *StorageEtcd) DelKeeperAddr(domain, nodeID, preAddr string) error {
	se.Lock()
	defer se.Unlock()

	// lock domain when write(update domain version)
	locker := NewLocker(se.session, getLockPath(domain))
	if err := locker.Lock(); err != nil {
		return err
	}
	defer locker.Unlock()

	path := getAddrKeyPath(domain, nodeID)
	if len(preAddr) == 0 {
		_, err := se.delete(path)
		return err
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
		tx := se.kvApi.Txn(ctx)
		_, err := tx.If(clientv3.Compare(clientv3.Value(path), "=", preAddr)).Then(clientv3.OpDelete(path)).Commit()
		cancel()
		return err
	}
}

//keepalive需要消费chan，如果不及时消费，keepalive会500ms请求一次心跳操作，导致qps上升
//如果keepalive返回的chan被关闭，需要重新申请leaseID，然后再进行重新keepalive。（注意，超过续约时间，leaseID会失效）
func (se *StorageEtcd) KeepAlive(id int64, addr string) {
	path := getNodeLeaseKeyPath(id)
retry:
	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	grantResp, err := se.client.Grant(ctx, keeperAddrLeaseTTL)
	cancel()
	if err != nil {
		se.logger.Error("KeepAlive Grant error:", err)
		time.Sleep(1 * time.Second)
		goto retry
	}
	_, err = se.set(path, addr, clientv3.WithLease(grantResp.ID))
	if err != nil {
		se.logger.Error("KeepAlive set error:", err, path, addr, grantResp.ID)
		time.Sleep(1 * time.Second)
		goto retry
	}
	//不消费chan时影响可以参考 go.etcd.io/etcd/issues/7446
	kpChan, err := se.client.KeepAlive(context.Background(), grantResp.ID)
	if err != nil {
		se.logger.Error("KeepAlive set error:", err, path, addr, grantResp.ID)
		time.Sleep(1 * time.Second)
		goto retry
	}
	for {
		select {
		case _, ok := <-kpChan:
			// chan 被关闭了，需要重新keepalive
			if !ok {
				se.logger.Error("KeepAlive chan is close, retry to keepalive again")
				time.Sleep(1 * time.Second)
				goto retry
			}
		}
	}
}

func (se *StorageEtcd) GetAliveKeeperNodes(withLock bool) (map[int64]string, error) {
	se.Lock()
	defer se.Unlock()

	path := getNodeLeaseDirPath()

	if withLock {
		locker := NewLocker(se.session, getLockPath(path))
		if err := locker.Lock(); err != nil {
			return nil, err
		}
		defer locker.Unlock()
	}

	resp, err := se.get(path, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	nodes := map[int64]string{}
	for _, ev := range resp.Kvs {
		id, err := parseNodeLeaseKeyPath(string(ev.Key))
		if err != nil {
			se.logger.Error("Etcd GetAliveKeeperNodes", "parseNodeLeaseKeyPath", string(ev.Key), string(ev.Value), err.Error())
			continue
		}
		nodes[id] = string(ev.Value)
	}
	return nodes, nil
}

func (se *StorageEtcd) isCron() bool {
	resp, err := se.get(getCronPath())
	if err != nil {
		return false
	}
	if resp.Count == 0 {
		return true
	}
	return false
}

func (se *StorageEtcd) leaseCron() error {
	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	grantResp, err := se.client.Grant(ctx, cronLeaseTTL)
	cancel()
	if err != nil {
		return err
	}
	_, err = se.set(getCronPath(), "cron", clientv3.WithLease(grantResp.ID))
	return err
}

func (se *StorageEtcd) Run() {
	se.Lock()
	defer se.Unlock()

	locker := NewLocker(se.session, getLockPath(getGoKeeperRootPath()))
	if err := locker.Lock(); err != nil {
		return
	}
	defer locker.Unlock()

	//多个节点在指定时间只执行一次
	if !se.isCron() {
		return
	}
	if err := se.leaseCron(); err != nil {
		return
	}

	domainNames, err := se.getDomainNamesAux()
	if err != nil {
		se.logger.Error("Etcd Run", "getDomainNamesAux", err.Error())
		return
	}
	latestEtcdVersion, err := se.getLatestEtcdVersion()
	if err != nil {
		se.logger.Error("Etcd Run", "getLatestEtcdVersion", err.Error())
		return
	}

	for _, domainName := range domainNames {
		err = se.syncDomain(domainName, latestEtcdVersion)
		if err != nil {
			se.logger.Error("Etcd Run", "syncDomain", domainName, latestEtcdVersion, err.Error())
			continue
		}
	}
}

func (se *StorageEtcd) syncDomain(domain string, latestEtcdVersion int64) error {
	dv, err := se.getCurrentVersionAux(domain)
	if err != nil {
		return err
	}
	recode, err := se.getDomainRecode(domain, dv.Version)
	if err != nil {
		return err
	}

	currentVersionData, err := se.getDomainAux(domain, dv.EtcdVersion)
	if err != nil {
		return err
	}
	latestVersionData, err := se.getDomainAux(domain, -1)
	if err != nil {
		return err
	}
	modifiedEtcdVersion, err := se.reviseData(domain, currentVersionData, latestVersionData)
	if err != nil {
		return err
	}
	if modifiedEtcdVersion == 0 {
		modifiedEtcdVersion = latestEtcdVersion
	}
	err = se.updateVersion(domain, dv.Version, modifiedEtcdVersion, recode)
	if err != nil {
		return err
	}
	return nil
}

func (se *StorageEtcd) reviseData(domain string, rightData, dirtyData map[string]map[string]map[string]string) (int64, error) {
	putData := mapDiff(rightData, dirtyData, true)
	deleteData := mapDiff(dirtyData, rightData, false)

	var ops []clientv3.Op
	if len(putData) > 0 {
		for file, fileData := range putData {
			for section, sectionData := range fileData {
				for key, value := range sectionData {
					ops = append(ops, clientv3.OpPut(getConfKeyPath(domain, file, section, key), value))
				}
			}
		}
	}
	if len(deleteData) > 0 {
		for file, fileData := range deleteData {
			for section, sectionData := range fileData {
				for key := range sectionData {
					ops = append(ops, clientv3.OpDelete(getConfKeyPath(domain, file, section, key)))
				}
			}
		}
	}
	if len(ops) > 0 {
		resp, err := se.commitTransaction(ops...)
		if err != nil {
			return -1, err
		}
		return resp.Header.Revision, nil
	} else {
		return 0, nil
	}
}

func (se *StorageEtcd) recoverDomain(domain string) error {
	latestEtcdVersion, err := se.getLatestEtcdVersion()
	if err != nil {
		se.logger.Error("Etcd recoverDomain", "getLatestEtcdVersion", domain, err.Error())
		return err
	}
	err = se.syncDomain(domain, latestEtcdVersion)
	if err != nil {
		se.logger.Error("Etcd recoverDomain", "syncDomain", domain, latestEtcdVersion, err.Error())
		return err
	}
	return nil
}

func (se *StorageEtcd) Watch(mode operate.EventModeType, storageEventChan chan<- operate.Event) {
	watcher := clientv3.NewWatcher(se.client)
	switch mode {
	case operate.EventModeConf:
		// watch conf change
		confWatcher := watcher.Watch(context.TODO(), getConfDirPath(), clientv3.WithPrefix())
		for {
			select {
			case confResp := <-confWatcher:
				for _, ev := range confResp.Events {
					var value string
					var opcode operate.Opcode
					domain, file, section, key, err := parseConfKeyPath(string(ev.Kv.Key))
					if err != nil {
						se.logger.Error("Etcd Watch", "parseConfKeyPath", string(ev.Kv.Key), string(ev.Kv.Value), err.Error())
						continue
					}
					switch ev.Type {
					case mvccpb.PUT:
						opcode = operate.OpcodeUpdateKey
						value = string(ev.Kv.Value)
					case mvccpb.DELETE:
						opcode = operate.OpcodeDeleteKey
					}
					event := operate.Event{Opcode: opcode, Domain: domain, File: file, Section: section, Key: key, Data: value}
					storageEventChan <- event
				}
			}
		}
	case operate.EventModeVersion:
		// watch version change
		versionWatcher := watcher.Watch(context.TODO(), getCurrentVersionDirPath(), clientv3.WithPrefix())
		for {
			select {
			case versionResp := <-versionWatcher:
				for _, ev := range versionResp.Events {
					domain, err := parseCurrentVersionKeyPath(string(ev.Kv.Key))
					if err != nil {
						se.logger.Error("Etcd Watch", "parseCurrentVersionKeyPath", string(ev.Kv.Key), string(ev.Kv.Value), err.Error())
						continue
					}
					domainVersion, err := decodeDomainVersion(string(ev.Kv.Value))
					if err != nil {
						se.logger.Error("Etcd Watch", "decodeDomainVersion", string(ev.Kv.Key), string(ev.Kv.Value), err.Error())
						continue
					}
					event := operate.Event{
						Opcode: operate.OpcodeUpdateDomain,
						Domain: domain,
						Data:   domainVersion.Version,
					}
					storageEventChan <- event
				}
			}
		}
	}
}
