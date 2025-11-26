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
	"testing"

	"github.com/vesoft-inc/nebula-go/v5/pkg/types"
)

func TestValueString(t *testing.T) {
	var (
		boolTrue     = NebulaBool(true)
		boolFalse    = NebulaBool(false)
		int8Val      = NebulaInt8(8)
		int8Val2     = NebulaInt8(9)
		int16Val     = NebulaInt16(16)
		floatVal     = NebulaFloat(16.02)
		float32Max   = NebulaFloat(float32(math.Inf(1)))
		float32Min   = NebulaFloat(float32(math.Inf(-1)))
		float64Val   = NebulaDouble(16.02)
		float64Max   = NebulaDouble(math.Inf(1))
		float64Min   = NebulaDouble(math.Inf(-1))
		float64NaN   = NebulaDouble(math.NaN())
		stringVal    = NebulaString("string")
		stringVal2   = NebulaString("dec")
		stringVal3   = NebulaString("abc")
		stringWithCR = NebulaString("abc\rde")
	)
	testcases := []struct {
		name   string
		value  *NebulaValue
		expect string
	}{
		{
			name:   "nil",
			value:  &NebulaValue{Data: nil},
			expect: "null",
		},
		{
			name:   "bool",
			value:  &NebulaValue{Data: &boolTrue},
			expect: "true",
		},
		{
			name:   "bool",
			value:  &NebulaValue{Data: &boolFalse},
			expect: "false",
		},
		{
			name:   "int8",
			value:  &NebulaValue{Data: &int8Val},
			expect: "8",
		},
		{
			name:   "int16",
			value:  &NebulaValue{Data: &int16Val},
			expect: "16",
		},
		{
			name:   "float",
			value:  &NebulaValue{Data: &floatVal},
			expect: "16.02",
		},
		{
			name:   "float",
			value:  &NebulaValue{Data: &float32Max},
			expect: "+Inf",
		},
		{
			name:   "float",
			value:  &NebulaValue{Data: &float32Min},
			expect: "-Inf",
		},
		{
			name:   "double",
			value:  &NebulaValue{Data: &float64Val},
			expect: "16.02",
		},
		{
			name:   "double",
			value:  &NebulaValue{Data: &float64Max},
			expect: "+Inf",
		},
		{
			name:   "double",
			value:  &NebulaValue{Data: &float64Min},
			expect: "-Inf",
		},
		{
			name:   "double",
			value:  &NebulaValue{Data: &float64NaN},
			expect: "NaN",
		},
		{
			name:   "string",
			value:  &NebulaValue{Data: &stringVal},
			expect: `string`,
		},
		{
			name:   "string",
			value:  &NebulaValue{Data: &stringWithCR},
			expect: "abc\rde",
		},
		{
			name: "list",
			value: &NebulaValue{Data: &NebulaList{
				Values: []*NebulaValue{
					{Data: &int8Val},
					{Data: &stringVal2},
				},
			}},
			expect: `[8,dec]`,
		},
		{
			name: "record",
			value: &NebulaValue{Data: &NebulaRecord{
				Values: map[string]*NebulaValue{
					"int": {Data: &int8Val},
					"str": {Data: &stringVal2},
				},
			}},
			expect: `{int:8,str:dec}`,
		},
		{
			name: "node",
			value: &NebulaValue{Data: &NebulaNode{
				NodeId:    1,
				PropNames: []string{"int", "str"},
				PropValues: []*NebulaValue{
					{Data: &int8Val},
					{Data: &stringVal2},
				},
				Type:   "node",
				Labels: []string{"label1", "label2"},
			}},
			expect: `(1@node:label1&label2{int:8,str:dec})`,
		},
		{
			name: "edge",
			value: &NebulaValue{Data: &NebulaEdge{
				SrcId:     1,
				DstId:     2,
				PropNames: []string{"int", "str"},
				PropValues: []*NebulaValue{
					{Data: &int8Val},
					{Data: &stringVal2},
				},
				Type:   "edge",
				Labels: []string{"label3", "label4"},
				Rank:   123456,
			}},
			expect: `(1)-[123456@edge:label3&label4{int:8,str:dec}]->(2)`,
		},
		{
			name: "edge",
			value: &NebulaValue{Data: &NebulaEdge{
				SrcId:     1,
				DstId:     2,
				PropNames: []string{"int", "str"},
				PropValues: []*NebulaValue{
					{Data: &int8Val},
					{Data: &stringVal2},
				},
				Type:      "edge",
				Labels:    []string{"label3", "label4"},
				Rank:      123456,
				Direction: edgeNoDirection,
			}},
			expect: `(1)~[123456@edge:label3&label4{int:8,str:dec}]~(2)`,
		},
		{
			name: "path",
			value: &NebulaValue{Data: &NebulaPath{
				Values: []*NebulaValue{
					{Data: &NebulaNode{
						NodeId:    1,
						PropNames: []string{"int", "str"},
						PropValues: []*NebulaValue{
							{Data: &int8Val},
							{Data: &stringVal2},
						},
						Type:   "node1",
						Labels: []string{"label1", "label2"},
					}},
					{Data: &NebulaEdge{
						SrcId:     1,
						DstId:     2,
						PropNames: []string{"int", "str"},
						PropValues: []*NebulaValue{
							{Data: &int8Val},
							{Data: &stringVal2},
						},
						Rank:   123456,
						Type:   "edge",
						Labels: []string{"label3", "label4"},
					}},
					{Data: &NebulaNode{
						NodeId:    2,
						PropNames: []string{"int", "str"},
						PropValues: []*NebulaValue{
							{Data: &int8Val2},
							{Data: &stringVal3},
						},
						Type:   "node2",
						Labels: []string{"label5", "label6"},
					},
					}}}},
			expect: `(1@node1:label1&label2{int:8,str:dec})-[123456@edge:label3&label4{int:8,str:dec}]->(2@node2:label5&label6{int:9,str:abc})`,
		},
		{
			name: "path",
			value: &NebulaValue{Data: &NebulaPath{
				Values: []*NebulaValue{
					{Data: &NebulaNode{
						NodeId: 1,
						Type:   "node1",
						Labels: []string{"label1", "label2"},
					},
					},
					{Data: &NebulaEdge{
						SrcId:  1,
						DstId:  2,
						Rank:   123456,
						Type:   "edge",
						Labels: []string{"label3", "label4"},
					},
					},
					{Data: &NebulaNode{
						NodeId: 2,
						Type:   "node2",
						Labels: []string{"label5", "label6"},
					},
					},
					{Data: &NebulaEdge{
						SrcId:  1,
						DstId:  2,
						Rank:   123456,
						Type:   "edge",
						Labels: []string{"label3", "label4"},
					},
					},
					{Data: &NebulaNode{
						NodeId: 1,
						Type:   "node1",
						Labels: []string{"label1", "label2"},
					},
					},
				}}},
			expect: `(1@node1:label1&label2{})-[123456@edge:label3&label4{}]->(2@node2:label5&label6{})<-[123456@edge:label3&label4{}]-(1@node1:label1&label2{})`,
		},
		{
			name: "path",
			value: &NebulaValue{Data: &NebulaPath{
				Values: []*NebulaValue{
					{Data: &NebulaNode{
						NodeId:    2,
						PropNames: []string{"int", "str"},
						PropValues: []*NebulaValue{
							{Data: &int8Val},
							{Data: &stringVal2},
						},
						Type:   "node1",
						Labels: []string{"label1", "label2"},
					}},
					{Data: &NebulaEdge{
						SrcId:     1,
						DstId:     2,
						PropNames: []string{"int", "str"},
						PropValues: []*NebulaValue{
							{Data: &int8Val},
							{Data: &stringVal2},
						},
						Rank:      123456,
						Type:      "edge",
						Labels:    []string{"label3", "label4"},
						Direction: edgeInComingDirection,
					}},
					{Data: &NebulaNode{
						NodeId:    1,
						PropNames: []string{"int", "str"},
						PropValues: []*NebulaValue{
							{Data: &int8Val2},
							{Data: &stringVal3},
						},
						Type:   "node2",
						Labels: []string{"label5", "label6"},
					},
					}}}},
			expect: `(2@node1:label1&label2{int:8,str:dec})<-[123456@edge:label3&label4{int:8,str:dec}]-(1@node2:label5&label6{int:9,str:abc})`,
		},
		{
			name: "localtime",
			value: &NebulaValue{Data: &NebulaLocalTime{
				Hour: 23, Minute: 59, Sec: 59, Microsec: 999999,
			},
			},
			expect: `23:59:59.999999`,
		},
		{
			name: "localdatetime",
			value: &NebulaValue{Data: &NebulaLocalDatetime{
				Year: 2024, Month: 12, Day: 31, Hour: 23, Minute: 59, Sec: 59, Microsec: 999999,
			}},
			expect: `2024-12-31T23:59:59.999999`,
		},
		{
			name: "zonedtime",
			value: &NebulaValue{Data: &NebulaZonedTime{
				Hour: 23, Minute: 59, Sec: 59, Microsec: 999999,
			}},
			expect: `23:59:59.999999Z`,
		},
		{
			name: "zoneddatetime",
			value: &NebulaValue{Data: &NebulaZonedDatetime{
				Year: 2024, Month: 12, Day: 31, Hour: 23, Minute: 59, Sec: 59, Microsec: 999999,
			}},
			expect: `2024-12-31T23:59:59.999999Z`,
		},
		{
			name: "duration",
			value: &NebulaValue{Data: &NebulaDuration{
				Sec: 23, Microsec: 999999,
			}},
			expect: `PT23.999999S`,
		},
		{
			name: "duration",
			value: &NebulaValue{Data: &NebulaDuration{
				Year: 1, Month: 1, MonthBased: true,
			}},
			expect: `P1Y1M`,
		},
		{
			name: "duration",
			value: &NebulaValue{Data: &NebulaDuration{
				Day: 1,
			}},
			expect: `P1D`,
		},
		{
			name: "duration",
			value: &NebulaValue{Data: &NebulaDuration{
				Sec: 59, Microsec: 0,
			}},
			expect: `PT59S`,
		},
		{
			name: "duration",
			value: &NebulaValue{Data: &NebulaDuration{
				Sec: -59, Microsec: -99000,
			}},
			expect: `PT-59.099S`,
		},
		{
			name: "duration",
			value: &NebulaValue{Data: &NebulaDuration{
				Sec: 0, Microsec: -99,
			}},
			expect: `PT-0.000099S`,
		},
		{
			name: "duration",
			value: &NebulaValue{Data: &NebulaDuration{
				Sec: 0, Microsec: 99,
			}},
			expect: `PT0.000099S`,
		},
	}
	for _, c := range testcases {
		v := c.value
		if v.String() != c.expect {
			t.Fatalf("name: %s expect %s, got %s", c.name, c.expect, v.String())
		}
	}
}

