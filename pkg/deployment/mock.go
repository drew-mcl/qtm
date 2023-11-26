package deployment

import (
	"context"
	"qtm/pkg/catalog"
	"qtm/pkg/session"
	"qtm/pkg/suite"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MockDeployer simulates the deployment process
type MockDeployer struct {
	deploymentResults            map[string]map[int]DeploymentResult // deploymentResults stores predefined results for specific app and phase combinations
	mu                           sync.Mutex
	logger                       *zap.Logger
	deployedApps                 map[string]bool
	sleep                        int
	session.SessionManagerHolder // Embedded struct to hold the session manager
	suite.SuiteSourceHolder
	catalog.CatalogSourceHolder
}

// NewMockDeployer creates a new MockDeployer instance
func NewMockDeployer(logger *zap.Logger, sleep int) *MockDeployer {
	return &MockDeployer{
		deploymentResults: make(map[string]map[int]DeploymentResult),
		deployedApps:      make(map[string]bool),
		logger:            logger,
		sleep:             sleep,
	}
}

// SetDeploymentResult allows setting a predefined result for a specific app and phase
func (m *MockDeployer) SetDeploymentResult(appID string, phase int, result DeploymentResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("Setting deployment result", zap.String("appID", appID), zap.Int("phase", phase), zap.Any("status", result.Status))

	if _, exists := m.deploymentResults[appID]; !exists {
		m.deploymentResults[appID] = make(map[int]DeploymentResult)
	}
	m.deploymentResults[appID][phase] = result
}

// Deploy simulates deploying an app in a phase
func (m *MockDeployer) Deploy(ctx context.Context, app suite.SuiteItem, phase int) DeploymentResult {
	m.mu.Lock()
	if result, exists := m.checkPredefinedResult(app.Name, phase); exists {
		m.mu.Unlock()
		return result
	}
	m.mu.Unlock()

	// Fetch version for the app
	catalogSource := m.CatalogSourceHolder.GetCatalogSource()
	data, err := catalogSource.FetchData(app.Name, app.Group)
	if err != nil {
		return DeploymentResult{AppID: app.Name, Phase: phase, Status: Fail, ErrorMsg: err.Error()}
	}
	m.logger.Info("Fetched deployment data", zap.String("appID", app.Name), zap.Int("phase", phase), zap.String("version", data.Version), zap.String("chart", data.HelmChart))

	// Check for cancellation before starting deployment
	if ctx.Err() != nil {
		return DeploymentResult{AppID: app.Name, Phase: phase, Status: Fail, ErrorMsg: "Deployment cancelled"}
	}

	// Perform the actual deployment as part of this instantiation of the deployer
	m.logger.Info("Mocking deploy", zap.String("appID", app.Name), zap.Int("phase", phase), zap.String("version", data.Version), zap.String("chart", data.Name))

	// Check for predefined results first
	if result, exists := m.checkPredefinedResult(app.Name, phase); exists {
		return result
	}
	if m.sleep > 0 {
		time.Sleep(time.Duration(m.sleep) * time.Second)
	}
	m.logger.Info("Mock deploy completed", zap.String("appID", app.Name), zap.Int("phase", phase), zap.String("version", data.Version), zap.String("chart", data.HelmChart))

	// Default to success
	m.deployedApps[app.Name] = true
	return DeploymentResult{AppID: app.Name, Phase: phase, Status: Success}
}

func (m *MockDeployer) checkPredefinedResult(appID string, phase int) (DeploymentResult, bool) {
	if phases, exists := m.deploymentResults[appID]; exists {
		if result, ok := phases[phase]; ok {
			return result, true
		}
	}
	return DeploymentResult{}, false
}

// SetPredefinedResult sets a predefined result for a specific app and phase.
func (m *MockDeployer) SetPredefinedResult(appName string, phase int, result DeploymentResult) {
	if _, exists := m.deploymentResults[appName]; !exists {
		m.deploymentResults[appName] = make(map[int]DeploymentResult)
	}
	m.deploymentResults[appName][phase] = result
}
