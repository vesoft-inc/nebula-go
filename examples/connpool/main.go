package main

import (
	"log"

	"github.com/vesoft-inc/nebula-go/examples/connpool/simple"
	"github.com/vesoft-inc/nebula-go/nebula/graph"
)

func main() {
	pool, err := simple.New(1, "127.0.0.1:3699", "user", "pass")
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	for i := 0; i < 10; i++ {
		respCh := pool.Execute("SHOW HOSTS;")
		respData := <-respCh
		if respData.Err != nil {
			log.Print(respData.Err)
		}
		if respData.Resp.GetErrorCode() != graph.ErrorCode_SUCCEEDED {
			log.Print(respData.Resp.GetErrorMsg())
		}
	}
}
