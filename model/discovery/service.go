package discovery

import (
	"encoding/json"
	"fmt"
	"time"

	pb "github.com/huajiao-tv/gokeeper/pb/go"
)

const (
	// 后台metadata key前缀
	BackendMetadataPrefix = "backend-metadata-"
	//机房权重key,用于service
	BackendMetadataZoneWeight = "backend-metadata-zone_weight"
	//节点权重key，用于instance
	BackendMetadataInstanceWeight = "backend-metadata-instance_weight"

	//实例是否在线，如果该字段为空，表示在线
	BackendMetadataInstanceOnline = "backend-metadata-online"
	//节点在线
	BackendInstanceOnlineYes = "Y"
	//节点被下掉，只有明确该字段为"N"才表示该节点被下掉
	BackendInstanceOnlineNo = "N"

	SchemaHttp = "http"
	SchemaRpc  = "rpc"

	//默认的polls时长
	DefaultPollsInterval = 60 * time.Second
)

// 实例信息，每次启动一个实例，都会注册一个Instance
type Instance struct {
	//实例的唯一标识，类似于keeper中的nodeId
	Id string `json:"id"`
	//服务名称，project.service[.sub_service],例如live.session
	ServiceName string `json:"service_name"`
	//机房，bjcc/bjyt/bjdt
	Zone string `json:"zone"`
	//环境，prod(线上)/pre(预发布)/test(测试)/dev(开发)
	Env string `json:"env"`
	//机器地址
	Hostname string `json:"hostname"`
	//地址 scheme -> address  例如： http -> 127.0.0.1:8080、rpc -> 127.0.0.1:8081
	Addrs map[string]string `json:"addrs"`
	//元组数据，例如weight权重信息
	Metadata MD `json:"metadata"`

	//注册时间
	RegTime int64 `json:"reg_time"`
	//更新时间，目前该字段没有实时更新，不变，值为RegTime
	UpdateTime int64 `json:"update_time"`
}

// 服务，根据ServiceName进行唯一标识
type Service struct {
	//服务名称，project.service[.sub_service],例如live.session
	ServiceName string `json:"service_name"`
	//实例信息，根据zone（机房）进行分组
	Instances map[string]map[string]*Instance `json:"instances"`
	//元组数据，用于扩展
	Metadata MD `json:"metadata"`
	//service更新时间，该字段暂时不维护
	UpdateTime int64 `json:"update_time"`
	//service版本，暂时采用时间戳，精确到纳秒
	Version int64 `json:"version"`
}

//机房权重信息，用于机房流量分配，采用类似于bilibili的两级分配机制
//权重的重新计算在client里边做，只要service有变更就重新计算一次，同属一个机房的实例保存的权重是一样的。
type ZoneWeight struct {
	//源机房
	Src string `json:"src"`
	//目标机房权重
	Dst map[string]uint64 `json:"dst"`
}
type ZoneWeights map[string]ZoneWeight

func EncodeZoneWeight(zws ZoneWeights) (string, error) {
	s, err := json.Marshal(zws)
	return string(s), err
}

func DecodeZoneWeight(s string) (ZoneWeights, error) {
	var zws ZoneWeights
	err := json.Unmarshal([]byte(s), &zws)
	return zws, err
}

//后台人工干预的属性信息
type Property struct {
	//服务名称，project.service[.sub_service],例如live.session
	ServiceName string `json:"service_name"`
	//机房权重
	ZoneWeights ZoneWeights `json:"zone_weights"`
	//节点属性信息  id -> MD
	Attrs map[string]MD `json:"attrs"`
}

func ParseInstance(in *pb.Instance) *Instance {
	if in == nil {
		return nil
	}
	return &Instance{
		Id:          in.Id,
		ServiceName: in.ServiceName,
		Zone:        in.Zone,
		Env:         in.Env,
		Hostname:    in.Hostname,
		Addrs:       in.Addrs,
		Metadata:    in.Metadata,
		RegTime:     in.RegTime,
		UpdateTime:  in.UpdateTime,
	}
}

func FormatInstance(in *Instance) *pb.Instance {
	if in == nil {
		return nil
	}
	return &pb.Instance{
		Id:          in.Id,
		ServiceName: in.ServiceName,
		Zone:        in.Zone,
		Env:         in.Env,
		Hostname:    in.Hostname,
		Addrs:       in.Addrs,
		Metadata:    in.Metadata,
		RegTime:     in.RegTime,
		UpdateTime:  in.UpdateTime,
	}
}

func ParseService(in *pb.Service) *Service {
	if in == nil {
		return nil
	}
	out := &Service{
		ServiceName: in.ServiceName,
		Metadata:    in.Metadata,
		UpdateTime:  in.UpdateTime,
		Version:     in.Version,
		Instances:   map[string]map[string]*Instance{},
	}

	for zone, group := range in.Instances {
		zoneInstances := map[string]*Instance{}
		for id, ins := range group.ZoneGroup {
			zoneInstances[id] = ParseInstance(ins)
		}
		out.Instances[zone] = zoneInstances
	}
	return out
}

//过滤掉被下掉的节点
func FilterOfflineInstance(service *Service) *Service {
	if service == nil {
		return nil
	}

	for _, group := range service.Instances {
		for id, ins := range group {
			if IsInstanceOffline(ins) {
				delete(group, id)
			}
		}
	}

	return service
}

//查看instance是否被下掉
func IsInstanceOffline(instance *Instance) bool {
	return instance.Metadata[BackendMetadataInstanceOnline] == BackendInstanceOnlineNo
}

func FormatService(in *Service) *pb.Service {
	if in == nil {
		return nil
	}
	out := &pb.Service{
		ServiceName: in.ServiceName,
		Metadata:    in.Metadata,
		UpdateTime:  in.UpdateTime,
		Version:     in.Version,
		Instances:   map[string]*pb.ZoneGroup{},
	}
	for zone, zoneInstances := range in.Instances {
		group := &pb.ZoneGroup{
			ZoneGroup: map[string]*pb.Instance{},
		}
		for id, ins := range zoneInstances {
			group.ZoneGroup[id] = FormatInstance(ins)
		}
		out.Instances[zone] = group
	}
	return out
}

func PrintService(action string, service *Service) {
	fmt.Println("action:", action)
	fmt.Println("ServiceName:", service.ServiceName, "Medatada:", service.Metadata,
		"UpdateTime:", time.Unix(service.UpdateTime, 0),
		"Version:", service.Version)
	for zone, zoneInstances := range service.Instances {
		fmt.Println("zone:", zone)
		for id, instance := range zoneInstances {
			fmt.Println("id:", id, "instance:", instance)
		}
	}
}
