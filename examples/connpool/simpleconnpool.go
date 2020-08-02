/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package connpool

import (
	"fmt"

	nebula "github.com/vesoft-inc/nebula-go"
)

type reqData struct {
	gql    string
	respCh chan<- RespData
}

type SimpleConnectionPool struct {
	stop             bool
	clients          []*nebula.GraphClient
	idleClientsQueue chan *nebula.GraphClient
	reqCh            chan reqData
}

func New(size int, conn Connection, spaceDesc SpaceDesc) (ConnectionPool, error) {
	if size <= 0 {
		return nil, fmt.Errorf("Invalid pool size: %d", size)
	}

	pool := &SimpleConnectionPool{
		stop:             false,
		idleClientsQueue: make(chan *nebula.GraphClient, size),
		reqCh:            make(chan reqData),
	}

	createSpace := true
	for i := 0; i < size; i++ {
		client, err := nebula.NewClient(fmt.Sprintf("%s:%d", conn.Ip, conn.Port))
		if err != nil {
			return nil, err
		}

		err = client.Connect(conn.User, conn.Password)
		if err != nil {
			return nil, err
		}

		pool.clients = append(pool.clients, client)

		if createSpace {
			resp, err := client.Execute(spaceDesc.CreateSpaceString())
			if err != nil || nebula.IsError(resp) {
				return nil, fmt.Errorf("Fail to create space %s", spaceDesc.Name)
			}
			createSpace = false
		}

		resp, err := client.Execute(spaceDesc.UseSpaceString())
		if err != nil || nebula.IsError(resp) {
			return nil, fmt.Errorf("Fail to use space %s", spaceDesc.Name)
		}

		pool.idleClientsQueue <- client
	}

	go pool.start()

	return pool, nil
}

func (p *SimpleConnectionPool) start() {
	for !p.stop {
		req := <-p.reqCh
		// wait an idle client from queue
		client := <-p.idleClientsQueue
		go func() {
			defer func() {
				// Enqueue client after execution
				p.idleClientsQueue <- client
			}()
			resp, err := client.Execute(req.gql)
			req.respCh <- RespData{
				Resp: resp,
				Err:  err,
			}
			close(req.respCh)
		}()
	}
}

func (p *SimpleConnectionPool) Execute(gql string) <-chan RespData {
	ch := make(chan RespData, 1)
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
