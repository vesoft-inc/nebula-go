// Copyright 2026 vesoft inc. All Rights Reserved.
package nebula_ng

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vesoft-inc/nebula-go/v5/pkg/types"
)

func streamingGraphClient(t *testing.T) types.Client {
	t.Helper()
	address := os.Getenv("NEBULA_STREAM_TEST_ADDRESS")
	if address == "" {
		t.Skip("NEBULA_STREAM_TEST_ADDRESS is unset")
	}
	client, err := NewNebulaClient(address, "root", "NebulaGraph01")
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// Check each row against an independent oracle; do not collect the stream to compare it.
func TestStreamingGraphScaleMatrix(t *testing.T) {
	client := streamingGraphClient(t)
	for _, batch := range []int{16, 1024} {
		for _, count := range []int{0, 1, 2, 10, 100, batch - 1, batch, batch + 1, 2*batch + 1, 10000, 100000, 1000000} {
			for _, width := range []int{1, 128} {
				t.Run(fmt.Sprintf("batch%d/rows%d/width%d", batch, count, width), func(t *testing.T) {
					query := fmt.Sprintf("/*+ SET_VAR(exec_batch_size=%d) */ FOR i IN range(0,%d) FILTER i < %d RETURN i AS i ORDER BY i NEXT RETURN repeat(CAST(i %% 10 AS STRING),%d) AS x", batch, max(count-1, 0), count, width)
					for _, stream := range []bool{false, true} {
						var result types.Result
						var err error
						if stream {
							result, err = client.(types.StreamingClient).StreamingExecute(context.Background(), query)
						} else {
							result, err = client.Execute(query)
						}
						require.NoError(t, err)
						expected := [10]string{}
						for digit := range expected {
							expected[digit] = strings.Repeat(strconv.Itoa(digit), width)
						}
						n := 0
						for result.HasNext() {
							row, err := result.Next()
							require.NoError(t, err)
							value, err := row.GetValueByIndex(0)
							require.NoError(t, err)
							actual, err := value.AsString()
							if err != nil || string(actual) != expected[n%10] {
								t.Fatalf("stream=%v row=%d: %q (%v)", stream, n, actual, err)
							}
							n++
						}
						if stream {
							sr := result.(types.StreamingResult)
							require.NoError(t, sr.Err())
							require.NoError(t, sr.Close())
						}
						require.Equal(t, count, n)
					}
				})
			}
		}
	}
}
func resultRows(t *testing.T, result types.Result) [][]string {
	t.Helper()
	var rows [][]string
	for result.HasNext() {
		row, err := result.Next()
		require.NoError(t, err)
		values := make([]string, len(row.Values()))
		for i, v := range row.Values() {
			values[i] = v.String()
		}
		rows = append(rows, values)
	}
	if stream, ok := result.(types.StreamingResult); ok {
		require.NoError(t, stream.Err())
	}
	return rows
}

func TestStreamingGraphIntegration(t *testing.T) {
	client := streamingGraphClient(t)
	streamClient := client.(types.StreamingClient)
	for _, query := range []string{
		`RETURN 1 AS x, true AS b, "hello" AS s, null AS n`,
		`RETURN LIST[1,2,3] AS xs, {a:1,b:"test"} AS rec`,
		`FOR i IN range(1,5000) RETURN i AS x`,
		`FOR i IN [1] FILTER i < 0 RETURN i AS x`,
		`EXPLAIN RETURN 1 AS x`,
		`SHOW GRAPHS`,
		`SHOW PARAMETERS`,
		`DESCRIBE GRAPH TYPE ldbc_type`,
		`DESCRIBE EDGE TYPE KNOWS OF ldbc_type`,
		`USE ldbc MATCH p=(n:Person{id:1})-[e:KNOWS]->(m:Person{id:1}) RETURN n,e,p,list[n,m] AS ns,{src:n,link:e} AS rec LIMIT 2`,
		`RETURN date("2020-01-01") AS d, local_datetime("2020-01-01T12:00:00") AS dt`,
		`CALL {VALUE x = 1 RETURN x AS x} RETURN *`,
		`VALUE x = 1
IF false THEN {RETURN x AS x} ELSE {RETURN 2 AS x}`,
		`START TRANSACTION
RETURN 1 AS x
COMMIT`,
		"VALUE x = 1\nRETURN x AS x",
		"IF true THEN {FOR i IN range(1,5000) RETURN i AS x} ELSE {RETURN 0 AS x}",
		"VALUE x = 0\nWHILE x < 3 THEN {SET x = x + 1}\nRETURN x AS x",
	} {
		t.Run(query, func(t *testing.T) {
			unary, err := client.Execute(query)
			require.NoError(t, err)
			expected := resultRows(t, unary)
			result, err := streamClient.StreamingExecute(context.Background(), query)
			require.NoError(t, err)
			defer result.Close()
			require.Equal(t, unary.Columns(), result.Columns())
			require.Equal(t, expected, resultRows(t, result))
			require.NotNil(t, result.Summary())
			next, err := client.Execute("RETURN 2")
			require.NoError(t, err)
			require.Len(t, resultRows(t, next), 1)
		})
	}
}

// Report actual transfer and row decoding time. Logical row bytes are also checked,
// so a constant vector cannot turn the large-result test into a tiny transfer.
func TestStreamingGraphLargeResult(t *testing.T) {
	client := streamingGraphClient(t)
	if os.Getenv("NEBULA_STREAM_TEST_LARGE") == "" {
		t.Skip("large result test is opt-in")
	}
	sc := client.(types.StreamingClient)
	totals := []int64{1 << 20, 1 << 30, 10 << 30}
	if selected := os.Getenv("NEBULA_STREAM_TEST_BYTES"); selected != "" {
		total, err := strconv.ParseInt(selected, 10, 64)
		require.NoError(t, err)
		totals = []int64{total}
	}
	for _, total := range totals {
		t.Run(fmt.Sprint(total), func(t *testing.T) {
			rowBytes := int64(128 << 10)
			if total == 1<<20 {
				rowBytes = 1024
			}
			count := total / rowBytes
			query := fmt.Sprintf("FOR i IN range(0,%d) RETURN i AS i ORDER BY i NEXT RETURN repeat(CAST(i %% 10 AS STRING),%d) AS x", count-1, rowBytes)
			if os.Getenv("NEBULA_STREAM_TEST_NO_SORT") != "" {
				query = fmt.Sprintf("FOR i IN range(0,%d) RETURN repeat(CAST(i %% 10 AS STRING),%d) AS x", count-1, rowBytes)
			}
			if hints := os.Getenv("NEBULA_STREAM_TEST_HINTS"); hints != "" {
				query = hints + " " + query
			}
			if os.Getenv("NEBULA_STREAM_TEST_PROFILE") != "" {
				query = "PROFILE " + query
			}
			started := time.Now()
			result, err := sc.StreamingExecute(context.Background(), query)
			require.NoError(t, err)
			defer result.Close()
			rows, bytes := int64(0), int64(0)
			var first time.Duration
			for result.HasNext() {
				if rows == 0 {
					first = time.Since(started)
				}
				row, err := result.Next()
				require.NoError(t, err)
				value, err := row.GetValueByIndex(0)
				require.NoError(t, err)
				str, err := value.AsString()
				require.NoError(t, err)
				require.Len(t, str, int(rowBytes))
				require.Equal(t, len(str), strings.Count(string(str), strconv.Itoa(int(rows%10))))
				bytes += int64(len(str))
				rows++
			}
			require.NoError(t, result.Err())
			require.Equal(t, count, rows)
			require.Equal(t, total, bytes)
			t.Logf("rows=%d bytes=%d first=%v total=%v schema=%v", rows, bytes, first, time.Since(started), reflect.TypeOf(result))
			if os.Getenv("NEBULA_STREAM_TEST_PROFILE") != "" {
				var logPlan func(types.PlanInfo)
				logPlan = func(plan types.PlanInfo) {
					t.Logf("operator=%s memory_kib=%.0f rows=%d batches=%d blocked_ms=%.3f", plan.Name(), plan.MemoryKib(), plan.Rows(), plan.Batches(), plan.BlockedMs())
					for _, child := range plan.Children() {
						logPlan(child)
					}
				}
				logPlan(result.Summary().PlanInfo())
			}
		})
	}
}

func TestStreamingGraphLargeRow(t *testing.T) {
	client := streamingGraphClient(t)
	if os.Getenv("NEBULA_STREAM_TEST_LARGE") == "" {
		t.Skip("large row test is opt-in")
	}
	rowBytes := (256 << 20) - 1024
	result, err := client.(types.StreamingClient).StreamingExecute(context.Background(),
		fmt.Sprintf("FOR i IN range(0,1) RETURN repeat(CAST(i AS STRING),%d) AS x", rowBytes))
	require.NoError(t, err)
	defer result.Close()
	for i := 0; i < 2; i++ {
		require.True(t, result.HasNext(), result.Err())
		row, err := result.Next()
		require.NoError(t, err)
		value, err := row.GetValueByIndex(0)
		require.NoError(t, err)
		str, err := value.AsString()
		require.NoError(t, err)
		require.Len(t, str, rowBytes)
		require.Equal(t, len(str), strings.Count(string(str), strconv.Itoa(i)))
	}
	require.False(t, result.HasNext())
	require.NoError(t, result.Err())
}

func TestStreamingGraphZeroColumns(t *testing.T) {
	client := streamingGraphClient(t)
	for _, query := range []string{"FINISH", "FOR x IN [] FINISH"} {
		t.Run(query, func(t *testing.T) {
			result, err := client.(types.StreamingClient).StreamingExecute(context.Background(), query)
			require.NoError(t, err)
			defer result.Close()
			require.Empty(t, result.Columns())
			require.False(t, result.HasNext())
			require.NoError(t, result.Err())
			require.NotNil(t, result.Summary())
		})
	}
}

// A successful ROLLBACK returns no result table: unary has no columns, streaming sends no
// metadata or batches before its successful final.
func TestStreamingRollbackReturnsNoResult(t *testing.T) {
	client := streamingGraphClient(t)
	query := "START TRANSACTION RETURN 1 AS x ROLLBACK"
	result, err := client.Execute(query)
	require.NoError(t, err)
	// A zero-column table would also have no columns and rows; only a missing table is correct.
	require.Nil(t, result.(*resultSet).table)

	stream, err := client.(types.StreamingClient).StreamingExecute(context.Background(), query)
	require.NoError(t, err)
	require.False(t, stream.HasNext())
	require.NoError(t, stream.Err())
	require.NotNil(t, stream.Summary())
	require.Nil(t, stream.(*streamingResult).metadata)
}
