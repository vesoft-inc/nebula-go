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
package decode

import (
	"bytes"
	"fmt"
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vesoft-inc/nebula-go/v5/internal/internal_error"
	"github.com/vesoft-inc/nebula-go/v5/pkg/types"
)

type Valuer interface {
	GetValue()
}
type edgeDirection uint8

const (
	edgeOutGoingDirection edgeDirection = 0
	edgeInComingDirection               = 1
	edgeNoDirection                     = 2
)

var _ types.Value = &NebulaValue{}

type (
	mapValue    map[string]types.Value
	NebulaValue struct {
		Data Valuer
	}
	NebulaBool  bool
	NebulaInt8  int8
	NebulaInt16 int16
	NebulaInt32 int32
	NebulaInt64 int64

	NebulaUint8  uint8
	NebulaUint16 uint16
	NebulaUint32 uint32
	NebulaUint64 uint64
	NebulaFloat  float32
	NebulaDouble float64
	NebulaString string

	NebulaLocalTime struct {
		Hour     int8
		Minute   int8
		Sec      int8
		Microsec int32
	}
	NebulaLocalDatetime struct {
		Year     int16
		Month    int8
		Day      int8
		Hour     int8
		Minute   int8
		Sec      int8
		Microsec int32
	}
	NebulaZonedTime struct {
		Hour     int8
		Minute   int8
		Sec      int8
		Microsec int32
		Offset   int32
	}
	NebulaZonedDatetime struct {
		Year     int16
		Month    int8
		Day      int8
		Hour     int8
		Minute   int8
		Sec      int8
		Microsec int32
		Offset   int32
	}
	NebulaDuration struct {
		MonthBased bool
		Year       int64
		Month      int8
		Day        int32
		Hour       int8
		Minute     int8
		Sec        int8
		Microsec   int32
	}
	NebulaDate struct {
		Year  int16
		Month int8
		Day   int8
	}
	NebulaList struct {
		Values []*NebulaValue
	}
	NebulaRecord struct {
		Values map[string]*NebulaValue
	}
	NebulaNode struct {
		NodeId     int64
		Graph      string
		Type       string
		Labels     []string
		PropNames  []string
		PropValues []*NebulaValue
	}
	NebulaEdge struct {
		SrcId      int64
		DstId      int64
		Direction  edgeDirection
		Graph      string
		Type       string
		Labels     []string
		Rank       int64
		PropNames  []string
		PropValues []*NebulaValue
	}
	NebulaPath struct {
		Values []*NebulaValue
	}
	NebulaDecimal struct {
		Sval string
	}
	NebulaGeography struct {
		SRID       int32
		Shape      types.GeoShape
		Point      *types.Point
		LineString types.LineString
		Polygon    types.Polygon
	}
	NebulaEmbeddingVector struct {
		Values []float32
	}
	NebulaSet struct {
		Values []*NebulaValue
	}
	NebulaMap struct {
		Values map[*NebulaValue]*NebulaValue
	}
)

func (v *NebulaBool) GetValue()            {}
func (v *NebulaInt8) GetValue()            {}
func (v *NebulaInt16) GetValue()           {}
func (v *NebulaInt32) GetValue()           {}
func (v *NebulaInt64) GetValue()           {}
func (v *NebulaUint8) GetValue()           {}
func (v *NebulaUint16) GetValue()          {}
func (v *NebulaUint32) GetValue()          {}
func (v *NebulaUint64) GetValue()          {}
func (v *NebulaFloat) GetValue()           {}
func (v *NebulaDouble) GetValue()          {}
func (v *NebulaString) GetValue()          {}
func (v *NebulaLocalTime) GetValue()       {}
func (v *NebulaLocalDatetime) GetValue()   {}
func (v *NebulaZonedTime) GetValue()       {}
func (v *NebulaZonedDatetime) GetValue()   {}
func (v *NebulaDuration) GetValue()        {}
func (v *NebulaDate) GetValue()            {}
func (v *NebulaList) GetValue()            {}
func (v *NebulaRecord) GetValue()          {}
func (v *NebulaNode) GetValue()            {}
func (v *NebulaEdge) GetValue()            {}
func (v *NebulaPath) GetValue()            {}
func (v *NebulaDecimal) GetValue()         {}
func (v *NebulaGeography) GetValue()       {}
func (v *NebulaEmbeddingVector) GetValue() {}
func (v *NebulaSet) GetValue()             {}
func (v *NebulaMap) GetValue()             {}
func (v *NebulaValue) GetValue()           {}

// formatFloat formats a float64 value to string with proper handling of special cases
func formatFloat(value float64, precision int) string {
	fStr := strconv.FormatFloat(value, 'f', -1, precision)
	if math.IsInf(value, 1) || math.IsInf(value, -1) || math.IsNaN(value) {
		return fStr
	}
	if !strings.Contains(fStr, ".") {
		fStr = fStr + ".0"
	}
	return fStr
}

