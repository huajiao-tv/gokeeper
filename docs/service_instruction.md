## Use service discovery

### registry service

```go
import(
	...
	
	"github.com/huajiao-tv/gokeeper/client/discovery"
)

...

instance := discovery.NewInstance(discovery.GenRandomId(), "demo.test.com", map[string]string{discoverry.SchemaHttp: "127.0.0.1:17000"})
instance.Id = "test_id_1"

client := discovery.New(
  //gokeeper server address（grpc address）
  "127.0.0.1:7001",
  //registry service
  discovery.WithRegistry(instance),
  discovery.WithRegistryTTL(60*time.Second),
	//schedule strategy，default：random
  discovery.WithScheduler(map[string]schedule.Scheduler{
    "demo.test.com": schedule.NewRoundRobinScheduler(),
  }),
)
```

### discovery service

```go
import(
	...
	
	"github.com/huajiao-tv/gokeeper/client/discovery"
)

...


//start gokeeper client
client := discoverry.New(
  //gokeeper server address（grpc address）
  "127.0.0.1:7001",
  //subscribe a group of service
	discovery.WithDiscovery("example_client1", []string{"demo.test.com"}),
)

...

//get service address
addr, err := client.GetServiceAddr("demo.test.com", discovery.SchemaHttp)
if err != nil{
  ...
}
```

