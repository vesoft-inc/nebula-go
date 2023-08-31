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
	"fmt"
	"strconv"
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
//
// Notice that all queries will be executed in the default space specified in the pool config.
type SessionPool struct {
	idleSessions   list.List
	activeSessions list.List
	conf           SessionPoolConf
	tz             timezoneInfo
	log            Logger
	closed         bool
	cleanerChan    chan struct{} //notify when pool is close
	rwLock         sync.RWMutex
}

// one pureSession binds to one connection and shares the same lifespan.
// If the underlying connection is broken, the session will be removed from the session pool.
type pureSession struct {
	sessionID  int64
	connection *connection
	sessPool   *SessionPool
	returnedAt time.Time // the timestamp that the session was created or returned.
	timezoneInfo
	spaceName string
}

// NewSessionPool creates a new session pool with the given configs.
// There must be an existing SPACE in the DB.
func NewSessionPool(conf SessionPoolConf, log Logger) (*SessionPool, error) {
	// check the config
	conf.checkBasicFields(log)

	newSessionPool := &SessionPool{
		conf: conf,
		log:  log,
	}

	// init the pool
	if err := newSessionPool.init(); err != nil {
		return nil, fmt.Errorf("failed to create a new session pool, %s", err.Error())
	}
	newSessionPool.startCleaner()
	return newSessionPool, nil
}

// init initializes the session pool.
func (pool *SessionPool) init() error {
	// check the hosts status
	if err := checkAddresses(pool.conf.timeOut, pool.conf.serviceAddrs, pool.conf.sslConfig, pool.conf.useHTTP2); err != nil {
		return fmt.Errorf("failed to initialize the session pool, %s", err.Error())
	}

	// create sessions to fulfill the min pool size
	for i := 0; i < pool.conf.minSize; i++ {
		session, err := pool.newSession()
		if err != nil {
			return fmt.Errorf("failed to initialize the session pool, %s", err.Error())
		}

		session.returnedAt = time.Now()
		pool.addSessionToIdle(session)
	}

	return nil
}

// Execute returns the result of the given query as a ResultSet
// Notice there are some limitations:
// 1. The query should not be a plain space switch statement, e.g. "USE test_space",
// but queries like "use space xxx; match (v) return v" are accepted.
// 2. If the query contains statements like "USE <space name>", the space will be set to the
// one in the pool config after the execution of the query.
// 3. The query should not change the user password nor drop a user.
func (pool *SessionPool) Execute(stmt string) (*ResultSet, error) {
	return pool.ExecuteWithParameter(stmt, map[string]interface{}{})
}

