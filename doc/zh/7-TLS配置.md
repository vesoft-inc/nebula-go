# TLS 配置

本文档介绍如何配置 TLS/SSL 安全连接。

## 启用 TLS

### 基本用法

```go
client, err := nebula.NewNebulaClient(
    "127.0.0.1:10025",
    "root",
    "nebula",
    nebula.WithClientTLS(
        "ca.crt",        // CA 证书
        "client.crt",    // 客户端证书
        "client.key",    // 客户端私钥
        false,           // 是否跳过证书验证
    ),
)
if err != nil {
    panic(err)
}
defer client.Close()
```

## 证书说明

| 参数 | 说明 |
|------|------|
| `ca.crt` | CA 证书，用于验证服务端证书 |
| `client.crt` | 客户端证书 |
| `client.key` | 客户端私钥 |
| `insecureSkipVerify` | 是否跳过证书验证（生产环境应设为 false） |

## 连接池 TLS 配置

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

## 自定义 TLS 配置

使用 `crypto/tls.Config` 进行高级配置：

```go
import "crypto/tls"

tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS12,
    MaxVersion: tls.VersionTLS13,
    ServerName: "nebula-graph",
    // 自定义证书验证
    VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
        // 自定义验证逻辑
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

## 跳过证书验证（测试环境）

```go
// 仅在测试环境使用
client, err := nebula.NewNebulaClient(
    "127.0.0.1:10025",
    "root",
    "nebula",
    nebula.WithClientTLS(
        "",      // CA 为空
        "",      // 证书为空
        "",      // 密钥为空
        true,    // 跳过验证
    ),
)
```

## 完整示例

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
    
    // 创建带 TLS 的客户端
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
    
    // 检查连接
    if err := client.Ping(); err != nil {
        panic(err.Error())
    }
    
    fmt.Println("TLS connection established successfully")
}
```

## 证书生成（用于测试）

```bash
# 生成 CA
openssl genrsa -out ca.key 2048
openssl req -new -x509 -days 365 -key ca.key -out ca.crt

# 生成客户端证书
openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr
openssl x509 -req -days 365 -in client.csr -CA ca.crt -CAkey ca.key -out client.crt
```

## 最佳实践

1. **生产环境不跳过验证** - `insecureSkipVerify` 仅用于测试
2. **使用 TLS 1.2+** - 确保最低版本要求
3. **正确配置证书** - 使用有效的 CA 签发证书
4. **保护私钥** - 确保 `client.key` 文件权限安全
