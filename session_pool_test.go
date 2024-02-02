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
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vesoft-inc/nebula-go/v3/nebula"
	"github.com/vesoft-inc/nebula-go/v3/nebula/graph"
	"golang.org/x/net/context"
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

func TestSessionPoolServerCheck(t *testing.T) {
	prepareSpace("client_test")
	defer dropSpace("client_test")
	hostAddress := HostAddress{Host: address, Port: port}
	testcases := []struct {
		conf   *SessionPoolConf
		errMsg string
	}{
		{
			conf: &SessionPoolConf{
				username:     "root",
				password:     "nebula",
				serviceAddrs: []HostAddress{hostAddress},
				spaceName:    "invalid_space",
				minSize:      1,
			},
			errMsg: "failed to create a new session pool, " +
				"failed to initialize the session pool, " +
				"failed to use space invalid_space: SpaceNotFound: SpaceName `invalid_space`",
		},
		{
			conf: &SessionPoolConf{
				username:     "root1",
				password:     "nebula",
				serviceAddrs: []HostAddress{hostAddress},
				spaceName:    "client_test",
				minSize:      1,
			},
			errMsg: "failed to create a new session pool, " +
				"failed to initialize the session pool, " +
				"failed to authenticate the user, error code: -1001, " +
				"error message: User not exist, the pool has been closed",
		},
		{
			conf: &SessionPoolConf{
				username:     "root",
				password:     "nebula1",
				serviceAddrs: []HostAddress{hostAddress},
				spaceName:    "client_test",
				minSize:      1,
			},
			errMsg: "failed to create a new session pool, " +
				"failed to initialize the session pool, " +
				"failed to authenticate the user, error code: -1001, " +
				"error message: Invalid password, the pool has been closed",
		},
		{
			conf: &SessionPoolConf{
				username:     "root",
				password:     "nebula1",
				serviceAddrs: []HostAddress{{"127.0.0.1", 1234}},
				spaceName:    "client_test",
				minSize:      1,
			},
			errMsg: "failed to create a new session pool, " +
				"failed to initialize the session pool, " +
				"failed to open transport, " +
				"error: dial tcp 127.0.0.1:1234: connect: connection refused",
		},
	}
	for _, tc := range testcases {
		_, err := NewSessionPool(*tc.conf, DefaultLogger{})
		if err == nil {
			t.Fatal("should return error")
		}
		assert.Equal(t, err.Error(), tc.errMsg,
			fmt.Sprintf("expected error: %s, but actual error: %s", tc.errMsg, err.Error()))
	}
}

func TestSessionPoolInvalidHandshakeKey(t *testing.T) {
	prepareSpace("client_test")
	defer dropSpace("client_test")
	hostAddress := HostAddress{Host: address, Port: port}

	// wrong handshakeKey info
	versionConfig, err := NewSessionPoolConf(
		"root",
		"nebula",
		[]HostAddress{hostAddress},
		"client_test",
	)
	versionConfig.handshakeKey = "INVALID_VERSION"
	versionConfig.minSize = 1

	// create session pool
	_, err = NewSessionPool(*versionConfig, DefaultLogger{})
	if err != nil {
		assert.Contains(t, err.Error(), "incompatible handshakeKey between client and server")
	}
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
	rsCh := make(chan *ResultSet, sessionPool.conf.maxSize)
	done := make(chan bool)
	wg.Add(sessionPool.conf.maxSize)

	// producer create sessions
	for i := 0; i < sessionPool.conf.maxSize; i++ {
		go func(sessCh chan<- *ResultSet, wg *sync.WaitGroup) {
			defer wg.Done()
			rs, err := sessionPool.Execute("yield 1")
			if err != nil {
				t.Errorf("fail to execute query, %s", err.Error())
			}

			rsCh <- rs
		}(rsCh, &wg)
	}

	// consumer consumes the session created
	var rsList []*ResultSet
	go func(rsCh <-chan *ResultSet) {
		for session := range rsCh {
			rsList = append(rsList, session)
		}
		done <- true
	}(rsCh)
	wg.Wait()
	close(rsCh)
	<-done

	assert.Equalf(t, 0, sessionPool.activeSessions.Len(),
		"Total number of active connections should be %d", config.maxSize)
	assert.Equalf(t, config.maxSize, len(rsList),
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

func TestSessionPoolApplySchema(t *testing.T) {
	err := prepareSpace("test_space_schema")
	if err != nil {
		t.Fatal(err)
	}
	defer dropSpace("test_space_schema")

	hostAddress := HostAddress{Host: address, Port: port}
	config, err := NewSessionPoolConf(
		"root",
		"nebula",
		[]HostAddress{hostAddress},
		"test_space_schema")
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

	tagSchema := LabelSchema{
		Name: "account",
		Fields: []LabelFieldSchema{
			{
				Field:    "name",
				Nullable: false,
			},
			{
				Field:    "email",
				Nullable: true,
			},
			{
				Field:    "phone",
				Type:     "int64",
				Nullable: true,
			},
		},
	}
	_, err = sessionPool.CreateTag(tagSchema)
	if err != nil {
		t.Fatal(err)
	}
	tags, err := sessionPool.ShowTags()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(tags), "should have 1 tags")
	assert.Equal(t, "account", tags[0].Name, "tag name should be account")
	labels, err := sessionPool.DescTag("account")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 3, len(labels), "should have 3 labels")
	assert.Equal(t, "name", labels[0].Field, "field name should be name")
	assert.Equal(t, "string", labels[0].Type, "field type should be string")
	assert.Equal(t, "email", labels[1].Field, "field name should be email")
	assert.Equal(t, "string", labels[1].Type, "field type should be string")
	assert.Equal(t, "phone", labels[2].Field, "field name should be phone")
	assert.Equal(t, "int64", labels[2].Type, "field type should be string")

	edgeSchema := LabelSchema{
		Name: "account_email",
		Fields: []LabelFieldSchema{
			{
				Field:    "email",
				Nullable: false,
			},
		},
	}
	_, err = sessionPool.CreateEdge(edgeSchema)
	if err != nil {
		t.Fatal(err)
	}
	edges, err := sessionPool.ShowEdges()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(edges), "should have 1 edges")
	assert.Equal(t, "account_email", edges[0].Name, "edge name should be account_email")
	labels, err = sessionPool.DescEdge("account_email")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(labels), "should have 1 labels")
	assert.Equal(t, "email", labels[0].Field, "field name should be email")
	assert.Equal(t, "string", labels[0].Type, "field type should be string")
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

