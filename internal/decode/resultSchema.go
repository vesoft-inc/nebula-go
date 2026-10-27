// Copyright 2026 vesoft inc. All Rights Reserved.
package decode

import (
	"time"

	"github.com/vesoft-inc/nebula-go/v5/internal/generated_code/v5.0.0/proto/vector"
)

type resultSchema struct {
	columnNames     []string
	columnTypes     []typeSchema
	columnTypeBytes [][]byte
	context         *decodeContext
}

func newResultSchema(meta *vector.VectorTableMetaData) (*resultSchema, error) {
	schema := &resultSchema{context: &decodeContext{
		timezoneOffset: int64(int32(meta.TimeZoneOffset)) * int64(time.Minute/time.Second),
		graphsSchema:   constructGraphsSchema(meta),
	}}
	if meta.RowType != nil {
		schema.columnNames = meta.RowType.ColumnNames
		for _, typ := range meta.RowType.ColumnTypes {
			parsed, err := newTypeSchema(newBytesReader(typ.ValueType))
			if err != nil {
				return nil, err
			}
			schema.columnTypes = append(schema.columnTypes, parsed)
			schema.columnTypeBytes = append(schema.columnTypeBytes, typ.ValueType)
		}
	}
	return schema, nil
}

func constructGraphsSchema(meta *vector.VectorTableMetaData) graphsSchema {
	if meta.GraphSchema == nil {
		return nil
	}
	gsm := make(graphsSchema)
	for _, g := range meta.GraphSchema {
		gs := graphSchema{
			name:        string(g.GraphName),
			id:          g.GraphId,
			nodesSchema: make(map[int32]*elementSchema),
			edgesSchema: make(map[int32]*elementSchema),
		}
		for _, n := range g.NodeType {
			lables := make([]string, 0, len(n.Label))
			for _, l := range n.Label {
				lables = append(lables, string(l))
			}
			gs.nodesSchema[n.NodeTypeId] = &elementSchema{
				typeName: string(n.NodeTypeName),
				typeId:   n.NodeTypeId,
				labels:   lables,
			}
		}
		for _, e := range g.EdgeType {
			lables := make([]string, 0, len(e.Label))
			for _, l := range e.Label {
				lables = append(lables, string(l))
			}
			gs.edgesSchema[e.EdgeTypeId] = &elementSchema{
				typeName: string(e.EdgeTypeName),
				typeId:   e.EdgeTypeId,
				labels:   lables,
			}
		}
		gsm[g.GraphId] = &gs
	}
	return gsm
}
