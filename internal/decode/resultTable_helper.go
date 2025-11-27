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
package decode

import (
	"math"
	"time"

	"github.com/vesoft-inc/nebula-go/v5/internal/generated_code/v5.0.0/proto/vector"
	internal_error "github.com/vesoft-inc/nebula-go/v5/internal/internal_error"
	"github.com/vesoft-inc/nebula-go/v5/pkg/errors"
	"github.com/vesoft-inc/nebula-go/v5/pkg/types"
)

var (
	errTypeAssertion       = internal_error.ErrInternal("type assertion failed")
	errOutOfRange          = internal_error.ErrInternal("index out of range")
	errInvalidColumnType   = internal_error.ErrInternal("invalid type")
	errInvalidVectorType   = internal_error.ErrInternal("invalid vector type")
	errNoZeroString        = internal_error.ErrInternal("no zero found in string")
	errElementTypeNotFount = internal_error.ErrInternal("element type not found")
	errGraphNotFound       = internal_error.ErrInternal("graph not found")
	errPropNotFound        = internal_error.ErrInternal("prop not found")
	errBatchNotEqual       = internal_error.ErrInternal("batch not equal")
	errColumnTypeIsNil     = internal_error.ErrInternal("column type is nil")
)

const (
	kMicrosecondsOfSecond = 1000000
	kMicrosecondsOfMinute = 60 * kMicrosecondsOfSecond
	kMicrosecondsOfHour   = 60 * kMicrosecondsOfMinute
	kMicrosecondsOfDay    = 24 * kMicrosecondsOfHour
)

var sizeMap = map[types.ColumnType]int{
	types.ColumnTypeBool:          1,
	types.ColumnTypeInt8:          1,
	types.ColumnTypeUint8:         1,
	types.ColumnTypeInt16:         2,
	types.ColumnTypeUint16:        2,
	types.ColumnTypeInt32:         4,
	types.ColumnTypeUint32:        4,
	types.ColumnTypeInt64:         8,
	types.ColumnTypeUint64:        8,
	types.ColumnTypeFloat32:       4,
	types.ColumnTypeFloat64:       8,
	types.ColumnTypeDate:          4,
	types.ColumnTypeLocalTime:     8,
	types.ColumnTypeZonedTime:     8,
	types.ColumnTypeLocalDatetime: 8,
	types.ColumnTypeZonedDatetime: 8,
	types.ColumnTypeDuration:      8,
}

var kOneBitmasks = []byte{
	1 << 0,
	1 << 1,
	1 << 2,
	1 << 3,
	1 << 4,
	1 << 5,
	1 << 6,
	1 << 7,
}

type decodeFlatFn func(dctx *decodeContext, value *NebulaValue, v *vector.NestedVector, index uint32, columnType typeSchema) error

type decoder interface {
	// decodeValue decodes the value from the vector
	decodeValue(dctx *decodeContext, value *NebulaValue, v *vector.NestedVector, t vectorType, index uint32, columnType typeSchema) error
}
type vectorDecoder struct {
	decodeFlatFns map[types.ColumnType]decodeFlatFn
}

var defaultDecoder decoder = &vectorDecoder{}

// for common value, the bytes size is fixed
// use a function to get the bytes
func (c *vectorDecoder) getCommonBytes(v *vector.NestedVector, typ types.ColumnType, index uint32) ([]byte, error) {
	size, ok := sizeMap[typ]
	if !ok {
		return nil, errors.Wrap(errInvalidColumnType, "")
	}
	if len(v.VectorData) < int(index)*size+size {
		return nil, errors.Wrap(errOutOfRange, "")
	}
	return v.VectorData[int(index)*size : int(index)*size+size], nil
}

func (c *vectorDecoder) decodeValue(dctx *decodeContext, value *NebulaValue, v *vector.NestedVector, t vectorType,
	index uint32, columnType typeSchema) error {
	if v.NullBitMap != nil {
		if v.NullBitMap[index/8]&kOneBitmasks[index%8] == 0 {
			value.Data = nil
			return nil
		}
	}
	switch t {
	case vectorTypeFlat:
		return c.decodeFlatValue(dctx, value, v, index, columnType)
	case vectorTypeConst:
		r := newBytesReader(v.VectorData)
		return c.decodeConstValue(dctx, value, r, columnType.getType())
	default:
		return errInvalidVectorType
	}
}

func (c *vectorDecoder) decodeFlatValue(dctx *decodeContext, value *NebulaValue, v *vector.NestedVector,
	index uint32, columnType typeSchema) error {
	typ := columnType.getType()
	if fn, ok := c.decodeFlatFns[typ]; ok {
		return fn(dctx, value, v, index, columnType)
	} else {
		return errors.Wrap(errInvalidColumnType, "")
	}
}

func (c *vectorDecoder) decodeConstValue(dctx *decodeContext, value *NebulaValue, r *bytesReader, columnType types.ColumnType) error {
	typ := columnType
	offset := dctx.timezoneOffset
	if isBasicColumnType(typ) {
		dataBytes := r.readN(sizeMap[typ])
		if r.error() != nil {
			return r.error()
		}
		decodeBasicValue(dataBytes, value, typ, offset)
		return nil
	} else {
		return decodeAnyCompositeValue(dctx, value, r, typ, true)
	}

}

func (c *vectorDecoder) decodeStringValue() decodeFlatFn {
	return func(dctx *decodeContext, value *NebulaValue, v *vector.NestedVector, index uint32, columnType typeSchema) error {
		// uint32 length + prefix string + uint32 chunk offset + uint32 chunk index
		// if the string is less then 12 bytes, no need to get data from chunk
		length := 4 + 4 + 4 + 4
		header := v.VectorData[index*uint32(length) : index*(uint32(length))+uint32(length)]
		strLen := bytesToUint32(header[:4])

		if strLen <= 12 {
			s := NebulaString(string(header[4 : 4+strLen]))
			value.Data = &s
		} else {
			chunkIndex := bytesToUint32(header[12:16])
			chunkOffset := bytesToUint32(header[8:12])
			chunk := v.NestedVectors[chunkIndex]
			data := chunk.VectorData[chunkOffset : chunkOffset+strLen]
			offset := dctx.timezoneOffset
			decodeBasicValue(data, value, types.ColumnTypeString, offset)
		}
		return nil
	}
}

