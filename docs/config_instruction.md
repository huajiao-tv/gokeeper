## Use config manage

### Concepts

- **node**: Identification of client to connect gokeeper service（usually use IP address as node）
- **domain**: Identification of a group of config（usually one project matches one domain）
- **file**:a child group config in a domain（define in a *.conf file）
- **section**:a child group of a file（different sections could have same key that has different values,define with "[]",eg:[127.0.0.1:80],default value:DEFAULT）
- **key**:a config key matches a config value
- **gokeeper-cli**: use to create golang struct from conf file（[conf file example](../example/keeper/data/config/mydomain/test.conf)、[conf file rules](conf_file_rules.md)）


### create config in gokeeper server
There are three ways:
- use the admin api:
```
curl "127.0.0.1:17000/add/file?domain=testDomain&file=testFile.conf&conf=test_key%20string%20%3D%20test_value&note=test"
```
  (needs urlEncode)
- init default server:
   Gokeeper server will init domains in the directory "/tmp/gokeeper/init/"(if you use docker-compose to start,it have init the domain "mydomain")
- use dashboard backend to add config:
   The dashboard have been started if you start gokeeper with docker-compose,default address:http://127.0.0.1:8000.
### use config in project:
#### Installation

```shell
go get -u github.com/huajiao-tv/gokeeper ...
```

#### Creating your struct

```shell
cd example/mycomponent
mkdir data && ./../../cmd/gokeeper-cli/gokeeper-cli -i ./../keeper/data/config/mydomain
```

#### start gokeeper client

```go
import (
  ...

  gokeeper "github.com/huajiao-tv/gokeeper/client"

  //the output directory of gokeepercli
  "github.com/huajiao-tv/gokeeper/example/mycomponent/data" 
)

...

//your file and sections to use
sections := []string{"test.conf/DEFAULT"}  

//gokeeper.WithGrpc() will use grpc to connect gokeeper server,otherwise use gorpc
client := gokeeper.New(keeperAddr, domain, nodeID, component, sections, nil, gokeeper.WithGrpc())
client.LoadData(data.ObjectsContainer).RegisterCallback(run)
if err := client.Work(); err != nil {
	panic(err)
}

...
```

#### use config

```go
import (
	//the output directory of gokeepercli
	"github.com/huajiao-tv/gokeeper/example/mycomponent/data" 
)

...

// CurrentTest() returns currnt config with a Test（the struct that gokeeper-cli create）
timeout := data.CurrentTest().Timeout

...
```
