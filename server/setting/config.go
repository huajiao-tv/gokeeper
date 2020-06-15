package setting

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/huajiao-tv/gokeeper/server/storage/operate"
	"github.com/huajiao-tv/gokeeper/utility/go-ini/ini"
)

const (
	Component        = "gokeeper"
	defaultEventMode = operate.EventModeVersion
)

var (
	ConfigFile string

	GoRpcListen string // gorpc listen（采用内部gorpc协议）
	GRpcListen  string // grpc listen（采用google grpc协议)
	AdminListen string // admin listen
	PromListen  string // prometheus metrics listen

	//配置存储
	StorageUrl      []string
	StorageUsername string
	StoragePassword string
	EventMode       operate.EventModeType //etcd 监听的事件方式 conf：监听domain conf  version：监听domain version，默认为version

	//服务发现后端存储
	RegistryUrl      []string
	RegistryUsername string
	RegistryPassword string

	BasePath string
	LogPath  = "log"
	TmpPath  = "tmp"

	EventInterval = 60
	TestMode      = false

	GrpcListener net.Listener //grpc中配置和服务发现的端口暂时共用
)

func InitConfig() error {
	flag.StringVar(&ConfigFile, "f", "/path/to/keeper.conf", "keeper config file path")
	flag.Int64Var(&KeeperID, "i", -1, "keeper id")
	flag.Parse()

	var err error
	c, err := ini.Load(ConfigFile)
	if err != nil {
		return err
	}

	if GoRpcListen = c.Section("").Key("gorpc_listen").MustString(""); GoRpcListen == "" {
		return errors.New("gorpc_listen is empty")
	}
	if GRpcListen = c.Section("").Key("grpc_listen").MustString(""); GRpcListen == "" {
		return errors.New("grpc_listen is empty")
	}
	if AdminListen = c.Section("").Key("admin_listen").MustString(""); AdminListen == "" {
		return errors.New("admin_listen is empty")
	}
	if PromListen = c.Section("").Key("prom_listen").MustString(""); PromListen == "" {
		return errors.New("prom_listen is empty")
	}
	if BasePath = c.Section("").Key("base_path").MustString(""); BasePath == "" {
		return errors.New("base_path is empty")
	}
	if StorageUrl = c.Section("").Key("storage_url").Strings(","); len(StorageUrl) == 0 {
		return errors.New("storage_url is empty")
	}
	StorageUsername = c.Section("").Key("storage_username").String()
	StoragePassword = c.Section("").Key("storage_password").String()
	if mode := c.Section("").Key("event_mode").MustString(""); operate.IsValidEventMode(mode) {
		EventMode = operate.EventModeType(mode)
	} else {
		EventMode = defaultEventMode
	}
	if RegistryUrl = c.Section("").Key("registry_url").Strings(","); len(RegistryUrl) == 0 {
		return errors.New("registry_url is empty")
	}
	RegistryUsername = c.Section("").Key("registry_username").String()
	RegistryPassword = c.Section("").Key("registry_password").String()
	LogPath = c.Section("").Key("log_path").MustString(filepath.Join(BasePath, LogPath))
	TmpPath = c.Section("").Key("tmp_path").MustString(filepath.Join(BasePath, TmpPath))

	if err = InitKeeperAddr(GoRpcListen, AdminListen); err != nil {
		return errors.New("KeeperAddr is invalid: " + err.Error())
	}

	EventInterval = c.Section("").Key("event_interval").MustInt(EventInterval)
	TestMode = c.Section("").Key("test_mode").MustBool(false)

	if err = mkDir(); err != nil {
		return err
	}

	//init grpc listener
	GrpcListener, err = net.Listen("tcp", GRpcListen)
	if err != nil {
		panic(fmt.Sprintf("failed to listen grpc addr:%s error:%s", GRpcListen, err.Error()))
	}

	return nil
}

func mkDir() error {
	var err error
	var per os.FileMode = 0744
	if err = os.MkdirAll(LogPath, per); err != nil {
		return errors.New("log_path " + LogPath + ": " + err.Error())
	}
	if err = os.MkdirAll(TmpPath, per); err != nil {
		return errors.New("tmp_path " + TmpPath + ": " + err.Error())
	}
	return nil
}