func (c *vectorDecoder) decodeDecimalValue() decodeFlatFn {
	// same with string
	return func(dctx *decodeContext, value *NebulaValue, v *vector.NestedVector, index uint32, columnType typeSchema) error {
		// uint32 length + prefix string + uint32 chunk offset + uint32 chunk index
		// if the string is less then 12 bytes, no need to get data from chunk
		length := 4 + 4 + 4 + 4
		header := v.VectorData[index*uint32(length) : index*(uint32(length))+uint32(length)]
		strLen := bytesToUint32(header[:4])

		if strLen <= 12 {
			value.Data = &NebulaDecimal{Sval: string(header[4 : 4+strLen])}
		} else {
			chunkIndex := bytesToUint32(header[12:16])
			chunkOffset := bytesToUint32(header[8:12])
			chunk := v.NestedVectors[chunkIndex]
			data := chunk.VectorData[chunkOffset : chunkOffset+strLen]
			offset := dctx.timezoneOffset
			decodeBasicValue(data, value, types.ColumnTypeString, offset)
		}
		return nil
	}
}

func (c *vectorDecoder) decodeGeographyValue() decodeFlatFn {
	return func(dctx *decodeContext, value *NebulaValue, v *vector.NestedVector, index uint32, columnType typeSchema) error {
		length := 8 // chunkIndex(int32) + chunkOffset(int32)
		header := v.VectorData[index*uint32(length) : index*(uint32(length))+uint32(length)]
		chunkIndex := bytesToUint32(header[:4])
		chunkOffset := bytesToInt32(header[4:])
		chunk := v.NestedVectors[chunkIndex]
		buf := chunk.VectorData[chunkOffset:]
		r := newBytesReader(buf)
		g, err := decodeGeographyData(r)
		if err != nil {
			return err
		}
		value.Data = g
		return nil
	}
}

func (c *vectorDecoder) decodeBasicValue(t types.ColumnType) decodeFlatFn {
	return func(dctx *decodeContext, value *NebulaValue, v *vector.NestedVector, index uint32, columnType typeSchema) error {
		bs, err := c.getCommonBytes(v, t, index)
		offset := dctx.timezoneOffset
		if err != nil {
			return err
		}
		decodeBasicValue(bs, value, t, offset)
		return nil
	}
}

func (c *vectorDecoder) decodeVectorValue() decodeFlatFn {
	return func(dctx *decodeContext, value *NebulaValue, v *vector.NestedVector, index uint32, columnType typeSchema) error {
		vt, ok := columnType.(*columnTypeSchemaVector)
		if !ok {
			return errTypeAssertion
		}

		dim := int(vt.dim)
		offset := int(index) * dim * 4
		data := v.VectorData
		l := &NebulaEmbeddingVector{Values: make([]float32, 0, dim)}
		for i := 0; i < dim; i++ {
			currOff := offset + i*4
			val := math.Float32frombits(order.Uint32(data[currOff : currOff+4]))
			l.Values = append(l.Values, val)
		}
		value.Data = l
		return nil
	}
}

func (c *vectorDecoder) decodeListValue() decodeFlatFn {
	return func(dctx *decodeContext, value *NebulaValue, v *vector.NestedVector, index uint32, columnType typeSchema) error {
		// offset + size
		length := 4 + 4
		header := v.VectorData[index*uint32(length) : index*(uint32(length))+uint32(length)]
		offset := bytesToUint32(header[:4])
		size := bytesToUint32(header[4:8])
		if value.needReset(types.ValueTypeList) {
			value.Data = &NebulaList{}
		}
		l := value.Data.(*NebulaList)
		l.Values = constructListValue(l.Values, int(size))
		schema, ok := columnType.(*columnTypeSchemaList)
		if !ok {
			return errTypeAssertion
		}

		for i := uint32(0); i < size; i++ {
			vv := l.Values[i]
			dataVector := v.NestedVectors[0]
			wrapper := &vectorWrapper{
				vector:        dataVector,
				decodeContext: dctx,
				decoder:       defaultDecoder,
				vectorType:    vectorTypeFlat,
			}
			wrapper.typ = schema.subSchema
			if err := wrapper.prepare(); err != nil {
				return err
			}
			if err := c.decodeValue(dctx, vv, v.NestedVectors[0], vectorTypeFlat, offset+i, schema.subSchema); err != nil {
				return err
			}
		}
		value.Data = l
		return nil
	}
}

func (c *vectorDecoder) decodeRecordValue() decodeFlatFn {
	return func(dctx *decodeContext, value *NebulaValue, v *vector.NestedVector, index uint32, columnType typeSchema) error {
		r := newBytesReader(v.SpecialMetaData)
		schema, ok := columnType.(*columnTypeSchemaRecord)
		if !ok {
			return errTypeAssertion
		}
		if value.needReset(types.ValueTypeRecord) {
			value.Data = &NebulaRecord{}
		}
		propSchemas := schema.propSchemas
		nameIndexes := make([]string, 0, len(propSchemas))
		for i := 0; i < len(propSchemas); i++ {
			sizeBytes := r.readN(2)
			if r.error() != nil {
				return r.error()
			}
			size := bytesToInt16(sizeBytes)
			propNameBytes := r.readN(int(size))
			if r.error() != nil {
				return r.error()
			}
			nameIndexes = append(nameIndexes, string(propNameBytes))
		}
		record := value.Data.(*NebulaRecord)
		record.Values = make(map[string]*NebulaValue, len(propSchemas))
		for i := 0; i < len(propSchemas); i++ {
			n := nameIndexes[i]
			s, ok := propSchemas[n]
			if !ok {
				return errPropNotFound
			}
			vv := &NebulaValue{}
			dataVector := v.NestedVectors[i]
			wrapper := &vectorWrapper{
				vector:        dataVector,
				decodeContext: dctx,
				decoder:       defaultDecoder,
				vectorType:    vectorTypeFlat,
			}
			wrapper.typ = s
			if err := wrapper.prepare(); err != nil {
				return err
			}
			if err := c.decodeValue(dctx, vv, v.NestedVectors[i], vectorTypeFlat, index, s); err != nil {
				return err
			}
			record.Values[n] = vv
		}
		value.Data = record
		return nil
	}
}