func TestNebulaGeography(t *testing.T) {
	// Test Point
	point := &NebulaGeography{
		SRID:  4326,
		Shape: types.GeoShapePoint,
		Point: &types.Point{Lng: 116.3974, Lat: 39.9093},
	}

	if point.GetSRID() != 4326 {
		t.Errorf("Expected SRID 4326, got %d", point.GetSRID())
	}

	if point.GetShape() != types.GeoShapePoint {
		t.Errorf("Expected shape Point, got %d", point.GetShape())
	}

	if point.GetPoint() == nil {
		t.Error("Expected point to be not nil")
	} else {
		if point.GetPoint().Lng != 116.3974 {
			t.Errorf("Expected Lng 116.3974, got %f", point.GetPoint().Lng)
		}
		if point.GetPoint().Lat != 39.9093 {
			t.Errorf("Expected Lat 39.9093, got %f", point.GetPoint().Lat)
		}
	}

	expectedStr := "POINT(116.3974 39.9093)"
	if point.String() != expectedStr {
		t.Errorf("Expected string %s, got %s", expectedStr, point.String())
	}

	// Test LineString
	lineString := &NebulaGeography{
		SRID:  4326,
		Shape: types.GeoShapeLineString,
		LineString: types.LineString{
			{Lng: 116.3974, Lat: 39.9093},
			{Lng: 116.3975, Lat: 39.9094},
			{Lng: 116.3976, Lat: 39.9095},
		},
	}
	if lineString.GetShape() != types.GeoShapeLineString {
		t.Errorf("Expected shape LineString, got %d", lineString.GetShape())
	}

	if lineString.GetLineString() == nil {
		t.Error("Expected lineString to be not nil")
	} else {
		if len(lineString.GetLineString()) != 3 {
			t.Errorf("Expected 3 coordinates, got %d", len(lineString.GetLineString()))
		}
	}

	expectedLineStr := "LINESTRING(116.3974 39.9093, 116.3975 39.9094, 116.3976 39.9095)"
	if lineString.String() != expectedLineStr {
		t.Errorf("Expected string %s, got %s", expectedLineStr, lineString.String())
	}

	// Test Polygon
	polygon := &NebulaGeography{
		SRID:  4326,
		Shape: types.GeoShapePolygon,
		Polygon: types.Polygon{
			{{Lng: 116.3974, Lat: 39.9093}, {Lng: 116.3975, Lat: 39.9093}, {Lng: 116.3975, Lat: 39.9094}, {Lng: 116.3974, Lat: 39.9094}, {Lng: 116.3974, Lat: 39.9093}},
		},
	}

	if polygon.GetShape() != types.GeoShapePolygon {
		t.Errorf("Expected shape Polygon, got %d", polygon.GetShape())
	}

	if polygon.GetPolygon() == nil {
		t.Error("Expected polygon to be not nil")
	} else {
		loops := polygon.GetPolygon()
		if len(loops) != 1 {
			t.Errorf("Expected 1 loop, got %d", len(loops))
		}
		if len(loops[0]) != 5 {
			t.Errorf("Expected 5 coordinates in first loop, got %d", len(loops[0]))
		}
	}

	expectedPolyStr := "POLYGON((116.3974 39.9093, 116.3975 39.9093, 116.3975 39.9094, 116.3974 39.9094, 116.3974 39.9093))"
	if polygon.String() != expectedPolyStr {
		t.Errorf("Expected string %s, got %s", expectedPolyStr, polygon.String())
	}

	// Test multi-loop polygon
	multiLoopPolygon := &NebulaGeography{
		SRID:  4326,
		Shape: types.GeoShapePolygon,
		Polygon: types.Polygon{
			{{Lng: 1, Lat: 1}, {Lng: 2, Lat: 2}, {Lng: 3, Lat: 3}, {Lng: 1, Lat: 1}}, // First loop
			{{Lng: 4, Lat: 4}, {Lng: 5, Lat: 5}, {Lng: 4, Lat: 4}},                   // Second loop
		},
	}

	expectedMultiLoopStr := "POLYGON((1.0 1.0, 2.0 2.0, 3.0 3.0, 1.0 1.0), (4.0 4.0, 5.0 5.0, 4.0 4.0))"
	if multiLoopPolygon.String() != expectedMultiLoopStr {
		t.Errorf("Expected multi-loop string %s, got %s", expectedMultiLoopStr, multiLoopPolygon.String())
	}
}

