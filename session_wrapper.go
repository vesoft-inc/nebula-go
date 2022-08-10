package nebula_go

import "fmt"

type SessionWrapper struct {
	session        *Session
	sessionManager *SessionManager
	sessionId      int64
	isRelease      bool
}

func (w *SessionWrapper) GetSessionID() (int64, error) {
	if w.isRelease {
		return -1, fmt.Errorf("can not get session id of a released session wrapper")
	}
	return w.session.sessionID, nil
}

func (w *SessionWrapper) Release() {
	if !w.isRelease {
		w.sessionManager.ReleaseSession(w)
	}
}

func (w *SessionWrapper) ExecuteWithParameter(stmt string, params map[string]interface{}) (*ResultSet, error) {
	if w.isRelease {
		return nil, fmt.Errorf("can not execute statement by a released session wrapper")
	}
	return w.session.ExecuteWithParameter(stmt, params)
}

func (w *SessionWrapper) Execute(stmt string) (*ResultSet, error) {
	if w.isRelease {
		return nil, fmt.Errorf("can not execute statement by a released session wrapper")
	}
	return w.ExecuteWithParameter(stmt, map[string]interface{}{})
}

func (w *SessionWrapper) ExecuteJson(stmt string) ([]byte, error) {
	if w.isRelease {
		return nil, fmt.Errorf("can not execute statement by a released session wrapper")
	}
	return w.session.ExecuteJsonWithParameter(stmt, map[string]interface{}{})
}

func (w *SessionWrapper) ExecuteJsonWithParameter(stmt string, params map[string]interface{}) ([]byte, error) {
	if w.isRelease {
		return nil, fmt.Errorf("can not execute statement by a released session wrapper")
	}
	return w.ExecuteJsonWithParameter(stmt, params)
}