func (c *vectorDecoder) decodeNullValue() decodeFlatFn {
	return func(dctx *decodeContext, value *NebulaValue, v *vector.NestedVector, index uint32, columnType typeSchema) error {
		value.Data = nil
		return nil
	}
}

func (c *vectorDecoder) decodeNodeValue() decodeFlatFn {
	return func(dctx *decodeContext, value *NebulaValue, v *vector.NestedVector, index uint32, columnType typeSchema) error {
		nodeSchema, ok := columnType.(*columnTypeSchemaElement)
		if !ok {
			return errTypeAssertion
		}
		allNodeProps := nodeSchema.graphElementProps
		// header = nodeID + graphId + padding
		length := 8 + 4 + 4
		header := v.VectorData[index*uint32(length) : index*(uint32(length))+uint32(length)]

		nodeID := bytesToInt64(header[:8])
		graphID := bytesToInt32(header[8:12])
		nodeTypeID := (int32)(nodeID >> 48)
		nodeType, ok := allNodeProps[graphID]
		if !ok {
			return errors.Wrap(errElementTypeNotFount, "")
		}
		nodeProps, ok := nodeType[nodeTypeID]
		if !ok {
			return errors.Wrap(errElementTypeNotFount, "")
		}
		gsm := dctx.graphsSchema
		graphName, typeName, labels, err := getSchemaName(gsm, graphID, nodeTypeID, true)
		if err != nil {
			return err
		}
		if value.needReset(types.ValueTypeNode) {
			value.Data = &NebulaNode{}
		}
		node := value.Data.(*NebulaNode)
		node.NodeId = nodeID
		node.Graph = graphName
		node.Type = typeName
		node.Labels = labels
		node.PropNames = constructListString(node.PropNames, len(nodeProps))
		node.PropValues = constructListValue(node.PropValues, len(nodeProps))
		loopIndex := 0
		for _, prop := range nodeProps {
			vw := v.NestedVectors[prop.vectorIndex]
			vv := node.PropValues[loopIndex]
			if err := c.decodeValue(dctx, vv, vw, vectorTypeFlat, index, prop.typ); err != nil {
				return err
			}
			// TODO: check if the prop name is valid
			node.PropNames[loopIndex] = prop.name
			loopIndex++
		}
		value.Data = node
		return nil
	}
}

func (c *vectorDecoder) decodeEdgeValue() decodeFlatFn {
	return func(dctx *decodeContext, value *NebulaValue, v *vector.NestedVector, index uint32, columnType typeSchema) error {
		schema, ok := columnType.(*columnTypeSchemaElement)
		if !ok {
			return errTypeAssertion
		}
		allGraphElementProps := schema.graphElementProps
		// header = src id + dst id + rank + graph id + edge type id
		length := 8 + 8 + 8 + 4 + 4
		header := v.VectorData[index*uint32(length) : index*(uint32(length))+uint32(length)]

		srcID := bytesToInt64(header[:8])
		dstID := bytesToInt64(header[8:16])
		rank := bytesToInt64(header[16:24])
		graphID := bytesToInt32(header[24:28])
		edgeTypeID := bytesToInt32(header[28:32])
		noDirectType := edgeTypeID & 0x3FFFFFFF
		direction := getEdgeDirection(uint8(edgeTypeID >> 30))
		edgeType, ok := allGraphElementProps[graphID]
		if !ok {
			return errors.Wrap(errElementTypeNotFount, "")
		}
		props, ok := edgeType[noDirectType]
		if !ok {
			return errors.Wrap(errElementTypeNotFount, "")
		}
		gsm := dctx.graphsSchema
		graphName, typeName, labels, err := getSchemaName(gsm, graphID, noDirectType, false)
		if err != nil {
			return err
		}
		if value.needReset(types.ValueTypeEdge) {
			value.Data = &NebulaEdge{}
		}
		e := value.Data.(*NebulaEdge)
		e.Direction = direction
		e.Rank = rank
		e.Graph = graphName
		e.Type = typeName
		e.Labels = labels
		e.PropNames = make([]string, len(props), len(props))
		e.PropValues = constructListValue(e.PropValues, len(props))

		switch direction {
		case edgeInComingDirection:
			e.SrcId = dstID
			e.DstId = srcID
		default:
			e.SrcId = srcID
			e.DstId = dstID
		}
		var loopIndex int
		for _, prop := range props {
			vw := v.NestedVectors[prop.vectorIndex]
			vv := e.PropValues[loopIndex]
			if err := c.decodeValue(dctx, vv, vw, vectorTypeFlat, index, prop.typ); err != nil {
				return err
			}
			e.PropNames[loopIndex] = prop.name
			loopIndex++
		}
		value.Data = e
		return nil
	}
}

