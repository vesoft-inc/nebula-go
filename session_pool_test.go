//go:build integration
// +build integration

/*
 *
 * Copyright (c) 2022 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 *
 */
package nebula_go

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionPool_Basic(t *testing.T) {
	err := prepareSpace(t)
	assert.Nil(t, err)

	hostAddress := HostAddress{Host: address, Port: 29562}
	config, err := NewSessionPoolConf("root", "nebula", []HostAddress{hostAddress}, "nba")
	if err != nil {
		t.Errorf("failed to create session pool config, %s", err.Error())
	}

	// create session pool
	sessionPool, err := NewSessionPool(*config, DefaultLogger{})
	if err != nil {
		t.Fatal(err)
	}
	defer sessionPool.Close()

	// execute query
	resultSet, err := sessionPool.Execute("RETURN 1")
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, resultSet.IsSucceed())
}

func prepareSpace(t *testing.T) error {
	hostAddress := HostAddress{Host: address, Port: 29562}
	conn := newConnection(hostAddress)
	testPoolConfig := GetDefaultConf()

	err := conn.open(hostAddress, testPoolConfig.TimeOut, nil)
	if err != nil {
		t.Fatalf("fail to open connection, address: %s, port: %d, %s", address, port, err.Error())
	}

	authResp, authErr := conn.authenticate(username, password)
	if authErr != nil {
		t.Fatalf("fail to authenticate, username: %s, password: %s, %s", username, password, authErr.Error())
	}

	sessionID := authResp.GetSessionID()

	defer logoutAndClose(conn, sessionID)

	resp, err := conn.execute(sessionID, "CREATE SPACE IF NOT EXISTS"+
		" client_test(partition_num=32, replica_factor=1, vid_type = FIXED_STRING(30));")
	if err != nil {
		return err
	}
	checkConResp(t, "create space", resp)
	return nil
}
