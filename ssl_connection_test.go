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

func TestSslConnection(t *testing.T) {
	// skip test when ssl_test is not set to true
	skipSsl(t)

	hostAdress := HostAddress{Host: address, Port: port}
	hostList := []HostAddress{}
	hostList = append(hostList, hostAdress)

	testPoolConfig = PoolConfig{
		TimeOut:         0 * time.Millisecond,
		IdleTime:        0 * time.Millisecond,
		MaxConnPoolSize: 10,
		MinConnPoolSize: 1,
	}

	sslConfig, err := GetDefaultSSLConfig(
		"./nebula-docker-compose/secrets/test.ca.pem",
		"./nebula-docker-compose/secrets/test.client.crt",
		"./nebula-docker-compose/secrets/test.client.key",
	)

	sslConfig.InsecureSkipVerify = true // This is only used for testing

	// Initialize connection pool
	pool, err := NewSslConnectionPool(hostList, testPoolConfig, sslConfig, nebulaLog)
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

	sslConfig, err := GetDefaultSSLConfig(
		"./nebula-docker-compose/secrets/test.self-signed.pem",
		"./nebula-docker-compose/secrets/test.self-signed.pem",
		"./nebula-docker-compose/secrets/test.self-signed.key",
	)

	sslConfig.InsecureSkipVerify = true // This is only used for testing

	// Initialize connection pool
	pool, err := NewSslConnectionPool(hostList, testPoolConfig, sslConfig, nebulaLog)
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

func openAndReadFileTest(t *testing.T, path string) []byte {
	b, err := openAndReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func skipSsl(t *testing.T) {
	if os.Getenv("ssl_test") != "true" {
		t.Skip("Skipping SSL testing in CI environment")
	}
}

func skipSslSelfSigned(t *testing.T) {
	if os.Getenv("self_signed") != "true" {
		t.Skip("Skipping SSL testing in CI environment")
	}
}