func (v *NebulaValue) String() string {
	switch v.GetType() {
	case types.ValueTypeNull:
		return "null"
	case types.ValueTypeBool:
		d, _ := v.AsBool()
		return fmt.Sprintf("%t", d)
	case types.ValueTypeInt8:
		d, _ := v.AsInt8()
		return fmt.Sprintf("%d", d)
	case types.ValueTypeInt16:
		d, _ := v.AsInt16()
		return fmt.Sprintf("%d", d)
	case types.ValueTypeInt32:
		d, _ := v.AsInt32()
		return fmt.Sprintf("%d", d)
	case types.ValueTypeInt64:
		d, _ := v.AsInt64()
		return fmt.Sprintf("%d", d)
	case types.ValueTypeUInt8:
		d, _ := v.AsUInt8()
		return fmt.Sprintf("%d", d)
	case types.ValueTypeUInt16:
		d, _ := v.AsUInt16()
		return fmt.Sprintf("%d", d)
	case types.ValueTypeUInt32:
		d, _ := v.AsUInt32()
		return fmt.Sprintf("%d", d)
	case types.ValueTypeUInt64:
		d, _ := v.AsUInt64()
		return fmt.Sprintf("%d", d)
	case types.ValueTypeFloat:
		d, _ := v.AsFloat()
		return formatFloat(float64(d), 32)
	case types.ValueTypeDouble:
		d, _ := v.AsDouble()
		return formatFloat(float64(d), 64)
	case types.ValueTypeString:
		s, _ := v.AsString()
		return s.String()
	case types.ValueTypeDuration:
		d, _ := v.AsDuration()
		return d.String()
	case types.ValueTypeDate:
		d, _ := v.AsDate()
		return d.String()
	case types.ValueTypeLocalDateTime:
		dt, _ := v.AsLocalDatetime()
		return dt.String()
	case types.ValueTypeLocalTime:
		t, _ := v.AsLocalTime()
		return t.String()
	case types.ValueTypeZonedTime:
		t, _ := v.AsZonedTime()
		return t.String()
	case types.ValueTypeZonedDateTime:
		dt, _ := v.AsZonedDatetime()
		return dt.String()
	case types.ValueTypeList:
		l, _ := v.AsList()
		return l.String()
	case types.ValueTypeRecord:
		r, _ := v.AsRecord()
		return r.String()
	case types.ValueTypeNode:
		n, _ := v.AsNode()
		return n.String()
	case types.ValueTypeEdge:
		e, _ := v.AsEdge()
		return e.String()
	case types.ValueTypePath:
		p, _ := v.AsPath()
		return p.String()
	case types.ValueTypeDecimal:
		d, _ := v.AsDecimal()
		return d.String()
	case types.ValueTypeGeography:
		g, _ := v.AsGeography()
		return g.String()
	case types.ValueTypeEmbeddingVector:
		ev, _ := v.AsEmbeddingVector()
		return ev.String()
	case types.ValueTypeSet:
		s, _ := v.AsSet()
		return s.String()
	case types.ValueTypeMap:
		m, _ := v.AsMap()
		return m.String()
	default:
		return fmt.Sprintf("%v", v.Data)
	}
}

func (v *NebulaValue) GetType() types.ValueType {
	if v.Data == nil {
		return types.ValueTypeNull
	}
	switch v.Data.(type) {
	case *NebulaBool:
		return types.ValueTypeBool
	case *NebulaInt8:
		return types.ValueTypeInt8
	case *NebulaInt16:
		return types.ValueTypeInt16
	case *NebulaInt32:
		return types.ValueTypeInt32
	case *NebulaInt64:
		return types.ValueTypeInt64
	case *NebulaUint8:
		return types.ValueTypeUInt8
	case *NebulaUint16:
		return types.ValueTypeUInt16
	case *NebulaUint32:
		return types.ValueTypeUInt32
	case *NebulaUint64:
		return types.ValueTypeUInt64
	case *NebulaFloat:
		return types.ValueTypeFloat
	case *NebulaDouble:
		return types.ValueTypeDouble
	case *NebulaString:
		return types.ValueTypeString
	case *NebulaDuration:
		return types.ValueTypeDuration
	case *NebulaLocalTime:
		return types.ValueTypeLocalTime
	case *NebulaLocalDatetime:
		return types.ValueTypeLocalDateTime
	case *NebulaZonedTime:
		return types.ValueTypeZonedTime
	case *NebulaZonedDatetime:
		return types.ValueTypeZonedDateTime
	case *NebulaDate:
		return types.ValueTypeDate
	case *NebulaList:
		return types.ValueTypeList
	case *NebulaRecord:
		return types.ValueTypeRecord
	case *NebulaNode:
		return types.ValueTypeNode
	case *NebulaEdge:
		return types.ValueTypeEdge
	case *NebulaPath:
		return types.ValueTypePath
	case *NebulaDecimal:
		return types.ValueTypeDecimal
	case *NebulaGeography:
		return types.ValueTypeGeography
	case *NebulaEmbeddingVector:
		return types.ValueTypeEmbeddingVector
	case *NebulaSet:
		return types.ValueTypeSet
	case *NebulaMap:
		return types.ValueTypeMap
	default:
		return types.ValueUnSupport
	}
}

func (v *NebulaValue) IsNull() bool {
	return v.Data == nil
}

func asValue[T Valuer](v *NebulaValue, valueTyp types.ValueType) (T, error) {
	var t T
	if v.GetType() != valueTyp {
		errMsg := fmt.Sprintf("value is not %s, but %s", valueTyp.String(), v.GetType().String())
		return t, internal_error.ErrType(errMsg)
	}
	data, ok := v.Data.(T)
	if !ok {
		errMsg := fmt.Sprintf("value is not %s, but %s", valueTyp.String(), v.GetType().String())
		return t, internal_error.ErrType(errMsg)
	}
	return data, nil
}

func (v *NebulaValue) AsBool() (types.Bool, error) {
	d, err := asValue[*NebulaBool](v, types.ValueTypeBool)
	if err != nil {
		return false, err
	}
	return types.Bool(*d), nil
}

func (v *NebulaValue) AsInt8() (types.Int8, error) {
	d, err := asValue[*NebulaInt8](v, types.ValueTypeInt8)
	if err != nil {
		return 0, err
	}
	return types.Int8(*d), nil
}

func (v *NebulaValue) AsInt16() (types.Int16, error) {
	d, err := asValue[*NebulaInt16](v, types.ValueTypeInt16)
	if err != nil {
		return 0, err
	}
	return types.Int16(*d), nil
}

func (v *NebulaValue) AsInt32() (types.Int32, error) {
	d, err := asValue[*NebulaInt32](v, types.ValueTypeInt32)
	if err != nil {
		return 0, err
	}
	return types.Int32(*d), nil
}

func (v *NebulaValue) AsInt64() (types.Int64, error) {
	d, err := asValue[*NebulaInt64](v, types.ValueTypeInt64)
	if err != nil {
		return 0, err
	}
	return types.Int64(*d), nil
}

