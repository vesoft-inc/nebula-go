/*
 *
 * Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 *
 */

package nebula_go

import (
	"container/list"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/vesoft-inc/nebula-go/v3/nebula"
)

type ConnectionPool struct {
	idleConnectionQueue   list.List
	activeConnectionQueue list.List
	addresses             []HostAddress
	conf                  PoolConfig
	hostIndex             int
	log                   Logger
	rwLock                sync.RWMutex
	cleanerChan           chan struct{} // notify when pool is close
	closed                bool
	sslConfig             *tls.Config
}

// NewConnectionPool constructs a new connection pool using the given addresses and configs
func NewConnectionPool(ctx context.Context, addresses []HostAddress, conf PoolConfig, log Logger) (*ConnectionPool, error) {
	return NewSslConnectionPool(ctx, addresses, conf, nil, log)
}

// NewSslConnectionPool constructs a new SSL connection pool using the given addresses and configs
func NewSslConnectionPool(
	ctx context.Context, addresses []HostAddress, conf PoolConfig, sslConfig *tls.Config, log Logger) (*ConnectionPool, error) {
	// Check input
	if len(addresses) == 0 {
		return nil, fmt.Errorf("failed to initialize connection pool: illegal address input")
	}

	// Check config
	conf.validateConf(log)

	newPool := &ConnectionPool{
		conf:      conf,
		log:       log,
		addresses: addresses,
		hostIndex: 0,
		sslConfig: sslConfig,
	}

	// Init pool with SSL socket
	if err := newPool.initPool(ctx); err != nil {
		return nil, err
	}
	newPool.startCleaner()
	return newPool, nil
}

// initPool initializes the connection pool
func (pool *ConnectionPool) initPool(ctx context.Context) error {
	if err := checkAddresses(ctx, pool.conf.TimeOut, pool.addresses, pool.sslConfig,
		pool.conf.UseHTTP2, pool.conf.HttpHeader, pool.conf.HandshakeKey); err != nil {
		return fmt.Errorf("failed to open connection, error: %s ", err.Error())
	}

	for i := 0; i < pool.conf.MinConnPoolSize; i++ {
		// Simple round-robin
		newConn := newConnection(pool.addresses[i%len(pool.addresses)])

		// Open connection to host
		if err := newConn.open(ctx, newConn.severAddress, pool.conf.TimeOut, pool.sslConfig,
			pool.conf.UseHTTP2, pool.conf.HttpHeader, pool.conf.HandshakeKey); err != nil {
			// If initialization failed, clean idle queue
			idleLen := pool.idleConnectionQueue.Len()
			for i := 0; i < idleLen; i++ {
				if err := pool.idleConnectionQueue.Front().Value.(*connection).close(); err != nil {
					pool.log.Error(fmt.Sprintf("failed to close connection, error: %s", err.Error()))
				}
				pool.idleConnectionQueue.Remove(pool.idleConnectionQueue.Front())
			}
			return fmt.Errorf("failed to open connection, error: %s ", err.Error())
		}
		// Mark connection as in use
		pool.idleConnectionQueue.PushBack(newConn)
	}
	return nil
}

// GetSession authenticates the username and password.
// It returns a session if the authentication succeed.
func (pool *ConnectionPool) GetSession(ctx context.Context, username, password string) (*Session, error) {
	// Get valid and usable connection
	var conn *connection = nil
	var err error = nil
	const retryTimes = 3
	for i := 0; i < retryTimes; i++ {
		conn, err = pool.getIdleConn(ctx)
		if err == nil {
			break
		}
	}
	if conn == nil {
		return nil, err
	}
	// Authenticate
	resp, err := conn.authenticate(ctx, username, password)
	if err != nil {
		// if authentication failed, put connection back
		pool.pushBack(conn)
		return nil, err
	}

	// Check auth response
	if resp.GetErrorCode() != nebula.ErrorCode_SUCCEEDED {
		// if authentication responded with unsuccessful code, put connection back
		pool.pushBack(conn)
		return nil, fmt.Errorf("failed to authenticate, error code: %d, error msg: %s",
			resp.GetErrorCode(), resp.GetErrorMsg())
	}

	sessID := resp.GetSessionID()
	timezoneOffset := resp.GetTimeZoneOffsetSeconds()
	timezoneName := resp.GetTimeZoneName()
	// Create new session
	newSession := Session{
		sessionID:    sessID,
		connection:   conn,
		connPool:     pool,
		log:          pool.log,
		timezoneInfo: timezoneInfo{timezoneOffset, timezoneName},
	}

	return &newSession, nil
}

