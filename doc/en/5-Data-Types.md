# Data Types

This guide details all supported data types and their usage.

## Type Mapping

| NebulaGraph Type | Go Type | Null Type |
|-----------------|---------|-----------|
| BOOL | `bool` | `nebula.NullBool` |
| INT8 | `int8` | `nebula.NullInt8` |
| INT16 | `int16` | `nebula.NullInt16` |
| INT32 | `int32` | `nebula.NullInt32` |
| INT64 | `int64` | `nebula.NullInt64` |
| UINT8 | `uint8` | `nebula.NullUInt8` |
| UINT16 | `uint16` | `nebula.NullUInt16` |
| UINT32 | `uint32` | `nebula.NullUInt32` |
| UINT64 | `uint64` | `nebula.NullUInt64` |
| FLOAT | `float32` | `nebula.NullFloat` |
| DOUBLE | `float64` | `nebula.NullDouble` |
| STRING | `string` | `nebula.NullString` |
| DATE | - | `nebula.NullDate` |
| DATETIME | - | `nebula.NullLocalDatetime` |
| TIME | - | `nebula.NullLocalTime` |
| DURATION | - | `nebula.NullDuration` |
| LIST | - | `nebula.NullList` |
| MAP | - | `nebula.NullMap` |
| SET | - | `nebula.NullSet` |
| NODE | - | `nebula.NullNode` |
| EDGE | - | `nebula.NullEdge` |
| PATH | - | `nebula.NullPath` |
| GEOGRAPHY | - | `nebula.NullGeography` |
| VECTOR | - | `nebula.NullEmbeddingVector` |

## Primitive Types

### Numeric Types

```go
var (
    id    nebula.NullInt64
    score nebula.NullDouble
)

resp.Scan(&id, &score)

if id.Valid {
    fmt.Printf("ID: %d\n", id.Data)
}
if score.Valid {
    fmt.Printf("Score: %.2f\n", score.Data)
}
```

### String

```go
var name nebula.NullString
resp.Scan(&name)

if name.Valid {
    fmt.Printf("Name: %s\n", name.Data)
}
```

### Boolean

```go
var isActive nebula.NullBool
resp.Scan(&isActive)

if isActive.Valid {
    fmt.Printf("Active: %v\n", isActive.Data)
}
```

## Temporal Types

### Date

```go
var birthday nebula.NullDate
resp.Scan(&birthday)

if birthday.Valid {
    fmt.Printf("Birthday: %d-%02d-%02d\n",
        birthday.Data.GetYear(),
        birthday.Data.GetMonth(),
        birthday.Data.GetDay())
}
```

### LocalTime

```go
var createdAt nebula.NullLocalTime
resp.Scan(&createdAt)

if createdAt.Valid {
    fmt.Printf("Time: %02d:%02d:%02d.%06d\n",
        createdAt.Data.GetHour(),
        createdAt.Data.GetMinute(),
        createdAt.Data.GetSec(),
        createdAt.Data.GetMicrosec())
}
```

### LocalDatetime

```go
var timestamp nebula.NullLocalDatetime
resp.Scan(&timestamp)

if timestamp.Valid {
    fmt.Printf("Datetime: %d-%02d-%02d %02d:%02d:%02d\n",
        timestamp.Data.GetYear(),
        timestamp.Data.GetMonth(),
        timestamp.Data.GetDay(),
        timestamp.Data.GetHour(),
        timestamp.Data.GetMinute(),
        timestamp.Data.GetSec())
}
```

### Duration

```go
var dur nebula.NullDuration
resp.Scan(&dur)

if dur.Valid {
    fmt.Printf("Duration: %d years, %d months, %d days\n",
        dur.Data.GetYear(),
        dur.Data.GetMonth(),
        dur.Data.GetDay())
    fmt.Printf("Time: %dh %dm %ds %dμs\n",
        dur.Data.GetHour(),
        dur.Data.GetMinute(),
        dur.Data.GetSecond(),
        dur.Data.GetMicrosecond())
}
```

## Graph Types

### Node

```go
var v nebula.NullNode
resp.Scan(&v)

if v.Valid {
    fmt.Printf("Graph: %s\n", v.Data.GetGraph())
    fmt.Printf("Type: %s\n", v.Data.GetType())
    fmt.Printf("ID: %d\n", v.Data.GetId())
    fmt.Printf("Labels: %v\n", v.Data.GetLabels())
    
    // Get properties
    props := v.Data.GetProperties()
    for k, val := range props {
        fmt.Printf("  %s: %s\n", k, val.String())
    }
}
```

### Edge

