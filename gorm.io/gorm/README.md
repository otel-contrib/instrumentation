# GORM

The fantastic ORM library for Golang, aims to be developer friendly.

## Overview

* Full-Featured ORM
* Associations (Has One, Has Many, Belongs To, Many To Many, Polymorphism, Single-table inheritance)
* Hooks (Before/After Create/Save/Update/Delete/Find)
* Eager loading with `Preload`, `Joins`
* Transactions, Nested Transactions, Save Point, RollbackTo to Saved Point
* Context, Prepared Statment Mode, DryRun Mode
* Batch Insert, FindInBatches, Find To Map
* SQL Builder, Upsert, Locking, Optimizer/Index/Comment Hints, NamedArg, Search/Update/Create with SQL Expr
* Composite Primary Key
* Auto Migrations
* Logger
* Extendable, flexible plugin API: Database Resolver (Multiple Databases, Read/Write Splitting) / Prometheus…
* Every feature comes with tests
* Developer Friendly

## Getting Started

```shell
go mod init github.com/my/repo
go get -u github.com/otel-contrib/instrumentation
```

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/otel-contrib/instrumentation/gorm.io/gorm"
)

type user struct {
    gorm.Model
    Name string
}

func main() {
    dsn := "root:password@tcp(127.0.0.1:3306)/test?parseTime=true&loc=Asia%2FShanghai"
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

    if err := db.AutoMigrate(&user{}); err != nil {
        log.Fatal(err)
    }
}
```
