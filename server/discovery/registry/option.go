package registry

import (
	"fmt"
	"time"
)

// 日志
type Log interface {
	Debug(args ...interface{})
	Trace(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

type logger struct{}

//@todo log修改
func (l *logger) Debug(args ...interface{}) {
	fmt.Println(args...)
}
func (l *logger) Trace(args ...interface{}) {
	fmt.Println(args...)
}
func (l *logger) Warn(args ...interface{}) {
	fmt.Println(args...)
}
func (l *logger) Error(args ...interface{}) {
	fmt.Println(args...)
}

//默认logger
var DefaultLogger = new(logger)

//后端存储option
type Option struct {
	Addrs    []string
	Timeout  time.Duration
	Username string
	Password string

	Logger Log
}
type OpOption func(op *Option)

func WithAddrs(addrs ...string) OpOption {
	return func(op *Option) {
		op.Addrs = addrs
	}
}

func WithTimeout(timeout time.Duration) OpOption {
	return func(op *Option) {
		op.Timeout = timeout
	}
}

func WithAuth(username, password string) OpOption {
	return func(op *Option) {
		op.Username = username
		op.Password = password
	}
}

func WithLogger(logger Log) OpOption {
	return func(op *Option) {
		op.Logger = logger
	}
}

type RegisterOption struct {
	//租约时间
	TTL time.Duration
	//是否强制更新
	Refresh bool
}

type OpRegisterOption func(op *RegisterOption)

//续约时间
func WithRegistryTTL(ttl time.Duration) OpRegisterOption {
	return func(op *RegisterOption) {
		op.TTL = ttl
	}
}

//是否强制刷新，强制刷新时，重置数据（etcd会重新put数据）
func WithRefresh() OpRegisterOption {
	return func(op *RegisterOption) {
		op.Refresh = true
	}
}
