/* Copyright (c) 2023 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 */

package nebula_go

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var ip = "127.0.0.1"
var metaPort = 9559
var graphPort = 3699
var metaAddress = HostAddress{ip, metaPort}
var client = NewMetaClient(metaAddress, 2*time.Second)
var isOpen = false

var spaceName = "test_meta"

func setup() {
	mockSchema()
	openClient()
}

func openClient() {
	if !isOpen {
		err := client.Open()
		if err != nil {
			log.Fatal(fmt.Sprintf("open meta client failed, host:%s, port:%d, %s", ip, metaPort, err.Error()))
		}
		isOpen = true
	}
}

func mockSchema() {
	// create schema in NebulaGraph
	hostAddress := HostAddress{Host: ip, Port: graphPort}
	hostList := []HostAddress{hostAddress}
	// Create configs for connection pool using default values
	testPoolConfig := GetDefaultConf()

	// Initialize connection pool
	pool, err := NewConnectionPool(hostList, testPoolConfig, DefaultLogger{})
	if err != nil {
		log.Fatal(fmt.Sprintf("Fail to initialize the connection pool, host: %s, port: %d, %s", ip, graphPort, err.Error()))
	}
	// Close all connections in the pool
	defer pool.Close()

	username := "root"
	password := "nebula"
	// Create session
	session, err := pool.GetSession(username, password)
	if err != nil {
		log.Fatal(fmt.Sprintf("Fail to create a new session from connection pool, username: %s, password: %s, %s",
			username, password, err.Error()))
	}
	// Release session and return connection back to connection pool
	defer session.Release()
	resultSet, err := session.Execute("CREATE SPACE IF NOT EXISTS test_meta(vid_type=int64, partition_num=10);" +
		"USE test_meta; CREATE TAG IF NOT EXISTS person(name string, age int32);" +
		"CREATE TAG IF NOT EXISTS player(name string, team string);" +
		"CREATE EDGE IF NOT EXISTS friend(degree int64)")
	if err != nil {
		log.Fatal("create schema error.", err.Error())

	}
	if !resultSet.IsSucceed() {
		log.Fatal(fmt.Sprintf("create schema failed, message:%s", resultSet.GetErrorMsg()))
	}
}

func teardown() {
	if client != nil && isOpen {
		client.Close()
	}
}

func TestMetaClientOpen(t *testing.T) {
	skipSslForMeta(t)

	address := HostAddress{Host: ip, Port: metaPort}
	localMetaClient := NewMetaClient(address, 2*time.Second)
	err := localMetaClient.Open()
	defer localMetaClient.Close()
	if err != nil {
		t.Errorf(fmt.Sprintf("expect open metaClient success, but failed, messgae:%s", err.Error()))
	}
	// test ipv6
	addrIpv6 := "::1"
	address = HostAddress{Host: addrIpv6, Port: metaPort}
	ipv6MetaClient := NewMetaClient(address, 2*time.Second)
	err = ipv6MetaClient.Open()
	defer ipv6MetaClient.Close()
	if err != nil {
		t.Errorf(fmt.Sprintf("expect open metaClient success with ipv6 host, but failed, message:%s", err.Error()))
	}

	// test wrong host
	address = HostAddress{Host: "1.1.1.1", Port: 9559}
	wrongMetaClient := NewMetaClient(address, 2*time.Second)
	err = wrongMetaClient.Open()
	defer wrongMetaClient.Close()
	if err != nil {
		assert.EqualError(t, err, "failed to open transport, error: dial tcp 1.1.1.1:9559: i/o timeout")
	}
}

func TestGetSpaces(t *testing.T) {
	skipSslForMeta(t)

	if !isOpen {
		openClient()
	}
	names, err := client.GetSpaces()
	if err != nil {
		t.Fatal(err)
	}
	assert.Contains(t, names, spaceName, "space names must contains %s, but not found.", spaceName)
}

func TestGetSpace(t *testing.T) {
	skipSslForMeta(t)

	if !isOpen {
		setup()
	}
	spaceItem, err := client.GetSpace(spaceName)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, string(spaceItem.GetProperties().GetSpaceName()), spaceName)
}

func TestGetTags(t *testing.T) {
	skipSslForMeta(t)

	if !isOpen {
		setup()
	}
	tags, err := client.GetTags(spaceName)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(tags), 2)
}

func TestGetTag(t *testing.T) {
	skipSslForMeta(t)

	if !isOpen {
		setup()
	}
	tagName := "person"
	schema, err := client.GetTag(spaceName, tagName)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(schema.GetColumns()), 2)
}

func TestGetEdges(t *testing.T) {
	skipSslForMeta(t)

	if !isOpen {
		setup()
	}
	edges, err := client.GetEdges(spaceName)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(edges), 1)
}

func TestGetEdge(t *testing.T) {
	skipSslForMeta(t)

	if !isOpen {
		setup()
	}
	edgeName := "friend"
	schema, err := client.GetEdge(spaceName, edgeName)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(schema.GetColumns()), 1)
}

func TestGetPartsAlloc(t *testing.T) {
	skipSslForMeta(t)

	if !isOpen {
		setup()
	}
	partsAllocMap, err := client.GetPartsAlloc(spaceName)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(partsAllocMap), 10, fmt.Sprintf("space test_meta has 10 parts, but got %d.", len(partsAllocMap)))
}

func TestGetPartLeaders(t *testing.T) {
	skipSslForMeta(t)

	if !isOpen {
		setup()
	}
	partToHostMap, err := client.GetPartsLeader(spaceName)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(partToHostMap), 10, fmt.Sprintf("space test_meta has 10 parts, but got %d.", len(partToHostMap)))
}

func TestGetHosts(t *testing.T) {
	skipSslForMeta(t)

	if !isOpen {
		setup()
	}
	hostList, err := client.ListStorageHosts()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(hostList), 3, fmt.Sprintf("storaged hosts num expect 3, but got %d", len(hostList)))
}

func TestMetaClientExecuteAfterClose(t *testing.T) {
	skipSslForMeta(t)

	address := HostAddress{Host: ip, Port: metaPort}
	localMetaClient := NewMetaClient(address, 2*time.Second)
	err := localMetaClient.Open()
	defer localMetaClient.Close()
	if err != nil {
		t.Errorf(fmt.Sprintf("expect open metaClient success, but failed, messgae:%s", err.Error()))
	}
	localMetaClient.Close()
	names, err := localMetaClient.GetSpaces()
	if err != nil {
		assert.EqualError(t, err, "Connection not open")
	} else {
		fmt.Printf(names[0])
		t.Errorf("expect error when getSpaces after client has been closed.")
	}
}

func skipSslForMeta(t *testing.T) {
	if os.Getenv("ssl_test") == "true" || os.Getenv("self_signed") == "true" {
		fmt.Printf("Skipping meta test with SSL in CI environment.")
		t.SkipNow()
	}
}

func TestMain(m *testing.M) {
	retCode := m.Run()
	teardown()
	os.Exit(retCode)
}
