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
	"io"
	"time"

	"github.com/vesoft-inc/nebula-go/v5/internal/generated_code/v5.0.0/proto/vector"
	"github.com/vesoft-inc/nebula-go/v5/pkg/errors"
	"github.com/vesoft-inc/nebula-go/v5/pkg/types"
)

var columnTypeMap = map[uint8]types.ColumnType{
	0x1: types.ColumnTypeNode,
	0x2: types.ColumnTypeEdge,
	0x3: types.ColumnTypeUnknown,
	0x4: types.ColumnTypeBool,
	0x5: types.ColumnTypeInt8,
	0x6: types.ColumnTypeUint8,
	0x7: types.ColumnTypeInt16,
	0x8: types.ColumnTypeUint16,
	0x9: types.ColumnTypeInt32,
	0xa: types.ColumnTypeUint32,
	0xb: types.ColumnTypeInt64,
	0xc: types.ColumnTypeUint64,
	0xd: types.ColumnTypeFloat32,
	0xe: types.ColumnTypeFloat64,
	// 0xf: ColumnTypeBytes, not support yet
	0x10: types.ColumnTypeString,
	0x11: types.ColumnTypeList,
	0x12: types.ColumnTypePath,
	0x13: types.ColumnTypeRecord,
	0x14: types.ColumnTypeVector,
	0x15: types.ColumnTypeLocalTime,
	0x16: types.ColumnTypeDuration,
	0x17: types.ColumnTypeDate,
	0x18: types.ColumnTypeLocalDatetime,
	0x19: types.ColumnTypeZonedTime,
	0x20: types.ColumnTypeZonedDatetime,
	0x22: types.ColumnTypeDecimal,
	0x24: types.ColumnTypeGeography,
	0x25: types.ColumnTypeSet,
	0x26: types.ColumnTypeMap,
	0xFE: types.ColumnTypeAny,
	0xFF: types.ColumnTypeInvalid,
}

type vectorType uint8

const (
	vectorTypeInvalid vectorType = iota
	vectorTypeConst
	vectorTypeFlat
	vectorTypeParallel
)

type ResultTable struct {
	table                *vector.VectorResultTable
	numBatches           int
	columnNames          []string
	columnTypes          []typeSchema
	ColumnTypeBytes      [][]byte
	batches              []batcher
	batchIndex           uint64
	currentBatch         batcher
	currentBatchRowIndex uint32
	decodeContext        *decodeContext
	decoder              decoder
}

type decodeContext struct {
	timezoneOffset int64 //offset in seconds
	graphsSchema   graphsSchema
}

type batcher interface {
	numRecords() uint32
	getRowByIndex(index uint32, values []types.Value) error
}

type batch struct {
	vectors []*vectorWrapper
}

// vectorWrapper is a wrapper for vector.NestedVector
// one column has one vectorWrapper
type vectorWrapper struct {
	vector        *vector.NestedVector
	nullAllSet    bool // allSet is true means all values in the vector are not null
	vectorType    vectorType
	typ           typeSchema
	elementProps  graphElementProps
	pathMeta      *pathMetaData
	decoder       decoder
	constValue    *NebulaValue
	decodeContext *decodeContext
	prepared      bool
}

type pathMetaData struct {
	nodeTypeIndex pathElementTypeIndexMap
	edgeTypeIndex pathElementTypeIndexMap
	nodeIndexPair map[pathPairIndex]*pathPair
	edgeIndexPair map[pathPairIndex]*pathPair
}

type pathPairIndex uint16
type pathElementTypeIndexMap map[int32]pathPairIndex
type pathPair struct {
	cur *vector.NestedVector
	adj *vector.NestedVector
}
type pathAdjHeader struct {
	isEnd           bool
	nextIsEdge      bool
	nextVectorIndex uint32
	nextOffset      uint32
}
type props map[string]*vectorProps
type elementProps map[int32]props
type graphElementProps map[int32]elementProps

type vectorProps struct {
	name        string
	typ         typeSchema
	vectorIndex int
}

func newPathAdjHeader(value int64) *pathAdjHeader {
	header := &pathAdjHeader{}
	header.isEnd = ((value >> 63) & 1) == 1
	header.nextIsEdge = ((value >> 62) & 1) == 1
	header.nextVectorIndex = (uint32)((value >> 32) & 0xFFFF)
	header.nextOffset = (uint32)(value & 0xFFFFFFFF)
	return header
}

