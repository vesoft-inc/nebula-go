package simple

import (
	"sync"

	ngdb "github.com/vesoft-inc/nebula-go"
)

type connection struct {
	client *ngdb.GraphClient
	idle   bool
	lock   sync.Mutex
}

func newConnection(addr, username, passwd string) (conn connection, err error) {
	client, err = ngdb.NewClient(addr)
	if err != nil {
		return
	}

	err = client.Connect(username, passwd)
	if err != nil {
		return
	}

	return connection{
		client: client,
		idle:   true,
	}, nil
}
