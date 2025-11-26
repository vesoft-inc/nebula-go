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
package e2e

import (
	"testing"

	nebula "github.com/vesoft-inc/nebula-go/v5"
)

func TestDecodeFlat(t *testing.T) {
	p, err := nebula.NewNebulaPool(nebulaAddress, nebulaUser, nebulaPassword)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	session, err := p.GetClient()
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if _, err = session.Execute(`
	      CREATE GRAPH TYPE IF NOT EXISTS ddl_test_type AS {
        NODE TYPE player (LABEL player {
        	id INT PRIMARY KEY, 
        	name STRING, 
        	prop_set set<int>,
        	prop_list list<string>,
        	prop_map  map<int, string>,
        	vec1 VECTOR<3, float>, 
        	vec2 VECTOR<128, float>
        })
      }
	`); err != nil {
		t.Fatal(err)
	}
	if _, err = session.Execute(`
			CREATE GRAPH ddl_test TYPED ddl_test_type
	`); err != nil {
		t.Fatal(err)
	}
	if _, err := session.Execute(`
	  use ddl_test insert (@player{id:1, prop_set: set{null,1}, prop_map: map{100:null}, vec1: VECTOR<3, FLOAT>([1.0, 2.0, 3.0]) })
		`); err != nil {
		t.Fatal(err)
	}
	result, err := session.Execute(`
	  use ddl_test match (a:player) return a.id, a.name, a.prop_set, a.prop_list, a.prop_map, a.vec1, a.vec2
	`)
	if err != nil {
		t.Fatal(err)
	}
	var id nebula.NullInt
	var name nebula.NullString
	var propSet nebula.NullSet
	var propList nebula.NullList
	var propMap nebula.NullMap
	var vec1 nebula.NullEmbeddingVector
	var vec2 nebula.NullEmbeddingVector
	if err := result.Scan(&id, &name, &propSet, &propList, &propMap, &vec1, &vec2); err != nil {
		t.Fatal(err)
	}
	if !id.Valid || id.Data != 1 {
		t.Fatalf("invalid id: %+v", id)
	}
}
