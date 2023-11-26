package lifecycle

import (
	"context"
	"fmt"
	"os"
	"qtm/pkg/catalog"
	"qtm/pkg/deployment"
	"qtm/pkg/rollback"
	"qtm/pkg/session"
	"qtm/pkg/suite"
	"testing"
	"time"

	"go.uber.org/zap"
)

var logger *zap.Logger

// TestMain initializes the zap logger
func TestMain(m *testing.M) {
	var err error
	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.DebugLevel), // Setting the log level to debug
		Development:      true,
		Encoding:         "console", // or "json", based on preference
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err = cfg.Build()
	if err != nil {
		fmt.Println("Error initializing zap logger:", err)
		os.Exit(1)
	}
	defer logger.Sync()

	defer logger.Sync()

	os.Exit(m.Run())
}

// setupTest creates a mock deployer and rollbacker for testing
func setupTest() (*deployment.MockDeployer, *rollback.MockRollbacker, context.Context, context.CancelFunc) {
	deployer := deployment.NewMockDeployer(logger, 0)
	rollbacker := rollback.NewMockRollbacker(logger)

	//deployer and rollbacker need to share this for moc data as its in memory not an actual source (this took an a bit to long to figure out lol)
	sessionMngr := session.NewMockSessionManager(logger)
	deployer.SetSuiteSource(suite.NewMockSuiteSource())
	deployer.SetCatalogSource(catalog.NewMockCatalogSource())
	deployer.SetSessionManager(sessionMngr)

	rollbacker.SetSessionManager(sessionMngr)
	rollbacker.SetSuiteSource(suite.NewMockSuiteSource())

	ctx, cancel := context.WithCancel(context.Background())
	return deployer, rollbacker, ctx, cancel
}

func TestDeploymentScenarios(t *testing.T) {
	scenarios := []struct {
		name           string
		setupFunc      func(*deployment.MockDeployer)
		decisionMaker  func(int, bool) bool
		expectSuccess  bool
		checkSession   bool             // Whether to check session for app versions
		expectRollback map[int][]string // Expected rollbacks, keyed by phase
	}{
		{
			name: "All Apps Succeed",
			setupFunc: func(deployer *deployment.MockDeployer) {
				// No additional setup required
			},
			decisionMaker: DefaultDecisionMaker,
			expectSuccess: true,
			checkSession:  true,
		},
		{
			name: "Non-Critical Failure",
			setupFunc: func(deployer *deployment.MockDeployer) {
				deployer.SetPredefinedResult("app2-phase", 2, deployment.DeploymentResult{
					AppID:    "app2-phase2",
					Phase:    2,
					Status:   deployment.Fail,
					ErrorMsg: "Simulated failure",
				})
			},
			decisionMaker: func(phase int, phaseSuccess bool) bool {
				return true // Always continue to the next phase
			},
			expectSuccess: true,
			checkSession:  false,
		},
		{
			name: "Halt On Failure",
			setupFunc: func(deployer *deployment.MockDeployer) {
				deployer.SetPredefinedResult("app2-phase2", 2, deployment.DeploymentResult{
					AppID:    "app2-phase2",
					Phase:    2,
					Status:   deployment.Fail,
					ErrorMsg: "Simulated failure",
				})
			},
			decisionMaker: func(phase int, phaseSuccess bool) bool {
				return phaseSuccess // Stop if the phase fails
			},
			expectSuccess: false,
			checkSession:  false,
		},
		// Additional scenarios...
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			deployer, rollbacker, ctx, cancel := setupTest()
			defer cancel()

			scenario.setupFunc(deployer)

			suiteSource := deployer.GetSuiteSource()
			s, err := suiteSource.FetchSuite()
			if err != nil {
				t.Fatalf("Error fetching suite: %v", err)
			}

			suite := suite.OrganizeSuiteData(s)

			success := DeployAllPhases(ctx, deployer, rollbacker, suite, scenario.decisionMaker, false, logger)
			if success != scenario.expectSuccess {
				t.Errorf("Expected success = %v, got %v", scenario.expectSuccess, success)
			}

			// Check session for app versions if required
			if scenario.checkSession {
				sessionManager := deployer.GetSessionManager()
				for _, phaseApps := range suite {
					for _, app := range phaseApps {
						version, err := sessionManager.GetAppVersion(app.Name)
						if err != nil || version == "" {
							t.Errorf("App %s was not added to the session as expected", app.Name)
						}
					}
				}
			}

			// Check rollback expectations
			for phase, apps := range scenario.expectRollback {
				for _, appID := range apps {
					if !rollbacker.IsRolledBack(appID, phase) {
						t.Errorf("Expected rollback of %s in phase %d", appID, phase)
					}
				}
			}
		})
	}
}

