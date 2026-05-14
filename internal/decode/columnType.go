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
	"encoding/binary"
	"fmt"
	"maps"

	"github.com/vesoft-inc/nebula-go/v5/pkg/types"
)

type (
	typeSchema interface {
		getType() types.ColumnType
	}

	columnTypeSchemaBasic struct {
		typ types.ColumnType
	}

	columnTypeSchemaList struct {
		typ       types.ColumnType
		subSchema typeSchema
	}
	columnTypeSchemaSet struct {
		typ       types.ColumnType
		subSchema typeSchema
	}
	columnTypeSchemaMap struct {
		typ         types.ColumnType
		keySchema   typeSchema
		valueSchema typeSchema
	}

	columnTypeSchemaVector struct {
		typ       types.ColumnType
		dim       uint32
		subSchema typeSchema
	}
	columnTypeSchemaRecord struct {
		typ         types.ColumnType
		propSchemas map[string]typeSchema
	}

	columnTypeSchemaElement struct {
		typ               types.ColumnType
		graphElementProps graphElementProps
	}

	columnTypeSchemaPath struct {
		typ        types.ColumnType
		nodeSchema *columnTypeSchemaElement
		edgeSchema *columnTypeSchemaElement
		meta       *pathMetaData
	}

	graphsSchema map[int32]*graphSchema
	graphSchema  struct {
		name        string
		id          int32
		nodesSchema map[int32]*elementSchema
		edgesSchema map[int32]*elementSchema
	}
	elementSchema struct {
		typeName string
		typeId   int32
		labels   []string
	}
)

// getType implements typeSchema.
func (c *columnTypeSchemaVector) getType() types.ColumnType {
	return c.typ
}

var _ typeSchema = (*columnTypeSchemaVector)(nil)

func newTypeSchema(r *bytesReader) (typeSchema, error) {
	b := r.readN(1)
	if r.error() != nil {
		return nil, r.error()
	}
	t, ok := columnTypeMap[b[0]]
	if !ok {
		return nil, fmt.Errorf("unknown column type: %d", b[0])
	}
	switch t {
	case types.ColumnTypeList:
		subSchema, err := newTypeSchema(r)
		if err != nil {
			return nil, err
		}
		typ := columnTypeSchemaList{
			typ:       t,
			subSchema: subSchema,
		}
		return &typ, nil
	case types.ColumnTypeSet:
		subSchema, err := newTypeSchema(r)
		if err != nil {
			return nil, err
		}
		typ := columnTypeSchemaSet{
			typ:       t,
			subSchema: subSchema,
		}
		return &typ, nil

	case types.ColumnTypeMap:
		keySchema, err := newTypeSchema(r)
		if err != nil {
			return nil, err
		}
		valueSchema, err := newTypeSchema(r)
		if err != nil {
			return nil, err
		}
		typ := columnTypeSchemaMap{
			typ:         t,
			keySchema:   keySchema,
			valueSchema: valueSchema,
		}
		return &typ, nil

	case types.ColumnTypeRecord:
		// filed num + [name + "0" + type]
		schema := make(map[string]typeSchema)
		numFieldsBytes := r.readN(4)
		if r.error() != nil {
			return nil, r.error()
		}
		numFields := bytesToInt32(numFieldsBytes)
		for range numFields {
			sizeBytes := r.readN(2)
			if r.error() != nil {
				return nil, r.error()
			}
			size := bytesToInt16(sizeBytes)
			nameBytes := r.readN(int(size))
			if r.error() != nil {
				return nil, r.error()
			}
			if r.error() != nil {
				return nil, r.error()
			}
			propSchema, err := newTypeSchema(r)
			if err != nil {
				return nil, err
			}
			schema[string(nameBytes)] = propSchema
		}
		typ := columnTypeSchemaRecord{
			typ:         t,
			propSchemas: schema,
		}
		return &typ, nil
	case types.ColumnTypeNode:
		pp, err := decodeElementTypes(r, true)
		if err != nil {
			return nil, err
		}
		typ := columnTypeSchemaElement{
			typ:               t,
			graphElementProps: pp,
		}
		return &typ, nil
	case types.ColumnTypeEdge:
		pp, err := decodeElementTypes(r, false)
		if err != nil {
			return nil, err
		}
		typ := columnTypeSchemaElement{
			typ:               t,
			graphElementProps: pp,
		}
		return &typ, nil
	case types.ColumnTypePath:
		// num of elements + [element type]
		typ := columnTypeSchemaPath{
			typ: t,
			nodeSchema: &columnTypeSchemaElement{
				typ:               types.ColumnTypeNode,
				graphElementProps: make(graphElementProps),
			},
			edgeSchema: &columnTypeSchemaElement{
				typ:               types.ColumnTypeEdge,
				graphElementProps: make(graphElementProps),
			},
		}
		elementNumBytes := r.readN(4)
		if r.error() != nil {
			return nil, r.error()
		}
		elementNum := bytesToInt32(elementNumBytes)
		for range elementNum {
			elementSchema, err := newTypeSchema(r)
			if err != nil {
				return nil, err
			}
			if elementSchema.getType() == types.ColumnTypeNode {
				s := elementSchema.(*columnTypeSchemaElement)
				for graphId, types := range s.graphElementProps {
					graph, ok := typ.nodeSchema.graphElementProps[graphId]
					if !ok {
						typ.nodeSchema.graphElementProps[graphId] = types
					} else {
						maps.Copy(graph, types)
					}
				}
			} else if elementSchema.getType() == types.ColumnTypeEdge {
				s := elementSchema.(*columnTypeSchemaElement)
				for graphId, types := range s.graphElementProps {
					graph, ok := typ.edgeSchema.graphElementProps[graphId]
					if !ok {
						typ.edgeSchema.graphElementProps[graphId] = types
					} else {
						maps.Copy(graph, types)
					}
				}
			} else {
				return nil, errInvalidColumnType
			}
		}
		return &typ, nil
	case types.ColumnTypeDecimal:
		// precision + scale
		_ = r.readN(2)
		_ = r.readN(2)
		if r.error() != nil {
			return nil, r.error()
		}
		return &columnTypeSchemaBasic{
			typ: t,
		}, nil
	case types.ColumnTypeVector:
		dim := binary.LittleEndian.Uint32(r.readN(4))
		subSchema, err := newTypeSchema(r)
		if err != nil {
			return nil, err
		}
		if subSchema.getType() != types.ColumnTypeFloat32 {
			return nil, errInvalidColumnType
		}
		typ := columnTypeSchemaVector{
			typ:       t,
			dim:       dim,
			subSchema: subSchema,
		}
		return &typ, nil
	case types.ColumnTypeGeography:
		return &columnTypeSchemaBasic{
			typ: t,
		}, nil
	default:
		return &columnTypeSchemaBasic{
			typ: t,
		}, nil
	}
}

