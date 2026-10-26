// Copyright 2026 vesoft inc. All Rights Reserved.
package nebula_ng

import (
	"context"
	"encoding/binary"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vesoft-inc/nebula-go/v5/internal/generated_code/v5.0.0/proto/common"
	"github.com/vesoft-inc/nebula-go/v5/internal/generated_code/v5.0.0/proto/graph"
	"github.com/vesoft-inc/nebula-go/v5/internal/generated_code/v5.0.0/proto/vector"
	"github.com/vesoft-inc/nebula-go/v5/pkg/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type streamTestServer struct {
	graph.UnimplementedGraphServiceServer
	stream func(*graph.ExecuteRequest, graph.GraphService_StreamingExecuteServer) error
}

func (s *streamTestServer) StreamingExecute(req *graph.ExecuteRequest, stream graph.GraphService_StreamingExecuteServer) error {
	return s.stream(req, stream)
}
func (s *streamTestServer) Execute(context.Context, *graph.ExecuteRequest) (*graph.ExecuteResponse, error) {
	return &graph.ExecuteResponse{Status: &common.Status{Code: []byte("00000")}}, nil
}
func streamTestConnection(t *testing.T, serve func(*graph.ExecuteRequest, graph.GraphService_StreamingExecuteServer) error) *connection {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	graph.RegisterGraphServiceServer(server, &streamTestServer{stream: serve})
	go func() { _ = server.Serve(listener) }()
	cc, err := grpc.NewClient("passthrough:///test", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }))
	require.NoError(t, err)
	t.Cleanup(func() { _ = cc.Close(); server.Stop(); _ = listener.Close() })
	return &connection{graphClient: graph.NewGraphServiceClient(cc), clientConn: cc, sessionId: 1}
}
func streamHeader() *graph.StreamingExecuteResponse {
	return &graph.StreamingExecuteResponse{Response: &graph.StreamingExecuteResponse_Metadata{Metadata: &vector.VectorTableMetaData{TableType: 1, IsLittleEndian: true, RowType: &vector.RowType{NumColumns: 1, ColumnNames: []string{"x"}, ColumnTypes: []*vector.ValueType{{ValueType: []byte{0xb}}}}}}}
}
func streamBatch(values ...int64) *graph.StreamingExecuteResponse {
	data := make([]byte, len(values)*8)
	for i, v := range values {
		binary.LittleEndian.PutUint64(data[i*8:], uint64(v))
	}
	return &graph.StreamingExecuteResponse{Response: &graph.StreamingExecuteResponse_Batch{Batch: &vector.VectorBatch{Vectors: []*vector.NestedVector{{CommonMetaData: &vector.VectorCommonMetaData{NumRecords: uint32(len(values)), VectorContentType: 0x102}, VectorData: data}}}}}
}
func streamFinal(code string) *graph.StreamingExecuteResponse {
	return &graph.StreamingExecuteResponse{Response: &graph.StreamingExecuteResponse_Final{Final: &graph.ExecuteFinal{Status: &common.Status{Code: []byte(code), Message: []byte("query failed")}, Summary: &graph.Summary{}, Cursor: []byte("cursor")}}}
}