// Rollback Current Phase: On failure in phase 2, rollback all successful deployments in that phase.
func TestRollbackCurrentPhase(t *testing.T) {

	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("Error initializing zap logger:", err)
		os.Exit(1)
	}
	defer logger.Sync()

	//Set logger to debug level
	logger = logger.WithOptions(zap.IncreaseLevel(zap.DebugLevel))
	logger.Debug("Debug level enabled")

	decisionMaker := func(phase int, phaseSuccess bool) bool {
		return phaseSuccess // Stop if the phase fails
	}

	deployer := deployment.NewMockDeployer(logger, 0)
	rollbacker := rollback.NewMockRollbacker(logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sessionManager := session.NewMockSessionManager(logger)
	deployer.SetSuiteSource(suite.NewMockSuiteSource())
	deployer.SetCatalogSource(catalog.NewMockCatalogSource())
	deployer.SetSessionManager(sessionManager)

	rollbacker.SetSessionManager(sessionManager)
	rollbacker.SetSuiteSource(suite.NewMockSuiteSource())

	suiteSource := deployer.GetSuiteSource()
	s, err := suiteSource.FetchSuite()
	if err != nil {
		t.Error("Error fetching suite:", err)
	}

	deployer.SetPredefinedResult("app2-phase2", 2, deployment.DeploymentResult{
		AppID:    "app2-phase2",
		Phase:    2,
		Status:   deployment.Fail,
		ErrorMsg: "Simulated failure",
	})

	suite := suite.OrganizeSuiteData(s)

	DeployAllPhases(ctx, deployer, rollbacker, suite, decisionMaker, false, logger)

	// Expected successful apps in each phase
	successfulApps := map[int][]string{
		2: {"app1-phase2", "app3-phase2"}, // app2 fails in phase 2
	}

	// Check if rollback was performed for successful apps in all phases
	for phase, apps := range successfulApps {
		for _, appID := range apps {
			if !rollbacker.IsRolledBack(appID, phase) {
				t.Errorf("Expected rollback of %s in phase %d", appID, phase)
			}
		}
	}

	// Verify app2 is removed from the session after rollback
	_, err = sessionManager.GetAppVersion("app1-phase2")
	if err == nil {
		t.Errorf("App app2 was not removed from the session as expected")
	}
	_, err = sessionManager.GetAppVersion("app2-phase2")
	if err == nil {
		t.Errorf("App app2 was not removed from the session as expected")
	}
	_, err = sessionManager.GetAppVersion("app3-phase2")
	if err == nil {
		t.Errorf("App app2 was not removed from the session as expected")
	}
}

