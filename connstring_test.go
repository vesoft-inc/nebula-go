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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	nebula_go "github.com/vesoft-inc/nebula-go/v3"
)

func TestParseString(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		label      string
		connString string
		cfg        *nebula_go.ConnectionConfig
		errMsg     string
	}{
		{
			label:      "empty connection string should fail",
			connString: "",
			errMsg:     "unexpected empty host/port",
		},
		{
			label:      "simple connection string should not fail",
			connString: "nebula://localhost",
			cfg: &nebula_go.ConnectionConfig{
				HostAddresses: []nebula_go.HostAddress{
					{
						Host: "localhost",
						Port: nebula_go.DEFAULT_PORT,
					},
				},
				PoolConfig: nebula_go.GetDefaultConf(),
			},
		},
		{
			label:      "can omit protocol nebula:// to define a host and port",
			connString: "localhost:1234",
			cfg: &nebula_go.ConnectionConfig{
				HostAddresses: []nebula_go.HostAddress{
					{
						Host: "localhost",
						Port: 1234,
					},
				},
				PoolConfig: nebula_go.GetDefaultConf(),
			},
		},
		{
			label:      "can omit protocol nebula:// to define a host without port",
			connString: "localhost",
			cfg: &nebula_go.ConnectionConfig{
				HostAddresses: []nebula_go.HostAddress{
					{
						Host: "localhost",
						Port: nebula_go.DEFAULT_PORT,
					},
				},
				PoolConfig: nebula_go.GetDefaultConf(),
			},
		},
		{
			label:      "must use protocol nebula:// to define a host",
			connString: "other://localhost",
			errMsg:     "connection string must start with \"nebula\":// instead \"other\"",
		},
		{
			label:      "empty connection string should fail",
			connString: "nebula://",
			errMsg:     "unexpected empty host/port",
		},
		{
			label:      "simple connection string with custom port",
			connString: "nebula://127.0.0.1:1234",
			cfg: &nebula_go.ConnectionConfig{
				HostAddresses: []nebula_go.HostAddress{
					{
						Host: "127.0.0.1",
						Port: 1234,
					},
				},
				PoolConfig: nebula_go.GetDefaultConf(),
			},
		},
		{
			label:      "simple connection string with custom port as ipv6",
			connString: "nebula://[fec0:bebe:cafe::01]:1234",
			cfg: &nebula_go.ConnectionConfig{
				HostAddresses: []nebula_go.HostAddress{
					{
						Host: "fec0:bebe:cafe::01",
						Port: 1234,
					},
				},
				PoolConfig: nebula_go.GetDefaultConf(),
			},
		},
		{
			label:      "simple connection string with query string",
			connString: "nebula://localhost?TimeOut=2s&IdleTime=1s&MaxConnPoolSize=5&MinConnPoolSize=2",
			cfg: &nebula_go.ConnectionConfig{
				HostAddresses: []nebula_go.HostAddress{
					{
						Host: "localhost",
						Port: nebula_go.DEFAULT_PORT,
					},
				},
				PoolConfig: nebula_go.PoolConfig{
					TimeOut:         2 * time.Second,
					IdleTime:        1 * time.Second,
					MaxConnPoolSize: 5,
					MinConnPoolSize: 2,
				},
			},
		},
		{
			label:      "simple connection string with user/pass and tls false",
			connString: "nebula://user:pass@localhost/myspace?tls=false",
			cfg: &nebula_go.ConnectionConfig{
				HostAddresses: []nebula_go.HostAddress{
					{
						Host: "localhost",
						Port: nebula_go.DEFAULT_PORT,
					},
				},
				Username:   "user",
				Password:   "pass",
				Space:      "myspace",
				TLS:        "false",
				PoolConfig: nebula_go.GetDefaultConf(),
			},
		},
		{
			label:      "simple connection string tls true",
			connString: "nebula://localhost?tls=true",
			cfg: &nebula_go.ConnectionConfig{
				HostAddresses: []nebula_go.HostAddress{
					{
						Host: "localhost",
						Port: nebula_go.DEFAULT_PORT,
					},
				},
				TLS:        "true",
				TLSConfig:  &tls.Config{},
				PoolConfig: nebula_go.GetDefaultConf(),
			},
		},
		{
			label:      "simple connection string tls skip-verify",
			connString: "nebula://localhost?tls=skip-verify",
			cfg: &nebula_go.ConnectionConfig{
				HostAddresses: []nebula_go.HostAddress{
					{
						Host: "localhost",
						Port: nebula_go.DEFAULT_PORT,
					},
				},
				TLS:        "skip-verify",
				TLSConfig:  &tls.Config{InsecureSkipVerify: true},
				PoolConfig: nebula_go.GetDefaultConf(),
			},
		},
		{
			label:      "simple connection string tls not-found should fail",
			connString: "nebula://localhost?tls=not-found",
			errMsg:     "tls configuration \"not-found\" not found",
		},
		{
			label:      "should support multiple ips and hostnames",
			connString: "nebula://[[fec0:bebe:cafe::01]:1234,[::1],1.1.1.1,2.2.2.2:9999,other]:1234?tls=false",
			cfg: &nebula_go.ConnectionConfig{
				HostAddresses: []nebula_go.HostAddress{
					{
						Host: "fec0:bebe:cafe::01",
						Port: 1234,
					},
					{
						Host: "::1",
						Port: 1234,
					},
					{
						Host: "1.1.1.1",
						Port: 1234,
					},
					{
						Host: "2.2.2.2",
						Port: 9999,
					},
					{
						Host: "other",
						Port: 1234,
					},
				},
				TLS:        "false",
				PoolConfig: nebula_go.GetDefaultConf(),
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()

			cfg, err := nebula_go.ParseConnectionString(tc.connString)
			if tc.errMsg != "" {
				assert.Nil(t, cfg, "connection configuration must be nil")
				assert.EqualError(t, err, tc.errMsg, "expected error string")
			} else {
				assert.Equal(t, tc.cfg, cfg, "must return the expected configuration object")
				assert.NoError(t, err, "must return no error")
			}
		})
	}
}

