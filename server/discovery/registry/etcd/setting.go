package etcd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	dm "github.com/huajiao-tv/gokeeper/model/discovery"
)

const (
	defaultTimeout  = 5 * time.Second  //默认超时时长
	defaultLeaseTTL = 30 * time.Second //默认租约时间

	serviceRootPath  = "/discovery"
	infoTypeInstance = "instance"
	infoTypeProperty = "property"

	//存储etcd leaseId，仅用于etcd，不对外暴露
	etcdMetadataLeaseId = "etcd-metadata-lease_id"
)

var (
	ErrInvalidPath = errors.New("path is invalid")
)

func getRootPath() string {
	return serviceRootPath
}

//获取服务的节点路径
//这里不要采用filepath.Join，etcd中的数据存储路径并不根据client的系统类型而变化
func getServicePath(service string) string {
	return fmt.Sprintf("%s/%s/", serviceRootPath, service)
}

//获取服务的节点路径
func getInstancePath(service, id string) string {
	return fmt.Sprintf("%s/%s/%s/%s/", serviceRootPath, service, infoTypeInstance, id)
}

//获取服务属性的路径,后台设置时，typ为backend
func getPropertyPath(service, typ string) string {
	return fmt.Sprintf("%s/%s/%s/%s/", serviceRootPath, service, infoTypeProperty, typ)
}

//解析路径，返回 service、infoType，id
func parseRawPath(path string) (string, string, string, error) {
	list := strings.Split(strings.Trim(path, "/"), "/")
	if len(list) != 4 {
		return "", "", "", ErrInvalidPath
	}
	return list[1], list[2], list[3], nil
}

func encodeInstance(i *dm.Instance) (string, error) {
	b, err := json.Marshal(i)
	return string(b), err
}

func decodeInstance(s string) (*dm.Instance, error) {
	var instance dm.Instance
	err := json.Unmarshal([]byte(s), &instance)
	if err != nil {
		return nil, err
	}
	if instance.Metadata == nil {
		instance.Metadata = dm.MD{}
	}
	return &instance, err
}

func encodeProperty(p *dm.Property) (string, error) {
	b, err := json.Marshal(p)
	return string(b), err
}

func decodeProperty(s string) (*dm.Property, error) {
	var p dm.Property
	err := json.Unmarshal([]byte(s), &p)
	return &p, err
}