// Rollback All Phases: On failure in phase 2, rollback all successful deployments in phases 1 and 2.
func TestRollbackAllPhases(t *testing.T) {

	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("Error initializing zap logger:", err)
		os.Exit(1)
	}
	defer logger.Sync()

	decisionMaker := func(phase int, phaseSuccess bool) bool {
		return phaseSuccess // Stop if the phase fails
	}

	deployer := deployment.NewMockDeployer(logger, 0)
	rollbacker := rollback.NewMockRollbacker(logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sessionManager := session.NewMockSessionManager(logger)
	deployer.SetSuiteSource(suite.NewMockSuiteSource())
	deployer.SetCatalogSource(catalog.NewMockCatalogSource())
	deployer.SetSessionManager(sessionManager)

	rollbacker.SetSessionManager(sessionManager)
	rollbacker.SetSuiteSource(suite.NewMockSuiteSource())

	deployer.SetPredefinedResult("app2-phase2", 2, deployment.DeploymentResult{
		AppID:    "app2-phase2",
		Phase:    2,
		Status:   deployment.Fail,
		ErrorMsg: "Simulated failure",
	})

	suiteSource := deployer.GetSuiteSource()
	s, err := suiteSource.FetchSuite()
	if err != nil {
		t.Error("Error fetching suite:", err)
	}
	suite := suite.OrganizeSuiteData(s)

	DeployAllPhases(ctx, deployer, rollbacker, suite, decisionMaker, true, logger)

	// Expected successful apps in each phase
	successfulApps := map[int][]string{
		1: {"app1-phase1", "app2-phase1", "app3-phase1"},
		2: {"app1-phase2", "app3-phase2"}, // app2 fails in phase 2
	}

	// Check if rollback was performed for successful apps in all phases
	for phase, apps := range successfulApps {
		for _, appID := range apps {
			if !rollbacker.IsRolledBack(appID, phase) {
				t.Errorf("Expected rollback of %s in phase %d", appID, phase)
			}
		}
	}
}

// Cancellation Signal: Test that a cancellation signal stops all running goroutines immediately.
func TestImmediateCancellation(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Error initializing zap logger: %v", err)
	}
	defer logger.Sync()

	deployer := deployment.NewMockDeployer(logger, 3)
	rollbacker := rollback.NewMockRollbacker(logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sessionManager := session.NewMockSessionManager(logger)
	deployer.SetSuiteSource(suite.NewMockSuiteSource())
	deployer.SetCatalogSource(catalog.NewMockCatalogSource())
	deployer.SetSessionManager(sessionManager)

	rollbacker.SetSessionManager(sessionManager)
	rollbacker.SetSuiteSource(suite.NewMockSuiteSource())

	suiteSource := deployer.GetSuiteSource()
	s, err := suiteSource.FetchSuite()
	if err != nil {
		t.Error("Error fetching suite:", err)
	}

	suite := suite.OrganizeSuiteData(s)
	// Start the deployment process in a separate goroutine
	go func() {
		DeployAllPhases(ctx, deployer, rollbacker, suite, DefaultDecisionMaker, false, logger)
	}()

	// Wait for a short duration before cancelling to ensure deployment starts
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Add a short delay to give time for the cancellation to propagate
	time.Sleep(100 * time.Millisecond)

	// Check for cancellation
	if ctx.Err() == nil {
		t.Errorf("Expected deployment to be cancelled")
	}
}

func TestCancellationWithRollback(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Error initializing zap logger: %v", err)
	}
	defer logger.Sync()

	deployer := deployment.NewMockDeployer(logger, 3)
	rollbacker := rollback.NewMockRollbacker(logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sessionManager := session.NewMockSessionManager(logger)
	deployer.SetSuiteSource(suite.NewMockSuiteSource())
	deployer.SetCatalogSource(catalog.NewMockCatalogSource())
	deployer.SetSessionManager(sessionManager)

	rollbacker.SetSessionManager(sessionManager)
	rollbacker.SetSuiteSource(suite.NewMockSuiteSource())

	suiteSource := deployer.GetSuiteSource()
	s, err := suiteSource.FetchSuite()
	if err != nil {
		t.Error("Error fetching suite:", err)
	}

	suite := suite.OrganizeSuiteData(s)

	// Start the deployment process in a separate goroutine
	go func() {
		DeployAllPhases(ctx, deployer, rollbacker, suite, DefaultDecisionMaker, false, logger)
	}()

	// Wait for a short duration before cancelling to ensure deployment starts
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Add a short delay to give time for the cancellation to propagate
	time.Sleep(100 * time.Millisecond)

	// Check for cancellation
	if ctx.Err() == nil {
		t.Errorf("Expected deployment to be cancelled")
	}

	// Verify if the apps from the last phase were rolled back
	for appID, rolledBack := range rollbacker.RolledBackApps {
		if !rolledBack {
			t.Errorf("App %s was not rolled back as expected", appID)
		}
	}
}
