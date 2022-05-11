/*
 *
 * Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 *
 */
package nebula_go

import (
	"crypto/tls"
	"fmt"
)

var (
	_ SessionGetter = (*ConnectionPool)(nil)
)

// ConnectionOption type.
type ConnectionOption func(*ConnectionConfig)

// WithTLSConfig functional option to set the tls configuration.
// use ClientConfigForX509 or GetDefaultSSLConfig to build based on files.
func WithTLSConfig(tlsConfig *tls.Config) ConnectionOption {
	return func(cfg *ConnectionConfig) {
		cfg.TLSConfig = tlsConfig
	}
}

// WithLogger functional option to substitute the default logger.
func WithLogger(log Logger) ConnectionOption {
	return func(cfg *ConnectionConfig) {
		cfg.Log = log
	}
}

// WithCredentials functional option to set a pair of username and password.
func WithCredentials(username, password string) ConnectionOption {
	return func(cfg *ConnectionConfig) {
		cfg.Username = username
		cfg.Password = password
	}
}

// WithConnectionPoolBuilder functional option allow to use a custom connection pool.
func WithConnectionPoolBuilder(connectionPoolBuilder ConnectionPoolBuilder) ConnectionOption {
	return func(cfg *ConnectionConfig) {
		cfg.ConnectionPoolBuilder = connectionPoolBuilder
	}
}

// WithConnectionPoolConfig functional option to override the connection pool configuration.
func WithConnectionPoolConfig(poolConfig PoolConfig) ConnectionOption {
	return func(cfg *ConnectionConfig) {
		cfg.PoolConfig = poolConfig
	}
}

// SessionGetter interface.
type SessionGetter interface {
	GetSession(username, password string) (*Session, error)
	Close()
}

// SessionPool type.
type SessionPool struct {
	ConnectionPool SessionGetter
	Username       string
	Password       string
	Space          string
	Log            Logger
}

// NewSessionPool constructs a new session pool using the given connection string and options.
// will use DefaultLogger as logger and no ssl by default
// Instead:
//   pool, err := NewSslConnectionPool(addresses, conf, tlsConfig, log)
//   session, err := pool.GetSession(user,pass)
// You can:
//   sessionPool, err := NewSessionPool("nebula://user:pass@host:port/space?options", ...)
//   session, err := sessionPool.Acquire()
// See: ParseConnectionString
func NewSessionPool(connectionString string, opts ...ConnectionOption) (*SessionPool, error) {
	cfg, err := ParseConnectionString(connectionString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %s", err.Error())
	}

	return newSessionFromFromConnectionConfig(cfg, opts...)
}

// NewSessionPoolFromHostAddresses uses the existing []HostAddress to build a session pool
// Instead:
//   pool, err := NewSslConnectionPool(addresses, conf, tlsConfig, log)
//   session, err := pool.GetSession(user,pass)
// You can:
//   sessionPool, err := NewSessionPoolFromHostAddresses(address,
//      WithTLSConfig(tlsConfig),
//      WithLogger(log)
//      WithConnectionPoolConfig(conf),
//      WithCredentials(user,pass),
//   )
//   session, err := sessionPool.Acquire()
func NewSessionPoolFromHostAddresses(addresses []HostAddress, opts ...ConnectionOption) (*SessionPool, error) {
	cfg := &ConnectionConfig{
		HostAddresses: addresses,
	}

	return newSessionFromFromConnectionConfig(cfg, opts...)
}

func newSessionFromFromConnectionConfig(cfg *ConnectionConfig, opts ...ConnectionOption) (*SessionPool, error) {

	cfg.ConnectionPoolBuilder = defaultConnectionPoolBuilder
	cfg.Log = DefaultLogger{}

	for _, opt := range opts {
		opt(cfg)
	}

	connPool, err := cfg.ConnectionPoolBuilder(cfg.HostAddresses, cfg.PoolConfig, cfg.TLSConfig, cfg.Log)
	if err != nil {
		return nil, fmt.Errorf("unable to build connection pool: %s", err.Error())
	}

	return &SessionPool{
		ConnectionPool: connPool,
		Username:       cfg.Username,
		Password:       cfg.Password,
		Space:          cfg.Space,
		Log:            cfg.Log,
	}, nil
}

// Acquire return an authenticated session or return an error.
func (s *SessionPool) Acquire() (*Session, error) {
	session, err := s.ConnectionPool.GetSession(s.Username, s.Password)
	if err != nil {
		return nil, fmt.Errorf("unable to get session: %s", err.Error())
	}

	if session.log == nil {
		session.log = s.Log // inject logger when it came from other connection pool
	}

	return session, nil
}

// WithSession execute a callback with an authenticated session, releasing it in the end.
func (s *SessionPool) WithSession(callback func(session *Session) error) error {
	session, err := s.Acquire()
	if err != nil {
		return err
	}

	defer session.Release()

	return callback(session)
}

// Close method.
func (s *SessionPool) Close() error {
	s.ConnectionPool.Close()

	return nil
}

func defaultConnectionPoolBuilder(addresses []HostAddress,
	conf PoolConfig,
	sslConfig *tls.Config,
	log Logger) (SessionGetter, error) {
	return NewSslConnectionPool(addresses, conf, sslConfig, log)
}
