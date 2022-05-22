/*
 *
 * Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 *
 */
package nebula_go_test

import (
	"crypto/tls"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	nebula_go "github.com/vesoft-inc/nebula-go/v3"
)

func TestSessionPool(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		label      string
		connString string

		opts []nebula_go.ConnectionOption

		prepare func(*mockConnectionPoolBuilder, *mockSessionGetter, *mockSession)
		verify  func(*testing.T, *nebula_go.SessionPool, nebula_go.NebulaSession)

		errMsg string
	}{
		{
			label:      "cant create session pool with bad connection string",
			connString: "mysql://host",
			errMsg:     "unable to parse connection string: connection string must start with \"nebula\":// instead \"mysql\"",
		},
		{
			label:      "should return error from connection pool constructor",
			connString: "nebula://localhost",
			prepare: func(connPollBuilder *mockConnectionPoolBuilder,
				_ *mockSessionGetter,
				_ *mockSession,
			) {
				connPollBuilder.On("CALL",
					[]nebula_go.HostAddress{
						{Host: "localhost", Port: 9669},
					},
					nebula_go.GetDefaultConf(),
					(*tls.Config)(nil),
					nebula_go.NoLogger{}).Return(nil, errors.New("ops"))
			},
			errMsg: "unable to build connection pool: ops",
		},
		{
			label:      "should return error from connection pool constructor with tls skip-verify",
			connString: "nebula://localhost?tls=skip-verify",
			prepare: func(connPollBuilder *mockConnectionPoolBuilder,
				_ *mockSessionGetter,
				_ *mockSession,
			) {
				connPollBuilder.On("CALL",
					[]nebula_go.HostAddress{
						{Host: "localhost", Port: 9669},
					},
					nebula_go.GetDefaultConf(),
					&tls.Config{InsecureSkipVerify: true},
					nebula_go.NoLogger{}).Return(nil, errors.New("ops"))
			},
			errMsg: "unable to build connection pool: ops",
		},
		{
			label:      "should return a session pool and close should call connection pool close",
			connString: "nebula://localhost?timeout=5s",
			prepare: func(connPollBuilder *mockConnectionPoolBuilder,
				sessGetter *mockSessionGetter,
				_ *mockSession,
			) {
				sessGetter.On("Close").Return()

				poolConf := nebula_go.GetDefaultConf()
				poolConf.TimeOut = 5 * time.Second

				connPollBuilder.On("CALL",
					[]nebula_go.HostAddress{
						{Host: "localhost", Port: 9669},
					},
					poolConf,
					(*tls.Config)(nil),
					nebula_go.NoLogger{}).Return(sessGetter, nil)
			},
			verify: func(t *testing.T, sessPool *nebula_go.SessionPool, _ nebula_go.NebulaSession) {
				sessPool.Close()
			},
		},
		{
			label:      "should return a session pool and but fail to return an authenticated session",
			connString: "nebula://user:pass@localhost",
			opts: []nebula_go.ConnectionOption{
				nebula_go.WithConnectionPoolConfig(nebula_go.PoolConfig{
					IdleTime: 5 * time.Second,
				}),
			},
			prepare: func(connPollBuilder *mockConnectionPoolBuilder,
				sessGetter *mockSessionGetter,
				_ *mockSession,
			) {
				sessGetter.On("GetSession", "user", "pass").Return(nil, errors.New("ops"))

				connPollBuilder.On("CALL",
					[]nebula_go.HostAddress{
						{Host: "localhost", Port: 9669},
					},
					nebula_go.PoolConfig{
						IdleTime: 5 * time.Second,
					},
					(*tls.Config)(nil),
					nebula_go.NoLogger{}).Return(sessGetter, nil)
			},
			verify: func(t *testing.T, sessPool *nebula_go.SessionPool, _ nebula_go.NebulaSession) {
				got, err := sessPool.Acquire()
				assert.Nil(t, got)
				assert.EqualError(t, err, "unable to get session: ops")
			},
		},
		{
			label: "should return a session pool and return an authenticated session with different credentials",
			opts: []nebula_go.ConnectionOption{
				nebula_go.WithCredentials("foo", "bar"),
			},
			connString: "nebula://user:pass@localhost",
			prepare: func(connPollBuilder *mockConnectionPoolBuilder,
				sessGetter *mockSessionGetter,
				session *mockSession,
			) {
				sessGetter.On("GetSession", "foo", "bar").Return(session, nil)

				connPollBuilder.On("CALL",
					[]nebula_go.HostAddress{
						{Host: "localhost", Port: 9669},
					},
					nebula_go.GetDefaultConf(),
					(*tls.Config)(nil),
					nebula_go.NoLogger{}).Return(sessGetter, nil)

				session.On("Release").Return()
			},
			verify: func(t *testing.T, sessPool *nebula_go.SessionPool, session nebula_go.NebulaSession) {
				got, err := sessPool.Acquire()
				assert.Nil(t, err)
				assert.Equal(t, session, got)

				sessPool.Release(got)
			},
		},
		{
			label: "should return error when execute default on acquire session stmt",
			opts: []nebula_go.ConnectionOption{
				nebula_go.WithCredentials("foo", "bar"),
			},
			connString: "nebula://user:pass@localhost/space",
			prepare: func(connPollBuilder *mockConnectionPoolBuilder,
				sessGetter *mockSessionGetter,
				session *mockSession,
			) {
				sessGetter.On("GetSession", "foo", "bar").Return(session, nil)

				connPollBuilder.On("CALL",
					[]nebula_go.HostAddress{
						{Host: "localhost", Port: 9669},
					},
					nebula_go.GetDefaultConf(),
					(*tls.Config)(nil),
					nebula_go.NoLogger{}).Return(sessGetter, nil)

				session.On("Execute", "USE space;").Return(nil, errors.New("ops"))
				session.On("Release").Return()
			},
			verify: func(t *testing.T, sessPool *nebula_go.SessionPool, session nebula_go.NebulaSession) {
				got, err := sessPool.Acquire()

				assert.EqualError(t, err, "unable to execute statement \"USE space;\" on acquire session id=0x1: ops")
				assert.Nil(t, got)
			},
		},
		{
			label:      "should return error with bad space name",
			connString: "nebula://localhost/skip-verify",
			prepare: func(_ *mockConnectionPoolBuilder,
				_ *mockSessionGetter,
				_ *mockSession,
			) {
			},
			errMsg: "unable to parse connection string: space name \"skip-verify\" is not valid",
		},
		{
			label: "should return failure when execute default on acquire session stmt",
			opts: []nebula_go.ConnectionOption{
				nebula_go.WithCredentials("foo", "bar"),
				nebula_go.WithOnAcquireSessionStmt(
					`CREATE SPACE IF NOT EXISTS %SPACE%(vid_type=FIXED_STRING(20)); USE %SPACE%;`,
				),
			},
			connString: "nebula://user:pass@localhost/space",
			prepare: func(connPollBuilder *mockConnectionPoolBuilder,
				sessGetter *mockSessionGetter,
				session *mockSession,
			) {
				sessGetter.On("GetSession", "foo", "bar").Return(session, nil)

				connPollBuilder.On("CALL",
					[]nebula_go.HostAddress{
						{Host: "localhost", Port: 9669},
					},
					nebula_go.GetDefaultConf(),
					(*tls.Config)(nil),
					nebula_go.NoLogger{}).Return(sessGetter, nil)

				rs := &nebula_go.ResultSet{}
				session.On("Execute",
					"CREATE SPACE IF NOT EXISTS space(vid_type=FIXED_STRING(20)); USE space;").Return(rs, nil)

				session.On("Release").Return()
			},
			verify: func(t *testing.T, sessPool *nebula_go.SessionPool, session nebula_go.NebulaSession) {
				got, err := sessPool.Acquire()

				assert.EqualError(t, err,
					"unable to execute statement "+
						"\"CREATE SPACE IF NOT EXISTS space(vid_type=FIXED_STRING(20)); USE space;\""+
						" on acquire session id=0x1: not succeed: \"\" (error code -8000)")
				assert.Nil(t, got)
			},
		},
		{
			label: "should return a session pool and return an authenticated with max session pool",
			opts: []nebula_go.ConnectionOption{
				nebula_go.WithCredentials("foo", "bar"),
			},
			connString: "nebula://user:pass@localhost?max_idle_session_pool_size=10",
			prepare: func(connPollBuilder *mockConnectionPoolBuilder,
				sessGetter *mockSessionGetter,
				session *mockSession,
			) {
				sessGetter.On("GetSession", "foo", "bar").Return(session, nil)

				connPollBuilder.On("CALL",
					[]nebula_go.HostAddress{
						{Host: "localhost", Port: 9669},
					},
					nebula_go.GetDefaultConf(),
					(*tls.Config)(nil),
					nebula_go.NoLogger{}).Return(sessGetter, nil)
			},
			verify: func(t *testing.T, sessPool *nebula_go.SessionPool, session nebula_go.NebulaSession) {
				got, err := sessPool.Acquire()
				assert.Nil(t, err)
				assert.Equal(t, session, got)

				sessPool.Release(got)
			},
		},
		{
			label: "should return a session pool and return an authenticated with max session pool, release on Close",
			opts: []nebula_go.ConnectionOption{
				nebula_go.WithCredentials("foo", "bar"),
			},
			connString: "nebula://user:pass@localhost?max_idle_session_pool_size=10",
			prepare: func(connPollBuilder *mockConnectionPoolBuilder,
				sessGetter *mockSessionGetter,
				session *mockSession,
			) {
				sessGetter.On("GetSession", "foo", "bar").Return(session, nil)

				connPollBuilder.On("CALL",
					[]nebula_go.HostAddress{
						{Host: "localhost", Port: 9669},
					},
					nebula_go.GetDefaultConf(),
					(*tls.Config)(nil),
					nebula_go.NoLogger{}).Return(sessGetter, nil)

				session.On("Release").Return()

				sessGetter.On("Close").Return()
			},
			verify: func(t *testing.T, sessPool *nebula_go.SessionPool, session nebula_go.NebulaSession) {
				got, err := sessPool.Acquire()
				assert.Nil(t, err)
				assert.Equal(t, session, got)

				sessPool.Release(got)

				sessPool.Close()
			},
		},
		{
			label:      "should return a session pool and return an authenticated session via WithSession",
			connString: "nebula://user:pass@localhost",
			opts: []nebula_go.ConnectionOption{
				nebula_go.WithTLSConfig(&tls.Config{}),
				nebula_go.WithDefaultLogger(),
			},
			prepare: func(connPollBuilder *mockConnectionPoolBuilder,
				sessGetter *mockSessionGetter,
				session *mockSession,
			) {
				sessGetter.On("GetSession", "user", "pass").Return(session, nil)

				connPollBuilder.On("CALL",
					[]nebula_go.HostAddress{
						{Host: "localhost", Port: 9669},
					},
					nebula_go.GetDefaultConf(),
					&tls.Config{},
					nebula_go.DefaultLogger{}).Return(sessGetter, nil)

				session.On("Release").Return()
			},
			verify: func(t *testing.T, sessPool *nebula_go.SessionPool, session nebula_go.NebulaSession) {
				err := sessPool.WithSession(func(got nebula_go.NebulaSession) error {
					assert.Equal(t, session, got)
					return nil
				})

				assert.Nil(t, err)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()

			var (
				connPollBuilder mockConnectionPoolBuilder
				sessGetter      mockSessionGetter
				session         mockSession
			)

			if tc.prepare != nil {
				tc.prepare(&connPollBuilder, &sessGetter, &session)
			}

			opts := []nebula_go.ConnectionOption{
				nebula_go.WithConnectionPoolBuilder(connPollBuilder.CALL),
			}

			opts = append(opts, tc.opts...)

			sp, err := nebula_go.NewSessionPool(tc.connString, opts...)

			if tc.errMsg != "" {
				assert.Nil(t, sp)
				assert.EqualError(t, err, tc.errMsg)
			} else {
				assert.Nil(t, err)
				tc.verify(t, sp, &session)
			}

			connPollBuilder.AssertExpectations(t)
			sessGetter.AssertExpectations(t)
			session.AssertExpectations(t)
		})
	}
}

