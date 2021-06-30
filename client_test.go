/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package nebula_go

import (
	"fmt"
	"log"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vesoft-inc/nebula-go/v2/nebula"
	"github.com/vesoft-inc/nebula-go/v2/nebula/graph"
)

const (
	address  = "127.0.0.1"
	port     = 3699
	username = "root"
	password = "nebula"
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
	hostAdress := HostAddress{Host: address, Port: port}

	conn := newConnection(hostAdress)
	err := conn.open(hostAdress, testPoolConfig.TimeOut)
	if err != nil {
		t.Fatalf("fail to open connection, address: %s, port: %d, %s", address, port, err.Error())
	}

	authresp, authErr := conn.authenticate(username, password)
	if authErr != nil {
		t.Fatalf("fail to authenticate, username: %s, password: %s, %s", username, password, authErr.Error())
	}

	sessionID := authresp.GetSessionID()

	defer logoutAndClose(conn, sessionID)

	resp, err := conn.execute(sessionID, "SHOW HOSTS;")
	if err != nil {
		t.Fatalf(err.Error())
		return
	}

	checkConResp(t, "show hosts", resp)

	resp, err = conn.execute(sessionID, "CREATE SPACE client_test(partition_num=1024, replica_factor=1, vid_type = FIXED_STRING(30));")
	if err != nil {
		t.Error(err.Error())
		return
	}
	checkConResp(t, "create space", resp)
	resp, err = conn.execute(sessionID, "DROP SPACE client_test;")
	if err != nil {
		t.Error(err.Error())
		return
	}
	checkConResp(t, "drop space", resp)

	res := conn.ping()
	if res != true {
		t.Error("Connectin ping failed")
		return
	}
}

func TestConfigs(t *testing.T) {
	hostAdress := HostAddress{Host: address, Port: port}
	hostList := []HostAddress{}
	hostList = append(hostList, hostAdress)

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
		// Initialize connectin pool
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
		// Excute a query
		resp, err := session.Execute("SHOW HOSTS;")
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		checkResSetResp(t, "show hosts", resp)
		// Create a new space
		resp, err = session.Execute("CREATE SPACE client_test(partition_num=1024, replica_factor=1, vid_type = FIXED_STRING(30));")
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		checkResSetResp(t, "create space", resp)

		resp, err = session.Execute("DROP SPACE client_test;")
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		checkResSetResp(t, "drop space", resp)
	}
}

func TestAuthentication(t *testing.T) {
	const (
		address  = "127.0.0.1"
		port     = 3699
		username = "dummy"
		password = "nebula"
	)

	hostAdress := HostAddress{Host: address, Port: port}

	conn := newConnection(hostAdress)
	err := conn.open(hostAdress, testPoolConfig.TimeOut)
	if err != nil {
		t.Fatalf("fail to open connection, address: %s, port: %d, %s", address, port, err.Error())
	}
	defer conn.close()

	_, authErr := conn.authenticate(username, password)
	assert.EqualError(t, authErr, "fail to authenticate, error: Bad username/password")
}

func TestInvalidHostTimeout(t *testing.T) {
	hostList := []HostAddress{
		{Host: "192.168.100.125", Port: 3699}, // Invalid host
		{Host: "127.0.0.1", Port: 3699},
	}

	// Initialize connectin pool
	pool, err := NewConnectionPool(hostList, testPoolConfig, nebulaLog)
	if err != nil {
		t.Fatalf("fail to initialize the connection pool, host: %s, port: %d, %s", address, port, err.Error())
	}
	// close all connections in the pool
	defer pool.Close()
	err = pool.Ping(hostList[0], 1000*time.Millisecond)
	assert.EqualError(t, err, "failed to open transport, error: dial tcp 192.168.100.125:3699: i/o timeout")
	err = pool.Ping(hostList[1], 1000*time.Millisecond)
	if err != nil {
		t.Error("failed to ping 127.0.0.1")
	}
}

