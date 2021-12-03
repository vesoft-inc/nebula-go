/* Copyright (c) 2021 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package main

import (
	"fmt"
	"strings"
	"sync"

	nebulago "github.com/vesoft-inc/nebula-go/v2"
	nebula "github.com/vesoft-inc/nebula-go/v2/nebula"
)

const (
	address = "127.0.0.1"
	// The default port of Nebula Graph 2.x is 9669.
	// 3699 is only for testing.
	port     = 1774
	username = "root"
	password = "nebula"
)

// Initialize logger
var log = nebulago.DefaultLogger{}

func main() {
	hostAddress := nebulago.HostAddress{Host: address, Port: port}
	hostList := []nebulago.HostAddress{hostAddress}
	// Create configs for connection pool using default values
	testPoolConfig := nebulago.GetDefaultConf()

	// Initialize connection pool
	pool, err := nebulago.NewConnectionPool(hostList, testPoolConfig, log)
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

		checkResultSet := func(prefix string, res *nebulago.ResultSet) {
			if !res.IsSucceed() {
				log.Fatal(fmt.Sprintf("%s, ErrorCode: %v, ErrorMsg: %s", prefix, res.GetErrorCode(), res.GetErrorMsg()))
			}
		}

		var params map[string]*nebula.Value
		params = make(map[string]*nebula.Value)

		var bVal bool = true
		var iVal int64 = 3
		// bool
		p1 := nebula.Value{BVal: &bVal}
		// int
		p2 := nebula.Value{IVal: &iVal}
		// list
		lSlice := []*nebula.Value{&p1,&p2}
		var lVal nebula.NList
		lVal.Values = lSlice
		p3 := nebula.Value{LVal: &lVal}
		// map
		var nmap map[string]*nebula.Value = map[string]*nebula.Value{"a": &p1, "b": &p2}
		var mVal nebula.NMap
		mVal.Kvs = nmap
		p4 := nebula.Value{MVal: &mVal}

		params["p1"] = &p1
		params["p2"] = &p2
		params["p3"] = &p3
		params["p4"] = &p4


		// Extract data from the resultSet
		{
			query := "RETURN abs($p2)+1 AS col1, toBoolean($p1) and false AS col2, $p3, $p4.a"
			// Send query
			// resultSet, err := session.ExecuteWithParameter(query, params)
			resultSet, err := session.ExecuteWithParameter(query,params)
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			checkResultSet(query, resultSet)

			// Get all column names from the resultSet
			colNames := resultSet.GetColNames()
			fmt.Printf("Column names: %s\n", strings.Join(colNames, ", "))
			fmt.Print(resultSet.AsStringTable())
			// Get a row from resultSet
			record, err := resultSet.GetRowValuesByIndex(0)
			if err != nil {
				log.Error(err.Error())
			}
			// Print whole row
			fmt.Printf("The first row elements: %s\n", record.String())
		}
	}(&wg)
	wg.Wait()

	fmt.Print("\n")
	log.Info("Nebula Go Client Gorountines Example Finished")
}
