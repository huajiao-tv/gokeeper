gokeeper 设计
==============

### 设计目标

1. 分布式高可用配置管理
2. 通过agent对管理的组件进行start/restart/stop等操作
3. 收集组件运行状态(内存, cpu等)

### 角色

系统定义了3个角色：
1. keeper - 配置管理中心节点
2. agent  - 可对node进行操作, 需要与被管理的node在同一台机器上
3. node   - 接入系统的用户组件，组件通过引入的client包从keeper订阅需要的配置

三者间的交互：
1. keeper与node间通过rpc通信，请求由node主动发起
2. keeper与agent间通过rpc通信，请求由node主动发起
3. node与agent不直接通信

基本概念：
1. domain     - 域，代表一个产品，每个产品间相互独立
2. component  - 组件，即接入node定义的角色，方便对node进行分类管理
3. gokeeper-cli - 结构体生成工具，可将配置文件按文件名生成结构体
4. data - 即gokeeper-cli生成的结构体
5. client - node实际上是client包的具体应用

### keeper

##### 配置
配置以文本文件的方式存放：

```
├── domain1
│   ├── bjsc
│   │   ├── center.conf
│   │   ├── global.conf
│   ├── zwt
│   │   ├── center.conf
│   │   ├── global.conf
│   ├── center.conf
│   ├── global.conf
```

1. 各目录下允许同名文件，同名文件间合并成一个结构体，如遇key将合并。
2. node订阅的文件如有父级同名文件，将一并订阅。例如如果客户端订阅了/domain1/bjsc/global.conf 那么同时也将订阅 /domain1/global.conf

##### 管理

keeper提供管理http接口

1. 获取domain列表
2. 配置add/update
3. 配置回滚
4. 查看版本记录
5. 查看节点信息

##### 分布式（未实现）

### node

node实际上是client包的具体应用，通过client包订阅需要的配置。node通过rpc与keeper保持通讯，以事件的方式进行数据交互


### agent




