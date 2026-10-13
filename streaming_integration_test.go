// Copyright 2026 vesoft inc. All Rights Reserved.
package nebula_ng

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vesoft-inc/nebula-go/v5/internal/generated_code/v5.0.0/proto/graph"
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
		`FINISH`,
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

// Each sample ends after normal EOF; Close and assertions are outside the measured interval.
// JSONL preserves every sample, its pair/order, and the distinct completion points.
func TestStreamingGraphTiming(t *testing.T) {
	client := streamingGraphClient(t)
	if os.Getenv("NEBULA_STREAM_TEST_BENCH") == "" {
		t.Skip("performance test is opt-in")
	}
	type sample struct {
		Case      string    `json:"case"`
		Round     int       `json:"round"`
		Pair      int       `json:"pair"`
		Order     int       `json:"order"`
		Streaming bool      `json:"streaming"`
		Rows      int       `json:"rows"`
		Elapsed   float64   `json:"elapsed_us"`
		Returned  float64   `json:"returned_us"`
		FirstRow  float64   `json:"first_row_us"`
		Server    [6]int64  `json:"server_us"`
		Wire      bool      `json:"wire"`
		Arrivals  []float64 `json:"frame_arrivals_us,omitempty"`
	}
	var output *json.Encoder
	if path := os.Getenv("NEBULA_STREAM_TEST_SAMPLES"); path != "" {
		file, err := os.Create(path)
		require.NoError(t, err)
		defer file.Close()
		output = json.NewEncoder(file)
	}
	wire := os.Getenv("NEBULA_STREAM_TEST_WIRE") != ""
	measure := func(query string, streaming bool) sample {
		if wire {
			cn := client.(*driverConn).conn.(*connection)
			req := &graph.ExecuteRequest{SessionId: cn.sessionId, Stmt: []byte(query)}
			s := sample{Streaming: streaming, Wire: true, FirstRow: -1, Arrivals: make([]float64, 0, 4)}
			var summary *graph.Summary
			start := time.Now()
			if streaming {
				call, err := cn.graphClient.StreamingExecute(context.Background(), req)
				if err != nil {
					t.Fatal(err)
				}
				final := false
				for {
					frame, err := call.Recv()
					arrival := float64(time.Since(start).Nanoseconds()) / 1000
					if err == io.EOF {
						s.Elapsed = arrival
						break
					}
					if err != nil {
						t.Fatal(err)
					}
					if final {
						t.Fatal("frame after final")
					}
					if len(s.Arrivals) == 0 {
						s.Returned = arrival
					}
					s.Arrivals = append(s.Arrivals, arrival)
					if batch := frame.GetBatch(); batch != nil {
						if s.FirstRow < 0 {
							s.FirstRow = arrival
						}
						if len(batch.Vectors) > 0 {
							s.Rows += int(batch.Vectors[0].CommonMetaData.NumRecords)
						} else {
							s.Rows += int(batch.NumRecords)
						}
					}
					if terminal := frame.GetFinal(); terminal != nil {
						if string(terminal.GetStatus().GetCode()) != "00000" {
							t.Fatal(terminal.GetStatus())
						}
						summary = terminal.Summary
						final = true
					}
				}
				require.True(t, final)
			} else {
				response, err := cn.graphClient.Execute(context.Background(), req)
				s.Elapsed = float64(time.Since(start).Nanoseconds()) / 1000
				require.NoError(t, err)
				require.Equal(t, "00000", string(response.GetStatus().GetCode()))
				s.Returned = s.Elapsed
				s.Rows = int(response.GetResult().GetMeta().GetNumRecords())
				if s.Rows > 0 {
					s.FirstRow = s.Elapsed
				}
				summary = response.Summary
			}
			e := summary.GetElapsedTime()
			s.Server = [6]int64{e.GetParseTimeUs(), e.GetBuildTimeUs(), e.GetOptimizeTimeUs(), e.GetExecutionTimeUs(), e.GetSerializeTimeUs(), e.GetTotalServerTimeUs()}
			return s
		}
		start := time.Now()
		var result types.Result
		var err error
		if streaming {
			result, err = client.(types.StreamingClient).StreamingExecute(context.Background(), query)
		} else {
			result, err = client.Execute(query)
		}
		if err != nil {
			t.Fatal(err)
		}
		returned := time.Since(start)
		first := time.Duration(-1)
		rows := 0
		for result.HasNext() {
			_, err = result.Next()
			if err != nil {
				t.Fatal(err)
			}
			if rows == 0 {
				first = time.Since(start)
			}
			rows++
		}
		if stream, ok := result.(types.StreamingResult); ok {
			if err := stream.Err(); err != nil {
				t.Fatal(err)
			}
		}
		elapsed := time.Since(start)
		summary := result.Summary()
		if stream, ok := result.(types.StreamingResult); ok {
			require.NoError(t, stream.Close())
		}
		firstUS := -1.0
		if rows > 0 {
			firstUS = float64(first.Nanoseconds()) / 1000
		}
		return sample{Streaming: streaming, Rows: rows, Elapsed: float64(elapsed.Nanoseconds()) / 1000,
			Returned: float64(returned.Nanoseconds()) / 1000, FirstRow: firstUS,
			Server: [6]int64{summary.ParseTimeUs(), summary.BuildTimeUs(), summary.OptimizeTimeUs(),
				summary.ExecutionTimeUs(), summary.SerializeTimeUs(), summary.TotalServerTimeUs()}}
	}
	count, rounds := 500, 3
	for name, target := range map[string]*int{"NEBULA_STREAM_TEST_BENCH_PAIRS": &count, "NEBULA_STREAM_TEST_ROUNDS": &rounds} {
		if value := os.Getenv(name); value != "" {
			n, err := strconv.Atoi(value)
			require.NoError(t, err)
			require.Greater(t, n, 0)
			*target = n
		}
	}
	type timingCase struct {
		name, query string
		rows        int
	}
	cases := []timingCase{
		{"no_table", "EXPLAIN RETURN 1 AS x", 0},
		{"empty", "FOR i IN [1] FILTER i < 0 RETURN i AS x", 0},
		{"one_row", "RETURN 1 AS x", 1},
	}
	for _, rows := range []int{2, 10, 100, 1023, 1024, 1025, 10000, 100000, 1000000} {
		cases = append(cases, timingCase{fmt.Sprintf("rows_%d", rows), fmt.Sprintf("FOR i IN range(1,%d) RETURN i AS x ORDER BY i", rows), rows})
	}
	for _, rows := range []int{1, 100, 1024, 10000} {
		cases = append(cases, timingCase{fmt.Sprintf("wide_%d", rows), fmt.Sprintf("FOR i IN range(1,%d) RETURN i AS i ORDER BY i NEXT RETURN repeat(CAST(i %% 10 AS STRING),128) AS x", rows), rows})
	}
	selected, mode := os.Getenv("NEBULA_STREAM_TEST_CASE"), os.Getenv("NEBULA_STREAM_TEST_MODE")
	require.Contains(t, []string{"", "unary", "streaming"}, mode)
	matched := false
	for round := 0; round < rounds; round++ {
		for index := range cases {
			ci := index
			if round%2 == 1 {
				ci = len(cases) - 1 - index
			}
			c := cases[ci]
			if selected != "" && selected != c.name {
				continue
			}
			matched = true
			pairs := count
			if c.rows >= 100000 {
				pairs = min(pairs, 10)
			}
			if c.rows >= 1000000 {
				pairs = min(pairs, 5)
			}
			for warm := 0; warm < min(pairs, 20); warm++ {
				if mode != "streaming" {
					require.Equal(t, c.rows, measure(c.query, false).Rows)
				}
				if mode != "unary" {
					require.Equal(t, c.rows, measure(c.query, true).Rows)
				}
			}
			samples := [2][]float64{}
			deltas := []float64{}
			for pair := 0; pair < pairs; pair++ {
				values := [2]float64{}
				for order := 0; order < 2; order++ {
					variant := (pair + round + order) % 2
					streaming := variant == 1
					if (mode == "unary" && streaming) || (mode == "streaming" && !streaming) {
						continue
					}
					s := measure(c.query, streaming)
					require.Equal(t, c.rows, s.Rows)
					s.Case = c.name
					s.Round = round
					s.Pair = pair
					s.Order = order
					if output != nil {
						require.NoError(t, output.Encode(s))
					}
					values[variant] = s.Elapsed
					samples[variant] = append(samples[variant], s.Elapsed)
				}
				if mode == "" {
					deltas = append(deltas, values[1]-values[0])
				}
			}
			for variant, values := range samples {
				if len(values) == 0 {
					continue
				}
				sort.Float64s(values)
				t.Logf("case=%s round=%d streaming=%v n=%d median_us=%.3f p90_us=%.3f", c.name, round, variant == 1, len(values), values[len(values)/2], values[len(values)*9/10])
			}
			if len(deltas) > 1 {
				mean := 0.0
				for _, d := range deltas {
					mean += d / float64(len(deltas))
				}
				variance := 0.0
				for _, d := range deltas {
					variance += (d - mean) * (d - mean) / float64(len(deltas)-1)
				}
				bound := 1.96 * math.Sqrt(variance/float64(len(deltas)))
				t.Logf("case=%s round=%d paired_mean_difference_us=%.3f confidence_95_us=[%.3f,%.3f]", c.name, round, mean, mean-bound, mean+bound)
			}
		}
	}
	require.True(t, matched, "unknown timing case: %s", selected)
	if os.Getenv("NEBULA_STREAM_TEST_TIMING_ONLY") != "" || selected != "" || mode != "" {
		return
	}
	query := "FOR i IN range(0,8191) RETURN i AS i ORDER BY i NEXT RETURN repeat(CAST(i % 10 AS STRING),131072) AS x"
	for round := 0; round < 3; round++ {
		for order := 0; order < 2; order++ {
			runtime.GC()
			streaming := (round+order)%2 == 1
			s := measure(query, streaming)
			require.Equal(t, 8192, s.Rows)
			s.Case = "one_gib"
			s.Round = round
			s.Order = order
			if output != nil {
				require.NoError(t, output.Encode(s))
			}
			t.Logf("one_gib streaming=%v duration_us=%.3f mib_per_sec=%.2f", streaming, s.Elapsed, 1024e6/s.Elapsed)
		}
	}
}
