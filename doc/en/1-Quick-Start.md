# Quick Start

This guide helps you get started with NebulaGraph Go SDK quickly.

## Installation

```bash
go get github.com/vesoft-inc/nebula-go/v5@v5.2.0
```

## Basic Usage Example

```go
package main

import (
    "fmt"
    nebula "github.com/vesoft-inc/nebula-go/v5"
)

const (
    address  = "127.0.0.1:10025"
    username = "root"
    password = "nebula"
)

func main() {
    // Create client
    client, err := nebula.NewNebulaClient(address, username, password)
    if err != nil {
        panic(err)
    }
    defer client.Close()

    // Execute query
    resp, err := client.Execute("RETURN 1 + 1 AS result")
    if err != nil {
        panic(err)
    }

    // Process results
    for resp.HasNext() {
        row, err := resp.Next()
        if err != nil {
            panic(err)
        }
        value, _ := row.GetValueByIndex(0)
        fmt.Println(value)
    }
}
```

## Using Connection Pool

```go
package main

import (
    "fmt"
    "time"
    nebula "github.com/vesoft-inc/nebula-go/v5"
)

const (
    address  = "127.0.0.1:10025"
    username = "root"
    password = "nebula"
)

func main() {
    // Create connection pool
    pool, err := nebula.NewNebulaPool(
        address, username, password,
        nebula.WithPoolConnectTime(10*time.Second),
        nebula.WithPoolRequestTimeout(10*time.Second),
        nebula.WithPoolMaxOpenConns(100),
    )
    if err != nil {
        panic(err)
    }
    defer pool.Close()

    // Get client from pool
    client, err := pool.GetClient()
    if err != nil {
        panic(err)
    }
    defer pool.PutClient(client)

    // Execute query
    resp, err := client.Execute("RETURN 1 + 1 AS result")
    if err != nil {
        panic(err)
    }

    for resp.HasNext() {
        row, _ := resp.Next()
        value, _ := row.GetValueByIndex(0)
        fmt.Println(value)
    }
}
```

## Core Concepts

| Concept | Description |
|---------|-------------|
| `Client` | Single client connection for executing queries |
| `Pool` | Connection pool managing multiple clients |
| `Result` | Query result set containing data and summary |
| `Row` | A single row in the result set |
| `Summary` | Query execution statistics |

## Next Steps

- [Connection Management](2-Connection-Management.md)
- [Query Execution](3-Query-Execution.md)
- [Result Processing](4-Result-Processing.md)