func (c *vectorDecoder) decodePathValue() decodeFlatFn {
	return func(dctx *decodeContext, value *NebulaValue, v *vector.NestedVector, index uint32, columnType typeSchema) error {
		schema, ok := columnType.(*columnTypeSchemaPath)
		if !ok {
			return errTypeAssertion
		}

		meta := schema.meta
		nodeSchema := schema.nodeSchema
		edgeSchema := schema.edgeSchema
		// header = totalNum + headerIdx + tailIdx + headOffset + tailOffset
		length := 4 + 2 + 2 + 4 + 4
		header := v.VectorData[index*uint32(length) : index*(uint32(length))+uint32(length)]
		totalNum := bytesToInt32(header[:4])
		headerIdx := bytesToUint16(header[4:6])
		tailIdx := bytesToUint16(header[6:8])
		headOffset := bytesToUint32(header[8:12])
		tailOffset := bytesToUint32(header[12:16])
		_, _ = tailIdx, tailOffset
		if value.needReset(types.ValueTypePath) {
			value.Data = &NebulaPath{}
		}
		path := value.Data.(*NebulaPath)
		path.Values = constructListValue(path.Values, int(totalNum))
		if totalNum == 0 {
			return nil
		}
		var (
			dataVector           *vector.NestedVector
			adjVector            *vector.NestedVector
			pathHeaderValue      *NebulaValue
			pathHeaderValueInt64 types.Int64
			pathHeader           *pathAdjHeader
			err                  error
		)
		pairIndex := pathPairIndex(headerIdx)

		sentinelPair, ok := meta.nodeIndexPair[pairIndex]
		if !ok {
			return errors.Wrap(errElementTypeNotFount, "")
		}
		sentinelType := types.ColumnTypeNode
		sentinelOffset := headOffset
		pathHeaderSchema := &columnTypeSchemaBasic{
			typ: types.ColumnTypeInt64,
		}
		for i := int32(0); i < totalNum; i++ {
			data := path.Values[i]
			dataVector = sentinelPair.cur
			adjVector = sentinelPair.adj

			wrapper := &vectorWrapper{
				vector:        dataVector,
				decodeContext: dctx,
				decoder:       defaultDecoder,
				vectorType:    vectorTypeFlat,
			}
			if sentinelType == types.ColumnTypeNode {
				wrapper.typ = nodeSchema
				data.Data = &NebulaNode{}
				if err := wrapper.prepare(); err != nil {
					return err
				}
				err = c.decodeValue(dctx, data, dataVector, vectorTypeFlat, sentinelOffset, nodeSchema)
			} else {
				wrapper.typ = edgeSchema
				data.Data = &NebulaEdge{}
				if err := wrapper.prepare(); err != nil {
					return err
				}
				err = c.decodeValue(dctx, data, dataVector, vectorTypeFlat, sentinelOffset, edgeSchema)
			}
			if err != nil {
				return err
			}

			pathHeaderValue = &NebulaValue{}
			if err = c.decodeValue(dctx, pathHeaderValue, adjVector, vectorTypeFlat, sentinelOffset, pathHeaderSchema); err != nil {
				return err
			}
			pathHeaderValueInt64, err = pathHeaderValue.AsInt64()
			if err != nil {
				return err
			}
			pathHeader = newPathAdjHeader(int64(pathHeaderValueInt64))
			sentinelOffset = pathHeader.nextOffset
			if pathHeader.nextIsEdge {
				sentinelType = types.ColumnTypeEdge
			} else {
				sentinelType = types.ColumnTypeNode
			}
			if sentinelType == types.ColumnTypeNode {
				sentinelPair, ok = meta.nodeIndexPair[pathPairIndex(pathHeader.nextVectorIndex)]
				if !ok {
					return errors.Wrap(errElementTypeNotFount, "")
				}
			} else {
				sentinelPair, ok = meta.edgeIndexPair[pathPairIndex(pathHeader.nextVectorIndex)]
				if !ok {
					return errors.Wrap(errElementTypeNotFount, "")
				}
			}
		}
		value.Data = path
		return nil
	}
}

func (c *vectorDecoder) decodeSetValue() decodeFlatFn {
	return func(dctx *decodeContext, value *NebulaValue, v *vector.NestedVector, index uint32, columnType typeSchema) error {
		// offset + size
		length := 4 + 4
		header := v.VectorData[index*uint32(length) : index*(uint32(length))+uint32(length)]
		offset := bytesToUint32(header[:4])
		size := bytesToUint32(header[4:8])
		if value.needReset(types.ValueTypeSet) {
			value.Data = &NebulaSet{}
		}
		l := value.Data.(*NebulaSet)
		l.Values = constructListValue(l.Values, int(size))
		schema, ok := columnType.(*columnTypeSchemaSet)
		if !ok {
			return errTypeAssertion
		}
		for i := uint32(0); i < size; i++ {
			vv := l.Values[i]
			dataVector := v.NestedVectors[0]
			wrapper := &vectorWrapper{
				vector:        dataVector,
				decodeContext: dctx,
				decoder:       defaultDecoder,
				vectorType:    vectorTypeFlat,
			}
			wrapper.typ = schema.subSchema
			if err := wrapper.prepare(); err != nil {
				return err
			}
			if err := c.decodeValue(dctx, vv, v.NestedVectors[0], vectorTypeFlat, offset+i, schema.subSchema); err != nil {
				return err
			}
		}
		value.Data = l
		return nil
	}
}

