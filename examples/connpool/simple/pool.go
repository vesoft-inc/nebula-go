package simple

import "github.com/vesoft-inc/nebula-go/nebula/graph"

type RespData struct {
	Resp *graph.ExecutionResponse
	Err  error
}

type reqData struct {
	gql    string
	respCh chan<- RespData
}

type ConnectionPool struct {
	pool   []*connection
	reqCh  chan reqData
	stopCh chan bool
}

func New(size int, addr, username, passwd string) (pool ConnectionPool, err error) {
	conns := make([]*connection, size)

	for i := range pool.pool {
		conns[i], err = newConnection(addr, username, passwd)
		if err != nil {
			return
		}
	}

	pool = ConnectionPool{
		pool:   conns,
		reqCh:  make(chan reqData),
		stopCh: make(chan bool),
	}

	go func() {
		for {
			select {
			case req := <-pool.reqCh:
				wait := true
				for wait {
					for _, conn := range pool.pool {
						if conn.idle {
							conn.idle = false
							go func() {
								defer func() { conn.idle = true }()
								resp, err := conn.client.Execute(req.gql)
								req.respCh <- RespData{
									Resp: resp,
									Err:  err,
								}
							}()
							wait = false
						}
					}
				}
			case <-pool.stopCh:
				for _, conn := range pool.pool {
					conn.client.Disconnect()
				}
				break
			}
		}
	}()

}

func (p *ConnectionPool) Execute(gql string) <-chan RespData {
	ch := make(chan RespData)
	p.reqCh <- reqData{
		gql:    gql,
		respCh: ch,
	}
	return ch
}

func (p *ConnectionPool) Close() {
	p.stopCh <- true
}
