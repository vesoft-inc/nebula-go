package nebula_go

import (
	"crypto/tls"
	"fmt"
	"github.com/bits-and-blooms/bitset"
	"sync"
)

type ManagerConfig struct {
	username   string
	password   string
	spaceName  string
	addresses  []HostAddress
	poolConfig PoolConfig
	sslConfig  *tls.Config
}

type SessionManager struct {
	config         ManagerConfig
	mutex          sync.Mutex
	isClose        bool
	log            Logger
	pool           *ConnectionPool
	idleBitSet     *bitset.BitSet
	idleSessionMap map[uint]*SessionWrapper
	sessionMap     map[int64]*SessionWrapper
}

func NewSessionManager(config ManagerConfig, log Logger) (*SessionManager, error) {
	if len(config.spaceName) == 0 {
		return nil, fmt.Errorf("fail to create session manager: space name can not be empty")
	}
	config.poolConfig.validateConf(log)

	manager := &SessionManager{
		config: config,
		log:    log,
	}
	err := manager.initManager()
	if err != nil {
		return nil, err
	}
	return manager, nil
}

func (m *SessionManager) initManager() error {

	var pool *ConnectionPool
	var err error
	if m.config.sslConfig != nil {
		pool, err = NewConnectionPool(m.config.addresses, m.config.poolConfig, m.log)
	} else {
		pool, err = NewSslConnectionPool(m.config.addresses, m.config.poolConfig, m.config.sslConfig, m.log)
	}

	if err != nil {
		return err
	}
	m.pool = pool

	// idleBitSet: bitset contains [config.poolConfig.MaxConnPoolSize] bits
	// idleBitSet: [1, 0, 1, .... 0], set value of position-i to 1 means there is an idle wrapper-session available in idleSessionMap
	m.idleBitSet = bitset.New(uint(m.config.poolConfig.MaxConnPoolSize))

	// sessionMap : <index of idleBitSet: *SessionWrapper>
	m.idleSessionMap = make(map[uint]*SessionWrapper)

	// all session get from manager, <sessionId, *SessionWrapper>
	m.sessionMap = make(map[int64]*SessionWrapper)

	return nil
}

// GetSession get a wrapper session from sessionManager
func (m *SessionManager) GetSession() (*SessionWrapper, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.isClose {
		return nil, fmt.Errorf("fail to get new seesion, session manager has closed")
	}

	// if there are any available session
	if m.idleBitSet.Count() != 0 {
		index, ok := m.idleBitSet.NextSet(0)
		if ok {
			m.idleBitSet.Clear(index)
			session := m.idleSessionMap[index]
			delete(m.idleSessionMap, index)
			return session, nil
		}
	}

	// create new Session from pool
	session, err := m.pool.GetSession(m.config.username, m.config.password)
	if err != nil {
		return nil, err
	}

	// change space
	result, err := session.Execute("USE " + m.config.spaceName)
	if err != nil {
		return nil, err
	}
	if !result.IsSucceed() {
		return nil, fmt.Errorf("fail to get new seesion, change space error: %s", err.Error())
	}

	sessionWrapper := &SessionWrapper{session: session, sessionManager: m, sessionId: session.GetSessionID()}
	m.sessionMap[session.GetSessionID()] = sessionWrapper
	return sessionWrapper, nil
}

// ReleaseSession release wrapper session to sessionManager
func (m *SessionManager) ReleaseSession(sessionWrapper *SessionWrapper) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// do nothing
	if sessionWrapper == nil || sessionWrapper.isRelease || m.isClose {
		return
	}

	// session is not created by manager, just return to pool
	if _, ok := m.sessionMap[sessionWrapper.session.GetSessionID()]; !ok {
		sessionWrapper.session.Release()
		return
	}

	// acquire position from idleBitSet
	setIndex, ok := m.idleBitSet.NextClear(0)
	if !ok {
		sessionWrapper.session.Release()
	} else {
		m.idleBitSet.Set(setIndex)
		newWrapper := &SessionWrapper{session: sessionWrapper.session, sessionManager: m, sessionId: sessionWrapper.sessionId}
		m.idleSessionMap[setIndex] = newWrapper
	}
	sessionWrapper.isRelease = true
	sessionWrapper.session = nil
	sessionWrapper.sessionManager = nil
}

// Close mark all wrapper session to release status and close connection pool
func (m *SessionManager) Close() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, v := range m.sessionMap {
		v.isRelease = true
	}
	m.sessionMap = nil
	m.idleSessionMap = nil
	m.pool.Close()
	m.isClose = true
}