func (c *vectorDecoder) decodeMapValue() decodeFlatFn {
	return func(dctx *decodeContext, value *NebulaValue, v *vector.NestedVector, index uint32, columnType typeSchema) error {
		// offset + size
		length := 4 + 4
		header := v.VectorData[index*uint32(length) : index*(uint32(length))+uint32(length)]
		offset := bytesToUint32(header[:4])
		size := bytesToUint32(header[4:8])
		if value.needReset(types.ValueTypeMap) {
			value.Data = &NebulaMap{}
		}
		m := value.Data.(*NebulaMap)
		m.Values = make(map[*NebulaValue]*NebulaValue)
		schema, ok := columnType.(*columnTypeSchemaMap)
		if !ok {
			return errTypeAssertion
		}
		for i := uint32(0); i < size; i++ {
			key := &NebulaValue{}
			val := &NebulaValue{}
			keyDataVector := v.NestedVectors[0]
			wrapperKey := &vectorWrapper{
				vector:        keyDataVector,
				decodeContext: dctx,
				decoder:       defaultDecoder,
				vectorType:    vectorTypeFlat,
			}
			wrapperKey.typ = schema.keySchema
			if err := wrapperKey.prepare(); err != nil {
				return err
			}
			valDataVector := v.NestedVectors[1]
			wrapperVal := &vectorWrapper{
				vector:        valDataVector,
				decodeContext: dctx,
				decoder:       defaultDecoder,
				vectorType:    vectorTypeFlat,
			}
			wrapperVal.typ = schema.valueSchema
			if err := wrapperVal.prepare(); err != nil {
				return err
			}
			if err := c.decodeValue(dctx, key, v.NestedVectors[0], vectorTypeFlat, offset+i, schema.keySchema); err != nil {
				return err
			}
			if err := c.decodeValue(dctx, val, v.NestedVectors[1], vectorTypeFlat, offset+i, schema.valueSchema); err != nil {
				return err
			}
			m.Values[key] = val
		}
		value.Data = m
		return nil
	}
}

func (c *vectorDecoder) decodeAnyValue() decodeFlatFn {
	return func(dctx *decodeContext, value *NebulaValue, v *vector.NestedVector, index uint32, columnType typeSchema) error {
		dataTypeVector := v.NestedVectors[0]
		typeLength := uint32(1)
		dateHeaderLength := uint32(8)
		typeBytes := dataTypeVector.VectorData[typeLength*index : typeLength*(index+1)]
		dataType, ok := columnTypeMap[typeBytes[0]]
		offset := dctx.timezoneOffset
		if !ok {
			return errors.Wrap(errInvalidColumnType, "")
		}
		dataBytes := v.VectorData[dateHeaderLength*index : dateHeaderLength*(index+1)]
		if isBasicColumnType(dataType) {
			v := &NebulaValue{}
			decodeBasicValue(dataBytes, v, dataType, offset)
			value.Data = v.Data
			return nil
		}
		chunkIndexBytes, chunkOffsetBytes := dataBytes[:4], dataBytes[4:8]
		chunkIndex := bytesToUint32(chunkIndexBytes)
		chunkOffset := bytesToUint32(chunkOffsetBytes)
		r := newBytesReader(v.NestedVectors[chunkIndex+1].VectorData)
		r.readN(int(chunkOffset))
		return decodeAnyCompositeValue(dctx, value, r, dataType, true)
	}
}

