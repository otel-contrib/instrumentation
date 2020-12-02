# Redis client for Golang

## Installation

go-redis supports 2 last Go versions and requires a Go version with
[modules](https://github.com/golang/go/wiki/Modules) support. So make sure to initialize a Go module:

```shell
go mod init github.com/my/repo
go get -u github.com/otel-contrib/instrumentation
```

## Quick start

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/otel-contrib/instrumentation/github.com/go-redis/redis"
)

func main() {
    c := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "", // no password set
        DB:       0,  // use default DB
    })

    if err := c.Set(context.Background(), "key", "value", time.Minute).Err(); err != nil {
        log.Fatal(err)
    }
}
```
