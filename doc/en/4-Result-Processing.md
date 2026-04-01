# Result Processing

This guide details how to process query result sets.

## Result Set Structure

```
Result
├── Columns()     -> []string       // Column names
├── ColumnTypes() -> []ColumnType   // Column types
├── RowSize()     -> int            // Number of rows
├── HasNext()     -> bool           // Has more rows
├── Next()        -> Row            // Get next row
├── Scan(...any)  -> error          // Scan to variables
└── Summary()     -> Summary        // Execution summary
```

## Iterating Result Sets

### Basic Iteration

```go
resp, err := client.Execute("MATCH (n:player) RETURN n.name, n.age LIMIT 5")

// Check for data
fmt.Println("Columns:", resp.Columns())

for resp.HasNext() {
    row, err := resp.Next()
    if err != nil {
        panic(err)
    }
    
    // Get all values
    values := row.Values()
    for _, v := range values {
        fmt.Printf("%s ", v.String())
    }
    fmt.Println()
}
```

### Get Value by Index

```go
row, _ := resp.Next()

// By column index (0-based)
name, err := row.GetValueByIndex(0)
age, err := row.GetValueByIndex(1)
```

### Get Value by Name

```go
row, _ := resp.Next()

// By column name
name, err := row.GetValueByName("n.name")
age, err := row.GetValueByName("n.age")
```

## Using Scan to Populate Structs

Scan is the most efficient way to get results, directly mapping columns to Go structs.

### Basic Usage

```go
var (
    name nebula.NullString
    age  nebula.NullInt64
)

for resp.HasNext() {
    if err := resp.Scan(&name, &age); err != nil {
        panic(err)
    }
    
    if name.Valid {
        fmt.Printf("Name: %s, Age: %d\n", name.Data, age.Data)
    }
}
```

### Using Structs (Defined Inside for Loop)

```go
type Player struct {
    ID   nebula.NullInt64
    Name nebula.NullString
    Age  nebula.NullInt
}

for resp.HasNext() {
    // Create new struct instance each iteration
    var p Player
    if err := resp.Scan(&p.ID, &p.Name, &p.Age); err != nil {
        panic(err)
    }
    
    // Each iteration has independent p object
    if p.ID.Valid && p.Name.Valid {
        fmt.Printf("ID: %d, Name: %s, Age: %d\n", 
            p.ID.Data, p.Name.Data, p.Age.Data)
    }
}
```

### Using Structs (Defined Outside for Loop)

```go
type Player struct {
    ID   nebula.NullInt64
    Name nebula.NullString
    Age  nebula.NullInt
}

// Defined outside, created once
var p Player

for resp.HasNext() {
    // Reuse the same struct instance, values overwritten each Scan
    if err := resp.Scan(&p.ID, &p.Name, &p.Age); err != nil {
        panic(err)
    }
    
    // Caution: if using p asynchronously (e.g., goroutine) inside the loop,
    // must define inside loop, otherwise values will be overwritten
    if p.ID.Valid && p.Name.Valid {
        fmt.Printf("ID: %d, Name: %s, Age: %d\n", 
            p.ID.Data, p.Name.Data, p.Age.Data)
    }
}
```

### Difference Between the Two Approaches

| Approach | Pros | Cons | Use Case |
|----------|------|------|----------|
| Inside for | Each iteration independent, safe | Memory allocation each iteration | Recommended, safer |
| Outside for | Less memory allocation | Values can be overwritten | Only when no async ops in loop |

```go
// ❌ Wrong: outside for + async use
var p Player
for resp.HasNext() {
    resp.Scan(&p.ID, &p.Name, &p.Age)
    // Wrong: goroutine captures same address, value will be overwritten
    go func() {
        fmt.Println(p.Name)  // May output wrong value
    }()
}

// ✅ Correct: inside for
for resp.HasNext() {
    var p Player
    resp.Scan(&p.ID, &p.Name, &p.Age)
    // Correct: new instance each iteration
    go func() {
        fmt.Println(p.Name)  // Value is correct
    }()
}
```

### Supported Types

| Go Type | Nebula Type |
|---------|-------------|
| `*int` | INT8, INT16, INT32, INT64 |
| `*uint` | UINT8, UINT16, UINT32, UINT64 |
| `*float32` | FLOAT |
| `*float64` | FLOAT, DOUBLE |
| `*string` | STRING |
| `*bool` | BOOL |
| `*nebula.NullXXX` | Corresponding type, supports NULL |

## Handling NULL Values

NebulaGraph NULL values are wrapped with `NullXXX` types:

```go
// NullString - String that may be NULL
var name nebula.NullString
if err := resp.Scan(&name); err != nil {
    panic(err)
}
if name.Valid {
    fmt.Println("Name:", name.Data)
} else {
    fmt.Println("Name is NULL")
}

// Other NULL types
var (
    id        nebula.NullInt64
    score     nebula.NullDouble
    createdAt nebula.NullDate
    isActive  nebula.NullBool
)
```

## Getting Execution Summary

```go
summary := resp.Summary()
if summary == nil {
    return
}

// Time stats (microseconds)
fmt.Printf("Parse time:    %d μs\n", summary.ParseTimeUs())
fmt.Printf("Build time:    %d μs\n", summary.BuildTimeUs())
fmt.Printf("Optimize time: %d μs\n", summary.OptimizeTimeUs())
fmt.Printf("Execute time:  %d μs\n", summary.TotalServerTimeUs())

// Query stats
stats := summary.QueryStats()
if stats != nil {
    fmt.Printf("Affected nodes: %d\n", stats.NumAffectedNodes())
    fmt.Printf("Affected edges: %d\n", stats.NumAffectedEdges())
}

// Warnings count
fmt.Printf("Warnings: %d\n", summary.NumWarnings())
```

## Best Practices

1. **Prefer Scan** - Better performance than GetValueByIndex/Name
2. **Use NullXXX for NULL** - Avoid type mismatch errors
3. **Check Valid field** - Always check before accessing data
4. **Consume results** - Result must be fully iterated or explicitly closed
