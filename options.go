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
	"crypto/tls"
	"fmt"
	"math"
	"time"
)

type ClientOptionsFn func(*driverConn)
type PoolOptionsFn func(*driverPool)

// WithClientConnectTimeout sets the timeout for connecting to the server
func WithClientConnectTimeout(timeout time.Duration) ClientOptionsFn {
	return func(ops *driverConn) {
		if timeout <= 0 {
			timeout = math.MaxInt64
		}
		ops.cfg.connectTimeout = timeout
	}
}

// WithClientRequestTimeout sets the timeout for a request to the server
func WithClientRequestTimeout(timeout time.Duration) ClientOptionsFn {
	return func(ops *driverConn) {
		if timeout <= 0 {
			timeout = math.MaxInt64
		}
		ops.cfg.requestTimeout = timeout
	}
}

// WithClientLogger sets the logger for the client
func WithClientLogger(logger Logger) ClientOptionsFn {
	return func(ops *driverConn) {
		ops.log = logger
	}
}

// used for testing
func withClientConnector(connector connector) ClientOptionsFn {
	return func(ops *driverConn) {
		ops.connector = connector
	}
}

// Experimental: TLS options
func WithClientTLS(ca, cert, key string, insecureSkipVerify bool) ClientOptionsFn {
	return func(conn *driverConn) {
		conn.cfg.enableTLS = true
		conn.cfg.ca = ca
		conn.cfg.cert = cert
		conn.cfg.key = key
		conn.cfg.insecureSkipVerify = insecureSkipVerify
	}
}

func WithClientTLSConfig(tlsConfig *tls.Config) ClientOptionsFn {
	return func(conn *driverConn) {
		conn.cfg.enableTLS = true
		conn.cfg.tlsConfig = tlsConfig
	}
}

// WithPoolConnectTimeout sets the timeout for connecting to the server
func WithPoolConnectTime(timeout time.Duration) PoolOptionsFn {
	return func(ops *driverPool) {
		if timeout <= 0 {
			timeout = math.MaxInt64
		}
		ops.connCfg.connectTimeout = timeout
	}
}

// WithPoolRequestTimeout sets the timeout for a request to the server
func WithPoolRequestTimeout(timeout time.Duration) PoolOptionsFn {
	return func(ops *driverPool) {
		if timeout <= 0 {
			timeout = math.MaxInt64
		}
		ops.connCfg.requestTimeout = timeout
	}
}

// WithPoolMaxWait sets max wait time
// that the pool will wait (when there are no available connections)
// for a connection to be returned
func WithPoolMaxWait(maxWait time.Duration) PoolOptionsFn {
	return func(ops *driverPool) {
		if maxWait <= 0 {
			maxWait = math.MaxInt64
		}
		ops.maxWait = maxWait
	}
}

// WithPoolMaxLifetime sets the maximum amount of time a connection may be reused
func WithPoolMaxLifetime(maxLifeTime time.Duration) PoolOptionsFn {
	return func(ops *driverPool) {
		if maxLifeTime <= 0 {
			maxLifeTime = math.MaxInt64
		}
		ops.connMaxLifeTime = maxLifeTime
	}
}

// WithPoolMinOpenConns sets the minimum number of open connections
func WithPoolMinOpenConns(minOpenConns int) PoolOptionsFn {
	return func(ops *driverPool) {
		ops.minOpen = minOpenConns
	}
}

// WithPoolMaxOpenConns sets the maximum number of open connections
func WithPoolMaxOpenConns(maxOpenConns int) PoolOptionsFn {
	return func(ops *driverPool) {
		ops.maxOpen = maxOpenConns
	}
}

// WithPoolMaxIdleConns sets the maximum number of idle connections
// would evicte the idle connections until the number of idle connections
func WithPoolMaxIdleConns(maxIdleConns int) PoolOptionsFn {
	return func(ops *driverPool) {
		ops.maxIdle = maxIdleConns
	}
}

// WithPoolTickerDuration sets the duration of the ticker
func WithPoolTickerDuration(ticker time.Duration) PoolOptionsFn {
	return func(ops *driverPool) {
		if ticker <= 0 {
			ticker = math.MaxInt64
		}
		ops.tickerDuration = ticker
	}
}

// WithPoolLogger sets the logger for the pool
func WithPoolLogger(logger Logger) PoolOptionsFn {
	return func(ops *driverPool) {
		ops.log = logger
	}
}

// WithPoolGraph sets the graph for the session
func WithPoolGraph(graph string) PoolOptionsFn {
	return func(ops *driverPool) {
		ops.sessionConfig.currentGraph = graph
	}
}

