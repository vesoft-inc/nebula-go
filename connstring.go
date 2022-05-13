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
	"regexp"
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

	defaultOnAcquireSession = `USE %SPACE%;`
)

// ConnectionConfig type.
type ConnectionConfig struct {
	// HostAddresses defines a list of host (string) and port (number)
	HostAddresses []HostAddress
	PoolConfig
	SessionPoolConfig
	Username  string
	Password  string
	Space     string
	TLS       string
	TLSConfig *tls.Config
	Log       Logger

	OnAcquireSession string
	OnReleaseSession string

	ConnectionPoolBuilder
}

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

// WithDefaultLogger functional option to substitute the NoLogger by DefaultLogger.
func WithDefaultLogger() ConnectionOption {
	return WithLogger(DefaultLogger{})
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

// WithSessionPoolConfig functional option to override the session pool configuration.
func WithSessionPoolConfig(sessionPoolConfig SessionPoolConfig) ConnectionOption {
	return func(cfg *ConnectionConfig) {
		cfg.SessionPoolConfig = sessionPoolConfig
	}
}

// WithOnAcquireSessionStmt functional option to override the default on acquire session stmt.
// This will be executed each time one session is acquired from the pool
// The default value if no Space is defined is none.
// Else, the default value is:
//    USE %SPACE%;
// where macro %SPACE% being substituted by the value of Space
func WithOnAcquireSessionStmt(stmt string) ConnectionOption {
	return func(cfg *ConnectionConfig) {
		cfg.OnAcquireSession = stmt
	}
}

// WithOnReleaseSessionStmt functional option to override the default on release session stmt.
// This will be executed each time one session is released to the pool
// Default is none.
// Use the same format as OnAcquireSession with macro %SPACE% being substituted
// by the value of of Space
func WithOnReleaseSessionStmt(stmt string) ConnectionOption {
	return func(cfg *ConnectionConfig) {
		cfg.OnReleaseSession = stmt
	}
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
//   "nebula://user:pass@hostname:port/space"              if defined, we run "USE <space>;" before each session acquire
//   "nebula://user:pass@hostname:port?TimeOut=2s"         set the pool conf timeout as 2s (default 0s)
//   "nebula://user:pass@hostname:port?IdleOut=2s"         set the pool conf idleout as 2s (default 0s)
//   "nebula://user:pass@hostname:port?MaxConnPoolSize=15" set max conn poll size to 15    (default 10)
//   "nebula://user:pass@hostname:port?MinConnPoolSize=4"  set min conn poll size to 4     (default  0)
//   "nebula://user:pass@hostname:port?tls=false"          use no TLS
//   "nebula://user:pass@hostname:port?tls=true"           use TLS &tls.Config{}
//   "nebula://user:pass@hostname:port?tls=skip-verify"    use TLS with InsecureSkipVerify true
//   "nebula://user:pass@hostname:port?tls=custom"         use config registered via RegisterTLSConfig
//   "nebula://user:pass@[host1,host2,...hostN]"           define multiple hosts
//   "nebula://hostname?MaxIdleSessionPoolSize=10"         set max idle session pool to 10 (default 0)
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

	sessionPoolConfig := GetDefaultSessionPoolConfig()

	err = peekIntFromQueryString(query, "MaxIdleSessionPoolSize", &sessionPoolConfig.MaxIdleSessionPoolSize)
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
		HostAddresses:     make([]HostAddress, len(hostPorts)),
		PoolConfig:        poolConfig,
		SessionPoolConfig: sessionPoolConfig,
		Username:          connectionURL.User.Username(),
	}

	if password, ok := connectionURL.User.Password(); ok {
		conf.Password = password
	}

	if space := strings.Replace(connectionURL.Path, "/", "", 1); space != "" {
		if !nebulaGraphSpaceNameFormat.MatchString(space) {
			return nil, fmt.Errorf("space name %q is not valid", space)
		}

		conf.Space = space
		conf.OnAcquireSession = defaultOnAcquireSession
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

		conf.TLSConfig, err = getTLSConfig(tlsOption)
		if err != nil {
			return nil, err
		}
	}

	return conf, nil
}

var nebulaGraphSpaceNameFormat = regexp.MustCompile("^[a-zA-Z0-9_]*$")

// Validate check the internal configuration consistency.
func (cfg *ConnectionConfig) Validate() error {
	cfg.SessionPoolConfig.validateConf(cfg.Log)

	if !nebulaGraphSpaceNameFormat.MatchString(cfg.Space) {
		return fmt.Errorf("space name %q is not valid", cfg.Space)
	}

	return nil
}

// String return a string representation of this configuration.
func (cfg *ConnectionConfig) String() string {
	uri := cfg.toURI()

	return uri.String()
}

// Redacted return a redacted string representation of this configuration to save password.
func (cfg *ConnectionConfig) Redacted() string {
	uri := cfg.toURI()

	if _, ok := uri.User.Password(); ok {
		uri.User = url.UserPassword(cfg.Username, "xxxxx")
	}

	return uri.String()
}

