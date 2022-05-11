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
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// DEFAULT_PORT is the default nebula db port.
	DEFAULT_PORT = 9669

	// NEBULA_SCHEME is the expected scheme / protocol in connection strings
	NEBULA_SCHEME = "nebula"
)

// ConnectionConfig type.
type ConnectionConfig struct {
	// HostAddresses defines a list of host (string) and port (number)
	HostAddresses []HostAddress
	PoolConfig    PoolConfig
	Username      string
	Password      string
	Space         string
	TLS           string
	TLSConfig     *tls.Config
	Log           Logger

	ConnectionPoolBuilder
}

// ConnectionPoolBuilder type.
type ConnectionPoolBuilder func([]HostAddress, PoolConfig, *tls.Config, Logger) (SessionGetter, error)

var (
	tlsConfigLock     sync.RWMutex
	tlsConfigRegistry map[string]*tls.Config
)

// ParseConnectionString builder function.
// This function parses a uri-like nebula graph connection string
// Examples:
//   "hostname"                                            represents a connection to host "hostname" using default port 9669
//   "hostname:port"                                       a connection to host "hostname" using port "port"
//   "nebula://hostname:port"                              same but explicit use protocol nebula://
//   "nebula://user:pass@hostname:port"                    define user and password to use in sessions
//   "nebula://user:pass@hostname:port/space"              reserved for future use
//   "nebula://user:pass@hostname:port?TimeOut=2s"         set the pool conf timeout as 2s
//   "nebula://user:pass@hostname:port?IdleOut=2s"         set the pool conf idleout as 2s
//   "nebula://user:pass@hostname:port?MaxConnPoolSize=10" set max conn poll size to 10
//   "nebula://user:pass@hostname:port?MinConnPoolSize=0"  set min conn poll size to 0
//   "nebula://user:pass@hostname:port?tls=false"          use no TLS
//   "nebula://user:pass@hostname:port?tls=true"           use TLS &tls.Config{}
//   "nebula://user:pass@hostname:port?tls=skip-verify"    use TLS with InsecureSkipVerify true
//   "nebula://user:pass@hostname:port?tls=custom"         use config registered via RegisterTLSConfig
//   "nebula://user:pass@[host1,host2,...hostN]"           define multiple hosts
func ParseConnectionString(connectionString string) (*ConnectionConfig, error) {
	return parseConnectionString(connectionString, true)
}

func parseConnectionString(connectionString string, canRetry bool) (*ConnectionConfig, error) {
	const protocolSeparator = "://"

	if canRetry && !strings.Contains(connectionString, protocolSeparator) {
		return parseConnectionString(NEBULA_SCHEME+protocolSeparator+connectionString, false)
	}

	connectionURL, err := url.Parse(connectionString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string %q as url: %v", connectionString, err)
	}

	if connectionURL.Scheme != NEBULA_SCHEME {
		return nil, fmt.Errorf("connection string must start with %q:// instead %q",
			NEBULA_SCHEME, connectionURL.Scheme)
	}

	query := connectionURL.Query()

	poolConfig := GetDefaultConf()
	err = peekDurationFromQueryString(query, "TimeOut", &poolConfig.TimeOut)
	if err != nil {
		return nil, err
	}

	err = peekDurationFromQueryString(query, "IdleTime", &poolConfig.IdleTime)
	if err != nil {
		return nil, err
	}

	err = peekIntFromQueryString(query, "MaxConnPoolSize", &poolConfig.MaxConnPoolSize)
	if err != nil {
		return nil, err
	}

	err = peekIntFromQueryString(query, "MinConnPoolSize", &poolConfig.MinConnPoolSize)
	if err != nil {
		return nil, err
	}

	defaultPort := DEFAULT_PORT
	if defaultPortOrService := connectionURL.Port(); defaultPortOrService != "" {
		defaultPort, err = convertToTCPPort(defaultPortOrService)
		if err != nil {
			return nil, err
		}
	}

	hostname := connectionURL.Host

	hostPorts := []string{hostname}

	if strings.ContainsRune(hostname, ',') {
		hostPorts = strings.Split(connectionURL.Hostname(), ",")
	}

	conf := &ConnectionConfig{
		HostAddresses: make([]HostAddress, len(hostPorts)),
		PoolConfig:    poolConfig,
		Username:      connectionURL.User.Username(),
	}

	if password, ok := connectionURL.User.Password(); ok {
		conf.Password = password
	}

	if space := strings.Replace(connectionURL.Path, "/", "", 1); space != "" {
		conf.Space = space
	}

	for i, hostPort := range hostPorts {
		if hostPort == "" {
			return nil, errors.New("unexpected empty host/port")
		}

		var portOrService string

		conf.HostAddresses[i].Port = defaultPort

		if stripIPv6Brackets, hasPort := checkTCPPort(hostPort); !hasPort {
			conf.HostAddresses[i].Host = stripIPv6Brackets

			continue
		}

		conf.HostAddresses[i].Host, portOrService, err = net.SplitHostPort(hostPort)
		if err != nil {
			return nil, fmt.Errorf("unable to parse host port %q: %v", hostPort, err)
		}

		if portOrService == "" {
			continue
		}

		conf.HostAddresses[i].Port, err = convertToTCPPort(portOrService)
		if err != nil {
			return nil, err
		}
	}

	if tlsOption := query.Get("tls"); tlsOption != "" {
		conf.TLS = tlsOption

		switch tlsOption {
		case "false", "0":
		case "true", "1":
			conf.TLSConfig = &tls.Config{}
		case "skip-verify":
			conf.TLSConfig = &tls.Config{InsecureSkipVerify: true}
		default:
			if tlsConfig, ok := getTLSConfig(tlsOption); ok {
				conf.TLSConfig = tlsConfig.Clone()
			} else {
				return nil, fmt.Errorf("tls configuration %q not found", tlsOption)
			}
		}
	}

	return conf, nil
}