func (v *NebulaValue) AsUInt8() (types.UInt8, error) {
	d, err := asValue[*NebulaUint8](v, types.ValueTypeUInt8)
	if err != nil {
		return 0, err
	}
	return types.UInt8(*d), nil
}

func (v *NebulaValue) AsUInt16() (types.UInt16, error) {
	d, err := asValue[*NebulaUint16](v, types.ValueTypeUInt16)
	if err != nil {
		return 0, err
	}
	return types.UInt16(*d), nil
}

func (v *NebulaValue) AsUInt32() (types.UInt32, error) {
	d, err := asValue[*NebulaUint32](v, types.ValueTypeUInt32)
	if err != nil {
		return 0, err
	}
	return types.UInt32(*d), nil
}

func (v *NebulaValue) AsUInt64() (types.UInt64, error) {
	d, err := asValue[*NebulaUint64](v, types.ValueTypeUInt64)
	if err != nil {
		return 0, err
	}
	return types.UInt64(*d), nil
}

func (v *NebulaValue) AsFloat() (types.Float, error) {
	d, err := asValue[*NebulaFloat](v, types.ValueTypeFloat)
	if err != nil {
		return 0, err
	}
	return types.Float(*d), nil
}

func (v *NebulaValue) AsDouble() (types.Double, error) {
	d, err := asValue[*NebulaDouble](v, types.ValueTypeDouble)
	if err != nil {
		return 0, err
	}
	return types.Double(*d), nil
}

func (v *NebulaValue) AsString() (types.String, error) {
	d, err := asValue[*NebulaString](v, types.ValueTypeString)
	if err != nil {
		return "", err
	}
	return types.String(*d), nil
}

func (v *NebulaValue) AsList() (types.List, error) {
	return asValue[*NebulaList](v, types.ValueTypeList)
}

func (v *NebulaValue) AsSet() (types.Set, error) {
	return asValue[*NebulaSet](v, types.ValueTypeSet)
}

func (v *NebulaValue) AsMap() (types.Map, error) {
	return asValue[*NebulaMap](v, types.ValueTypeMap)
}

func (v *NebulaValue) AsRecord() (types.Record, error) {
	return asValue[*NebulaRecord](v, types.ValueTypeRecord)
}

func (v *NebulaValue) AsDuration() (types.Duration, error) {
	return asValue[*NebulaDuration](v, types.ValueTypeDuration)
}

func (v *NebulaValue) AsNode() (types.Node, error) {
	return asValue[*NebulaNode](v, types.ValueTypeNode)
}

func (v *NebulaValue) AsEdge() (types.Edge, error) {
	return asValue[*NebulaEdge](v, types.ValueTypeEdge)
}

func (v *NebulaValue) AsPath() (types.Path, error) {
	return asValue[*NebulaPath](v, types.ValueTypePath)
}

func (v *NebulaValue) AsDecimal() (types.Decimal, error) {
	return asValue[*NebulaDecimal](v, types.ValueTypeDecimal)
}

func (v *NebulaValue) AsGeography() (types.Geography, error) {
	return asValue[*NebulaGeography](v, types.ValueTypeGeography)
}

func (v *NebulaValue) AsLocalDatetime() (types.LocalDatetime, error) {
	return asValue[*NebulaLocalDatetime](v, types.ValueTypeLocalDateTime)
}

func (v *NebulaValue) AsDate() (types.Date, error) {
	return asValue[*NebulaDate](v, types.ValueTypeDate)
}

func (v *NebulaValue) AsLocalTime() (types.LocalTime, error) {
	return asValue[*NebulaLocalTime](v, types.ValueTypeLocalTime)
}

func (v *NebulaValue) AsZonedTime() (types.ZonedTime, error) {
	return asValue[*NebulaZonedTime](v, types.ValueTypeZonedTime)
}

func (v *NebulaValue) AsZonedDatetime() (types.ZonedDatetime, error) {
	return asValue[*NebulaZonedDatetime](v, types.ValueTypeZonedDateTime)
}

func (v *NebulaValue) AsEmbeddingVector() (types.EmbeddingVector, error) {
	return asValue[*NebulaEmbeddingVector](v, types.ValueTypeEmbeddingVector)
}

func (l *NebulaList) String() string {
	valuesStr := make([]string, 0, len(l.Values))
	for _, v := range l.GetValues() {
		valuesStr = append(valuesStr, v.String())
	}
	return fmt.Sprintf("[%s]", strings.Join(valuesStr, ","))
}

func (l *NebulaList) GetValues() []types.Value {
	values := make([]types.Value, 0, len(l.Values))
	for _, v := range l.Values {
		values = append(values, v)
	}
	return values
}

func (l *NebulaList) Size() int {
	return len(l.Values)
}
func (s *NebulaSet) String() string {
	valuesStr := make([]string, 0, len(s.Values))
	for _, v := range s.Values {
		valuesStr = append(valuesStr, v.String())
	}
	return fmt.Sprintf("{%s}", strings.Join(valuesStr, ","))
}

func (s *NebulaSet) GetValues() []types.Value {
	values := make([]types.Value, 0, len(s.Values))
	for _, v := range s.Values {
		values = append(values, v)
	}
	return values
}

func (r *NebulaRecord) String() string {
	mv := mapValue(r.GetValues())
	return fmt.Sprintf("{%s}", mv.string())
}

func (r *NebulaRecord) GetValues() map[string]types.Value {
	values := make(map[string]types.Value)
	for k, v := range r.Values {
		values[k] = v
	}
	return values
}

func (m *NebulaMap) String() string {
	mm := m.GetValues()
	keys := make([]types.Value, 0, len(mm))
	values := make([]types.Value, 0, len(mm))
	for key := range mm {
		keys = append(keys, key)
		values = append(values, mm[key])
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].String() < keys[j].String()
	})
	kvStr := make([]string, 0, len(mm))
	for _, k := range keys {
		v := mm[k]
		kvTemp := fmt.Sprintf(`%s:%s`, k, v)
		kvStr = append(kvStr, kvTemp)
	}
	return "{" + strings.Join(kvStr, ",") + "}"
}

