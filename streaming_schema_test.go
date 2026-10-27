// Copyright 2026 vesoft inc. All Rights Reserved.
package nebula_ng

import (
	"context"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vesoft-inc/nebula-go/v5/internal/generated_code/v5.0.0/proto/graph"
	"github.com/vesoft-inc/nebula-go/v5/internal/generated_code/v5.0.0/proto/vector"
	"github.com/vesoft-inc/nebula-go/v5/pkg/types"
)

func streamNodeType() []byte {
	data := []byte{1}
	data = binary.LittleEndian.AppendUint32(data, 1)
	data = binary.LittleEndian.AppendUint32(data, 1)
	data = binary.LittleEndian.AppendUint16(data, 1)
	data = binary.LittleEndian.AppendUint32(data, 2)
	for _, name := range []byte{'a', 'b'} {
		data = binary.LittleEndian.AppendUint16(data, 1)
		data = append(data, name, 0x0b)
	}
	return data
}

func streamGraphBatch(id int64, pair uint16) *graph.StreamingExecuteResponse {
	names := []byte{'a', 'b'}
	values := []int64{id, 100 + id}
	if pair == 1 {
		names[0], names[1] = names[1], names[0]
		values[0], values[1] = values[1], values[0]
	}
	special := binary.LittleEndian.AppendUint32(nil, 2)
	for _, name := range names {
		special = binary.LittleEndian.AppendUint16(special, 1)
		special = append(special, name)
	}
	special = binary.LittleEndian.AppendUint32(special, 1)
	special = binary.LittleEndian.AppendUint32(special, 1)
	special = binary.LittleEndian.AppendUint16(special, 1)
	special = binary.LittleEndian.AppendUint32(special, 2)
	special = binary.LittleEndian.AppendUint32(special, 0)
	special = binary.LittleEndian.AppendUint32(special, 1)
	data := binary.LittleEndian.AppendUint64(nil, uint64(1<<48)+uint64(id))
	data = binary.LittleEndian.AppendUint32(data, 1)
	data = binary.LittleEndian.AppendUint32(data, 0)
	node := &vector.NestedVector{
		CommonMetaData: &vector.VectorCommonMetaData{NumRecords: 1, VectorContentType: 0x102},
		VectorData:     data, SpecialMetaData: special,
		NestedVectors: []*vector.NestedVector{streamBatch(values[0]).GetBatch().Vectors[0], streamBatch(values[1]).GetBatch().Vectors[0]},
	}
	pathData := binary.LittleEndian.AppendUint32(nil, 1)
	pathData = binary.LittleEndian.AppendUint16(pathData, pair)
	pathData = binary.LittleEndian.AppendUint16(pathData, pair)
	pathData = binary.LittleEndian.AppendUint64(pathData, 0)
	pathMeta := binary.LittleEndian.AppendUint32(nil, 1)
	pathMeta = binary.LittleEndian.AppendUint32(pathMeta, 1)
	pathMeta = binary.LittleEndian.AppendUint16(pathMeta, 1)
	pathMeta = binary.LittleEndian.AppendUint16(pathMeta, pair)
	pathMeta = binary.LittleEndian.AppendUint32(pathMeta, 0)
	path := &vector.NestedVector{
		CommonMetaData: &vector.VectorCommonMetaData{NumRecords: 1, VectorContentType: 0x102},
		VectorData:     pathData, SpecialMetaData: pathMeta,
		NestedVectors: []*vector.NestedVector{node, streamBatch(int64(uint64(1)<<63 | uint64(pair)<<32)).GetBatch().Vectors[0]},
	}
	list := &vector.NestedVector{
		CommonMetaData: &vector.VectorCommonMetaData{NumRecords: 1, VectorContentType: 0x102},
		VectorData:     binary.LittleEndian.AppendUint32(binary.LittleEndian.AppendUint32(nil, 0), 1),
		NestedVectors:  []*vector.NestedVector{path},
	}
	return &graph.StreamingExecuteResponse{Response: &graph.StreamingExecuteResponse_Batch{Batch: &vector.VectorBatch{Vectors: []*vector.NestedVector{node, path, list}}}}
}

