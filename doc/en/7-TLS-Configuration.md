# TLS Configuration

This guide explains how to configure TLS/SSL secure connections.

## Enabling TLS

### Basic Usage

```go
client, err := nebula.NewNebulaClient(
    "127.0.0.1:10025",
    "root",
    "nebula",
    nebula.WithClientTLS(
        "ca.crt",        // CA certificate
        "client.crt",    // Client certificate
        "client.key",    // Client private key
        false,           // Skip certificate verification
    ),
)
if err != nil {
    panic(err)
}
defer client.Close()
```

## Certificate Reference

| Parameter | Description |
|-----------|-------------|
| `ca.crt` | CA certificate, verifies server certificate |
| `client.crt` | Client certificate |
| `client.key` | Client private key |
| `insecureSkipVerify` | Skip verification (should be false in production) |

## Pool TLS Configuration

```go
pool, err := nebula.NewNebulaPool(
    "127.0.0.1:10025",
    "root",
    "nebula",
    nebula.WithPoolTLS(
        "ca.crt",
        "client.crt",
        "client.key",
        false,
    ),
)
```

## Custom TLS Configuration

Use `crypto/tls.Config` for advanced configuration:

```go
import "crypto/tls"

tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS12,
    MaxVersion: tls.VersionTLS13,
    ServerName: "nebula-graph",
    // Custom certificate verification
    VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
        // Custom verification logic
        return nil
    },
}

client, err := nebula.NewNebulaClient(
    "127.0.0.1:10025",
    "root",
    "nebula",
    nebula.WithClientTLSConfig(tlsConfig),
)
```

## Skip Certificate Verification (Testing Only)

```go
// Only for testing environments
client, err := nebula.NewNebulaClient(
    "127.0.0.1:10025",
    "root",
    "nebula",
    nebula.WithClientTLS(
        "",      // Empty CA
        "",      // Empty cert
        "",      // Empty key
        true,    // Skip verification
    ),
)
```

## Complete Example

```go
package main

import (
    "fmt"
    "time"
    
    nebula "github.com/vesoft-inc/nebula-go/v5"
)

const (
    address  = "127.0.0.1"
    port     = 16720
    username = "root"
    password = "nebula"
)

func main() {
    fmt.Println("TLS example starts ...")
    addresses := fmt.Sprintf("%s:%d", address, port)
    
    // Create client with TLS
    client, err := nebula.NewNebulaClient(
        addresses,
        username,
        password,
        nebula.WithClientTLS(
            "ca.crt",
            "client.crt",
            "client.key",
            false,
        ),
        nebula.WithClientConnectTimeout(5*time.Second),
    )
    if err != nil {
        panic(err.Error())
    }
    defer client.Close()
    
    // Check connection
    if err := client.Ping(); err != nil {
        panic(err.Error())
    }
    
    fmt.Println("TLS connection established successfully")
}
```

## Certificate Generation (For Testing)

```bash
# Generate CA
openssl genrsa -out ca.key 2048
openssl req -new -x509 -days 365 -key ca.key -out ca.crt

# Generate client certificate
openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr
openssl x509 -req -days 365 -in client.csr -CA ca.crt -CAkey ca.key -out client.crt
```

## Best Practices

1. **Don't skip verification in production** - `insecureSkipVerify` only for testing
2. **Use TLS 1.2+** - Ensure minimum version requirements
3. **Configure certificates correctly** - Use valid CA-signed certificates
4. **Protect private keys** - Ensure `client.key` file permissions are secure