// ExecuteWithParameter returns the result of the given query as a ResultSet
func (pool *SessionPool) ExecuteWithParameter(stmt string, params map[string]interface{}) (*ResultSet, error) {
	// Check if the pool is closed
	if pool.closed {
		return nil, fmt.Errorf("failed to execute: Session pool has been closed")
	}

	// Get a session from the pool
	session, err := pool.getSessionFromIdle()
	if err != nil {
		return nil, err
	}
	// if there's no idle session, create a new one
	if session == nil {
		session, err = pool.newSession()
		if err != nil {
			return nil, err
		}
		pool.addSessionToActive(session)
	} else {
		pool.removeSessionFromIdle(session)
		pool.addSessionToActive(session)
	}

	// Execute the query
	execFunc := func(s *pureSession) (*ResultSet, error) {
		rs, err := s.executeWithParameter(stmt, params)
		if err != nil {
			return nil, err
		}
		return rs, nil
	}

	rs, err := pool.executeWithRetry(session, execFunc, pool.conf.retryGetSessionTimes)
	if err != nil {
		session.close()
		pool.removeSessionFromActive(session)
		return nil, err
	}

	// if the space was changed after the execution of the given query,
	// change it back to the default space specified in the pool config
	if rs.GetSpaceName() != "" && rs.GetSpaceName() != pool.conf.spaceName {
		err := session.setSessionSpaceToDefault()
		if err != nil {
			pool.log.Warn(err.Error())
			session.close()
			pool.removeSessionFromActive(session)
			return nil, err
		}
	}

	// Return the session to the idle list
	pool.returnSession(session)

	return rs, nil
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
func (pool *SessionPool) ExecuteJson(stmt string) ([]byte, error) {
	return pool.ExecuteJsonWithParameter(stmt, map[string]interface{}{})
}

// ExecuteJson returns the result of the given query as a json string
// Date and Datetime will be returned in UTC
// The result is a JSON string in the same format as ExecuteJson()
// TODO(Aiee) check the space name
func (pool *SessionPool) ExecuteJsonWithParameter(stmt string, params map[string]interface{}) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

// Close logs out all sessions and closes bonded connection.
func (pool *SessionPool) Close() {
	pool.rwLock.Lock()
	defer pool.rwLock.Unlock()

	//TODO(Aiee) append 2 lists
	idleLen := pool.idleSessions.Len()
	activeLen := pool.activeSessions.Len()

	// iterate all sessions
	for i := 0; i < idleLen; i++ {
		session := pool.idleSessions.Front().Value.(*pureSession)
		session.close()
		pool.idleSessions.Remove(pool.idleSessions.Front())
	}
	for i := 0; i < activeLen; i++ {
		session := pool.activeSessions.Front().Value.(*pureSession)
		session.close()
		pool.activeSessions.Remove(pool.activeSessions.Front())
	}

	pool.closed = true
	if pool.cleanerChan != nil {
		close(pool.cleanerChan)
	}
}

// GetTotalSessionCount returns the total number of sessions in the pool
func (pool *SessionPool) GetTotalSessionCount() int {
	pool.rwLock.RLock()
	defer pool.rwLock.RUnlock()
	return pool.activeSessions.Len() + pool.idleSessions.Len()
}

// newSession creates a new session and returns it.
// `use <space>` will be executed so that the new session will be in the default space.
func (pool *SessionPool) newSession() (*pureSession, error) {
	graphAddr := pool.getNextAddr()
	cn := connection{
		severAddress: graphAddr,
		timeout:      0 * time.Millisecond,
		returnedAt:   time.Now(),
		sslConfig:    pool.conf.sslConfig,
		useHTTP2:     pool.conf.useHTTP2,
		graph:        nil,
	}

	// open a new connection
	if err := cn.open(cn.severAddress, pool.conf.timeOut, pool.conf.sslConfig, pool.conf.useHTTP2); err != nil {
		return nil, fmt.Errorf("failed to create a net.Conn-backed Transport,: %s", err.Error())
	}

	// authenticate with username and password to get a new session
	authResp, err := cn.authenticate(pool.conf.username, pool.conf.password)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new session: %s", err.Error())
	}

	// If the authentication failed, close the session pool because the pool must have a valid user to work
	if authResp.GetErrorCode() != 0 {
		if authResp.GetErrorCode() == nebula.ErrorCode_E_BAD_USERNAME_PASSWORD ||
			authResp.GetErrorCode() == nebula.ErrorCode_E_USER_NOT_FOUND {
			pool.Close()
			return nil, fmt.Errorf(
				"failed to authenticate the user, error code: %d, error message: %s, the pool has been closed",
				authResp.ErrorCode, authResp.ErrorMsg)
		}
		return nil, fmt.Errorf("failed to create a new session: %s", authResp.GetErrorMsg())
	}

	sessID := authResp.GetSessionID()
	timezoneOffset := authResp.GetTimeZoneOffsetSeconds()
	timezoneName := authResp.GetTimeZoneName()
	// Create new session
	newSession := pureSession{
		sessionID:    sessID,
		connection:   &cn,
		sessPool:     pool,
		timezoneInfo: timezoneInfo{timezoneOffset, timezoneName},
		spaceName:    pool.conf.spaceName,
	}

	// Switch to the default space
	stmt := fmt.Sprintf("USE %s", pool.conf.spaceName)
	useSpaceRs, err := newSession.execute(stmt)
	if err != nil {
		return nil, err
	}

	if useSpaceRs.GetErrorCode() != ErrorCode_SUCCEEDED {
		newSession.close()
		return nil, fmt.Errorf("failed to use space %s: %s",
			pool.conf.spaceName, useSpaceRs.GetErrorMsg())
	}
	return &newSession, nil
}

// getNextAddr returns the next address in the address list using simple round robin approach.
func (pool *SessionPool) getNextAddr() HostAddress {
	pool.rwLock.Lock()
	defer pool.rwLock.Unlock()
	if pool.conf.hostIndex >= len(pool.conf.serviceAddrs) {
		pool.conf.hostIndex = 0
	}
	host := pool.conf.serviceAddrs[pool.conf.hostIndex]
	pool.conf.hostIndex++
	return host
}

