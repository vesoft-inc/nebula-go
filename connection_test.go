// Copyright 2025 vesoft inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// 
//     http://www.apache.org/licenses/LICENSE-2.0
// 
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package nebula_ng

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vesoft-inc/nebula-go/v5/internal/generated_code/v5.0.0/proto/common"
	"github.com/vesoft-inc/nebula-go/v5/internal/generated_code/v5.0.0/proto/graph"
	"github.com/vesoft-inc/nebula-go/v5/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type dummyGRPConn struct {
	authResp *graph.AuthResponse
	resp     *graph.ExecuteResponse
	err      *grpcError
}

type grpcError struct {
	errMsg string
	code   codes.Code
}

func (e *grpcError) Error() string {
	return e.errMsg
}

func (e *grpcError) GRPCStatus() *status.Status {
	return status.New(e.code, e.errMsg)
}

var _ graph.GraphServiceClient = &dummyGRPConn{}

func (d *dummyGRPConn) Authenticate(ctx context.Context, in *graph.AuthRequest, opts ...grpc.CallOption) (*graph.AuthResponse, error) {
	if d.err != nil {
		return nil, d.err
	}
	return d.authResp, nil
}

func (d *dummyGRPConn) Execute(ctx context.Context, in *graph.ExecuteRequest, opts ...grpc.CallOption) (*graph.ExecuteResponse, error) {
	if d.err != nil {
		return nil, d.err
	}
	return d.resp, nil
}
func (d *dummyGRPConn) StreamingExecute(ctx context.Context, in *graph.ExecuteRequest, opts ...grpc.CallOption) (graph.GraphService_StreamingExecuteClient, error) {
	return nil, nil
}

func TestPingError(t *testing.T) {
	c := &connection{
		address: "localhost:9669",
	}
	testcases := []struct {
		grpcCode    codes.Code
		grpcErrMsg  string
		respCode    errors.ErrorCode
		respErrMsg  string
		expectedMsg string
	}{
		{
			grpcCode:    codes.OK,
			respCode:    errors.ERROR_INVALID_SYNTAX,
			respErrMsg:  "syntax error",
			expectedMsg: "[42001]: syntax error",
		},
		// catch the error, and format the error message
		// by default, ping timeout is 1 sec
		{
			grpcCode:    codes.DeadlineExceeded,
			grpcErrMsg:  "deadline exceeded",
			expectedMsg: "[99004]: request to localhost:9669 timeout after 1000ms, deadline exceeded",
		},
		// not catch the error
		{
			grpcCode:    codes.OutOfRange,
			grpcErrMsg:  "out of range",
			expectedMsg: "out of range",
		},
	}
	for _, tc := range testcases {
		grpcConn := &dummyGRPConn{}
		if tc.grpcCode != codes.OK {
			grpcConn.err = &grpcError{errMsg: tc.grpcErrMsg, code: tc.grpcCode}
		} else {
			grpcConn.resp = &graph.ExecuteResponse{
				Status: &common.Status{
					Code:    []byte(tc.respCode),
					Message: []byte(tc.respErrMsg),
				},
			}
		}
		c.graphClient = grpcConn
		err := c.Ping()
		if err == nil {
			t.Fatal("expected error")
		}
		assert.Equal(t, tc.expectedMsg, err.Error())
	}
}
