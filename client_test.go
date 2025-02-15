//go:build integration
// +build integration

/*
 *
 * Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 *
 */
package nebula_go

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vesoft-inc/nebula-go/v3/nebula"
	"github.com/vesoft-inc/nebula-go/v3/nebula/graph"
)

const (
	address  = "127.0.0.1"
	port     = 3699
	username = "root"
	password = "nebula"

	addressIPv6 = "::1"
)

var poolAddress = []HostAddress{
	{
		Host: "127.0.0.1",
		Port: 3699,
	},
	{
		Host: "127.0.0.1",
		Port: 3700,
	},
	{
		Host: "127.0.0.1",
		Port: 3701,
	},
}

var nebulaLog = DefaultLogger{}

// Create default configs
var testPoolConfig = GetDefaultConf()

// Before run `go test -v`, you should start a nebula server listening on 3699 port.
// Using docker-compose is the easiest way and you can reference this file:
//   https://github.com/vesoft-inc/nebula/blob/master/docker/docker-compose.yaml

func logoutAndClose(conn *connection, sessionID int64) {
	conn.signOut(sessionID)
	conn.close()
}

func TestConnection(t *testing.T) {
	hostAddress := HostAddress{Host: address, Port: port}
	conn := newConnection(hostAddress)
	err := conn.open(hostAddress, testPoolConfig.TimeOut, nil, false, nil, "")
	if err != nil {
		t.Fatalf("fail to open connection, address: %s, port: %d, %s", address, port, err.Error())
	}

	authResp, authErr := conn.authenticate(username, password)
	if authErr != nil {
		t.Fatalf("fail to authenticate, username: %s, password: %s, %s", username, password, authErr.Error())
	}

	if authResp.GetErrorCode() != nebula.ErrorCode_SUCCEEDED {
		t.Fatalf("failed to authenticate, error code: %d, error msg: %s",
			authResp.GetErrorCode(), authResp.GetErrorMsg())
	}

	sessionID := authResp.GetSessionID()

	defer logoutAndClose(conn, sessionID)

	resp, err := conn.execute(sessionID, "SHOW HOSTS;")
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	checkConResp("show hosts", resp)

	resp, err = conn.execute(sessionID, "CREATE SPACE client_test(partition_num=1024, replica_factor=1, vid_type = FIXED_STRING(30));")
	if err != nil {
		t.Error(err.Error())
		return
	}
	checkConResp("create space", resp)

	resp, err = conn.execute(sessionID, "return 1")
	if err != nil {
		t.Error(err.Error())
		return
	}
	checkConResp("return 1", resp)

	resp, err = conn.execute(sessionID, "DROP SPACE client_test;")
	if err != nil {
		t.Error(err.Error())
		return
	}
	checkConResp("drop space", resp)

	res := conn.ping()
	if res != true {
		t.Error("Connection ping failed")
		return
	}
}

func TestConnectionIPv6(t *testing.T) {
	hostAddress := HostAddress{Host: addressIPv6, Port: port}
	conn := newConnection(hostAddress)
	err := conn.open(hostAddress, testPoolConfig.TimeOut, nil, false, nil, "")
	if err != nil {
		t.Fatalf("fail to open connection, address: %s, port: %d, %s", address, port, err.Error())
	}

	authResp, authErr := conn.authenticate(username, password)
	if authErr != nil {
		t.Fatalf("fail to authenticate, username: %s, password: %s, %s", username, password, authErr.Error())
	}

	if authResp.GetErrorCode() != nebula.ErrorCode_SUCCEEDED {
		t.Fatalf("failed to authenticate, error code: %d, error msg: %s",
			authResp.GetErrorCode(), authResp.GetErrorMsg())
	}

	sessionID := authResp.GetSessionID()

	defer logoutAndClose(conn, sessionID)

	resp, err := conn.execute(sessionID, "SHOW HOSTS;")
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	checkConResp("show hosts", resp)

	resp, err = conn.execute(sessionID, "CREATE SPACE client_test(partition_num=1024, replica_factor=1, vid_type = FIXED_STRING(30));")
	if err != nil {
		t.Error(err.Error())
		return
	}
	checkConResp("create space", resp)
	resp, err = conn.execute(sessionID, "DROP SPACE client_test;")
	if err != nil {
		t.Error(err.Error())
		return
	}
	checkConResp("drop space", resp)

	res := conn.ping()
	if res != true {
		t.Error("Connection ping failed")
		return
	}
}

func TestConfigs(t *testing.T) {
	hostAddress := HostAddress{Host: address, Port: port}
	hostList := []HostAddress{}
	hostList = append(hostList, hostAddress)

	var configList = []PoolConfig{
		// default
		{
			TimeOut:         0 * time.Millisecond,
			IdleTime:        0 * time.Millisecond,
			MaxConnPoolSize: 10,
			MinConnPoolSize: 1,
		},
		// timeout < 0
		{
			TimeOut:         -1 * time.Millisecond,
			IdleTime:        0 * time.Millisecond,
			MaxConnPoolSize: 10,
			MinConnPoolSize: 1,
		},
		// MaxConnPoolSize < 0
		{
			TimeOut:         0 * time.Millisecond,
			IdleTime:        0 * time.Millisecond,
			MaxConnPoolSize: -1,
			MinConnPoolSize: 1,
		},
		// MinConnPoolSize < 0
		{
			TimeOut:         0 * time.Millisecond,
			IdleTime:        0 * time.Millisecond,
			MaxConnPoolSize: 1,
			MinConnPoolSize: -1,
		},
	}

	for _, testPoolConfig := range configList {
		// Initialize connection pool
		pool, err := NewConnectionPool(hostList, testPoolConfig, nebulaLog)
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
		// Execute a query
		resp, err := tryToExecute(session, "SHOW HOSTS;")
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		checkResultSet(t, "show hosts", resp)
		// Create a new space
		resp, err = tryToExecute(
			session,
			"CREATE SPACE client_test(partition_num=1024, replica_factor=1, vid_type = FIXED_STRING(30));")
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		checkResultSet(t, "create space", resp)

		err = dropSpace("client_test")
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
	}
}

