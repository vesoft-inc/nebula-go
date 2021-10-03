/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package nebula_go

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestSSLConnection(t *testing.T) {
	// skip test when ssl_test is not set to true
	skipCI(t)

	hostAdress := HostAddress{Host: address, Port: port}
	// hostAdress := HostAddress{Host: "192.168.8.6", Port: 29562}
	hostList := []HostAddress{}
	hostList = append(hostList, hostAdress)

	testPoolConfig = PoolConfig{
		TimeOut:         0 * time.Millisecond,
		IdleTime:        0 * time.Millisecond,
		MaxConnPoolSize: 10,
		MinConnPoolSize: 1,
	}

	var (
		rootCA     = openAndReadFile(t, "./nebula-docker-compose/secrets/test.ca.pem")
		cert       = openAndReadFile(t, "./nebula-docker-compose/secrets/test.client.crt")
		privateKey = openAndReadFile(t, "./nebula-docker-compose/secrets/test.client.key")
	)

	// generate the client certificate
	clientCert, err := tls.X509KeyPair(cert, privateKey)
	if err != nil {
		panic(err)
	}

	// parse root CA pem and add into CA pool
	rootCAPool := x509.NewCertPool()
	ok := rootCAPool.AppendCertsFromPEM(rootCA)
	if !ok {
		t.Fatal("unable to append supplied cert into tls.Config, are you sure it is a valid certificate")
	}

	// set tls config
	// InsecureSkipVerify is set to true for test purpose ONLY. DO NOT use it in production.
	sslConfig := &tls.Config{
		Certificates:       []tls.Certificate{clientCert},
		RootCAs:            rootCAPool,
		InsecureSkipVerify: true, // This is only used for testing
	}

	// Initialize connectin pool
	pool, err := NewSslConnectionPool(hostList, testPoolConfig, sslConfig, nebulaLog)
	if err != nil {
		t.Fatalf("fail to initialize the connection pool, host: %s, port: %d, %s", address, port, err.Error())
	}
	// close all connections in the pool
	defer pool.Close()

	// Create session
	session, err := pool.GetSession(username, password)
	if err != nil {
		t.Fatalf("fail to create a new session from connection pool, username: %s, password: %s, %s",
			username, password, err.Error())
	}
	defer session.Release()
	// Excute a query
	resp, err := tryToExecute(session, "SHOW HOSTS;")
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	checkResultSet(t, "show hosts", resp)
	// Create a new space
	resp, err = tryToExecute(session, "CREATE SPACE client_test(partition_num=1024, replica_factor=1, vid_type = FIXED_STRING(30));")
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	checkResultSet(t, "create space", resp)

	resp, err = tryToExecute(session, "DROP SPACE client_test;")
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	checkResultSet(t, "drop space", resp)
}

func openAndReadFile(t *testing.T, path string) []byte {
	// open file
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf(fmt.Sprintf("unable to open test file %s: %s", path, err))
	}
	// read file
	b, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatalf(fmt.Sprintf("unable to ReadAll of test file %s: %s", path, err))
	}
	return b
}

func skipCI(t *testing.T) {
	if os.Getenv("ssl_test") != "true" {
		t.Skip("Skipping SSL testing in CI environment")
	}
}