// getSession returns an available session.
// This method should move an available session to the active list and should be MT-safe.
func (pool *SessionPool) getSessionFromIdle() (*pureSession, error) {
	pool.rwLock.Lock()
	defer pool.rwLock.Unlock()
	// Get a session from the idle queue if possible
	if pool.idleSessions.Len() > 0 {
		session := pool.idleSessions.Front().Value.(*pureSession)
		pool.idleSessions.Remove(pool.idleSessions.Front())
		return session, nil
	} else if pool.activeSessions.Len() < pool.conf.maxSize {
		return nil, nil
	}
	// There is no available session in the pool and the total session count has reached the limit
	return nil, fmt.Errorf("failed to get session: no session available in the" +
		" session pool and the total session count has reached the limit")
}

// retryGetSession tries to create a new session when:
// 1. the current session is invalid.
// 2. connection is invalid.
// and then change the original session to the new one.
func (pool *SessionPool) executeWithRetry(
	session *pureSession,
	f func(*pureSession) (*ResultSet, error),
	retry int) (*ResultSet, error) {
	rs, err := f(session)
	if err == nil {
		if rs.GetErrorCode() == ErrorCode_SUCCEEDED {
			return rs, nil
		} else if rs.GetErrorCode() != ErrorCode_E_SESSION_INVALID { // only retry when the session is invalid
			return rs, err
		}
	}

	// If the session is invalid, close it first
	session.close()
	// get a new session
	for i := 0; i < retry; i++ {
		pool.log.Info("retry to get sessions")
		newSession, err := pool.newSession()
		if err != nil {
			return nil, err
		}

		pingErr := newSession.ping()
		if pingErr != nil {
			pool.log.Error("failed to ping the session, error: " + pingErr.Error())
			continue
		}
		pool.log.Info("retry to get sessions successfully")
		*session = *newSession

		return f(session)
	}
	pool.log.Error(fmt.Sprintf("failed to get session after " + strconv.Itoa(retry) + " retries"))
	return nil, fmt.Errorf("failed to get session after %d retries", retry)
}

// startCleaner starts sessionCleaner if idleTime > 0.
func (pool *SessionPool) startCleaner() {
	if pool.conf.idleTime > 0 && pool.cleanerChan == nil {
		pool.cleanerChan = make(chan struct{}, 1)
		go pool.sessionCleaner()
	}
}

func (pool *SessionPool) sessionCleaner() {
	const minInterval = time.Minute

	d := pool.conf.idleTime

	if d < minInterval {
		d = minInterval
	}
	t := time.NewTimer(d)

	for {
		select {
		case <-t.C:
		case <-pool.cleanerChan: // pool was closed.
		}

		if pool.closed {
			pool.cleanerChan = nil
			return
		}

		closing := pool.timeoutSessionList()
		//release expired session from the pool
		for _, session := range closing {
			session.close()
		}
		t.Reset(d)
	}
}

// timeoutSessionList returns a list of sessions that have been idle for longer than the idle time.
func (pool *SessionPool) timeoutSessionList() (closing []*pureSession) {
	if pool.conf.idleTime == 0 {
		return
	}
	pool.rwLock.Lock()
	defer pool.rwLock.Unlock()
	expiredSince := time.Now().Add(-pool.conf.idleTime)
	var newEle *list.Element = nil

	maxCleanSize := pool.idleSessions.Len() + pool.activeSessions.Len() - pool.conf.minSize

	for ele := pool.idleSessions.Front(); ele != nil; {
		if maxCleanSize == 0 {
			return
		}

		newEle = ele.Next()
		// Check Session is expired
		if !ele.Value.(*pureSession).returnedAt.Before(expiredSince) {
			return
		}
		closing = append(closing, ele.Value.(*pureSession))
		pool.idleSessions.Remove(ele)
		ele = newEle
		maxCleanSize--
	}
	return
}

// parseParams converts the params map to a map of nebula.Value
func parseParams(params map[string]interface{}) (map[string]*nebula.Value, error) {
	paramsMap := make(map[string]*nebula.Value)
	for k, v := range params {
		nv, err := value2Nvalue(v)
		if err != nil {
			return nil, fmt.Errorf("failed to parse params: %s", err.Error())
		}
		paramsMap[k] = nv
	}
	return paramsMap, nil
}

