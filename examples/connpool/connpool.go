/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package connpool

import (
	"fmt"

	ngdb "github.com/vesoft-inc/nebula-go"
	"github.com/vesoft-inc/nebula-go/nebula/graph"
)

type RespData struct {
	Resp *graph.ExecutionResponse
	Err  error
}

func (rsp RespData) IsError() bool {
	return rsp.Err != nil || ngdb.IsError(rsp.Resp)
}

func (rsp RespData) String() string {
	if rsp.Err != nil {
		return fmt.Sprintf("Error: %s", rsp.Err.Error())
	}
	return rsp.Resp.String()
}

type ConnectionPool interface {
	Close()
	Execute(stmt string) <-chan RespData
}
