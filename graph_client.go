/* Copyright (c) 2019 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package nebula_client

import (
	"github.com/facebook/fbthrift/thrift/lib/go/thrift"
	"github.com/vesoft-inc/nebula-go/gen-go/nebula/graph"
	"log"
)

type GraphClient struct {
	graph   graph.GraphServiceClient
	option  Options
	session int64
}

func NewGraphClient(address string, opts ...Option) (client *GraphClient, err error) {
	options := defaultOptions
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
		// TODO(yee): Check whether to close transport
		log.Printf("ErrorCode: %v, ErrorMsg: %s", resp.GetErrorCode(), resp.GetErrorMsg())
		return err
	} else {
		client.session = resp.GetSessionID()
		return nil
	}
}

// Signout and close transport
func (client *GraphClient) Disconnect() {
	if err := client.graph.Signout(client.session); err != nil {
		log.Println("Fail to signout")
	}

	if err := client.graph.Close(); err != nil {
		log.Println("Fail to close transport")
	}
}

func (client *GraphClient) Execute(statement string) (*graph.ExecutionResponse, error) {
	return client.graph.Execute(client.session, statement)
}