func TestVersionVerify(t *testing.T) {
	const (
		username = "root"
		password = "nebula"
	)

	hostAddress := HostAddress{Host: address, Port: port}

	conn := newConnection(hostAddress)
	err := conn.open(hostAddress, testPoolConfig.TimeOut, nil, false, nil, "INVALID_VERSION")
	if err != nil {
		assert.Contains(t, err.Error(), "incompatible handshakeKey between client and server")
	}
	defer conn.close()
}

func TestAuthentication(t *testing.T) {
	const (
		username = "dummy"
		password = "nebula"
	)

	hostAddress := HostAddress{Host: address, Port: port}

	conn := newConnection(hostAddress)
	err := conn.open(hostAddress, testPoolConfig.TimeOut, nil, false, nil, "")
	if err != nil {
		t.Fatalf("fail to open connection, address: %s, port: %d, %s", address, port, err.Error())
	}
	defer conn.close()

	resp, authErr := conn.authenticate(username, password)
	if authErr != nil {
		t.Fatalf("fail to authenticate, username: %s, password: %s, %s", username, password, authErr.Error())
	}

	assert.Equal(t, string(resp.GetErrorMsg()), "User not exist")
}

func TestInvalidHostTimeout(t *testing.T) {
	hostAddress := HostAddress{Host: address, Port: port}
	// invalid host
	invalidHostAddress := HostAddress{Host: "192.168.100.125", Port: 3699}
	hostList := []HostAddress{hostAddress}

	invalidHostList := []HostAddress{
		invalidHostAddress, // Invalid host
		hostAddress,
	}

	// Initialize connection pool
	pool, err := NewConnectionPool(hostList, testPoolConfig, nebulaLog)
	if err != nil {
		t.Fatalf("fail to initialize the connection pool, host: %s, port: %d, %s", address, port, err.Error())
	}
	// close all connections in the pool
	defer pool.Close()
	err = pool.Ping(invalidHostList[0], 1000*time.Millisecond)
	assert.EqualError(t, err, "failed to open transport, error: dial tcp 192.168.100.125:3699: i/o timeout")
	err = pool.Ping(invalidHostList[1], 1000*time.Millisecond)
	if err != nil {
		t.Error("failed to ping 127.0.0.1")
	}
}

