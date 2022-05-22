/*
 *
 * Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 *
 */
package nebula_go

import (
	"container/list"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

var (
	_ SessionGetter = (*connectionPoolWrapper)(nil)

	_ NebulaSession = (*Session)(nil)

	_ Releaser = (*Session)(nil)
)

// SessionGetter interface.
type SessionGetter interface {
	GetSession(username, password string) (NebulaSession, error)
	Close()
}

// NebulaSession subset interface of Session.
type NebulaSession interface {
	Execute(stmt string) (*ResultSet, error)
	ExecuteWithParameter(stmt string, params map[string]interface{}) (*ResultSet, error)
	ExecuteJson(stmt string) ([]byte, error)
	ExecuteJsonWithParameter(stmt string, params map[string]interface{}) ([]byte, error)
	GetSessionID() int64
}

// Releaser interface.
type Releaser interface {
	Release()
}

type connectionPoolWrapper struct {
	*ConnectionPool
}

// GetSession method adapter.
func (cpw *connectionPoolWrapper) GetSession(username, password string) (NebulaSession, error) {
	session, err := cpw.ConnectionPool.GetSession(username, password)
	if err != nil {
		return nil, err
	}

	session.log = cpw.ConnectionPool.log

	return session, nil
}

// SessionPool type.
type SessionPool struct {
	ConnectionPool   SessionGetter
	Username         string
	Password         string
	Space            string
	OnAcquireSession string
	Log              Logger
	sessionQueue     *sessionQueue
	sync.Mutex
}

// SessionPoolConfig type.
type SessionPoolConfig struct {
	// The max idle sessions in pool
	MaxIdleSessionPoolSize int
}

func (conf *SessionPoolConfig) validateConf(log Logger) {
	if conf.MaxIdleSessionPoolSize < 0 {
		conf.MaxIdleSessionPoolSize = defaultMaxIdleSessionPoolSize
		log.Warn(fmt.Sprintf("Invalid MaxIdleSessionPoolSize value, the default value of %d has been applied",
			defaultMaxIdleSessionPoolSize))
	}
}

const defaultMaxIdleSessionPoolSize = 0

// GetDefaultSessionPoolConfig return the default session pool configuration.
func GetDefaultSessionPoolConfig() SessionPoolConfig {
	return SessionPoolConfig{
		MaxIdleSessionPoolSize: defaultMaxIdleSessionPoolSize,
	}
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

	return newSessionFromFromConnectionConfig(cfg, nil, opts...)
}

// NewSessionPoolFromHostAddresses uses the existing []HostAddress to build a session pool
// Instead:
//   pool, err := NewSslConnectionPool(addresses, conf, tlsConfig /*or nil*/, log)
//   session, err := pool.GetSession(user,pass)
// You can:
//   sessionPool, err := NewSessionPoolFromHostAddresses(address,
//      WithTLSConfig(tlsConfig), // omit for no tls configuration
//      WithLogger(log),          // or use WithDefaultLogger()
//      WithConnectionPoolConfig(conf),
//      WithCredentials(user,pass),
//   )
//   session, err := sessionPool.Acquire()
func NewSessionPoolFromHostAddresses(addresses []HostAddress, opts ...ConnectionOption) (*SessionPool, error) {
	cfg := &ConnectionConfig{
		HostAddresses: addresses,
	}

	return newSessionFromFromConnectionConfig(cfg, nil, opts...)
}

// NewSessionPoolFromConnectionPool is an alternative way to build a session pool
// from an existing connection pool object.
// Usage:
//   pool, err := NewSslConnectionPool(addresses, conf, tlsConfig, log)
//   sessionPool, err := NewSessionPoolFromConnectionPool(pool,
//      WithCredentials(user,pass),
//   )
//   session, err := sessionPool.Acquire()
func NewSessionPoolFromConnectionPool(connPool *ConnectionPool, opts ...ConnectionOption) (*SessionPool, error) {
	cfg := &ConnectionConfig{
		HostAddresses: connPool.addresses,
	}

	sessionGetter := &connectionPoolWrapper{
		ConnectionPool: connPool,
	}
	return newSessionFromFromConnectionConfig(cfg, sessionGetter, opts...)
}

func newSessionFromFromConnectionConfig(cfg *ConnectionConfig,
	connPool SessionGetter,
	opts ...ConnectionOption) (*SessionPool, error) {
	cfg.Apply(opts)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	if connPool == nil {
		var err error
		connPool, err = cfg.BuildConnectionPool()
		if err != nil {
			return nil, fmt.Errorf("unable to build connection pool: %s", err.Error())
		}
	}

	var queue *sessionQueue

	if max := cfg.SessionPoolConfig.MaxIdleSessionPoolSize; max > 0 {
		queue = &sessionQueue{
			max: max,
		}
	}

	onAcquireSession := expandMacros(cfg.OnAcquireSession, map[string]string{
		"%SPACE%": cfg.Space,
	})

	return &SessionPool{
		ConnectionPool:   connPool,
		Username:         cfg.Username,
		Password:         cfg.Password,
		Space:            cfg.Space,
		OnAcquireSession: onAcquireSession,
		Log:              cfg.Log,
		sessionQueue:     queue,
	}, nil
}

func expandMacros(str string, macros map[string]string) string {
	if str == "" {
		return str
	}

	for macro, value := range macros {
		str = strings.ReplaceAll(str, macro, value)
	}

	return str
}

// Acquire return an authenticated session or return an error.
// If MaxIdleSessionPoolSize is not zero, we may return an existing session.
// Usage:
//   session, err := sessionPool.Acquire()
//   ... use session ...
//   sessionPool.Release(session)
func (s *SessionPool) Acquire() (session NebulaSession, err error) {
	var ok bool
	session, ok = s.sessionQueue.Dequeue()
	if ok {
		return session, nil
	}

	session, err = s.ConnectionPool.GetSession(s.Username, s.Password)
	if err != nil {
		return nil, fmt.Errorf("unable to get session: %s", err.Error())
	}

	err = s.executeStatement(session, s.OnAcquireSession)
	if err != nil {
		return nil, fmt.Errorf("unable to execute statement %q on acquire session id=%#x: %s",
			s.OnAcquireSession, session.GetSessionID(), err)
	}

	return session, nil
}

func (s *SessionPool) executeStatement(session NebulaSession, stmt string) error {
	if stmt == "" {
		return nil
	}

	rs, err := session.Execute(stmt)
	if err != nil {
		s.release(session)

		return err
	}

	if !rs.IsSucceed() {
		s.release(session)

		return fmt.Errorf("not succeed: %q (error code %d)", rs.GetErrorMsg(), rs.GetErrorCode())
	}

	return nil
}

// Release will handle the release of the nebula session.
// If MaxIdleSessionPoolSize is not zero, we may keep the session in memory until Close.
func (s *SessionPool) Release(session NebulaSession) {
	if session == nil {
		return
	}

	oldest := s.sessionQueue.Enqueue(session)

	s.release(oldest)
}

func (s *SessionPool) release(session NebulaSession) {
	if session == nil {
		return
	}

	sessionID := session.GetSessionID()

	if releaser, ok := session.(Releaser); ok {
		s.Log.Info("releasing session id: Ox" + strconv.FormatInt(sessionID, 16))

		releaser.Release()

		return
	}

	s.Log.Warn(fmt.Sprintf("unable to release session id %d: no method release available on %T",
		sessionID, session))

}

// WithSession execute a callback with an authenticated session, releasing it in the end.
// Usage:
//   err := sessionPool.WithSession(func(session NebulaSession)error{
//	    ... use session ... no need for release
//      return nil // or an error depending on the logic
//   })
// Equivalent of
//   session, err := sessionPool.Acquire()
//   ... use session ...
//   sessionPool.Release(session)
func (s *SessionPool) WithSession(callback func(session NebulaSession) error) error {
	session, err := s.Acquire()
	if err != nil {
		return err
	}

	defer s.Release(session)

	return callback(session)
}

// Close will release all remaining sessions and close the connection pool.
func (s *SessionPool) Close() error {
	s.releaseAllRemainingSessions(func() {
		s.Log.Info("closing connection pool")

		s.ConnectionPool.Close()
	})

	return nil
}

func (s *SessionPool) releaseAllRemainingSessions(finally func()) {
	s.sessionQueue.ForEach(s.release, finally)
}

type sessionQueue struct {
	sync.Mutex
	queue list.List
	max   int
}

func (q *sessionQueue) Dequeue() (NebulaSession, bool) {
	if q == nil || q.max == 0 {
		return nil, false
	}

	q.Lock()
	defer q.Unlock()

	return q.dequeue()
}

func (q *sessionQueue) dequeue() (NebulaSession, bool) {
	front := q.queue.Front()
	if front != nil {
		session, _ := q.queue.Remove(front).(NebulaSession)
		return session, true
	}

	return nil, false
}

func (q *sessionQueue) Enqueue(session NebulaSession) (oldest NebulaSession) {
	if session == nil || q == nil || q.max == 0 {
		return session
	}

	q.Lock()
	defer q.Unlock()

	if q.max > 0 && q.queue.Len() == q.max {
		front := q.queue.Front()
		oldest, _ = q.queue.Remove(front).(NebulaSession)
	}

	q.queue.PushBack(session)

	return
}

func (q *sessionQueue) ForEach(callback func(session NebulaSession), finally func()) {
	if q == nil || q.max == 0 {
		finally()

		return
	}

	q.Lock()
	defer q.Unlock()

	defer finally()

	for {
		session, ok := q.dequeue()
		if !ok {
			return
		}

		callback(session)
	}
}
