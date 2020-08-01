/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	nebula "github.com/vesoft-inc/nebula-go"
	"github.com/vesoft-inc/nebula-go/examples/connpool"
	"github.com/vesoft-inc/nebula-go/nebula/graph"
)

func check(respCh <-chan connpool.RespData) {
	resp := <-respCh
	if resp.IsError() {
		if resp.Err != nil {
			log.Fatalf("error: %s", resp.Err.Error())
		}
		log.Fatalf("resp code: %s, msg: %s", resp.Resp.GetErrorCode().String(), resp.Resp.GetErrorMsg())
	}
}

func printResp(resp *graph.ExecutionResponse) {
	if nebula.IsError(resp) {
		return
	}

	if resp.IsSetColumnNames() {
		var columnNames []string
		for _, columnName := range resp.GetColumnNames() {
			columnNames = append(columnNames, string(columnName))
		}
		fmt.Println(strings.Join(columnNames, ","))
	}

	if resp.IsSetRows() {
		for _, row := range resp.GetRows() {
			var columns []string
			for _, column := range row.GetColumns() {
				var str string
				switch {
				case column.IsSetBoolVal():
					str = fmt.Sprintf("%t", column.GetBoolVal())
				case column.IsSetDate():
					date := column.GetDate()
					str = fmt.Sprintf("%d-%d-%d", date.GetYear(), date.GetMonth(), date.GetDay())
				case column.IsSetDatetime():
					datetime := column.GetDatetime()
					str = fmt.Sprintf("%d-%d-%d %d:%d:%d %d %d", datetime.GetYear(), datetime.GetMonth(), datetime.GetDay(),
						datetime.GetHour(), datetime.GetMinute(), datetime.GetSecond(), datetime.GetMillisec(), datetime.GetMicrosec())
				case column.IsSetDoublePrecision():
					str = fmt.Sprintf("%f", column.GetDoublePrecision())
				case column.IsSetId():
					str = fmt.Sprintf("%d", int64(column.GetId()))
				case column.IsSetInteger():
					str = fmt.Sprintf("%d", column.GetInteger())
				case column.IsSetMonth():
					str = fmt.Sprintf("%d-%d", column.GetMonth().GetYear(), column.GetMonth().GetMonth())
				case column.IsSetPath():
					path := column.GetPath()
					arrow := ""
					var builder strings.Builder
					// Path format: (1)->[like:0]->(2)->[like:1]->(4)
					for i, element := range path.GetEntryList() {
						vert := element.GetVertex()
						edge := element.GetEdge()
						if i < len(path.GetEntryList())-1 {
							builder.WriteString(fmt.Sprintf("%s(%d)->[%s:%d]", arrow, int64(vert.GetId()), string(edge.GetType()), int64(edge.GetRanking())))
						} else {
							builder.WriteString(fmt.Sprintf("%s(%d)", arrow, int64(vert.GetId())))
						}
						arrow = "->"
					}
					str = builder.String()
				case column.IsSetSinglePrecision():
					str = fmt.Sprintf("%f", column.GetSinglePrecision())
				case column.IsSetStr():
					str = string(column.GetStr())
				case column.IsSetTimestamp():
					str = fmt.Sprintf("%d", int64(column.GetTimestamp()))
				case column.IsSetYear():
					str = fmt.Sprintf("%d", int16(column.GetYear()))
				default:
					log.Fatalln("Invalid column")
				}
				columns = append(columns, str)
			}
			fmt.Println(strings.Join(columns, ","))
		}
	}
}

func main() {
	pool, err := connpool.New(3, "127.0.0.1:4910", "user", "password")
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	// create schema
	check(pool.Execute("CREATE SPACE IF NOT EXISTS nba2"))
	check(pool.Execute("USE nba2"))
	check(pool.Execute("CREATE TAG IF NOT EXISTS person(name string, age int)"))
	check(pool.Execute("CREATE EDGE IF NOT EXISTS like(likeness double)"))

	time.Sleep(6 * time.Second)

	// insert vertices and edges
	check(pool.Execute(`INSERT VERTEX person(name, age) VALUES
	1:("Bob", 30),
	2:("Lily", 29),
	3:("Tom", 31),
	4:("Jerry", 24),
	5:("John", 27)`))
	check(pool.Execute(`INSERT EDGE like(likeness) VALUES
	1->2:(80.0),
	1->3:(70.0),
	2->4:(84.0),
	3->5:(68.3),
	1->5:(97.2)`))

	respCh := pool.Execute("GO FROM 1 OVER like YIELD $$.person.name AS name, $$.person.age AS age, like.likeness AS likeness")
	respData := <-respCh
	if respData.IsError() {
		log.Println(respData.String())
		return
	}

	printResp(respData.Resp)
}
