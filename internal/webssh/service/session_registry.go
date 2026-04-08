package service

import "sync"

var sessionRegistry = &SessionRegistry{
	sessions: make(map[string]*SSHSession),
}

type SessionRegistry struct {
	mu       sync.RWMutex
	sessions map[string]*SSHSession
}

func RegisterSession(sessionID string, session *SSHSession) {
	if sessionID == "" || session == nil {
		return
	}

	sessionRegistry.mu.Lock()
	defer sessionRegistry.mu.Unlock()
	sessionRegistry.sessions[sessionID] = session
}

func UnregisterSession(sessionID string) {
	if sessionID == "" {
		return
	}

	sessionRegistry.mu.Lock()
	defer sessionRegistry.mu.Unlock()
	delete(sessionRegistry.sessions, sessionID)
}

func GetSession(sessionID string) (*SSHSession, bool) {
	sessionRegistry.mu.RLock()
	defer sessionRegistry.mu.RUnlock()
	session, ok := sessionRegistry.sessions[sessionID]
	return session, ok
}
