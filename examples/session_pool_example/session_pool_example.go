/*
 *
 * Copyright (c) 2022 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 *
 */

package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	nebula "github.com/vesoft-inc/nebula-go/v3"
)

const (
	address = "127.0.0.1"
	// The default port of NebulaGraph 2.x is 9669.
	// 3699 is only for testing.
	port     = 3699
	username = "root"
	password = "nebula"
	useHTTP2 = false
)

// Initialize logger
var log = nebula.DefaultLogger{}

func main() {
	prepareSpace()
	hostAddress := nebula.HostAddress{Host: address, Port: port}

	// Create configs for session pool
	config, err := nebula.NewSessionPoolConf(
		"root",
		"nebula",
		[]nebula.HostAddress{hostAddress},
		"example_space",
		nebula.WithHTTP2(useHTTP2),
	)
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to create session pool config, %s", err.Error()))
	}

	// create session pool
	sessionPool, err := nebula.NewSessionPool(*config, nebula.DefaultLogger{})
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to initialize session pool, %s", err.Error()))
	}
	defer sessionPool.Close()

	checkResultSet := func(prefix string, res *nebula.ResultSet) {
		if !res.IsSucceed() {
			log.Fatal(fmt.Sprintf("%s, ErrorCode: %v, ErrorMsg: %s", prefix, res.GetErrorCode(), res.GetErrorMsg()))
		}
	}

	// execute query
	{
		insertVertexes := "INSERT VERTEX person(name, age) VALUES " +
			"'Bob':('Bob', 10), " +
			"'Lily':('Lily', 9), " +
			"'Tom':('Tom', 10), " +
			"'Jerry':('Jerry', 13), " +
			"'John':('John', 11);"

		// Insert multiple vertexes
		resultSet, err := sessionPool.Execute(insertVertexes)
		if err != nil {
			fmt.Print(err.Error())
			return
		}
		checkResultSet(insertVertexes, resultSet)
	}
	{
		// Insert multiple edges
		insertEdges := "INSERT EDGE like(likeness) VALUES " +
			"'Bob'->'Lily':(80.0), " +
			"'Bob'->'Tom':(70.0), " +
			"'Lily'->'Jerry':(84.0), " +
			"'Tom'->'Jerry':(68.3), " +
			"'Bob'->'John':(97.2);"

		resultSet, err := sessionPool.Execute(insertEdges)
		if err != nil {
			fmt.Print(err.Error())
			return
		}
		checkResultSet(insertEdges, resultSet)
	}
	// Extract data from the resultSet
	{
		query := "GO FROM 'Bob' OVER like YIELD $^.person.name, $^.person.age, like.likeness"
		// Send query in goroutine
		wg := sync.WaitGroup{}
		wg.Add(1)
		var resultSet *nebula.ResultSet
		go func(wg *sync.WaitGroup) {
			defer wg.Done()

			resultSet, err = sessionPool.Execute(query)
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			checkResultSet(query, resultSet)
		}(&wg)
		wg.Wait()

		// Get all column names from the resultSet
		colNames := resultSet.GetColNames()
		fmt.Printf("column names: %s\n", strings.Join(colNames, ", "))

		// Get a row from resultSet
		record, err := resultSet.GetRowValuesByIndex(0)
		if err != nil {
			log.Error(err.Error())
		}
		// Print whole row
		fmt.Printf("row elements: %s\n", record.String())
		// Get a value in the row by column index
		valueWrapper, err := record.GetValueByIndex(0)
		if err != nil {
			log.Error(err.Error())
		}
		// Get type of the value
		fmt.Printf("valueWrapper type: %s \n", valueWrapper.GetType())
		// Check if valueWrapper is a string type
		if valueWrapper.IsString() {
			// Convert valueWrapper to a string value
			v1Str, err := valueWrapper.AsString()
			if err != nil {
				log.Error(err.Error())
			}
			fmt.Printf("Result of ValueWrapper.AsString(): %s\n", v1Str)
		}
		// Print ValueWrapper using String()
		fmt.Printf("Print using ValueWrapper.String(): %s", valueWrapper.String())
	}
	// Drop space
	{
		query := "DROP SPACE IF EXISTS example_space"
		// Send query
		resultSet, err := sessionPool.Execute(query)
		if err != nil {
			fmt.Print(err.Error())
			return
		}
		checkResultSet(query, resultSet)
	}
	fmt.Print("\n")
	log.Info("Nebula Go Client Session Pool Example Finished")
}

// Just a helper function to create a space for this example to run.
func prepareSpace() {
	hostAddress := nebula.HostAddress{Host: address, Port: port}
	hostList := []nebula.HostAddress{hostAddress}
	// Create configs for connection pool using default values
	testPoolConfig := nebula.GetDefaultConf()
	testPoolConfig.UseHTTP2 = useHTTP2

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
		createSchema := "CREATE SPACE IF NOT EXISTS example_space(vid_type=FIXED_STRING(20)); " +
			"USE example_space;" +
			"CREATE TAG IF NOT EXISTS person(name string, age int);" +
			"CREATE EDGE IF NOT EXISTS like(likeness double)"

		// Execute a query
		resultSet, err := session.Execute(createSchema)
		if err != nil {
			fmt.Print(err.Error())
			return
		}
		checkResultSet(createSchema, resultSet)
	}
	time.Sleep(5 * time.Second)
	log.Info("Space example_space was created")
}