func TestStreamingGraphSchemaAcrossBatches(t *testing.T) {
	nodeType := streamNodeType()
	pathType := append(binary.LittleEndian.AppendUint32([]byte{0x12}, 1), nodeType...)
	header := &graph.StreamingExecuteResponse{Response: &graph.StreamingExecuteResponse_Metadata{Metadata: &vector.VectorTableMetaData{
		IsLittleEndian: true,
		RowType:        &vector.RowType{NumColumns: 3, ColumnNames: []string{"n", "p", "paths"}, ColumnTypes: []*vector.ValueType{{ValueType: nodeType}, {ValueType: pathType}, {ValueType: append([]byte{0x11}, pathType...)}}},
		GraphSchema:    []*vector.PropertyGraphSchema{{GraphId: 1, GraphName: []byte("test_graph"), NodeType: []*vector.NodeType{{NodeTypeId: 1, NodeTypeName: []byte("person"), Label: [][]byte{[]byte("Person")}}}}},
	}}}
	cn := streamTestConnection(t, func(_ *graph.ExecuteRequest, s graph.GraphService_StreamingExecuteServer) error {
		for _, m := range []*graph.StreamingExecuteResponse{header, streamGraphBatch(1, 0), streamGraphBatch(2, 1), streamFinal("00000")} {
			if err := s.Send(m); err != nil {
				return err
			}
		}
		return nil
	})
	rs, err := cn.StreamingExecute(context.Background(), "query")
	require.NoError(t, err)
	defer rs.Close()
	checkTypes := func() {
		require.Equal(t, []string{"n", "p", "paths"}, rs.Columns())
		got := rs.ColumnTypes()
		require.Equal(t, []types.ColumnType{types.ColumnTypeNode, types.ColumnTypePath, types.ColumnTypeList}, got)
		got[0] = types.ColumnTypeUnknown
		require.Equal(t, types.ColumnTypeNode, rs.ColumnTypes()[0])
	}
	checkRow := func(row types.Row, id int64) {
		for _, name := range []string{"n", "p", "paths"} {
			value, err := row.GetValueByName(name)
			require.NoError(t, err)
			if name == "paths" {
				list, err := value.AsList()
				require.NoError(t, err)
				require.Len(t, list.GetValues(), 1)
				value = list.GetValues()[0]
			}
			if name != "n" {
				path, err := value.AsPath()
				require.NoError(t, err)
				require.Len(t, path.GetValues(), 1)
				value = path.GetValues()[0]
			}
			node, err := value.AsNode()
			require.NoError(t, err)
			a, err := node.GetProperties()["a"].AsInt64()
			require.NoError(t, err)
			b, err := node.GetProperties()["b"].AsInt64()
			require.NoError(t, err)
			require.EqualValues(t, id, a)
			require.EqualValues(t, 100+id, b)
		}
	}
	checkTypes()
	first, err := rs.Next()
	require.NoError(t, err)
	checkRow(first, 1)
	second, err := rs.Next()
	require.NoError(t, err)
	checkRow(second, 2)
	checkTypes()
	require.False(t, rs.HasNext())
	require.NoError(t, rs.Err())
	checkRow(first, 1)
	checkRow(second, 2)
	checkTypes()
}

func TestStreamingEmptySchemaResults(t *testing.T) {
	for _, mode := range []string{"empty table", "zero columns", "only final"} {
		t.Run(mode, func(t *testing.T) {
			cn := streamTestConnection(t, func(_ *graph.ExecuteRequest, s graph.GraphService_StreamingExecuteServer) error {
				if mode != "only final" {
					h := streamHeader()
					if mode == "zero columns" {
						h.GetMetadata().RowType = &vector.RowType{}
					}
					if err := s.Send(h); err != nil {
						return err
					}
				}
				return s.Send(streamFinal("00000"))
			})
			rs, err := cn.StreamingExecute(context.Background(), "query")
			require.NoError(t, err)
			defer rs.Close()
			require.False(t, rs.HasNext())
			require.NoError(t, rs.Err())
			if mode == "empty table" {
				require.Equal(t, []types.ColumnType{types.ColumnTypeInt64}, rs.ColumnTypes())
			} else if mode == "zero columns" {
				require.Equal(t, []types.ColumnType{}, rs.ColumnTypes())
			} else {
				require.Nil(t, rs.ColumnTypes())
			}
		})
	}
}