// WithPoolSchema sets the schema for the session
func WithPoolSchema(schema string) PoolOptionsFn {
	return func(ops *driverPool) {
		ops.sessionConfig.currentSchema = schema
	}
}

// WithPoolTimezone sets the timezone for the session
func WithPoolTimezone(timezone string) PoolOptionsFn {
	return func(ops *driverPool) {
		ops.sessionConfig.configs["timezone"] = fmt.Sprintf("\"%s\"", timezone)
	}
}

func WithPoolZonedDatetimeFormat(format string) PoolOptionsFn {
	return func(ops *driverPool) {
		ops.sessionConfig.configs["zoned_datetime_format"] = fmt.Sprintf("\"%s\"", format)
	}
}

func WithPoolZonedTimeFormat(format string) PoolOptionsFn {
	return func(ops *driverPool) {
		ops.sessionConfig.configs["zoned_time_format"] = fmt.Sprintf("\"%s\"", format)
	}
}

func WithPoolLocalTimeFormat(format string) PoolOptionsFn {
	return func(ops *driverPool) {
		ops.sessionConfig.configs["local_time_format"] = fmt.Sprintf("\"%s\"", format)
	}
}

func WithPoolLocalDatetimeFormat(format string) PoolOptionsFn {
	return func(ops *driverPool) {
		ops.sessionConfig.configs["local_datetime_format"] = fmt.Sprintf("\"%s\"", format)
	}
}

func WithPoolDateFormat(format string) PoolOptionsFn {
	return func(ops *driverPool) {
		ops.sessionConfig.configs["date_format"] = fmt.Sprintf("\"%s\"", format)
	}
}

// WithPoolParameters sets the parameters for the session
func WithPoolParameters(parameters map[string]string) PoolOptionsFn {
	return func(ops *driverPool) {
		ops.sessionConfig.parameters = parameters
	}
}

func WithPoolSessionConfigs(configs map[string]string) PoolOptionsFn {
	return func(ops *driverPool) {
		for k, v := range configs {
			ops.sessionConfig.configs[k] = v
		}
	}
}

func WithPoolExecuteOnOpenSession(statements []string) PoolOptionsFn {
	return func(ops *driverPool) {
		ops.sessionConfig.preStatements = statements
	}
}

// WithPoolStrictlyServerHealthy sets the pool to strictly check the server health
// if strictly is false, the pool will be created successfully if any of the servers is healthy
// if strictly is true, the pool will be created successfully only if all of the servers are healthy
// default is false
func WithPoolStrictlyServerHealthy(strictly bool) PoolOptionsFn {
	return func(ops *driverPool) {
		ops.strictlyServerHealthy = strictly
	}
}

// used for testing
func withPoolConnector(connector connector) PoolOptionsFn {
	return func(ops *driverPool) {
		ops.connector = connector
	}
}

// TLS options
func WithPoolTLS(ca, cert, key string, insecureSkipVerify bool) PoolOptionsFn {
	return func(pool *driverPool) {
		pool.connCfg.enableTLS = true
		pool.connCfg.ca = ca
		pool.connCfg.cert = cert
		pool.connCfg.key = key
		pool.connCfg.insecureSkipVerify = insecureSkipVerify
	}
}

func WithPoolTLSConfig(tlsConfig *tls.Config) PoolOptionsFn {
	return func(pool *driverPool) {
		pool.connCfg.enableTLS = true
		pool.connCfg.tlsConfig = tlsConfig
	}
}

// WithPoolWithOutPing sets the pool to not ping the server when getting a connection
// default is false, meaning the pool will ping the server when getting a connection
func WithPoolWithOutPing() PoolOptionsFn {
	return func(pool *driverPool) {
		pool.withoutPing = true
	}
}

// WithPoolPingTimeout sets the timeout for pinging the server
func WithPoolPingTimeout(timeout time.Duration) PoolOptionsFn {
	return func(pool *driverPool) {
		if timeout <= 0 {
			timeout = math.MaxInt64
		}
		pool.pingTimeout = timeout
	}
}

// WithClientAuthInfo sets the authInfo for the client connection
func WithClientAuthInfo(authInfo map[string]string) ClientOptionsFn {
	return func(ops *driverConn) {
		ops.cfg.authInfo = authInfo
	}
}

// WithPoolAuthInfo sets the authInfo for the pool connection
func WithPoolAuthInfo(authInfo map[string]string) PoolOptionsFn {
	return func(ops *driverPool) {
		ops.connCfg.authInfo = authInfo
	}
}
