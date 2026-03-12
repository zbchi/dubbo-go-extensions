# Hystrix 过滤器

[English](./README.md) | 简体中文

基于 [hystrix-go](https://github.com/afex/hystrix-go) 的 dubbo-go 熔断器过滤器。

## 安装

```bash
go get github.com/apache/dubbo-go-extensions/filter/hystrix
```

## 使用方法

### 消费者端 (Consumer Side)

```go
import (
    "context"
    "github.com/afex/hystrix-go/hystrix"
    _ "github.com/apache/dubbo-go-extensions/filter/hystrix"
    "dubbo.apache.org/dubbo-go/v3"
    "dubbo.apache.org/dubbo-go/v3/client"
    "dubbo.apache.org/dubbo-go/v3/registry"
)

func init() {
    // 为服务方法配置 hystrix 命令
    // 资源名称格式: dubbo:consumer:接口名:group:version:方法名
    cmdName := "dubbo:consumer:greet.GreetService:::Greet"

    hystrix.ConfigureCommand(cmdName, hystrix.CommandConfig{
        Timeout:                1000,
        MaxConcurrentRequests:  10,
        RequestVolumeThreshold: 5,
        SleepWindow:            5000,
        ErrorPercentThreshold:  50,
    })
}

func main() {
    ins, _ := dubbo.NewInstance(
        dubbo.WithRegistry(
            registry.WithZookeeper(),
            registry.WithAddress("127.0.0.1:2181"),
        ),
    )
    cli, _ := ins.NewClient()
    svc, _ := greet.NewGreetService(cli, client.WithFilter("hystrix_consumer"))

    resp, err := svc.Greet(context.Background(), &greet.GreetRequest{Name: "test"})
}
```

### 提供者端 (Provider Side)

```go
import (
    _ "github.com/apache/dubbo-go-extensions/filter/hystrix"
    "dubbo.apache.org/dubbo-go/v3/server"
)

func main() {
    srv, _ := server.NewServer(
        server.WithFilter("hystrix_provider"),
    )
    // ... 服务器其余配置
}
```

## 过滤器名称

- **消费者端**: `"hystrix_consumer"`
- **提供者端**: `"hystrix_provider"`

## 资源命名规则

Hystrix 命令命名格式如下：
```
dubbo:{consumer|provider}:{interface}:{group}:{version}:{method}
```

- `{consumer|provider}`: 消费者端或提供者端
- `{interface}`: 完整接口名（例如 `greet.GreetService`）
- `{group}`: 服务分组（未指定时为空，显示为 `::`）
- `{version}`: 服务版本（未指定时为空）
- `{method}`: 方法名

示例：
- `dubbo:consumer:greet.GreetService:::Greet`
- `dubbo:consumer:com.example.UserProvider:::GetUser`
- `dubbo:consumer:com.example.UserService:group:v1.0:CreateUser`