func TestConnStringConversion(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		label    string
		cfg      nebula_go.ConnectionConfig
		str      string
		redacted string
	}{
		{
			label: "basic configuration",
			cfg: nebula_go.ConnectionConfig{
				HostAddresses: []nebula_go.HostAddress{
					{
						Host: "localhost",
						Port: 9669,
					},
				},
				PoolConfig: nebula_go.GetDefaultConf(),
			},
			str:      "nebula://localhost:9669",
			redacted: "nebula://localhost:9669",
		},
		{
			label: "basic ipv4 configuration",
			cfg: nebula_go.ConnectionConfig{
				HostAddresses: []nebula_go.HostAddress{
					{
						Host: "127.0.0.1",
						Port: 9669,
					},
				},
				PoolConfig: nebula_go.GetDefaultConf(),
			},
			str:      "nebula://127.0.0.1:9669",
			redacted: "nebula://127.0.0.1:9669",
		},
		{
			label: "basic ipv6 configuration",
			cfg: nebula_go.ConnectionConfig{
				HostAddresses: []nebula_go.HostAddress{
					{
						Host: "fec0:bebe:cafe::01",
						Port: 9669,
					},
				},
				PoolConfig: nebula_go.GetDefaultConf(),
			},
			str:      "nebula://[fec0:bebe:cafe::01]:9669",
			redacted: "nebula://[fec0:bebe:cafe::01]:9669",
		},
		{
			label: "tls configuration",
			cfg: nebula_go.ConnectionConfig{
				HostAddresses: []nebula_go.HostAddress{
					{
						Host: "localhost",
						Port: 9669,
					},
				},
				PoolConfig: nebula_go.GetDefaultConf(),
				TLS:        "skip-verify",
			},
			str:      "nebula://localhost:9669?tls=skip-verify",
			redacted: "nebula://localhost:9669?tls=skip-verify",
		},
		{
			label: "space configuration",
			cfg: nebula_go.ConnectionConfig{
				HostAddresses: []nebula_go.HostAddress{
					{
						Host: "localhost",
						Port: 9669,
					},
				},
				PoolConfig: nebula_go.GetDefaultConf(),
				Space:      "test",
			},
			str:      "nebula://localhost:9669/test",
			redacted: "nebula://localhost:9669/test",
		},
		{
			label: "credentials no password configuration",
			cfg: nebula_go.ConnectionConfig{
				HostAddresses: []nebula_go.HostAddress{
					{
						Host: "localhost",
						Port: 9669,
					},
				},
				Username:   "foo",
				PoolConfig: nebula_go.GetDefaultConf(),
			},
			str:      "nebula://foo@localhost:9669",
			redacted: "nebula://foo@localhost:9669",
		},
		{
			label: "credentials with password configuration",
			cfg: nebula_go.ConnectionConfig{
				HostAddresses: []nebula_go.HostAddress{
					{
						Host: "localhost",
						Port: 9669,
					},
				},
				Username:   "foo",
				Password:   "bar",
				PoolConfig: nebula_go.GetDefaultConf(),
			},
			str:      "nebula://foo:bar@localhost:9669",
			redacted: "nebula://foo:xxxxx@localhost:9669",
		},
		{
			label: "pool config configuration MaxConnPoolSize",
			cfg: nebula_go.ConnectionConfig{
				HostAddresses: []nebula_go.HostAddress{
					{
						Host: "localhost",
						Port: 9669,
					},
				},
				PoolConfig: nebula_go.PoolConfig{
					MaxConnPoolSize: 7,
				},
			},
			str:      "nebula://localhost:9669?MaxConnPoolSize=7",
			redacted: "nebula://localhost:9669?MaxConnPoolSize=7",
		},
		{
			label: "pool config configuration MinConnPoolSize",
			cfg: nebula_go.ConnectionConfig{
				HostAddresses: []nebula_go.HostAddress{
					{
						Host: "localhost",
						Port: 9669,
					},
				},
				PoolConfig: nebula_go.PoolConfig{
					MaxConnPoolSize: 10,
					MinConnPoolSize: 7,
				},
			},
			str:      "nebula://localhost:9669?MinConnPoolSize=7",
			redacted: "nebula://localhost:9669?MinConnPoolSize=7",
		},
		{
			label: "pool config configuration TimeOut",
			cfg: nebula_go.ConnectionConfig{
				HostAddresses: []nebula_go.HostAddress{
					{
						Host: "localhost",
						Port: 9669,
					},
				},
				PoolConfig: nebula_go.PoolConfig{
					MaxConnPoolSize: 10,
					TimeOut:         7 * time.Second,
				},
			},
			str:      "nebula://localhost:9669?TimeOut=7s",
			redacted: "nebula://localhost:9669?TimeOut=7s",
		},
		{
			label: "pool config configuration IdleTime",
			cfg: nebula_go.ConnectionConfig{
				HostAddresses: []nebula_go.HostAddress{
					{
						Host: "localhost",
						Port: 9669,
					},
				},
				PoolConfig: nebula_go.PoolConfig{
					MaxConnPoolSize: 10,
					IdleTime:        7 * time.Second,
				},
			},
			str:      "nebula://localhost:9669?IdleTime=7s",
			redacted: "nebula://localhost:9669?IdleTime=7s",
		},
		{
			label: "many hosts same port configuration",
			cfg: nebula_go.ConnectionConfig{
				HostAddresses: []nebula_go.HostAddress{
					{
						Host: "host1",
						Port: 9669,
					},
					{
						Host: "host2",
						Port: 9669,
					},
				},
				PoolConfig: nebula_go.GetDefaultConf(),
			},
			str:      "nebula://[host1,host2]:9669",
			redacted: "nebula://[host1,host2]:9669",
		},
		{
			label: "many hosts different ports configuration",
			cfg: nebula_go.ConnectionConfig{
				HostAddresses: []nebula_go.HostAddress{
					{
						Host: "host1",
						Port: 9669,
					},
					{
						Host: "host2",
						Port: 9665,
					},
				},
				PoolConfig: nebula_go.GetDefaultConf(),
			},
			str:      "nebula://[host1:9669,host2:9665]",
			redacted: "nebula://[host1:9669,host2:9665]",
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.str, tc.cfg.String())
			assert.Equal(t, tc.redacted, tc.cfg.Redacted())
		})
	}
}