func NewResultTable(table *vector.VectorResultTable) (*ResultTable, error) {
	if table == nil || table.Meta == nil {
		return nil, nil
	}
	if table.Batch != nil && len(table.Batch) > 0 {
		// check if the length of table.Batch is equal to the length of table.Meta.VectorBatchMetaData
		if int(table.Meta.NumBatches) != len(table.Batch) {
			return nil, errBatchNotEqual
		}
	}
	t := &ResultTable{table: table, decoder: defaultDecoder}
	// construct graph schema
	gsm := t.constructGraphsSchema()
	timeZone := int32(table.Meta.TimeZoneOffset)
	dctx := &decodeContext{
		timezoneOffset: int64(timeZone) * int64(time.Minute/time.Second),
		graphsSchema:   gsm,
	}
	t.decodeContext = dctx

	if table.Meta.RowType != nil {
		t.columnNames = table.Meta.RowType.ColumnNames
		typs := table.Meta.RowType.ColumnTypes
		for _, typ := range typs {
			r := newBytesReader(typ.ValueType)
			columnType, err := newTypeSchema(r)
			if err != nil {
				return nil, err
			}
			t.columnTypes = append(t.columnTypes, columnType)
			t.ColumnTypeBytes = append(t.ColumnTypeBytes, typ.ValueType)
		}
	}
	t.numBatches = len(table.Batch)
	for i := 0; i < t.numBatches; i++ {
		vs := make([]*vectorWrapper, 0, len(table.Batch[i].Vectors))
		for colIndex, v := range table.Batch[i].Vectors {
			typ := t.columnTypes[colIndex]
			wrapper := &vectorWrapper{
				vector:        v,
				typ:           typ,
				decodeContext: dctx,
				decoder:       defaultDecoder,
			}
			if typ.getType() != types.ColumnTypeUnknown {
				vectorTypeCode := v.CommonMetaData.VectorContentType
				vt := uint8(vectorTypeCode & 0xFF)
				nullAllSet := (vectorTypeCode & 1 << 8) != 0
				wrapper.vectorType = vectorType(vt)
				wrapper.nullAllSet = nullAllSet
			}
			vs = append(vs, wrapper)
		}

		t.batches = append(t.batches, &batch{
			vectors: vs,
		})
	}

	return t, nil
}

func (rt *ResultTable) NumRecords() uint64 {
	return rt.table.Meta.NumRecords
}

func (rt *ResultTable) ColumnNames() []string {
	return rt.columnNames
}

func (rt *ResultTable) ColumnTypes() []types.ColumnType {
	types := make([]types.ColumnType, 0, len(rt.columnTypes))
	for _, t := range rt.columnTypes {
		types = append(types, t.getType())
	}
	return types
}

