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

	nebulago "github.com/vesoft-inc/nebula-go/v3"
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
var log = nebulago.DefaultLogger{}

func main() {
	hostAddress := nebulago.HostAddress{Host: address, Port: port}
	hostList := []nebulago.HostAddress{hostAddress}
	// Create configs for connection pool using default values
	testPoolConfig := nebulago.GetDefaultConf()
	testPoolConfig.UseHTTP2 = useHTTP2

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

		params := make(map[string]interface{})
		params["p1"] = true
		params["p2"] = 3
		params["p3"] = []interface{}{true, 3}
		params["p4"] = map[string]interface{}{"a": true, "b": 3}
		params["p5"] = int64(9223372036854775807)

		// Extract data from the resultSet
		{
			query := "RETURN abs($p2)+1 AS col1, toBoolean($p1) and false AS col2, $p3, $p4.a, $p5 AS col5"
			// Send query
			// resultSet, err := session.ExecuteWithParameter(query, params)
			resultSet, err := session.ExecuteWithParameter(query, params)
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

			// Specifically check the int64 value
			valWrap, err := record.GetValueByIndex(4) // Column 'col5'
			if err != nil {
				log.Error(err.Error())
			} else {
				int64Val, err := valWrap.AsInt()
				if err != nil {
					log.Error(err.Error())
				} else {
					fmt.Printf("The int64 value (col5): %d\n", int64Val)
				}
			}
		}
	}(&wg)
	wg.Wait()

	fmt.Print("\n")
	log.Info("Nebula Go Parameter Example Finished")
}
