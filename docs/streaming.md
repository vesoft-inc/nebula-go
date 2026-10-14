# 流式查询

当前源码新增 `types.StreamingClient`。服务端需要实现当前流式协议。已发布的旧 graphd 虽声明了该 RPC，但没有处理它，不能用 `UNIMPLEMENTED` 判断支持情况。现有 unary `Execute` 继续可用。

```go
sc := client.(types.StreamingClient)
result, err := sc.StreamingExecute(ctx, "FOR i IN range(1,5000) RETURN i")
if err != nil {
    return err
}
defer result.Close()
for result.HasNext() {
    row, err := result.Next()
    if err != nil {
        return err
    }
    consume(row)
}
return result.Err()
```

调用读到第一条响应消息后返回。第一条为表头时列信息已经可用；第一条为终帧时返回无表结果或错误。当前批次读取完才接收下一批。结果仍实现 `types.Result`，可交给现有读行函数。`RowSize()` 返回 `-1`；最终成功状态和统计需要读到成功终帧及正常 EOF 后才能取得。

结果由一个 goroutine 消费，逐行读取和列信息访问不支持并发调用；与 unary 结果的消费方式一致。`Close()` 可以与消费并发执行。

遍历后必须检查 `Err()`。中途失败时，已经处理的行不完整。`Close()` 可以重复调用，也可以从另一 goroutine 中断阻塞读取；关闭只取消本次 RPC。本地在结果读完或关闭之前拒绝该 client 的其他查询和 ping，连接池也拒绝归还仍有活跃流的 client。

`ctx` 的 deadline 覆盖完整遍历。关闭后，服务端 kill 尚未完成时，下次查询可能返回 `SESSION_IS_BUSY`。