func TestServiceDataIO(t *testing.T) {
	hostAddress := HostAddress{Host: address, Port: port}
	hostList := []HostAddress{}
	hostList = append(hostList, hostAddress)

	testPoolConfig = PoolConfig{
		TimeOut:         0 * time.Millisecond,
		IdleTime:        0 * time.Millisecond,
		MaxConnPoolSize: 10,
		MinConnPoolSize: 1,
	}

	// Initialize connection pool
	pool, err := NewConnectionPool(hostList, testPoolConfig, nebulaLog)
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
	// Save session create time
	loc, _ := time.LoadLocation("Asia/Shanghai")
	sessionCreatedTime := time.Now().In(loc)
	defer session.Release()

	// Create schemas
	createTestDataSchema(t, session)
	// Load data
	loadTestData(t, session)

	// test base type
	{
		query :=
			"FETCH PROP ON person \"Bob\" YIELD vertex as VertexID," +
				"person.name, person.age, person.grade," +
				"person.friends, person.book_num, person.birthday, " +
				"person.start_school, person.morning, " +
				"person.property, person.is_girl, person.child_name, " +
				"person.expend, person.first_out_city, person.hobby"
		resp, err := tryToExecute(session, query)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		checkResultSet(t, query, resp)

		assert.True(t, resp.GetLatency() > 0)
		assert.Empty(t, resp.GetComment())
		assert.Equal(t, "test_data", resp.GetSpaceName())
		assert.False(t, resp.IsEmpty())
		assert.Equal(t, 1, resp.GetRowSize())
		names := []string{"VertexID",
			"person.name",
			"person.age",
			"person.grade",
			"person.friends",
			"person.book_num",
			"person.birthday",
			"person.start_school",
			"person.morning",
			"person.property",
			"person.is_girl",
			"person.child_name",
			"person.expend",
			"person.first_out_city",
			"person.hobby"}
		assert.Equal(t, names, resp.GetColNames())

		// test datetime
		record, err := resp.GetRowValuesByIndex(0)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		valWrap, err := record.GetValueByIndex(6)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		dateTimeWrapper, err := valWrap.AsDateTime()
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		// local time
		assert.Equal(t, "2010-09-10T10:08:02.000000", valWrap.String())
		// UTC time
		UTCDatetime := dateTimeWrapper.getRawDateTime()
		expectedDatetime := nebula.DateTime{2010, 9, 10, 2, 8, 2, 0}
		assert.Equal(t, expectedDatetime, *UTCDatetime)

		// test date
		valWrap, err = record.GetValueByIndex(7)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.True(t, valWrap.IsDate())
		assert.Equal(t, "2017-09-10", valWrap.String())

		// test time
		valWrap, err = record.GetValueByIndex(8)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		timeWrapper, err := valWrap.AsTime()
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.True(t, valWrap.IsTime())
		assert.Equal(t, "07:10:00.000000", valWrap.String())

		UTCTime := timeWrapper.getRawTime()
		expected := nebula.Time{23, 10, 0, 0}
		assert.Equal(t, expected, *UTCTime)
	}

	// test node
	{
		resp, err := tryToExecute(session, "MATCH (v:person {name: \"Bob\"}) RETURN v")
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.Equal(t, 1, resp.GetRowSize())
		record, err := resp.GetRowValuesByIndex(0)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		valWrap, err := record.GetValueByIndex(0)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		node, err := valWrap.AsNode()
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.Equal(t,
			"(\"Bob\" :student{interval: P1MT100.000020000S, name: \"Bob\"} "+
				":person{age: 10, birthday: 2010-09-10T10:08:02.000000, book_num: 100, "+
				"child_name: \"Hello Worl\", expend: 100.0, "+
				"first_out_city: 1111, friends: 10, grade: 3, "+
				"hobby: __NULL__, is_girl: false, "+
				"morning: 07:10:00.000000, name: \"Bob\", "+
				"property: 1000.0, start_school: 2017-09-10})",
			node.String())
		props, _ := node.Properties("person")
		datetime := props["birthday"]
		dtWrapper, err := datetime.AsDateTime()
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		utcTime := dtWrapper.getRawDateTime()
		expected := nebula.DateTime{2010, 9, 10, 2, 8, 2, 0}
		assert.Equal(t, expected, *utcTime)

		localTime, _ := dtWrapper.getLocalDateTime()
		expected = nebula.DateTime{2010, 9, 10, 10, 8, 2, 0}
		assert.Equal(t, expected, *localTime)
	}

	// test edge
	{
		resp, err := tryToExecute(session, "MATCH (:person{name: \"Bob\"}) -[e:friend]-> (:person{name: \"Lily\"}) RETURN e")
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.Equal(t, 1, resp.GetRowSize())
		record, err := resp.GetRowValuesByIndex(0)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		valWrap, err := record.GetValueByIndex(0)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		relationship, err := valWrap.AsRelationship()
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.Equal(t,
			"[:friend \"Bob\"->\"Lily\" @0 {end_Datetime: 2010-09-10T10:08:02.000000, start_Datetime: 2008-09-10T10:08:02.000000}]",
			relationship.String())
		props := relationship.Properties()
		datetime := props["end_Datetime"]
		dtWrapper, err := datetime.AsDateTime()
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		utcTime := dtWrapper.getRawDateTime()
		expected := nebula.DateTime{2010, 9, 10, 2, 8, 2, 0}
		assert.Equal(t, expected, *utcTime)

		localTime, _ := dtWrapper.getLocalDateTime()
		expected = nebula.DateTime{2010, 9, 10, 10, 8, 2, 0}
		assert.Equal(t, expected, *localTime)
	}

	// Test path
	{
		resp, err := tryToExecute(session, "MATCH p = (:person{name: \"Bob\"}) -[e:friend]-> (:person{name: \"Lily\"}) RETURN p")
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.Equal(t, 1, resp.GetRowSize())
		record, err := resp.GetRowValuesByIndex(0)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		valWrap, err := record.GetValueByIndex(0)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		path, err := valWrap.AsPath()
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.Equal(t,
			"<(\"Bob\" :student{interval: P1MT100.000020000S, name: \"Bob\"} :person{age: 10, birthday: 2010-09-10T10:08:02.000000, book_num: 100, child_name: \"Hello Worl\", expend: 100.0, first_out_city: 1111, friends: 10, grade: 3, hobby: __NULL__, is_girl: false, morning: 07:10:00.000000, name: \"Bob\", property: 1000.0, start_school: 2017-09-10})-[:friend@0 {end_Datetime: 2010-09-10T10:08:02.000000, start_Datetime: 2008-09-10T10:08:02.000000}]->(\"Lily\" :student{interval: P12MT0.000000000S, name: \"Lily\"} :person{age: 9, birthday: 2010-09-10T10:08:02.000000, book_num: 100, child_name: \"Hello Worl\", expend: 100.0, first_out_city: 1111, friends: 10, grade: 3, hobby: __NULL__, is_girl: false, morning: 07:10:00.000000, name: \"Lily\", property: 1000.0, start_school: 2017-09-10})>",
			path.String())
	}

	// Check timestamp
	{
		// test show jobs
		_, err := tryToExecute(session, "SUBMIT JOB STATS")
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		expected := int8(time.Now().In(loc).Hour())
		time.Sleep(5 * time.Second)

		resp, err := tryToExecute(session, "SHOW JOBS")
		if err != nil {
			t.Fatalf(err.Error())
			return
		}

		// Row[0][3] is the Start Time of the job
		record, err := resp.GetRowValuesByIndex(0)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		valWrap, err := record.GetValueByColName("Start Time")
		if err != nil {
			t.Fatalf(err.Error())
			return
		}

		dtWrapper, err := valWrap.AsDateTime()
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		localTime, err := dtWrapper.getLocalDateTime()
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.Equal(t, expected, localTime.GetHour())

		// test show sessions
		resp, err = tryToExecute(session, "SHOW SESSIONS")
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		// Row[0][4] is the CreateTime of the session
		record, err = resp.GetRowValuesByIndex(0)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		valWrap, err = record.GetValueByColName("CreateTime")
		if err != nil {
			t.Fatalf(err.Error())
			return
		}

		dtWrapper, err = valWrap.AsDateTime()
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		localTime, err = dtWrapper.getLocalDateTime()
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.Equal(t, int8(sessionCreatedTime.Hour()), localTime.GetHour())
	}
	err = dropSpace("client_test")
	if err != nil {
		t.Fatalf(err.Error())
	}

}

