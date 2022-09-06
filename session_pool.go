/*
 *
 * Copyright (c) 2022 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 *
 */

package nebula_go

import (
	"container/list"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/vesoft-inc/nebula-go/v3/nebula"
)

// SessionPool is a pool that manages sessions internally.
//
// Usage:
// Construct
// sessionPool = newSessionPool(conf)
//
// Initialize
// sessionPool.init()
//
// Execute query
// result = sessionPool.execute("query")
//
// Release:
// sessionPool.close()
type SessionPool struct {
	idleSessions   list.List
	activeSessions list.List
	conf           SessionPoolConf
	tz             timezoneInfo
	log            Logger
	closed         bool
	cleanerChan    chan struct{} //notify when pool is close
	mu             sync.Mutex
	sslConfig      *tls.Config
}

// NewSessionPool creates a new session pool with the given configs.
func NewSessionPool(conf SessionPoolConf, log Logger) (*SessionPool, error) {
	// check the config
	conf.checkBasicFields(log)

	newSessionPool := &SessionPool{
		conf: conf,
	}

	// init the pool
	if err := newSessionPool.init(); err != nil {
		return nil, fmt.Errorf("failed to create a new session pool, %s", err.Error())
	}

	return newSessionPool, nil
}

// init initializes the session pool.
func (pool *SessionPool) init() error {
	// check the hosts status
	if err := checkAddresses(pool.conf.TimeOut, pool.conf.ServiceAddrs, pool.sslConfig); err != nil {
		return fmt.Errorf("failed to initialize the session pool, %s", err.Error())
	}

	// create sessions to fulfill the min connection size

	return nil
}

// Execute returns the result of the given query as a ResultSet
func (pool *SessionPool) Execute(stmt string) (*ResultSet, error) {
	return pool.ExecuteWithParameter(stmt, map[string]interface{}{})
}

// ExecuteWithParameter returns the result of the given query as a ResultSet
func (pool *SessionPool) ExecuteWithParameter(stmt string, params map[string]interface{}) (*ResultSet, error) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	// Get a session from the pool
	session, err := pool.getIdleSession()
	if err != nil {
		return nil, err
	}
	// check the session is valid
	if session.connection == nil {
		return nil, fmt.Errorf("failed to execute: Session has been released")
	}
	// parse params
	paramsMap, err := parseParams(params)
	if err != nil {
		return nil, err
	}

	// Execute the query
	resp, err := session.connection.executeWithParameter(session.sessionID, stmt, paramsMap)
	if err != nil {
		return nil, err
	}
	resSet, err := genResultSet(resp, session.timezoneInfo)
	if err != nil {
		return nil, err
	}
	return resSet, err
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
func (pool *SessionPool) ExecuteJson(stmt string) ([]byte, error) {
	return pool.ExecuteJsonWithParameter(stmt, map[string]interface{}{})
}

// ExecuteJson returns the result of the given query as a json string
// Date and Datetime will be returned in UTC
// The result is a JSON string in the same format as ExecuteJson()
func (pool *SessionPool) ExecuteJsonWithParameter(stmt string, params map[string]interface{}) ([]byte, error) {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	// Get a session from the pool
	session, err := pool.getIdleSession()
	if err != nil {
		return nil, err
	}
	// check the session is valid
	if session.connection == nil {
		return nil, fmt.Errorf("failed to execute: Session has been released")
	}
	// parse params
	paramsMap, err := parseParams(params)
	if err != nil {
		return nil, err
	}

	resp, err := session.connection.ExecuteJsonWithParameter(session.sessionID, stmt, paramsMap)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Close logs out all sessions and closes bonded connection.
func (pool *SessionPool) Close() {
	idleLen := pool.idleSessions.Len()
	activeLen := pool.activeSessions.Len()

	// iterate all sessions
	for i := 0; i < idleLen; i++ {
		session := pool.idleSessions.Front().Value.(*Session)
		if session.connection == nil {
			session.log.Warn("Session has been released")
		} else if err := session.connection.signOut(session.sessionID); err != nil {
			session.log.Warn(fmt.Sprintf("Sign out failed, %s", err.Error()))
		}
		// close connection
		session.connection.close()
		pool.idleSessions.Remove(pool.idleSessions.Front())
	}
	for i := 0; i < activeLen; i++ {
		session := pool.activeSessions.Front().Value.(*Session)
		if session.connection == nil {
			session.log.Warn("Session has been released")
		} else if err := session.connection.signOut(session.sessionID); err != nil {
			session.log.Warn(fmt.Sprintf("Sign out failed, %s", err.Error()))
		}
		// close connection
		session.connection.close()
		pool.activeSessions.Remove(pool.activeSessions.Front())
	}

	pool.closed = true
	if pool.cleanerChan != nil {
		close(pool.cleanerChan)
	}
}

// newSession creates a new session and returns it.
func (pool *SessionPool) newSession() (*Session, error) {

	graphAddr := pool.getNextAddr()
	cn := connection{
		severAddress: graphAddr,
		timeout:      0 * time.Millisecond,
		returnedAt:   time.Now(),
		sslConfig:    nil,
		graph:        nil,
	}

	// open a new connection
	if err := cn.open(cn.severAddress, pool.conf.TimeOut, nil); err != nil {
		return nil, fmt.Errorf("failed to create a net.Conn-backed Transport,: %s", err.Error())
	}

	// authenticate with username and password to get a new session
	resp, err := cn.authenticate(pool.conf.Username, pool.conf.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new session: %s", err.Error())
	}
	sessID := resp.GetSessionID()
	timezoneOffset := resp.GetTimeZoneOffsetSeconds()
	timezoneName := resp.GetTimeZoneName()
	// Create new session
	newSession := Session{
		sessionID:    sessID,
		connection:   &cn,
		connPool:     nil,
		log:          pool.log,
		timezoneInfo: timezoneInfo{timezoneOffset, timezoneName},
	}
	err = newSession.Ping()
	if err != nil {
		return nil, err
	}

	return &newSession, nil
}

// getNextAddr returns the next address in the address list using simple round robin approach.
func (pool *SessionPool) getNextAddr() HostAddress {
	if pool.conf.hostIndex == len(pool.conf.ServiceAddrs) {
		pool.conf.hostIndex = 0
	}
	host := pool.conf.ServiceAddrs[pool.conf.hostIndex]
	pool.conf.hostIndex++
	return host
}

// getSession returns a available session.
func (pool *SessionPool) getIdleSession() (*Session, error) {
	// Get a session from the idle queue if possible
	if pool.idleSessions.Len() > 0 {
		session := pool.idleSessions.Remove(pool.idleSessions.Front()).(*Session)
		pool.activeSessions.PushBack(session)
		return session, nil
	} else if pool.activeSessions.Len() < pool.conf.MaxSize {
		// Create a new session if the total number of sessions is less than the max size
		session, err := pool.newSession()
		if err != nil {
			return nil, err
		}
		pool.activeSessions.PushBack(session)
		return session, nil
	}
	// There is no available session in the pool and the total session count has reached the limit
	return nil, fmt.Errorf("failed to get session: no session available in the" +
		" session pool and the total session count has reached the limit")
}

// parseParams converts the params map to a map of nebula.Value
func parseParams(params map[string]interface{}) (map[string]*nebula.Value, error) {
	paramsMap := make(map[string]*nebula.Value)
	for k, v := range params {
		nv, err := value2Nvalue(v)
		if err != nil {
			return nil, err
		}
		paramsMap[k] = nv
	}
	return paramsMap, nil
}
