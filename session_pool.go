/*
 *
 * Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 *
 */
package nebula_go

import (
	"bytes"
	"container/list"
	"fmt"
	"strconv"
	"sync"
	"text/template"

	"github.com/vesoft-inc/nebula-go/v3/nebula"
	"github.com/vesoft-inc/nebula-go/v3/nebula/graph"
)

var (
	_ SessionGetter = (*connectionPoolWrapper)(nil)

	_ NebulaSession = (*Session)(nil)
)

// SessionGetter interface.
type SessionGetter interface {
	GetSession(username, password string) (NebulaSession, error)
	Close()
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
	OnReleaseSession string
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

	type Data struct {
		Space string
	}

	data := Data{
		Space: cfg.Space,
	}

	onAcquireSession, err := processTemplate("onAcquire", cfg.OnAcquireSession, data)
	if err != nil {
		return nil, err
	}

	onReleaseSession, err := processTemplate("onRelease", cfg.OnReleaseSession, data)
	if err != nil {
		return nil, err
	}

	return &SessionPool{
		ConnectionPool:   connPool,
		Username:         cfg.Username,
		Password:         cfg.Password,
		Space:            cfg.Space,
		OnAcquireSession: onAcquireSession,
		OnReleaseSession: onReleaseSession,
		Log:              cfg.Log,
		sessionQueue:     queue,
	}, nil
}

func processTemplate(name, raw string, data interface{}) (string, error) {
	if raw == "" {
		return "", nil
	}

	t, err := template.New(name).Parse(raw)
	if err != nil {
		return "", fmt.Errorf("unable to process template %s: %v", name, err)
	}

	var buff bytes.Buffer
	err = t.Execute(&buff, data)
	if err != nil {
		return "", fmt.Errorf("unable to execute template %s: %v", name, err)
	}

	return buff.String(), nil
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

	if s.OnAcquireSession != "" {
		rs, err := session.Execute(s.OnAcquireSession)
		if err != nil {
			s.release(session)

			return nil, fmt.Errorf("unable to execute statement %q on acquire session: %v",
				s.OnAcquireSession, err)
		}

		if rs.resp == nil {
			rs.resp = &graph.ExecutionResponse{
				ErrorCode: nebula.ErrorCode_E_UNKNOWN,
				ErrorMsg:  []byte("unknown error"),
			}
		}

		if !rs.IsSucceed() {
			s.release(session)

			return nil, fmt.Errorf("execute statement %q on acquire session does not succeed: %s (error code %d)",
				s.OnAcquireSession, rs.GetErrorMsg(), rs.GetErrorCode())
		}
	}

	return session, nil
}

// Release will handle the release of the nebula session.
// If MaxIdleSessionPoolSize is not zero, we may keep the session in memory until Close.
func (s *SessionPool) Release(session NebulaSession) error {
	if session == nil {
		return nil
	}

	if s.OnReleaseSession != "" {
		rs, err := session.Execute(s.OnReleaseSession)
		if err != nil {
			s.release(session)

			return fmt.Errorf("unable to execute statement %q on release session: %v",
				s.OnReleaseSession, err)
		}

		if rs.resp == nil {
			rs.resp = &graph.ExecutionResponse{
				ErrorCode: nebula.ErrorCode_E_UNKNOWN,
				ErrorMsg:  []byte("unknown error"),
			}
		}

		if !rs.IsSucceed() {
			s.release(session)

			return fmt.Errorf("unable to execute statement %q on release session: %s (error code %d)",
				s.OnReleaseSession, rs.GetErrorMsg(), rs.GetErrorCode())
		}
	}

	oldest := s.sessionQueue.Enqueue(session)

	s.release(oldest)

	return nil
}

func (s *SessionPool) release(session NebulaSession) {
	if session == nil {
		return
	}

	if releaser, ok := session.(interface{ Release() }); ok {
		s.Log.Info("releasing session id: " + strconv.FormatInt(session.GetSessionID(), 16))

		releaser.Release()
	} else {
		s.Log.Warn(fmt.Sprintf("unable to release session id %d: no method release on %T",
			session.GetSessionID(), session))
	}
}

// NebulaSession subset interface of Session.
type NebulaSession interface {
	Execute(stmt string) (*ResultSet, error)
	ExecuteWithParameter(stmt string, params map[string]interface{}) (*ResultSet, error)
	ExecuteJson(stmt string) ([]byte, error)
	ExecuteJsonWithParameter(stmt string, params map[string]interface{}) ([]byte, error)
	GetSessionID() int64
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

	defer func() {
		if err := s.Release(session); err != nil {
			s.Log.Warn(fmt.Sprintf("unexpected error while release session id %d: %v",
				session.GetSessionID(), err))
		}
	}()

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
