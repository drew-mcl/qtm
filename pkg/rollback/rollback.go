package rollback

import (
	"context"
	"qtm/pkg/session"
	"qtm/pkg/suite"

	"go.uber.org/zap"
)

// RollbackResult stores the result of a rollback operation
type RollbackResult struct {
	AppID    string
	Phase    int
	Status   RollbackStatus
	ErrorMsg string
}

// RollbackStatus represents the status of a rollback
type RollbackStatus int

const (
	RollbackSuccess RollbackStatus = iota
	RollbackFail
)

// Rollbacker defines the interface for rolling back deployments
type Rollbacker interface {
	Rollback(ctx context.Context, appName string, phase int, logger *zap.Logger) RollbackResult
	IsRolledBack(releaseName string, phase int) bool
	SetSessionManager(manager session.SessionManager)
	GetSessionManager() session.SessionManager
	SetSuiteSource(src suite.SuiteSource)
	GetSuiteSource() suite.SuiteSource
}

// RollbackApp performs the rollback of a single app
func RollbackApp(ctx context.Context, rb Rollbacker, appName string, phase int, logger *zap.Logger) {

	// Check for cancellation before starting deployment
	if ctx.Err() != nil {
		logger.Info("Rollback cancelled", zap.String("releaseName", appName), zap.Int("phase", phase))
	}

	// Perform the actual rollback as part of this instantiation of the deployer
	result := rb.Rollback(ctx, appName, phase, logger)

	if result.Status == RollbackSuccess {
		rb.GetSessionManager().RemoveApp(appName)
		logger.Info("Removed app from session", zap.String("appID", appName))
	} else {
		logger.Error("Rollback failed not removing from session", zap.String("releaseName", appName), zap.Int("phase", phase), zap.Any("status", result.Status))
	}
	return
}
