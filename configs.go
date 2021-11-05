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
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// PoolConfig is the configs of connection pool
type PoolConfig struct {
	// Socket timeout and Socket connection timeout, unit: seconds
	TimeOut time.Duration
	// The idleTime of the connection, unit: seconds
	// If connection's idle time is longer than idleTime, it will be delete
	// 0 value means the connection will not expire
	IdleTime time.Duration
	// The max connections in pool for all addresses
	MaxConnPoolSize int
	// The min connections in pool for all addresses
	MinConnPoolSize int
}

// validateConf validates config
func (conf *PoolConfig) validateConf(log Logger) {
	if conf.TimeOut < 0 {
		conf.TimeOut = 0 * time.Millisecond
		log.Warn("Illegal Timeout value, the default value of 0 second has been applied")
	}
	if conf.IdleTime < 0 {
		conf.IdleTime = 0 * time.Millisecond
		log.Warn("Invalid IdleTime value, the default value of 0 second has been applied")
	}
	if conf.MaxConnPoolSize < 1 {
		conf.MaxConnPoolSize = 10
		log.Warn("Invalid MaxConnPoolSize value, the default value of 10 has been applied")
	}
	if conf.MinConnPoolSize < 0 {
		conf.MinConnPoolSize = 0
		log.Warn("Invalid MinConnPoolSize value, the default value of 0 has been applied")
	}
}

// SslConfig is a warpper of tls.Config which will be used to initialize SSL connection
// Notice that while users can use NewCaSignedSslConf() or NewSelfSignedSslConf() to generate SSL configs,
// it is also possible to manually a construct SslConfig by setting a customized SslConf using SetConfig()
type SslConfig struct {
	SslConf    tls.Config
	IsCaSigned bool
	Password   string
}

// NewCaSignedSslConf reads the given file path and generate SslConfig for CA-signed SSL connection
func NewCaSignedSslConf(rootCAPath string, certPath string, priKeyPath string) (*SslConfig, error) {
	rootCA, err := openAndReadFile(rootCAPath)
	if err != nil {
		return nil, err
	}
	cert, err := openAndReadFile(certPath)
	if err != nil {
		return nil, err
	}
	privateKey, err := openAndReadFile(priKeyPath)
	if err != nil {
		return nil, err
	}

	// Generate the client certificate
	clientCert, err := tls.X509KeyPair(cert, privateKey)
	if err != nil {
		return nil, err
	}

	// Parse root CA pem and add into CA pool
	rootCAPool := x509.NewCertPool()
	ok := rootCAPool.AppendCertsFromPEM(rootCA)
	if !ok {
		return nil, fmt.Errorf("unable to append supplied cert into tls.Config, are you sure it is a valid certificate")
	}

	return &SslConfig{
		SslConf: tls.Config{
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      rootCAPool,
		},
		IsCaSigned: true,
	}, nil
}

// NewSelfSignedSslConf reads the given file path and generate SslConfig for self-signed SSL connection
func NewSelfSignedSslConf(certPath string, priKeyPath string, passwordPath string) (*SslConfig, error) {
	cert, err := openAndReadFile(certPath)
	if err != nil {
		return nil, err
	}
	// For self-signed cert, use the local cert as the root ca
	rootCA := cert

	privateKey, err := openAndReadFile(priKeyPath)
	if err != nil {
		return nil, err
	}
	password, err := openAndReadFile(passwordPath)
	if err != nil {
		return nil, err
	}
	keyBlock, rest := pem.Decode(privateKey)
	if keyBlock == nil {
		return nil, fmt.Errorf("failed to decode pem, content: %s", rest)
	}
	// Remove the newline after the password
	// password = []byte(strings.TrimRight(string(password), "\n"))
	// Decrypt private key using password
	keyDER, err := x509.DecryptPEMBlock(keyBlock, password)
	if err != nil {
		return nil, fmt.Errorf("failed to DecryptPEMBlock: %s", err.Error())
	}
	keyBlock.Bytes = keyDER
	keyPEM := pem.EncodeToMemory(keyBlock)

	// Generate the client certificate
	clientCert, err := tls.X509KeyPair(cert, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to generate the client certificate: %s", err.Error())
	}

	// Parse root CA pem and add into CA pool
	rootCAPool := x509.NewCertPool()
	ok := rootCAPool.AppendCertsFromPEM(rootCA)
	if !ok {
		return nil, fmt.Errorf("unable to append supplied cert into tls.Config, are you sure it is a valid certificate")
	}

	return &SslConfig{
		SslConf: tls.Config{
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      rootCAPool,
		},
		IsCaSigned: true,
	}, nil
}

func (s *SslConfig) SetConfig(conf tls.Config) {
	s.SslConf = conf
}

// GetDefaultConf returns the default config
func GetDefaultConf() PoolConfig {
	return PoolConfig{
		TimeOut:         0 * time.Millisecond,
		IdleTime:        0 * time.Millisecond,
		MaxConnPoolSize: 10,
		MinConnPoolSize: 0,
	}
}

func openAndReadFile(path string) ([]byte, error) {
	// open file
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("unable to open test file %s: %s", path, err))
	}
	// read file
	out, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("unable to ReadAll of test file %s: %s", path, err))
	}
	return out, nil
}
