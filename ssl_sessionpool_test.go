//go:build integration
// +build integration

/*
 *
 * Copyright (c) 2023 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 *
 */

package nebula_go

import (
	"testing"
	"time"
)

func TestSslSessionPool(t *testing.T) {
	skipSsl(t)

	hostAddress := HostAddress{Host: address, Port: port}
	hostList := []HostAddress{}
	hostList = append(hostList, hostAddress)
	sslConfig, err := GetDefaultSSLConfig(
		"./nebula-docker-compose/secrets/root.crt",
		"./nebula-docker-compose/secrets/client.crt",
		"./nebula-docker-compose/secrets/client.key",
	)
	if err != nil {
		t.Fatal(err)
	}
	sslConfig.InsecureSkipVerify = true // This is only used for testing
	conf, err := NewSessionPoolConf(
		username,
		password,
		hostList,
		"session_pool",
		WithMaxSize(10),
		WithMinSize(1),
		WithTimeOut(0*time.Millisecond),
		WithIdleTime(0*time.Millisecond),
		WithSSLConfig(sslConfig),
	)

	if err != nil {
		t.Fatal(err)
	}
	pool, err := NewSessionPool(*conf, nebulaLog)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	resp, err := pool.Execute("SHOW HOSTS;")
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	checkResultSet(t, "show hosts", resp)
	// Create a new space
	resp, err = pool.Execute("CREATE SPACE client_test(partition_num=1024, replica_factor=1, vid_type = FIXED_STRING(30));")
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	checkResultSet(t, "create space", resp)

	resp, err = pool.Execute("DROP SPACE client_test;")
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	checkResultSet(t, "drop space", resp)
}
