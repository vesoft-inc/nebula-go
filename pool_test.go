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
package nebula_ng

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vesoft-inc/nebula-go/v5/pkg/types"
	"golang.org/x/sync/errgroup"
)

type dummyConnector struct {
	sleep time.Duration
}

type dummyConn struct {
	id     int
	closed bool
}

func (c *dummyConnector) connect(address string, cfg *connConfig) (types.Client, error) {
	if c.sleep > 0 {
		<-time.After(c.sleep)
	}
	return &dummyConn{closed: false}, nil
}

func (d *dummyConn) Execute(stmt string) (types.Result, error) {
	return nil, nil
}

func (d *dummyConn) ExecuteContext(ctx context.Context, stmt string) (types.Result, error) {
	return nil, nil
}
func (d *dummyConn) Ping() error {
	return nil
}
func (d *dummyConn) PingContext(ctx context.Context) error {
	return nil
}

func (d *dummyConn) GetSessionId() (int64, error) {
	return 0, nil
}
func (d *dummyConn) GetVersion() (string, error) {
	return "v5.0.0", nil
}
func (d *dummyConn) Close() error {
	d.closed = true
	return nil
}
func (d *dummyConn) IsClosed() bool {
	return d.closed == true
}

func TestCleanIdle(t *testing.T) {
	testcases := []struct {
		maxIdle  int
		conns    []types.Client
		expected int
	}{
		{5,
			[]types.Client{
				&dummyConn{id: 0},
				&dummyConn{id: 1},
				&dummyConn{id: 2},
				&dummyConn{id: 3},
				&dummyConn{id: 4},
			},
			5,
		},
		{4,
			[]types.Client{
				&dummyConn{id: 0},
				&dummyConn{id: 1},
				&dummyConn{id: 2},
				&dummyConn{id: 3},
				&dummyConn{id: 4},
			},
			4,
		},
		{2,
			[]types.Client{
				&dummyConn{id: 0},
				&dummyConn{id: 1},
			},
			2,
		},
	}

	for _, tc := range testcases {
		connMap := make(map[types.Client]struct{})
		for _, conn := range tc.conns {
			c := conn
			connMap[c] = struct{}{}
		}
		pool := &driverPool{
			maxIdle:  tc.maxIdle,
			freeConn: tc.conns,
			connMap:  connMap,
		}
		pool.clearIdleConn()
		assert.Equal(t, tc.expected, len(pool.freeConn))
		assert.Equal(t, tc.expected, len(pool.connMap))
	}
}

func TestPool(t *testing.T) {
	var (
		maxOpen = 100
		maxIdle = 50
		minIdle = 10
	)
	p, err := NewNebulaPool("127.0.0.1:9669", "", "",
		WithPoolMaxOpenConns(maxOpen),
		WithPoolMaxIdleConns(maxIdle),
		WithPoolMinOpenConns(minIdle),
		withPoolConnector(&dummyConnector{}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	pool, _ := p.(*driverPool)
	defer pool.Close()
	pool.mu.Lock()
	pool.tickerDuration = 100 * time.Millisecond
	pool.connCfg = &connConfig{
		requestTimeout: 10 * time.Second,
	}
	pool.mu.Unlock()

	c, err := pool.GetClient()
	if err != nil {
		t.Fatal(err)
	}
	<-time.After(10 * time.Millisecond)
	if err := pool.PutClient(c); err != nil {
		t.Fatal(err)
	}
	//test min open
	pool.mu.Lock()
	assert.Equal(t, pool.minOpen, len(pool.connMap))
	assert.Equal(t, pool.minOpen, len(pool.freeConn))
	pool.mu.Unlock()
	//test max open
	var wg sync.WaitGroup
	for i := 0; i < maxOpen; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, err := pool.GetClient()
			if err != nil {
				t.Log(err)
			}
			<-time.After(30 * time.Millisecond)
			err = pool.PutClient(c)
			if err != nil {
				t.Log(err)
			}
		}()
	}
	wg.Wait()

	pool.mu.Lock()
	assert.Equal(t, maxOpen, len(pool.connMap))
	assert.Equal(t, maxOpen, len(pool.freeConn))
	pool.mu.Unlock()
	//test idle conn
	<-time.After(150 * time.Millisecond)

	pool.mu.Lock()
	assert.Equal(t, maxIdle, len(pool.freeConn))
	assert.Equal(t, maxIdle, len(pool.connMap))
	pool.mu.Unlock()
}

