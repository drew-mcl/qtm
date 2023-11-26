package rollback

import (
	"context"
	"qtm/pkg/session"
	"qtm/pkg/suite"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MockRollbacker simulates the rollback process
type MockRollbacker struct {
	// rolledBack keeps track of rollback status of apps by phase
	rolledBack     map[string]map[int]bool
	mu             sync.Mutex
	logger         *zap.Logger
	RolledBackApps map[string]bool
	session.SessionManagerHolder
	suite.SuiteSourceHolder
	sleep int
}

// NewMockRollbacker creates a new MockRollbacker instance
func NewMockRollbacker(logger *zap.Logger) *MockRollbacker {
	return &MockRollbacker{
		rolledBack:     make(map[string]map[int]bool),
		RolledBackApps: make(map[string]bool),
		logger:         logger,
		sleep:          0,
	}
}

// Rollback simulates rolling back an app deployment
func (m *MockRollbacker) Rollback(ctx context.Context, appName string, phase int, logger *zap.Logger) RollbackResult {
	logger.Info("Performing mock rollback", zap.String("releaseName", appName), zap.Int("phase", phase))
	if ctx.Err() != nil {
		return RollbackResult{AppID: appName, Phase: phase, Status: RollbackFail, ErrorMsg: "Rollback cancelled"}
	}

	if m.sleep > 0 {
		time.Sleep(time.Duration(m.sleep) * time.Second)
	}

	if _, exists := m.rolledBack[appName]; !exists {
		m.rolledBack[appName] = make(map[int]bool)
	}

	// Simulating the rollback action
	m.mu.Lock()
	m.rolledBack[appName][phase] = true
	m.RolledBackApps[appName] = true
	m.mu.Unlock()

	// Simulate the rollback action
	logger.Info("Rolling back app", zap.String("releaseName", appName), zap.Int("phase", phase))
	return RollbackResult{AppID: appName, Phase: phase, Status: RollbackSuccess}

}

// IsRolledBack checks if a specific app has been rolled back in a specific phase
func (m *MockRollbacker) IsRolledBack(releaseName string, phase int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("Checking for rollback", zap.String("release_name", releaseName), zap.Int("phase", phase))

	if phases, exists := m.rolledBack[releaseName]; exists {
		return phases[phase]
	}
	return false
}