func TestRegisterTLSConf(t *testing.T) {
	t.Parallel()

	tlsConfig := &tls.Config{
		ServerName: "test",
	}

	defer nebula_go.DeregisterTLSConfig("foo")

	err := nebula_go.RegisterTLSConfig("foo", tlsConfig)

	assert.NoError(t, err, "should register with success")

	connString := "nebula://user:pass@localhost/myspace?tls=foo"
	expected := &nebula_go.ConnectionConfig{
		HostAddresses: []nebula_go.HostAddress{
			{
				Host: "localhost",
				Port: nebula_go.DEFAULT_PORT,
			},
		},
		Username:   "user",
		Password:   "pass",
		Space:      "myspace",
		TLS:        "foo",
		TLSConfig:  tlsConfig.Clone(),
		PoolConfig: nebula_go.GetDefaultConf(),
	}

	actual, err := nebula_go.ParseConnectionString(connString)

	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestRegisterTLSConfReservedWords(t *testing.T) {
	t.Parallel()

	tlsConfig := &tls.Config{
		ServerName: "test",
	}

	keys := []string{"0", "1", "true", "false", "skip-verify"}
	for _, key := range keys {
		key := key
		t.Run("test reserved key "+key, func(t *testing.T) {
			t.Parallel()

			err := nebula_go.RegisterTLSConfig(key, tlsConfig)

			msg := fmt.Sprintf("key '%s' is reserved", key)
			assert.EqualError(t, err, msg)
		})
	}

	t.Run("empty key", func(t *testing.T) {
		t.Parallel()

		err := nebula_go.RegisterTLSConfig("", tlsConfig)

		assert.EqualError(t, err, "missing key")
	})
}