func (pool *ConnectionPool) getIdleConn(ctx context.Context) (*connection, error) {
	pool.rwLock.Lock()
	defer pool.rwLock.Unlock()

	// Take an idle valid connection if possible
	if pool.idleConnectionQueue.Len() > 0 {
		var newConn *connection = nil
		var newEle *list.Element = nil
		var tmpNextEle *list.Element
		for ele := pool.idleConnectionQueue.Front(); ele != nil; ele = tmpNextEle {
			// Check if connection is valid
			if res := ele.Value.(*connection).ping(ctx); res {
				newConn = ele.Value.(*connection)
				newEle = ele
				break
			} else {
				tmpNextEle = ele.Next()
				pool.idleConnectionQueue.Remove(ele)
				if err := ele.Value.(*connection).close(); err != nil {
					pool.log.Error(fmt.Sprintf("failed to close connection, error: %s", err.Error()))
				}
			}
		}
		if newConn == nil {
			return pool.createConnection(ctx)
		}
		// Remove new connection from idle and add to active if found
		pool.idleConnectionQueue.Remove(newEle)
		pool.activeConnectionQueue.PushBack(newConn)
		return newConn, nil
	}

	// Create a new connection if there is no idle connection and total connection < pool max size
	newConn, err := pool.createConnection(ctx)
	// TODO: If no idle available, wait for timeout and reconnect
	return newConn, err
}

// Release connection to pool
func (pool *ConnectionPool) release(conn *connection) {
	pool.deactivate(conn, true, true)
}

func (pool *ConnectionPool) pushBack(conn *connection) {
	pool.deactivate(conn, true, false)
}

// Deactivate connection. Add to idleConnectionQueue if pushBack is true.
// Release connection if release is true.
func (pool *ConnectionPool) deactivate(conn *connection, pushBack, release bool) {
	pool.rwLock.Lock()
	defer pool.rwLock.Unlock()
	// Remove connection from active queue and add into idle queue
	removeFromList(&pool.activeConnectionQueue, conn)
	if release {
		conn.release()
	}
	if pushBack {
		pool.idleConnectionQueue.PushBack(conn)
	}
}

// Ping checks availability of host
func (pool *ConnectionPool) Ping(ctx context.Context, host HostAddress, timeout time.Duration) error {
	return pingAddress(ctx, host, timeout, pool.sslConfig, pool.conf.UseHTTP2, pool.conf.HttpHeader, pool.conf.HandshakeKey)
}

// Close closes all connection
func (pool *ConnectionPool) Close() {
	pool.rwLock.Lock()
	defer pool.rwLock.Unlock()

	// TODO(Aiee) merge 2 lists and close all connections
	idleLen := pool.idleConnectionQueue.Len()
	activeLen := pool.activeConnectionQueue.Len()

	for i := 0; i < idleLen; i++ {
		if err := pool.idleConnectionQueue.Front().Value.(*connection).close(); err != nil {
			pool.log.Error(fmt.Sprintf("failed to close connection, error: %s", err.Error()))
		}
		pool.idleConnectionQueue.Remove(pool.idleConnectionQueue.Front())
	}
	for i := 0; i < activeLen; i++ {
		if err := pool.activeConnectionQueue.Front().Value.(*connection).close(); err != nil {
			pool.log.Error(fmt.Sprintf("failed to close connection, error: %s", err.Error()))
		}

		pool.activeConnectionQueue.Remove(pool.activeConnectionQueue.Front())
	}

	pool.closed = true
	if pool.cleanerChan != nil {
		close(pool.cleanerChan)
	}
}

func (pool *ConnectionPool) getActiveConnCount() int {
	return pool.activeConnectionQueue.Len()
}

func (pool *ConnectionPool) getIdleConnCount() int {
	return pool.idleConnectionQueue.Len()
}

// Get a valid host (round robin)
func (pool *ConnectionPool) getHost() HostAddress {
	if pool.hostIndex == len(pool.addresses) {
		pool.hostIndex = 0
	}
	host := pool.addresses[pool.hostIndex]
	pool.hostIndex++
	return host
}

