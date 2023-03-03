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
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSessionPoolInvalidConfig(t *testing.T) {
	hostAddress := HostAddress{Host: address, Port: port}

	// No space name
	_, err := NewSessionPoolConf(
		"root",
		"nebula",
		[]HostAddress{hostAddress},
		"",
	)
	assert.Contains(t, err.Error(), "invalid session pool config: Space name is empty",
		"error message should contain Space name is empty")

	// No username and password
	_, err = NewSessionPoolConf(
		"",
		"",
		[]HostAddress{hostAddress},
		"client_test",
	)
	assert.Contains(t, err.Error(), "Username is empty", "error message should contain Username is empty")

	// No service address
	_, err = NewSessionPoolConf("root",
		"nebula",
		[]HostAddress{},
		"client_test",
	)
	assert.Contains(t, err.Error(), "invalid session pool config: Service address is empty",
		"error message should contain Service address is empty")
}

func TestSessionPoolBasic(t *testing.T) {
	prepareSpace("client_test")
	defer dropSpace("client_test")

	hostAddress := HostAddress{Host: address, Port: port}
	config, err := NewSessionPoolConf(
		"root",
		"nebula",
		[]HostAddress{hostAddress},
		"client_test")
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

func TestSessionPoolMultiThreadGetSession(t *testing.T) {
	err := prepareSpace("client_test")
	if err != nil {
		t.Fatal(err)
	}
	defer dropSpace("client_test")

	hostList := poolAddress
	config, err := NewSessionPoolConf(
		"root",
		"nebula",
		hostList,
		"client_test")
	if err != nil {
		t.Errorf("failed to create session pool config, %s", err.Error())
	}
	config.maxSize = 333

	// create session pool
	sessionPool, err := NewSessionPool(*config, DefaultLogger{})
	if err != nil {
		t.Fatal(err)
	}
	defer sessionPool.Close()

	var wg sync.WaitGroup
	sessCh := make(chan *Session)
	done := make(chan bool)
	wg.Add(sessionPool.conf.maxSize)

	// producer create sessions
	for i := 0; i < sessionPool.conf.maxSize; i++ {
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

	assert.Equalf(t, config.maxSize, sessionPool.activeSessions.Len(),
		"Total number of active connections should be %d", config.maxSize)
	assert.Equalf(t, config.maxSize, len(sessionList),
		"Total number of result returned should be %d", config.maxSize)
}

func TestSessionPoolMultiThreadExecute(t *testing.T) {
	err := prepareSpace("client_test")
	if err != nil {
		t.Fatal(err)
	}
	defer dropSpace("client_test")

	hostList := poolAddress
	config, err := NewSessionPoolConf(
		"root",
		"nebula",
		hostList,
		"client_test")
	if err != nil {
		t.Errorf("failed to create session pool config, %s", err.Error())
	}
	config.maxSize = 300

	// create session pool
	sessionPool, err := NewSessionPool(*config, DefaultLogger{})
	if err != nil {
		t.Fatal(err)
	}
	defer sessionPool.Close()

	var wg sync.WaitGroup
	wg.Add(sessionPool.conf.maxSize)
	respCh := make(chan *ResultSet)
	done := make(chan bool)

	for i := 0; i < sessionPool.conf.maxSize; i++ {
		go func(respCh chan<- *ResultSet, wg *sync.WaitGroup) {
			defer wg.Done()
			resp, err := sessionPool.Execute("SHOW HOSTS")
			if err != nil {
				t.Errorf(err.Error())
			}
			respCh <- resp
		}(respCh, &wg)
	}

	var respList []*ResultSet
	go func(respCh <-chan *ResultSet) {
		for resp := range respCh {
			respList = append(respList, resp)
		}
		done <- true
	}(respCh)
	wg.Wait()
	close(respCh)
	<-done

	// should generate config.maxSize results
	assert.Equalf(t, config.maxSize, len(respList),
		"Total number of response should be %d", config.maxSize)

	// should be 0 active sessions because they are put back to idle session list after
	// query execution
	assert.Equal(t, 0, sessionPool.activeSessions.Len(),
		"Total number of active sessions should be 0")
	// Note that here the idle session number may not be equal to the max size because once the query execution
	// finished, the session will be put back to the idle session list and could be reused by other goroutines.
}

// This test is used to test if the space bond to session is the same as the space in the session pool config after executing
// a query contains `USE <space_name>` statement.
func TestSessionPoolSpaceChange(t *testing.T) {
	err := prepareSpace("test_space_1")
	if err != nil {
		t.Fatal(err)
	}
	defer dropSpace("test_space_1")

	err = prepareSpace("test_space_2")
	if err != nil {
		t.Fatal(err)
	}
	defer dropSpace("test_space_2")

	hostAddress := HostAddress{Host: address, Port: port}
	config, err := NewSessionPoolConf(
		"root",
		"nebula",
		[]HostAddress{hostAddress},
		"test_space_1")
	if err != nil {
		t.Errorf("failed to create session pool config, %s", err.Error())
	}

	// allow only one session in the pool so it is easier to test
	config.maxSize = 1

	// create session pool
	sessionPool, err := NewSessionPool(*config, DefaultLogger{})
	if err != nil {
		t.Fatal(err)
	}
	defer sessionPool.Close()

	// execute query in test_space_2
	resultSet, err := sessionPool.Execute("USE test_space_2; SHOW HOSTS;")
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, resultSet.IsSucceed(), fmt.Errorf("error code: %d, error msg: %s",
		resultSet.GetErrorCode(), resultSet.GetErrorMsg()))

	// this query should be executed in test_space_1
	resultSet, err = sessionPool.Execute("SHOW HOSTS;")
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, resultSet.IsSucceed(), fmt.Errorf("error code: %d, error msg: %s",
		resultSet.GetErrorCode(), resultSet.GetErrorMsg()))
	assert.Equal(t, resultSet.GetSpaceName(), "test_space_1", "space name should be test_space_1")
}

