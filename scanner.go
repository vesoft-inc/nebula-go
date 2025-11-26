// Copyright 2025 vesoft inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// 
//     http://www.apache.org/licenses/LICENSE-2.0
// 
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package nebula_ng

import (
	"fmt"

	"github.com/vesoft-inc/nebula-go/v5/internal/decode"
	"github.com/vesoft-inc/nebula-go/v5/internal/internal_error"
	"github.com/vesoft-inc/nebula-go/v5/pkg/types"
)

type scanner interface {
	scan(value types.Value) error
}

type nullable[T any] struct {
	Data  T
	Valid bool
	value *decode.NebulaValue
}

type (
	NullString struct {
		nullable[types.String]
	}
	NullBool struct {
		nullable[types.Bool]
	}
	NullInt struct {
		nullable[types.Int64]
	}
	NullInt8 struct {
		nullable[types.Int8]
	}
	NullInt16 struct {
		nullable[types.Int16]
	}
	NullInt32 struct {
		nullable[types.Int32]
	}
	NullInt64 struct {
		nullable[types.Int64]
	}
	NullUInt struct {
		nullable[types.UInt64]
	}
	NullUInt8 struct {
		nullable[types.UInt8]
	}
	NullUInt16 struct {
		nullable[types.UInt16]
	}
	NullUInt32 struct {
		nullable[types.UInt32]
	}
	NullUInt64 struct {
		nullable[types.UInt64]
	}
	NullFloat struct {
		nullable[types.Float]
	}
	NullDouble struct {
		nullable[types.Double]
	}
	NullList struct {
		nullable[types.List]
	}
	NullRecord struct {
		nullable[types.Record]
	}
	NullDuration struct {
		nullable[types.Duration]
	}
	NullLocalTime struct {
		nullable[types.LocalTime]
	}
	NullLocalDatetime struct {
		nullable[types.LocalDatetime]
	}
	NullDate struct {
		nullable[types.Date]
	}
	NullZonedDatetime struct {
		nullable[types.ZonedDatetime]
	}
	NullZonedTime struct {
		nullable[types.ZonedTime]
	}
	NullNode struct {
		nullable[types.Node]
	}
	NullEdge struct {
		nullable[types.Edge]
	}
	NullPath struct {
		nullable[types.Path]
	}
	NullGeography struct {
		nullable[types.Geography]
	}
	NullSet struct {
		nullable[types.Set]
	}
	NullMap struct {
		nullable[types.Map]
	}
	NullEmbeddingVector struct {
		nullable[types.EmbeddingVector]
	}
	NullValue struct {
		nullable[types.Value]
	}
)

func (n *nullable[T]) scanBasic(value types.Value, fn func(types.Value) (T, error)) error {
	if value == nil || value.IsNull() {
		n.Valid = false
		return nil
	}
	d, err := fn(value)
	if err != nil {
		return err
	}
	n.Valid = true
	n.Data = d
	return nil
}

func (n *nullable[T]) scanComposite(value types.Value, fn func() decode.Valuer) error {
	if value == nil || value.IsNull() {
		n.Valid = false
		return nil
	}
	if n.value == nil {
		n.value = &decode.NebulaValue{}
	}
	dst := n.value
	if dst.Data == nil {
		dst.Data = fn()
	}
	src := value.(*decode.NebulaValue)
	decode.DeepCopyValue(src, dst)
	n.Data = n.value.Data.(T)
	n.Valid = true
	return nil
}

// used for testing
func (n *nullable[T]) getData() any {
	return n.Data
}

// used for testing
func (n *nullable[T]) isValid() bool {
	return n.Valid
}

func (n *NullString) scan(value types.Value) error {
	return n.scanBasic(value, func(value types.Value) (types.String, error) {
		return value.AsString()
	})
}

func (n *NullBool) scan(value types.Value) error {
	return n.scanBasic(value, func(value types.Value) (types.Bool, error) {
		return value.AsBool()
	})
}

func (n *NullInt) scan(value types.Value) error {
	return n.scanBasic(value, func(value types.Value) (types.Int64, error) {
		var (
			i   types.Int64
			err error
		)
		switch value.GetType() {
		case types.ValueTypeInt8:
			var d types.Int8
			d, err = value.AsInt8()
			i = types.Int64(d)
		case types.ValueTypeInt16:
			var d types.Int16
			d, err = value.AsInt16()
			i = types.Int64(d)
		case types.ValueTypeInt32:
			var d types.Int32
			d, err = value.AsInt32()
			i = types.Int64(d)
		case types.ValueTypeInt64:
			i, err = value.AsInt64()
		default:
			return 0, internal_error.ErrType(fmt.Sprintf("value type not match"))
		}
		if err != nil {
			return 0, err
		}
		return i, nil
	})
}

func (n *NullInt8) scan(value types.Value) error {
	return n.scanBasic(value, func(value types.Value) (types.Int8, error) {
		return value.AsInt8()
	})
}

func (n *NullInt16) scan(value types.Value) error {
	return n.scanBasic(value, func(value types.Value) (types.Int16, error) {
		return value.AsInt16()
	})
}

func (n *NullInt32) scan(value types.Value) error {
	return n.scanBasic(value, func(value types.Value) (types.Int32, error) {
		return value.AsInt32()
	})
}