// Select a new host to create a new connection
func (pool *ConnectionPool) newConnToHost(ctx context.Context) (*connection, error) {
	// Get a valid host (round robin)
	host := pool.getHost()
	newConn := newConnection(host)
	// Open connection to host
	if err := newConn.open(ctx, newConn.severAddress, pool.conf.TimeOut, pool.sslConfig,
		pool.conf.UseHTTP2, pool.conf.HttpHeader, pool.conf.HandshakeKey); err != nil {
		return nil, err
	}
	// Add connection to active queue
	pool.activeConnectionQueue.PushBack(newConn)
	// TODO: update workload
	return newConn, nil
}

// Remove a connection from list
func removeFromList(l *list.List, conn *connection) {
	for ele := l.Front(); ele != nil; ele = ele.Next() {
		if ele.Value.(*connection) == conn {
			l.Remove(ele)
		}
	}
}

// Compare total connection number with pool max size and return a connection if capable
func (pool *ConnectionPool) createConnection(ctx context.Context) (*connection, error) {
	totalConn := pool.idleConnectionQueue.Len() + pool.activeConnectionQueue.Len()
	// If no idle available and the number of total connection reaches the max pool size, return error/wait for timeout
	if totalConn >= pool.conf.MaxConnPoolSize {
		return nil, fmt.Errorf("failed to get connection: No valid connection" +
			" in the idle queue and connection number has reached the pool capacity")
	}

	newConn, err := pool.newConnToHost(ctx)
	if err != nil {
		return nil, err
	}
	// TODO: update workload
	return newConn, nil
}

// startCleaner starts connectionCleaner if idleTime > 0.
func (pool *ConnectionPool) startCleaner() {
	if pool.conf.IdleTime > 0 && pool.cleanerChan == nil {
		pool.cleanerChan = make(chan struct{}, 1)
		go pool.connectionCleaner()
	}
}

func (pool *ConnectionPool) connectionCleaner() {
	const minInterval = time.Minute

	d := pool.conf.IdleTime

	if d < minInterval {
		d = minInterval
	}
	t := time.NewTimer(d)

	for {
		select {
		case <-t.C:
		case <-pool.cleanerChan: // pool was closed.
		}

		pool.rwLock.Lock()

		if pool.closed {
			pool.cleanerChan = nil
			pool.rwLock.Unlock()
			return
		}

		closing := pool.timeoutConnectionList()
		pool.rwLock.Unlock()
		for _, c := range closing {
			if err := c.close(); err != nil {
				pool.log.Error(fmt.Sprintf("failed to close connection, error: %s", err.Error()))
			}
		}

		t.Reset(d)
	}
}

func (pool *ConnectionPool) timeoutConnectionList() (closing []*connection) {
	if pool.conf.IdleTime > 0 {
		expiredSince := time.Now().Add(-pool.conf.IdleTime)
		var newEle *list.Element

		maxCleanSize := pool.idleConnectionQueue.Len() + pool.activeConnectionQueue.Len() - pool.conf.MinConnPoolSize

		for ele := pool.idleConnectionQueue.Front(); ele != nil; {
			if maxCleanSize == 0 {
				return
			}

			newEle = ele.Next()
			// Check connection is expired
			if !ele.Value.(*connection).returnedAt.Before(expiredSince) {
				return
			}
			closing = append(closing, ele.Value.(*connection))
			pool.idleConnectionQueue.Remove(ele)
			ele = newEle
			maxCleanSize--
		}
	}
	return
}

// checkAddresses checks addresses availability
// It opens a temporary connection to each address and closes it immediately.
// If no error is returned, the addresses are available.
func checkAddresses(
	ctx context.Context,
	confTimeout time.Duration, addresses []HostAddress, sslConfig *tls.Config,
	useHTTP2 bool, httpHeader http.Header, handshakeKey string,
) error {
	timeout := 3 * time.Second
	if confTimeout != 0 && confTimeout < timeout {
		timeout = confTimeout
	}
	for _, address := range addresses {
		if err := pingAddress(ctx, address, timeout, sslConfig, useHTTP2, httpHeader, handshakeKey); err != nil {
			return err
		}
	}
	return nil
}

func pingAddress(
	ctx context.Context,
	address HostAddress, timeout time.Duration, sslConfig *tls.Config,
	useHTTP2 bool, httpHeader http.Header, handshakeKey string,
) error {
	newConn := newConnection(address)
	// Open connection to host
	if err := newConn.open(ctx, newConn.severAddress, timeout, sslConfig, useHTTP2, httpHeader, handshakeKey); err != nil {
		return err
	}
	defer newConn.close()
	return nil
}
