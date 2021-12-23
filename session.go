/*
 *
 * Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 *
 */

package nebula_go

import (
	"fmt"
	"sync"

	"github.com/facebook/fbthrift/thrift/lib/go/thrift"
	"github.com/vesoft-inc/nebula-go/v2/nebula"
	graph "github.com/vesoft-inc/nebula-go/v2/nebula/graph"
)

type timezoneInfo struct {
	offset int32
	name   []byte
}

type Session struct {
	sessionID  int64
	connection *connection
	connPool   *ConnectionPool
	log        Logger
	mu         sync.Mutex
	timezoneInfo
}

func (session *Session) reconnectWithExecuteErr(err error) error {
	// Reconnect only if the tranport is closed
	err2, ok := err.(thrift.TransportException)
	if !ok {
		return err
	}
	if err2.TypeID() != thrift.END_OF_FILE {
		return err
	}
	if _err := session.reConnect(); _err != nil {
		return fmt.Errorf("failed to reconnect, %s", _err.Error())
	}
	session.log.Info(fmt.Sprintf("Successfully reconnect to host: %s, port: %d",
		session.connection.severAddress.Host, session.connection.severAddress.Port))
	return nil
}

func (session *Session) executeWithReconnect(f func() (interface{}, error)) (interface{}, error) {
	resp, err := f()
	if err == nil {
		return resp, nil
	}
	if err2 := session.reconnectWithExecuteErr(err); err2 != nil {
		return nil, err2
	}
	// Execute with the new connetion
	return f()

}

// Execute returns the result of the given query as a ResultSet
func (session *Session) Execute(stmt string) (*ResultSet, error) {
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.connection == nil {
		return nil, fmt.Errorf("failed to execute: Session has been released")
	}

	execFunc := func() (interface{}, error) {
		resp, err := session.connection.execute(session.sessionID, stmt)
		if err != nil {
			return nil, err
		}
		resSet, err := genResultSet(resp, session.timezoneInfo)
		if err != nil {
			return nil, err
		}
		return resSet, nil
	}
	resp, err := session.executeWithReconnect(execFunc)
	if err != nil {
		return nil, err
	}
	return resp.(*ResultSet), err
}

// Execute returns the result of given query as a ResultSet
func (session *Session) ExecuteWithParameter(stmt string, params map[string]*nebula.Value) (*ResultSet, error) {
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.connection == nil {
		return nil, fmt.Errorf("failed to execute: Session has been released")
	}
	execFunc := func() (interface{}, error) {
		resp, err := session.connection.executeWithParameter(session.sessionID, stmt, params)
		if err != nil {
			return nil, err
		}
		resSet, err := genResultSet(resp, session.timezoneInfo)
		if err != nil {
			return nil, err
		}
		return resSet, nil
	}

	resp, err := session.executeWithReconnect(execFunc)
	if err != nil {
		return nil, err
	}
	return resp.(*ResultSet), err

}

// ExecuteJson returns the result of the given query as a json string
// Date and Datetime will be returned in UTC
//	JSON struct:
// {
//     "results":[
//         {
//             "columns":[
//             ],
//             "data":[
//                 {
//                     "row":[
//                         "row-data"
//                     ],
//                     "meta":[
//                         "metadata"
//                     ]
//                 }
//             ],
//             "latencyInUs":0,
//             "spaceName":"",
//             "planDesc ":{
//                 "planNodeDescs":[
//                     {
//                         "name":"",
//                         "id":0,
//                         "outputVar":"",
//                         "description":{
//                             "key":""
//                         },
//                         "profiles":[
//                             {
//                                 "rows":1,
//                                 "execDurationInUs":0,
//                                 "totalDurationInUs":0,
//                                 "otherStats":{}
//                             }
//                         ],
//                         "branchInfo":{
//                             "isDoBranch":false,
//                             "conditionNodeId":-1
//                         },
//                         "dependencies":[]
//                     }
//                 ],
//                 "nodeIndexMap":{},
//                 "format":"",
//                 "optimize_time_in_us":0
//             },
//             "comment ":""
//         }
//     ],
//     "errors":[
//         {
//       		"code": 0,
//       		"message": ""
//         }
//     ]
// }
func (session *Session) ExecuteJson(stmt string) ([]byte, error) {
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.connection == nil {
		return nil, fmt.Errorf("failed to execute: Session has been released")
	}

	execFunc := func() (interface{}, error) {
		resp, err := session.connection.executeJson(session.sessionID, stmt)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
	resp, err := session.executeWithReconnect(execFunc)
	if err != nil {
		return nil, err
	}
	return resp.([]byte), err
}

func (session *Session) reConnect() error {
	newconnection, err := session.connPool.getIdleConn()
	if err != nil {
		err = fmt.Errorf(err.Error())
		return err
	}

	// Release connection to pool
	session.connPool.release(session.connection)
	session.connection = newconnection
	return nil
}

// Release logs out and releases connetion hold by session.
// The connection will be added into the activeConnectionQueue of the connection pool
// so that it could be reused.
func (session *Session) Release() {
	if session == nil {
		return
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.connection == nil {
		session.log.Warn("Session has been released")
		return
	}
	if err := session.connection.signOut(session.sessionID); err != nil {
		session.log.Warn(fmt.Sprintf("Sign out failed, %s", err.Error()))
	}
	// Release connection to pool
	session.connPool.release(session.connection)
	session.connection = nil
}

func IsError(resp *graph.ExecutionResponse) bool {
	return resp.GetErrorCode() != nebula.ErrorCode_SUCCEEDED
}
