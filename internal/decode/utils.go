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
	"encoding/binary"
	"fmt"
	"math"

	"github.com/vesoft-inc/nebula-go/v5/pkg/errors"
	"github.com/vesoft-inc/nebula-go/v5/pkg/types"
)

var order = binary.LittleEndian

type bytesReader struct {
	bs  []byte
	err error
}

func newBytesReader(bs []byte) *bytesReader {
	return &bytesReader{
		bs: bs,
	}
}

func (r *bytesReader) readN(n int) []byte {
	if n > len(r.bs) {
		r.err = errors.Wrap(errOutOfRange, "")
		return nil
	}
	bs := r.bs[:n:n]
	r.bs = r.bs[n:]
	return bs
}

func (r *bytesReader) error() error {
	return r.err
}

func bytesToInt8(bs []byte) int8 {
	return int8(bs[0])
}

func bytesToInt16(bs []byte) int16 {
	return int16(order.Uint16(bs))
}

func bytesToInt32(bs []byte) int32 {
	return int32(order.Uint32(bs))
}

func bytesToInt64(bs []byte) int64 {
	return int64(order.Uint64(bs))
}

func bytesToUint8(bs []byte) uint8 {
	return uint8(bs[0])
}

func bytesToUint16(bs []byte) uint16 {
	return order.Uint16(bs)
}

func bytesToUint32(bs []byte) uint32 {
	return order.Uint32(bs)
}

func bytesToUint64(bs []byte) uint64 {
	return order.Uint64(bs)
}

func bytesToFloat64(bs []byte) float64 {
	return math.Float64frombits(order.Uint64(bs))
}

func isBasicColumnType(typ types.ColumnType) bool {
	switch typ {
	case types.ColumnTypeBool:
		fallthrough
	case types.ColumnTypeInt8, types.ColumnTypeInt16, types.ColumnTypeInt32, types.ColumnTypeInt64:
		fallthrough
	case types.ColumnTypeUint8, types.ColumnTypeUint16, types.ColumnTypeUint32, types.ColumnTypeUint64:
		fallthrough
	case types.ColumnTypeFloat32, types.ColumnTypeFloat64:
		fallthrough
	case types.ColumnTypeLocalTime, types.ColumnTypeLocalDatetime:
		fallthrough
	case types.ColumnTypeZonedTime, types.ColumnTypeZonedDatetime:
		fallthrough
	case types.ColumnTypeDate, types.ColumnTypeDuration:
		return true
	default:
		return false
	}
}

func getSchemaName(gsm graphsSchema, graphId int32, elementTypeId int32, isNode bool) (graphName, typeName string, labels []string, err error) {
	gs, ok := gsm[graphId]
	if !ok {
		return "", "", nil, errors.Wrap(errGraphNotFound, "")
	}
	var elementsSchema map[int32]*elementSchema
	if isNode {
		elementsSchema = gs.nodesSchema
	} else {
		elementsSchema = gs.edgesSchema
	}
	es, ok := elementsSchema[elementTypeId]
	if !ok {
		return "", "", nil, errors.Wrap(errElementTypeNotFount, fmt.Sprintf("element id: %d", elementTypeId))
	}
	return gs.name, es.typeName, es.labels, nil
}

func getEdgeDirection(d uint8) edgeDirection {
	switch d {
	case 0:
		return edgeOutGoingDirection
	case 1:
		return edgeInComingDirection
	default:
		return edgeNoDirection
	}
}

func decodeGeographyData(r *bytesReader) (*NebulaGeography, error) {
	// Read shape type (4 bytes)
	shapeBytes := r.readN(1)
	if r.error() != nil {
		return nil, r.error()
	}
	shapeType := bytesToInt8(shapeBytes)

	// Read SRID (4 bytes)
	sridBytes := r.readN(4)
	if r.error() != nil {
		return nil, r.error()
	}
	srid := bytesToInt32(sridBytes)

	geography := &NebulaGeography{
		SRID:  srid,
		Shape: types.GeoShape(shapeType),
	}

	switch types.GeoShape(shapeType) {
	case types.GeoShapePoint:
		xBytes := r.readN(8)
		yBytes := r.readN(8)
		if r.error() != nil {
			return nil, r.error()
		}
		x := bytesToFloat64(xBytes)
		y := bytesToFloat64(yBytes)
		geography.Point = &types.Point{Lng: x, Lat: y}

	case types.GeoShapeLineString:
		if len(r.bs) < 4 {
			return nil, errors.Wrap(errOutOfRange, "linestring data too short")
		}
		numCoordsBytes := r.readN(4)
		if r.error() != nil {
			return nil, r.error()
		}
		numCoords := bytesToInt32(numCoordsBytes)

		if len(r.bs) < int(numCoords)*16 {
			return nil, errors.Wrap(errOutOfRange, "linestring coordinates data too short")
		}

		coords := make([]*types.Point, 0, numCoords)
		for i := int32(0); i < numCoords; i++ {
			xBytes := r.readN(8)
			yBytes := r.readN(8)
			if r.error() != nil {
				return nil, r.error()
			}
			x := bytesToFloat64(xBytes)
			y := bytesToFloat64(yBytes)
			coords = append(coords, &types.Point{Lng: x, Lat: y})
		}
		geography.LineString = types.LineString(coords)

	case types.GeoShapePolygon:
		if len(r.bs) < 8 {
			return nil, errors.Wrap(errOutOfRange, "polygon data too short")
		}

		// Read number of row indexes
		numLoopBytes := r.readN(4)
		if r.error() != nil {
			return nil, r.error()
		}
		loops := bytesToInt32(numLoopBytes)

		// Read row indexes
		rowIndexes := make([]int32, 0, loops+1)
		for i := int32(0); i < loops+1; i++ {
			rowIndexBytes := r.readN(4)
			if r.error() != nil {
				return nil, r.error()
			}
			rowIndexes = append(rowIndexes, bytesToInt32(rowIndexBytes))
		}
		numCoords := rowIndexes[loops]
		if len(r.bs) < int(numCoords)*16 {
			return nil, errors.Wrap(errOutOfRange, "polygon coordinates data too short")
		}

		coords := make([]*types.Point, 0, numCoords)
		for i := int32(0); i < numCoords; i++ {
			xBytes := r.readN(8)
			yBytes := r.readN(8)
			if r.error() != nil {
				return nil, r.error()
			}
			x := bytesToFloat64(xBytes)
			y := bytesToFloat64(yBytes)
			coords = append(coords, &types.Point{Lng: x, Lat: y})
		}

		// Create loops from coordinates and row indexes
		polygonLoops := make([][]*types.Point, 0, loops)
		for i := int32(0); i < loops; i++ {
			start := int(rowIndexes[i])
			end := int(rowIndexes[i+1])
			if start < end && start < len(coords) && end <= len(coords) {
				loop := coords[start:end]
				polygonLoops = append(polygonLoops, loop)
			}
		}
		geography.Polygon = polygonLoops
	default:
		return nil, errors.Wrap(errInvalidColumnType, fmt.Sprintf("unsupported geography shape: %d", shapeType))
	}

	return geography, nil
}