func (cfg *ConnectionConfig) toURI() *url.URL {
	var userinfo *url.Userinfo
	if cfg.Username != "" {
		if cfg.Password != "" {
			userinfo = url.UserPassword(cfg.Username, cfg.Password)
		} else {
			userinfo = url.User(cfg.Username)
		}
	}

	var (
		hostPort string
		path     string
	)

	if n := len(cfg.HostAddresses); n > 0 {
		hosts := make([]string, n)
		ports := make([]int, n)
		var hasDifferentPorts bool

		firstPort := cfg.HostAddresses[0].Port

		for i, hp := range cfg.HostAddresses {
			hosts[i], ports[i] = hp.Host, hp.Port

			if hp.Port != firstPort {
				hasDifferentPorts = true
			}
		}

		if hasDifferentPorts {
			for i, host := range hosts {
				hosts[i] = net.JoinHostPort(host, strconv.Itoa(ports[i]))
			}

			hostPort = fmt.Sprintf("[%s]", strings.Join(hosts, ","))
		} else if n > 1 {
			hostPort = fmt.Sprintf("[%s]:%d", strings.Join(hosts, ","), firstPort)
		} else {
			hostPort = net.JoinHostPort(hosts[0], strconv.Itoa(ports[0]))
		}
	}

	query := url.Values{}

	defaultConf := GetDefaultConf()
	if cfg.PoolConfig.TimeOut != defaultConf.TimeOut {
		query.Add("TimeOut", cfg.PoolConfig.TimeOut.String())
	}
	if cfg.PoolConfig.IdleTime != defaultConf.IdleTime {
		query.Add("IdleTime", cfg.PoolConfig.IdleTime.String())
	}
	if cfg.PoolConfig.MaxConnPoolSize != defaultConf.MaxConnPoolSize {
		query.Add("MaxConnPoolSize", strconv.Itoa(cfg.PoolConfig.MaxConnPoolSize))
	}
	if cfg.PoolConfig.MinConnPoolSize != defaultConf.MinConnPoolSize {
		query.Add("MinConnPoolSize", strconv.Itoa(cfg.PoolConfig.MinConnPoolSize))
	}

	defaultSessConf := GetDefaultSessionPoolConfig()
	if cfg.SessionPoolConfig.MaxIdleSessionPoolSize != defaultSessConf.MaxIdleSessionPoolSize {
		query.Add("MaxIdleSessionPoolSize", strconv.Itoa(cfg.SessionPoolConfig.MaxIdleSessionPoolSize))
	}

	if cfg.TLS != "" {
		query.Add("tls", cfg.TLS)
	}

	if cfg.Space != "" {
		path = "/" + cfg.Space
	}

	uri := &url.URL{
		Scheme:   "nebula",
		User:     userinfo,
		Host:     hostPort,
		RawQuery: query.Encode(),
		Path:     path,
	}

	return uri
}

// Apply method.
func (cfg *ConnectionConfig) Apply(opts []ConnectionOption) {
	for _, opt := range opts {
		opt(cfg)
	}
}

// BuildConnectionPool return an interface SessionGetter of ConnectionPool
// based on the configuration / connection string.
func (cfg *ConnectionConfig) BuildConnectionPool() (SessionGetter, error) {
	if cfg.TLS != "" && cfg.TLSConfig == nil {
		tlsConfig, err := getTLSConfig(cfg.TLS)
		if err != nil {
			return nil, err
		}

		cfg.TLSConfig = tlsConfig
	}
	if cfg.Log == nil {
		cfg.Log = NoLogger{}
	}

	if cfg.ConnectionPoolBuilder == nil {
		cfg.ConnectionPoolBuilder = defaultConnectionPoolBuilder
	}

	return cfg.ConnectionPoolBuilder(cfg.HostAddresses, cfg.PoolConfig, cfg.TLSConfig, cfg.Log)
}

func defaultConnectionPoolBuilder(addresses []HostAddress,
	conf PoolConfig,
	sslConfig *tls.Config,
	log Logger,
) (SessionGetter, error) {
	connPool, err := NewSslConnectionPool(addresses, conf, sslConfig, log)
	if err != nil {
		return nil, err
	}

	return &connectionPoolWrapper{
		ConnectionPool: connPool,
	}, nil
}

func getTLSConfig(key string) (*tls.Config, error) {
	switch key {
	case "false", "0":
		return nil, nil
	case "true", "1":
		return &tls.Config{}, nil
	case "skip-verify":
		return &tls.Config{InsecureSkipVerify: true}, nil
	default:
		tlsConfig, err := getTLSConfigFromRegistry(key)
		if err != nil {
			return nil, err
		}

		return tlsConfig.Clone(), nil
	}
}
func getTLSConfigFromRegistry(key string) (*tls.Config, error) {
	tlsConfigLock.RLock()
	defer tlsConfigLock.RUnlock()

	if tlsConfig, ok := tlsConfigRegistry[key]; ok {
		return tlsConfig.Clone(), nil
	}

	return nil, fmt.Errorf("tls configuration %q not found", key)
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
