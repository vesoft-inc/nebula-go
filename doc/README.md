# NebulaGraph Go SDK Documentation

This is the documentation repository for the NebulaGraph Go client.

## Documentation Structure

```
doc/
├── README.md           # Documentation overview
├── zh/                 # Chinese documentation
│   ├── 1-快速开始.md
│   ├── 2-连接管理.md
│   ├── 3-查询执行.md
│   ├── 4-结果处理.md
│   ├── 5-数据类型.md
│   ├── 6-连接池.md
│   ├── 7-TLS配置.md
│   └── 8-日志配置.md
└── en/                 # English documentation
    ├── 1-Quick-Start.md
    ├── 2-Connection-Management.md
    ├── 3-Query-Execution.md
    ├── 4-Result-Processing.md
    ├── 5-Data-Types.md
    ├── 6-Connection-Pool.md
    ├── 7-TLS-Configuration.md
    └── 8-Logging.md
```

## Features

| Feature | Description |
|---------|-------------|
| **Connection Management** | Single client and connection pool modes |
| **Query Execution** | Sync/async execution with context support |
| **Result Processing** | Row iteration, Scan, NULL handling |
| **Data Types** | Full support for all NebulaGraph data types |
| **Connection Pool** | Configurable pool with health checks |
| **TLS Security** | TLS/SSL encrypted connections |
| **Logging** | Configurable logging interface |

## Quick Links

- [中文文档 (Chinese)](zh/)
- [Quick Start (English)](en/1-Quick-Start.md)
- [Connection Pool](en/6-Connection-Pool.md)
- [TLS Configuration](en/7-TLS-Configuration.md)
