// Copyright 2026 vesoft inc. All Rights Reserved.
package decode

import (
	"github.com/vesoft-inc/nebula-go/v5/internal/generated_code/v5.0.0/proto/vector"
	"github.com/vesoft-inc/nebula-go/v5/pkg/types"
)

// StreamDecoder owns one stream's parsed types. A single consumer binds and reads one
// batch at a time; retained rows contain decoded values, not a cursor into this schema.
// Type schemas also hold vector-layout bindings, which must not survive a batch.
type StreamDecoder struct {
	schema      *resultSchema
	columnTypes []types.ColumnType
}

func NewStreamDecoder(meta *vector.VectorTableMetaData) (*StreamDecoder, error) {
	schema, err := newResultSchema(meta)
	if err != nil {
		return nil, err
	}
	d := &StreamDecoder{schema: schema}
	for _, typ := range schema.columnTypes {
		d.columnTypes = append(d.columnTypes, typ.getType())
	}
	return d, nil
}

func (d *StreamDecoder) BindBatch(batch *vector.VectorBatch, rows uint64) *ResultTable {
	d.ClearBatch()
	return newResultTable(d.schema, []*vector.VectorBatch{batch}, rows)
}

// ClearBatch releases path vector references as soon as the consumer exhausts a batch.
// Close must not call it concurrently with decoding.
func (d *StreamDecoder) ClearBatch() {
	for _, typ := range d.schema.columnTypes {
		resetBatchBindings(typ)
	}
}

func (d *StreamDecoder) ColumnTypes() []types.ColumnType {
	result := make([]types.ColumnType, len(d.columnTypes))
	copy(result, d.columnTypes)
	return result
}

// Keep all batch-dependent fields here when adding a typeSchema implementation.
// Parsing builds an acyclic type tree; shared subtrees may be reset more than once.
func resetBatchBindings(typ typeSchema) {
	switch t := typ.(type) {
	case *columnTypeSchemaBasic:
	case *columnTypeSchemaList:
		resetBatchBindings(t.subSchema)
	case *columnTypeSchemaSet:
		resetBatchBindings(t.subSchema)
	case *columnTypeSchemaVector:
		resetBatchBindings(t.subSchema)
	case *columnTypeSchemaMap:
		resetBatchBindings(t.keySchema)
		resetBatchBindings(t.valueSchema)
	case *columnTypeSchemaRecord:
		for _, field := range t.propSchemas {
			resetBatchBindings(field)
		}
	case *columnTypeSchemaElement:
		for _, graph := range t.graphElementProps {
			for _, props := range graph {
				for _, prop := range props {
					prop.vectorIndex = 0
					resetBatchBindings(prop.typ)
				}
			}
		}
	case *columnTypeSchemaPath:
		t.meta = nil
		resetBatchBindings(t.nodeSchema)
		resetBatchBindings(t.edgeSchema)
	}
}
