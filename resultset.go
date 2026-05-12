// Copyright 2025 vesoft inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package nebula_ng

import (
	"fmt"
	"io"
	"iter"

	"github.com/vesoft-inc/nebula-go/v5/internal/decode"
	"github.com/vesoft-inc/nebula-go/v5/internal/generated_code/v5.0.0/proto/graph"
	"github.com/vesoft-inc/nebula-go/v5/internal/internal_error"
	"github.com/vesoft-inc/nebula-go/v5/pkg/types"
)

type resultSet struct {
	index      int
	table      *decode.ResultTable
	summary    *graph.Summary
	cursor     []byte
	values     []types.Value // values is used to store the values of the current row
	columnsMap map[string]int
}

type rowData struct {
	resultSet *resultSet
	values    []types.Value
}

func (rs *resultSet) HasNext() bool {
	if rs.table == nil {
		return false
	}
	return rs.index < int(rs.table.NumRecords())
}

func (rs *resultSet) Next() (types.Row, error) {
	if !rs.HasNext() {
		return nil, io.EOF
	}
	vs, err := rs.table.Next()
	if err != nil {
		return nil, err
	}
	row := &rowData{
		resultSet: rs,
	}
	rs.index++
	row.values = vs

	return row, nil
}

func (rs *resultSet) RowSize() int {
	if rs.table == nil {
		return 0
	}
	return int(rs.table.NumRecords())
}

func (rs *resultSet) ColumnTypes() []types.ColumnType {
	if rs.table == nil {
		return nil
	}
	return rs.table.ColumnTypes()
}

func (rs *resultSet) Summary() types.Summary {
	if rs.summary == nil {
		return nil
	}
	return &summary{summary: rs.summary}
}

func (rs *resultSet) Cursor() []byte {
	return rs.cursor
}

func (rs *resultSet) Columns() []string {
	if rs.table == nil {
		return nil
	}
	return rs.table.ColumnNames()
}

func (rs *resultSet) Scan(dsts ...any) error {
	if !rs.HasNext() {
		return io.EOF
	}
	if len(dsts) != len(rs.Columns()) {
		return internal_error.ErrInternal("scanner length not match values length")
	}
	if rs.values == nil {
		rs.values = make([]types.Value, 0, len(dsts))
		for i := 0; i < len(dsts); i++ {
			rs.values = append(rs.values, &decode.NebulaValue{})
		}
	}

	if err := rs.table.Scan(rs.values); err != nil {
		return err
	}
	values := rs.values
	for i, dst := range dsts {
		if err := rs.convertValue(values[i], dst); err != nil {
			return err
		}
	}
	rs.index++
	return nil
}

func (rs *resultSet) All() iter.Seq2[types.Row, error] {
	return func(yield func(types.Row, error) bool) {
		for rs.HasNext() {
			row, err := rs.Next()
			if !yield(row, err) {
				return
			}
			if err != nil {
				return
			}
		}
	}
}

func (rs *resultSet) convertValue(src types.Value, dst any) error {
	switch dst.(type) {
	case *int, *uint, *float32, *float64, *string, *bool:
		return rs.convertBasicValue(src, dst)
	}
	if sn, ok := dst.(scanner); ok {
		return sn.scan(src)
	}
	return internal_error.ErrInternal(fmt.Sprintf("unsupported type %T", dst))
}

func (rs *resultSet) convertBasicValue(src types.Value, dst any) error {
	if src.IsNull() {
		return fmt.Errorf("value is null")
	}
	switch d := dst.(type) {
	case *int:
		switch src.GetType() {
		case types.ValueTypeInt8:
			v, _ := src.AsInt8()
			*d = int(v)
		case types.ValueTypeInt16:
			v, _ := src.AsInt16()
			*d = int(v)
		case types.ValueTypeInt32:
			v, _ := src.AsInt32()
			*d = int(v)
		case types.ValueTypeInt64:
			v, _ := src.AsInt64()
			*d = int(v)
		case types.ValueTypeUInt8:
			v, _ := src.AsUInt8()
			*d = int(v)
		case types.ValueTypeUInt16:
			v, _ := src.AsUInt16()
			*d = int(v)
		case types.ValueTypeUInt32:
			v, _ := src.AsInt32()
			*d = int(v)
		case types.ValueTypeUInt64:
			v, _ := src.AsInt64()
			*d = int(v)
		default:
			return internal_error.ErrInternal("value type not match")
		}
	case *uint:
		switch src.GetType() {
		case types.ValueTypeUInt8:
			v, _ := src.AsUInt8()
			*d = uint(v)
		case types.ValueTypeUInt16:
			v, _ := src.AsUInt16()
			*d = uint(v)
		case types.ValueTypeUInt32:
			v, _ := src.AsInt32()
			*d = uint(v)
		case types.ValueTypeUInt64:
			v, _ := src.AsInt64()
			*d = uint(v)
		default:
			return internal_error.ErrInternal("value type not match")
		}
	case *float32:
		switch src.GetType() {
		case types.ValueTypeFloat:
			v, _ := src.AsFloat()
			*d = float32(v)
		default:
			return internal_error.ErrInternal("value type not match")
		}
	case *float64:
		switch src.GetType() {
		case types.ValueTypeFloat:
			v, _ := src.AsFloat()
			*d = float64(v)
		case types.ValueTypeDouble:
			v, _ := src.AsDouble()
			*d = float64(v)
		default:
			return internal_error.ErrInternal("value type not match")
		}
	case *string:
		switch src.GetType() {
		case types.ValueTypeString:
			v, _ := src.AsString()
			*d = string(v)
		default:
			return internal_error.ErrInternal("value type not match")
		}
	case *bool:
		switch src.GetType() {
		case types.ValueTypeBool:
			v, _ := src.AsBool()
			*d = bool(v)
		default:
			return internal_error.ErrInternal("value type not match")
		}
	}
	return nil
}

func (rd *rowData) Values() []types.Value {
	return rd.values
}

func (rd *rowData) GetValueByName(name string) (types.Value, error) {
	names := rd.resultSet.Columns()
	var index int
	if rd.resultSet.columnsMap == nil {
		rd.resultSet.columnsMap = make(map[string]int)
		for i, n := range names {
			rd.resultSet.columnsMap[string(n)] = i
		}
	}
	if idx, ok := rd.resultSet.columnsMap[name]; ok {
		index = idx
	} else {
		return nil, internal_error.ErrInternal(fmt.Sprintf("column %s not found", name))
	}

	return rd.values[index], nil
}

func (rd *rowData) GetValueByIndex(index int) (types.Value, error) {
	if index < 0 || index >= len(rd.values) {
		return nil, internal_error.ErrInternal("index out of range")
	}
	return rd.values[index], nil
}
