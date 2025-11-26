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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	nebula_ng "github.com/vesoft-inc/nebula-go/v5"
	"github.com/vesoft-inc/nebula-go/v5/pkg/types"
)

var addr = fmt.Sprintf("%s:%d", nebulaHost, nebulaPort)

func TestReturnDecode(t *testing.T) {
	testcases := []struct {
		name     string
		stmt     string
		expected string
	}{
		{name: "null", stmt: "return null", expected: "null"},
		{name: "bool", stmt: "return true", expected: "true"},
		{name: "string", stmt: "return 'hello world'", expected: "hello world"},
		{name: "int", stmt: "return 123", expected: "123"},
		{name: "float", stmt: "return 123.456", expected: "123.456"},
		{name: "list", stmt: "return [1, 2, 3]", expected: "[1,2,3]"},
		{name: "record", stmt: "return {a:1, b:2}", expected: "{a:1,b:2}"},
		{name: "local_time", stmt: "return LOCAL_TIME('20:01:02')", expected: "20:01:02.000000"},
		{name: "zoned_time", stmt: "return ZONED_TIME('20:01:02+0000')", expected: "20:01:02.000000Z"},
		{name: "zoned_time", stmt: "return ZONED_TIME('20:01:02+0800')", expected: "12:01:02.000000Z"},
		{name: "data", stmt: "return DATE('2020-01-02')", expected: "2020-01-02"},
		{name: "datetime", stmt: "return Zoned_DATETIME('2020-01-02T20:01:02+0800')", expected: "2020-01-02T12:01:02.000000Z"},
		{name: "datetime", stmt: "return Zoned_DATETIME('2020-01-02T20:01:02+0000')", expected: "2020-01-02T20:01:02.000000Z"},
		{name: "datetime", stmt: "return Zoned_DATETIME('2020-01-02T20:01:02-0800')", expected: "2020-01-03T04:01:02.000000Z"},
		{name: "duration", stmt: "return DURATION('PT0H1M')", expected: "PT1M"},
		{name: "duration", stmt: "return DURATION('P1Y2M')", expected: "P1Y2M"},
	}
	c, err := nebula_ng.NewNebulaClient(
		addr,
		nebulaUser,
		nebulaPassword,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := c.Execute(tc.stmt)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, 1, resp.RowSize())
			row, _ := resp.Next()
			v, err := row.GetValueByIndex(0)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.expected, v.String())
		})
	}
}

func TestReturnNode(t *testing.T) {
	stmt := "USE movie match(v@Actor{id:6217}) return v "
	c, err := nebula_ng.NewNebulaClient(
		addr,
		nebulaUser,
		nebulaPassword,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	resp, err := c.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, resp.RowSize())
	row, _ := resp.Next()
	v, err := row.GetValueByIndex(0)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, types.ValueTypeNode, v.GetType())
	n, err := v.AsNode()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "movie", n.GetGraph())
	assert.Equal(t, []string{"Person"}, n.GetLabels())
	assert.Equal(t, "Actor", n.GetType())
	props := n.GetProperties()
	assert.Equal(t, 3, len(props))
	name := props["name"]
	birthday := props["birthDate"]
	id := props["id"]
	assert.Equal(t, "Anupam Kher", name.String())
	assert.Equal(t, "1955-03-07", birthday.String())
	assert.Equal(t, "6217", id.String())
	assert.Equal(t, types.ValueTypeString, name.GetType())
	assert.Equal(t, types.ValueTypeDate, birthday.GetType())
	assert.Equal(t, types.ValueTypeInt64, id.GetType())
	assert.Equal(t, "NODE", types.ValueTypeNode.String())
}

func TestResultEdge(t *testing.T) {
	stmt := "use movie match(v@`User`{id:540})-[e]->() order by e.rate return e"
	c, err := nebula_ng.NewNebulaClient(
		addr,
		nebulaUser,
		nebulaPassword,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	resp, err := c.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 7, resp.RowSize())
	row, _ := resp.Next()
	v, err := row.GetValueByIndex(0)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, types.ValueTypeEdge, v.GetType())
	e, err := v.AsEdge()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "movie", e.GetGraph())
	assert.Equal(t, []string{"Watch"}, e.GetLabels())
	assert.Equal(t, "Watch", e.GetType())
	props := e.GetProperties()
	assert.Equal(t, 1, len(props))
	rate := props["rate"]
	assert.Equal(t, "2.5", rate.String())
	assert.Equal(t, types.ValueTypeFloat, rate.GetType())
}

func TestSummary(t *testing.T) {
	c, err := nebula_ng.NewNebulaClient(
		addr,
		nebulaUser,
		nebulaPassword,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	stmt := "return 1"
	resp, err := c.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	summary := resp.Summary()
	assert.NotNil(t, summary)
	assert.Greater(t, summary.SerializeTimeUs(), int64(0))

}