// retry when return the error code *ErrorCode_E_SESSION_INVALID*
func TestSessionPoolRetry(t *testing.T) {
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
	testcaes := []struct {
		name    string
		retryFn func(*pureSession) (*ResultSet, error)
		retry   bool
	}{
		{
			name: "success",
			retryFn: func(s *pureSession) (*ResultSet, error) {
				return &ResultSet{
					resp: &graph.ExecutionResponse{
						ErrorCode: nebula.ErrorCode_SUCCEEDED,
					},
				}, nil
			},
			retry: false,
		},
		{
			name: "error",
			retryFn: func(s *pureSession) (*ResultSet, error) {
				return nil, fmt.Errorf("error")
			},
			retry: true,
		},
		{
			name: "invalid session error code",
			retryFn: func(s *pureSession) (*ResultSet, error) {
				return &ResultSet{
					resp: &graph.ExecutionResponse{
						ErrorCode: nebula.ErrorCode_E_SESSION_INVALID,
					},
				}, nil
			},
			retry: true,
		},
		{
			name: "execution error code",
			retryFn: func(s *pureSession) (*ResultSet, error) {
				return &ResultSet{
					resp: &graph.ExecutionResponse{
						ErrorCode: nebula.ErrorCode_E_EXECUTION_ERROR,
					},
				}, nil
			},
			retry: false,
		},
	}
	for _, tc := range testcaes {
		session, err := sessionPool.newSession()
		if err != nil {
			t.Fatal(err)
		}
		original := session.sessionID
		conn := session.connection
		_, _ = sessionPool.executeWithRetry(session, tc.retryFn, 2)
		if tc.retry {
			assert.NotEqual(t, original, session.sessionID, fmt.Sprintf("test case: %s", tc.name))
			assert.NotEqual(t, conn, nil, fmt.Sprintf("test case: %s", tc.name))
		} else {
			assert.Equal(t, original, session.sessionID, fmt.Sprintf("test case: %s", tc.name))
		}
	}
}

func TestSessionPoolClose(t *testing.T) {
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
	sessionPool.Close()

	assert.Equal(t, 0, sessionPool.activeSessions.Len(), "Total number of active connections should be 0")
	assert.Equal(t, 0, sessionPool.idleSessions.Len(), "Total number of active connections should be 0")
	_, err = sessionPool.Execute("SHOW HOSTS;")
	assert.Equal(t, err.Error(), "failed to execute: Session pool has been closed", "session pool should be closed")
}

// TestSessionPoolGetSessionTimeout tests the scenario that if all requests are timeout,
// the session pool should return timeout error, not reach the pool size limit.
func TestQueryTimeout(t *testing.T) {
	hostAddress := HostAddress{Host: address, Port: port}
	config, err := NewSessionPoolConf(
		"root",
		"nebula",
		[]HostAddress{hostAddress},
		"test_data")
	if err != nil {
		t.Errorf("failed to create session pool config, %s", err.Error())
	}
	config.minSize = 0
	config.maxSize = 10
	config.retryGetSessionTimes = 1
	config.timeOut = 100 * time.Millisecond
	// create session pool
	sessionPool, err := NewSessionPool(*config, DefaultLogger{})
	if err != nil {
		t.Fatal(err)
	}
	defer sessionPool.Close()
	createTestDataSchema(t, sessionPool)
	loadTestData(t, sessionPool)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	errCh := make(chan error, 1)
	defer cancel()
	var wg sync.WaitGroup
	for i := 0; i < config.maxSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				select {
				case <-ctx.Done():
					return
				default:
					_, err := sessionPool.Execute(`go 2000 step from "Bob" over like yield tags($$)`)
					if err == nil {
						errCh <- fmt.Errorf("should return error")
						return
					}
					errMsg := "i/o timeout"
					if !strings.Contains(err.Error(), errMsg) {
						errCh <- fmt.Errorf("expect error contains: %s, but actual: %s", errMsg, err.Error())
						return
					}
				}
			}
		}()
	}
	wg.Wait()
	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}
}
