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
package nebula_ng

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vesoft-inc/nebula-go/v5/pkg/types"
)

type connectorWithErr struct {
	connectTime time.Duration
	executeTime time.Duration
	dummyConnector
	connectTimes int
	err          error
}

type connWithErr struct {
	dummyConn
	executeTime time.Duration
	err         error
}

type connectorWithFunc struct {
	dummyConn
	fn func(address string, cfg *connConfig) (types.Client, error)
}

func (c *connectorWithErr) connect(address string, cfg *connConfig) (types.Client, error) {
	c.connectTimes++
	return &connWithErr{executeTime: c.executeTime, err: c.err}, nil
}

func (c *connectorWithFunc) connect(address string, cfg *connConfig) (types.Client, error) {
	return c.fn(address, cfg)
}

func (d *connWithErr) Execute(stmt string) (types.Result, error) {
	time.Sleep(d.executeTime)
	return nil, d.err
}

func (d *connWithErr) ExecuteContext(ctx context.Context, stmt string) (types.Result, error) {
	time.Sleep(d.executeTime)
	return nil, d.err
}

func TestClientRetry(t *testing.T) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	testcases := []struct {
		connectTimes int
		ctx          context.Context
		maxLifeTime  time.Duration
		executionErr error
		err          error
	}{
		{1, timeoutCtx, 150 * time.Millisecond, fmt.Errorf("error"), timeoutCtx.Err()},
		{1, context.Background(), 15 * time.Millisecond, nil, nil},
		{1, context.Background(), 15 * time.Millisecond, fmt.Errorf("error"), fmt.Errorf("error")},
		{1, context.Background(), 15 * time.Millisecond, fmt.Errorf("not connnection error"), fmt.Errorf("not connnection error")},
	}
	for _, tc := range testcases {
		connector := &connectorWithErr{
			connectTime: 3 * time.Millisecond,
			executeTime: 20 * time.Millisecond,
			err:         tc.executionErr,
		}
		client, err := NewNebulaClient("127.0.0.1:9669", "", "", withClientConnector(connector))
		if err != nil {
			t.Fatal(err)
		}

		_, err = client.ExecuteContext(tc.ctx, "")
		if tc.err != nil {
			assert.EqualError(t, err, tc.err.Error())
		}
		assert.Equal(t, tc.connectTimes, connector.connectTimes)
	}
}

func TestPoolRetry(t *testing.T) {
	// when using pool, user should put the client back to pool
	// means alway create a connection at last.
	// so if retry times is 2, pool would open connection 4 times if all connections are broken
	timeoutCtx, _ := context.WithTimeout(context.Background(), 25*time.Millisecond)
	testcases := []struct {
		connectTimes int
		ctx          context.Context
		maxLifeTime  time.Duration
		executionErr error
		err          error
	}{
		// timeout is 25ms, connectTime is 3ms, executeTime is 20ms
		{1, timeoutCtx, 150 * time.Millisecond, fmt.Errorf("error"), timeoutCtx.Err()},
		{1, context.Background(), 15 * time.Millisecond, nil, nil},
		{1, context.Background(), 15 * time.Millisecond, fmt.Errorf("error"), fmt.Errorf("error")},
		{1, context.Background(), 15 * time.Millisecond, fmt.Errorf("not connnection error"), fmt.Errorf("not connnection error")},
	}

	for _, tc := range testcases {
		connector := &connectorWithErr{
			connectTime: 3 * time.Millisecond,
			executeTime: 20 * time.Millisecond,
			err:         tc.executionErr,
		}
		p, err := NewNebulaPool("127.0.0.1:9669", "", "", withPoolConnector(connector))
		if err != nil {
			t.Fatal(err)
		}
		defer p.Close()
		pool, _ := p.(*driverPool)
		pool.connMaxLifeTime = tc.maxLifeTime
		pool.mu.Lock()
		pool.minOpen = 0
		pool.mu.Unlock()
		pool.connCfg = &connConfig{
			requestTimeout: 10 * time.Second,
		}
		client, err := pool.GetClient()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 1, len(pool.connMap))
		assert.Equal(t, 0, len(pool.freeConn))
		_, err = client.ExecuteContext(tc.ctx, "")
		if tc.err != nil {
			assert.EqualError(t, err, tc.err.Error())
		}
		assert.Equal(t, tc.connectTimes, connector.connectTimes)
	}
}