func (n *NullInt64) scan(value types.Value) error {
	return n.scanBasic(value, func(value types.Value) (types.Int64, error) {
		return value.AsInt64()
	})
}

func (n *NullUInt) scan(value types.Value) error {
	return n.scanBasic(value, func(value types.Value) (types.UInt64, error) {
		var (
			i   types.UInt64
			err error
		)
		switch value.GetType() {
		case types.ValueTypeInt8:
			var d types.UInt8
			d, err = value.AsUInt8()
			i = types.UInt64(d)
		case types.ValueTypeInt16:
			var d types.UInt16
			d, err = value.AsUInt16()
			i = types.UInt64(d)
		case types.ValueTypeInt32:
			var d types.UInt32
			d, err = value.AsUInt32()
			i = types.UInt64(d)
		case types.ValueTypeInt64:
			i, err = value.AsUInt64()
		default:
			return 0, internal_error.ErrType(fmt.Sprintf("value type not match"))
		}
		if err != nil {
			return 0, err
		}
		return i, nil
	})
}

func (n *NullUInt8) scan(value types.Value) error {
	return n.scanBasic(value, func(value types.Value) (types.UInt8, error) {
		return value.AsUInt8()
	})
}

func (n *NullUInt16) scan(value types.Value) error {
	return n.scanBasic(value, func(value types.Value) (types.UInt16, error) {
		return value.AsUInt16()
	})
}

func (n *NullUInt32) scan(value types.Value) error {
	return n.scanBasic(value, func(value types.Value) (types.UInt32, error) {
		return value.AsUInt32()
	})
}

func (n *NullUInt64) scan(value types.Value) error {
	return n.scanBasic(value, func(value types.Value) (types.UInt64, error) {
		return value.AsUInt64()
	})
}

func (n *NullFloat) scan(value types.Value) error {
	return n.scanBasic(value, func(value types.Value) (types.Float, error) {
		return value.AsFloat()
	})
}
func (n *NullDouble) scan(value types.Value) error {
	return n.scanBasic(value, func(value types.Value) (types.Double, error) {
		var (
			d   types.Double
			err error
		)
		switch value.GetType() {
		case types.ValueTypeFloat:
			var f types.Float
			f, err = value.AsFloat()
			d = types.Double(f)
		case types.ValueTypeDouble:
			d, err = value.AsDouble()
		default:
			return 0, internal_error.ErrType(fmt.Sprintf("value type not match"))
		}
		if err != nil {
			return 0, err
		}
		return d, nil
	})
}

func (n *NullList) scan(value types.Value) error {
	return n.scanComposite(value, func() decode.Valuer {
		return &decode.NebulaList{}
	})
}

func (n *NullRecord) scan(value types.Value) error {
	return n.scanComposite(value, func() decode.Valuer {
		return &decode.NebulaRecord{}
	})
}

func (n *NullSet) scan(value types.Value) error {
	return n.scanComposite(value, func() decode.Valuer {
		return &decode.NebulaSet{}
	})
}

func (n *NullMap) scan(value types.Value) error {
	return n.scanComposite(value, func() decode.Valuer {
		return &decode.NebulaMap{}
	})
}

func (n *NullDuration) scan(value types.Value) error {
	return n.scanComposite(value, func() decode.Valuer {
		return &decode.NebulaDuration{}
	})
}

func (n *NullLocalTime) scan(value types.Value) error {
	return n.scanComposite(value, func() decode.Valuer {
		return &decode.NebulaLocalTime{}
	})
}

func (n *NullLocalDatetime) scan(value types.Value) error {
	return n.scanComposite(value, func() decode.Valuer {
		return &decode.NebulaLocalDatetime{}
	})
}

func (n *NullDate) scan(value types.Value) error {
	return n.scanComposite(value, func() decode.Valuer {
		return &decode.NebulaDate{}
	})
}

func (n *NullZonedDatetime) scan(value types.Value) error {
	return n.scanComposite(value, func() decode.Valuer {
		return &decode.NebulaZonedDatetime{}
	})
}

func (n *NullZonedTime) scan(value types.Value) error {
	return n.scanComposite(value, func() decode.Valuer {
		return &decode.NebulaZonedTime{}
	})
}

func (n *NullNode) scan(value types.Value) error {
	return n.scanComposite(value, func() decode.Valuer {
		return &decode.NebulaNode{}
	})
}

func (n *NullEdge) scan(value types.Value) error {
	return n.scanComposite(value, func() decode.Valuer {
		return &decode.NebulaEdge{}
	})
}

func (n *NullPath) scan(value types.Value) error {
	return n.scanComposite(value, func() decode.Valuer {
		return &decode.NebulaPath{}
	})
}

func (n *NullGeography) scan(value types.Value) error {
	return n.scanComposite(value, func() decode.Valuer {
		return &decode.NebulaGeography{}
	})
}

func (n *NullEmbeddingVector) scan(value types.Value) error {
	return n.scanComposite(value, func() decode.Valuer {
		return &decode.NebulaEmbeddingVector{}
	})
}

func (n *NullValue) scan(value types.Value) error {
	if value == nil || value.IsNull() {
		n.Valid = false
		return nil
	}
	if n.Data == nil {
		n.Data = &decode.NebulaValue{}
	}
	dst := n.Data.(*decode.NebulaValue)
	src := value.(*decode.NebulaValue)
	decode.DeepCopyValue(src, dst)
	n.Valid = true
	return nil
}
