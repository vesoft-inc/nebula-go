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

	"github.com/stretchr/testify/assert"
	nebula "github.com/vesoft-inc/nebula-go/v5"
)

func TestLexical(t *testing.T) {
	c, err := nebula.NewNebulaClient(nebulaAddress, nebulaUser, nebulaPassword)
	if err != nil {
		t.Fatalf("NewNebulaClient failed, %s", err.Error())
	}
	defer c.Close()
	testcases := []struct {
		stmt     string
		expected string
	}{
		{"return 1", "1"},
		{"return 1.1", "1.1"},
		{"return \"hello\"", "hello"},
		{"return true", "true"},
		{"return UNKNOWN", "null"},
		{"return [1,2,3]", "[1,2,3]"},
		{"return {\"a\" : 1}", "{a:1}"},
		{"return datetime '2015-07-21T07:12:23.123+0800'", "2015-07-20T23:12:23.123000Z"},
		{"return datetime '2015-07-21T23:12:23.123'", "2015-07-21T23:12:23.123000"},
		{"return duration 'P0Y'", "P0M"},
		{"return duration 'P0D'", "PT0S"},
	}

	for _, tc := range testcases {
		resp, err := c.Execute(tc.stmt)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 1, resp.RowSize())
		row, err := resp.Next()
		if err != nil {
			t.Fatal(err)
		}
		v, err := row.GetValueByIndex(0)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, tc.expected, v.String())
	}
}

func TestConstVector(t *testing.T) {
	c, err := nebula.NewNebulaClient(nebulaAddress, nebulaUser, nebulaPassword)
	if err != nil {
		t.Fatalf("NewNebulaClient failed, %s", err.Error())
	}
	defer c.Close()
	stmt := `
	 use movie match p=(v1:Person{id:15})-[e:Watch]->(v2:Movie)  
	 match (v:Person{id:15}) return p,v2,e
	`
	resp, err := c.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	assert.Greater(t, resp.RowSize(), 1)
	var p nebula.NullPath
	var v nebula.NullNode
	var e nebula.NullEdge
	if err := resp.Scan(&p, &v, &e); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, p.Valid, true)
	assert.Equal(t, v.Valid, true)
	assert.Equal(t, e.Valid, true)
	pValue := p.Data.GetValues()[0]
	pNode, err := pValue.AsNode()
	if err != nil {
		t.Fatal(err)
	}
	idValue := pNode.GetProperties()["id"]
	id, err := idValue.AsInt64()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int(id), 15)
	assert.Equal(t, e.Data.GetType(), "Watch")
	assert.Equal(t, v.Data.GetType(), "Movie")
}

func TestFlatVector(t *testing.T) {
	c, err := nebula.NewNebulaClient(nebulaAddress, nebulaUser, nebulaPassword)
	if err != nil {
		t.Fatalf("NewNebulaClient failed, %s", err.Error())
	}
	defer c.Close()
	stmt := `
	 use movie match p=(v1:Person{id:15})-[e:Watch]->(v2:Movie) limit 10 return p,v2,e
	`
	resp, err := c.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	assert.LessOrEqual(t, resp.RowSize(), 10)
	var p nebula.NullPath
	var v nebula.NullNode
	var e nebula.NullEdge
	if err := resp.Scan(&p, &v, &e); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, p.Valid, true)
	assert.Equal(t, v.Valid, true)
	assert.Equal(t, e.Valid, true)
	pValue := p.Data.GetValues()[0]
	pNode, err := pValue.AsNode()
	if err != nil {
		t.Fatal(err)
	}
	idValue := pNode.GetProperties()["id"]
	id, err := idValue.AsInt64()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int(id), 15)
	assert.Equal(t, pNode.GetLabels(), []string{"Person"})
	assert.Equal(t, e.Data.GetType(), "Watch")
	assert.Equal(t, v.Data.GetType(), "Movie")
}
