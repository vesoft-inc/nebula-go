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
	"strings"
	"sync"
	"time"

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
	// Create session and send query in go routine
	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		// Create session
		session, err := pool.GetSession(username, password)
		if err != nil {
			log.Fatal(fmt.Sprintf("Fail to create a new session from connection pool, username: %s, password: %s, %s",
				username, password, err.Error()))
		}
		// Release session and return connection back to connection pool
		defer session.Release()
		// Method used to check execution response
		checkResultSet := func(prefix string, res *nebula.ResultSet) {
			if !res.IsSucceed() {
				log.Fatal(fmt.Sprintf("%s, ErrorCode: %v, ErrorMsg: %s", prefix, res.GetErrorCode(), res.GetErrorMsg()))
			}
		}
		{
			createSchema := "CREATE SPACE IF NOT EXISTS example_space(vid_type=FIXED_STRING(20)); " +
				"USE example_space;" +
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
		time.Sleep(5 * time.Second)
		{
			insertVertexes := "INSERT VERTEX person(name, age) VALUES " +
				"'Bob':('Bob', 10), " +
				"'Lily':('Lily', 9), " +
				"'Tom':('Tom', 10), " +
				"'Jerry':('Jerry', 13), " +
				"'John':('John', 11);"

			// Insert multiple vertexes
			resultSet, err := session.Execute(insertVertexes)
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

			resultSet, err := session.Execute(insertEdges)
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			checkResultSet(insertEdges, resultSet)
		}
		// Extract data from the resultSet
		{
			query := "GO FROM 'Bob' OVER like YIELD $^.person.name, $^.person.age, like.likeness"
			// Send query
			resultSet, err := session.Execute(query)
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			checkResultSet(query, resultSet)

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
			resultSet, err := session.Execute(query)
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			checkResultSet(query, resultSet)
		}
	}(&wg)
	wg.Wait()

	fmt.Print("\n")
	log.Info("Nebula Go Client Gorountines Example Finished")
}