func TestPool_SingleHost(t *testing.T) {
	hostAddress := HostAddress{Host: address, Port: port}
	hostList := []HostAddress{}
	hostList = append(hostList, hostAddress)

	testPoolConfig = PoolConfig{
		TimeOut:         0 * time.Millisecond,
		IdleTime:        0 * time.Millisecond,
		MaxConnPoolSize: 10,
		MinConnPoolSize: 1,
	}

	// Initialize connection pool
	pool, err := NewConnectionPool(hostList, testPoolConfig, nebulaLog)
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
	// Execute a query
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

	err = dropSpace("client_test")
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestPool_MultiHosts(t *testing.T) {
	hostList := poolAddress
	// Minimum pool size < hosts number
	multiHostsConfig := PoolConfig{
		TimeOut:         0 * time.Millisecond,
		IdleTime:        0 * time.Millisecond,
		MaxConnPoolSize: 3,
		MinConnPoolSize: 1,
	}

	// Initialize connection pool
	pool, err := NewConnectionPool(hostList, multiHostsConfig, nebulaLog)
	if err != nil {
		log.Fatal(fmt.Sprintf("fail to initialize the connection pool, host: %s, port: %d, %s", address, port, err.Error()))
	}

	var sessionList []*Session

	// Take all idle connection and try to create a new session
	for i := 0; i < multiHostsConfig.MaxConnPoolSize; i++ {
		session, err := pool.GetSession(username, password)
		if err != nil {
			t.Errorf("fail to create a new session from connection pool, %s", err.Error())
		}
		sessionList = append(sessionList, session)
	}

	assert.Equal(t, 0, pool.idleConnectionQueue.Len())
	assert.Equal(t, 3, pool.activeConnectionQueue.Len())

	_, err = pool.GetSession(username, password)
	assert.EqualError(t, err, "failed to get connection: No valid connection in the idle queue and connection number has reached the pool capacity")

	// Release 1 connection back to pool
	sessionToRelease := sessionList[0]
	sessionToRelease.Release()
	sessionList = sessionList[1:]
	assert.Equal(t, 1, pool.idleConnectionQueue.Len())
	assert.Equal(t, 2, pool.activeConnectionQueue.Len())
	// Try again to get connection
	newSession, err := pool.GetSession(username, password)
	if err != nil {
		t.Errorf("fail to create a new session, %s", err.Error())
	}
	assert.Equal(t, 0, pool.idleConnectionQueue.Len())
	assert.Equal(t, 3, pool.activeConnectionQueue.Len())

	resp, err := tryToExecute(newSession, "SHOW HOSTS;")
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	checkResultSet(t, "show hosts", resp)

	// Try to get more session when the pool is full
	_, err = pool.GetSession(username, password)
	assert.EqualError(t, err, "failed to get connection: No valid connection in the idle queue and connection number has reached the pool capacity")

	for i := 0; i < len(sessionList); i++ {
		sessionList[i].Release()
	}
}

func TestMultiThreads(t *testing.T) {
	hostList := poolAddress

	testPoolConfig := PoolConfig{
		TimeOut:         0 * time.Millisecond,
		IdleTime:        0 * time.Millisecond,
		MaxConnPoolSize: 666,
		MinConnPoolSize: 1,
	}

	// Initialize connection pool
	pool, err := NewConnectionPool(hostList, testPoolConfig, nebulaLog)
	if err != nil {
		log.Fatal(fmt.Sprintf("fail to initialize the connection pool, host: %s, port: %d, %s",
			address, port, err.Error()))
	}
	defer pool.Close()

	var sessionList []*Session

	// Create multiple session
	var wg sync.WaitGroup
	sessCh := make(chan *Session)
	done := make(chan bool)
	wg.Add(testPoolConfig.MaxConnPoolSize)
	for i := 0; i < testPoolConfig.MaxConnPoolSize; i++ {
		go func(sessCh chan<- *Session, wg *sync.WaitGroup) {
			defer wg.Done()
			session, err := pool.GetSession(username, password)
			if err != nil {
				t.Errorf("fail to create a new session from connection pool, %s", err.Error())
			}
			sessCh <- session
		}(sessCh, &wg)

	}
	go func(sessCh <-chan *Session) {
		for session := range sessCh {
			sessionList = append(sessionList, session)
		}
		done <- true
	}(sessCh)
	wg.Wait()
	close(sessCh)
	<-done

	assert.Equal(t, 666, pool.getActiveConnCount(), "Total number of active connections should be 666")
	assert.Equal(t, 666, len(sessionList), "Total number of sessions should be 666")

	for i := 0; i < testPoolConfig.MaxConnPoolSize; i++ {
		sessionList[i].Release()
	}
	assert.Equal(t, 666, pool.getIdleConnCount(), "Total number of idle connections should be 666")
}

func TestLoadBalancer(t *testing.T) {
	hostList := poolAddress
	var loadPerHost = make(map[HostAddress]int)
	testPoolConfig := PoolConfig{
		TimeOut:         0 * time.Millisecond,
		IdleTime:        0 * time.Millisecond,
		MaxConnPoolSize: 999,
		MinConnPoolSize: 0,
	}

	// Initialize connection pool
	pool, err := NewConnectionPool(hostList, testPoolConfig, nebulaLog)
	if err != nil {
		t.Fatalf("fail to initialize the connection pool, host: %s, port: %d, %s", address, port, err.Error())
	}
	defer pool.Close()

	var sessionList []*Session

	// Create multiple sessions
	for i := 0; i < 999; i++ {
		session, err := pool.GetSession(username, password)
		if err != nil {
			t.Errorf("fail to create a new session from connection pool, %s", err.Error())
		}
		loadPerHost[session.connection.severAddress]++
		sessionList = append(sessionList, session)
	}
	assert.Equal(t, 999, len(sessionList), "Total number of sessions should be 666")

	for _, v := range loadPerHost {
		assert.Equal(t, 333, v, "Total number of sessions should be 333")
	}
	for i := 0; i < len(sessionList); i++ {
		sessionList[i].Release()
	}
}

func TestIdleTimeoutCleaner(t *testing.T) {
	hostList := poolAddress

	idleTimeoutConfig := PoolConfig{
		TimeOut:         0 * time.Millisecond,
		IdleTime:        2 * time.Second,
		MaxConnPoolSize: 30,
		MinConnPoolSize: 6,
	}

	// Initialize connection pool
	pool, err := NewConnectionPool(hostList, idleTimeoutConfig, nebulaLog)
	if err != nil {
		t.Fatalf("fail to initialize the connection pool, host: %s, port: %d, %s", address, port, err.Error())
	}
	defer pool.Close()

	var sessionList []*Session

	// Create session
	for i := 0; i < idleTimeoutConfig.MaxConnPoolSize; i++ {
		session, err := pool.GetSession(username, password)
		if err != nil {
			t.Errorf("fail to create a new session from connection pool, %s", err.Error())
		}
		sessionList = append(sessionList, session)
	}

	for i := range sessionList {
		_, err := sessionList[i].Execute("SHOW HOSTS;")
		if err != nil {
			t.Errorf("Error info: %s", err.Error())
			return
		}
		sessionList[i].Release()
	}

	time.Sleep(idleTimeoutConfig.IdleTime)
	pool.cleanerChan <- struct{}{} // The minimum interval for cleanup is 1 minute, so in CI we need to trigger cleanup manually
	time.Sleep(idleTimeoutConfig.IdleTime)

	pool.rwLock.RLock()
	assert.Equal(t, idleTimeoutConfig.MinConnPoolSize, pool.idleConnectionQueue.Len())
	assert.Equal(t, 0, pool.activeConnectionQueue.Len())
	pool.rwLock.RUnlock()
}

func TestTimeout(t *testing.T) {
	hostAddress := HostAddress{Host: address, Port: port}
	hostList := []HostAddress{}
	hostList = append(hostList, hostAddress)

	testPoolConfig = PoolConfig{
		TimeOut:         1000 * time.Millisecond,
		IdleTime:        0 * time.Millisecond,
		MaxConnPoolSize: 10,
		MinConnPoolSize: 1,
	}

	// Initialize connection pool
	pool, err := NewConnectionPool(hostList, testPoolConfig, nebulaLog)
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
	assert.NotEmptyf(t, session, "session is nil")

	// Create schemas
	{
		createSchema :=
			"CREATE SPACE IF NOT EXISTS test_timeout(VID_TYPE=FIXED_STRING(32));" +
				"USE test_timeout;" +
				"CREATE TAG IF NOT EXISTS person (name string, age int);" +
				"CREATE EDGE IF NOT EXISTS like(likeness int);"
		resultSet, err := tryToExecute(session, createSchema)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.True(t, resultSet.IsSucceed())
	}
	time.Sleep(3 * time.Second)

	// Load data
	{
		query := "INSERT VERTEX person (name, age) VALUES" +
			"'A':('A', 10), " +
			"'B':('B', 10), " +
			"'C':('C', 10), " +
			"'D':('D', 10)"
		resultSet, err := tryToExecute(session, query)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.Truef(t, resultSet.IsSucceed(), resultSet.GetErrorMsg())
		query =
			"INSERT EDGE like(likeness) VALUES " +
				"'A'->'B':(80), " +
				"'B'->'C':(70), " +
				"'C'->'D':(84), " +
				"'D'->'A':(68)"
		resultSet, err = tryToExecute(session, query)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.Truef(t, resultSet.IsSucceed(), resultSet.GetErrorMsg())
	}

	// trigger timeout
	_, err = session.Execute("GO 10000 STEPS FROM 'A' OVER * YIELD like.likeness")
	assert.Contains(t, err.Error(), "timeout")

	resultSet, err := tryToExecute(session, "YIELD 999")
	assert.Empty(t, err)
	assert.True(t, resultSet.IsSucceed())
	assert.Contains(t, resultSet.AsStringTable(), []string{"999"})

	// Drop space
	err = dropSpace("client_test")
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestExecuteJson(t *testing.T) {
	hostList := []HostAddress{{Host: address, Port: port}}

	testPoolConfig = PoolConfig{
		TimeOut:         0 * time.Millisecond,
		IdleTime:        0 * time.Millisecond,
		MaxConnPoolSize: 10,
		MinConnPoolSize: 1,
	}

	// Initialize connection pool
	pool, err := NewConnectionPool(hostList, testPoolConfig, nebulaLog)
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

	// Create schemas
	createTestDataSchema(t, session)
	// Load data
	loadTestData(t, session)

	// Simple query
	{
		jsonStrResult, err := session.ExecuteJson(`YIELD 1, 2.2, "hello", [1,2,"abc"], {key: "value"}, "汉字"`)
		if err != nil {
			t.Fatalf("fail to get the result in json format, %s", err.Error())
		}
		var jsonObj map[string]any
		exp := []any{
			float64(1), float64(2.2), "hello",
			[]any{float64(1), float64(2), "abc"},
			map[string]any{"key": "value"},
			"汉字"}

		// Parse JSON
		json.Unmarshal(jsonStrResult, &jsonObj)

		// Get errorcode
		errorCode := float64(0)
		respErrorCode := jsonObj["errors"].([]any)[0].(map[string]any)["code"]
		assert.Equal(t, errorCode, respErrorCode)

		// Get data
		rowData := jsonObj["results"].([]any)[0].(map[string]any)["data"].([]any)[0].(map[string]any)["row"]
		assert.Equal(t, exp, rowData)

		// Get space name
		respSpace := jsonObj["results"].([]any)[0].(map[string]any)["spaceName"]
		assert.Equal(t, "test_data", respSpace)

	}

	// Complex result
	{
		jsonStrResult, err := session.ExecuteJson("MATCH (v:person {name: \"Bob\"}) RETURN v")
		if err != nil {
			t.Fatalf("fail to get the result in json format, %s", err.Error())
		}
		var jsonObj map[string]any
		exp := []any{
			map[string]any{
				"person.age":            float64(10),
				"person.birthday":       `2010-09-10T02:08:02.000000000Z`,
				"person.book_num":       float64(100),
				"person.child_name":     "Hello Worl",
				"person.expend":         float64(100),
				"person.first_out_city": float64(1111),
				"person.friends":        float64(10),
				"person.grade":          float64(3),
				"person.hobby":          nil,
				"person.is_girl":        false,
				"person.morning":        `23:10:00.000000000Z`,
				"person.name":           "Bob",
				"person.property":       float64(1000),
				"person.start_school":   `2017-09-10`,
				"student.name":          "Bob",
				"student.interval":      `P1MT100.000020000S`,
			},
		}

		// Parse JSON
		json.Unmarshal(jsonStrResult, &jsonObj)
		rowData := jsonObj["results"].([]any)[0].(map[string]any)["data"].([]any)[0].(map[string]any)["row"]
		assert.Equal(t, exp, rowData)
	}

	// Error test
	{
		jsonStrResult, err := session.ExecuteJson("MATCH (v:invalidTag {name: \"Bob\"}) RETURN v")
		if err != nil {
			t.Fatalf("fail to get the result in json format, %s", err.Error())
		}
		var jsonObj map[string]any

		// Parse JSON
		json.Unmarshal(jsonStrResult, &jsonObj)

		errorCode := float64(-1009)
		respErrorCode := jsonObj["errors"].([]any)[0].(map[string]any)["code"]
		assert.Equal(t, errorCode, respErrorCode)

		errorMsg := "SemanticError: `invalidTag': Unknown tag"
		respErrorMsg := jsonObj["errors"].([]any)[0].(map[string]any)["message"]
		assert.Equal(t, errorMsg, respErrorMsg)
	}
}

func TestExecuteWithParameter(t *testing.T) {
	hostList := []HostAddress{{Host: address, Port: port}}

	testPoolConfig = PoolConfig{
		TimeOut:         0 * time.Millisecond,
		IdleTime:        0 * time.Millisecond,
		MaxConnPoolSize: 10,
		MinConnPoolSize: 1,
	}

	// Initialize connection pool
	pool, err := NewConnectionPool(hostList, testPoolConfig, nebulaLog)
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

	// Create schemas
	createTestDataSchema(t, session)
	// Load data
	loadTestData(t, session)

	// Update the params map
	params := make(map[string]any)
	params["p1"] = true
	params["p2"] = 3
	params["p3"] = []any{true, 3}
	params["p4"] = map[string]any{"a": true, "b": "Bob"}
	params["p5"] = int64(9223372036854775807)  // Max int64
	params["p6"] = int64(-9223372036854775808) // Min int64
	params["p7"] = int64(42)                   // Normal int64 value

	// Simple result
	{
		query := `RETURN
			toBoolean($p1) and false,
			$p2+3,
			$p3[1]>3,
			$p5,
			$p6,
			$p7,
			$p5 + 1 AS overflow_add,
			$p6 - 1 AS overflow_subtract,
			$p7 * 2 AS normal_multiply`
		resp, err := tryToExecuteWithParameter(session, query, params)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.Equal(t, 1, resp.GetRowSize())
		record, err := resp.GetRowValuesByIndex(0)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}

		// Test existing cases
		// col0 toBoolean($p1) and false, p1 = true
		valWrap, err := record.GetValueByIndex(0)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		col1, err := valWrap.AsBool()
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.Equal(t, false, col1)

		// col1 $p2+3, p2 = 3
		valWrap, err = record.GetValueByIndex(1)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		col2, err := valWrap.AsInt()
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.Equal(t, int64(6), col2)

		// col2 $p3[1]>3, p3 = [true,3]
		valWrap, err = record.GetValueByIndex(2)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		col3, err := valWrap.AsBool()
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.Equal(t, false, col3)

		// Test int64 cases
		// Max int64
		valWrap, err = record.GetValueByIndex(3)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.False(t, valWrap.IsNull())
		maxInt64, err := valWrap.AsInt()
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.Equal(t, int64(9223372036854775807), maxInt64)

		// Min int64
		valWrap, err = record.GetValueByIndex(4)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.False(t, valWrap.IsNull())
		minInt64, err := valWrap.AsInt()
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.Equal(t, int64(-9223372036854775808), minInt64)

		// Normal int64
		valWrap, err = record.GetValueByIndex(5)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.False(t, valWrap.IsNull())
		normalInt64, err := valWrap.AsInt()
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.Equal(t, int64(42), normalInt64)

		// Overflow addition
		valWrap, err = record.GetValueByIndex(6)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.True(t, valWrap.IsNull(), "Overflow addition should result in null")

		// Overflow subtraction
		valWrap, err = record.GetValueByIndex(7)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.True(t, valWrap.IsNull(), "Overflow subtraction should result in null")

		// Normal multiplication
		valWrap, err = record.GetValueByIndex(8)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.False(t, valWrap.IsNull())
		normalMultiply, err := valWrap.AsInt()
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.Equal(t, int64(84), normalMultiply)
	}
	// Complex result
	{
		query := "MATCH (v:person {name: $p4.b}) WHERE v.person.age>$p2-3 and $p1==true RETURN v ORDER BY $p3[0] LIMIT $p2"
		resp, err := tryToExecuteWithParameter(session, query, params)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		checkResultSet(t, query, resp)

		assert.Equal(t, 1, resp.GetRowSize())
		record, err := resp.GetRowValuesByIndex(0)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		valWrap, err := record.GetValueByIndex(0)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		node, err := valWrap.AsNode()
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		assert.Equal(t,
			"(\"Bob\" :student{interval: P1MT100.000020000S, name: \"Bob\"} "+
				":person{age: 10, birthday: 2010-09-10T10:08:02.000000, book_num: 100, "+
				"child_name: \"Hello Worl\", expend: 100.0, "+
				"first_out_city: 1111, friends: 10, grade: 3, "+
				"hobby: __NULL__, is_girl: false, "+
				"morning: 07:10:00.000000, name: \"Bob\", "+
				"property: 1000.0, start_school: 2017-09-10})",
			node.String())
	}
}

func TestReconnect(t *testing.T) {
	hostList := poolAddress

	timeoutConfig := PoolConfig{
		TimeOut:         0 * time.Millisecond,
		IdleTime:        0 * time.Millisecond,
		MaxConnPoolSize: 10,
		MinConnPoolSize: 6,
	}

	// Initialize connection pool
	pool, err := NewConnectionPool(hostList, timeoutConfig, nebulaLog)
	if err != nil {
		t.Fatalf("fail to initialize the connection pool, host: %s, port: %d, %s", address, port, err.Error())
	}
	defer pool.Close()

	// Create session
	session, err := pool.GetSession(username, password)
	if err != nil {
		t.Errorf("fail to create a new session from connection pool, %s", err.Error())
	}
	defer session.Release()

	// Send query to server periodically
	for i := 0; i < timeoutConfig.MaxConnPoolSize; i++ {
		time.Sleep(200 * time.Millisecond)
		if i == 3 {
			stopContainer(t, "nebula-docker-compose_graphd0_1")
		}
		if i == 7 {
			stopContainer(t, "nebula-docker-compose_graphd1_1")
		}
		_, err := session.Execute("SHOW HOSTS;")
		fmt.Println("Sending query...")

		if err != nil {
			t.Errorf("Error info: %s", err.Error())
			return
		}
	}

	resp, err := session.Execute("SHOW HOSTS;")
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	checkResultSet(t, "SHOW HOSTS;", resp)

	startContainer(t, "nebula-docker-compose_graphd0_1")
	startContainer(t, "nebula-docker-compose_graphd1_1")

	// Wait for graphd to be up
	time.Sleep(5 * time.Second)
}

// Method used to check execution response
func checkResultSet(t *testing.T, prefix string, err *ResultSet) {
	t.Helper()
	if !err.IsSucceed() {
		t.Errorf("%s, ErrorCode: %v, ErrorMsg: %s", prefix, err.GetErrorCode(), err.GetErrorMsg())
	}
}

func checkConResp(prefix string, err *graph.ExecutionResponse) {
	if IsError(err) {
		log.Fatalf("%s, ErrorCode: %v, ErrorMsg: %s", prefix, err.ErrorCode, err.ErrorMsg)
	}
}

func stopContainer(t *testing.T, containerName string) {
	cmd := exec.Command("docker", "stop", containerName)
	err := cmd.Run()
	if err != nil {
		t.Fatalf("failed to stop container, name: %s, error code: %s", containerName, err.Error())
	}
}

func startContainer(t *testing.T, containerName string) {
	cmd := exec.Command("docker", "start", containerName)
	err := cmd.Run()
	if err != nil {
		t.Fatalf("failed to start container, name: %s, error code: %s", containerName, err.Error())
	}
}

type executer interface {
	Execute(query string) (*ResultSet, error)
}

func tryToExecute(e executer, query string) (resp *ResultSet, err error) {
	for i := 3; i > 0; i-- {
		resp, err = e.Execute(query)
		if err == nil && resp.IsSucceed() {
			return
		}
		time.Sleep(2 * time.Second)
	}
	return
}

func tryToExecuteWithParameter(session *Session, query string, params map[string]any) (resp *ResultSet, err error) {
	for i := 3; i > 0; i-- {
		resp, err = session.ExecuteWithParameter(query, params)
		if err == nil && resp.IsSucceed() {
			return
		}
		time.Sleep(2 * time.Second)
	}
	return
}

// creates schema
func createTestDataSchema(t *testing.T, executor executer) {
	createSchema := "CREATE SPACE IF NOT EXISTS test_data(vid_type = FIXED_STRING(30));" +
		"USE test_data; " +
		"CREATE TAG IF NOT EXISTS person(name string, age int8, grade int16, " +
		"friends int32, book_num int64, birthday datetime, " +
		"start_school date, morning time, property double, " +
		"is_girl bool, child_name fixed_string(10), expend float, " +
		"first_out_city timestamp, hobby string); " +
		"CREATE TAG IF NOT EXISTS student(name string, interval duration); " +
		"CREATE EDGE IF NOT EXISTS like(likeness double); " +
		"CREATE EDGE IF NOT EXISTS friend(start_Datetime datetime, end_Datetime datetime); " +
		"CREATE TAG INDEX IF NOT EXISTS person_name_index ON person(name(8));"
	resultSet, err := tryToExecute(executor, createSchema)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	checkResultSet(t, createSchema, resultSet)

	time.Sleep(3 * time.Second)
}

// inserts data that used in tests
func loadTestData(t *testing.T, e executer) {
	query := "INSERT VERTEX person(name, age, grade, friends, book_num," +
		"birthday, start_school, morning, property," +
		"is_girl, child_name, expend, first_out_city) VALUES" +
		"'Bob':('Bob', 10, 3, 10, 100, datetime('2010-09-10T10:08:02')," +
		"date('2017-09-10'), time('07:10:00'), " +
		"1000.0, false, \"Hello World!\", 100.0, 1111)," +
		"'Lily':('Lily', 9, 3, 10, 100, datetime('2010-09-10T10:08:02'), " +
		"date('2017-09-10'), time('07:10:00'), " +
		"1000.0, false, \"Hello World!\", 100.0, 1111)," +
		"'Tom':('Tom', 10, 3, 10, 100, datetime('2010-09-10T10:08:02'), " +
		"date('2017-09-10'), time('07:10:00'), " +
		"1000.0, false, \"Hello World!\", 100.0, 1111)," +
		"'Jerry':('Jerry', 9, 3, 10, 100, datetime('2010-09-10T10:08:02')," +
		"date('2017-09-10'), time('07:10:00'), " +
		"1000.0, false, \"Hello World!\", 100.0, 1111), " +
		"'John':('John', 10, 3, 10, 100, datetime('2010-09-10T10:08:02'), " +
		"date('2017-09-10'), time('07:10:00'), " +
		"1000.0, false, \"Hello World!\", 100.0, 1111)"
	resultSet, err := tryToExecute(e, query)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	checkResultSet(t, query, resultSet)

	query =
		"INSERT VERTEX student(name, interval) VALUES " +
			"'Bob':('Bob', duration({months:1, seconds:100, microseconds:20})), 'Lily':('Lily', duration({years: 1, seconds: 0})), " +
			"'Tom':('Tom', duration({years: 1, seconds: 0})), 'Jerry':('Jerry', duration({years: 1, seconds: 0})), 'John':('John', duration({years: 1, seconds: 0}))"
	resultSet, err = tryToExecute(e, query)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	checkResultSet(t, query, resultSet)

	query =
		"INSERT EDGE like(likeness) VALUES " +
			"'Bob'->'Lily':(80.0), " +
			"'Bob'->'Tom':(70.0), " +
			"'Jerry'->'Lily':(84.0)," +
			"'Tom'->'Jerry':(68.3), " +
			"'Bob'->'John':(97.2), " +
			"'Lily'->'Tom':(80.0)"
	resultSet, err = tryToExecute(e, query)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	checkResultSet(t, query, resultSet)

	query =
		"INSERT EDGE friend(start_Datetime, end_Datetime) VALUES " +
			"'Bob'->'Lily':(datetime('2008-09-10T10:08:02'), datetime('2010-09-10T10:08:02')), " +
			"'Bob'->'Tom':(datetime('2008-09-10T10:08:02'), datetime('2010-09-10T10:08:02')), " +
			"'Jerry'->'Lily':(datetime('2008-09-10T10:08:02'), datetime('2010-09-10T10:08:02')), " +
			"'Tom'->'Jerry':(datetime('2008-09-10T10:08:02'), datetime('2010-09-10T10:08:02')), " +
			"'Bob'->'John':(datetime('2008-09-10T10:08:02'), datetime('2010-09-10T10:08:02'))"
	resultSet, err = tryToExecute(e, query)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	checkResultSet(t, query, resultSet)
}

// prepareSpace creates a space for test
func prepareSpace(spaceName string) error {
	hostAddress := HostAddress{Host: address, Port: port}
	conn := newConnection(hostAddress)
	testPoolConfig := GetDefaultConf()

	err := conn.open(hostAddress, testPoolConfig.TimeOut, nil, false, nil, "")
	if err != nil {
		return fmt.Errorf("fail to open connection, address: %s, port: %d, %s", address, port, err.Error())
	}

	authResp, authErr := conn.authenticate(username, password)
	if authErr != nil {
		return fmt.Errorf("fail to authenticate, username: %s, password: %s, %s", username, password, authErr.Error())
	}

	if authResp.GetErrorCode() != nebula.ErrorCode_SUCCEEDED {
		return fmt.Errorf("failed to authenticate, error code: %d, error msg: %s",
			authResp.GetErrorCode(), authResp.GetErrorMsg())
	}

	sessionID := authResp.GetSessionID()

	defer logoutAndClose(conn, sessionID)

	query := fmt.Sprintf("CREATE SPACE IF NOT EXISTS"+
		" %s(partition_num=32, replica_factor=1, vid_type = FIXED_STRING(30));", spaceName)
	resp, err := conn.execute(sessionID, query)
	if err != nil {
		log.Fatalf(err.Error())
	}
	checkConResp(query, resp)
	time.Sleep(5 * time.Second)

	return nil
}

// dropSpace drops a space. The space name should be the same as the one created in prepareSpace
func dropSpace(spaceName string) error {
	hostAddress := HostAddress{Host: address, Port: port}
	conn := newConnection(hostAddress)
	testPoolConfig := GetDefaultConf()

	err := conn.open(hostAddress, testPoolConfig.TimeOut, nil, false, nil, "")
	if err != nil {
		return fmt.Errorf("fail to open connection, address: %s, port: %d, %s", address, port, err.Error())
	}

	authResp, authErr := conn.authenticate(username, password)
	if authErr != nil {
		return fmt.Errorf("fail to authenticate, username: %s, password: %s, %s", username, password, authErr.Error())
	}

	if authResp.GetErrorCode() != nebula.ErrorCode_SUCCEEDED {
		return fmt.Errorf("failed to authenticate, error code: %d, error msg: %s",
			authResp.GetErrorCode(), authResp.GetErrorMsg())
	}

	sessionID := authResp.GetSessionID()

	defer logoutAndClose(conn, sessionID)

	query := fmt.Sprintf("DROP SPACE IF EXISTS %s;", spaceName)
	resp, err := conn.execute(sessionID, query)
	if err != nil {
		return err
	}
	checkConResp(query, resp)
	return nil
}
