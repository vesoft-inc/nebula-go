/*
 *
 * Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 *
 */

package nebula_go

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/facebook/fbthrift/thrift/lib/go/thrift"
	"github.com/vesoft-inc/nebula-go/v3/nebula"
	"github.com/vesoft-inc/nebula-go/v3/nebula/meta"
)

type metaClient struct {
	address             HostAddress
	latestSchemaVersion int64
	timeout             time.Duration
	meta                *meta.MetaServiceClient
}

func NewMetaClient(address HostAddress, timeout time.Duration) *metaClient {
	return &metaClient{
		address:             address,
		timeout:             timeout,
		latestSchemaVersion: -1,
		meta:                nil,
	}
}

// open the metaClient socket connection
func (client *metaClient) Open() error {
	ip := client.address.Host
	port := client.address.Port
	return client.doOpen(ip, port)
}

// do open operation for metaClient
func (client *metaClient) doOpen(ip string, port int) error {
	newAdd := net.JoinHostPort(ip, strconv.Itoa(port))

	sock, err := thrift.NewSocket(thrift.SocketAddr(newAdd), thrift.SocketTimeout(client.timeout))
	if err != nil {
		return fmt.Errorf("failed to create a net.Conn-backed Transport,: %s", err.Error())
	}

	bufferSize := 128 << 10
	bufferedTranFactory := thrift.NewBufferedTransportFactory(bufferSize)
	transport := thrift.NewHeaderTransport(bufferedTranFactory.GetTransport(sock))
	pf := thrift.NewHeaderProtocolFactory()

	client.meta = meta.NewMetaServiceClientFactory(transport, pf)
	if err = client.meta.Open(); err != nil {
		return fmt.Errorf("failed to open transport, error: %s", err.Error())
	}

	if !client.meta.IsOpen() {
		return fmt.Errorf("transport is off")
	}
	return client.verifyClientVersion()
}

// verify the compatibility of client and server
func (client *metaClient) verifyClientVersion() error {
	req := meta.NewVerifyClientVersionReq()
	resp, err := client.meta.VerifyClientVersion(req)
	if err != nil {
		client.close()
		return fmt.Errorf("failed to verify client version: %s", err.Error())
	}
	if resp.GetCode() != nebula.ErrorCode_SUCCEEDED {
		return fmt.Errorf("incompatible version between client and server: %s", string(resp.GetErrorMsg()))
	}
	return nil
}

// get all space names
func (client *metaClient) GetSpaces() ([]string, error) {
	req := meta.NewListSpacesReq()

	resp, err := client.meta.ListSpaces(req)
	if err != nil {
		return nil, err
	}

	// fresh metaClient with real leader and re-execute the request
	if resp.GetCode() == nebula.ErrorCode_E_LEADER_CHANGED {
		client.freshClient(resp.GetLeader())
		resp, err = client.meta.ListSpaces(req)
		if err != nil {
			return nil, err
		}
	}
	var spaces []string
	for _, name := range resp.GetSpaces() {
		spaces = append(spaces, string(name.GetName()))
	}
	return spaces, nil
}

// get one specific space info
func (client *metaClient) GetSpace(spaceName string) (*meta.SpaceItem, error) {
	req := meta.NewGetSpaceReq()
	req.SetSpaceName([]byte(spaceName))
	resp, err := client.meta.GetSpace(req)
	if err != nil {
		return nil, err
	}

	// fresh metaClient with real leader and re-execute the request
	if resp.GetCode() == nebula.ErrorCode_E_LEADER_CHANGED {
		client.freshClient(resp.GetLeader())
		resp, err = client.meta.GetSpace(req)
		if err != nil {
			return nil, err
		}
	}
	return resp.GetItem(), nil
}

// get all tag names of specific space name
func (client *metaClient) GetTags(spaceName string) ([]string, error) {
	req := meta.NewListTagsReq()
	spaceItem, err := client.GetSpace(spaceName)
	if err != nil {
		return nil, err
	}
	req.SetSpaceID(spaceItem.GetSpaceID())

	resp, err := client.meta.ListTags(req)
	if err != nil {
		return nil, err
	}

	// fresh metaClient with real leader and re-execute the request
	if resp.GetCode() == nebula.ErrorCode_E_LEADER_CHANGED {
		client.freshClient(resp.GetLeader())
		resp, err = client.meta.ListTags(req)
		if err != nil {
			return nil, err
		}
	}
	var tags []string
	for _, name := range resp.GetTags() {
		tags = append(tags, string(name.GetTagName()))
	}
	return tags, nil
}

// get schema of specifc tag
func (client *metaClient) GetTag(spaceName string, tag string) (*meta.Schema, error) {
	req := meta.NewGetTagReq()

	spaceItem, err := client.GetSpace(spaceName)
	if err != nil {
		return nil, err
	}
	req.SetSpaceID(spaceItem.GetSpaceID())
	req.SetTagName([]byte(tag))
	req.SetVersion(client.latestSchemaVersion)
	resp, err := client.meta.GetTag(req)
	if err != nil {
		return nil, err
	}

	// fresh metaClient with real leader and re-execute the request
	if resp.GetCode() == nebula.ErrorCode_E_LEADER_CHANGED {
		client.freshClient(resp.GetLeader())
		resp, err = client.meta.GetTag(req)
		if err != nil {
			return nil, err
		}
	}
	return resp.GetSchema(), nil
}

