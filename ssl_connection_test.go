/*
 *
 * Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 *
 */

package nebula_go

import (
	"os"
	"testing"
	"time"
)

func TestSslConnectionCaSigned(t *testing.T) {
	// skip test when ssl_test is not set to true
	skipSslCaSigned(t)

	hostAdress := HostAddress{Host: address, Port: port}
	hostList := []HostAddress{}
	hostList = append(hostList, hostAdress)

	testPoolConfig = PoolConfig{
		TimeOut:         0 * time.Millisecond,
		IdleTime:        0 * time.Millisecond,
		MaxConnPoolSize: 10,
		MinConnPoolSize: 1,
	}

	sslConf, err := NewCaSignedSslConf("./nebula-docker-compose/secrets/test.ca.pem",
		"./nebula-docker-compose/secrets/test.client.crt",
		"./nebula-docker-compose/secrets/test.client.key")
	if err != nil {
		t.Fatalf(err.Error())
	}
	// InsecureSkipVerify is set to true for test purpose ONLY. DO NOT use it in production.
	sslConf.SslConf.InsecureSkipVerify = true

	// Initialize connection pool
	pool, err := NewSslConnectionPool(hostList, testPoolConfig, sslConf, nebulaLog)
	if err != nil {
		t.Fatalf("fail to initialize the connection pool, host: %s, port: %d, %s", address, port, err.Error())
	}
	// Close all connections in the pool
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

// TODO: generate certificate with hostName info and disable InsecureSkipVerify
func TestSslConnectionSelfSigned(t *testing.T) {
	// skip test when ssl_test is not set to true
	skipSslSelfSigned(t)

	hostAdress := HostAddress{Host: address, Port: port}
	hostList := []HostAddress{}
	hostList = append(hostList, hostAdress)

	testPoolConfig = PoolConfig{
		TimeOut:         0 * time.Millisecond,
		IdleTime:        0 * time.Millisecond,
		MaxConnPoolSize: 10,
		MinConnPoolSize: 1,
	}

	sslConf, err := NewSelfSignedSslConf(
		"./nebula-docker-compose/secrets/test.self-signed.pem",
		"./nebula-docker-compose/secrets/test.self-signed.key",
		"./nebula-docker-compose/secrets/test.self-signed.password")
	if err != nil {
		t.Fatalf(err.Error())
	}
	// InsecureSkipVerify is set to true for test purpose ONLY. DO NOT use it in production.
	sslConf.SslConf.InsecureSkipVerify = true

	// Initialize connection pool
	pool, err := NewSslConnectionPool(hostList, testPoolConfig, sslConf, nebulaLog)
	if err != nil {
		t.Fatalf("fail to initialize the connection pool, host: %s, port: %d, %s", address, port, err.Error())
	}
	// Close all connections in the pool
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

func skipSslCaSigned(t *testing.T) {
	if os.Getenv("ca_signed") != "true" {
		t.Skip("Skipping CA-signed SSL testing in CI environment")
	}
}

func skipSslSelfSigned(t *testing.T) {
	if os.Getenv("self_signed") != "true" {
		t.Skip("Skipping self-signed SSL testing in CI environment")
	}
}