func TestDecodeGeographyData(t *testing.T) {
	// Test Point data - using little endian byte order
	// Format: shape type (1 byte) + SRID (4 bytes) + coordinates (16 bytes)
	pointData := []byte{
		0x01,                   // Shape Point (1 byte)
		0x0E, 0x11, 0x00, 0x00, // SRID 4366 (little endian)
		0x9A, 0x99, 0x99, 0x99, 0x99, 0x59, 0x40, 0x00, // X: 116.3974 (little endian)
		0xCD, 0xCC, 0xCC, 0xCC, 0xCC, 0x4C, 0x40, 0x00, // Y: 39.9093 (little endian)
	}

	geoReader := newBytesReader(pointData)
	geography, err := decodeGeographyData(geoReader)
	if err != nil {
		t.Fatalf("Failed to decode point data: %v", err)
	}

	if geography.GetSRID() != 4366 { // 0x1110 in little endian is 4366
		t.Errorf("Expected SRID 4366, got %d", geography.GetSRID())
	}

	if geography.GetShape() != types.GeoShapePoint {
		t.Errorf("Expected shape Point, got %d", geography.GetShape())
	}

	if geography.GetPoint() == nil {
		t.Error("Expected point to be not nil")
	}

	// Test LineString data
	lineStringData := []byte{
		0x05,                   // Shape LineString (1 byte)
		0x0E, 0x11, 0x00, 0x00, // SRID 4366 (little endian)
		0x03, 0x00, 0x00, 0x00, // Number of coordinates: 3 (little endian)
		0x9A, 0x99, 0x99, 0x99, 0x99, 0x59, 0x40, 0x00, // X1: 116.3974
		0xCD, 0xCC, 0xCC, 0xCC, 0xCC, 0x4C, 0x40, 0x00, // Y1: 39.9093
		0x9B, 0x99, 0x99, 0x99, 0x99, 0x59, 0x40, 0x00, // X2: 116.3975
		0xCE, 0xCC, 0xCC, 0xCC, 0xCC, 0x4C, 0x40, 0x00, // Y2: 39.9094
		0x9C, 0x99, 0x99, 0x99, 0x99, 0x59, 0x40, 0x00, // X3: 116.3976
		0xCF, 0xCC, 0xCC, 0xCC, 0xCC, 0x4C, 0x40, 0x00, // Y3: 39.9095
	}

	geoReader = newBytesReader(lineStringData)
	lineString, err := decodeGeographyData(geoReader)
	if err != nil {
		t.Fatalf("Failed to decode lineString data: %v", err)
	}

	if lineString.GetSRID() != 4366 {
		t.Errorf("Expected SRID 4366, got %d", lineString.GetSRID())
	}

	if lineString.GetShape() != types.GeoShapeLineString {
		t.Errorf("Expected shape LineString, got %d", lineString.GetShape())
	}

	if lineString.GetLineString() == nil {
		t.Error("Expected lineString to be not nil")
	} else {
		if len(lineString.GetLineString()) != 3 {
			t.Errorf("Expected 3 coordinates, got %d", len(lineString.GetLineString()))
		}
	}

	// Test Polygon data
	polygonData := []byte{
		0x09,                   // Shape Polygon (1 byte)
		0x0E, 0x11, 0x00, 0x00, // SRID 4366 (little endian)
		0x01, 0x00, 0x00, 0x00, // Number of loops: 1 (little endian)
		0x00, 0x00, 0x00, 0x00, // Row index 0: 0 (little endian)
		0x05, 0x00, 0x00, 0x00, // Row index 1: 5 (little endian)
		0x9A, 0x99, 0x99, 0x99, 0x99, 0x59, 0x40, 0x00, // X1: 116.3974
		0xCD, 0xCC, 0xCC, 0xCC, 0xCC, 0x4C, 0x40, 0x00, // Y1: 39.9093
		0x9B, 0x99, 0x99, 0x99, 0x99, 0x59, 0x40, 0x00, // X2: 116.3975
		0xCD, 0xCC, 0xCC, 0xCC, 0xCC, 0x4C, 0x40, 0x00, // Y2: 39.9093
		0x9B, 0x99, 0x99, 0x99, 0x99, 0x59, 0x40, 0x00, // X3: 116.3975
		0xCE, 0xCC, 0xCC, 0xCC, 0xCC, 0x4C, 0x40, 0x00, // Y3: 39.9094
		0x9A, 0x99, 0x99, 0x99, 0x99, 0x59, 0x40, 0x00, // X4: 116.3974
		0xCE, 0xCC, 0xCC, 0xCC, 0xCC, 0x4C, 0x40, 0x00, // Y4: 39.9094
		0x9A, 0x99, 0x99, 0x99, 0x99, 0x59, 0x40, 0x00, // X5: 116.3974
		0xCD, 0xCC, 0xCC, 0xCC, 0xCC, 0x4C, 0x40, 0x00, // Y5: 39.9093
	}

	geoReader = newBytesReader(polygonData)
	polygon, err := decodeGeographyData(geoReader)
	if err != nil {
		t.Fatalf("Failed to decode polygon data: %v", err)
	}

	if polygon.GetSRID() != 4366 {
		t.Errorf("Expected SRID 4366, got %d", polygon.GetSRID())
	}

	if polygon.GetShape() != types.GeoShapePolygon {
		t.Errorf("Expected shape Polygon, got %d", polygon.GetShape())
	}

	if polygon.GetPolygon() == nil {
		t.Error("Expected polygon to be not nil")
	} else {
		loops := polygon.GetPolygon()
		if len(loops) != 1 {
			t.Errorf("Expected 1 loop, got %d", len(loops))
		}
		if len(loops[0]) != 5 {
			t.Errorf("Expected 5 coordinates in first loop, got %d", len(loops[0]))
		}
	}
}