// after retry, the connection should be put back to pool
// verify the broken connection would be removed from pool
func TestPoolRetry2(t *testing.T) {
	connector := &connectorWithErr{
		connectTime: 3 * time.Millisecond,
		executeTime: 20 * time.Millisecond,
		err:         fmt.Errorf("error"),
	}
	p, err := NewNebulaPool("127.0.0.1:9669", "", "", withPoolConnector(connector))
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	pool, _ := p.(*driverPool)
	pool.mu.Lock()
	pool.connector = connector
	pool.minOpen = 0
	pool.connCfg = &connConfig{
		requestTimeout: 10 * time.Second,
	}
	pool.mu.Unlock()
	client, err := pool.GetClient()
	if err != nil {
		t.Fatal(err)
	}
	_, _ = client.Execute("")
	if err := pool.PutClient(client); err != nil {
		t.Fatal(err)
	}
	pool.mu.Lock()
	assert.Equal(t, 1, len(pool.connMap))
	assert.Equal(t, 1, len(pool.freeConn))
	for _, conn := range pool.freeConn {
		c := conn
		if _, ok := pool.connMap[c]; !ok {
			t.Fatal("conn not in connMap")
		}
	}
	pool.mu.Unlock()
}

func TestPoolStrictlyServerHealthy(t *testing.T) {
	connector := &connectorWithFunc{
		fn: func(address string, cfg *connConfig) (types.Client, error) {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				return nil, err
			}
			if host == "127.0.0.1" && port == "9669" {
				return nil, fmt.Errorf("broken")
			} else {
				return &dummyConn{}, nil
			}
		},
	}
	testcases := []struct {
		strictly  bool
		addresses string
		hasErr    bool
		err       string
	}{
		{true, "127.0.0.4:9669,127.0.0.2:9669,127.0.0.3:9669", false, ""},
		{false, "127.0.0.4:9669,127.0.0.2:9669,127.0.0.3:9669", false, ""},
		{true, "127.0.0.1:9669,127.0.0.2:9669,127.0.0.3:9669", true, "broken"},
		{false, "127.0.0.1:9669,127.0.0.2:9669,127.0.0.3:9669", false, ""},
		{false, "127.0.0.1:9669", true, "broken"},
	}
	for _, tc := range testcases {
		_, err := NewNebulaPool(tc.addresses, "", "",
			withPoolConnector(connector),
			WithPoolStrictlyServerHealthy(tc.strictly),
		)
		assert.Equal(t, tc.hasErr, err != nil)
		if tc.hasErr {
			assert.Equal(t, tc.err, err.Error())
		}
	}

}

// if session set error, new pool would return error
func TestPoolStrictlyServerHealthy2(t *testing.T) {
	connector := &connectorWithErr{
		err: fmt.Errorf("session set error"),
	}
	testcases := []struct {
		graphName string
		hasErr    bool
		errMsg    string
	}{
		{"test", true, "session set error"},
		{"", false, ""},
	}
	var err error
	for _, tc := range testcases {
		if tc.graphName != "" {
			_, err = NewNebulaPool("127.0.0.1:9669", "", "", withPoolConnector(connector),
				WithPoolGraph(tc.graphName))
		} else {
			_, err = NewNebulaPool("127.0.0.1:9669", "", "", withPoolConnector(connector))
		}
		assert.Equal(t, tc.hasErr, err != nil)
		if tc.hasErr {
			assert.Equal(t, tc.errMsg, err.Error())
		}
	}
}
