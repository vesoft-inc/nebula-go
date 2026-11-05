// Copyright 2026 vesoft inc. All Rights Reserved.
package decode

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vesoft-inc/nebula-go/v5/pkg/types"
)

func TestStreamDecoderClearsNestedBatchBindings(t *testing.T) {
	prop := &vectorProps{vectorIndex: 7, typ: &columnTypeSchemaBasic{typ: types.ColumnTypeInt64}}
	path := &columnTypeSchemaPath{
		meta:       &pathMetaData{},
		nodeSchema: &columnTypeSchemaElement{graphElementProps: graphElementProps{1: elementProps{1: props{"a": prop}}}},
		edgeSchema: &columnTypeSchemaElement{},
	}
	decoder := &StreamDecoder{schema: &resultSchema{columnTypes: []typeSchema{
		&columnTypeSchemaList{subSchema: path},
	}}}
	decoder.ClearBatch()
	require.Nil(t, path.meta)
	require.Zero(t, prop.vectorIndex)
}
