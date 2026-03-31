# Logging

This guide explains how to configure SDK logging.

## Default Loggers

SDK provides ready-to-use loggers:

```go
// Use default logger (outputs to stdout)
log := nebula.DefaultLogger

// Use empty logger (silent mode)
log := nebula.EmptyLogger
```

## Logger Interface

```go
type Logger interface {
    Info(msg string)
    Warn(msg string)
    Error(msg string)
}
```

## Custom Logger

Implement `Logger` interface to create custom logger:

```go
import (
    "fmt"
    "log"
    "runtime"
)

type MyLogger struct {
    prefix string
}

func (l *MyLogger) Info(msg string) {
    fmt.Printf("[INFO] [%s] %s\n", l.prefix, msg)
}

func (l *MyLogger) Warn(msg string) {
    fmt.Printf("[WARN] [%s] %s\n", l.prefix, msg)
}

func (l *MyLogger) Error(msg string) {
    fmt.Printf("[ERROR] [%s] %s\n", l.prefix, msg)
}
```

## Using with Client

```go
client, err := nebula.NewNebulaClient(
    "127.0.0.1:10025",
    "root",
    "nebula",
    nebula.WithClientLogger(&MyLogger{prefix: "nebula-client"}),
)
```

## Using with Pool

```go
pool, err := nebula.NewNebulaPool(
    "127.0.0.1:10025",
    "root",
    "nebula",
    nebula.WithPoolLogger(&MyLogger{prefix: "nebula-pool"}),
)
```

## Integrating Logging Frameworks

### Integrating logrus

```go
import "github.com/sirupsen/logrus"

type LogrusLogger struct {
    log *logrus.Logger
}

func (l *LogrusLogger) Info(msg string) {
    l.log.Info(msg)
}

func (l *LogrusLogger) Warn(msg string) {
    l.log.Warn(msg)
}

func (l *LogrusLogger) Error(msg string) {
    l.log.Error(msg)
}

pool, err := nebula.NewNebulaPool(
    "127.0.0.1:10025",
    "root",
    "nebula",
    nebula.WithPoolLogger(&LogrusLogger{log: logrus.New()}),
)
```

### Integrating zap

```go
import "go.uber.org/zap"

type ZapLogger struct {
    sugar *zap.SugaredLogger
}

func (l *ZapLogger) Info(msg string) {
    l.sugar.Info(msg)
}

func (l *ZapLogger) Warn(msg string) {
    l.sugar.Warn(msg)
}

func (l *ZapLogger) Error(msg string) {
    l.sugar.Error(msg)
}

// Usage
logger, _ := zap.NewProduction()
defer logger.Close()

pool, err := nebula.NewNebulaPool(
    "127.0.0.1:10025",
    "root",
    "nebula",
    nebula.WithPoolLogger(&ZapLogger{sugar: logger.Sugar()}),
)
```

## Complete Example

```go
package main

import (
    "fmt"
    
    nebula "github.com/vesoft-inc/nebula-go/v5"
)

// CustomLogger implements nebula Logger interface
type CustomLogger struct {
    prefix string
}

func (l *CustomLogger) Info(msg string) {
    fmt.Printf("[INFO] [%s] %s\n", l.prefix, msg)
}

func (l *CustomLogger) Warn(msg string) {
    fmt.Printf("[WARN] [%s] %s\n", l.prefix, msg)
}

func (l *CustomLogger) Error(msg string) {
    fmt.Printf("[ERROR] [%s] %s\n", l.prefix, msg)
}

func main() {
    // Use custom logger
    logger := &CustomLogger{prefix: "myapp"}
    
    pool, err := nebula.NewNebulaPool(
        "127.0.0.1:10025",
        "root",
        "nebula",
        nebula.WithPoolLogger(logger),
    )
    if err != nil {
        logger.Error(fmt.Sprintf("Failed to create pool: %v", err))
        return
    }
    defer pool.Close()

    client, err := pool.GetClient()
    if err != nil {
        logger.Error(fmt.Sprintf("Failed to get client: %v", err))
        return
    }
    defer pool.PutClient(client)

    resp, err := client.Execute("RETURN 1")
    if err != nil {
        logger.Error(fmt.Sprintf("Query failed: %v", err))
        return
    }
    
    logger.Info("Query executed successfully")
    fmt.Printf("Columns: %v\n", resp.Columns())
}
```

## Silent Mode

For production or testing, use empty logger to silence logs:

```go
pool, err := nebula.NewNebulaPool(
    "127.0.0.1:10025",
    "root",
    "nebula",
    nebula.WithPoolLogger(nebula.EmptyLogger),
)
```

## Best Practices

1. **Use professional logging libraries in production** - logrus, zap, zerolog
2. **Add prefixes** - Easier log tracing
3. **Set log levels** - Implement custom filtering
4. **Async logging** - Avoid log I/O impacting performance
