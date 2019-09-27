/* Copyright (c) 2019 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package nebula

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/facebook/fbthrift/thrift/lib/go/thrift"
	"github.com/vesoft-inc/nebula-go/nebula/graph"
)

const (
	kColumnTypeEmpty = iota
	kColumnTypeBool
	kColumnTypeInteger
	kColumnTypeID
	kColumnTypeSinglePrecision
	kColumnTypeDoublePrecision
	kColumnTypeStr
	kColumnTypeTimestamp
	kColumnTypeYear
	kColumnTypeMonth
	kColumnTypeDate
	kColumnTypeDatetime
)

type GraphOptions struct {
	Timeout time.Duration
}

type GraphOption func(*GraphOptions)

var defaultGraphOptions = GraphOptions{
	Timeout: 30 * time.Second,
}

type GraphClient struct {
	graph     graph.GraphServiceClient
	option    GraphOptions
	sessionID int64
}

func WithTimeout(duration time.Duration) GraphOption {
	return func(options *GraphOptions) {
		options.Timeout = duration
	}
}

func NewClient(address string, opts ...GraphOption) (client *GraphClient, err error) {
	options := defaultGraphOptions
	for _, opt := range opts {
		opt(&options)
	}

	timeoutOption := thrift.SocketTimeout(options.Timeout)
	addressOption := thrift.SocketAddr(address)
	transport, err := thrift.NewSocket(timeoutOption, addressOption)
	if err != nil {
		return nil, err
	}

	protocol := thrift.NewBinaryProtocolFactoryDefault()
	graph := &GraphClient{
		graph: *graph.NewGraphServiceClientFactory(transport, protocol),
	}
	return graph, nil
}

// Open transport and authenticate
func (client *GraphClient) Connect(username, password string) error {
	if err := client.graph.Transport.Open(); err != nil {
		return err
	}

	if resp, err := client.graph.Authenticate(username, password); err != nil {
		log.Printf("Authentication fails, ErrorCode: %v, ErrorMsg: %s", resp.GetErrorCode(), resp.GetErrorMsg())
		if e := client.graph.Close(); e != nil {
			log.Printf("Fail to close transport, error: %s", e.Error())
		}
		return err
	} else {
		client.sessionID = resp.GetSessionID()
		return nil
	}
}

// Signout and close transport
func (client *GraphClient) Disconnect() {
	if err := client.graph.Signout(client.sessionID); err != nil {
		log.Println("Fail to signout")
	}

	if err := client.graph.Close(); err != nil {
		log.Println("Fail to close transport")
	}
}

func (client *GraphClient) Execute(stmt string) (*graph.ExecutionResponse, error) {
	return client.graph.Execute(client.sessionID, stmt)
}

func (client *GraphClient) PrintResult(response *graph.ExecutionResponse) string {
	widths, formats := computeColumnWidths(response)
	if len(widths) == 0 {
		return ""
	}

	sum := 0
	for _, width := range widths {
		sum += width
	}
	len := sum + 3*len(widths) + 1
	headerLine := strings.Repeat("=", len)
	rowLine := strings.Repeat("-", len)

	builder := strings.Builder{}
	builder.WriteString(headerLine)
	builder.WriteString("\n|")
	builder.WriteString(printHeader(response.GetColumnNames(), widths))
	builder.WriteString(headerLine)
	builder.WriteString("\n")

	builder.WriteString(printData(response.GetRows(), rowLine, widths, formats))

	return builder.String()
}

func computeColumnWidths(resp *graph.ExecutionResponse) (widths []int, formats []string) {
	widths = make([]int, len(resp.ColumnNames))
	for idx, columnName := range resp.ColumnNames {
		widths[idx] = len(string(columnName))
	}

	formats = make([]string, len(widths))
	if len(widths) == 0 || len(resp.Rows) == 0 {
		return
	}

	types := make([]int, len(widths))
	for idx := range widths {
		types[idx] = kColumnTypeEmpty
		formats[idx] = " "
	}

	for rowIdx, row := range resp.Rows {
		if len(widths) != len(row.GetColumns()) {
			log.Fatalf("Wrong number of columns(%d) in row(%d), expected %d", len(row.GetColumns()), rowIdx, len(widths))
		}
		for idx, column := range row.GetColumns() {
			genFmt := types[idx] == kColumnTypeEmpty
			if column.IsSetBoolVal() {
				if types[idx] == kColumnTypeEmpty {
					types[idx] = kColumnTypeBool
				} else {
					if types[idx] != kColumnTypeBool {
						log.Fatalf("%s is not bool column type", columnTypeString(types[idx]))
					}
				}

				if widths[idx] < 5 {
					widths[idx] = 5
					genFmt = true
				}
				if genFmt {
					formats[idx] = fmt.Sprintf(" %%-%ds |", widths[idx])
				}
			} else if column.IsSetInteger() {
				if types[idx] == kColumnTypeEmpty {
					types[idx] = kColumnTypeInteger
				} else {
					if types[idx] != kColumnTypeInteger {
						log.Fatalf("%s is not integer column type", columnTypeString(types[idx]))
					}
				}

				val := column.GetInteger()
				len := len(fmt.Sprintf("%d", val))
				if widths[idx] < len {
					widths[idx] = len
					genFmt = true
				}

				if genFmt {
					formats[idx] = fmt.Sprintf(" %%-%dd |", widths[idx])
				}
			} else if column.IsSetId() {
				if types[idx] == kColumnTypeEmpty {
					types[idx] = kColumnTypeID
				} else {
					if types[idx] != kColumnTypeID {
						log.Fatalf("%s is not id column type", columnTypeString(types[idx]))
					}
				}

				val := column.GetId()
				len := len(fmt.Sprintf("%d", val))
				if widths[idx] < len {
					widths[idx] = len
					genFmt = true
				}

				if genFmt {
					formats[idx] = fmt.Sprintf(" %%-%dd |", widths[idx])
				}
			} else if column.IsSetSinglePrecision() {
				if types[idx] == kColumnTypeEmpty {
					types[idx] = kColumnTypeSinglePrecision
				} else {
					if types[idx] != kColumnTypeSinglePrecision {
						log.Fatalf("%s is not single precision column type", columnTypeString(types[idx]))
					}
				}

				val := column.GetSinglePrecision()
				len := len(fmt.Sprintf("%f", val))
				if widths[idx] < len {
					widths[idx] = len
					genFmt = true
				}
				if genFmt {
					formats[idx] = fmt.Sprintf(" %%-%df |", widths[idx])
				}
			} else if column.IsSetDoublePrecision() {
				if types[idx] == kColumnTypeEmpty {
					types[idx] = kColumnTypeDoublePrecision
				} else {
					if types[idx] != kColumnTypeDoublePrecision {
						log.Fatalf("%s is not double precision column type", columnTypeString(types[idx]))
					}
				}

				val := column.GetDoublePrecision()
				len := len(fmt.Sprintf("%f", val))
				if widths[idx] < len {
					widths[idx] = len
					genFmt = true
				}
				if genFmt {
					formats[idx] = fmt.Sprintf(" %%-%df |", widths[idx])
				}
			} else if column.IsSetStr() {
				if types[idx] == kColumnTypeEmpty {
					types[idx] = kColumnTypeStr
				} else {
					if types[idx] != kColumnTypeStr {
						log.Fatalf("%s is not str column type", columnTypeString(types[idx]))
					}
				}

				val := column.GetStr()
				len := len(string(val))
				if widths[idx] < len {
					widths[idx] = len
					genFmt = true
				}

				if genFmt {
					formats[idx] = fmt.Sprintf(" %%-%ds |", widths[idx])
				}
			} else if column.IsSetTimestamp() {
				if types[idx] == kColumnTypeEmpty {
					types[idx] = kColumnTypeTimestamp
				} else {
					if types[idx] != kColumnTypeTimestamp {
						log.Fatalf("%s is not timestamp column type", columnTypeString(types[idx]))
					}
				}

				if widths[idx] < 19 {
					widths[idx] = 19
					genFmt = true
				}

				if genFmt {
					formats[idx] = fmt.Sprintf(" %%%dd-%%02d-%%02d %%02d:%%02d:%%02d |", widths[idx]-15)
				}
			} else if column.IsSetYear() {
				if types[idx] == kColumnTypeEmpty {
					types[idx] = kColumnTypeYear
				} else {
					if types[idx] != kColumnTypeYear {
						log.Fatalf("%s is not year column type", columnTypeString(types[idx]))
					}
				}

				if widths[idx] < 4 {
					widths[idx] = 4
					genFmt = true
				}

				if genFmt {
					formats[idx] = fmt.Sprintf(" %%-%dd |", widths[idx])
				}
			} else if column.IsSetMonth() {
				if types[idx] == kColumnTypeEmpty {
					types[idx] = kColumnTypeMonth
				} else {
					if types[idx] != kColumnTypeMonth {
						log.Fatalf("%s is not month column type", columnTypeString(types[idx]))
					}
				}

				if widths[idx] < 7 {
					widths[idx] = 7
					genFmt = true
				}
				if genFmt {
					formats[idx] = fmt.Sprintf(" %%%dd/%%02d |", widths[idx]-3)
				}
			} else if column.IsSetDate() {
				if types[idx] == kColumnTypeEmpty {
					types[idx] = kColumnTypeDate
				} else {
					if types[idx] != kColumnTypeDate {
						types[idx] = kColumnTypeDate
					}
				}

				if widths[idx] < 10 {
					widths[idx] = 10
					genFmt = true
				}

				if genFmt {
					formats[idx] = fmt.Sprintf(" %%%dd/%%02d/%%02d |", widths[idx]-6)
				}
			} else if column.IsSetDatetime() {
				if types[idx] == kColumnTypeEmpty {
					types[idx] = kColumnTypeDatetime
				} else {
					if types[idx] != kColumnTypeDatetime {
						log.Fatalf("%s is not datetime column type", columnTypeString(types[idx]))
					}
				}

				formats[idx] = fmt.Sprintf(" %%%dd/%%02d/%%02d %%02d:%%02d:%%02d.%%03d%%03d |", widths[idx]-22)
			} else {
				if types[idx] != kColumnTypeEmpty {
					log.Fatalf("Wrong column type: %s", columnTypeString(types[idx]))
				}
			}
		}
	}

	return
}

func columnTypeString(columnType int) string {
	switch columnType {
	case kColumnTypeEmpty:
		return "Empty"
	case kColumnTypeBool:
		return "Bool"
	case kColumnTypeInteger:
		return "Integer"
	case kColumnTypeID:
		return "ID"
	case kColumnTypeSinglePrecision:
		return "Single precision"
	case kColumnTypeDoublePrecision:
		return "Double precision"
	case kColumnTypeStr:
		return "Str"
	case kColumnTypeTimestamp:
		return "Timestamp"
	case kColumnTypeYear:
		return "Year"
	case kColumnTypeMonth:
		return "Month"
	case kColumnTypeDate:
		return "Date"
	case kColumnTypeDatetime:
		return "Datetime"
	default:
		log.Printf("Invalid column type %d", columnType)
		return ""
	}
}

func printHeader(columnNames [][]byte, widths []int) string {
	if len(columnNames) == 0 {
		return ""
	}

	builder := strings.Builder{}

	for idx, columnName := range columnNames {
		format := fmt.Sprintf(" %%-%ds |", widths[idx])
		builder.WriteString(fmt.Sprintf(format, string(columnName)))
	}

	builder.WriteString("\n")

	return builder.String()
}

func printData(rows []*graph.RowValue, rowLine string, widths []int, formats []string) string {
	if len(rows) == 0 {
		return ""
	}

	builder := strings.Builder{}

	for _, row := range rows {
		builder.WriteString("|")
		for colIdx, column := range row.GetColumns() {
			var str string
			if column.IsSetBoolVal() {
				if column.GetBoolVal() {
					str = fmt.Sprintf(formats[colIdx], "true")
				} else {
					str = fmt.Sprintf(formats[colIdx], "false")
				}
			} else if column.IsSetInteger() {
				str = fmt.Sprintf(formats[colIdx], column.GetInteger())
			} else if column.IsSetId() {
				str = fmt.Sprintf(formats[colIdx], column.GetId())
			} else if column.IsSetSinglePrecision() {
				str = fmt.Sprintf(formats[colIdx], column.GetSinglePrecision())
			} else if column.IsSetDoublePrecision() {
				str = fmt.Sprintf(formats[colIdx], column.GetDoublePrecision())
			} else if column.IsSetStr() {
				str = fmt.Sprintf(formats[colIdx], column.GetStr())
			} else if column.IsSetTimestamp() {
				timestamp := column.GetTimestamp()
				tm := time.Unix(int64(timestamp), 0)
				str = fmt.Sprintf(formats[colIdx], tm.Year()+1900, tm.Month()+1, tm.Day(), tm.Hour(), tm.Minute(), tm.Second())
			} else if column.IsSetYear() {
				str = fmt.Sprintf(formats[colIdx], column.GetYear())
			} else if column.IsSetMonth() {
				month := column.GetMonth()
				str = fmt.Sprintf(formats[colIdx], month.GetYear(), month.GetMonth())
			} else if column.IsSetDate() {
				date := column.GetDate()
				str = fmt.Sprintf(formats[colIdx], date.GetYear(), date.GetMonth(), date.GetDay())
			} else if column.IsSetDatetime() {
				dt := column.GetDatetime()
				str = fmt.Sprintf(formats[colIdx], dt.GetYear(), dt.GetMonth(), dt.GetDay(), dt.GetHour(), dt.GetMinute(), dt.GetSecond(), dt.GetMicrosec())
			} else {
				format := fmt.Sprintf(" %%-%dc |", widths[colIdx])
				str = fmt.Sprintf(format, " ")
			}
			builder.WriteString(str)
		}
		builder.WriteString("\n")
		builder.WriteString(rowLine)
		builder.WriteString("\n")
	}

	return builder.String()
}