func TestStreamingMetadataBeforeRowsAndReuse(t *testing.T) {
	produce := make(chan struct{})
	cn := streamTestConnection(t, func(_ *graph.ExecuteRequest, s graph.GraphService_StreamingExecuteServer) error {
		if err := s.Send(streamHeader()); err != nil {
			return err
		}
		<-produce
		for _, m := range []*graph.StreamingExecuteResponse{streamBatch(1, 2), streamBatch(3), streamFinal("00000")} {
			if err := s.Send(m); err != nil {
				return err
			}
		}
		return nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	rs, err := cn.StreamingExecute(ctx, "RETURN 1")
	require.NoError(t, err)
	require.Equal(t, []string{"x"}, rs.Columns())
	require.Equal(t, -1, rs.RowSize())
	require.Nil(t, rs.Summary())
	_, err = cn.ExecuteContext(ctx, "RETURN 2")
	require.ErrorIs(t, err, errActiveStream)
	close(produce)
	require.Equal(t, []string{"x"}, rs.Columns())
	var got []int64
	for rs.HasNext() {
		row, err := rs.Next()
		require.NoError(t, err)
		value, err := row.GetValueByName("x")
		require.NoError(t, err)
		n, err := value.AsInt64()
		require.NoError(t, err)
		got = append(got, int64(n))
	}
	require.Equal(t, []int64{1, 2, 3}, got)
	require.NoError(t, rs.Err())
	require.NotNil(t, rs.Summary())
	require.Equal(t, []byte("cursor"), rs.Cursor())
	require.NoError(t, rs.Close())
	require.NoError(t, rs.Close())
	_, err = cn.ExecuteContext(ctx, "RETURN 2")
	require.NoError(t, err)
}

func TestStreamingExecuteWaitsForFirstResponseMessage(t *testing.T) {
	produce := make(chan struct{})
	started := make(chan struct{})
	cn := streamTestConnection(t, func(_ *graph.ExecuteRequest, s graph.GraphService_StreamingExecuteServer) error {
		if err := s.SendHeader(metadata.Pairs("nebula-stream-accepted", "1")); err != nil {
			return err
		}
		close(started)
		select {
		case <-produce:
		case <-s.Context().Done():
			return s.Context().Err()
		}
		if err := s.Send(streamHeader()); err != nil {
			return err
		}
		return s.Send(streamFinal("00000"))
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	done := make(chan struct{})
	var result types.StreamingResult
	var executeErr error
	go func() {
		result, executeErr = cn.StreamingExecute(ctx, "query")
		close(done)
	}()
	<-started
	select {
	case <-done:
		t.Fatal("StreamingExecute returned before the first response message")
	case <-time.After(100 * time.Millisecond):
	}
	close(produce)
	<-done
	require.NoError(t, executeErr)
	require.Equal(t, []string{"x"}, result.Columns())
	require.False(t, result.HasNext())
	require.NoError(t, result.Err())
}

func TestStreamingRejectsIncompleteOrMalformedResults(t *testing.T) {
	for _, test := range []struct {
		name     string
		messages []*graph.StreamingExecuteResponse
		rpcErr   error
	}{
		{"missing final", []*graph.StreamingExecuteResponse{streamHeader(), streamBatch(1)}, nil},
		{"error final", []*graph.StreamingExecuteResponse{streamHeader(), streamBatch(1), streamFinal("42000")}, nil},
		{"transport after final", []*graph.StreamingExecuteResponse{streamFinal("00000")}, status.Error(codes.Unavailable, "lost")},
		{"message after final", []*graph.StreamingExecuteResponse{streamFinal("00000"), streamHeader()}, nil},
		{"duplicate metadata", []*graph.StreamingExecuteResponse{streamHeader(), streamHeader()}, nil},
		{"batch before metadata", []*graph.StreamingExecuteResponse{streamBatch(1)}, nil},
		{"empty batch", []*graph.StreamingExecuteResponse{streamHeader(), streamBatch()}, nil},
		{"metadata without column types", []*graph.StreamingExecuteResponse{{Response: &graph.StreamingExecuteResponse_Metadata{Metadata: &vector.VectorTableMetaData{RowType: &vector.RowType{NumColumns: 1, ColumnNames: []string{"x"}}}}}}, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			cn := streamTestConnection(t, func(_ *graph.ExecuteRequest, s graph.GraphService_StreamingExecuteServer) error {
				_ = s.SendHeader(metadata.Pairs("nebula-stream-accepted", "1"))
				for _, m := range test.messages {
					if err := s.Send(m); err != nil {
						return err
					}
				}
				return test.rpcErr
			})
			rs, err := cn.StreamingExecute(context.Background(), "query")
			if test.name == "transport after final" || test.name == "message after final" || test.name == "batch before metadata" || test.name == "metadata without column types" {
				require.Error(t, err)
				require.Nil(t, rs)
				require.Nil(t, cn.activeStream)
				return
			}
			require.NoError(t, err)
			for _, err := range rs.All() {
				if err != nil {
					break
				}
			}
			require.Error(t, rs.Err())
			require.Nil(t, rs.Summary())
			require.NoError(t, rs.Close())
		})
	}
}

func TestStreamingValidationAndUnsupported(t *testing.T) {
	for _, unsupported := range []bool{false, true} {
		cn := streamTestConnection(t, func(_ *graph.ExecuteRequest, s graph.GraphService_StreamingExecuteServer) error {
			if unsupported {
				return status.Error(codes.Unimplemented, "old server")
			}
			return s.Send(streamFinal("42000"))
		})
		rs, err := cn.StreamingExecute(context.Background(), "bad")
		require.Error(t, err)
		require.Nil(t, rs)
		require.Nil(t, cn.activeStream)
	}
}

func TestStreamingCloseInterruptsReceive(t *testing.T) {
	started := make(chan struct{})
	var once sync.Once
	cn := streamTestConnection(t, func(_ *graph.ExecuteRequest, s graph.GraphService_StreamingExecuteServer) error {
		_ = s.SendHeader(metadata.Pairs("nebula-stream-accepted", "1"))
		_ = s.Send(streamHeader())
		once.Do(func() { close(started) })
		<-s.Context().Done()
		return s.Context().Err()
	})
	rs, err := cn.StreamingExecute(context.Background(), "query")
	require.NoError(t, err)
	<-started
	done := make(chan bool)
	go func() { done <- rs.HasNext() }()
	require.NoError(t, rs.Close())
	require.NoError(t, rs.Close())
	select {
	case value := <-done:
		require.False(t, value)
	case <-time.After(time.Second):
		t.Fatal("Close did not interrupt Recv")
	}
	require.Error(t, rs.Err())
	_, err = rs.Next()
	require.NotEqual(t, io.EOF, err)
}

func TestStreamingRowsOutliveBatchesAndScanCrossesBatches(t *testing.T) {
	cn := streamTestConnection(t, func(_ *graph.ExecuteRequest, s graph.GraphService_StreamingExecuteServer) error {
		for _, m := range []*graph.StreamingExecuteResponse{streamHeader(), streamBatch(1), streamBatch(2, 3), streamBatch(4), streamFinal("00000")} {
			if err := s.Send(m); err != nil {
				return err
			}
		}
		return nil
	})
	rs, err := cn.StreamingExecute(context.Background(), "query")
	require.NoError(t, err)
	first, err := rs.Next()
	require.NoError(t, err)
	var got []int
	for {
		var x int
		if err := rs.Scan(&x); err != nil {
			require.ErrorIs(t, err, io.EOF)
			break
		}
		got = append(got, x)
	}
	require.Equal(t, []int{2, 3, 4}, got)
	require.NoError(t, rs.Err())
	value, err := first.GetValueByName("x")
	require.NoError(t, err)
	n, err := value.AsInt64()
	require.NoError(t, err)
	require.EqualValues(t, 1, n)
	_, err = first.GetValueByName("y")
	require.Error(t, err)
}

func TestStreamingActiveStreamBlocksPingAndPoolReturn(t *testing.T) {
	cn := streamTestConnection(t, func(_ *graph.ExecuteRequest, s graph.GraphService_StreamingExecuteServer) error {
		if err := s.Send(streamHeader()); err != nil {
			return err
		}
		<-s.Context().Done()
		return s.Context().Err()
	})
	rs, err := cn.StreamingExecute(context.Background(), "query")
	require.NoError(t, err)
	require.ErrorIs(t, cn.PingContext(context.Background()), errActiveStream)
	dc := &driverConn{conn: cn, createAt: time.Now()}
	pool := &driverPool{connMap: map[types.Client]struct{}{dc: {}}, maxOpen: 1}
	// A rejected client stays with the caller; the pool is unchanged.
	require.ErrorIs(t, pool.PutClient(dc), errActiveStream)
	require.Empty(t, pool.freeConn)
	require.NoError(t, rs.Close())
	require.NoError(t, cn.PingContext(context.Background()))
	require.NoError(t, pool.PutClient(dc))
	require.Equal(t, []types.Client{dc}, pool.freeConn)
}