func getTLSConfig(key string) (*tls.Config, bool) {
	tlsConfigLock.RLock()
	defer tlsConfigLock.RUnlock()

	if tlsConfig, ok := tlsConfigRegistry[key]; ok {
		return tlsConfig.Clone(), true
	}

	return nil, false
}

// RegisterTLSConfig adds the tls.Config associated with key.
func RegisterTLSConfig(key string, config *tls.Config) error {
	switch key {
	case "":
		return errors.New("missing key")
	case "true", "false", "0", "1", "skip-verify":
		return fmt.Errorf("key '%s' is reserved", key)
	}

	tlsConfigLock.Lock()

	defer tlsConfigLock.Unlock()

	if tlsConfigRegistry == nil {
		tlsConfigRegistry = make(map[string]*tls.Config)
	}

	tlsConfigRegistry[key] = config.Clone()

	return nil
}

// DeregisterTLSConfig removes the tls.Config associated with key.
func DeregisterTLSConfig(key string) {
	tlsConfigLock.Lock()

	defer tlsConfigLock.Unlock()

	if tlsConfigRegistry != nil {
		delete(tlsConfigRegistry, key)
	}
}

func checkTCPPort(hostPort string) (stripIPv6Brackets string, hasPort bool) {
	// check if ipv6
	stripIPv6Brackets = hostPort
	if pos := strings.IndexByte(hostPort, ']'); pos > 1 {
		stripIPv6Brackets = hostPort[1:pos]
		hostPort = hostPort[pos:]
	}

	hasPort = strings.IndexByte(hostPort, ':') != -1

	return
}

func convertToTCPPort(portOrService string) (int, error) {
	port, err := strconv.Atoi(portOrService)
	if err != nil {
		port, err = net.LookupPort("tcp", portOrService)
		if err != nil {
			return 0, fmt.Errorf("unable to parse service %q as port: %v", portOrService, err)
		}
	}

	return port, nil
}

func peekDurationFromQueryString(query url.Values, key string, dest *time.Duration) (err error) {
	if duration := query.Get(key); duration != "" {
		*dest, err = time.ParseDuration(duration)
		if err != nil {
			err = fmt.Errorf("unable to parse query string '%s' %q as duration: %v", key, duration, err)
		}
	}

	return
}

func peekIntFromQueryString(query url.Values, key string, dest *int) (err error) {
	if duration := query.Get(key); duration != "" {
		*dest, err = strconv.Atoi(duration)
		if err != nil {
			err = fmt.Errorf("unable to parse query string '%s' %q as int: %v", key, duration, err)
		}
	}

	return
}