// removeSessionFromIdleList Removes a session from list
func (pool *SessionPool) removeSessionFromActive(session *pureSession) {
	pool.rwLock.Lock()
	defer pool.rwLock.Unlock()
	l := &pool.activeSessions
	for ele := l.Front(); ele != nil; ele = ele.Next() {
		if ele.Value.(*pureSession) == session {
			l.Remove(ele)
		}
	}
}

func (pool *SessionPool) addSessionToActive(session *pureSession) {
	pool.rwLock.Lock()
	defer pool.rwLock.Unlock()
	l := &pool.activeSessions
	l.PushBack(session)
}

func (pool *SessionPool) removeSessionFromIdle(session *pureSession) {
	pool.rwLock.Lock()
	defer pool.rwLock.Unlock()
	l := &pool.idleSessions
	for ele := l.Front(); ele != nil; ele = ele.Next() {
		if ele.Value.(*pureSession) == session {
			l.Remove(ele)
		}
	}
}

func (pool *SessionPool) addSessionToIdle(session *pureSession) {
	pool.rwLock.Lock()
	defer pool.rwLock.Unlock()
	l := &pool.idleSessions
	l.PushBack(session)
}

// returnSession returns a session from active list to the idle list.
func (pool *SessionPool) returnSession(session *pureSession) {
	pool.rwLock.Lock()
	defer pool.rwLock.Unlock()
	l := &pool.activeSessions
	for ele := l.Front(); ele != nil; ele = ele.Next() {
		if ele.Value.(*pureSession) == session {
			l.Remove(ele)
		}
	}
	l = &pool.idleSessions
	l.PushBack(session)
	session.returnedAt = time.Now()
}

func (pool *SessionPool) setSessionSpaceToDefault(session *pureSession) error {
	stmt := fmt.Sprintf("USE %s", pool.conf.spaceName)
	rs, err := session.execute(stmt)
	if err != nil {
		return err
	}

	if rs.GetErrorCode() == ErrorCode_SUCCEEDED {
		return nil
	}
	// if failed to change back to the default space, send a warning log
	// and remove the session from the pool because it is malformed.
	pool.log.Warn(fmt.Sprintf("failed to reset the space of the session: errorCode: %d, errorMsg: %s, session removed",
		rs.GetErrorCode(), rs.GetErrorMsg()))
	session.close()
	pool.removeSessionFromActive(session)
	return fmt.Errorf("failed to reset the space of the session: errorCode: %d, errorMsg: %s",
		rs.GetErrorCode(), rs.GetErrorMsg())
}

func (session *pureSession) execute(stmt string) (*ResultSet, error) {
	return session.executeWithParameter(stmt, nil)
}

func (session *pureSession) executeWithParameter(stmt string, params map[string]interface{}) (*ResultSet, error) {
	paramsMap, err := parseParams(params)
	if err != nil {
		return nil, err
	}
	if session.connection == nil {
		return nil, fmt.Errorf("failed to execute: Session has been released")
	}
	resp, err := session.connection.executeWithParameter(session.sessionID, stmt, paramsMap)
	if err != nil {
		return nil, err
	}
	rs, err := genResultSet(resp, session.timezoneInfo)
	if err != nil {
		return nil, err
	}
	return rs, nil
}

func (session *pureSession) close() {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()
	if session.connection != nil {
		// ignore signout error
		_ = session.connection.signOut(session.sessionID)
		session.connection.close()
		session.connection = nil
	}
}

// Ping checks if the session is valid
func (session *pureSession) ping() error {
	if session.connection == nil {
		return fmt.Errorf("failed to ping: Session has been released")
	}
	// send ping request
	rs, err := session.execute(`RETURN "NEBULA GO PING"`)
	// check connection level error
	if err != nil {
		return fmt.Errorf("session ping failed, %s" + err.Error())
	}
	// check session level error
	if !rs.IsSucceed() {
		return fmt.Errorf("session ping failed, %s" + rs.GetErrorMsg())
	}
	return nil
}

func (session *pureSession) setSessionSpaceToDefault() error {
	stmt := fmt.Sprintf("USE %s", session.spaceName)
	rs, err := session.execute(stmt)
	if err != nil {
		return err
	}

	if rs.GetErrorCode() == ErrorCode_SUCCEEDED {
		return nil
	}
	return fmt.Errorf("failed to reset the space of the session: errorCode: %d, errorMsg: %s",
		rs.GetErrorCode(), rs.GetErrorMsg())
}
