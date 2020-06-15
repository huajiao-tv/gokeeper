package etcd

import (
	"context"

	"go.etcd.io/etcd/clientv3/concurrency"
)

type lockerMutex struct{ *concurrency.Mutex }

func NewLocker(session *concurrency.Session, pfx string) *lockerMutex {
	lockKey := getLockPath(pfx)
	return &lockerMutex{concurrency.NewMutex(session, lockKey)}
}

func (lm *lockerMutex) Lock() error {
	ctx, cancel := context.WithTimeout(context.Background(), lockTimeout)
	defer cancel()
	if err := lm.Mutex.Lock(ctx); err != nil {
		return err
	}
	return nil
}

func (lm *lockerMutex) Unlock() error {
	ctx, cancel := context.WithTimeout(context.Background(), lockTimeout)
	defer cancel()
	return lm.Mutex.Unlock(ctx)
}
