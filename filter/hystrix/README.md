# Hystrix Filter

Circuit breaker filter for dubbo-go using [hystrix-go](https://github.com/afex/hystrix-go).

## Installation

```bash
go get github.com/apache/dubbo-go-extensions/filter/hystrix
```

## Usage

### Consumer Side

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
    // Configure hystrix command for the service method
    // Resource name format: dubbo:consumer:InterfaceName:group:version:Method
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

### Provider Side

```go
import (
    _ "github.com/apache/dubbo-go-extensions/filter/hystrix"
    "dubbo.apache.org/dubbo-go/v3/server"
)

func main() {
    srv, _ := server.NewServer(
        server.WithFilter("hystrix_provider"),
    )
    // ... rest of server setup
}
```

## Filter Keys

- **Consumer**: `"hystrix_consumer"`
- **Provider**: `"hystrix_provider"`

## Resource Naming

Hystrix commands are named as:
```
dubbo:{consumer|provider}:{interface}:{group}:{version}:{method}
```

- `{consumer|provider}`: Consumer or provider side
- `{interface}`: Full interface name (e.g., `greet.GreetService`)
- `{group}`: Service group (empty if not specified, shown as `::`)
- `{version}`: Service version (empty if not specified)
- `{method}`: Method name

Examples:
- `dubbo:consumer:greet.GreetService:::Greet`
- `dubbo:consumer:com.example.UserProvider:::GetUser`
- `dubbo:consumer:com.example.UserService:group:v1.0:CreateUser`