func (m *NebulaMap) GetValues() map[types.Value]types.Value {
	values := make(map[types.Value]types.Value)
	for k, v := range m.Values {
		values[k] = v
	}
	return values
}

func (g *NebulaDecimal) String() string {
	return g.Sval
}

func (g *NebulaGeography) String() string {
	switch g.Shape {
	case types.GeoShapePoint:
		if g.Point != nil {
			return fmt.Sprintf("POINT(%s %s)", formatFloat(g.Point.Lng, 64), formatFloat(g.Point.Lat, 64))
		}
	case types.GeoShapeLineString:
		if g.LineString != nil {
			coords := make([]string, 0, len(g.LineString))
			for _, coord := range g.LineString {
				coords = append(coords, fmt.Sprintf("%s %s", formatFloat(coord.Lng, 64), formatFloat(coord.Lat, 64)))
			}
			return fmt.Sprintf("LINESTRING(%s)", strings.Join(coords, ", "))
		}
	case types.GeoShapePolygon:
		if g.Polygon != nil {
			loopStrings := make([]string, 0, len(g.Polygon))
			for _, loop := range g.Polygon {
				coords := make([]string, 0, len(loop))
				for _, coord := range loop {
					coords = append(coords, fmt.Sprintf("%s %s", formatFloat(coord.Lng, 64), formatFloat(coord.Lat, 64)))
				}
				loopStrings = append(loopStrings, fmt.Sprintf("(%s)", strings.Join(coords, ", ")))
			}
			return fmt.Sprintf("POLYGON(%s)", strings.Join(loopStrings, ", "))
		}
	}
	return "GEOGRAPHY"
}

func (g *NebulaGeography) GetSRID() int32 {
	return g.SRID
}

func (g *NebulaGeography) GetShape() types.GeoShape {
	return g.Shape
}

func (g *NebulaGeography) GetPoint() *types.Point {
	return g.Point
}

func (g *NebulaGeography) GetLineString() types.LineString {
	return g.LineString
}

func (g *NebulaGeography) GetPolygon() types.Polygon {
	return g.Polygon
}

// (288314845273522179@City:City&Place{id:32,name:Norway,url:http://dbpedia.org/resource/Norway})
func (n *NebulaNode) String() string {
	mv := mapValue(n.GetProperties())
	return fmt.Sprintf("(%d@%s:%s{%s})",
		n.GetId(),
		n.GetType(),
		strings.Join(n.GetLabels(), "&"),
		mv.string(),
	)
}

func (n *NebulaNode) GetProperties() map[string]types.Value {
	properties := make(map[string]types.Value)
	for i, key := range n.PropNames {
		properties[key] = n.PropValues[i]
	}
	return properties
}

func (n *NebulaNode) GetId() int {
	return int(n.NodeId)
}

func (n *NebulaNode) GetGraph() string {
	return n.Graph
}

func (n *NebulaNode) GetType() string {
	return n.Type
}

func (n *NebulaNode) GetLabels() []string {
	return n.Labels
}

// (288314845273522179)<-[288314845273522179@connected_Sub_Load:
// connected_Sub_Load{cid:115967690673232363}]-(288314982712475649)
func (e *NebulaEdge) String() string {
	mv := mapValue(e.GetProperties())
	var (
		leftBracket  string
		rightBracket string
	)
	if e.IsDirected() {
		leftBracket = "-"
		rightBracket = "->"
	} else {
		leftBracket = "~"
		rightBracket = "~"
	}

	return fmt.Sprintf("(%d)%s[%d@%s:%s{%s}]%s(%d)",
		e.GetSrcId(),
		leftBracket,
		e.GetRank(),
		e.GetType(),
		strings.Join(e.GetLabels(), "&"),
		mv.string(),
		rightBracket,
		e.GetDstId(),
	)
}

func (e *NebulaEdge) GetProperties() map[string]types.Value {
	properties := make(map[string]types.Value)
	for i, key := range e.PropNames {
		properties[key] = e.PropValues[i]
	}
	return properties
}

func (e *NebulaEdge) GetSrcId() int {
	return int(e.SrcId)
}

func (e *NebulaEdge) GetDstId() int {
	return int(e.DstId)
}

func (e *NebulaEdge) GetGraph() string {
	return e.Graph
}

func (e *NebulaEdge) GetType() string {
	return e.Type
}

func (e *NebulaEdge) GetLabels() []string {
	return e.Labels
}

func (e *NebulaEdge) GetRank() int {
	return int(e.Rank)
}

func (e *NebulaEdge) IsDirected() bool {
	return e.Direction != edgeNoDirection
}

// (288314845273522179@City:City&Place{id:3,kind:3,name:org3,url:https://org3.com})-[288314845273522179@connected_Sub_Load:connected_Sub_Load{}]-(288315214640709633@City:City&Place{id:3,kind:city,name:Hangzhou,url:https://hangzhou.com})
// -[288315214640709633@connected_Sub_Load:connected_Sub_Load{}]
// -(288315309129990145@City:City&Place{id:6,kind:city,name:Shenzhen,url:https://shenzhen.com})
func (p *NebulaPath) String() string {
	values := p.GetValues()
	buf := bytes.NewBuffer(nil)
	defer buf.Reset()
	var preN types.Node
	for _, v := range values {
		if v.GetType() == types.ValueTypeNode {
			n, _ := v.AsNode()
			preN = n
			buf.WriteString(n.String())
		} else if v.GetType() == types.ValueTypeEdge {
			e, _ := v.AsEdge()
			mv := mapValue(e.GetProperties())
			estr := fmt.Sprintf("[%d@%s:%s{%s}]",
				e.GetRank(),
				e.GetType(),
				strings.Join(e.GetLabels(), "&"),
				mv.string(),
			)

			if e.IsDirected() {
				if e.GetSrcId() == preN.GetId() {
					buf.WriteString(fmt.Sprintf("-%s->", estr))
				} else {
					buf.WriteString(fmt.Sprintf("<-%s-", estr))
				}
			} else {
				buf.WriteString(fmt.Sprintf("~%s~", estr))
			}
		} else {
			//no other type
		}
	}
	return buf.String()
}

