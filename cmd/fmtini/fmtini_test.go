package main

import (
	"bytes"
	// "os"
	"strings"
	"testing"
)

var src = `
# 用户消息槽，记录格式：<product>:<username>:$messageIds。记录按 product 分区，按 username 散列到各个 slot server 实例
db_user_slots         []string = 10.16.57.170:7611
# 消息存储桶，记录格式：$messages:<messageId>。按 messageId 散列
db_buckets             []string = 10.16.57.170:7611

# 存储tag
db_product_tags         []string = 10.16.57.170:7611,10.16.57.170:7611
db_blacklist             []string = 10.16.57.170:7611
area_net_message_addrs     []string = 10.16.57.170:7611
db_strom_counter         []string = 10.16.57.170:7611@default,10.16.57.170:7611@baohe


# 消息存储桶过期时间 单位s 604800(86400*7)
message_expire         map[string]int = desktop:604800,baohe:604801,zhuomian:604800,pchelper:604800,safe:604800,msdk:604800,mse:604800,mgamebox:604800,newsreader:604800,miaopai:604800,360tray:604800,watch:604800,360video:259200,zhuanjia:259200,oem_hw_rongyao:259200,cplive:259200,cloudcontrol:259200,ipcam:259200,qucmsec:86400

#单播和多播离线补偿产品列表
coordinator_unicast_product     map[string]bool = 11,123,13,zhuomian,autoTest2,autoTest,autoTest3,autoTest1,autoTest0,80001,autoTest4,autoTest5,120,121,122,124
coordinator_broadcast_product     map[string]bool = 11,13,123,zhuomian,autoTest2,autoTest,autoTest3,autoTest1,autoTest0,80001,autoTest4,autoTest5,120,121,122,124

#离线补偿全局开关false-关；true-开
coordinator_on_off                 bool = true

gorpc_listen     = 10.16.57.170:7321
gorpc_inner = 10.16.57.170:7321
admin_listen = 10.16.57.170:27321

[10.16.57.170:17301]
admin_listen     = 10.16.57.170:17301
abandoned_users []string = 10.16.57.170:7611

admin_listen     = 10.16.57.170:17301
abandoned_users []string = 10.16.57.170:7611
`
var dst = `
# 用户消息槽，记录格式：<product>:<username>:$messageIds。记录按 product 分区，按 username 散列到各个 slot server 实例
db_user_slots []string = 10.16.57.170:7611
# 消息存储桶，记录格式：$messages:<messageId>。按 messageId 散列
db_buckets    []string = 10.16.57.170:7611

# 存储tag
db_product_tags        []string = 10.16.57.170:7611,10.16.57.170:7611
db_blacklist           []string = 10.16.57.170:7611
area_net_message_addrs []string = 10.16.57.170:7611
db_strom_counter       []string = 10.16.57.170:7611@default,10.16.57.170:7611@baohe


# 消息存储桶过期时间 单位s 604800(86400*7)
message_expire map[string]int = desktop:604800,baohe:604801,zhuomian:604800,pchelper:604800,safe:604800,msdk:604800,mse:604800,mgamebox:604800,newsreader:604800,miaopai:604800,360tray:604800,watch:604800,360video:259200,zhuanjia:259200,oem_hw_rongyao:259200,cplive:259200,cloudcontrol:259200,ipcam:259200,qucmsec:86400

#单播和多播离线补偿产品列表
coordinator_unicast_product   map[string]bool = 11,123,13,zhuomian,autoTest2,autoTest,autoTest3,autoTest1,autoTest0,80001,autoTest4,autoTest5,120,121,122,124
coordinator_broadcast_product map[string]bool = 11,13,123,zhuomian,autoTest2,autoTest,autoTest3,autoTest1,autoTest0,80001,autoTest4,autoTest5,120,121,122,124

#离线补偿全局开关false-关；true-开
coordinator_on_off bool = true

gorpc_listen = 10.16.57.170:7321
gorpc_inner  = 10.16.57.170:7321
admin_listen = 10.16.57.170:27321

[10.16.57.170:17301]
admin_listen             = 10.16.57.170:17301
abandoned_users []string = 10.16.57.170:7611

admin_listen             = 10.16.57.170:17301
abandoned_users []string = 10.16.57.170:7611
`

func TestFmtIni(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := format(strings.NewReader(src), buf); err != nil {
		t.Error(err)
	}
	if dst != buf.String() {
		t.Fatal("fomat fail")
	}
	t.Log("format success")
}
