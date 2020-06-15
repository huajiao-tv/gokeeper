//TODO 后期需要将该文件改成pb插件生成

package client

import (
	"context"
	"errors"
	"time"

	pb "github.com/huajiao-tv/gokeeper/pb/go"
	"github.com/silenceper/pool"
)

var (
	ErrTypeInValid = errors.New("stream from pool is invalid")
	ErrSendTimeout = errors.New("stream send timeout")
)

type SyncStream struct {
	pb.Sync_SyncClient
	cancel context.CancelFunc
}

type syncCallOption struct {
	timeout    time.Duration
	retryTimes int
}

type OpSyncCallOption func(o *syncCallOption)

func WithSyncCallTimeout(timeout time.Duration) OpSyncCallOption {
	return func(o *syncCallOption) {
		o.timeout = timeout
	}
}

func WithSyncCallRetryTimes(retryTimes int) OpSyncCallOption {
	if retryTimes <= 0 {
		retryTimes = 0
	}
	return func(o *syncCallOption) {
		o.retryTimes = retryTimes
	}
}

func NewSyncStreamPool(client pb.SyncClient, initCap, maxCap int, idleTimeout time.Duration) (pool.Pool, error) {
	streamFactory := func() (interface{}, error) {
		ctx, cancel := context.WithCancel(context.Background())
		s, err := client.Sync(ctx)
		if err != nil {
			return nil, err
		}
		return &SyncStream{
			s,
			cancel,
		}, nil
	}

	streamClose := func(v interface{}) error {
		stream, ok := v.(*SyncStream)
		if !ok {
			return ErrTypeInValid
		}
		stream.cancel()
		return stream.CloseSend()
	}

	poolConfig := &pool.Config{
		InitialCap:  initCap,
		MaxCap:      maxCap,
		Factory:     streamFactory,
		Close:       streamClose,
		IdleTimeout: idleTimeout,
	}
	return pool.NewChannelPool(poolConfig)
}

// @todo 如果server端处理能力不足，导致server recv buffer full（stream level），此时Send会阻塞，这时应该如何处理？？？
func Sync(p pool.Pool, evtReq *pb.ConfigEvent, opts ...OpSyncCallOption) (evtResp *pb.ConfigEvent, err error) {
	var (
		rawStream interface{}
		option    syncCallOption
	)

	for _, opt := range opts {
		opt(&option)
	}

RETRY:
	option.retryTimes--
	rawStream, err = p.Get()
	if err != nil {
		// 如果设置了重试次数，并且失败次数没有达到上限，则重试
		if option.retryTimes >= 0 {
			goto RETRY
		}
		return
	}

	stream, ok := rawStream.(*SyncStream)
	if !ok {
		err = ErrTypeInValid
		// 如果设置了重试次数，并且失败次数没有达到上限，则重试
		if option.retryTimes >= 0 {
			goto RETRY
		}
		return
	}

	//err为nil的情况下将连接放回连接池，否则直接关闭
	defer func() {
		if err == nil {
			err = p.Put(rawStream)
		} else {
			p.Close(stream)
		}
	}()

	if option.timeout <= 0 {
		err = stream.Send(evtReq)
		if err == nil {
			evtResp, err = stream.Recv()
		}
	} else {
		done := make(chan struct{})
		var streamErr error
		go func() {
			streamErr = stream.Send(evtReq)
			if streamErr == nil {
				evtResp, streamErr = stream.Recv()
			}
			close(done)
		}()
		timer := time.NewTimer(option.timeout)
		select {
		case <-timer.C:
			stream.cancel()
			err = ErrSendTimeout
		case <-done:
			if !timer.Stop() {
				<-timer.C
			}
			err = streamErr
		}
	}

	if err != nil {
		// 如果设置了重试次数，并且失败次数没有达到上限，则重试
		if option.retryTimes >= 0 {
			goto RETRY
		}
		return
	}

	return
}
