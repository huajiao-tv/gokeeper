gokeeper 开发文档
=================

### keeper

Keeper内部有几个重要的变量
1. KeeperFiles   - 保存所有配置文件信息
2. KeeperDomains - 保存所有已注册的domain和node信息
3. KeeperSaver   - 存储文件版本记录


`
keeperDomains = &DomainBook{}
`
```
type DomainBook struct {
	Domains map[string]*Domain
	lock    sync.RWMutex
}

// Domain is a collection of nodes
type Domain struct {
	Name    string
	Version int
	Nodes   map[string]*model.Node
	lock    sync.RWMutex
}

// Node info
type Node struct {
	ID        string
	Addr      string
	Domain    string
	Component string
	// register datetime
	Rtime int64
	// 当前配置版本
	Version int
	// 订阅信息
	Sections []string
	// 根据订阅信息生成的订阅文件
	SubSections map[string][]Section
	// 根据订阅信息解析订阅文件中kv值
	StructData []StructData
	// node status, 1-online 0-offline
	Status int
	// 机器运行状态信息
	Proc *ProcInfo

	Event chan *Event `json:"-"`
	lock  sync.RWMutex
}
```

`
KeeperFiles = &Files{}
`

```
type Files struct {
	path   string
	info   os.FileInfo
	config *ini.File
	child  []*Files
	lock   sync.RWMutex
}
```

##### 工作流程

初始化：
1. 程序初始化时扫描配置目录下的所有文件到KeeperFiles中，所有对文件的操作都通过KeeperFiles进行
2. 同时将已有的domain信息保存到 KeeperDomains中
3. 找出domain的最新版本号，保存在KeeperDomains中

node注册：TODO


配置更新：
1. 解析node注册的文件和section，找出符合订阅关系的所有文件和section，并按照目录层次排序（父级排前面，方便后续生成结构体时重复key覆盖）
2. 解析文件中的kv，并根据key的类型对v进行类型转换，生成ConfigData结构
3. 根据文件名合并生成的数据，生成[]ConfigStruct结构
4. 将数据返回给node
```
type ConfigData struct {
	Type     string
	Key      string
	RawValue string      // 原始值
	Value    interface{} //处理后的值
}
type StructData struct {
	Name    string // 结构体名
	Version int
	Data    map[string]ConfigData
}
```








