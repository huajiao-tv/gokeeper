package discovery

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

type PollsStream struct {
	pb.Discovery_PollsClient
	cancel context.CancelFunc
}

type pollsCallOption struct {
	timeout    time.Duration
	retryTimes int
}

type OpPollsCallOption func(o *pollsCallOption)

func WithPollsCallTimeout(timeout time.Duration) OpPollsCallOption {
	return func(o *pollsCallOption) {
		o.timeout = timeout
	}
}

func WithPollsCallRetryTimes(retryTimes int) OpPollsCallOption {
	if retryTimes <= 0 {
		retryTimes = 0
	}
	return func(o *pollsCallOption) {
		o.retryTimes = retryTimes
	}
}

func NewPollsStreamPool(client pb.DiscoveryClient, initCap, maxCap int, idleTimeout time.Duration) (pool.Pool, error) {
	streamFactory := func() (interface{}, error) {
		ctx, cancel := context.WithCancel(context.Background())
		s, err := client.Polls(ctx)
		if err != nil {
			return nil, err
		}
		return &PollsStream{
			s,
			cancel,
		}, nil
	}

	streamClose := func(v interface{}) error {
		stream, ok := v.(*PollsStream)
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
func Polls(p pool.Pool, req *pb.PollsReq, opts ...OpPollsCallOption) (resp *pb.PollsResp, err error) {
	var (
		rawStream interface{}
		option    pollsCallOption
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

	stream, ok := rawStream.(*PollsStream)
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
		err = stream.Send(req)
		if err == nil {
			resp, err = stream.Recv()
		}
	} else {
		done := make(chan struct{})
		var streamErr error
		go func() {
			streamErr = stream.Send(req)
			if streamErr == nil {
				resp, streamErr = stream.Recv()
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