func (p *NebulaPath) GetValues() []types.Value {
	values := make([]types.Value, 0, len(p.Values))
	for _, v := range p.Values {
		values = append(values, v)
	}
	return values
}

func (l *NebulaLocalDatetime) String() string {
	//RFC3339 without timezone
	return fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d.%06d",
		l.Year, l.Month, l.Day,
		l.Hour, l.Minute, l.Sec, l.Microsec)
}

func (l *NebulaLocalDatetime) GetYear() int {
	return int(l.Year)
}

func (l *NebulaLocalDatetime) GetMonth() int {
	return int(l.Month)
}

func (l *NebulaLocalDatetime) GetDay() int {
	return int(l.Day)
}

func (l *NebulaLocalDatetime) GetHour() int {
	return int(l.Hour)
}

func (l *NebulaLocalDatetime) GetMinute() int {
	return int(l.Minute)
}

func (l *NebulaLocalDatetime) GetSec() int {
	return int(l.Sec)
}

func (l *NebulaLocalDatetime) GetMicrosec() int {
	return int(l.Microsec)
}

func (d *NebulaDate) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

func (d *NebulaDate) GetYear() int {
	return int(d.Year)
}

func (d *NebulaDate) GetMonth() int {
	return int(d.Month)
}

func (d *NebulaDate) GetDay() int {
	return int(d.Day)
}

func (t *NebulaLocalTime) String() string {
	return fmt.Sprintf("%02d:%02d:%02d.%06d",
		t.Hour, t.Minute, t.Sec, t.Microsec)
}

func (t *NebulaLocalTime) GetHour() int {
	return int(t.Hour)
}

func (t *NebulaLocalTime) GetMinute() int {
	return int(t.Minute)
}

func (t *NebulaLocalTime) GetSec() int {
	return int(t.Sec)
}

func (t *NebulaLocalTime) GetMicrosec() int {
	return int(t.Microsec)
}

func (d *NebulaDuration) String() string {
	var prefix string = "P"
	if d.IsMonthBased() {
		if d.GetYear() != 0 {
			prefix += fmt.Sprintf("%dY", d.GetYear())
		}
		if d.GetMonth() != 0 {
			prefix += fmt.Sprintf("%dM", d.GetMonth())
		}
		if d.GetYear() == 0 && d.GetMonth() == 0 {
			prefix += "0M"
		}
	} else {
		if d.GetDay() != 0 {
			prefix += fmt.Sprintf("%dD", d.GetDay())
		}
		if d.GetHour() != 0 || d.GetMinute() != 0 || d.GetSecond() != 0 || d.GetMicrosecond() != 0 {
			prefix += "T"
		}
		if d.GetHour() != 0 {
			prefix += fmt.Sprintf("%dH", d.GetHour())
		}
		if d.GetMinute() != 0 {
			prefix += fmt.Sprintf("%dM", d.GetMinute())
		}
		if d.GetSecond() != 0 || d.GetMicrosecond() != 0 {
			if d.GetMicrosecond() == 0 {
				prefix += fmt.Sprintf("%dS", d.GetSecond())
			} else {
				ms := d.GetSecond()*1e6 + d.GetMicrosecond()
				isMinus := d.GetSecond() < 0 || d.GetMicrosecond() < 0
				if isMinus {
					ms = -ms
				}
				s, ss := ms/1e6, ms%1e6
				if isMinus {
					prefix += fmt.Sprintf("-%d.%06d", s, ss)
				} else {
					prefix += fmt.Sprintf("%d.%06d", s, ss)
				}
				prefix = strings.TrimRight(prefix, "0")
				prefix += "S"
			}
		}
		if d.GetDay() == 0 && d.GetHour() == 0 && d.GetMinute() == 0 && d.GetSecond() == 0 && d.GetMicrosecond() == 0 {
			prefix += "T0S"
		}
	}
	return prefix
}

func (d *NebulaDuration) IsMonthBased() bool {
	return d.MonthBased
}

func (d *NebulaDuration) GetYear() int {
	return int(d.Year)
}

func (d *NebulaDuration) GetMonth() int {
	return int(d.Month)
}

func (d *NebulaDuration) GetDay() int {
	return int(d.Day)
}

func (d *NebulaDuration) GetHour() int {
	return int(d.Hour)
}

func (d *NebulaDuration) GetMinute() int {
	return int(d.Minute)
}

func (d *NebulaDuration) GetSecond() int {
	return int(d.Sec)
}

func (d *NebulaDuration) GetMicrosecond() int {
	return int(d.Microsec)
}

func (zt *NebulaZonedTime) String() string {
	//TODO server would return offset with seconds
	offset := zt.GetOffset()
	var zone string
	if offset < 0 {
		zone = fmt.Sprintf("-%02d:%02d", -offset/3600, (-offset%3600)/60)
	} else if offset > 0 {
		zone = fmt.Sprintf("+%02d:%02d", offset/3600, (offset%3600)/60)
	} else {
		zone = "Z"
	}

	return fmt.Sprintf("%02d:%02d:%02d.%06d%s",
		zt.Hour,
		zt.Minute,
		zt.Sec,
		zt.Microsec,
		zone,
	)
}

func (zt *NebulaZonedTime) GetHour() int {
	return int(zt.Hour)
}

func (zt *NebulaZonedTime) GetMinute() int {
	return int(zt.Minute)
}

func (zt *NebulaZonedTime) GetSec() int {
	return int(zt.Sec)
}

func (zt *NebulaZonedTime) GetMicrosec() int {
	return int(zt.Microsec)
}

func (zt *NebulaZonedTime) GetOffset() int {
	return int(zt.Offset)
}