var (
	_ nebula_go.ConnectionPoolBuilder = (*mockConnectionPoolBuilder)(nil).CALL

	_ nebula_go.SessionGetter = (*mockSessionGetter)(nil)

	_ nebula_go.NebulaSession = (*mockSession)(nil)
)

type mockConnectionPoolBuilder struct {
	mock.Mock
}

func (m *mockConnectionPoolBuilder) CALL(addresses []nebula_go.HostAddress,
	poolConfig nebula_go.PoolConfig,
	tlsConfig *tls.Config,
	log nebula_go.Logger,
) (nebula_go.SessionGetter, error) {
	args := m.Called(addresses, poolConfig, tlsConfig, log)

	connPool, _ := args.Get(0).(nebula_go.SessionGetter)

	return connPool, args.Error(1)
}

type mockSessionGetter struct {
	mock.Mock
}

func (m *mockSessionGetter) Close() {
	m.Called()
}

func (m *mockSessionGetter) GetSession(user, pass string) (nebula_go.NebulaSessionReleaser, error) {
	args := m.Called(user, pass)

	session, _ := args.Get(0).(nebula_go.NebulaSessionReleaser)

	return session, args.Error(1)
}

type mockSession struct {
	mock.Mock
}

func (m *mockSession) Execute(stmt string) (*nebula_go.ResultSet, error) {
	args := m.Called(stmt)

	resultSet, _ := args.Get(0).(*nebula_go.ResultSet)

	return resultSet, args.Error(1)
}

func (m *mockSession) ExecuteWithParameter(stmt string, params map[string]interface{}) (*nebula_go.ResultSet, error) {
	args := m.Called(stmt, params)

	resultSet, _ := args.Get(0).(*nebula_go.ResultSet)

	return resultSet, args.Error(1)
}

func (m *mockSession) ExecuteJson(stmt string) ([]byte, error) {
	args := m.Called(stmt)

	result, _ := args.Get(0).([]byte)

	return result, args.Error(1)
}

func (m *mockSession) ExecuteJsonWithParameter(stmt string, params map[string]interface{}) ([]byte, error) {
	args := m.Called(stmt, params)

	result, _ := args.Get(0).([]byte)

	return result, args.Error(1)
}

func (m *mockSession) GetSessionID() int64 {
	return 1
}

func (m *mockSession) Release() {
	m.Called()
}
