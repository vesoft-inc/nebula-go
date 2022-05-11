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

		prepare func(*mockConnectionPoolBuilder, *mockSessionGetter, *nebula_go.Session)
		verify  func(*testing.T, *nebula_go.SessionPool, *nebula_go.Session)

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
				_ *nebula_go.Session) {

				connPollBuilder.On("CALL",
					[]nebula_go.HostAddress{
						{Host: "localhost", Port: 9669},
					},
					nebula_go.GetDefaultConf(),
					(*tls.Config)(nil),
					nebula_go.DefaultLogger{},
				).Return(nil, errors.New("ops"))
			},
			errMsg: "unable to build connection pool: ops",
		},
		{
			label:      "should return error from connection pool constructor with tls skip-verify",
			connString: "nebula://localhost?tls=skip-verify",
			opts: []nebula_go.ConnectionOption{
				nebula_go.WithLogger(nil),
			},
			prepare: func(connPollBuilder *mockConnectionPoolBuilder,
				_ *mockSessionGetter,
				_ *nebula_go.Session) {

				connPollBuilder.On("CALL",
					[]nebula_go.HostAddress{
						{Host: "localhost", Port: 9669},
					},
					nebula_go.GetDefaultConf(),
					&tls.Config{InsecureSkipVerify: true},
					nil,
				).Return(nil, errors.New("ops"))
			},
			errMsg: "unable to build connection pool: ops",
		},
		{
			label:      "should return a session pool and close should call connection pool close",
			connString: "nebula://localhost?TimeOut=5s",
			prepare: func(connPollBuilder *mockConnectionPoolBuilder,
				sessGetter *mockSessionGetter,
				_ *nebula_go.Session) {

				sessGetter.On("Close").Return()

				poolConf := nebula_go.GetDefaultConf()
				poolConf.TimeOut = 5 * time.Second

				connPollBuilder.On("CALL",
					[]nebula_go.HostAddress{
						{Host: "localhost", Port: 9669},
					},
					poolConf,
					(*tls.Config)(nil),
					nebula_go.DefaultLogger{},
				).Return(sessGetter, nil)
			},
			verify: func(t *testing.T, sessPool *nebula_go.SessionPool, _ *nebula_go.Session) {
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
				_ *nebula_go.Session) {

				sessGetter.On("GetSession", "user", "pass").Return(nil, errors.New("ops"))

				connPollBuilder.On("CALL",
					[]nebula_go.HostAddress{
						{Host: "localhost", Port: 9669},
					},
					nebula_go.PoolConfig{
						IdleTime: 5 * time.Second,
					},
					(*tls.Config)(nil),
					nebula_go.DefaultLogger{},
				).Return(sessGetter, nil)
			},
			verify: func(t *testing.T, sessPool *nebula_go.SessionPool, _ *nebula_go.Session) {
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
				session *nebula_go.Session) {

				sessGetter.On("GetSession", "foo", "bar").Return(session, nil)

				connPollBuilder.On("CALL",
					[]nebula_go.HostAddress{
						{Host: "localhost", Port: 9669},
					},
					nebula_go.GetDefaultConf(),
					(*tls.Config)(nil),
					nebula_go.DefaultLogger{},
				).Return(sessGetter, nil)
			},
			verify: func(t *testing.T, sessPool *nebula_go.SessionPool, session *nebula_go.Session) {
				got, err := sessPool.Acquire()
				assert.Nil(t, err)
				assert.Equal(t, session, got)
			},
		},
		{
			label:      "should return a session pool and return an authenticated session via WithSession",
			connString: "nebula://user:pass@localhost",
			opts: []nebula_go.ConnectionOption{
				nebula_go.WithTLSConfig(&tls.Config{}),
			},
			prepare: func(connPollBuilder *mockConnectionPoolBuilder,
				sessGetter *mockSessionGetter,
				session *nebula_go.Session) {

				sessGetter.On("GetSession", "user", "pass").Return(session, nil)

				connPollBuilder.On("CALL",
					[]nebula_go.HostAddress{
						{Host: "localhost", Port: 9669},
					},
					nebula_go.GetDefaultConf(),
					&tls.Config{},
					nebula_go.DefaultLogger{},
				).Return(sessGetter, nil)
			},
			verify: func(t *testing.T, sessPool *nebula_go.SessionPool, session *nebula_go.Session) {
				err := sessPool.WithSession(func(got *nebula_go.Session) error {
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
				session         nebula_go.Session
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
		})
	}
}

var (
	_ nebula_go.ConnectionPoolBuilder = (*mockConnectionPoolBuilder)(nil).CALL

	_ nebula_go.SessionGetter = (*mockSessionGetter)(nil)
)

type mockConnectionPoolBuilder struct {
	mock.Mock
}

func (m *mockConnectionPoolBuilder) CALL(addresses []nebula_go.HostAddress,
	poolConfig nebula_go.PoolConfig,
	tlsConfig *tls.Config,
	log nebula_go.Logger) (nebula_go.SessionGetter, error) {

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

func (m *mockSessionGetter) GetSession(user, pass string) (*nebula_go.Session, error) {
	args := m.Called(user, pass)

	session, _ := args.Get(0).(*nebula_go.Session)

	return session, args.Error(1)
}
