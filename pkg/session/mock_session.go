package session

import (
	"errors"
	"sync"
	"time"
)

type MockSessionManager struct {
	sessions  map[string]SessionData
	sessionID string
	mu        sync.Mutex
}

func NewMockSessionManager() *MockSessionManager {
	return &MockSessionManager{
		sessions:  make(map[string]SessionData),
		sessionID: "mock-session-id",
	}
}

func (m *MockSessionManager) SetSessionID(sessionID string) {
	m.sessionID = sessionID
}

func (m *MockSessionManager) GetSessionID() string {
	return m.sessionID
}

func (m *MockSessionManager) GetSessions() ([]string, error) {
	var sessionIDs []string
	for id := range m.sessions {
		sessionIDs = append(sessionIDs, id)
	}
	return sessionIDs, nil
}

func (m *MockSessionManager) LocateSession(sessionID string) (SessionData, error) {
	if session, exists := m.sessions[sessionID]; exists {
		return session, nil
	}
	return SessionData{}, errors.New("session not found")
}

func (m *MockSessionManager) GetData(sessionID string) (SessionData, error) {
	return m.LocateSession(sessionID)
}

func (m *MockSessionManager) GetEndpoints(sessionID string) (map[string]string, error) {
	session, err := m.LocateSession(sessionID)
	if err != nil {
		return nil, err
	}
	return session.Endpoints, nil
}

func (m *MockSessionManager) AddApp(appName, version string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, exists := m.sessions[m.sessionID]; exists {
		if session.Apps == nil {
			session.Apps = make(map[string]AppData)
		}
		session.Apps[appName] = AppData{Version: version}
		m.sessions[m.sessionID] = session
		return nil
	}
	return errors.New("session not found")
}

func (m *MockSessionManager) AddEndpoint(sessionID, endpointName, address string) error {
	if session, exists := m.sessions[sessionID]; exists {
		if session.Endpoints == nil {
			session.Endpoints = make(map[string]string)
		}
		session.Endpoints[endpointName] = address
		m.sessions[sessionID] = session
		return nil
	}
	return errors.New("session not found")
}

func (m *MockSessionManager) AddConfigAdjustment(sessionID, app, filename, data string) error {
	if session, exists := m.sessions[sessionID]; exists {
		session.ConfigChanges = append(session.ConfigChanges, ConfigChange{
			App:       app,
			Filename:  filename,
			Data:      data,
			Timestamp: time.Now().Format(time.RFC3339),
		})
		m.sessions[sessionID] = session
		return nil
	}
	return errors.New("session not found")
}
