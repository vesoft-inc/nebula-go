/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package connpool

import (
	"strings"

	ngdb "github.com/vesoft-inc/nebula-go"
)

type reqData struct {
	gql    string
	respCh chan<- RespData
}

type SimpleConnectionPool struct {
	stop             bool
	clients          []*ngdb.GraphClient
	idleClientsQueue chan *ngdb.GraphClient
	reqCh            chan reqData
}

func New(size int, addr, username, passwd string) (ConnectionPool, error) {
	pool := &SimpleConnectionPool{
		stop:             false,
		idleClientsQueue: make(chan *ngdb.GraphClient, size),
		reqCh:            make(chan reqData),
	}

	for i := 0; i < size; i++ {
		client, err := ngdb.NewClient(addr)
		if err != nil {
			return nil, err
		}

		err = client.Connect(username, passwd)
		if err != nil {
			return nil, err
		}

		pool.clients = append(pool.clients, client)
		pool.idleClientsQueue <- client
	}

	go pool.start()

	return pool, nil
}

func (p *SimpleConnectionPool) start() {
	var useSpace string
	for !p.stop {
		req := <-p.reqCh
		gql := req.gql
		// switch space
		if strings.HasPrefix(strings.ToUpper(req.gql), "USE ") {
			useSpace = req.gql
		} else {
			gql = useSpace + ";" + req.gql
		}
		// wait an idle client from queue
		client := <-p.idleClientsQueue
		go func() {
			defer func() {
				// Enqueue client after execution
				p.idleClientsQueue <- client
			}()
			resp, err := client.Execute(gql)
			req.respCh <- RespData{
				Resp: resp,
				Err:  err,
			}
		}()
	}
}

func (p *SimpleConnectionPool) Execute(gql string) <-chan RespData {
	ch := make(chan RespData)
	p.reqCh <- reqData{
		gql:    gql,
		respCh: ch,
	}
	return ch
}

func (p *SimpleConnectionPool) Close() {
	p.stop = true
	numConnectedClients := len(p.clients)
	for numConnectedClients > 0 {
		client := <-p.idleClientsQueue
		client.Disconnect()
		numConnectedClients--
	}
	close(p.reqCh)
	close(p.idleClientsQueue)
}
