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
package main

import (
	"fmt"
	"log"

	nebula "github.com/vesoft-inc/nebula-go/v5"
)

func main() {
	// Create a client
	client, err := nebula.NewNebulaClient("127.0.0.1:7188", "root", "NebulaGraph01")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Test Geography Point with debug info
	fmt.Println("=== Testing Geography Point ===")
	result, err := client.Execute("RETURN ST_GeogFromText('POINT(116.3974 39.9093)') AS p")
	if err != nil {
		log.Printf("Failed to execute query: %v", err)
	} else {
		for result.HasNext() {
			row, err := result.Next()
			if err != nil {
				log.Printf("Failed to get next row: %v", err)
				continue
			}

			pointValue, err := row.GetValueByIndex(0)
			if err != nil {
				log.Printf("Failed to get point value: %v", err)
				continue
			}

			// Debug: print the raw value
			fmt.Printf("Raw value type: %s\n", pointValue.GetType())
			fmt.Printf("Raw value string: %s\n", pointValue.String())

			geography, err := pointValue.AsGeography()
			if err != nil {
				log.Printf("Failed to convert to geography: %v", err)
				continue
			}

			fmt.Printf("Point: %s\n", geography.String())
			fmt.Printf("SRID: %d\n", geography.GetSRID())
			fmt.Printf("Shape: %d\n", geography.GetShape())

			if point := geography.GetPoint(); point != nil {
				fmt.Printf("Coordinates: (%f, %f)\n", point.Lng, point.Lat)
			}
		}
	}

	// Test Geography LineString with debug info
	fmt.Println("\n=== Testing Geography LineString ===")
	result, err = client.Execute("RETURN ST_GeogFromText('LINESTRING(116.3974 39.9093, 116.4074 39.9193, 116.4174 39.9293)') AS l")
	if err != nil {
		log.Printf("Failed to execute query: %v", err)
	} else {
		for result.HasNext() {
			row, err := result.Next()
			if err != nil {
				log.Printf("Failed to get next row: %v", err)
				continue
			}

			lineValue, err := row.GetValueByIndex(0)
			if err != nil {
				log.Printf("Failed to get line value: %v", err)
				continue
			}

			// Debug: print the raw value
			fmt.Printf("Raw value type: %s\n", lineValue.GetType())
			fmt.Printf("Raw value string: %s\n", lineValue.String())

			geography, err := lineValue.AsGeography()
			if err != nil {
				log.Printf("Failed to convert to geography: %v", err)
				continue
			}

			fmt.Printf("LineString: %s\n", geography.String())
			fmt.Printf("SRID: %d\n", geography.GetSRID())
			fmt.Printf("Shape: %d\n", geography.GetShape())

			if line := geography.GetLineString(); line != nil {
				fmt.Printf("Number of points: %d\n", len(line))
				for i, coord := range line {
					fmt.Printf("Point %d: (%f, %f)\n", i, coord.Lng, coord.Lat)
				}
			}
		}
	}

	// Test Geography Polygon with debug info
	fmt.Println("\n=== Testing Geography Polygon ===")
	result, err = client.Execute("RETURN ST_GeogFromText('POLYGON((116.3974 39.9093, 116.4074 39.9093, 116.4074 39.9193, 116.3974 39.9193, 116.3974 39.9093))') AS poly")
	if err != nil {
		log.Printf("Failed to execute query: %v", err)
	} else {
		for result.HasNext() {
			row, err := result.Next()
			if err != nil {
				log.Printf("Failed to get next row: %v", err)
				continue
			}

			polyValue, err := row.GetValueByIndex(0)
			if err != nil {
				log.Printf("Failed to get polygon value: %v", err)
				continue
			}

			// Debug: print the raw value
			fmt.Printf("Raw value type: %s\n", polyValue.GetType())
			fmt.Printf("Raw value string: %s\n", polyValue.String())

			geography, err := polyValue.AsGeography()
			if err != nil {
				log.Printf("Failed to convert to geography: %v", err)
				continue
			}

			fmt.Printf("Polygon: %s\n", geography.String())
			fmt.Printf("SRID: %d\n", geography.GetSRID())
			fmt.Printf("Shape: %d\n", geography.GetShape())

			if polygon := geography.GetPolygon(); polygon != nil {
				fmt.Printf("Number of rings: %d\n", len(polygon))
				for i, loop := range polygon {
					fmt.Printf("Ring %d has %d points:\n", i, len(loop))
					for j, coord := range loop {
						fmt.Printf("  Point %d: (%f, %f)\n", j, coord.Lng, coord.Lat)
					}
				}
			}
		}
	}
}
