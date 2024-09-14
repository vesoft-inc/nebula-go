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

	"github.com/vesoft-inc/nebula-go/v3/nebula"
	graph "github.com/vesoft-inc/nebula-go/v3/nebula/graph"
)

type timezoneInfo struct {
	offset int32
	name   []byte
}

type Session struct {
	sessionID  int64
	connection *connection
	connPool   *ConnectionPool // the connection pool which the session belongs to. could be nil if the Session is store in the SessionPool
	log        Logger
	mu         sync.Mutex
	timezoneInfo
}

func (session *Session) reconnectWithExecuteErr(err error) error {
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
	// Execute with the new connection
	return f()
}

// ExecuteWithParameter returns the result of the given query as a ResultSet
func (session *Session) ExecuteWithParameter(stmt string, params map[string]interface{}) (*ResultSet, error) {
	session.mu.Lock()
	defer session.mu.Unlock()
	paramsMap, err := parseParams(params)
	if err != nil {
		return nil, err
	}

	fn := func() (*graph.ExecutionResponse, error) {
		return session.connection.executeWithParameter(session.sessionID, stmt, paramsMap)
	}
	return session.tryExecuteLocked(fn)

}

// Execute returns the result of the given query as a ResultSet
func (session *Session) Execute(stmt string) (*ResultSet, error) {
	return session.ExecuteWithParameter(stmt, map[string]interface{}{})
}