func (zdt *NebulaZonedDatetime) String() string {
	offset := zdt.GetOffset()
	var zone string
	if offset < 0 {
		zone = fmt.Sprintf("-%02d:%02d", -offset/3600, (-offset%3600)/60)
	} else if offset > 0 {
		zone = fmt.Sprintf("+%02d:%02d", offset/3600, (offset%3600)/60)
	} else {
		zone = "Z"
	}
	return fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d.%06d%s",
		zdt.Year, zdt.Month, zdt.Day,
		zdt.Hour, zdt.Minute, zdt.Sec, zdt.Microsec,
		zone)
}

func (zdt *NebulaZonedDatetime) GetOffset() int {
	return int(zdt.Offset)
}

func (zdt *NebulaZonedDatetime) Time() *time.Time {
	if zdt == nil {
		return nil
	}
	timezone := time.FixedZone("", int(zdt.GetOffset()))
	t := time.Date(
		int(zdt.GetYear()),
		time.Month(zdt.GetMonth()),
		int(zdt.GetDay()),
		int(zdt.GetHour()),
		int(zdt.GetMinute()),
		int(zdt.GetSec()),
		int(zdt.GetMicrosec())*int(time.Microsecond),
		timezone,
	)
	return &t
}

func (zdt *NebulaZonedDatetime) GetYear() int {
	return int(zdt.Year)
}

func (zdt *NebulaZonedDatetime) GetMonth() int {
	return int(zdt.Month)
}

func (zdt *NebulaZonedDatetime) GetDay() int {
	return int(zdt.Day)
}

func (zdt *NebulaZonedDatetime) GetHour() int {
	return int(zdt.Hour)
}

func (zdt *NebulaZonedDatetime) GetMinute() int {
	return int(zdt.Minute)
}

func (zdt *NebulaZonedDatetime) GetSec() int {
	return int(zdt.Sec)
}

func (zdt *NebulaZonedDatetime) GetMicrosec() int {
	return int(zdt.Microsec)
}

func (ev *NebulaEmbeddingVector) String() string {
	valuesStr := make([]string, 0, len(ev.Values))
	for _, v := range ev.Values {
		valuesStr = append(valuesStr, formatFloat(float64(v), 32))
	}
	return fmt.Sprintf("[%s]", strings.Join(valuesStr, ","))
}

func (ev *NebulaEmbeddingVector) GetValues() []float32 {
	return ev.Values
}

func (m mapValue) string() string {
	var kvStr []string = make([]string, 0, len(m))
	var keys []string = make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, k := range keys {
		v := m[k]
		kvTemp := fmt.Sprintf(`%s:%s`, k, v)
		kvStr = append(kvStr, kvTemp)
	}
	return strings.Join(kvStr, ",")
}

var valuePool = sync.Pool{
	New: func() any {
		return &NebulaValue{}
	},
}

func (value *NebulaValue) needReset(typ types.ValueType) bool {
	if value.Data == nil || value.GetType() != typ {
		return true
	}
	return false
}

