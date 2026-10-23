# Streaming queries

`types.StreamingClient` executes a statement through the `StreamingExecute` RPC and returns a forward-only result. The server must implement the current streaming protocol. Released graphd versions declare this RPC but do not handle it, so a call to them waits instead of failing with `UNIMPLEMENTED`. The unary `Execute` remains available.

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

The call returns after the first response message. If that message is the header, column information is already available; if it is the final message, the call returns a result without a table, or the error. The next batch is received only after the current batch has been read. The result still implements `types.Result`, so existing row-reading functions accept it. `RowSize()` returns `-1`. The final status and statistics are available only after a successful final message followed by a normal end of stream.

A result is consumed by one goroutine: row iteration and column access must not be called concurrently, the same as unary results. `Close()` may run concurrently with consumption.

Check `Err()` after iteration. If the stream fails midway, the rows already processed are incomplete. `Close()` may be called more than once, and from another goroutine to interrupt a blocked read; it only cancels this RPC. Until the result is fully read or closed, the client rejects other queries and pings, and the pool rejects returning the client.

The deadline of `ctx` covers the whole iteration, including the time spent processing rows. After `Close()`, the next query may return `SESSION_IS_BUSY` until the server finishes killing the statement.