func TestFormatFloat(t *testing.T) {
	testCases := []struct {
		name      string
		value     float64
		precision int
		expected  string
	}{
		{
			name:      "normal float value",
			value:     123.456,
			precision: 64,
			expected:  "123.456",
		},
		{
			name:      "normal float value with 32 bit precision",
			value:     123.456,
			precision: 32,
			expected:  "123.456",
		},
		{
			name:      "integer value should have .0 suffix",
			value:     42.0,
			precision: 64,
			expected:  "42.0",
		},
		{
			name:      "zero value should have .0 suffix",
			value:     0.0,
			precision: 64,
			expected:  "0.0",
		},
		{
			name:      "positive infinity",
			value:     math.Inf(1),
			precision: 64,
			expected:  "+Inf",
		},
		{
			name:      "negative infinity",
			value:     math.Inf(-1),
			precision: 64,
			expected:  "-Inf",
		},
		{
			name:      "NaN value",
			value:     math.NaN(),
			precision: 64,
			expected:  "NaN",
		},
		{
			name:      "very small positive value",
			value:     0.000001,
			precision: 64,
			expected:  "0.000001",
		},
		{
			name:      "very small negative value",
			value:     -0.000001,
			precision: 64,
			expected:  "-0.000001",
		},
		{
			name:      "large positive value",
			value:     1234567890.123456,
			precision: 64,
			expected:  "1234567890.123456",
		},
		{
			name:      "large negative value",
			value:     -1234567890.123456,
			precision: 64,
			expected:  "-1234567890.123456",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatFloat(tc.value, tc.precision)
			if result != tc.expected {
				t.Errorf("formatFloat(%f, %d) = %s, expected %s", tc.value, tc.precision, result, tc.expected)
			}
		})
	}
}
