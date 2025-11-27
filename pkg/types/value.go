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
package types

import (
	"fmt"
	"time"
)

type Bool bool
type Int8 int8
type Int16 int16
type Int32 int32
type Int64 int64
type UInt8 uint8
type UInt16 uint16
type UInt32 uint32
type UInt64 uint64
type Float float32
type Double float64
type String string
type ValueType uint8

type (
	Value interface {
		String() string
		GetType() ValueType
		IsNull() bool
		AsBool() (Bool, error)
		AsInt8() (Int8, error)
		AsInt16() (Int16, error)
		AsInt32() (Int32, error)
		AsInt64() (Int64, error)
		AsUInt8() (UInt8, error)
		AsUInt16() (UInt16, error)
		AsUInt32() (UInt32, error)
		AsUInt64() (UInt64, error)
		AsFloat() (Float, error)
		AsDouble() (Double, error)
		AsString() (String, error)
		AsList() (List, error)
		AsSet() (Set, error)
		AsRecord() (Record, error)
		AsMap() (Map, error)
		AsDuration() (Duration, error)
		AsLocalTime() (LocalTime, error)
		AsLocalDatetime() (LocalDatetime, error)
		AsDate() (Date, error)
		AsZonedDatetime() (ZonedDatetime, error)
		AsZonedTime() (ZonedTime, error)
		AsNode() (Node, error)
		AsEdge() (Edge, error)
		AsPath() (Path, error)
		AsDecimal() (Decimal, error)
		AsGeography() (Geography, error)
		AsEmbeddingVector() (EmbeddingVector, error)
	}
	List interface {
		String() string
		GetValues() []Value
		Size() int
	}
	Record interface {
		String() string
		GetValues() map[string]Value
	}
	Set interface {
		String() string
		GetValues() []Value
	}
	Map interface {
		String() string
		GetValues() map[Value]Value
	}

	Duration interface {
		IsMonthBased() bool
		String() string
		GetYear() int
		GetMonth() int
		GetDay() int
		GetHour() int
		GetMinute() int
		GetSecond() int
		GetMicrosecond() int
	}

	LocalDatetime interface {
		String() string
		Date
		LocalTime
	}

	LocalTime interface {
		String() string
		GetHour() int
		GetMinute() int
		GetSec() int
		GetMicrosec() int
	}
	ZonedDatetime interface {
		LocalDatetime
		TimeZone
		Time() *time.Time
	}
	ZonedTime interface {
		LocalTime
		TimeZone
	}

	TimeZone interface {
		//GetOffset return the time zone offset in seconds
		GetOffset() int
	}

	Date interface {
		String() string
		GetYear() int
		GetMonth() int
		GetDay() int
	}

	Node interface {
		String() string
		GetProperties() map[string]Value
		GetGraph() string
		GetType() string
		GetLabels() []string
		GetId() int
	}
	Edge interface {
		String() string
		GetProperties() map[string]Value
		GetSrcId() int
		GetDstId() int
		GetGraph() string
		GetType() string
		GetLabels() []string
		GetRank() int
		IsDirected() bool
	}
	Path interface {
		String() string
		GetValues() []Value
	}

	Decimal interface {
		String() string
	}

	Geography interface {
		String() string
		GetSRID() int32
		GetShape() GeoShape
		GetPoint() *Point
		GetLineString() LineString
		GetPolygon() Polygon
	}

	EmbeddingVector interface {
		String() string
		GetValues() []float32
	}

	Point struct {
		Lng float64
		Lat float64
	}

	LineString []*Point

	Polygon [][]*Point

	GeoShape int32
)

const (
	GeoShapePoint      GeoShape = 1
	GeoShapeLineString GeoShape = 5
	GeoShapePolygon    GeoShape = 9
)

const (
	ValueTypeNull ValueType = iota
	ValueTypeBool
	ValueTypeInt8
	ValueTypeInt16
	ValueTypeInt32
	ValueTypeInt64
	ValueTypeUInt8
	ValueTypeUInt16
	ValueTypeUInt32
	ValueTypeUInt64
	ValueTypeFloat
	ValueTypeDouble
	ValueTypeString
	ValueTypeList
	ValueTypeRecord
	ValueTypeNode
	ValueTypeEdge
	ValueTypePath
	ValueTypeDuration
	ValueTypeDate
	ValueTypeLocalTime
	ValueTypeLocalDateTime
	ValueTypeZonedTime
	ValueTypeZonedDateTime
	ValueTypeDecimal
	ValueTypeGeography
	ValueTypeEmbeddingVector
	ValueTypeSet
	ValueTypeMap
	ValueUnSupport = 0xff
)

