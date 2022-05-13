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

	nebula "github.com/vesoft-inc/nebula-go/v3"
)

const (
	// The default port of Nebula Graph 2.x is 9669.
	// 3699 is only for testing.
	connString = "nebula://root:nebula@127.0.0.1:3699/basic_example_space"
)

// Initialize logger
var log = nebula.DefaultLogger{}

func main() {
	sessPool, err := nebula.NewSessionPool(connString,
		nebula.WithDefaultLogger(),
		nebula.WithOnAcquireSessionStmt(
			`CREATE SPACE IF NOT EXISTS {{.Space}}(vid_type=FIXED_STRING(20)); USE {{.Space}};`,
		),
	)
	if err != nil {
		log.Fatal(fmt.Sprintf("Fail to initialize the session pool with string %q: %s",
			connString, err.Error()))
	}

	defer sessPool.Close()

	checkResultSet := func(prefix string, res *nebula.ResultSet) {
		if !res.IsSucceed() {
			log.Fatal(fmt.Sprintf("%s, ErrorCode: %v, ErrorMsg: %s", prefix, res.GetErrorCode(), res.GetErrorMsg()))
		}
	}

	{
		session, err := sessPool.Acquire()
		if err != nil {
			log.Fatal(fmt.Sprintf("Fail to acquire a new session from session pool with string %q: %s",
				connString, err.Error()))
		}

		createSchema := `CREATE TAG IF NOT EXISTS person(name string, age int); 
CREATE EDGE IF NOT EXISTS like(likeness double);
`

		// Excute a query
		resultSet, err := session.Execute(createSchema)
		if err != nil {
			log.Error(err.Error())
			return
		}
		checkResultSet(createSchema, resultSet)

		defer sessPool.Release(session)
	}

	err = sessPool.WithSession(func(session nebula.NebulaSession) error {
		query := "DROP SPACE IF EXISTS basic_example_space"
		// Send query
		resultSet, err := session.Execute(query)
		if err != nil {
			return err
		}
		checkResultSet(query, resultSet)

		return nil
	})

	if err != nil {
		log.Fatal(fmt.Sprintf("Fail to use a new session from session pool with string %q: %s",
			connString, err.Error()))
	}

	fmt.Print("\n")
	log.Info("Nebula Go Sesion Pool Example Finished")
}