func TestServiceDataIO(t *testing.T) {
	hostAdress := HostAddress{Host: address, Port: port}
	hostList := []HostAddress{}
	hostList = append(hostList, hostAdress)

	testPoolConfig = PoolConfig{
		TimeOut:         0 * time.Millisecond,
		IdleTime:        0 * time.Millisecond,
		MaxConnPoolSize: 10,
		MinConnPoolSize: 1,
	}

	// Initialize connectin pool
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

	// Method used to check execution response
	checkResultSet := func(prefix string, res *ResultSet) {
		if !res.IsSucceed() {
			t.Fatalf("%s, ErrorCode: %v, ErrorMsg: %s", prefix, res.GetErrorCode(), res.GetErrorMsg())
		}
	}
	// Create schemas
	{
		createSchema := "CREATE SPACE IF NOT EXISTS test_data(vid_type = FIXED_STRING(30));" +
			"USE test_data; " +
			"CREATE TAG IF NOT EXISTS person(name string, age int8, grade int16, " +
			"friends int32, book_num int64, birthday datetime, " +
			"start_school date, morning time, property double, " +
			"is_girl bool, child_name fixed_string(10), expend float, " +
			"first_out_city timestamp, hobby string); " +
			"CREATE TAG IF NOT EXISTS student(name string); " +
			"CREATE EDGE IF NOT EXISTS like(likeness double); " +
			"CREATE EDGE IF NOT EXISTS friend(start_year int, end_year int); " +
			"CREATE TAG INDEX IF NOT EXISTS person_name_index ON person(name(8));"
		resultSet, err := session.Execute(createSchema)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		checkResultSet(createSchema, resultSet)
	}
	time.Sleep(10 * time.Second)

	// Load data
	{
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
		resultSet, err := session.Execute(query)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		checkResultSet(query, resultSet)

		query =
			"INSERT VERTEX student(name) VALUES " +
				"'Bob':('Bob'), 'Lily':('Lily'), " +
				"'Tom':('Tom'), 'Jerry':('Jerry'), 'John':('John')"
		resultSet, err = session.Execute(query)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		checkResultSet(query, resultSet)

		query =
			"INSERT EDGE like(likeness) VALUES " +
				"'Bob'->'Lily':(80.0), " +
				"'Bob'->'Tom':(70.0), " +
				"'Jerry'->'Lily':(84.0)," +
				"'Tom'->'Jerry':(68.3), " +
				"'Bob'->'John':(97.2)"
		resultSet, err = session.Execute(query)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		checkResultSet(query, resultSet)

		query =
			"INSERT EDGE friend(start_year, end_year) VALUES " +
				"'Bob'->'Lily':(2018, 2020), " +
				"'Bob'->'Tom':(2018, 2020), " +
				"'Jerry'->'Lily':(2018, 2020)," +
				"'Tom'->'Jerry':(2018, 2020), " +
				"'Bob'->'John':(2018, 2020)"
		resultSet, err = session.Execute(query)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		checkResultSet(query, resultSet)
	}

	// test base type
	{
		query :=
			"FETCH PROP ON person \"Bob\" YIELD person.name, person.age, person.grade," +
				"person.friends, person.book_num, person.birthday, " +
				"person.start_school, person.morning, " +
				"person.property, person.is_girl, person.child_name, " +
				"person.expend, person.first_out_city, person.hobby"
		resp, err := session.Execute(query)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
		checkResultSet(query, resp)

		assert.Equal(t, resp.GetLatency() > 0, true)
		assert.Equal(t, "", resp.GetComment(), true)
		assert.Equal(t, "test_data", resp.GetSpaceName())
		assert.Equal(t, !resp.IsEmpty(), true)
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
		assert.Equal(t, true, valWrap.IsDate())
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
		assert.Equal(t, true, valWrap.IsTime())
		assert.Equal(t, "07:10:00.000000", valWrap.String())

		UTCTime := timeWrapper.getRawTime()
		expected := nebula.Time{23, 10, 0, 0}
		assert.Equal(t, expected, *UTCTime)
	}
	// test node
	{
		resp, err := session.Execute("MATCH (v:person {name: \"Bob\"}) RETURN v")
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
			"(\"Bob\" :student{name: \"Bob\"} "+
				":person{age: 10, birthday: 2010-09-10T10:08:02.000000, book_num: 100, "+
				"child_name: \"Hello Worl\", expend: 100.0, "+
				"first_out_city: 1111, friends: 10, grade: 3, "+
				"hobby: __NULL__, is_girl: false, "+
				"morning: 07:10:00.000000, name: \"Bob\", "+
				"property: 1000.0, start_school: 2017-09-10})",
			node.String())
	}
	// Drop space
	{
		query := "DROP SPACE test_data;"
		_, err := session.Execute(query)
		if err != nil {
			t.Fatalf(err.Error())
			return
		}
	}
}

func TestPool_SingleHost(t *testing.T) {
	hostAdress := HostAddress{Host: address, Port: port}
	hostList := []HostAddress{}
	hostList = append(hostList, hostAdress)

	testPoolConfig = PoolConfig{
		TimeOut:         0 * time.Millisecond,
		IdleTime:        0 * time.Millisecond,
		MaxConnPoolSize: 10,
		MinConnPoolSize: 1,
	}

	// Initialize connectin pool
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
	// Excute a query
	resp, err := session.Execute("SHOW HOSTS;")
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	checkResSetResp(t, "show hosts", resp)
	// Create a new space
	resp, err = session.Execute("CREATE SPACE client_test(partition_num=1024, replica_factor=1, vid_type = FIXED_STRING(30));")
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	checkResSetResp(t, "create space", resp)

	resp, err = session.Execute("DROP SPACE client_test;")
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	checkResSetResp(t, "drop space", resp)
}

