package main

import (
	"context"
	"fmt"
	nebula "github.com/vesoft-inc/nebula-go/v3"
	"strings"
)

const (
	address = "127.0.0.1"
	// The default port of NebulaGraph 2.x is 9669.
	// 3699 is only for testing.
	port     = 3699
	username = "root"
	password = "nebula"
	useHTTP2 = false
)

// Initialize logger
var log = nebula.DefaultLogger{}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	hostAddress := nebula.HostAddress{Host: address, Port: port}
	hostList := []nebula.HostAddress{hostAddress}
	// Create configs for connection pool using default values
	testPoolConfig := nebula.GetDefaultConf()
	testPoolConfig.UseHTTP2 = useHTTP2

	// Initialize connection pool
	pool, err := nebula.NewConnectionPool(ctx, hostList, testPoolConfig, log)
	if err != nil {
		log.Fatal(fmt.Sprintf("Fail to initialize the connection pool, host: %s, port: %d, %s", address, port, err.Error()))
	}
	// Close all connections in the pool
	defer pool.Close()

	// Create session
	session, _ := pool.GetSession(ctx, username, password)

	// Release session and return connection back to connection pool
	defer session.Release(ctx)

	// Intentionally call cancel()
	cancel()

	yieldQuery := "YIELD 5;"

	// Execute a query
	_, err = session.Execute(ctx, yieldQuery)
	if err != nil && strings.Contains(err.Error(), context.Canceled.Error()) {
		fmt.Println(err.Error())
		return
	}

	fmt.Print("\n")
	log.Info("Nebula Go Client Goroutines Example Finished")
}
