// Copyright 2026 vesoft inc. All Rights Reserved.
package nebula_ng

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"regexp"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/vesoft-inc/nebula-go/v5/internal/decode"
	"github.com/vesoft-inc/nebula-go/v5/internal/generated_code/v5.0.0/proto/common"
	"github.com/vesoft-inc/nebula-go/v5/internal/generated_code/v5.0.0/proto/graph"
	"github.com/vesoft-inc/nebula-go/v5/internal/generated_code/v5.0.0/proto/vector"
	"github.com/vesoft-inc/nebula-go/v5/internal/internal_error"
	"github.com/vesoft-inc/nebula-go/v5/pkg/types"
)

// Authenticate returns the release name, optionally with a development suffix.
var streamingVersionPattern = regexp.MustCompile(`^[vV]?([0-9]+)\.([0-9]+)(\.[0-9]+)?(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)

func supportsStreaming(version string) bool {
	parts := streamingVersionPattern.FindStringSubmatch(version)
	if parts == nil {
		return false
	}
	major, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return false
	}
	minor, err := strconv.ParseUint(parts[2], 10, 64)
	return err == nil && (major > 5 || (major == 5 && minor >= 4))
}

var errActiveStream = errors.New("the client has an unread streaming result; consume or close it first")

// readMu serializes decoding. Close only uses cancellation and stateMu, so another goroutine
// can interrupt Recv without waiting for readMu. There is no background receive loop.
type streamingResult struct {
	resultSet
	readMu      sync.Mutex
	stateMu     sync.Mutex
	stream      graph.GraphService_StreamingExecuteClient
	metadata    *vector.VectorTableMetaData
	cancel      context.CancelFunc
	release     func()
	releaseOnce sync.Once
	closed      atomic.Bool
	done        atomic.Bool
	err         error
}

var _ types.StreamingResult = (*streamingResult)(nil)
var _ types.StreamingClient = (*driverConn)(nil)
var _ types.StreamingClient = (*connection)(nil)

func (cn *connection) StreamingExecute(ctx context.Context, stmt string) (types.StreamingResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !supportsStreaming(cn.version) {
		return nil, fmt.Errorf("streaming execution requires NebulaGraph 5.4 or later; server version is %q", cn.version)
	}
	ctx, cancel := context.WithCancel(ctx)
	rs := &streamingResult{cancel: cancel}
	rs.release = func() {
		cn.mu.Lock()
		if cn.activeStream == rs {
			cn.activeStream = nil
		}
		cn.mu.Unlock()
	}
	cn.mu.Lock()
	if cn.activeStream != nil {
		cn.mu.Unlock()
		cancel()
		return nil, errActiveStream
	}
	cn.activeStream = rs
	cn.mu.Unlock()
	stream, err := cn.graphClient.StreamingExecute(ctx, &graph.ExecuteRequest{SessionId: cn.sessionId, Stmt: []byte(stmt)})
	if err != nil {
		rs.finish(err)
		return nil, err
	}
	rs.stream = stream
	if err = rs.readResponse(); err != nil {
		return nil, err
	}
	return rs, nil
}

func streamStatusError(status *common.Status) error {
	if status == nil {
		return errors.New("stream final has no status")
	}
	if string(status.Code) != "00000" {
		return internal_error.ErrServerResponse(string(status.Code), string(status.Message))
	}
	return nil
}

func (rs *streamingResult) finish(err error) {
	rs.stateMu.Lock()
	if !rs.done.Load() {
		rs.err = err
		rs.done.Store(true)
	}
	rs.stateMu.Unlock()
	rs.releaseOnce.Do(func() {
		rs.cancel()
		if rs.release != nil {
			rs.release()
		}
	})
}

func (rs *streamingResult) Err() error {
	rs.stateMu.Lock()
	defer rs.stateMu.Unlock()
	return rs.err
}

func (rs *streamingResult) Close() error {
	rs.closed.Store(true)
	rs.finish(context.Canceled)
	return nil
}

func (rs *streamingResult) RowSize() int { return -1 }

func (rs *streamingResult) readResponse() error {
	response, err := rs.stream.Recv()
	if err != nil {
		if err == io.EOF {
			err = errors.New("stream ended before final; result is incomplete")
		}
		rs.finish(err)
		return err
	}
	switch part := response.Response.(type) {
	case *graph.StreamingExecuteResponse_Metadata:
		if rs.metadata != nil || part.Metadata == nil || part.Metadata.RowType == nil {
			err = errors.New("invalid or duplicate streaming metadata")
			break
		}
		if part.Metadata.NumRecords != 0 || part.Metadata.NumBatches != 0 {
			err = errors.New("streaming metadata contains result totals")
			break
		}
		rs.metadata = part.Metadata
		rs.table, err = decode.NewResultTable(&vector.VectorResultTable{DataLayoutVersion: []byte("5.0.0"), Meta: rs.metadata})
	case *graph.StreamingExecuteResponse_Batch:
		if rs.metadata == nil || part.Batch == nil || len(part.Batch.Vectors) != len(rs.metadata.RowType.ColumnNames) {
			err = errors.New("streaming batch does not match metadata")
			break
		}
		rows := part.Batch.NumRecords
		if len(part.Batch.Vectors) > 0 {
			rows = part.Batch.Vectors[0].GetCommonMetaData().GetNumRecords()
			for _, v := range part.Batch.Vectors {
				if v.GetCommonMetaData().GetNumRecords() != rows {
					err = errors.New("inconsistent streaming vector row counts")
					break
				}
			}
		}
		if err != nil {
			break
		}
		if rows == 0 {
			err = errors.New("stream contains an empty batch")
			break
		}
		meta := &vector.VectorTableMetaData{
			TableType: rs.metadata.TableType, RowType: rs.metadata.RowType,
			NumRecords: uint64(rows), NumBatches: 1, TimeZoneOffset: rs.metadata.TimeZoneOffset,
			IsLittleEndian: rs.metadata.IsLittleEndian, GraphSchema: rs.metadata.GraphSchema,
		}
		rs.table, err = decode.NewResultTable(&vector.VectorResultTable{DataLayoutVersion: []byte("5.0.0"), Meta: meta, Batch: []*vector.VectorBatch{part.Batch}})
		rs.index = 0
	case *graph.StreamingExecuteResponse_Final:
		err = streamStatusError(part.Final.GetStatus())
		_, endErr := rs.stream.Recv()
		if endErr != io.EOF {
			if endErr == nil {
				endErr = errors.New("stream has a response after final")
			}
			if err == nil {
				err = fmt.Errorf("stream did not end normally: %w", endErr)
			}
		}
		if err == nil {
			rs.stateMu.Lock()
			if !rs.done.Load() {
				rs.summary = part.Final.GetSummary()
				rs.cursor = part.Final.GetCursor()
			}
			rs.stateMu.Unlock()
		}
		rs.table = nil
		rs.finish(err)
		return err
	default:
		err = errors.New("stream response has no recognized payload")
	}
	if err != nil {
		rs.finish(err)
	}
	return err
}

func (rs *streamingResult) hasNext() bool {
	for {
		if rs.done.Load() || rs.closed.Load() {
			return false
		}
		if rs.resultSet.HasNext() {
			return true
		}
		rs.table = nil
		if rs.readResponse() != nil {
			return false
		}
	}
}

func (rs *streamingResult) HasNext() bool {
	rs.readMu.Lock()
	defer rs.readMu.Unlock()
	return rs.hasNext()
}

func (rs *streamingResult) Next() (types.Row, error) {
	rs.readMu.Lock()
	defer rs.readMu.Unlock()
	if !rs.hasNext() {
		if err := rs.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}
	row, err := rs.resultSet.Next()
	if row != nil {
		row.(*rowData).columnNames = rs.metadata.RowType.ColumnNames
	}
	if err != nil {
		rs.finish(err)
	}
	return row, err
}

func (rs *streamingResult) Scan(dest ...any) error {
	rs.readMu.Lock()
	defer rs.readMu.Unlock()
	if !rs.hasNext() {
		if err := rs.Err(); err != nil {
			return err
		}
		return io.EOF
	}
	err := rs.resultSet.Scan(dest...)
	if err != nil {
		rs.finish(err)
	}
	return err
}

func (rs *streamingResult) All() iter.Seq2[types.Row, error] {
	return func(yield func(types.Row, error) bool) {
		for rs.HasNext() {
			row, err := rs.Next()
			if !yield(row, err) || err != nil {
				return
			}
		}
		if err := rs.Err(); err != nil {
			yield(nil, err)
		}
	}
}

func (rs *streamingResult) readMetadata() {
	for rs.metadata == nil {
		if rs.done.Load() || rs.closed.Load() || rs.readResponse() != nil {
			return
		}
	}
}

func (rs *streamingResult) Columns() []string {
	rs.readMu.Lock()
	defer rs.readMu.Unlock()
	rs.readMetadata()
	if rs.metadata == nil {
		return nil
	}
	return rs.metadata.RowType.ColumnNames
}

func (rs *streamingResult) ColumnTypes() []types.ColumnType {
	rs.readMu.Lock()
	defer rs.readMu.Unlock()
	rs.readMetadata()
	if rs.metadata == nil {
		return nil
	}
	table, err := decode.NewResultTable(&vector.VectorResultTable{Meta: rs.metadata})
	if err != nil {
		rs.finish(err)
		return nil
	}
	return table.ColumnTypes()
}

func (rs *streamingResult) Summary() types.Summary {
	rs.stateMu.Lock()
	defer rs.stateMu.Unlock()
	return rs.resultSet.Summary()
}

func (rs *streamingResult) Cursor() []byte {
	rs.stateMu.Lock()
	defer rs.stateMu.Unlock()
	return rs.resultSet.Cursor()
}
