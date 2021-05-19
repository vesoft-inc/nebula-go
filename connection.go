/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package nebula_go

import (
	"fmt"
	"math"
	"time"

	"github.com/facebook/fbthrift/thrift/lib/go/thrift"
	"github.com/vesoft-inc/nebula-go/v2/nebula"
	"github.com/vesoft-inc/nebula-go/v2/nebula/graph"
)

type connection struct {
	severAddress HostAddress
	returnedAt   time.Time // the connection was created or returned.
	graph        *graph.GraphServiceClient
}

func newConnection(severAddress HostAddress) *connection {
	return &connection{
		severAddress: severAddress,
		returnedAt:   time.Now(),
		graph:        nil,
	}
}

func (cn *connection) open(hostAddress HostAddress, timeout time.Duration) error {
	ip := hostAddress.Host
	port := hostAddress.Port
	newAdd := fmt.Sprintf("%s:%d", ip, port)
	timeoutOption := thrift.SocketTimeout(timeout)
	bufferSize := 128 << 10
	frameMaxLength := uint32(math.MaxUint32)
	addressOption := thrift.SocketAddr(newAdd)
	sock, err := thrift.NewSocket(timeoutOption, addressOption)
	if err != nil {
		return fmt.Errorf("Failed to create a net.Conn-backed Transport,: %s", err.Error())
	}
	// Set transport buffer
	bufferedTranFactory := thrift.NewBufferedTransportFactory(bufferSize)
	transport := thrift.NewFramedTransportMaxLength(bufferedTranFactory.GetTransport(sock), frameMaxLength)
	pf := thrift.NewBinaryProtocolFactoryDefault()
	cn.graph = graph.NewGraphServiceClientFactory(transport, pf)
	if err = cn.graph.Open(); err != nil {
		return fmt.Errorf("Failed to open transport, error: %s", err.Error())
	}
	if !cn.graph.IsOpen() {
		return fmt.Errorf("Transport is off")
	}
	return nil
}

// Authenticate
func (cn *connection) authenticate(username, password string) (*graph.AuthResponse, error) {
	resp, err := cn.graph.Authenticate([]byte(username), []byte(password))
	if err != nil {
		err = fmt.Errorf("Authentication fails, %s", err.Error())
		if e := cn.graph.Close(); e != nil {
			err = fmt.Errorf("Fail to close transport, error: %s", e.Error())
		}
		return nil, err
	}
	if resp.ErrorCode != nebula.ErrorCode_SUCCEEDED {
		return nil, fmt.Errorf("Fail to authenticate, error: %s", resp.ErrorMsg)
	}
	return resp, err
}

func (cn *connection) execute(sessionID int64, stmt string) (*graph.ExecutionResponse, error) {
	return cn.graph.Execute(sessionID, []byte(stmt))
}

// unsupported
// func (client *GraphClient) ExecuteJson((sessionID int64, stmt string) (*graph.ExecutionResponse, error) {
// 	return cn.graph.ExecuteJson(sessionID, []byte(stmt))
// }

// Check connection to host address
func (cn *connection) ping() bool {
	_, err := cn.execute(0, "YIELD 1")
	return err == nil
}

// Sign out and release seesin ID
func (cn *connection) signOut(sessionID int64) error {
	// Release session ID to graphd
	return cn.graph.Signout(sessionID)
}

// Update returnedAt for cleaner
func (cn *connection) release() {
	cn.returnedAt = time.Now()
}

// Close transport
func (cn *connection) close() {
	cn.graph.Close()
}