func (session *Session) tryExecuteLocked(fn func() (*graph.ExecutionResponse, error)) (*ResultSet, error) {
	if session.connection == nil {
		return nil, fmt.Errorf("failed to execute: Session has been released")
	}
	execFunc := func() (interface{}, error) {
		resp, err := fn()
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

func (session *Session) ExecuteWithTimeout(stmt string, timeoutMs int64) (*ResultSet, error) {
	return session.ExecuteWithParameterTimeout(stmt, map[string]interface{}{}, timeoutMs)
}

func (session *Session) ExecuteWithParameterTimeout(stmt string, params map[string]interface{}, timeoutMs int64) (*ResultSet, error) {
	session.mu.Lock()
	defer session.mu.Unlock()
	if timeoutMs <= 0 {
		return nil, fmt.Errorf("timeout should be a positive number")
	}
	paramsMap, err := parseParams(params)
	if err != nil {
		return nil, err
	}

	fn := func() (*graph.ExecutionResponse, error) {
		return session.connection.executeWithParameterTimeout(session.sessionID, stmt, paramsMap, timeoutMs)
	}
	return session.tryExecuteLocked(fn)
}

// ExecuteJson returns the result of the given query as a json string
// Date and Datetime will be returned in UTC
//
//	JSON struct:
//
//	{
//	    "results":[
//	        {
//	            "columns":[
//	            ],
//	            "data":[
//	                {
//	                    "row":[
//	                        "row-data"
//	                    ],
//	                    "meta":[
//	                        "metadata"
//	                    ]
//	                }
//	            ],
//	            "latencyInUs":0,
//	            "spaceName":"",
//	            "planDesc ":{
//	                "planNodeDescs":[
//	                    {
//	                        "name":"",
//	                        "id":0,
//	                        "outputVar":"",
//	                        "description":{
//	                            "key":""
//	                        },
//	                        "profiles":[
//	                            {
//	                                "rows":1,
//	                                "execDurationInUs":0,
//	                                "totalDurationInUs":0,
//	                                "otherStats":{}
//	                            }
//	                        ],
//	                        "branchInfo":{
//	                            "isDoBranch":false,
//	                            "conditionNodeId":-1
//	                        },
//	                        "dependencies":[]
//	                    }
//	                ],
//	                "nodeIndexMap":{},
//	                "format":"",
//	                "optimize_time_in_us":0
//	            },
//	            "comment ":""
//	        }
//	    ],
//	    "errors":[
//	        {
//	      		"code": 0,
//	      		"message": ""
//	        }
//	    ]
//	}
func (session *Session) ExecuteJson(stmt string) ([]byte, error) {
	return session.ExecuteJsonWithParameter(stmt, map[string]interface{}{})
}

// ExecuteJson returns the result of the given query as a json string
// Date and Datetime will be returned in UTC
// The result is a JSON string in the same format as ExecuteJson()
func (session *Session) ExecuteJsonWithParameter(stmt string, params map[string]interface{}) ([]byte, error) {
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.connection == nil {
		return nil, fmt.Errorf("failed to execute: Session has been released")
	}

	paramsMap := make(map[string]*nebula.Value)
	for k, v := range params {
		nv, er := value2Nvalue(v)
		if er != nil {
			return nil, er
		}
		paramsMap[k] = nv
	}
	execFunc := func() (interface{}, error) {
		resp, err := session.connection.ExecuteJsonWithParameter(session.sessionID, stmt, paramsMap)
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

func (session *Session) ExecuteAndCheck(stmt string) (*ResultSet, error) {
	rs, err := session.Execute(stmt)
	if err != nil {
		return nil, err
	}

	if !rs.IsSucceed() {
		errMsg := rs.GetErrorMsg()
		return nil, fmt.Errorf("fail to execute query. %s", errMsg)
	}

	return rs, nil
}

type SpaceConf struct {
	Name           string
	Partition      uint
	Replica        uint
	VidType        string
	IgnoreIfExists bool
	Comment        string
}

func (session *Session) CreateSpace(conf SpaceConf) (*ResultSet, error) {
	if conf.Partition == 0 {
		conf.Partition = 100
	}
	if conf.Replica == 0 {
		conf.Replica = 1
	}
	if conf.VidType == "" {
		conf.VidType = "FIXED_STRING(8)"
	}

	var q string
	if conf.IgnoreIfExists {
		q = fmt.Sprintf(
			"CREATE SPACE IF NOT EXISTS %s (partition_num = %d, replica_factor = %d, vid_type = %s)",
			conf.Name,
			conf.Partition,
			conf.Replica,
			conf.VidType,
		)
	} else {
		q = fmt.Sprintf(
			"CREATE SPACE %s (partition_num = %d, replica_factor = %d, vid_type = %s)",
			conf.Name,
			conf.Partition,
			conf.Replica,
			conf.VidType,
		)
	}

	if conf.Comment != "" {
		q += fmt.Sprintf(` COMMENT = "%s"`, conf.Comment)
	}

	return session.ExecuteAndCheck(q + ";")
}

func (session *Session) ShowSpaces() ([]SpaceName, error) {
	rs, err := session.ExecuteAndCheck("SHOW SPACES;")
	if err != nil {
		return nil, err
	}

	var names []SpaceName
	rs.Scan(&names)

	return names, nil
}

func (session *Session) reConnect() error {
	newConnection, err := session.connPool.getIdleConn()
	if err != nil {
		err = fmt.Errorf(err.Error())
		return err
	}

	session.connPool.deactivate(session.connection, false, false)
	session.connection = newConnection
	return nil
}

// Release logs out and releases connection hold by session.
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

	// if the session is created from the connection pool, return the connection to the pool
	if session.connPool != nil {
		session.connPool.release(session.connection)
	}
	session.connection = nil
}

func (session *Session) GetSessionID() int64 {
	return session.sessionID
}

func IsError(resp *graph.ExecutionResponse) bool {
	return resp.GetErrorCode() != nebula.ErrorCode_SUCCEEDED
}

// Ping checks if the session is valid
func (session *Session) Ping() error {
	if session.connection == nil {
		return fmt.Errorf("failed to ping: Session has been released")
	}
	// send ping request
	resp, err := session.Execute(`RETURN "NEBULA GO PING"`)
	// check connection level error
	if err != nil {
		return fmt.Errorf("session ping failed, %s" + err.Error())
	}
	// check session level error
	if !resp.IsSucceed() {
		return fmt.Errorf("session ping failed, %s" + resp.GetErrorMsg())
	}
	return nil
}

// construct Slice to nebula.NList
func slice2Nlist(list []interface{}) (*nebula.NList, error) {
	sv := []*nebula.Value{}
	var ret nebula.NList
	for _, item := range list {
		nv, er := value2Nvalue(item)
		if er != nil {
			return nil, er
		}
		sv = append(sv, nv)
	}
	ret.Values = sv
	return &ret, nil
}

// construct map to nebula.NMap
func map2Nmap(m map[string]interface{}) (*nebula.NMap, error) {
	var ret nebula.NMap
	kvs, err := parseParams(m)
	if err != nil {
		return nil, err
	}
	ret.Kvs = kvs
	return &ret, nil
}

// construct go-type to nebula.Value
func value2Nvalue(param interface{}) (value *nebula.Value, err error) {
	value = nebula.NewValue()
	switch v := param.(type) {
	case bool:
		value.BVal = &v
	case int:
		ival := int64(v)
		value.IVal = &ival
	case int64:
		value.IVal = &v
	case float64:
		if v == float64(int64(v)) {
			iv := int64(v)
			value.IVal = &iv
		} else {
			value.FVal = &v
		}
	case float32:
		if v == float32(int64(v)) {
			iv := int64(v)
			value.IVal = &iv
		} else {
			fval := float64(v)
			value.FVal = &fval
		}
	case string:
		value.SVal = []byte(v)
	case nil:
		nval := nebula.NullType___NULL__
		value.NVal = &nval
	case []interface{}:
		nv, er := slice2Nlist(v)
		if er != nil {
			err = er
		}
		value.LVal = nv
	case map[string]interface{}:
		nv, er := map2Nmap(v)
		if er != nil {
			err = er
		}
		value.MVal = nv
	case nebula.Value:
		value = &v
	case nebula.Date:
		value.SetDVal(&v)
	case nebula.DateTime:
		value.SetDtVal(&v)
	case nebula.Duration:
		value.SetDuVal(&v)
	case nebula.Time:
		value.SetTVal(&v)
	case nebula.Geography:
		value.SetGgVal(&v)
	default:
		// unsupported other Value type, use this function carefully
		err = fmt.Errorf("only support convert boolean/float/int/int64/string/map/list to nebula.Value but %T", param)
	}
	return
}
