// Copyright 2025 vesoft inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package types

import (
	"context"
	"iter"
)

type (
	Result interface {
		Summary() Summary
		Cursor() []byte
		Table
	}

	Table interface {
		RowSize() int
		HasNext() bool
		Next() (Row, error)
		// Scan copies the columns in the current row into the values pointed at by dest.
		// If there's no more row, return io.EOF
		Scan(...any) error
		// All returns a sequence of all remaining rows in the table.
		// If there's an error during iteration,
		// the error will be returned as the second value of the sequence.
		All() iter.Seq2[Row, error]
		Columns() []string
		ColumnTypes() []ColumnType //not support yet
	}

	Row interface {
		Values() []Value
		GetValueByName(name string) (Value, error)
		GetValueByIndex(index int) (Value, error)
	}

	Summary interface {
		ParseTimeUs() int64
		BuildTimeUs() int64
		OptimizeTimeUs() int64
		ExecutionTimeUs() int64
		SerializeTimeUs() int64
		TotalServerTimeUs() int64
		ExplainType() string
		PlanInfo() PlanInfo
		QueryStats() QueryStats
		NumWarnings() int
		Log() string
	}
	PlanInfo interface {
		Id() string
		Name() string
		Details() string
		Columns() []string
		TimeMs() float64
		Rows() int64
		MemoryKib() float64
		BlockedMs() float64
		QueuedMs() float64
		ConsumeMs() float64
		ProduceMs() float64
		FinishMs() float64
		Batches() int64
		Concurrency() int64
		OtherStatsJson() []byte
		Children() []PlanInfo
	}
	QueryStats interface {
		NumAffectedNodes() int64
		NumAffectedEdges() int64
		ExportedRowSize() int
		ExportedPaths() []string
	}

	Client interface {
		Execute(stmt string) (Result, error)
		ExecuteContext(ctx context.Context, stmt string) (Result, error)
		Ping() error // by default, timeout is 1s.
		PingContext(ctx context.Context) error
		IsClosed() bool
		Close() error
		GetSessionId() (int64, error)
		GetVersion() (string, error)
	}

	Pool interface {
		GetClient() (Client, error)
		PutClient(Client) error
		Close() error
	}

	ColumnType int
)

const (
	ColumnTypeInvalid ColumnType = iota
	ColumnTypeNode
	ColumnTypeEdge
	ColumnTypePath
	ColumnTypeUnknown
	ColumnTypeBool
	ColumnTypeInt8
	ColumnTypeUint8
	ColumnTypeInt16
	ColumnTypeUint16
	ColumnTypeInt32
	ColumnTypeUint32
	ColumnTypeInt64
	ColumnTypeUint64
	ColumnTypeFloat32
	ColumnTypeFloat64
	ColumnTypeString
	ColumnTypeList
	ColumnTypeRecord
	ColumnTypeLocalTime
	ColumnTypeLocalDatetime
	ColumnTypeZonedTime
	ColumnTypeZonedDatetime
	ColumnTypeDate
	ColumnTypeDuration
	ColumnTypeDecimal
	ColumnTypeVector
	ColumnTypeGeography
	ColumnTypeSet
	ColumnTypeMap
	ColumnTypeAny ColumnType = 0xFF
)
