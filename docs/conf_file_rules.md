#### 解析方式

gokeeper配置解析在[go-ini](https://github.com/go-ini/ini)的基础之上，做了修改（比如，注释仅'#'有效，';'不作为注释啦），采用键值对的方式

#### 配置规则：

键 类型 = 值
'键'与'类型'之间用空格分隔， 如果不指定'类型'， 则默认为 string 类型
以下是定义变量front_listen为 string 类型，值为 :8088

```
front_listen       string   = :8088
```

#### 目前支持类型

```
bool 
例如：test_mode := true 
配置：test_mode  bool = true 

int
例如：var log_level int = 1 
配置：log_level int = 1

int64

float64

string

[]string   #数组的元素采用','分隔
例如：user_cache_range := []string{"0:0", "1:40000100", "2:40000200"} 
配置：user_cache_range []string = 0:0,1:40000100,2:40000200

[]int

[]int64

[]float64

[]bool

map[string]string  #采用','分隔单个键值对，采用':'分隔键/值
例如：city_province_map := map[string]string{"济南"："山东", "沈阳":"辽宁"}
配置：city_province_map map[string]string = 济南:山东,沈阳:辽宁

map[string]int

map[string]bool

map[int]string

map[int]int

map[int]bool

map[string][]string  #采用';'分隔单个键值对, 采用':'分隔键/值(仅以第一个':'为分隔符，第一个':'前为key，后续的作为value)， 采用','分隔[]string 元素
例如：user_cache_redis_master := map[string][]string{"0":[]string{"10.208.53.13:1527:79ac01c004e298e1"}, "1":[]string{"10.208.53.13:1527:79ac01c004e298e1"}}
user_cache_redis_master map[string][]string = 0:10.208.53.13:1527:79ac01c004e298e1;1:10.208.53.13:1527:79ac01c004e298e1

map[string]struct{} #采用','分隔单个键值对，struct数据为空

map[int]struct{}


time.Duration 
例如： sleep_time := 30 * time.Second 可表示为
sleep_time time.Duration= 30s 


json：配置格式： 变量名 变量类型 json = json字符串 
例如想配置 type Person struct 结构体的json的数据，可以如下配置

pk_person Person json = {"name":"zhangsan","age":20,"cr":{"score":{"english":88,"math":99},"id":2011}}
可生成go代码：
GkPerson Person
然后， 需要在 data 文件夹下添加 Person的结构声明，例如添加 struct.go 文件，里边声明Person 
package data

type Course struct {
    Score map[string]uint64 `json:"score"`
    Id    uint64            `json:"id"`
}

type Person struct {
    Name string `json:"name"`
    Age  uint64 `json:"age"`
    Cr   Course `json:"cr"`
}
```