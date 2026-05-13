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
package e2e

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	nebula "github.com/vesoft-inc/nebula-go/v5"
)

func TestPoolSessionSet(t *testing.T) {
	params := map[string]string{
		"s": "\"1\"",
		"i": "2",
	}
	p, err := nebula.NewNebulaPool(nebulaAddress, nebulaUser, nebulaPassword,
		nebula.WithPoolTimezone("Asia/Shanghai"),
		nebula.WithPoolParameters(params),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	c, err := p.GetClient()
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.Execute(`return zoned_datetime("2020-03-02T23:12:00+0000")`)
	if err != nil {
		t.Fatal(err)
	}
	var dt nebula.NullZonedDatetime
	if err := resp.Scan(&dt); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, dt.Valid, true)
	assert.Equal(t, dt.Data.GetOffset(), 8*3600)
	assert.Equal(t, dt.Data.GetDay(), 3)
	assert.Equal(t, dt.Data.GetHour(), 7)
	assert.Equal(t, dt.Data.GetMinute(), 12)

	resp, err = c.Execute(`return $s, $i`)
	if err != nil {
		t.Fatal(err)
	}
	var s nebula.NullString
	var i nebula.NullInt
	if err := resp.Scan(&s, &i); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, s.Valid, true)
	assert.Equal(t, string(s.Data), "1")
	assert.Equal(t, i.Valid, true)
	assert.Equal(t, int(i.Data), 2)
}

func TestSessionSet(t *testing.T) {
	p, err := nebula.NewNebulaPool(nebulaAddress, nebulaUser, nebulaPassword,
		nebula.WithPoolGraph("test_graph"),
		nebula.WithPoolSchema("/test_schema"),
		nebula.WithPoolTimezone("Asia/Shanghai"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	c, err := p.GetClient()
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.Execute(`show current_schema`)
	if err != nil {
		t.Fatal(err)
	}
	var path, owner nebula.NullString
	if err := resp.Scan(&path, &owner); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, path.Valid, true)
	assert.Equal(t, string(path.Data), "/test_schema")
	assert.Equal(t, owner.Valid, true)
	assert.Equal(t, string(owner.Data), "root")
	resp, err = c.Execute(`show current_session`)
	if err != nil {
		t.Fatal(err)
	}
	row, err := resp.Next()
	if err != nil {
		t.Fatal(err)
	}
	v, err := row.GetValueByName("graph")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "/test_schema/test_graph", v.String())
	v, err = row.GetValueByName("timezone")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, v.String(), "Asia/Shanghai")
}

func TestSessionConfigs(t *testing.T) {
	configs := map[string]string{
		"timezone":    "\"Asia/Shanghai\"",
		"date_format": "\"%Y%m%d\"",
	}
	p, err := nebula.NewNebulaPool(nebulaAddress, nebulaUser, nebulaPassword,
		nebula.WithPoolSessionConfigs(configs),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	c, err := p.GetClient()
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Execute(`return DATE("2023-01-01")`)
	assert.NotNil(t, err)
	t.Log(err)
	resp, err := c.Execute(`return DATE("20230101")`)
	if err != nil {
		t.Fatal(err)
	}
	var d nebula.NullDate
	if err := resp.Scan(&d); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, d.Valid, true)
	assert.Equal(t, d.Data.GetYear(), 2023)
	assert.Equal(t, d.Data.GetMonth(), 1)
	assert.Equal(t, d.Data.GetDay(), 1)
	resp, err = c.Execute("call show_session_configs() yield name as n, `value` as v return n,v")
	if err != nil {
		t.Fatal(err)
	}
	var name, value nebula.NullString
	for resp.HasNext() {
		if err := resp.Scan(&name, &value); err != nil {
			t.Fatal(err)
		}
		if !name.Valid || !value.Valid {
			t.Fatal("invalid session config")
		}
		switch name.Data {
		case "schema":
			assert.Equal(t, string(value.Data), "/test_schema")
		case "graph":
			assert.Equal(t, string(value.Data), "test_graph")
		case "timezone":
			assert.Equal(t, string(value.Data), "Asia/Shanghai")
		case "date_format":
			assert.Equal(t, string(value.Data), "%Y%m%d")
		default:
			t.Fatalf("unknown session config %s", name.Data)
		}
	}
}

func TestPoolOnExecute(t *testing.T) {
	addr := fmt.Sprintf("%s:%d", nebulaHost, nebulaPort)
	pool, err := nebula.NewNebulaPool(addr, nebulaUser, nebulaPassword,
		nebula.WithPoolExecuteOnOpenSession([]string{
			`session set schema "/test_schema"`,
			`session set graph "test_graph"`,
		}))
	if err != nil {
		t.Fatal(err)
	}
	s, err := pool.GetClient()
	if err != nil {
		t.Fatal(err)
	}
	resp, err := s.Execute("call show_session_configs() yield name as n, `value` as v return n,v")
	if err != nil {
		t.Fatal(err)
	}
	var name, value nebula.NullString
	for resp.HasNext() {
		if err := resp.Scan(&name, &value); err != nil {
			t.Fatal(err)
		}
		if !name.Valid || !value.Valid {
			t.Fatal("invalid session config")
		}
		switch name.Data {
		case "schema":
			assert.Equal(t, string(value.Data), "/test_schema")
		case "graph":
			assert.Equal(t, string(value.Data), "test_graph")
		case "timezone":
			assert.Equal(t, string(value.Data), "Asia/Shanghai")
		case "date_format":
			assert.Equal(t, string(value.Data), "%Y%m%d")
		default:
			t.Fatalf("unknown session config %s", name.Data)
		}
	}
}

func TestPoolResultSetAll(t *testing.T) {
	addr := fmt.Sprintf("%s:%d", nebulaHost, nebulaPort)
	pool, err := nebula.NewNebulaPool(addr, nebulaUser, nebulaPassword,
		nebula.WithPoolExecuteOnOpenSession([]string{
			`session set schema "/test_schema"`,
			`session set graph "test_graph"`,
		}))
	if err != nil {
		t.Fatal(err)
	}
	s, err := pool.GetClient()
	if err != nil {
		t.Fatal(err)
	}
	resp, err := s.Execute("call show_session_configs() yield name as n, `value` as v return n,v")
	if err != nil {
		t.Fatal(err)
	}

	for row, err := range resp.All() {
		assert.NoError(t, err)

		n, err := row.GetValueByName("n")
		assert.NoError(t, err)
		name, err := n.AsString()
		assert.NoError(t, err)
		v, err := row.GetValueByName("v")
		assert.NoError(t, err)
		value, err := v.AsString()
		assert.NoError(t, err)

		switch name {
		case "schema":
			assert.Equal(t, value, "/test_schema")
		case "graph":
			assert.Equal(t, value, "test_graph")
		case "timezone":
			assert.Equal(t, value		, "Asia/Shanghai")
		case "date_format":
			assert.Equal(t, value, "%Y%m%d")
		default:
			t.Fatalf("unknown session config %s", name)
		}
	}
}
