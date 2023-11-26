package rollback

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// MockRollbacker simulates the rollback process
type MockRollbacker struct {
	// rolledBack keeps track of rollback status of apps by phase
	rolledBack     map[string]map[int]bool
	mu             sync.Mutex
	logger         *zap.Logger
	RolledBackApps map[string]bool
}

// NewMockRollbacker creates a new MockRollbacker instance
func NewMockRollbacker(logger *zap.Logger) *MockRollbacker {
	return &MockRollbacker{
		rolledBack:     make(map[string]map[int]bool),
		RolledBackApps: make(map[string]bool),
		logger:         logger,
	}
}

// Rollback simulates rolling back an app deployment
func (m *MockRollbacker) Rollback(releaseName string, phase int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("Performing rollback", zap.String("releaseName", releaseName), zap.Int("phase", phase))

	if _, exists := m.rolledBack[releaseName]; !exists {
		m.rolledBack[releaseName] = make(map[int]bool)
	}
	m.rolledBack[releaseName][phase] = true
	m.RolledBackApps[releaseName] = true

	// Simulate the rollback action
	fmt.Printf("Rolling back app %s from phase %d\n", releaseName, phase)
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