func (vt ValueType) String() string {
	switch vt {
	case ValueTypeNull:
		return "NULL"
	case ValueTypeBool:
		return "BOOL"
	case ValueTypeInt8:
		return "INT8"
	case ValueTypeInt16:
		return "INT16"
	case ValueTypeInt32:
		return "INT32"
	case ValueTypeInt64:
		return "INT64"
	case ValueTypeUInt8:
		return "UINT8"
	case ValueTypeUInt16:
		return "UINT16"
	case ValueTypeUInt32:
		return "UINT32"
	case ValueTypeUInt64:
		return "UINT64"
	case ValueTypeFloat:
		return "FLOAT"
	case ValueTypeDouble:
		return "DOUBLE"
	case ValueTypeString:
		return "STRING"
	case ValueTypeDuration:
		return "DURATION"
	case ValueTypeDate:
		return "DATE"
	case ValueTypeLocalTime:
		return "LOCALTIME"
	case ValueTypeLocalDateTime:
		return "LOCALDATETIME"
	case ValueTypeZonedTime:
		return "ZONEDTIME"
	case ValueTypeZonedDateTime:
		return "ZONEDDATETIME"
	case ValueTypeList:
		return "LIST"
	case ValueTypeRecord:
		return "RECORD"
	case ValueTypeNode:
		return "NODE"
	case ValueTypeEdge:
		return "EDGE"
	case ValueTypePath:
		return "PATH"
	case ValueTypeDecimal:
		return "Decimal"
	case ValueTypeGeography:
		return "GEOGRAPHY"
	case ValueTypeEmbeddingVector:
		return "EMBEDDINGVECTOR"
	}
	return "UNKNOWN"
}

// EmptyValue just for test
type EmptyValue struct{}

var _ Value = &EmptyValue{}

func (nv *EmptyValue) String() string {
	return ""
}

func (nv *EmptyValue) GetType() ValueType {
	return ValueTypeNull
}

func (nv *EmptyValue) IsNull() bool {
	return true
}

func (nv *EmptyValue) AsBool() (Bool, error) {
	return false, nil
}

func (nv *EmptyValue) AsInt8() (Int8, error) {
	return 0, nil
}

func (nv *EmptyValue) AsInt16() (Int16, error) {
	return 0, nil
}

func (nv *EmptyValue) AsInt32() (Int32, error) {
	return 0, nil
}

func (nv *EmptyValue) AsInt64() (Int64, error) {
	return 0, nil
}

func (nv *EmptyValue) AsUInt8() (UInt8, error) {
	return 0, nil
}

func (nv *EmptyValue) AsUInt16() (UInt16, error) {
	return 0, nil
}

func (nv *EmptyValue) AsUInt32() (UInt32, error) {
	return 0, nil
}

func (nv *EmptyValue) AsUInt64() (UInt64, error) {
	return 0, nil
}

func (nv *EmptyValue) AsFloat() (Float, error) {
	return 0, nil
}

func (nv *EmptyValue) AsDouble() (Double, error) {
	return 0, nil
}

func (nv *EmptyValue) AsString() (String, error) {
	return "", nil
}

func (nv *EmptyValue) AsList() (List, error) {
	return nil, nil
}

func (nv *EmptyValue) AsMap() (Map, error) {
	return nil, nil
}

func (nv *EmptyValue) AsRecord() (Record, error) {
	return nil, nil
}
func (nv *EmptyValue) AsSet() (Set, error) {
	return nil, nil
}

func (nv *EmptyValue) AsDuration() (Duration, error) {
	return nil, nil
}

// AsTime() (time.Time, error)
// AsDatetime() (Datetime, error)
func (nv *EmptyValue) AsNode() (Node, error) {
	return nil, nil
}

func (nv *EmptyValue) AsEdge() (Edge, error) {
	return nil, nil
}

func (nv *EmptyValue) AsPath() (Path, error) {
	return nil, nil
}
func (nv *EmptyValue) AsLocalDatetime() (LocalDatetime, error) {
	return nil, nil
}

func (nv *EmptyValue) AsLocalTime() (LocalTime, error) {
	return nil, nil
}

func (nv *EmptyValue) AsZonedTime() (ZonedTime, error) {
	return nil, nil
}
func (nv *EmptyValue) AsZonedDatetime() (ZonedDatetime, error) {
	return nil, nil
}

func (nv *EmptyValue) AsDate() (Date, error) {
	return nil, nil
}

func (nv *EmptyValue) AsDecimal() (Decimal, error) {
	return nil, nil
}

func (nv *EmptyValue) AsGeography() (Geography, error) {
	return nil, nil
}

func (nv *EmptyValue) AsEmbeddingVector() (EmbeddingVector, error) {
	return nil, nil
}

func (s *String) String() string {
	return fmt.Sprintf(`%s`, *s)
}
