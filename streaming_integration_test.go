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

// Each sample ends after normal EOF; Close and assertions are outside the measured interval.
// JSONL preserves every sample, its pair/order, and the distinct completion points.
func TestStreamingGraphTiming(t *testing.T) {
	client := streamingGraphClient(t)
	if os.Getenv("NEBULA_STREAM_TEST_BENCH") == "" {
		t.Skip("performance test is opt-in")
	}
	type sample struct {
		Case         string    `json:"case"`
		Round        int       `json:"round"`
		Pair         int       `json:"pair"`
		Order        int       `json:"order"`
		Streaming    bool      `json:"streaming"`
		Rows         int       `json:"rows"`
		LogicalBytes int64     `json:"logical_bytes"`
		Elapsed      float64   `json:"elapsed_us"`
		Returned     float64   `json:"returned_us"`
		FirstRow     float64   `json:"first_row_us"`
		Server       [6]int64  `json:"server_us"`
		Wire         bool      `json:"wire"`
		Arrivals     []float64 `json:"frame_arrivals_us,omitempty"`
	}
	var output *json.Encoder
	if path := os.Getenv("NEBULA_STREAM_TEST_SAMPLES"); path != "" {
		file, err := os.Create(path)
		require.NoError(t, err)
		defer file.Close()
		output = json.NewEncoder(file)
	}
	wire := os.Getenv("NEBULA_STREAM_TEST_WIRE") != ""
	consumeNs := 0
	if value := os.Getenv("NEBULA_STREAM_TEST_CONSUME_NS"); value != "" {
		var err error
		consumeNs, err = strconv.Atoi(value)
		require.NoError(t, err)
		require.GreaterOrEqual(t, consumeNs, 0)
	}
	require.False(t, wire && consumeNs > 0, "business consumption requires decoded rows")
	withHints := func(query string) string {
		if hints := os.Getenv("NEBULA_STREAM_TEST_HINTS"); hints != "" {
			if strings.HasPrefix(query, "EXPLAIN ") {
				query = "EXPLAIN " + hints + " " + strings.TrimPrefix(query, "EXPLAIN ")
			} else {
				query = hints + " " + query
			}
		}
		return query
	}
	measure := func(query string, streaming bool) sample {
		query = withHints(query)
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
			if consumeNs > 0 {
				until := time.Now().Add(time.Duration(consumeNs))
				for time.Now().Before(until) {
				}
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
	largePairs := 0
	if value := os.Getenv("NEBULA_STREAM_TEST_LARGE_PAIRS"); value != "" {
		var err error
		largePairs, err = strconv.Atoi(value)
		require.NoError(t, err)
		require.Greater(t, largePairs, 0)
	}
	for name, target := range map[string]*int{"NEBULA_STREAM_TEST_BENCH_PAIRS": &count, "NEBULA_STREAM_TEST_ROUNDS": &rounds} {
		if value := os.Getenv(name); value != "" {
			n, err := strconv.Atoi(value)
			require.NoError(t, err)
			require.Greater(t, n, 0)
			*target = n
		}
	}
	type timingCase struct {
		name, query  string
		rows         int
		logicalBytes int64
	}
	cases := []timingCase{
		{"no_table", "EXPLAIN RETURN 1 AS x", 0, 0},
		{"empty", "FOR i IN [1] FILTER i < 0 RETURN i AS x", 0, 0},
		{"one_row", "RETURN 1 AS x", 1, 4},
	}
	for _, rows := range []int{2, 10, 100, 1023, 1024, 1025, 10000, 100000, 1000000} {
		cases = append(cases, timingCase{fmt.Sprintf("rows_%d", rows), fmt.Sprintf("FOR i IN range(1,%d) RETURN i AS x ORDER BY i", rows), rows, int64(rows) * 4})
	}
	for _, rows := range []int{1, 100, 1024, 10000} {
		cases = append(cases, timingCase{fmt.Sprintf("wide_%d", rows), fmt.Sprintf("FOR i IN range(1,%d) RETURN i AS i ORDER BY i NEXT RETURN repeat(CAST(i %% 10 AS STRING),128) AS x", rows), rows, int64(rows) * 128})
	}
	if os.Getenv("NEBULA_STREAM_TEST_BYTE_MATRIX") != "" {
		for _, total := range []int{1 << 20, 16 << 20} {
			for _, width := range []int{16, 128, 4096, 131072} {
				rows := total / width
				cases = append(cases, timingCase{
					fmt.Sprintf("bytes_%d_width_%d", total, width),
					fmt.Sprintf("FOR i IN range(0,%d) RETURN i AS i ORDER BY i NEXT RETURN repeat(CAST(i %% 10 AS STRING),%d) AS x", rows-1, width),
					rows, int64(total),
				})
			}
		}
	}
	selected, mode := os.Getenv("NEBULA_STREAM_TEST_CASE"), os.Getenv("NEBULA_STREAM_TEST_MODE")
	if selected == "storage_16m" {
		cases = append(cases, timingCase{
			"storage_16m", "USE stream_bench MATCH (n@n) RETURN n.payload AS x ORDER BY n.id",
			8192, 16 << 20,
		})
	}
	if selected == "one_gib" {
		cases = append(cases, timingCase{
			"one_gib", "FOR i IN range(0,8191) RETURN i AS i ORDER BY i NEXT RETURN repeat(CAST(i % 10 AS STRING),131072) AS x",
			8192, 1 << 30,
		})
	}
	t.Logf("business consumption per decoded row: %d ns", consumeNs)
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
			// Validate full values before timing, without adding validation cost to samples.
			if round == 0 && os.Getenv("NEBULA_STREAM_TEST_BYTE_MATRIX") != "" {
				for _, streaming := range []bool{false, true} {
					var result types.Result
					var err error
					if streaming {
						result, err = client.(types.StreamingClient).StreamingExecute(context.Background(), withHints(c.query))
					} else {
						result, err = client.Execute(withHints(c.query))
					}
					require.NoError(t, err)
					n := 0
					for result.HasNext() {
						row, err := result.Next()
						require.NoError(t, err)
						value, err := row.GetValueByIndex(0)
						require.NoError(t, err)
						if strings.HasPrefix(c.name, "wide_") || strings.HasPrefix(c.name, "bytes_") || c.name == "one_gib" || c.name == "storage_16m" {
							first := 1
							if strings.HasPrefix(c.name, "bytes_") || c.name == "one_gib" || c.name == "storage_16m" {
								first = 0
							}
							actual, err := value.AsString()
							require.NoError(t, err)
							expected := strings.Repeat(strconv.Itoa((n+first)%10), int(c.logicalBytes/int64(c.rows)))
							require.Equal(t, expected, string(actual), "%s row %d", c.name, n)
						} else {
							require.Equal(t, strconv.Itoa(n+1), value.String(), "%s row %d", c.name, n)
						}
						n++
					}
					require.Equal(t, c.rows, n)
					if stream, ok := result.(types.StreamingResult); ok {
						require.NoError(t, stream.Err())
						require.NoError(t, stream.Close())
					}
				}
			}
			pairs := count
			if c.rows >= 100000 {
				pairs = min(pairs, 10)
			}
			if c.rows >= 1000000 {
				pairs = min(pairs, 5)
			}
			if largePairs > 0 && (c.rows >= 100000 || c.logicalBytes >= 1<<20) {
				pairs = largePairs
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
					s.LogicalBytes = c.logicalBytes
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
			s.LogicalBytes = 1 << 30
			s.Round = round
			s.Order = order
			if output != nil {
				require.NoError(t, output.Encode(s))
			}
			t.Logf("one_gib streaming=%v duration_us=%.3f mib_per_sec=%.2f", streaming, s.Elapsed, 1024e6/s.Elapsed)
		}
	}
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

func TestStreamingGraphConcurrentDelivery(t *testing.T) {
	value := os.Getenv("NEBULA_STREAM_TEST_CONCURRENT")
	if value == "" {
		t.Skip("concurrent delivery test is opt-in")
	}
	concurrency, err := strconv.Atoi(value)
	require.NoError(t, err)
	require.Contains(t, []int{0, 1, 16, 64}, concurrency)
	duration := time.Duration(0)
	if selected := os.Getenv("NEBULA_STREAM_TEST_CONCURRENT_DURATION"); selected != "" {
		duration, err = time.ParseDuration(selected)
		require.NoError(t, err)
		require.Greater(t, duration, time.Duration(0))
		require.LessOrEqual(t, duration, time.Minute)
	}
	if concurrency == 0 {
		require.Greater(t, duration, time.Duration(0), "unary baseline needs an observation window")
	}
	clients := make([]types.Client, concurrency)
	for i := range clients {
		clients[i] = streamingGraphClient(t)
	}
	observer := streamingGraphClient(t)
	expected := [10]string{}
	for digit := range expected {
		expected[digit] = strings.Repeat(strconv.Itoa(digit), 32768)
	}
	query := "FOR i IN range(0,1023) RETURN repeat(CAST(i % 10 AS STRING),32768) AS x"
	if hints := os.Getenv("NEBULA_STREAM_TEST_HINTS"); hints != "" {
		query = hints + " " + query
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	start := make(chan struct{})
	consume := func(client types.Client) error {
		result, err := client.(types.StreamingClient).StreamingExecute(ctx, query)
		if err != nil {
			return err
		}
		defer result.Close()
		n := 0
		for result.HasNext() {
			row, err := result.Next()
			if err != nil {
				return err
			}
			column, err := row.GetValueByIndex(0)
			if err != nil {
				return err
			}
			actual, err := column.AsString()
			if err != nil || string(actual) != expected[n%10] {
				return fmt.Errorf("row %d mismatch: %v", n, err)
			}
			n++
		}
		if result.Err() != nil {
			return result.Err()
		}
		if n != 1024 {
			return fmt.Errorf("read %d rows, expected 1024", n)
		}
		return result.Close()
	}
	type outcome struct {
		queries int
		err     error
	}
	done := make(chan outcome, concurrency)
	for _, client := range clients {
		go func(client types.Client) {
			<-start
			deadline := time.Now().Add(duration)
			queries := 0
			for {
				if err := consume(client); err != nil {
					done <- outcome{queries: queries, err: err}
					return
				}
				queries++
				if !time.Now().Before(deadline) {
					done <- outcome{queries: queries}
					return
				}
			}
		}(client)
	}
	stopObserver := make(chan struct{})
	type observation struct {
		latencies []float64
		err       error
	}
	observed := make(chan observation, 1)
	go func() {
		var samples []float64
		<-start
		for {
			select {
			case <-stopObserver:
				observed <- observation{latencies: samples}
				return
			default:
			}
			before := time.Now()
			result, err := observer.Execute("RETURN 1 AS x")
			if err != nil || !result.HasNext() {
				observed <- observation{err: fmt.Errorf("mixed unary query failed: %v", err)}
				return
			}
			_, err = result.Next()
			if err != nil {
				observed <- observation{err: err}
				return
			}
			samples = append(samples, float64(time.Since(before).Nanoseconds())/1000)
		}
	}()
	before := time.Now()
	close(start)
	completedQueries := 0
	if concurrency == 0 {
		time.Sleep(duration)
	}
	for range clients {
		result := <-done
		if result.err != nil {
			t.Error(result.err)
		}
		completedQueries += result.queries
	}
	elapsed := time.Since(before)
	close(stopObserver)
	observationResult := <-observed
	require.NoError(t, observationResult.err)
	require.NotEmpty(t, observationResult.latencies)
	sort.Float64s(observationResult.latencies)
	count := len(observationResult.latencies)
	t.Logf("streams=%d completed_queries=%d logical_mib=%d elapsed_seconds=%.6f aggregate_mib_per_sec=%.3f unary_samples=%d unary_p50_us=%.3f unary_p95_us=%.3f",
		concurrency, completedQueries, completedQueries*32, elapsed.Seconds(), float64(completedQueries*32)/elapsed.Seconds(), count,
		observationResult.latencies[count/2], observationResult.latencies[(95*count+99)/100-1])
}