```go
var e nebula.NullEdge
resp.Scan(&e)

if e.Valid {
    fmt.Printf("Graph: %s\n", e.Data.GetGraph())
    fmt.Printf("Type: %s\n", e.Data.GetType())
    fmt.Printf("SrcID: %d\n", e.Data.GetSrcId())
    fmt.Printf("DstID: %d\n", e.Data.GetDstId())
    fmt.Printf("Rank: %d\n", e.Data.GetRank())
    fmt.Printf("Directed: %v\n", e.Data.IsDirected())
    
    // Get properties
    props := e.Data.GetProperties()
    for k, val := range props {
        fmt.Printf("  %s: %s\n", k, val.String())
    }
}
```

### Path

Path consists of alternating Node and Edge sequence: Node-Edge-Node-Edge-Node...

```go
var path nebula.NullPath
resp.Scan(&path)

if path.Valid {
    elements := path.Data.GetValues()
    for i, elem := range elements {
        // Determine if current element is Node or Edge
        switch elem.GetType() {
        case nebula.ValueTypeNode:
            node, _ := elem.AsNode()
            fmt.Printf("Node[%d]: ID=%d, Labels=%v\n", i, node.GetId(), node.GetLabels())
        case nebula.ValueTypeEdge:
            edge, _ := elem.AsEdge()
            fmt.Printf("Edge[%d]: SrcID=%d, DstID=%d, Type=%s\n", i, edge.GetSrcId(), edge.GetDstId(), edge.GetType())
        }
    }
    
    // Or check by index: even index = Node, odd index = Edge
    for i := 0; i < len(elements); i++ {
        if i%2 == 0 {
            // Even index: Node
            node, _ := elements[i].AsNode()
            fmt.Printf("Node: %s\n", node)
        } else {
            // Odd index: Edge
            edge, _ := elements[i].AsEdge()
            fmt.Printf("Edge: %s\n", edge)
        }
    }
}
```

## Geography Types

### Point

```go
var geo nebula.NullGeography
resp.Scan(&geo)

if geo.Valid {
    if geo.Data.GetShape() == nebula.GeoShapePoint {
        point := geo.Data.GetPoint()
        fmt.Printf("Point: (%f, %f)\n", point.Lng, point.Lat)
    }
}
```

### LineString

```go
if geo.Data.GetShape() == nebula.GeoShapeLineString {
    line := geo.Data.GetLineString()
    for i, pt := range line {
        fmt.Printf("Point %d: (%f, %f)\n", i, pt.Lng, pt.Lat)
    }
}
```

### Polygon

```go
if geo.Data.GetShape() == nebula.GeoShapePolygon {
    polygon := geo.Data.GetPolygon()
    for i, ring := range polygon {
        fmt.Printf("Ring %d:\n", i)
        for j, pt := range ring {
            fmt.Printf("  Point %d: (%f, %f)\n", j, pt.Lng, pt.Lat)
        }
    }
}
```

### Creating from WKT

```go
// Create point
resp, _ := client.Execute("RETURN ST_GeogFromText('POINT(116.3974 39.9093)')")

// Create line
resp, _ := client.Execute("RETURN ST_GeogFromText('LINESTRING(116.3974 39.9093, 116.4074 39.9193)')")

// Create polygon
resp, _ := client.Execute("RETURN ST_GeogFromText('POLYGON((116.3974 39.9093, 116.4074 39.9093, 116.4074 39.9193, 116.3974 39.9193, 116.3974 39.9093))')")
```

## Vector Type

```go
var vec nebula.NullEmbeddingVector
resp.Scan(&vec)

if vec.Valid {
    values := vec.Data.GetValues()
    fmt.Printf("Vector dimension: %d\n", len(values))
    fmt.Printf("First 5 values: %v\n", values[:5])
}
```

## Collection Types

### List

```go
var list nebula.NullList
resp.Scan(&list)

if list.Valid {
    values := list.Data.GetValues()
    for i, v := range values {
        fmt.Printf("[%d]: %s\n", i, v.String())
    }
}
```

### Set

```go
var set nebula.NullSet
resp.Scan(&set)

if set.Valid {
    values := set.Data.GetValues()
    for _, v := range values {
        fmt.Println(v.String())
    }
}
```

### Map

```go
var m nebula.NullMap
resp.Scan(&m)

if m.Valid {
    values := m.Data.GetValues()
    for k, v := range values {
        fmt.Printf("%s: %s\n", k.String(), v.String())
    }
}
```

## Using Value Interface Directly

```go
row, _ := resp.Next()
value, _ := row.GetValueByIndex(0)

// Type checking and conversion
switch value.GetType() {
case nebula.ValueTypeString:
    str, _ := value.AsString()
    fmt.Println("String:", str)
case nebula.ValueTypeInt64:
    num, _ := value.AsInt64()
    fmt.Println("Int64:", num)
case nebula.ValueTypeNode:
    node, _ := value.AsNode()
    fmt.Println("Node:", node)
}
```

## Best Practices

1. **Use NullXXX types** - Safe NULL handling, avoid panic
2. **Use Scan** - Best performance, cleanest code
3. **Check Valid field** - Always check before accessing data
4. **Understand type mapping** - Choose correct Go types
