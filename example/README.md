## Getting Started With Gokeeper

### Concepts

* **gokeeper**: 
* **agent**: 
* **node**: 
* **domain**: 
* **gokeeper-cli**: 


### Installation 

```
go get -u github.com/huajiao-tv/gokeeper ...
```

**Creating your struct**

```
cd mycomponent
gokeeper-cli -in=../keeper/data/config/mydomain
```

**Use in your project**

```

```



### Start gokeeper

**Start gokeeper server**

```
cd keeper
gokeeper -f keeper.conf
```

**Start agent**

```
cd agent
agent -f agent.conf -d=mydomain -k=127.0.0.1:7000 -n=127.0.0.1
```

### Usage

#### Node manage

start
```
curl "http://127.0.0.1:17000/node/manage" -d "domain=mydomain&nodeid=127.0.0.1:80&component=mycomponent&operate=start"
```

restart
```
curl "http://127.0.0.1:17000/node/manage" -d "domain=mydomain&nodeid=127.0.0.1:80&component=mycomponent&operate=restart"
```

stop
```
curl "http://127.0.0.1:17000/node/manage" -d "domain=mydomain&nodeid=127.0.0.1:80&component=mycomponent&operate=stop"
```

#### Config manage

**Show config**

```
curl "http://127.0.0.1:17000/conf/list" -d "domain=mydomain"

```
**Reload config**
```
curl "http://127.0.0.1:17000/conf/reload" -d "domain=mydomain"
```


**Add key**

```
curl "http://127.0.0.1:17000/conf/manage" -d 'domain=mydomain&note=test&operates=[{"opcode":"add","file":"test.conf","key":"fortest","value":"30","section":"","type":""}]'
```

**Update key**

```
curl "http://127.0.0.1:17000/conf/manage" -d 'domain=mydomain&note=test&operates=[{"opcode":"update","file":"test.conf","key":"fortest","value":"30","section":"","type":""}]'
```

**Batch operates**

```

```

**Package list**

```
curl "http://127.0.0.1:17000/package/list" -d "domain=mydomain"
```

**Rollback**

```
curl "http://127.0.0.1:17000/package/list" -d "domain=mydomain&id=2"
```


### Next steps



