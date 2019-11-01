/* Copyright (c) 2019 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package nebula_client

import (
	"github.com/facebook/fbthrift/thrift/lib/go/thrift"
	"github.com/vesoft-inc/nebula-go/gen-go/nebula"
	"github.com/vesoft-inc/nebula-go/gen-go/nebula/storage"
)

type StorageClient struct {
	storage storage.StorageServiceClient
	option  Options
	session int64
}

func NewStorageClient(address string, opts ...Option) (client *StorageClient, err error) {
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
	storage := &StorageClient{
		storage: *storage.NewStorageServiceClientFactory(transport, protocol),
	}
	return storage, nil
}

func (client *StorageClient) Put(space nebula.GraphSpaceID, part nebula.PartitionID, key, value string) (*storage.ExecResponse, error) {
	pairs := make([]*nebula.Pair, 0)
	pairs = append(pairs, &nebula.Pair{Key: key, Value: value})
	parts := make(map[nebula.PartitionID][]*nebula.Pair)
	parts[part] = pairs
	request := storage.PutRequest{SpaceID: space, Parts: parts}
	return client.storage.Put(&request)
}

func (client *StorageClient) Get(space nebula.GraphSpaceID, part nebula.PartitionID, key string) (*storage.GeneralResponse, error) {
	parts := make(map[nebula.PartitionID][]string)
	keys := []string{key}
	parts[part] = keys
	request := storage.GetRequest{SpaceID: space, Parts: parts}
	return client.storage.Get(&request)
}
