package session

type SessionData struct {
	SessionID     string
	Apps          map[string]AppData
	Endpoints     map[string]string
	ConfigChanges []ConfigChange
}

type AppData struct {
	Version      string // Version of the app
	IsDeployed   bool   // Indicates whether the app was successfully deployed
	RolloutPhase int    // Indicates the phase of the rollout
}

type ConfigChange struct {
	App       string
	Filename  string
	Data      string
	Timestamp string
}

type SessionManager interface {
	GetSessions() ([]string, error)
	LocateSession(sessionID string) (SessionData, error)
	SetSessionID(sessionID string)
	NewSession()
	RemoveSession() error
	IsEmpty() bool
	GetData() (SessionData, error)
	GetEndpoints(sessionID string) (map[string]string, error)
	GetAppVersion(appName string) (string, error)
	AddApp(appName, version string, rolloutPhase int) error
	RemoveApp(appName string) error
	AddEndpoint(sessionID, endpointName, address string) error
	AddConfigAdjustment(sessionID, app, filename, data string) error
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
