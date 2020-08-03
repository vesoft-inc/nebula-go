/* Copyright (c) 2019 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package ngdb

import (
	"fmt"
	"log"
	"time"

	"github.com/facebook/fbthrift/thrift/lib/go/thrift"
	graph "github.com/vesoft-inc/nebula-go/nebula/graph"
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
	sock, err := thrift.NewSocket(timeoutOption, addressOption)
	if err != nil {
		return nil, err
	}

	transport := thrift.NewBufferedTransport(sock, 128<<10)

	pf := thrift.NewBinaryProtocolFactoryDefault()
	graph := &GraphClient{
		graph: *graph.NewGraphServiceClientFactory(transport, pf),
	}
	return graph, nil
}

// Open transport and authenticate
func (client *GraphClient) Connect(username, password string) error {
	if err := client.graph.Transport.Open(); err != nil {
		return err
	}

	resp, err := client.graph.Authenticate(username, password)
	if err != nil {
		log.Printf("Authentication fails, %s", err.Error())
		if e := client.graph.Close(); e != nil {
			log.Printf("Fail to close transport, error: %s", e.Error())
		}
		return err
	}

	if resp.GetErrorCode() != graph.ErrorCode_SUCCEEDED {
		log.Printf("Authentication fails, ErrorCode: %v, ErrorMsg: %s", resp.GetErrorCode(), resp.GetErrorMsg())
		return fmt.Errorf(resp.GetErrorMsg())
	}

	client.sessionID = resp.GetSessionID()

	return nil
}

// Signout and close transport
func (client *GraphClient) Disconnect() {
	if err := client.graph.Signout(client.sessionID); err != nil {
		log.Printf("Fail to signout, error: %s", err.Error())
	}

	if err := client.graph.Close(); err != nil {
		log.Printf("Fail to close transport, error: %s", err.Error())
	}
}

func (client *GraphClient) Execute(stmt string) (*graph.ExecutionResponse, error) {
	return client.graph.Execute(client.sessionID, stmt)
}

func (client *GraphClient) GetSessionID() int64 {
	return client.sessionID
}

func IsError(resp *graph.ExecutionResponse) bool {
	return resp.GetErrorCode() != graph.ErrorCode_SUCCEEDED
}