func (rt *ResultTable) constructGraphsSchema() graphsSchema {
	if rt.table.Meta.GraphSchema == nil {
		return nil
	}
	gsm := make(graphsSchema)
	for _, g := range rt.table.Meta.GraphSchema {
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

func (rt *ResultTable) getCurrentBatch() error {
	for {
		if len(rt.batches) == 0 || rt.batchIndex >= uint64(rt.numBatches) {
			return io.EOF
		}
		if rt.currentBatch == nil {
			rt.currentBatch = rt.batches[rt.batchIndex]
		}
		if rt.currentBatchRowIndex < rt.currentBatch.numRecords() {
			return nil
		}

		batch, ok := rt.currentBatch.(*batch)
		if ok {
			batch.vectors = nil
		}

		rt.batchIndex++
		rt.currentBatchRowIndex = 0
		rt.currentBatch = nil
	}
}

func (rt *ResultTable) Scan(values []types.Value) error {
	if err := rt.getCurrentBatch(); err != nil {
		return err
	}

	if err := rt.currentBatch.getRowByIndex(rt.currentBatchRowIndex, values); err != nil {
		return err
	}
	rt.currentBatchRowIndex++
	return nil
}

func (rt *ResultTable) Next() ([]types.Value, error) {
	values := make([]types.Value, 0, len(rt.columnNames))
	for i := 0; i < len(rt.columnNames); i++ {
		values = append(values, &NebulaValue{})
	}
	if err := rt.getCurrentBatch(); err != nil {
		return nil, err
	}

	if err := rt.currentBatch.getRowByIndex(rt.currentBatchRowIndex, values); err != nil {
		return nil, err
	}
	rt.currentBatchRowIndex++
	return values, nil
}

func (b *batch) numRecords() uint32 {
	if len(b.vectors) == 0 {
		return 0
	}
	// all vectors in a batch have the same number of records
	v := b.vectors[0]
	return v.vector.CommonMetaData.GetNumRecords()
}

func (b *batch) getRowByIndex(index uint32, values []types.Value) error {
	if index >= b.numRecords() {
		return errors.Wrap(errOutOfRange, "")
	}
	for i, v := range b.vectors {
		if err := v.prepare(); err != nil {
			return err
		}
		nvs := values[i].(*NebulaValue)
		if err := v.decodeValue(nvs, index); err != nil {
			return err
		}
	}
	return nil
}

// prepare some meta data for flat vector
func (v *vectorWrapper) prepare() error {
	var err error
	if v.vectorType != vectorTypeFlat {
		return nil
	}
	if v.prepared {
		return nil
	}
	switch v.typ.getType() {
	case types.ColumnTypeNode, types.ColumnTypeEdge:
		isNode := v.typ.getType() == types.ColumnTypeNode
		v.elementProps, err = v.getElementPropVectorIndex(isNode)
		if err != nil {
			return err
		}
		schema := v.typ.(*columnTypeSchemaElement)
		schema.graphElementProps = v.elementProps
	case types.ColumnTypePath:
		v.pathMeta, err = v.getPathSpecialData()
		if err != nil {
			return err
		}
		schema := v.typ.(*columnTypeSchemaPath)
		schema.meta = v.pathMeta
	default:
		// do nothing, other types do not need to prepare
	}
	v.prepared = true
	return nil
}

func (v *vectorWrapper) decodeValue(value *NebulaValue, index uint32) error {
	if v.typ == nil {
		return errColumnTypeIsNil
	}
	if v.typ.getType() == types.ColumnTypeUnknown {
		value.Data = nil
		return nil
	}
	if v.vectorType == vectorTypeConst && v.constValue != nil {
		value.Data = v.constValue.Data
		return nil
	}

	if err := v.decoder.decodeValue(v.decodeContext, value, v.vector, v.vectorType, index, v.typ); err != nil {
		return err
	}
	if v.vectorType == vectorTypeConst {
		v.constValue = value
	}
	return nil
}

func (v *vectorWrapper) getElementPropVectorIndex(isNode bool) (graphElementProps, error) {
	if v.elementProps == nil {
		elementSchema, ok := v.typ.(*columnTypeSchemaElement)
		if !ok {
			return nil, errTypeAssertion
		}
		allElementProps := elementSchema.graphElementProps
		if err := decodePropVectorIndex(allElementProps, v.vector.SpecialMetaData, isNode); err != nil {
			return nil, err
		}
		v.elementProps = allElementProps
	}
	return v.elementProps, nil
}

func (v *vectorWrapper) getPathSpecialData() (*pathMetaData, error) {
	if v.pathMeta == nil {
		v.pathMeta = &pathMetaData{
			nodeTypeIndex: make(pathElementTypeIndexMap),
			edgeTypeIndex: make(pathElementTypeIndexMap),
			nodeIndexPair: make(map[pathPairIndex]*pathPair),
			edgeIndexPair: make(map[pathPairIndex]*pathPair),
		}
		if err := decodePathSpecialData(v.vector, v.pathMeta); err != nil {
			return nil, err
		}
	}
	return v.pathMeta, nil
}

func decodePathSpecialData(v *vector.NestedVector, meta *pathMetaData) error {
	// special meta data
	// num of node type + [graph id + node type id + pair index] + num of edge type + [graph id + edge type id + pair index]
	r := newBytesReader(v.SpecialMetaData)
	nodeTypeNumBytes := r.readN(4)
	if r.error() != nil {
		return r.error()
	}
	nodeTypeNum := bytesToInt32(nodeTypeNumBytes)
	nestedVectorIndex := 0
	for i := 0; i < int(nodeTypeNum); i++ {
		graphIdTypes := r.readN(4)
		nodeTypeIdBytes := r.readN(2)
		pairIndexBytes := r.readN(2)
		if r.error() != nil {
			return r.error()
		}
		_ = bytesToInt32(graphIdTypes)
		nodeTypeId := int32(bytesToUint16(nodeTypeIdBytes))
		pairIndex := bytesToUint16(pairIndexBytes)
		meta.nodeTypeIndex[nodeTypeId] = pathPairIndex(pairIndex)
		meta.nodeIndexPair[pathPairIndex(pairIndex)] = &pathPair{
			cur: v.NestedVectors[nestedVectorIndex],
			adj: v.NestedVectors[nestedVectorIndex+1],
		}
		nestedVectorIndex += 2
	}
	edgeTypeNumBytes := r.readN(4)
	if r.error() != nil {
		return r.error()
	}
	edgeTypeNum := bytesToInt32(edgeTypeNumBytes)
	for i := 0; i < int(edgeTypeNum); i++ {
		graphIdTypes := r.readN(4)
		edgeTypeIdBytes := r.readN(4)
		pairIndexBytes := r.readN(2)
		if r.error() != nil {
			return r.error()
		}
		_ = bytesToInt32(graphIdTypes)
		edgeTypeId := bytesToInt32(edgeTypeIdBytes)
		pairIndex := bytesToInt16(pairIndexBytes)
		meta.edgeTypeIndex[edgeTypeId] = pathPairIndex(pairIndex)
		meta.edgeIndexPair[pathPairIndex(pairIndex)] = &pathPair{
			cur: v.NestedVectors[nestedVectorIndex],
			adj: v.NestedVectors[nestedVectorIndex+1],
		}
		nestedVectorIndex += 2
	}
	return nil
}

func decodePropVectorIndex(graphElementTypes graphElementProps, bs []byte, isNode bool) error {
	// properties num + [prop names] + element type num + [graph id + element type id + prop num + [vector index]]
	// vector index is same as the order of prop names
	elementTypeSize := 2
	if !isNode {
		elementTypeSize = 4
	}
	r := newBytesReader(bs)
	propsNumBytes := r.readN(4)
	if r.error() != nil {
		return r.error()
	}
	propsNum := bytesToInt32(propsNumBytes)
	var propList = make([]string, 0, propsNum)
	for i := 0; i < int(propsNum); i++ {
		sizeBytes := r.readN(2)
		if r.error() != nil {
			return r.error()
		}
		size := bytesToInt16(sizeBytes)
		propNameBytes := r.readN(int(size))
		if r.error() != nil {
			return r.error()
		}
		propList = append(propList, string(propNameBytes))
	}
	nodeTypeNumBytes := r.readN(4)
	if r.error() != nil {
		return r.error()
	}
	nodeTypeNum := bytesToInt32(nodeTypeNumBytes)
	for i := 0; i < int(nodeTypeNum); i++ {
		graphIdBytes := r.readN(4)
		elementTypeIdBytes := r.readN(elementTypeSize)
		propNumBytes := r.readN(4)
		if r.error() != nil {
			return r.error()
		}
		graphId := bytesToInt32(graphIdBytes)
		var elementId int32
		if isNode {
			elementId = int32(bytesToInt16(elementTypeIdBytes))
		} else {
			elementId = bytesToInt32(elementTypeIdBytes)
		}
		propNum := int(bytesToInt32(propNumBytes))
		for j := 0; j < propNum; j++ {
			vectorIndexBytes := r.readN(4)
			if r.error() != nil {
				return r.error()
			}
			vectorIndex := bytesToInt32(vectorIndexBytes)
			elementTypes, ok := graphElementTypes[graphId]
			if !ok {
				return errors.Wrap(errElementTypeNotFount, "")
			}
			et, ok := elementTypes[elementId]
			if !ok {
				return errors.Wrap(errElementTypeNotFount, "")
			}

			prop, ok := et[propList[vectorIndex]]
			if !ok {
				return errPropNotFound
			}
			prop.vectorIndex = int(vectorIndex)
		}
	}
	return nil
}
