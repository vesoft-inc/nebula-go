/* Copyright (c) 2019 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package ngdb

import (
	"log"
	"time"

	"github.com/facebook/fbthrift/thrift/lib/go/thrift"
	"github.com/vesoft-inc/nebula-go/graph"
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
	transport, err := thrift.NewSocket(timeoutOption, addressOption)
	if err != nil {
		return nil, err
	}

	protocol := thrift.NewBinaryProtocolFactoryDefault()
	graph := &GraphClient{
		graph: *graph.NewGraphServiceClientFactory(transport, protocol),
	}
	return graph, nil
}

// Open transport and authenticate
func (client *GraphClient) Connect(username, password string) error {
	if err := client.graph.Transport.Open(); err != nil {
		return err
	}

	if resp, err := client.graph.Authenticate(username, password); err != nil {
		log.Printf("Authentication fails, ErrorCode: %v, ErrorMsg: %s", resp.GetErrorCode(), resp.GetErrorMsg())
		if e := client.graph.Close(); e != nil {
			log.Printf("Fail to close transport, error: %s", e.Error())
		}
		return err
	} else {
		client.sessionID = resp.GetSessionID()
		return nil
	}
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