func TestPool_MultiHosts(t *testing.T) {
	hostList := poolAddress
	// Minimun pool size < hosts number
	multiHostsConfig := PoolConfig{
		TimeOut:         0 * time.Millisecond,
		IdleTime:        0 * time.Millisecond,
		MaxConnPoolSize: 3,
		MinConnPoolSize: 1,
	}

	// Initialize connectin pool
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

	// Release 1 connectin back to pool
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

	resp, err := newSession.Execute("SHOW HOSTS;")
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	checkResSetResp(t, "show hosts", resp)

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

	// Initialize connectin pool
	pool, err := NewConnectionPool(hostList, testPoolConfig, nebulaLog)
	if err != nil {
		log.Fatal(fmt.Sprintf("fail to initialize the connection pool, host: %s, port: %d, %s", address, port, err.Error()))
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

	// for i := 0; i < len(hostList); i++ {
	// 	assert.Equal(t, 222, pool.GetServerWorkload(i))
	// }
	for i := 0; i < testPoolConfig.MaxConnPoolSize; i++ {
		sessionList[i].Release()
	}
	assert.Equal(t, 666, pool.getIdleConnCount(), "Total number of idle connections should be 666")
}

func TestLoadbalancer(t *testing.T) {
	hostList := poolAddress
	var loadPerHost = make(map[HostAddress]int)
	testPoolConfig := PoolConfig{
		TimeOut:         0 * time.Millisecond,
		IdleTime:        0 * time.Millisecond,
		MaxConnPoolSize: 999,
		MinConnPoolSize: 0,
	}

	// Initialize connectin pool
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
	assert.Equal(t, len(sessionList), 999, "Total number of sessions should be 666")

	for _, v := range loadPerHost {
		assert.Equal(t, v, 333, "Total number of sessions should be 333")
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

func TestReconnect(t *testing.T) {
	hostList := poolAddress

	timeoutConfig := PoolConfig{
		TimeOut:         0 * time.Millisecond,
		IdleTime:        0 * time.Millisecond,
		MaxConnPoolSize: 10,
		MinConnPoolSize: 6,
	}

	// Initialize connectin pool
	pool, err := NewConnectionPool(hostList, timeoutConfig, nebulaLog)
	if err != nil {
		t.Fatalf("fail to initialize the connection pool, host: %s, port: %d, %s", address, port, err.Error())
	}

	var sessionList []*Session

	// Create session
	for i := 0; i < 3; i++ {
		session, err := pool.GetSession(username, password)
		if err != nil {
			t.Errorf("fail to create a new session from connection pool, %s", err.Error())
		}
		sessionList = append(sessionList, session)
	}

	// Send query to server periodically
	for i := 0; i < timeoutConfig.MaxConnPoolSize; i++ {
		time.Sleep(200 * time.Millisecond)
		if i == 3 {
			stopContainer(t, "nebula-docker-compose_graphd_1")
		}
		if i == 7 {
			stopContainer(t, "nebula-docker-compose_graphd1_1")
		}
		_, err := sessionList[0].Execute("SHOW HOSTS;")
		fmt.Println("Sending query...")

		if err != nil {
			t.Errorf("Error info: %s", err.Error())
			return
		}
	}

	resp, err := sessionList[0].Execute("SHOW HOSTS;")
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	checkResSetResp(t, "SHOW HOSTS;", resp)

	startContainer(t, "nebula-docker-compose_graphd_1")
	startContainer(t, "nebula-docker-compose_graphd1_1")

	for i := 0; i < len(sessionList); i++ {
		sessionList[i].Release()
	}
	pool.Close()
}

func TestIpLookup(t *testing.T) {
	hostAddress := HostAddress{Host: "192.168.10.105", Port: 3699}
	hostList := []HostAddress{hostAddress}
	_, err := DomainToIP(hostList)
	if err != nil {
		t.Errorf(err.Error())
	}
}

// Method used to check execution response
func checkResSetResp(t *testing.T, prefix string, err *ResultSet) {
	if !err.IsSucceed() {
		t.Errorf("%s, ErrorCode: %v, ErrorMsg: %s", prefix, err.GetErrorCode(), err.GetErrorMsg())
	}
}
func checkConResp(t *testing.T, prefix string, err *graph.ExecutionResponse) {
	if IsError(err) {
		t.Errorf("%s, ErrorCode: %v, ErrorMsg: %s", prefix, err.ErrorCode, err.ErrorMsg)
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
