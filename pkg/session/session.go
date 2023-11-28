package session

import (
	"fmt"
	"qtm/internal/prompt"
	"qtm/pkg/suite"

	"go.uber.org/zap"
)

type SessionData struct {
	Username      string
	SessionID     string
	Apps          map[string]AppData
	Endpoints     map[string]string
	ConfigChanges []ConfigChange
}

type AppData struct {
	version string
	obj     suite.SuiteItem
}

type ConfigChange struct {
	App       string
	Filename  string
	Data      string
	Timestamp string
}

type SessionOptions struct {
	Session    string
	NewSession bool
}

type SessionManager interface {
	GetSessions() ([]string, error)
	CreateSessionID() (string, error)
	SetSessionID(sessionID string)
	RegisterNewSession(sesisonID string) error
	RemoveSession() error
	ValidateSession() (bool, error)
	AddApp(app suite.SuiteItem, version string) error
	RemoveApp(appName string) error
	AddEndpoint(endpointName, address string) error
	AddConfigAdjustment(app, filename, data string) error
	IsEmpty() bool
	GetAppVersion(appName string) (string, error)
}

// SessionManagerHolder holds a reference to a SessionManager
type SessionManagerHolder struct {
	Manager SessionManager
}

func (h *SessionManagerHolder) SetSessionManager(manager SessionManager) {
	h.Manager = manager
}

func (h *SessionManagerHolder) GetSessionManager() SessionManager {
	return h.Manager
}

// CreateOrFetchSession handles the creation or fetching of a session.
func CreateOrFetchSession(logger *zap.Logger, sm SessionManager, opts SessionOptions) (string, error) {

	if opts.Session != "" {
		return opts.Session, nil
	}

	if opts.NewSession {
		return sm.CreateSessionID()
	}

	return ChooseSession(logger, sm)
}

// Choose handles the logic of choosing an existing session or creating a new one.
func ChooseSession(logger *zap.Logger, sessionManager SessionManager) (string, error) {
	sessions, err := sessionManager.GetSessions()
	if err != nil {
		return "", fmt.Errorf("error fetching sessions: %w", err)
	}

	if len(sessions) == 0 {
		logger.Warn("No sessions found, creating a new one")
		return sessionManager.CreateSessionID()
	}

	sessionID, err := prompt.ShowSelectionPrompt(sessions, "Please select a session to use")
	if err != nil {
		return "", err
	}

	return sessionID, nil
}
