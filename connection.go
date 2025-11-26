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
	"crypto/tls"
	"encoding/json"
	"math"
	"net"
	"sync"
	"time"

	"github.com/vesoft-inc/nebula-go/v5/internal/decode"
	"github.com/vesoft-inc/nebula-go/v5/internal/generated_code/v5.0.0/proto"
	"github.com/vesoft-inc/nebula-go/v5/internal/generated_code/v5.0.0/proto/common"
	"github.com/vesoft-inc/nebula-go/v5/internal/generated_code/v5.0.0/proto/graph"
	"github.com/vesoft-inc/nebula-go/v5/internal/grpcutil"
	"github.com/vesoft-inc/nebula-go/v5/internal/internal_error"
	"github.com/vesoft-inc/nebula-go/v5/pkg/errors"
	"github.com/vesoft-inc/nebula-go/v5/pkg/types"
	"github.com/vesoft-inc/nebula-go/v5/pkg/version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

var defaultConnector = &graphConnector{}
var defaultMsgSize = math.MaxInt64

const defaultPingTimeout = 1 * time.Second
const defaultCloseTimeout = 1 * time.Second

type graphConnector struct{}

type connection struct {
	mu          sync.Mutex
	graphClient graph.GraphServiceClient
	clientConn  *grpc.ClientConn
	sessionId   int64
	version     string
	timeout     time.Duration
	tlsConfig   *tls.Config
	address     string
}

func (c *graphConnector) connect(address string, cfg *connConfig) (types.Client, error) {
	cn := &connection{
		address: address,
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.connectTimeout)
	defer cancel()

	if cfg.enableTLS {
		if cfg.tlsConfig == nil {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return nil, err
			}
			tlsCfg, err := grpcutil.NewTLSConfig(host, cfg.ca, cfg.cert, cfg.key, cfg.insecureSkipVerify)
			if err != nil {
				return nil, err
			}
			cfg.tlsConfig = tlsCfg
		}
	}

	if err := cn.open(address, cfg.connectTimeout, cfg.tlsConfig); err != nil {
		return nil, err
	}
	if err := cn.authenticate(ctx, cfg.username, cfg.password, cfg.authInfo); err != nil {
		_ = cn.clientConn.Close()
		return nil, err
	}
	cn.timeout = cfg.requestTimeout
	return cn, nil
}

func (cn *connection) open(address string, timeout time.Duration, tlsCfg *tls.Config) error {
	grpcConn, err := grpcutil.NewGrpcClient(address, timeout, tlsCfg)
	if err != nil {
		return err
	}
	cn.clientConn = grpcConn
	cn.graphClient = graph.NewGraphServiceClient(grpcConn)
	return nil
}

func (cn *connection) authenticate(ctx context.Context, username, password string, authInfo map[string]string) error {
	d := grpcutil.GetCtxDuration(ctx)
	if authInfo == nil {
		authInfo = make(map[string]string)
		authInfo["password"] = password
	}

	bs, err := json.Marshal(authInfo)
	if err != nil {
		return err
	}
	clientInfo := &common.ClientInfo{
		Lang:            common.ClientInfo_GO,
		ProtocolVersion: proto.PROTOCOL_VERSION,
		Version:         []byte(version.ClientVersion),
	}
	in := graph.AuthRequest{
		Username:   []byte(username),
		AuthInfo:   bs,
		ClientInfo: clientInfo,
	}
	resp, err := cn.graphClient.Authenticate(ctx, &in)
	if err != nil {
		_ = cn.closeConn()
		return grpcutil.GetGrpcError(cn.address, err, d)
	}
	respErr := resp.GetStatus()
	if string(respErr.GetCode()) != string(errors.ERROR_SUCCESSFUL_COMPLETION) {
		_ = cn.closeConn()
		return internal_error.ErrServerResponse(string(respErr.GetCode()), string(respErr.GetMessage()))
	}
	cn.sessionId = resp.GetSessionId()

	cn.version = string(resp.GetVersion())
	return nil
}

func (cn *connection) Execute(stmt string) (types.Result, error) {
	if cn.timeout == 0 {
		return cn.ExecuteContext(context.Background(), stmt)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), cn.timeout)
		defer cancel()
		return cn.ExecuteContext(ctx, stmt)
	}
}

func (cn *connection) ExecuteContext(ctx context.Context, stmt string) (types.Result, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	d := grpcutil.GetCtxDuration(ctx)
	cn.mu.Lock()
	defer cn.mu.Unlock()
	in := &graph.ExecuteRequest{
		SessionId: cn.sessionId,
		Stmt:      []byte(stmt),
	}
	resp, err := cn.graphClient.Execute(ctx, in)
	if err != nil {
		return nil, grpcutil.GetGrpcError(cn.address, err, d)
	}
	t, err := decode.NewResultTable(resp.Result)
	if err != nil {
		return nil, internal_error.ErrDecodeFailed(err.Error())
	}
	resultResp := resultSet{
		index:   0,
		table:   t,
		summary: resp.Summary,
		cursor:  resp.Cursor,
	}
	if err := cn.isSucceed(resp); err != nil {
		return &resultResp, err
	}

	return &resultResp, nil
}

func (cn *connection) isSucceed(resp *graph.ExecuteResponse) error {
	respErr := resp.GetStatus()
	if string(respErr.GetCode()) != string(errors.ERROR_SUCCESSFUL_COMPLETION) {
		return internal_error.ErrServerResponse(string(respErr.GetCode()), string(respErr.GetMessage()))
	}
	return nil
}

func (cn *connection) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultPingTimeout)
	defer cancel()

	return cn.PingContext(ctx)

}

func (cn *connection) PingContext(ctx context.Context) error {
	d := grpcutil.GetCtxDuration(ctx)
	cn.mu.Lock()
	defer cn.mu.Unlock()
	stmt := []byte("RETURN 1")
	in := &graph.ExecuteRequest{
		SessionId: cn.sessionId,
		Stmt:      stmt,
	}
	resp, err := cn.graphClient.Execute(ctx, in)
	if err != nil {
		return grpcutil.GetGrpcError(cn.address, err, d)
	}
	if err := cn.isSucceed(resp); err != nil {
		return err
	}
	return nil
}

func (cn *connection) closeConn() error {
	return cn.clientConn.Close()
}

func (cn *connection) Close() error {
	cn.mu.Lock()
	defer cn.mu.Unlock()
	if cn.IsClosed() {
		return nil
	}
	// logout via statement, ignore the logout error
	in := &graph.ExecuteRequest{
		SessionId: cn.sessionId,
		Stmt:      []byte("SESSION CLOSE"),
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultCloseTimeout)
	defer cancel()
	_, err := cn.graphClient.Execute(ctx, in)
	_ = cn.closeConn()
	if err != nil {
		return grpcutil.GetGrpcError(cn.address, err, defaultCloseTimeout)
	}
	return nil
}

func (cn *connection) GetSessionId() (int64, error) {
	return cn.sessionId, nil
}

func (cn *connection) GetVersion() (string, error) {
	return cn.version, nil
}

func (cn *connection) IsClosed() bool {
	return cn.clientConn.GetState() == connectivity.Shutdown
}