func decodeBasicValue(bs []byte, value *NebulaValue, typ types.ColumnType, offset int64) {
	switch typ {
	case types.ColumnTypeBool:
		data := NebulaBool(bs[0] == 1)
		value.Data = &data
	case types.ColumnTypeInt8:
		data := NebulaInt8(bytesToInt8(bs[:1]))
		value.Data = &data
	case types.ColumnTypeInt16:
		data := NebulaInt16(bytesToInt16(bs[:2]))
		value.Data = &data
	case types.ColumnTypeInt32:
		data := NebulaInt32(bytesToInt32(bs[:4]))
		value.Data = &data
	case types.ColumnTypeInt64:
		data := NebulaInt64(bytesToInt64(bs[:8]))
		value.Data = &data
	case types.ColumnTypeUint8:
		data := NebulaUint8(bytesToUint8(bs[:1]))
		value.Data = &data
	case types.ColumnTypeUint16:
		data := NebulaUint16(bytesToUint16(bs[:2]))
		value.Data = &data
	case types.ColumnTypeUint32:
		data := NebulaUint32(bytesToUint32(bs[:4]))
		value.Data = &data
	case types.ColumnTypeUint64:
		data := NebulaUint64(bytesToUint64(bs[:8]))
		value.Data = &data
	case types.ColumnTypeFloat32:
		data := NebulaFloat(math.Float32frombits(order.Uint32(bs[:4])))
		value.Data = &data
	case types.ColumnTypeFloat64:
		data := NebulaDouble(math.Float64frombits(order.Uint64(bs[:8])))
		value.Data = &data
	case types.ColumnTypeString:
		data := NebulaString(string(bs))
		value.Data = &data
	case types.ColumnTypeDecimal:
		if value.needReset(types.ValueTypeDecimal) {
			value.Data = &NebulaDecimal{}
		}
		data := value.Data.(*NebulaDecimal)
		data.Sval = string(bs)
	case types.ColumnTypeDate:
		if value.needReset(types.ValueTypeDate) {
			value.Data = &NebulaDate{}
		}
		data := value.Data.(*NebulaDate)
		year := int16(bytesToInt16(bs[:2]))
		month := bytesToInt8(bs[2:3])
		day := bytesToInt8(bs[3:4])
		data.Year = year
		data.Month = month
		data.Day = day

	case types.ColumnTypeLocalTime:
		if value.needReset(types.ValueTypeLocalTime) {
			value.Data = &NebulaLocalTime{}
		}
		data := value.Data.(*NebulaLocalTime)

		hour := bytesToInt8(bs[:1])
		minute := bytesToInt8(bs[1:2])
		second := bytesToInt8(bs[2:3])
		// padding
		microsecond := bytesToInt32(bs[4:8])
		data.Hour = hour
		data.Minute = minute
		data.Sec = second
		data.Microsec = microsecond
	case types.ColumnTypeZonedTime:
		if value.needReset(types.ValueTypeZonedTime) {
			value.Data = &NebulaZonedTime{}
		}
		data := value.Data.(*NebulaZonedTime)
		hour := bytesToInt8(bs[:1])
		minute := bytesToInt8(bs[1:2])
		second := bytesToInt8(bs[2:3])
		// padding
		microsecond := bytesToInt32(bs[4:8])
		n := time.Now()
		var h int
		if hour < 0 {
			h = -int(hour)
		} else {
			h = int(hour)
		}
		t := time.Date(n.Year(), n.Month(), n.Day(), h, int(minute), int(second), int(microsecond)*int(time.Microsecond), time.UTC)
		zonedT := t.In(time.FixedZone("", int(offset)))
		data.Hour = int8(zonedT.Hour())
		data.Minute = int8(zonedT.Minute())
		data.Sec = int8(zonedT.Second())
		data.Microsec = int32(zonedT.Nanosecond() / int(time.Microsecond))
		data.Offset = int32(offset)

	case types.ColumnTypeLocalDatetime:
		if value.needReset(types.ValueTypeLocalDateTime) {
			value.Data = &NebulaLocalDatetime{}
		}
		data := value.Data.(*NebulaLocalDatetime)
		qword := bytesToInt64(bs)
		year := int16(qword & 0xffff)
		month := int8(qword >> 16 & 0xf)
		day := int8(qword >> 20 & 0x1f)
		hour := int8(qword >> 25 & 0x1f)
		minute := int8(qword >> 30 & 0x3f)
		second := int8(qword >> 36 & 0x3f)
		microsecond := int32(qword >> 42 & 0x3ffffff)
		data.Year = year
		data.Month = month
		data.Day = day
		data.Hour = hour
		data.Minute = minute
		data.Sec = second
		data.Microsec = microsecond

	case types.ColumnTypeZonedDatetime:
		if value.needReset(types.ValueTypeZonedDateTime) {
			value.Data = &NebulaZonedDatetime{}
		}
		data := value.Data.(*NebulaZonedDatetime)
		qword := bytesToInt64(bs)
		year := int16(qword & 0xffff)
		month := int8(qword >> 16 & 0xf)
		day := int8(qword >> 20 & 0x1f)
		hour := int8(qword >> 25 & 0x1f)
		minute := int8(qword >> 30 & 0x3f)
		second := int8(qword >> 36 & 0x3f)
		microsecond := int32(qword >> 42 & 0x3ffffff)
		t := time.Date(int(year), time.Month(month), int(day), int(hour), int(minute),
			int(second), int(microsecond)*int(time.Microsecond), time.UTC)
		zonedT := t.In(time.FixedZone("", int(offset)))
		data.Year = int16(zonedT.Year())
		data.Month = int8(zonedT.Month())
		data.Day = int8(zonedT.Day())
		data.Hour = int8(zonedT.Hour())
		data.Minute = int8(zonedT.Minute())
		data.Sec = int8(zonedT.Second())
		data.Microsec = int32(zonedT.Nanosecond() / int(time.Microsecond))
		data.Offset = int32(offset)
	case types.ColumnTypeDuration:
		if value.needReset(types.ValueTypeDuration) {
			value.Data = &NebulaDuration{}
		}
		data := value.Data.(*NebulaDuration)
		value := bytesToInt64(bs)
		isMonthBased := value&0x1 == 1
		value >>= 1
		var (
			year        int64
			month       int8
			day         int32
			hour        int8
			minute      int8
			second      int8
			microsecond int32
		)
		if isMonthBased {
			year = value / 12
			month = int8(value % 12)
		} else {
			day = int32(value / kMicrosecondsOfDay)
			hour = int8(value % kMicrosecondsOfDay / kMicrosecondsOfHour)
			minute = int8(value % kMicrosecondsOfHour / kMicrosecondsOfMinute)
			second = int8(value % kMicrosecondsOfMinute / kMicrosecondsOfSecond)
			microsecond = int32(value % kMicrosecondsOfSecond)
		}
		data.MonthBased = isMonthBased
		data.Year = year
		data.Month = month
		data.Day = day
		data.Hour = hour
		data.Minute = minute
		data.Sec = second
		data.Microsec = microsecond

	default:
		value.Data = nil
	}
}