func DeepCopyValue(src *NebulaValue, dst *NebulaValue) {
	switch src.GetType() {
	case types.ValueTypeNull:
		dst.Data = nil
	case types.ValueTypeBool:
		if dst.needReset(types.ValueTypeBool) {
			var b NebulaBool
			dst.Data = &b
		}
		d := dst.Data.(*NebulaBool)
		*d = NebulaBool(*src.Data.(*NebulaBool))
	case types.ValueTypeInt8:
		if dst.needReset(types.ValueTypeInt8) {
			var i NebulaInt8
			dst.Data = &i
		}
		d := dst.Data.(*NebulaInt8)
		*d = NebulaInt8(*src.Data.(*NebulaInt8))
	case types.ValueTypeInt16:
		if dst.needReset(types.ValueTypeInt16) {
			var i NebulaInt16
			dst.Data = &i
		}
		d := dst.Data.(*NebulaInt16)
		*d = NebulaInt16(*src.Data.(*NebulaInt16))
	case types.ValueTypeInt32:
		if dst.needReset(types.ValueTypeInt32) {
			var i NebulaInt32
			dst.Data = &i
		}
		d := dst.Data.(*NebulaInt32)
		*d = NebulaInt32(*src.Data.(*NebulaInt32))
	case types.ValueTypeInt64:
		if dst.needReset(types.ValueTypeInt64) {
			var i NebulaInt64
			dst.Data = &i
		}
		d := dst.Data.(*NebulaInt64)
		*d = NebulaInt64(*src.Data.(*NebulaInt64))
	case types.ValueTypeUInt8:
		if dst.needReset(types.ValueTypeUInt8) {
			var i NebulaUint8
			dst.Data = &i
		}
		d := dst.Data.(*NebulaUint8)
		*d = NebulaUint8(*src.Data.(*NebulaUint8))
	case types.ValueTypeUInt16:
		if dst.needReset(types.ValueTypeUInt16) {
			var i NebulaUint16
			dst.Data = &i
		}
		d := dst.Data.(*NebulaUint16)
		*d = NebulaUint16(*src.Data.(*NebulaUint16))
	case types.ValueTypeUInt32:
		if dst.needReset(types.ValueTypeUInt32) {
			var i NebulaUint32
			dst.Data = &i
		}
		d := dst.Data.(*NebulaUint32)
		*d = NebulaUint32(*src.Data.(*NebulaUint32))
	case types.ValueTypeUInt64:
		if dst.needReset(types.ValueTypeUInt64) {
			var i NebulaUint64
			dst.Data = &i
		}
		d := dst.Data.(*NebulaUint64)
		*d = NebulaUint64(*src.Data.(*NebulaUint64))
	case types.ValueTypeFloat:
		if dst.needReset(types.ValueTypeFloat) {
			var f NebulaFloat
			dst.Data = &f
		}
		d := dst.Data.(*NebulaFloat)
		*d = NebulaFloat(*src.Data.(*NebulaFloat))
	case types.ValueTypeDouble:
		if dst.needReset(types.ValueTypeDouble) {
			var d NebulaDouble
			dst.Data = &d
		}
		d := dst.Data.(*NebulaDouble)
		*d = NebulaDouble(*src.Data.(*NebulaDouble))
	case types.ValueTypeString:
		if dst.needReset(types.ValueTypeString) {
			var s NebulaString
			dst.Data = &s
		}
		d := dst.Data.(*NebulaString)
		*d = NebulaString(*src.Data.(*NebulaString))
	case types.ValueTypeDuration:
		srcDuration := src.Data.(*NebulaDuration)
		if dst.needReset(types.ValueTypeDuration) {
			dst.Data = &NebulaDuration{}
		}
		dstDuration := dst.Data.(*NebulaDuration)
		dstDuration.MonthBased = srcDuration.MonthBased
		dstDuration.Year = srcDuration.Year
		dstDuration.Month = srcDuration.Month
		dstDuration.Day = srcDuration.Day
		dstDuration.Hour = srcDuration.Hour
		dstDuration.Minute = srcDuration.Minute
		dstDuration.Sec = srcDuration.Sec
		dstDuration.Microsec = srcDuration.Microsec
	case types.ValueTypeLocalTime:
		srcTime := src.Data.(*NebulaLocalTime)
		if dst.needReset(types.ValueTypeLocalTime) {
			dst.Data = &NebulaLocalTime{}
		}
		dstTime := dst.Data.(*NebulaLocalTime)
		dstTime.Hour = srcTime.Hour
		dstTime.Minute = srcTime.Minute
		dstTime.Sec = srcTime.Sec
		dstTime.Microsec = srcTime.Microsec
	case types.ValueTypeLocalDateTime:
		srcDatetime := src.Data.(*NebulaLocalDatetime)
		if dst.needReset(types.ValueTypeLocalDateTime) {
			dst.Data = &NebulaLocalDatetime{}
		}
		dstDatetime := dst.Data.(*NebulaLocalDatetime)
		dstDatetime.Year = srcDatetime.Year
		dstDatetime.Month = srcDatetime.Month
		dstDatetime.Day = srcDatetime.Day
		dstDatetime.Hour = srcDatetime.Hour
		dstDatetime.Minute = srcDatetime.Minute
		dstDatetime.Sec = srcDatetime.Sec
		dstDatetime.Microsec = srcDatetime.Microsec

	case types.ValueTypeZonedTime:
		srcZonedTime := src.Data.(*NebulaZonedTime)
		if dst.needReset(types.ValueTypeZonedTime) {
			dst.Data = &NebulaZonedTime{}
		}
		dstZonedTime := dst.Data.(*NebulaZonedTime)
		dstZonedTime.Hour = srcZonedTime.Hour
		dstZonedTime.Minute = srcZonedTime.Minute
		dstZonedTime.Sec = srcZonedTime.Sec
		dstZonedTime.Microsec = srcZonedTime.Microsec
		dstZonedTime.Offset = srcZonedTime.Offset

	case types.ValueTypeZonedDateTime:
		srcZonedDatetime := src.Data.(*NebulaZonedDatetime)
		if dst.needReset(types.ValueTypeZonedDateTime) {
			dst.Data = &NebulaZonedDatetime{}
		}
		dstZonedDatetime := dst.Data.(*NebulaZonedDatetime)
		dstZonedDatetime.Year = srcZonedDatetime.Year
		dstZonedDatetime.Month = srcZonedDatetime.Month
		dstZonedDatetime.Day = srcZonedDatetime.Day
		dstZonedDatetime.Hour = srcZonedDatetime.Hour
		dstZonedDatetime.Minute = srcZonedDatetime.Minute
		dstZonedDatetime.Sec = srcZonedDatetime.Sec
		dstZonedDatetime.Microsec = srcZonedDatetime.Microsec
		dstZonedDatetime.Offset = srcZonedDatetime.Offset
	case types.ValueTypeDate:
		srcDate := src.Data.(*NebulaDate)
		if dst.needReset(types.ValueTypeDate) {
			dst.Data = &NebulaDate{}
		}
		dstDate := dst.Data.(*NebulaDate)
		dstDate.Year = srcDate.Year
		dstDate.Month = srcDate.Month
		dstDate.Day = srcDate.Day
	case types.ValueTypeList:
		srcList := src.Data.(*NebulaList)
		if dst.needReset(types.ValueTypeList) {
			dst.Data = &NebulaList{}
		}
		dstList := dst.Data.(*NebulaList)
		dstList.Values = constructListValue(dstList.Values, len(srcList.Values))
		for i, v := range srcList.Values {
			DeepCopyValue(v, dstList.Values[i])
		}
	case types.ValueTypeSet:
		srcSet := src.Data.(*NebulaSet)
		if dst.needReset(types.ValueTypeSet) {
			dst.Data = &NebulaSet{}
		}
		dstSet := dst.Data.(*NebulaSet)
		dstSet.Values = constructListValue(dstSet.Values, len(srcSet.Values))
		for i, v := range srcSet.Values {
			DeepCopyValue(v, dstSet.Values[i])
		}
	case types.ValueTypeMap:
		srcMap := src.Data.(*NebulaMap)
		if dst.needReset(types.ValueTypeMap) {
			dst.Data = &NebulaMap{}
		}
		dstMap := dst.Data.(*NebulaMap)
		dstMap.Values = make(map[*NebulaValue]*NebulaValue)
		for key, v := range srcMap.Values {
			dstMap.Values[key] = &NebulaValue{}
			DeepCopyValue(v, dstMap.Values[key])
		}
	case types.ValueTypeRecord:
		srcRecord := src.Data.(*NebulaRecord)
		if dst.needReset(types.ValueTypeRecord) {
			dst.Data = &NebulaRecord{}
		}
		dstRecord := dst.Data.(*NebulaRecord)
		dstRecord.Values = constructMapValue(dstRecord.Values, srcRecord.Values)
		for key, v := range srcRecord.Values {
			DeepCopyValue(v, dstRecord.Values[key])
		}
	case types.ValueTypeEmbeddingVector:
		srcEv := src.Data.(*NebulaEmbeddingVector)
		if dst.needReset(types.ValueTypeEmbeddingVector) {
			dst.Data = &NebulaEmbeddingVector{}
		}
		dstEv := dst.Data.(*NebulaEmbeddingVector)
		dstEv.Values = make([]float32, len(srcEv.Values))
		copy(dstEv.Values, srcEv.Values)
	case types.ValueTypeNode:
		srcNode := src.Data.(*NebulaNode)
		if dst.needReset(types.ValueTypeNode) {
			dst.Data = &NebulaNode{}
		}
		dstNode := dst.Data.(*NebulaNode)
		dstNode.NodeId = srcNode.NodeId
		dstNode.Graph = srcNode.Graph
		dstNode.Type = srcNode.Type
		dstNode.Labels = constructListString(dstNode.Labels, len(srcNode.Labels))
		copy(dstNode.Labels, srcNode.Labels)
		dstNode.PropNames = constructListString(dstNode.PropNames, len(srcNode.PropNames))
		copy(dstNode.PropNames, srcNode.PropNames)
		dstNode.PropValues = constructListValue(dstNode.PropValues, len(srcNode.PropValues))
		for i, v := range srcNode.PropValues {
			DeepCopyValue(v, dstNode.PropValues[i])
		}
	case types.ValueTypeEdge:
		srcEdge := src.Data.(*NebulaEdge)
		if dst.needReset(types.ValueTypeEdge) {
			dst.Data = &NebulaEdge{}
		}
		dstEdge := dst.Data.(*NebulaEdge)
		dstEdge.SrcId = srcEdge.SrcId
		dstEdge.DstId = srcEdge.DstId
		dstEdge.Graph = srcEdge.Graph
		dstEdge.Type = srcEdge.Type
		dstEdge.Labels = constructListString(dstEdge.Labels, len(srcEdge.Labels))
		copy(dstEdge.Labels, srcEdge.Labels)
		dstEdge.Rank = srcEdge.Rank
		dstEdge.Direction = srcEdge.Direction
		dstEdge.PropNames = constructListString(dstEdge.PropNames, len(srcEdge.PropNames))
		copy(dstEdge.PropNames, srcEdge.PropNames)
		dstEdge.PropValues = constructListValue(dstEdge.PropValues, len(srcEdge.PropValues))
		for i, v := range srcEdge.PropValues {
			DeepCopyValue(v, dstEdge.PropValues[i])
		}
	case types.ValueTypePath:
		srcPath := src.Data.(*NebulaPath)
		if dst.needReset(types.ValueTypePath) {
			dst.Data = &NebulaPath{}
		}
		dstPath := dst.Data.(*NebulaPath)
		dstPath.Values = constructListValue(dstPath.Values, len(srcPath.Values))
		for i, v := range srcPath.Values {
			DeepCopyValue(v, dstPath.Values[i])
		}
	case types.ValueTypeDecimal:
		srcDecimal := src.Data.(*NebulaDecimal)
		if dst.needReset(types.ValueTypeDecimal) {
			dst.Data = &NebulaDecimal{}
		}
		dstDecimal := dst.Data.(*NebulaDecimal)
		dstDecimal.Sval = srcDecimal.Sval
	case types.ValueTypeGeography:
		srcGeography := src.Data.(*NebulaGeography)
		if dst.needReset(types.ValueTypeGeography) {
			dst.Data = &NebulaGeography{}
		}
		dstGeography := dst.Data.(*NebulaGeography)
		dstGeography.SRID = srcGeography.SRID
		dstGeography.Shape = srcGeography.Shape
		if srcGeography.Point != nil {
			dstGeography.Point = &types.Point{
				Lng: srcGeography.Point.Lng,
				Lat: srcGeography.Point.Lat,
			}
		}
		if srcGeography.LineString != nil {
			dstGeography.LineString = make(types.LineString, len(srcGeography.LineString))
			for i, coord := range srcGeography.LineString {
				dstGeography.LineString[i] = &types.Point{
					Lng: coord.Lng,
					Lat: coord.Lat,
				}
			}
		}
		if srcGeography.Polygon != nil {
			dstGeography.Polygon = make(types.Polygon, len(srcGeography.Polygon))
			for i, ring := range srcGeography.Polygon {
				dstGeography.Polygon[i] = make(types.LineString, len(ring))
				for j, coord := range ring {
					dstGeography.Polygon[i][j] = &types.Point{
						Lng: coord.Lng,
						Lat: coord.Lat,
					}
				}
			}
		}
	}
}