func TestPoolPut(t *testing.T) {
	testcases := []struct {
		isClosed bool
		exceed   bool
		errMsg   string
	}{
		{false, false, ""},
		{true, false, "[99005]: connection to 127.0.0.1:9669 is closed"},
		{false, true, ""},
	}
	for _, tc := range testcases {
		p, err := NewNebulaPool("127.0.0.1:9669", "", "", withPoolConnector(&dummyConnector{}))
		if err != nil {
			t.Fatal(err)
		}
		defer p.Close()
		pool, _ := p.(*driverPool)
		pool.mu.Lock()
		pool.minOpen = 0
		pool.connCfg = &connConfig{
			requestTimeout: 10 * time.Second,
		}
		pool.mu.Unlock()
		c, err := pool.GetClient()
		if err != nil {
			t.Fatal(err)
		}
		if tc.isClosed {
			c.Close()
		}
		if tc.exceed {
			pool.maxOpen = 0
		}
		err = pool.PutClient(c)
		if tc.errMsg != "" {
			assert.EqualError(t, err, tc.errMsg)
		} else {
			assert.NoError(t, err)
		}
		pool.mu.Lock()
		if tc.exceed || tc.isClosed {
			assert.Equal(t, 0, len(pool.freeConn))
		} else {
			assert.Equal(t, 1, len(pool.freeConn))
		}
		pool.mu.Unlock()
		pool.Close()
	}
}

func TestPoolGet(t *testing.T) {
	testcases := []struct {
		concurrency int
		maxOpen     int
		runTimes    int
	}{
		{10, 10, 10},
		{20, 10, 30},
		{10, 30, 20},
	}
	for _, tc := range testcases {
		p, err := NewNebulaPool("127.0.0.1:9669", "", "",
			withPoolConnector(&dummyConnector{}),
			WithPoolMinOpenConns(0),
		)
		if err != nil {
			t.Fatal(err)
		}
		defer p.Close()
		pool, _ := p.(*driverPool)
		pool.maxOpen = tc.maxOpen
		pool.connCfg = &connConfig{
			requestTimeout: 10 * time.Second,
		}
		var eg errgroup.Group
		for i := 0; i < tc.concurrency; i++ {
			eg.Go(func() error {
				for j := 0; j < tc.runTimes; j++ {
					c, err := pool.GetClient()
					if err != nil {
						return err
					}
					<-time.After(10 * time.Millisecond)
					if err := pool.PutClient(c); err != nil {
						return err
					}
				}
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestPoolConcurrency(t *testing.T) {
	p, err := NewNebulaPool("127.0.0.1:9669", "", "",
		withPoolConnector(&dummyConnector{sleep: 2 * time.Millisecond}),
		WithPoolMaxOpenConns(400),
		WithPoolMaxWait(200*time.Millisecond),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	concurrency := 200
	loopsPerGoroutine := 1000
	var eg errgroup.Group
	for i := 0; i < concurrency; i++ {
		eg.Go(func() error {
			for l := 0; l < loopsPerGoroutine; l++ {
				run(t, p)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}
}

func run(t *testing.T, p types.Pool) {
	c, err := p.GetClient()
	if err != nil {
		t.Fatal(err)
	}
	defer p.PutClient(c)
	_, err = c.Execute("Return 1")

	if err != nil {
		t.Fatal(err)
	}
}

func TestPoolMaxWait(t *testing.T) {
	p, err := NewNebulaPool("127.0.0.1:9669", "", "",
		withPoolConnector(&dummyConnector{sleep: 2 * time.Second}),
		WithPoolMaxOpenConns(400),
		WithPoolMaxWait(200*time.Millisecond),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	// ignore validate connection
	_, _ = p.GetClient()
	_, err = p.GetClient()
	assert.True(t, err != nil)
	assert.Equal(t, "[99009]: Internal error, cannot get the valid connection", err.Error())
}

func TestPoolMaxLifeTime(t *testing.T) {
	p, err := NewNebulaPool("127.0.0.1:9669", "", "",
		withPoolConnector(&dummyConnector{sleep: 20 * time.Millisecond}),
		WithPoolMaxOpenConns(400),
		WithPoolMaxLifetime(100*time.Millisecond),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	c, err := p.GetClient()
	if err != nil {
		t.Fatal(err)
	}
	pool, _ := p.(*driverPool)
	pool.mu.Lock()
	assert.Equal(t, 1, len(pool.connMap))
	assert.Equal(t, 0, len(pool.freeConn))
	pool.mu.Unlock()
	<-time.After(200 * time.Millisecond)
	// the connection should be removed from pool
	if err := p.PutClient(c); err != nil {
		t.Fatal(err)
	}
	conn := c.(*driverConn)
	assert.True(t, conn.IsClosed())
	pool.mu.Lock()
	assert.Equal(t, 0, len(pool.freeConn))
	assert.Equal(t, 0, len(pool.connMap))
	pool.mu.Unlock()
}
