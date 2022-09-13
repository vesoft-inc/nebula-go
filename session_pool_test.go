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
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionPoolInvalidConfig(t *testing.T) {
	hostAddress := HostAddress{Host: address, Port: port}
	_, err := NewSessionPoolConf("root", "nebula", []HostAddress{hostAddress}, "")
	assert.Contains(t, err.Error(), "invalid session pool config: Space name is empty",
		"error message should contain Space name is empty")

	_, err = NewSessionPoolConf("", "", []HostAddress{hostAddress}, "test")
	assert.Contains(t, err.Error(), "Username is empty", "error message should contain Username is empty")

	_, err = NewSessionPoolConf("root", "nebula", []HostAddress{}, "")
	assert.Contains(t, err.Error(), "invalid session pool config: Service address is empty",
		"error message should contain Service address is empty")
}

func TestSessionPoolBasic(t *testing.T) {
	prepareSpace(t, "client_test")
	defer dropSpace(t, "client_test")

	hostAddress := HostAddress{Host: address, Port: port}
	config, err := NewSessionPoolConf("root", "nebula", []HostAddress{hostAddress}, "client_test")
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
	assert.True(t, resultSet.IsSucceed(), fmt.Errorf("error code: %d, error msg: %s",
		resultSet.GetErrorCode(), resultSet.GetErrorMsg()))

	assert.Equal(t, 0, sessionPool.activeSessions.Len(), "Total number of active connections should be 0")
	assert.Equal(t, 1, sessionPool.idleSessions.Len(), "Total number of active connections should be 1")
}

func TestSessionPoolMultiThread(t *testing.T) {
	prepareSpace(t, "client_test")
	defer dropSpace(t, "client_test")

	hostList := poolAddress
	config, err := NewSessionPoolConf("root", "nebula", hostList, "client_test")
	if err != nil {
		t.Errorf("failed to create session pool config, %s", err.Error())
	}
	config.MaxSize = 333

	// test get idle session
	{
		// create session pool
		sessionPool, err := NewSessionPool(*config, DefaultLogger{})
		if err != nil {
			t.Fatal(err)
		}
		defer sessionPool.Close()

		var wg sync.WaitGroup
		sessCh := make(chan *Session)
		done := make(chan bool)
		wg.Add(sessionPool.conf.MaxSize)

		// producer creates sessions
		for i := 0; i < sessionPool.conf.MaxSize; i++ {
			go func(sessCh chan<- *Session, wg *sync.WaitGroup) {
				defer wg.Done()
				session, err := sessionPool.getIdleSession()
				if err != nil {
					t.Errorf("fail to create a new session from connection pool, %s", err.Error())
				}
				sessCh <- session
			}(sessCh, &wg)
		}

		// consumer consumes the session created
		var sessionList []*Session
		go func(sessCh <-chan *Session) {
			for session := range sessCh {
				sessionList = append(sessionList, session)
			}
			done <- true
		}(sessCh)
		wg.Wait()
		close(sessCh)
		<-done

		assert.Equal(t, 333, sessionPool.activeSessions.Len(), "Total number of active connections should be 333")
		assert.Equal(t, 333, len(sessionList), "Total number of result returned should be 333")
	}

	// test Execute()
	{
		// create session pool
		sessionPool, err := NewSessionPool(*config, DefaultLogger{})
		if err != nil {
			t.Fatal(err)
		}
		defer sessionPool.Close()

		var wg sync.WaitGroup
		wg.Add(sessionPool.conf.MaxSize)

		for i := 0; i < sessionPool.conf.MaxSize; i++ {
			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				_, err := sessionPool.Execute("RETURN 1")
				if err != nil {
					t.Errorf(err.Error())
				}
			}(&wg)
		}
		wg.Wait()
		assert.Equal(t, 0, sessionPool.activeSessions.Len(), "Total number of active connections should be 0")
	}
}

// TODO(Aiee): add more test cases
func TestSessionPoolSpaceChange(t *testing.T) {
	err := prepareSpace(t, "test_space_1")
	if err != nil {
		t.Fatal(err)
	}
	defer dropSpace(t, "test_space_1")

	err = prepareSpace(t, "test_space_2")
	if err != nil {
		t.Fatal(err)
	}
	defer dropSpace(t, "test_space_2")

	hostAddress := HostAddress{Host: address, Port: port}
	config, err := NewSessionPoolConf("root", "nebula", []HostAddress{hostAddress}, "test_space_1")
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
	resultSet, err := sessionPool.Execute("USE test_space_2; SHOW HOSTS;")
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, resultSet.IsSucceed(), fmt.Errorf("error code: %d, error msg: %s",
		resultSet.GetErrorCode(), resultSet.GetErrorMsg()))
}
