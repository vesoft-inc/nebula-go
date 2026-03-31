# Connection Management

This guide details how to manage connections using the SDK.

## Creating a Single Client

### Basic Usage

```go
client, err := nebula.NewNebulaClient(
    "127.0.0.1:10025",  // Address (comma-separated for multiple)
    "root",              // Username
    "nebula",            // Password
)
if err != nil {
    // Handle error
}
defer client.Close()
```

### Multiple Addresses (Failover)

```go
// Multiple addresses separated by comma, SDK auto-selects
addresses := "192.168.1.1:10025,192.168.1.2:10025,192.168.1.3:10025"
client, err := nebula.NewNebulaClient(addresses, "root", "nebula")
```

## Client Configuration Options

```go
// Connection timeout (default 3 seconds)
nebula.WithClientConnectTimeout(5 * time.Second)

// Request timeout (default 60 seconds)
nebula.WithClientRequestTimeout(30 * time.Second)

// Custom logger
nebula.WithClientLogger(customLogger)

// Custom auth info
nebula.WithClientAuthInfo(map[string]string{
    "password": "nebula",
    "tenant": "tenant_name",
})
```

## Checking Connection Status

```go
// Ping to check if connection is valid
if err := client.Ping(); err != nil {
    fmt.Println("Connection invalid:", err)
}

// Ping with context
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
if err := client.PingContext(ctx); err != nil {
    fmt.Println("Connection invalid:", err)
}
```

## Getting Connection Info

```go
// Get Session ID
sessionId, err := client.GetSessionId()

// Get server version
version, err := client.GetVersion()
```

## Closing Connection

```go
// Close client connection
if err := client.Close(); err != nil {
    // Handle error
}
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
    // Create client with config
    client, err := nebula.NewNebulaClient(
        "127.0.0.1:10025",
        "root",
        "nebula",
        nebula.WithClientConnectTimeout(5*time.Second),
        nebula.WithClientRequestTimeout(30*time.Second),
    )
    if err != nil {
        panic(err)
    }
    defer client.Close()

    // Check connection
    if err := client.Ping(); err != nil {
        panic(err)
    }

    // Get version info
    version, _ := client.GetVersion()
    fmt.Printf("Connected to NebulaGraph %s\n", version)

    // Execute query
    resp, err := client.Execute("SHOW HOSTS")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Columns: %v\n", resp.Columns())
}
```

## Best Practices

1. **Always use defer Close()** - Ensures resources are released
2. **Set reasonable timeouts** - Configure based on your needs
3. **Handle connection failures** - Use Ping to check health
4. **Multi-address failover** - Recommended for production
