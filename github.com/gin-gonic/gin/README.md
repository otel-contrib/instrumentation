# Gin Web Framework

Gin is a web framework written in Go (Golang), based on [`gin`](https://github.com/gin-gonic/gin/).

## Installation

To install Gin package, you need to install Go and set your Go workspace first.

1. The first need [Go](https://golang.org/) installed (**version 1.11+ is required**), then you can use the below Go
   command to install Gin.

    ```sh
    go get -u github.com/otel-contrib/instrumentation/github.com/gin-gonic/gin
    ```

2. Import it in your code:

    ```go
    import "github.com/otel-contrib/instrumentation/github.com/gin-gonic/gin"
    ```

3. (Optional) Import `net/http`. This is required for example if using constants such as `http.StatusOK`.

    ```go
    import "net/http"
    ```

## Quick start

```sh
# assume the following codes in example.go file
cat example.go
```

```go
package main

import "github.com/otel-contrib/instrumentation/github.com/gin-gonic/gin"

func main() {
    r := gin.Default()
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
        })
    })
    r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
```

```sh
# run example.go and visit 0.0.0.0:8080/ping (for windows "localhost:8080/ping") on browser
go run example.go
```
