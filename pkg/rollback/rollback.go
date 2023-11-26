package rollback

import (
	"qtm/pkg/session"
	"qtm/pkg/suite"
)

// Rollbacker defines the interface for rolling back deployments
type Rollbacker interface {
	Rollback(releaseName string, phase int)
	IsRolledBack(releaseName string, phase int) bool
	SetSessionManager(manager session.SessionManager)
	GetSessionManager() session.SessionManager
	SetSuiteSource(src suite.SuiteSource)
	GetSuiteSource() suite.SuiteSource
}
