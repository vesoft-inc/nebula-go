/*
 *
 * Copyright (c) 2021 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 *
 */

package main

import (
	"fmt"

	nebula "github.com/vesoft-inc/nebula-go/v2"
)

const (
	address = "127.0.0.1"
	// The default port of Nebula Graph 2.x is 9669.
	// 3699 is only for testing.
	port     = 3699
	username = "root"
	password = "nebula"
)

// Initialize logger
var log = nebula.DefaultLogger{}

func main() {
	hostAddress := nebula.HostAddress{Host: address, Port: port}
	hostList := []nebula.HostAddress{hostAddress}
	// Create configs for connection pool using default values
	testPoolConfig := nebula.GetDefaultConf()

	// Initialize connection pool
	pool, err := nebula.NewConnectionPool(hostList, testPoolConfig, log)
	if err != nil {
		log.Fatal(fmt.Sprintf("Fail to initialize the connection pool, host: %s, port: %d, %s", address, port, err.Error()))
	}
	// Close all connections in the pool
	defer pool.Close()

	// Create session
	session, err := pool.GetSession(username, password)
	if err != nil {
		log.Fatal(fmt.Sprintf("Fail to create a new session from connection pool, username: %s, password: %s, %s",
			username, password, err.Error()))
	}
	// Release session and return connection back to connection pool
	defer session.Release()

	checkResultSet := func(prefix string, res *nebula.ResultSet) {
		if !res.IsSucceed() {
			log.Fatal(fmt.Sprintf("%s, ErrorCode: %v, ErrorMsg: %s", prefix, res.GetErrorCode(), res.GetErrorMsg()))
		}
	}

	{
		// Prepare the query
		createSchema := "CREATE SPACE IF NOT EXISTS basic_example_space(vid_type=FIXED_STRING(20)); " +
			"USE basic_example_space;" +
			"CREATE TAG IF NOT EXISTS person(name string, age int);" +
			"CREATE EDGE IF NOT EXISTS like(likeness double)"

		// Excute a query
		resultSet, err := session.Execute(createSchema)
		if err != nil {
			fmt.Print(err.Error())
			return
		}
		checkResultSet(createSchema, resultSet)
	}
	// Drop space
	{
		query := "DROP SPACE IF EXISTS basic_example_space"
		// Send query
		resultSet, err := session.Execute(query)
		if err != nil {
			fmt.Print(err.Error())
			return
		}
		checkResultSet(query, resultSet)
	}

	fmt.Print("\n")
	log.Info("Nebula Go Client Basic Example Finished")
}
