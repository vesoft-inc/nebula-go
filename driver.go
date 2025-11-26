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
	"crypto/tls"
	"math"
	"math/rand"
	"time"

	"github.com/vesoft-inc/nebula-go/v5/internal/internal_error"
	"github.com/vesoft-inc/nebula-go/v5/pkg/types"
)

type (
	// an internal interface to get the connection
	connector interface {
		connect(address string, cfg *connConfig) (types.Client, error)
	}
	connConfig struct {
		username           string
		password           string
		graph              string
		requestTimeout     time.Duration
		connectTimeout     time.Duration
		timezone           string
		enableTLS          bool
		cert               string
		key                string
		ca                 string
		insecureSkipVerify bool
		authInfo           map[string]string
		tlsConfig          *tls.Config
	}
	// driverConn is a client wrapper
	driverConn struct {
		cfg            *connConfig
		hostAddresses  []string
		currentAddress string
		connector      connector
		pool           *driverPool
		createAt       time.Time
		conn           types.Client
		isClosed       bool
		log            Logger
	}
)

const (
	defaultMaxOpenConns   = 100
	defaultMinOpenConns   = 1
	defaultMaxIdleConns   = 5
	defaultMaxLieTime     = 30 * time.Minute
	defaultRequestTimeout = 1 * time.Minute
	defaultConnectTimeout = 3 * time.Second
	defaultTicker         = 5 * time.Second
	defaultMaxWait        = math.MaxInt64
)

func NewNebulaClient(addresses, username, password string, opts ...ClientOptionsFn) (types.Client, error) {
	hostAddresses, err := parseAddresses(addresses)
	if err != nil {
		return nil, err
	}
	cfg := newConnConfig(username, password)
	dc := &driverConn{
		hostAddresses: hostAddresses,
		connector:     defaultConnector,
		cfg:           cfg,
	}
	for _, o := range opts {
		o(dc)
	}
	if err := dc.open(); err != nil {
		return nil, err
	}

	return dc, nil
}

func NewNebulaPool(addresses, username, password string, opts ...PoolOptionsFn) (types.Pool, error) {
	hostAddresses, err := parseAddresses(addresses)
	if err != nil {
		return nil, err
	}
	connCfg := newConnConfig(username, password)
	ctx, cancel := context.WithCancel(context.Background())
	pool := &driverPool{
		ctx:             ctx,
		stop:            cancel,
		hostAddresses:   hostAddresses,
		connCfg:         connCfg,
		connMap:         make(map[types.Client]struct{}),
		requestConnChan: make(map[uint64]chan types.Client),
		requestCount:    0,
		openerCh:        make(chan struct{}, openConnChannelSize),
		connector:       defaultConnector,
		connMaxLifeTime: defaultMaxLieTime,
		maxOpen:         defaultMaxOpenConns,
		minOpen:         defaultMinOpenConns,
		maxIdle:         defaultMaxIdleConns,
		maxWait:         defaultMaxWait,
		sessionConfig: &sessionConfig{
			configs:    make(map[string]string),
			parameters: make(map[string]string),
		},
		tickerDuration: defaultTicker,
		log:            DefaultLogger,
		pingTimeout:    1 * time.Second,
	}
	for _, o := range opts {
		o(pool)
	}
	var (
		succeeded = 0
		dc        types.Client
	)
	for _, address := range pool.hostAddresses {
		if !pool.strictlyServerHealthy && succeeded > 0 {
			break
		}
		dc, err = pool.openNewConn(address)
		if err != nil {
			if pool.strictlyServerHealthy {
				break
			}
		} else {
			succeeded++
			pool.putNewConn(dc)
		}
	}
	if succeeded == 0 || (pool.strictlyServerHealthy && succeeded != len(pool.hostAddresses)) {
		return nil, err
	}

	go pool.connectionOpener(ctx)
	// start ticker
	go pool.ticker(pool.ctx)
	return pool, nil
}

func newConnConfig(username, password string) *connConfig {
	return &connConfig{
		username:       username,
		password:       password,
		requestTimeout: defaultRequestTimeout,
		connectTimeout: defaultConnectTimeout,
	}
}

func (dc *driverConn) Execute(stmt string) (types.Result, error) {
	return dc.conn.Execute(stmt)
}

func (dc *driverConn) ExecuteContext(ctx context.Context, stmt string) (types.Result, error) {
	if dc.isClosed {
		return nil, internal_error.ErrConnIsClosed(dc.currentAddress)
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	result, err := dc.conn.ExecuteContext(ctx, stmt)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (dc *driverConn) Ping() error {
	if dc.IsClosed() {
		return internal_error.ErrConnIsClosed(dc.currentAddress)
	}
	return dc.conn.Ping()
}

func (dc *driverConn) PingContext(ctx context.Context) error {
	if dc.IsClosed() {
		return internal_error.ErrConnIsClosed(dc.currentAddress)
	}
	return dc.conn.PingContext(ctx)
}

func (dc *driverConn) Close() error {
	if dc.IsClosed() {
		return nil
	}
	err := dc.conn.Close()
	if err != nil {
		return err
	}
	dc.isClosed = true
	dc.conn = nil
	return nil
}

func (dc *driverConn) open() error {
	// random select one host
	if dc.currentAddress == "" {
		rand.Seed(time.Now().UnixNano())
		hostIndex := rand.Intn(len(dc.hostAddresses))
		dc.currentAddress = dc.hostAddresses[hostIndex]
	}

	conn, err := dc.connector.connect(dc.currentAddress, dc.cfg)
	if err != nil {
		return err
	}
	dc.createAt = time.Now()
	dc.conn = conn
	return nil
}

func (dc *driverConn) replaceFromPool() error {
	if dc.pool == nil {
		return nil
	}
	// return the old connection
	// pool would delete it if invalid.
	_ = dc.pool.PutClient(dc)
	conn, err := dc.pool.GetClient()
	if err != nil {
		return err
	}
	newConn, ok := conn.(*driverConn)
	if !ok {
		// not reachable
		return internal_error.ErrInternal("invalid connection type")
	}
	dc.conn = newConn.conn
	dc.currentAddress = newConn.currentAddress
	dc.createAt = newConn.createAt
	dc.isClosed = newConn.isClosed
	return nil
}

func (dc *driverConn) GetSessionId() (int64, error) {
	if dc.IsClosed() {
		return 0, internal_error.ErrConnIsClosed(dc.currentAddress)
	}

	return dc.conn.GetSessionId()
}

func (dc *driverConn) GetVersion() (string, error) {
	if dc.IsClosed() {
		return "", internal_error.ErrConnIsClosed(dc.currentAddress)
	}

	return dc.conn.GetVersion()
}

func (dc *driverConn) IsClosed() bool {
	return dc.conn == nil || dc.isClosed
}