func TestIdleSessionCleaner(t *testing.T) {
	err := prepareSpace("client_test")
	if err != nil {
		t.Fatal(err)
	}
	defer dropSpace("client_test")

	hostAddress := HostAddress{Host: address, Port: port}
	idleTimeoutConfig, err := NewSessionPoolConf(
		"root",
		"nebula",
		[]HostAddress{hostAddress},
		"client_test")
	if err != nil {
		t.Errorf("failed to create session pool config, %s", err.Error())
	}

	idleTimeoutConfig.idleTime = 2 * time.Second
	idleTimeoutConfig.minSize = 5
	idleTimeoutConfig.maxSize = 100

	// create session pool
	sessionPool, err := NewSessionPool(*idleTimeoutConfig, DefaultLogger{})
	if err != nil {
		t.Fatal(err)
	}
	defer sessionPool.Close()
	assert.Equal(t, 5, sessionPool.activeSessions.Len()+sessionPool.idleSessions.Len(),
		"Total number of sessions should be 5")

	// execute multiple queries so more sessions will be created
	var wg sync.WaitGroup
	wg.Add(sessionPool.conf.maxSize)

	for i := 0; i < sessionPool.conf.maxSize; i++ {
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			_, err := sessionPool.Execute("RETURN 1")
			if err != nil {
				t.Errorf(err.Error())
			}
		}(&wg)
	}
	wg.Wait()

	// wait for sessions to be idle
	time.Sleep(idleTimeoutConfig.idleTime)

	// the minimum interval for cleanup is 1 minute, so in CI we need to trigger cleanup manually
	sessionPool.cleanerChan <- struct{}{}
	time.Sleep(idleTimeoutConfig.idleTime + 1)

	// after cleanup, the total session should be 5 which is the minSize
	assert.Truef(t, sessionPool.GetTotalSessionCount() == sessionPool.conf.minSize,
		"Total number of session should be %d, but got %d",
		sessionPool.conf.minSize, sessionPool.GetTotalSessionCount())
}

func TestRetryGetSession(t *testing.T) {
	err := prepareSpace("client_test")
	if err != nil {
		t.Fatal(err)
	}
	defer dropSpace("client_test")

	hostAddress := HostAddress{Host: address, Port: port}
	config, err := NewSessionPoolConf(
		"root",
		"nebula",
		[]HostAddress{hostAddress},
		"client_test")
	if err != nil {
		t.Errorf("failed to create session pool config, %s", err.Error())
	}
	config.minSize = 2
	config.maxSize = 2
	config.retryGetSessionTimes = 1

	// create session pool
	sessionPool, err := NewSessionPool(*config, DefaultLogger{})
	if err != nil {
		t.Fatal(err)
	}
	defer sessionPool.Close()

	// kill all sessions in the cluster
	resultSet, err := sessionPool.Execute("SHOW SESSIONS | KILL SESSIONS $-.SessionId")
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, resultSet.IsSucceed(), fmt.Errorf("error code: %d, error msg: %s",
		resultSet.GetErrorCode(), resultSet.GetErrorMsg()))

	// execute query, it should retry to get session
	resultSet, err = sessionPool.Execute("SHOW HOSTS;")
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, resultSet.IsSucceed(), fmt.Errorf("error code: %d, error msg: %s",
		resultSet.GetErrorCode(), resultSet.GetErrorMsg()))
}

func BenchmarkConcurrency(b *testing.B) {
	err := prepareSpace("client_test")
	if err != nil {
		b.Fatal(err)
	}
	defer dropSpace("client_test")

	// create session pool config
	hostAddress := HostAddress{Host: address, Port: port}
	config, err := NewSessionPoolConf(
		"root",
		"nebula",
		[]HostAddress{hostAddress},
		"client_test",
		WithMaxSize(1200),
		WithMinSize(1000))
	if err != nil {
		b.Errorf("failed to create session pool config, %s", err.Error())
	}

	// create session pool
	sessionPool, err := NewSessionPool(*config, DefaultLogger{})
	if err != nil {
		b.Fatal(err)
	}
	defer sessionPool.Close()

	concurrencyLevels := []int{5, 10, 100, 1000}
	for _, clients := range concurrencyLevels {
		start := time.Now()
		b.Run(fmt.Sprintf("%d_clients", clients), func(b *testing.B) {
			wg := sync.WaitGroup{}
			for n := 0; n < clients; n++ {
				wg.Add(1)
				go func() {
					_, err := sessionPool.Execute("SHOW HOSTS;")
					if err != nil {
						b.Errorf(err.Error())
					}
					wg.Done()
				}()
			}
			wg.Wait()
		})
		end := time.Now()
		b.Logf("Concurrency: %d, Total time cost: %v", clients, end.Sub(start))
	}
}
