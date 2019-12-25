package conf

import (
	"flag"

	"github.com/go-redis/redis"
)

var (
	Redis       redis.Cmdable
	Listen      string
	KeeperAdmin string
)

func Init() error {
	flag.StringVar(&Listen, "p", ":8080", "listen")
	flag.StringVar(&KeeperAdmin, "k", "127.0.0.1:17000", "keeper admin address")
	flag.Parse()

	return nil
}
