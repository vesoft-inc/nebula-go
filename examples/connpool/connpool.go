/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package connpool

import (
	"fmt"

	nebula "github.com/vesoft-inc/nebula-go"
	"github.com/vesoft-inc/nebula-go/nebula/graph"
)

type RespData struct {
	Resp *graph.ExecutionResponse
	Err  error
}

type SpaceDesc struct {
	Name          string
	NumPartitions int
	ReplicaFactor int
	Charset       string
	Collate       string
}

type Connection struct {
	Ip       string
	Port     int
	User     string
	Password string
}

type ConnectionPool interface {
	Close()
	Execute(stmt string) <-chan RespData
}

func (rsp RespData) IsError() bool {
	return rsp.Err != nil || nebula.IsError(rsp.Resp)
}

func (rsp RespData) String() string {
	if rsp.Err != nil {
		return fmt.Sprintf("Error: %s", rsp.Err.Error())
	}
	return rsp.Resp.String()
}

func (s SpaceDesc) CreateSpaceString() string {
	return fmt.Sprintf("CREATE SPACE IF NOT EXISTS `%s`(partition_num=%d, replica_factor=%d, charset=%s, collate=%s)",
		s.Name, s.NumPartitions, s.ReplicaFactor, s.Charset, s.Collate)
}

func (s SpaceDesc) UseSpaceString() string {
	return fmt.Sprintf("Use `%s`", s.Name)
}