func (s *columnTypeSchemaBasic) getType() types.ColumnType {
	return s.typ
}

func (s *columnTypeSchemaList) getType() types.ColumnType {
	return s.typ
}

func (s *columnTypeSchemaSet) getType() types.ColumnType {
	return s.typ
}

func (s *columnTypeSchemaMap) getType() types.ColumnType {
	return s.typ
}

func (s *columnTypeSchemaRecord) getType() types.ColumnType {
	return s.typ
}

func (s *columnTypeSchemaElement) getType() types.ColumnType {
	return s.typ
}

func (s *columnTypeSchemaElement) getElementProps() graphElementProps {
	return s.graphElementProps
}

func (s *columnTypeSchemaPath) getType() types.ColumnType {
	return s.typ
}

func (s *columnTypeSchemaPath) getNodeSchema() *columnTypeSchemaElement {
	return s.nodeSchema
}

func (s *columnTypeSchemaPath) getEdgeSchema() *columnTypeSchemaElement {
	return s.edgeSchema
}

func decodeElementTypes(r *bytesReader, isNode bool) (graphElementProps, error) {
	// num of element type + [graphid + element type id + num of props + [prop name + prop type]]
	// ignore the first byte nodeType
	elementTypeSize := 2
	if !isNode {
		elementTypeSize = 4
	}
	numElementTypeBytes := r.readN(4)
	if r.error() != nil {
		return nil, r.error()
	}
	numElementType := bytesToInt32(numElementTypeBytes)
	graphElementTypes := make(graphElementProps)
	for i := 0; i < int(numElementType); i++ {
		var elementTypeId int32
		graphIdTypes := r.readN(4)
		if r.error() != nil {
			return nil, r.error()
		}
		graphId := bytesToInt32(graphIdTypes)
		elementTypeIdBytes := r.readN(elementTypeSize)
		if r.error() != nil {
			return nil, r.error()
		}
		if isNode {
			elementTypeId = int32(bytesToInt16(elementTypeIdBytes))
		} else {
			elementTypeId = bytesToInt32(elementTypeIdBytes)
		}
		var graphType elementProps
		graphType, ok := graphElementTypes[graphId]
		if !ok {
			graphType = make(elementProps)
		}
		nt := make(props)
		numPropsBytes := r.readN(4)
		if r.error() != nil {
			return nil, r.error()
		}
		numProps := bytesToInt32(numPropsBytes)
		for j := 0; j < int(numProps); j++ {
			sizeBytes := r.readN(2)
			if r.error() != nil {
				return nil, r.error()
			}
			size := bytesToInt16(sizeBytes)
			propNameBytes := r.readN(int(size))
			if r.error() != nil {
				return nil, r.error()
			}
			if r.error() != nil {
				return nil, r.error()
			}
			prop, err := decodePropNameAndType(propNameBytes, r)
			if err != nil {
				return nil, err
			}
			nt[prop.name] = prop
		}
		graphType[elementTypeId] = nt
		graphElementTypes[graphId] = graphType
	}

	return graphElementTypes, nil
}

func decodePropNameAndType(name []byte, r *bytesReader) (*vectorProps, error) {
	var prps vectorProps

	prps.name = string(name)
	tt, err := newTypeSchema(r)
	if err != nil {
		return nil, err
	}
	prps.typ = tt
	return &prps, nil
}
