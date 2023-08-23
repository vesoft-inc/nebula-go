/*
 *
 * Copyright (c) 2023 vesoft inc. All rights reserved.
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
	// Process domain to IP
	var addresses []HostAddress
	addresses = append(addresses, client.address)
	convAddress, err := DomainToIP(addresses)
	if err != nil {
		return fmt.Errorf("failed to find IP, error: %s ", err.Error())
	}

	// Check input
	if len(convAddress) == 0 {
		return fmt.Errorf("failed to initialize connection pool: illegal address input %s:%d", client.address.Host, client.address.Port)
	}

	ip := convAddress[0].Host
	port := convAddress[0].Port
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
		client.Close()
		return fmt.Errorf("failed to verify client version: %s", err.Error())
	}
	if resp.GetCode() != nebula.ErrorCode_SUCCEEDED {
		client.Close()
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
	if resp.GetCode() == nebula.ErrorCode_SUCCEEDED {
		var spaces []string
		for _, name := range resp.GetSpaces() {
			spaces = append(spaces, string(name.GetName()))
		}
		return spaces, nil
	} else {
		return nil, fmt.Errorf("GetSpaces failed, code:%s", resp.GetCode())
	}
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
	if resp.GetCode() == nebula.ErrorCode_SUCCEEDED {
		return resp.GetItem(), nil
	} else {
		return nil, fmt.Errorf("GetSpace failed, code:%s", resp.GetCode())
	}
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
	if resp.GetCode() == nebula.ErrorCode_SUCCEEDED {
		var tags []string
		for _, name := range resp.GetTags() {
			tags = append(tags, string(name.GetTagName()))
		}
		return tags, nil
	} else {
		return nil, fmt.Errorf("GetTags failed, code:%s", resp.GetCode())
	}
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
	if resp.GetCode() == nebula.ErrorCode_SUCCEEDED {
		return resp.GetSchema(), nil
	} else {
		return nil, fmt.Errorf("GetTag failed, code:%s", resp.GetCode())
	}
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
func (client *metaClient) GetPartsAlloc(spaceName string) (map[nebula.PartitionID][]HostAddress, error) {
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
	partsAlloc := make(map[nebula.PartitionID][]HostAddress)
	if resp.GetCode() == nebula.ErrorCode_SUCCEEDED {
		partsAllocMap := resp.GetParts()
		for part, hosts := range partsAllocMap {
			var allocHosts []HostAddress
			for _, host := range hosts {
				allocHosts = append(allocHosts, HostAddress{host.GetHost(), int(host.GetPort())})
			}
			partsAlloc[part] = allocHosts
		}
		return partsAlloc, nil

	} else {
		return nil, fmt.Errorf("GetPartsAlloc failed, code:%s", resp.GetCode())
	}
}

// get the leader host of all parts
func (client *metaClient) GetPartsLeader(spaceName string) (map[nebula.PartitionID]HostAddress, error) {
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
		partLeaders := make(map[nebula.PartitionID]HostAddress)
		for _, hostItem := range hostItems {
			parts := hostItem.GetLeaderParts()[spaceName]
			for _, part := range parts {
				partLeaders[part] = HostAddress{Host: hostItem.GetHostAddr().GetHost(), Port: int(hostItem.GetHostAddr().GetPort())}
			}
		}
		return partLeaders, nil
	} else {
		return nil, fmt.Errorf("GetPartsLeader failed, code:%s", resp.GetCode())
	}
}

// list all storaged hosts
func (client *metaClient) ListStorageHosts() ([]HostAddress, error) {
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
		var hosts []HostAddress
		for _, host := range resp.GetHosts() {
			hosts = append(hosts, HostAddress{Host: host.GetHostAddr().GetHost(), Port: int(host.GetHostAddr().GetPort())})
		}
		return hosts, nil
	} else {
		return nil, fmt.Errorf("ListStorageHosts failed, code:%s", resp.GetCode())
	}

}

// fresh the metaClient with meta leader host
func (client *metaClient) freshClient(leader *nebula.HostAddr) {
	client.Close()
	ip := leader.Host
	port := leader.Port
	client.doOpen(ip, int(port))
}

// close the metaClient
func (client *metaClient) Close() {
	if client.meta != nil && client.meta.IsOpen() {
		client.meta.Close()
	}
}