func decodeAnyCompositeValue(dctx *decodeContext, value *NebulaValue, r *bytesReader, typ types.ColumnType, first bool) error {
	if first {
		// first byte is the type
		if isBasicColumnType(typ) {
			return errors.Wrap(errInvalidColumnType, "")
		}
	}
	offset := dctx.timezoneOffset
	gsm := dctx.graphsSchema
	if isBasicColumnType(typ) {
		size, ok := sizeMap[typ]
		if !ok {
			return errors.Wrap(errInvalidColumnType, "")
		}
		bs := r.readN(size)
		if r.error() != nil {
			return r.error()
		}
		v := &NebulaValue{}
		decodeBasicValue(bs, v, typ, offset)
		value.Data = v
		return nil
	}
	switch typ {
	case types.ColumnTypeUnknown:
		value.Data = nil
	case types.ColumnTypeString, types.ColumnTypeDecimal:
		sizeBytes := r.readN(2)
		if r.error() != nil {
			return r.error()
		}
		size := int(bytesToUint16(sizeBytes))
		bs := r.readN(size)
		if r.error() != nil {
			return r.error()
		}
		if typ == types.ColumnTypeDecimal {
			value.Data = &NebulaDecimal{Sval: string(bs)}
		} else {
			s := NebulaString(string(bs))
			value.Data = &s
		}
	case types.ColumnTypeList:
		typeBytes := r.readN(1)
		sizeBytes := r.readN(2)
		if r.error() != nil {
			return r.error()
		}
		subType, ok := columnTypeMap[typeBytes[0]]
		if !ok {
			return errors.Wrap(errInvalidColumnType, "")
		}
		size := int(bytesToUint16(sizeBytes))
		l := make([]*NebulaValue, 0, size)
		var bitSize int
		if size%8 != 0 {
			bitSize = size/8 + 1
		} else {
			bitSize = size / 8
		}
		nullBitByte := r.readN(bitSize)
		if r.error() != nil {
			return r.error()
		}
		if r.error() != nil {
			return r.error()
		}
		for i := 0; i < size; i++ {
			if nullBitByte[i/8]&(1<<(i%8)) == 0 {
				l = append(l, &NebulaValue{Data: nil})
			} else {
				v := &NebulaValue{}
				if err := decodeAnyCompositeValue(dctx, v, r, subType, false); err != nil {
					return err
				}
				l = append(l, v)
			}
		}
		value.Data = &NebulaList{
			Values: l,
		}
	case types.ColumnTypeSet:
		typeBytes := r.readN(1)
		sizeBytes := r.readN(4)
		if r.error() != nil {
			return r.error()
		}
		subType, ok := columnTypeMap[typeBytes[0]]
		if !ok {
			return errors.Wrap(errInvalidColumnType, "")
		}
		size := int(bytesToUint32(sizeBytes))
		l := make([]*NebulaValue, 0, size)
		var bitSize int
		if size%8 != 0 {
			bitSize = size/8 + 1
		} else {
			bitSize = size / 8
		}
		nullBitByte := r.readN(bitSize)
		if r.error() != nil {
			return r.error()
		}
		if r.error() != nil {
			return r.error()
		}
		for i := 0; i < size; i++ {
			if nullBitByte[i/8]&(1<<(i%8)) == 0 {
				l = append(l, &NebulaValue{Data: nil})
			} else {
				v := &NebulaValue{}
				if err := decodeAnyCompositeValue(dctx, v, r, subType, false); err != nil {
					return err
				}
				l = append(l, v)
			}
		}
		value.Data = &NebulaSet{
			Values: l,
		}
	case types.ColumnTypeMap:
		kValue := &NebulaValue{}
		vValue := &NebulaValue{}
		if err := decodeAnyCompositeValue(dctx, kValue, r, types.ColumnTypeSet, false); err != nil {
			return err
		}
		if err := decodeAnyCompositeValue(dctx, vValue, r, types.ColumnTypeSet, false); err != nil {
			return err
		}
		keys := kValue.Data.(*NebulaSet)
		values := vValue.Data.(*NebulaSet)
		data := &NebulaMap{}
		data.Values = make(map[*NebulaValue]*NebulaValue)
		if len(keys.Values) != len(values.Values) {
			return internal_error.ErrDecodeFailed("map key size not equal to value size")
		}
		for i := 0; i < len(keys.Values); i++ {
			data.Values[keys.Values[i]] = values.Values[i]
		}
		value.Data = data
	case types.ColumnTypeRecord:
		sizeBytes := r.readN(2)
		size := int(bytesToUint16(sizeBytes))
		m := make(map[string]*NebulaValue, 0)
		for i := 0; i < size; i++ {
			bs := r.readN(2)
			if r.error() != nil {
				return r.error()
			}
			nameLength := int(bytesToInt16(bs))
			nameBytes := r.readN(nameLength)
			typeBytes := r.readN(1)
			if r.error() != nil {
				return r.error()
			}
			subType, ok := columnTypeMap[typeBytes[0]]
			if !ok {
				return errors.Wrap(errInvalidColumnType, "")
			}
			v := &NebulaValue{}
			if err := decodeAnyCompositeValue(dctx, v, r, subType, false); err != nil {
				return err
			}
			m[string(nameBytes)] = v
		}
		value.Data = &NebulaRecord{
			Values: m,
		}
	case types.ColumnTypeNode:
		// nodeID 8B + graphId 4B + prop_Size 2B
		nodeIdBytes, graphIdBytes, propSizeBytes := r.readN(8), r.readN(4), r.readN(2)
		if r.error() != nil {
			return r.error()
		}
		nodeId := bytesToInt64(nodeIdBytes)
		nodeTypeId := int32(nodeId >> 48)
		graphId := bytesToInt32(graphIdBytes)
		_ = graphId
		propSize := bytesToUint16(propSizeBytes)
		keys := make([]string, 0, propSize)
		values := make([]*NebulaValue, 0, propSize)
		for i := 0; i < int(propSize); i++ {
			nameSizeBytes := r.readN(2)
			if r.error() != nil {
				return r.error()
			}
			nameSize := int(bytesToInt16(nameSizeBytes))
			nameBytes := r.readN(nameSize)
			typeBytes := r.readN(1)
			if r.error() != nil {
				return r.error()
			}
			subType, ok := columnTypeMap[typeBytes[0]]
			if !ok {
				return errors.Wrap(errInvalidColumnType, "")
			}
			v := &NebulaValue{}
			if err := decodeAnyCompositeValue(dctx, v, r, subType, false); err != nil {
				return err
			}
			k := string(nameBytes)
			keys = append(keys, k)
			values = append(values, v)
		}
		graphName, typeName, labels, err := getSchemaName(gsm, graphId, nodeTypeId, true)
		if err != nil {
			return err
		}
		value.Data = &NebulaNode{
			NodeId:     nodeId,
			Graph:      graphName,
			Type:       typeName,
			Labels:     labels,
			PropNames:  keys,
			PropValues: values,
		}
	case types.ColumnTypeEdge:
		// src nodeID 8B + dst nodeID 8B + edge rank 8B + graphId 4B + edge type ID 4B  + prop_size 2B
		srcNodeIDBytes, dstNodeIDBytes, edgeRankBytes, graphIdBytes, edgeTypeIdBytes,
			propSizeBytes := r.readN(8), r.readN(8), r.readN(8), r.readN(4), r.readN(4), r.readN(2)
		if r.error() != nil {
			return r.error()
		}
		srcNodeId := bytesToInt64(srcNodeIDBytes)
		dstNodeId := bytesToInt64(dstNodeIDBytes)
		graphId := bytesToInt32(graphIdBytes)
		edgeTypeID := bytesToInt32(edgeTypeIdBytes)
		edgeRank := bytesToInt64(edgeRankBytes)
		propSize := bytesToUint16(propSizeBytes)
		propNames := make([]string, 0, propSize)
		propValues := make([]*NebulaValue, 0, propSize)
		noDirectType := edgeTypeID & 0x3FFFFFFF
		direction := getEdgeDirection(uint8(edgeTypeID >> 30))
		for i := 0; i < int(propSize); i++ {
			nameSizeBytes := r.readN(2)
			if r.error() != nil {
				return r.error()
			}
			nameSize := int(bytesToInt16(nameSizeBytes))
			nameBytes := r.readN(nameSize)
			typeBytes := r.readN(1)
			if r.error() != nil {
				return r.error()
			}
			subType, ok := columnTypeMap[typeBytes[0]]
			if !ok {
				return errors.Wrap(errInvalidColumnType, "")
			}
			v := &NebulaValue{}
			if err := decodeAnyCompositeValue(dctx, v, r, subType, false); err != nil {
				return err
			}
			propNames = append(propNames, string(nameBytes))
			propValues = append(propValues, v)
		}
		graphName, typeName, labels, err := getSchemaName(gsm, graphId, noDirectType, false)
		if err != nil {
			return errors.Wrap(err, "")
		}
		e := &NebulaEdge{
			Rank:       edgeRank,
			Graph:      graphName,
			Type:       typeName,
			Labels:     labels,
			PropNames:  propNames,
			PropValues: propValues,
			Direction:  direction,
		}
		switch direction {
		case edgeInComingDirection:
			e.SrcId = dstNodeId
			e.DstId = srcNodeId
		default:
			e.SrcId = srcNodeId
			e.DstId = dstNodeId
		}
		value.Data = e
	case types.ColumnTypePath:
		elementNumBytes := r.readN(2)
		if r.error() != nil {
			return r.error()
		}
		elementNum := int(bytesToInt16(elementNumBytes))
		p := &NebulaPath{
			Values: make([]*NebulaValue, 0, elementNum),
		}
		for i := 0; i < elementNum; i++ {
			subTypeBytes := r.readN(1)
			if r.error() != nil {
				return r.error()
			}
			subType, ok := columnTypeMap[subTypeBytes[0]]
			if !ok {
				return errors.Wrap(errInvalidColumnType, "")
			}
			element := &NebulaValue{}
			if err := decodeAnyCompositeValue(dctx, element, r, subType, false); err != nil {
				return err
			}
			p.Values = append(p.Values, element)
		}
		value.Data = p
	case types.ColumnTypeVector:
		sizeBytes := r.readN(2)
		if r.error() != nil {
			return r.error()
		}
		size := int(bytesToInt16(sizeBytes))
		l := &NebulaEmbeddingVector{Values: make([]float32, 0, size)}
		for i := 0; i < size; i++ {
			fval := math.Float32frombits(order.Uint32(r.readN(4)))
			l.Values = append(l.Values, fval)
		}
		value.Data = l
	case types.ColumnTypeGeography:
		geography, err := decodeGeographyData(r)
		if err != nil {
			return err
		}

		value.Data = geography
	default:
		return errors.Wrap(errInvalidColumnType, "")
	}
	return nil
}

