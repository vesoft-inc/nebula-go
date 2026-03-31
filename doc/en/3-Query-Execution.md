# Query Execution

This guide details how to execute nGQL queries.

## Basic Queries

### Execute (Sync)

```go
// Execute a single query
resp, err := client.Execute("RETURN 1 + 1 AS result")
if err != nil {
    panic(err)
}
```

### ExecuteContext (With Context)

```go
// Query with cancellation and timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

resp, err := client.ExecuteContext(ctx, "MATCH (n) RETURN n LIMIT 100")
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        fmt.Println("Query timeout")
    } else {
        panic(err)
    }
}
```

## Common Query Examples

### Creating Space and Graph

```go
// Create graph space
_, err := client.Execute("CREATE GRAPH TYPE graph_type IF NOT EXISTS AS GRAPH TYPE {(node_type(id) LABEL player {id INT, name STRING}),(node_type)-[edge_type LABEL follow {followness INT}]->(node_type)}")

// Create graph instance
_, err := client.Execute("CREATE GRAPH nba IF NOT EXISTS OF graph_type")

// Use graph
_, err := client.Execute("USE nba")
```

### Inserting Data

```go
// Insert nodes
_, err := client.Execute(`
    USE nba 
    INSERT NODE node_type ({id:1, name:"Tim"}),({id:2, name:"Jerry"})
`)

// Insert edges
_, err := client.Execute(`
    USE nba 
    INSERT EDGE edge_type ({id:1})-[{followness:90}]->({id:2})
`)
```

### Querying Data

```go
// Simple query
resp, err := client.Execute("USE nba RETURN 1 + 1")

// MATCH query
resp, err := client.Execute("USE nba MATCH (v) RETURN v LIMIT 10")

// Conditional query
resp, err := client.Execute(`
    USE nba 
    MATCH (p:player) 
    WHERE p.name == "Tim" 
    RETURN p
`)
```

## Processing Query Results

```go
resp, err := client.Execute("MATCH (n:player) RETURN n.name, n.age LIMIT 5")
if err != nil {
    panic(err)
}

// Get column names
fmt.Println("Columns:", resp.Columns())

// Iterate rows
for resp.HasNext() {
    row, err := resp.Next()
    if err != nil {
        panic(err)
    }
    
    // Get by index
    name, _ := row.GetValueByIndex(0)
    age, _ := row.GetValueByIndex(1)
    
    fmt.Printf("Name: %s, Age: %s\n", name, age)
}
```

## Using Scan to Populate Structs

```go
// Define struct
type Player struct {
    ID   nebula.NullInt64
    Name nebula.NullString
    Age  nebula.NullInt
}

resp, err := client.Execute("MATCH (p:player) RETURN p.id, p.name, p.age LIMIT 5")
if err != nil {
    panic(err)
}

for resp.HasNext() {
    var player Player
    if err := resp.Scan(&player.ID, &player.Name, &player.Age); err != nil {
        panic(err)
    }
    
    if player.ID.Valid && player.Name.Valid {
        fmt.Printf("ID: %d, Name: %s\n", player.ID.Data, player.Name.Data)
    }
}
```

## Getting Query Summary

```go
resp, err := client.Execute("MATCH (n) RETURN n LIMIT 1000")
if err != nil {
    panic(err)
}

// Get execution summary
summary := resp.Summary()
if summary != nil {
    fmt.Printf("Parse time: %d μs\n", summary.ParseTimeUs())
    fmt.Printf("Execute time: %d μs\n", summary.TotalServerTimeUs())
    
    // Get query stats
    stats := summary.QueryStats()
    if stats != nil {
        fmt.Printf("Affected nodes: %d\n", stats.NumAffectedNodes())
        fmt.Printf("Affected edges: %d\n", stats.NumAffectedEdges())
    }
}
```

## Best Practices

1. **Use context for timeout control** - Avoid blocking queries
2. **Use parameterized queries** - Reduce SQL injection risk
3. **Close results properly** - Ensure all rows are consumed
4. **Use Scan** - Better performance than GetValueByIndex
