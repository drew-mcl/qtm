package session

import (
	"errors"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

type MockSessionManager struct {
	sessionID string
	apps      map[string]AppData
	endpoints map[string]string
	mu        sync.Mutex
	logger    *zap.Logger
}

func NewMockSessionManager(l *zap.Logger) *MockSessionManager {
	return &MockSessionManager{
		apps:      make(map[string]AppData),
		endpoints: make(map[string]string),
		logger:    l,
	}
}

func (m *MockSessionManager) NewSession() {
	m.logger.Info("Creating new session")
	m.sessionID = "mock-session-id"
}

func (m *MockSessionManager) SetSessionID(sessionID string) {
	m.logger.Info("Setting session ID", zap.String("sessionID", sessionID))
	m.sessionID = sessionID
}

func (m *MockSessionManager) GetSessionID() string {
	return m.sessionID
}

func (m *MockSessionManager) GetSessions() ([]string, error) {
	return []string{"mock-session-id"}, nil
}

func (m *MockSessionManager) RemoveApp(appName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.apps[appName]; !exists {
		return fmt.Errorf("app %s does not exist in the session", appName)
	}

	delete(m.apps, appName)
	return nil
}

func (m *MockSessionManager) SetSession(sessionID string) error {
	m.logger.Info("Setting session", zap.String("sessionID", sessionID))
	m.sessionID = sessionID
	return nil
}

func (m *MockSessionManager) RemoveEndpoint(endpointName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.endpoints, endpointName)
	return nil
}

func (m *MockSessionManager) IsEmpty() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.apps) == 0
}

func (m *MockSessionManager) RemoveSession() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.apps = make(map[string]AppData)
	m.endpoints = make(map[string]string)
	return nil
}

func (m *MockSessionManager) UpdateAppDeploymentStatus(appName string, isDeployed bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.apps[appName]; !exists {
		return fmt.Errorf("app %s does not exist in the session", appName)
	}

	app := m.apps[appName]
	app.IsDeployed = isDeployed
	m.apps[appName] = app

	return nil
}

func (m *MockSessionManager) LocateSession(sessionID string) (SessionData, error) {
	return SessionData{
		SessionID:     m.sessionID,
		Apps:          m.apps,
		Endpoints:     m.endpoints,
		ConfigChanges: make([]ConfigChange, 0),
	}, nil
}

func (m *MockSessionManager) GetData() (SessionData, error) {
	m.logger.Info("Getting session data", zap.String("sessionID", m.sessionID))
	return m.LocateSession(m.sessionID)
}

func (m *MockSessionManager) GetEndpoints(sessionID string) (map[string]string, error) {
	m.logger.Info("Getting session endpoints", zap.String("sessionID", sessionID))
	return m.endpoints, nil
}

func (m *MockSessionManager) AddApp(appName string, version string, rolloutPhase int) error {
	m.logger.Info("Adding app", zap.String("appName", appName), zap.String("version", version))
	m.mu.Lock()
	defer m.mu.Unlock()

	m.apps[appName] = AppData{
		Version:      version,
		IsDeployed:   false,
		RolloutPhase: 0,
	}

	return nil
}

func (m *MockSessionManager) AddEndpoint(sessionID, endpointName, address string) error {
	m.logger.Info("Adding endpoint", zap.String("sessionID", sessionID), zap.String("endpointName", endpointName), zap.String("address", address))
	m.mu.Lock()
	defer m.mu.Unlock()
	m.endpoints[endpointName] = address
	return nil
}

func (m *MockSessionManager) AddConfigAdjustment(sessionID, app, filename, data string) error {
	//m.logger.Info("Adding config adjustment", zap.String("sessionID", sessionID), zap.String("app", app), zap.String("filename", filename), zap.String("data", data))
	//if session, exists := m.sessions[sessionID]; exists {
	//	session.ConfigChanges = append(session.ConfigChanges, ConfigChange{
	//		App:       app,
	//		Filename:  filename,
	//		Data:      data,
	//		Timestamp: time.Now().Format(time.RFC3339),
	//	})
	//	m.sessions[sessionID] = session
	//	return nil
	//}
	return errors.New("session not found")
}

func (m *MockSessionManager) GetAppVersion(appName string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if app, exists := m.apps[appName]; exists {
		return app.Version, nil
	}
	return "", errors.New("app not found")
}
