# Connection Pool

This guide details how to configure and use connection pools.

## Why Use Connection Pool

- **Connection Reuse** - Reduces connection creation/destruction overhead
- **Concurrency Support** - Supports multiple concurrent queries
- **Connection Management** - Auto health check and connection recycling
- **Failover** - Multi-address automatic switching

## Creating Connection Pool

### Basic Usage

```go
pool, err := nebula.NewNebulaPool(
    "127.0.0.1:10025",
    "root",
    "nebula",
)
if err != nil {
    panic(err)
}
defer pool.Close()
```

### With Configuration

```go
pool, err := nebula.NewNebulaPool(
    "127.0.0.1:10025",
    "root",
    "nebula",
    nebula.WithPoolConnectTime(10*time.Second),    // Connection timeout
    nebula.WithPoolRequestTimeout(30*time.Second),  // Request timeout
    nebula.WithPoolMaxOpenConns(100),               // Max connections
    nebula.WithPoolMaxIdleConns(10),                // Max idle
    nebula.WithPoolMinOpenConns(5),                 // Min idle
)
```

## Pool Configuration Options

### Connection Count

```go
// Max open connections (default 100)
nebula.WithPoolMaxOpenConns(100)

// Max idle connections (default 5)
nebula.WithPoolMaxIdleConns(10)

// Min idle connections (default 1)
nebula.WithPoolMinOpenConns(5)

// Connection max lifetime (default 30 minutes)
nebula.WithPoolMaxLifetime(30 * time.Minute)

// Max wait time (default infinite)
nebula.WithPoolMaxWait(10 * time.Second)
```

### Timeout Configuration

```go
// Connection timeout (default 3 seconds)
nebula.WithPoolConnectTime(5 * time.Second)

// Request timeout (default 60 seconds)
nebula.WithPoolRequestTimeout(30 * time.Second)

// Ping timeout (default 1 second)
nebula.WithPoolPingTimeout(2 * time.Second)

// Skip ping check
nebula.WithPoolWithOutPing()
```

### Session Configuration

```go
// Set default graph space
nebula.WithPoolGraph("nba")

// Set default Schema
nebula.WithPoolSchema("player")

// Set timezone
nebula.WithPoolTimezone("UTC")

// Set datetime formats
nebula.WithPoolDateFormat("%Y-%m-%d")
nebula.WithPoolLocalDatetimeFormat("%Y-%m-%d %H:%M:%S")

// Set session parameters
nebula.WithPoolParameters(map[string]string{
    "timeout": "30000",
})
```

### Health Check Configuration

```go
// Strict mode: all servers must be healthy
nebula.WithPoolStrictlyServerHealthy(true)

// Heartbeat interval (default 5 seconds)
nebula.WithPoolTickerDuration(10 * time.Second)
```

### Logging Configuration

```go
nebula.WithPoolLogger(nebula.DefaultLogger)
```

### TLS Configuration

```go
nebula.WithPoolTLS("ca.crt", "client.crt", "client.key", false)
```

## Using Connection Pool

### Getting and Returning Connections

```go
// Get client
client, err := pool.GetClient()
if err != nil {
    panic(err)
}

// Execute query
resp, err := client.Execute("MATCH (n) RETURN n LIMIT 10")

// Return connection to pool
if err := pool.PutClient(client); err != nil {
    // Handle error
}
```

### Using Defer for Safety

```go
client, err := pool.GetClient()
if err != nil {
    panic(err)
}
defer pool.PutClient(client)

// Execute query
resp, err := client.Execute("...")
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    nebula "github.com/vesoft-inc/nebula-go/v5"
)

func main() {
    // Create connection pool
    pool, err := nebula.NewNebulaPool(
        "192.168.1.1:10025,192.168.1.2:10025,192.168.1.3:10025",
        "root",
        "nebula",
        nebula.WithPoolMaxOpenConns(100),
        nebula.WithPoolMaxIdleConns(20),
        nebula.WithPoolMinOpenConns(10),
        nebula.WithPoolMaxLifetime(30*time.Minute),
        nebula.WithPoolConnectTime(5*time.Second),
        nebula.WithPoolRequestTimeout(30*time.Second),
        nebula.WithPoolGraph("nba"),
        nebula.WithPoolLogger(nebula.DefaultLogger),
    )
    if err != nil {
        panic(err)
    }
    defer pool.Close()

    // Concurrent queries
    for i := 0; i < 10; i++ {
        go func(id int) {
            client, err := pool.GetClient()
            if err != nil {
                fmt.Printf("Goroutine %d: failed to get client: %v\n", id, err)
                return
            }
            defer pool.PutClient(client)

            resp, err := client.Execute("RETURN 1")
            if err != nil {
                fmt.Printf("Goroutine %d: failed to execute: %v\n", id, err)
                return
            }
            fmt.Printf("Goroutine %d: success\n", id)
        }(i)
    }

    // Keep running
    time.Sleep(5 * time.Second)
}
```

## Multi-Address Failover

```go
// Multiple addresses, comma-separated
addresses := "192.168.1.1:10025,192.168.1.2:10025,192.168.1.3:10025"

pool, err := nebula.NewNebulaPool(
    addresses,
    "root",
    "nebula",
    // Strict mode: all addresses must be available
    nebula.WithPoolStrictlyServerHealthy(true),
)
```

## Monitoring Pool Status

The pool automatically:
- Maintains minimum idle connections
- Cleans up connections exceeding max lifetime
- Cleans up connections exceeding max idle count
- Pings to check connection health

## Best Practices

1. **Set reasonable connection count** - Based on concurrency needs
2. **Keep minimum connections** - Reduce cold start latency
3. **Set connection lifetime** - Avoid stale connections
4. **Use defer PutClient** - Ensure connection return
5. **Enable strict mode** - Recommended for production
