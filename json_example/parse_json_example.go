/* Copyright (c) 2021 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package main

import (
	"encoding/json"
	"fmt"
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

// Struct used for storing the parsed object
type JsonObj struct {
	Results []struct {
		Columns []string `json:"columns"`
		Data    []struct {
			Row  []interface{} `json:"row"`
			Meta []interface{} `json:"meta"`
		} `json:"data"`
		LatencyInUs int    `json:"latencyInUs"`
		SpaceName   string `json:"spaceName"`
		PlanDesc    struct {
			PlanNodeDescs []struct {
				Name        string `json:"name"`
				ID          int    `json:"id"`
				OutputVar   string `json:"outputVar"`
				Description struct {
					Key string `json:"key"`
				} `json:"description"`
				Profiles []struct {
					Rows              int `json:"rows"`
					ExecDurationInUs  int `json:"execDurationInUs"`
					TotalDurationInUs int `json:"totalDurationInUs"`
					OtherStats        struct {
					} `json:"otherStats"`
				} `json:"profiles"`
				BranchInfo struct {
					IsDoBranch      bool `json:"isDoBranch"`
					ConditionNodeID int  `json:"conditionNodeId"`
				} `json:"branchInfo"`
				Dependencies []interface{} `json:"dependencies"`
			} `json:"planNodeDescs"`
			NodeIndexMap struct {
			} `json:"nodeIndexMap"`
			Format           string `json:"format"`
			OptimizeTimeInUs int    `json:"optimize_time_in_us"`
		} `json:"planDesc "`
		Comment string `json:"comment "`
	} `json:"results"`
	Errors []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

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

	// Create schemas
	createTestDataSchema(session)
	// Load data
	loadTestData(session)

	// Complex result
	{
		jsonStrResult, err := session.ExecuteJson("MATCH (v:person {name: \"Bob\"}) RETURN v")
		if err != nil {
			log.Fatal(fmt.Sprintf("fail to get the result in json format, %s", err.Error()))
		}

		var jsonObj JsonObj
		// Parse JSON
		json.Unmarshal(jsonStrResult, &jsonObj)
		// Get row
		rowData := jsonObj.Results[0].Data[0].Row[0]
		row, err := json.MarshalIndent(rowData, "", "  ")
		if err != nil {
			fmt.Println("error:", err)
		}
		fmt.Println(string(row))

		// Get a property
		birthday := jsonObj.Results[0].Data[0].Row[0].(map[string]interface{})["person.birthday"]
		fmt.Printf("person.birthday is %s \n\n", birthday)
	}
	// With error
	{
		jsonStrResult, err := session.ExecuteJson("MAT (v:person {name: \"Bob\"}) RETURN v")
		if err != nil {
			log.Fatal(fmt.Sprintf("fail to get the result in json format, %s", err.Error()))
		}

		var jsonObj JsonObj
		// Parse JSON
		json.Unmarshal(jsonStrResult, &jsonObj)
		// Get error message
		errorMsg := jsonObj.Errors[0].Message
		msg, err := json.MarshalIndent(errorMsg, "", "  ")
		if err != nil {
			fmt.Println("error:", err)
		}
		fmt.Println("Error message: ", string(msg))
	}
	dropSpace(session, "client_test")

	fmt.Print("\n")
	log.Info("Nebula Go JSON parsing Example Finished")
}

// creates schema
func createTestDataSchema(session *nebula.Session) {
	createSchema := "CREATE SPACE IF NOT EXISTS test_data(vid_type = FIXED_STRING(30));" +
		"USE test_data; " +
		"CREATE TAG IF NOT EXISTS person(name string, age int8, grade int16, " +
		"friends int32, book_num int64, birthday datetime, " +
		"start_school date, morning time, property double, " +
		"is_girl bool, child_name fixed_string(10), expend float, " +
		"first_out_city timestamp, hobby string); " +
		"CREATE TAG IF NOT EXISTS student(name string); " +
		"CREATE EDGE IF NOT EXISTS like(likeness double); " +
		"CREATE EDGE IF NOT EXISTS friend(start_Datetime datetime, end_Datetime datetime); " +
		"CREATE TAG INDEX IF NOT EXISTS person_name_index ON person(name(8));"
	resultSet, err := session.Execute(createSchema)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	checkResultSet(createSchema, resultSet)

	time.Sleep(3 * time.Second)
}

// inserts data that used in tests
func loadTestData(session *nebula.Session) {
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
		log.Fatal(err.Error())
		return
	}
	checkResultSet(query, resultSet)

	query =
		"INSERT VERTEX student(name) VALUES " +
			"'Bob':('Bob'), 'Lily':('Lily'), " +
			"'Tom':('Tom'), 'Jerry':('Jerry'), 'John':('John')"
	resultSet, err = session.Execute(query)
	if err != nil {
		log.Fatal(err.Error())
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
		log.Fatal(err.Error())
		return
	}
	checkResultSet(query, resultSet)

	query =
		"INSERT EDGE friend(start_Datetime, end_Datetime) VALUES " +
			"'Bob'->'Lily':(datetime('2008-09-10T10:08:02'), datetime('2010-09-10T10:08:02')), " +
			"'Bob'->'Tom':(datetime('2008-09-10T10:08:02'), datetime('2010-09-10T10:08:02')), " +
			"'Jerry'->'Lily':(datetime('2008-09-10T10:08:02'), datetime('2010-09-10T10:08:02')), " +
			"'Tom'->'Jerry':(datetime('2008-09-10T10:08:02'), datetime('2010-09-10T10:08:02')), " +
			"'Bob'->'John':(datetime('2008-09-10T10:08:02'), datetime('2010-09-10T10:08:02'))"
	resultSet, err = session.Execute(query)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	checkResultSet(query, resultSet)
}

func dropSpace(session *nebula.Session, spaceName string) {
	query := fmt.Sprintf("DROP SPACE IF EXISTS %s;", spaceName)
	resultSet, err := session.Execute(query)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	checkResultSet(query, resultSet)
}

func checkResultSet(prefix string, res *nebula.ResultSet) {
	if !res.IsSucceed() {
		log.Fatal(fmt.Sprintf("%s, ErrorCode: %v, ErrorMsg: %s", prefix, res.GetErrorCode(), res.GetErrorMsg()))
	}
}