func init() {
	d := defaultDecoder.(*vectorDecoder)
	d.decodeFlatFns = make(map[types.ColumnType]decodeFlatFn)
	d.decodeFlatFns = map[types.ColumnType]decodeFlatFn{
		types.ColumnTypeBool:          d.decodeBasicValue(types.ColumnTypeBool),
		types.ColumnTypeInt8:          d.decodeBasicValue(types.ColumnTypeInt8),
		types.ColumnTypeInt16:         d.decodeBasicValue(types.ColumnTypeInt16),
		types.ColumnTypeInt32:         d.decodeBasicValue(types.ColumnTypeInt32),
		types.ColumnTypeInt64:         d.decodeBasicValue(types.ColumnTypeInt64),
		types.ColumnTypeUint8:         d.decodeBasicValue(types.ColumnTypeUint8),
		types.ColumnTypeUint16:        d.decodeBasicValue(types.ColumnTypeUint16),
		types.ColumnTypeUint32:        d.decodeBasicValue(types.ColumnTypeUint32),
		types.ColumnTypeUint64:        d.decodeBasicValue(types.ColumnTypeUint64),
		types.ColumnTypeFloat32:       d.decodeBasicValue(types.ColumnTypeFloat32),
		types.ColumnTypeFloat64:       d.decodeBasicValue(types.ColumnTypeFloat64),
		types.ColumnTypeString:        d.decodeStringValue(),
		types.ColumnTypeNode:          d.decodeNodeValue(),
		types.ColumnTypeEdge:          d.decodeEdgeValue(),
		types.ColumnTypePath:          d.decodePathValue(),
		types.ColumnTypeUnknown:       d.decodeNullValue(),
		types.ColumnTypeList:          d.decodeListValue(),
		types.ColumnTypeRecord:        d.decodeRecordValue(),
		types.ColumnTypeLocalTime:     d.decodeBasicValue(types.ColumnTypeLocalTime),
		types.ColumnTypeLocalDatetime: d.decodeBasicValue(types.ColumnTypeLocalDatetime),
		types.ColumnTypeZonedTime:     d.decodeBasicValue(types.ColumnTypeZonedTime),
		types.ColumnTypeZonedDatetime: d.decodeBasicValue(types.ColumnTypeZonedDatetime),
		types.ColumnTypeDate:          d.decodeBasicValue(types.ColumnTypeDate),
		types.ColumnTypeDuration:      d.decodeBasicValue(types.ColumnTypeDuration),
		types.ColumnTypeDecimal:       d.decodeDecimalValue(),
		types.ColumnTypeVector:        d.decodeVectorValue(),
		types.ColumnTypeGeography:     d.decodeGeographyValue(),
		types.ColumnTypeSet:           d.decodeSetValue(),
		types.ColumnTypeMap:           d.decodeMapValue(),
		types.ColumnTypeAny:           d.decodeAnyValue(),
	}
}