func constructMapValue(originalMap map[string]*NebulaValue, srcMap map[string]*NebulaValue) map[string]*NebulaValue {
	if originalMap == nil {
		originalMap = make(map[string]*NebulaValue, len(srcMap))
		for key := range srcMap {
			originalMap[key] = &NebulaValue{}
		}
		return originalMap
	}

	newMap := make(map[string]*NebulaValue, len(srcMap))
	for key := range srcMap {
		if v, ok := originalMap[key]; ok {
			newMap[key] = v
		} else {
			newMap[key] = &NebulaValue{}
		}
	}
	return newMap
}

func constructListString(originalList []string, expectedLen int) []string {
	if originalList == nil {
		originalList = make([]string, 0, expectedLen)
		for i := 0; i < expectedLen; i++ {
			originalList = append(originalList, "")
		}
		return originalList
	}
	if len(originalList) >= expectedLen {
		return originalList[:expectedLen]
	}

	newList := originalList[:]
	for i := len(originalList); i < expectedLen; i++ {
		newList = append(newList, "")
	}
	return newList
}

func constructListValue(originalList []*NebulaValue, expectedLen int) []*NebulaValue {
	if originalList == nil {
		originalList = make([]*NebulaValue, 0, expectedLen)
		for i := 0; i < expectedLen; i++ {
			originalList = append(originalList, &NebulaValue{})
		}
		return originalList
	}
	if len(originalList) >= expectedLen {
		return originalList[:expectedLen]
	}

	newList := originalList[:]
	for i := len(originalList); i < expectedLen; i++ {
		newList = append(newList, &NebulaValue{})
	}
	return newList
}