// get all edge names of specific space name
func (client *metaClient) GetEdges(spaceName string) ([]string, error) {
	req := meta.NewListEdgesReq()
	spaceItem, err := client.GetSpace(spaceName)
	if err != nil {
		return nil, err
	}
	req.SetSpaceID(spaceItem.GetSpaceID())

	resp, err := client.meta.ListEdges(req)
	if err != nil {
		return nil, err
	}

	// fresh metaClient with real leader and re-execute the request
	if resp.GetCode() == nebula.ErrorCode_E_LEADER_CHANGED {
		client.freshClient(resp.GetLeader())
		resp, err = client.meta.ListEdges(req)
		if err != nil {
			return nil, err
		}
	}
	var edges []string
	for _, name := range resp.GetEdges() {
		edges = append(edges, string(name.GetEdgeName()))
	}
	return edges, nil
}

// get schema of specifc tag
func (client *metaClient) GetEdge(spaceName string, edge string) (*meta.Schema, error) {
	req := meta.NewGetEdgeReq()

	spaceItem, err := client.GetSpace(spaceName)
	if err != nil {
		return nil, err
	}
	req.SetSpaceID(spaceItem.GetSpaceID())
	req.SetEdgeName([]byte(edge))
	req.SetVersion(client.latestSchemaVersion)

	resp, err := client.meta.GetEdge(req)
	if err != nil {
		return nil, err
	}

	// fresh metaClient with real leader and re-execute the request
	if resp.GetCode() == nebula.ErrorCode_E_LEADER_CHANGED {
		client.freshClient(resp.GetLeader())
		resp, err = client.meta.GetEdge(req)
		if err != nil {
			return nil, err
		}
	}
	return resp.GetSchema(), nil
}

// get the allocation of all parts.
func (client *metaClient) GetPartsAlloc(spaceName string) (map[nebula.PartitionID][]*nebula.HostAddr, error) {
	req := meta.NewGetPartsAllocReq()
	spaceItem, err := client.GetSpace(spaceName)
	if err != nil {
		return nil, err
	}
	req.SetSpaceID(spaceItem.GetSpaceID())
	resp, err := client.meta.GetPartsAlloc(req)
	if err != nil {
		return nil, err
	}

	// fresh metaClient with real leader and re-execute the request
	if resp.GetCode() == nebula.ErrorCode_E_LEADER_CHANGED {
		client.freshClient(resp.GetLeader())
		resp, err = client.meta.GetPartsAlloc(req)
		if err != nil {
			return nil, err
		}
	}
	if resp.GetCode() == nebula.ErrorCode_SUCCEEDED {
		return resp.GetParts(), nil
	} else {
		return nil, fmt.Errorf("GetPartsAlloc failed, code:%s", resp.GetCode())
	}
}

// get the leader host of all parts
func (client *metaClient) GetPartsLeader(spaceName string) (map[nebula.PartitionID]*nebula.HostAddr, error) {
	req := meta.NewListHostsReq()
	req.SetType(meta.ListHostType_ALLOC)

	resp, err := client.meta.ListHosts(req)
	if err != nil {
		return nil, err
	}

	// fresh metaClient with real leader and re-execute the request
	if resp.GetCode() == nebula.ErrorCode_E_LEADER_CHANGED {
		client.freshClient(resp.GetLeader())
		resp, err = client.meta.ListHosts(req)
		if err != nil {
			return nil, err
		}
	}

	if resp.GetCode() == nebula.ErrorCode_SUCCEEDED {
		hostItems := resp.GetHosts()
		var partLeaders map[nebula.PartitionID]*nebula.HostAddr
		for _, hostItem := range hostItems {
			parts := hostItem.GetLeaderParts()[spaceName]
			for _, part := range parts {
				partLeaders[part] = hostItem.GetHostAddr()
			}
		}
		return partLeaders, nil
	} else {
		return nil, fmt.Errorf("GetPartsLeader failed, code:%s", resp.GetCode())
	}
}

// list all storaged hosts
func (client *metaClient) ListStorageHosts() ([]*nebula.HostAddr, error) {
	req := meta.NewListHostsReq()
	req.SetType(meta.ListHostType_STORAGE)

	resp, err := client.meta.ListHosts(req)
	if err != nil {
		return nil, err
	}
	// fresh metaClient with real leader and re-execute the request
	if resp.GetCode() == nebula.ErrorCode_E_LEADER_CHANGED {
		client.freshClient(resp.GetLeader())
		resp, err = client.meta.ListHosts(req)
		if err != nil {
			return nil, err
		}
	}

	if resp.GetCode() == nebula.ErrorCode_SUCCEEDED {
		var hosts []*nebula.HostAddr
		for _, host := range resp.GetHosts() {
			hosts = append(hosts, host.GetHostAddr())
		}
		return hosts, nil
	} else {
		return nil, fmt.Errorf("ListStorageHosts failed, code:%s", resp.GetCode())
	}

}

// fresh the metaClient with meta leader host
func (client *metaClient) freshClient(leader *nebula.HostAddr) {
	client.close()
	ip := leader.Host
	port := leader.Port
	client.doOpen(ip, int(port))
}

// close the metaClient
func (client *metaClient) close() {
	client.meta.Close()
}
