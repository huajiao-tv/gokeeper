### update 2015-11-24

1.删除agent启动配置文件，相关参数从keeper获取，需要在项目配置添加agent.conf，内容见agent/agent.sample.conf

2.删除agent.conf中的component_pid_path, 组件pid文件不需要组件写，由agent写入到base_path/tmp/component


### update 2015-11-23

1.调整目录结构

* server代码->server
* 配置文件解析->server/conf
* http接口-> server/api/httpapi

2.http接口调整

* conf/manage 原有orders->operates, operate->opcode
* conf/package -> package/list
* node/manage 不需要keeper参数
* conf/list 返回数据结构有调整，见文档

3.配置调整

* keeper.conf 默认配置去除 log_path,tmp_path,conf_path,conf_backup_path,但是还是可以配置。
* agent.conf 去除 component_base_path

