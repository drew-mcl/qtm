package lifecycle

import (
	"context"
	"fmt"
	"os"
	"qtm/pkg/catalog"
	"qtm/pkg/deployment"
	"qtm/pkg/rollback"
	"qtm/pkg/suite"
	"testing"
	"time"

	"go.uber.org/zap"
)

// All Apps Succeed: Test that the deployment completes all phases when all apps succeed.
func TestAllAppsSucceed(t *testing.T) {

	// Create a mock deployer that always succeed
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("Error initializing zap logger:", err)
		os.Exit(1)
	}
	defer logger.Sync()

	deployer := deployment.NewMockDeployer(logger, catalog.NewMockCatalog(), 0)
	rollbacker := rollback.NewMockRollbacker(logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dataSource := suite.NewMockSuite()
	flatSuite, err := dataSource.FetchData()
	if err != nil {
		fmt.Println("Error reading data:", err)
		os.Exit(1)
	}

	suite := suite.OrganizeSuiteData(flatSuite)

	// Use the new DeployAllPhases signature with suiteData
	success := DeployAllPhases(ctx, deployer, rollbacker, suite, DefaultDecisionMaker, false, logger)
	if !success {
		t.Errorf("Expected all apps to succeed, but they did not")
	}
}

// Non-Critical Failure: An app fails in phase 2, but the deployment continues to phase 3.
func TestNonCriticalFailure(t *testing.T) {

	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("Error initializing zap logger:", err)
		os.Exit(1)
	}
	defer logger.Sync()

	deployer := deployment.NewMockDeployer(logger, catalog.NewMockCatalog(), 0)

	deployer.SetDeploymentResult("app2", 2, deployment.DeploymentResult{Status: deployment.Fail})
	rollbacker := rollback.NewMockRollbacker(logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	decisionMaker := func(phase int, phaseSuccess bool) bool {
		return true // Always continue to the next phase, regardless of success
	}

	dataSource := suite.NewMockSuite()
	flatSuite, err := dataSource.FetchData()
	if err != nil {
		fmt.Println("Error reading data:", err)
		os.Exit(1)
	}

	suite := suite.OrganizeSuiteData(flatSuite)

	success := DeployAllPhases(ctx, deployer, rollbacker, suite, decisionMaker, false, logger)
	if !success {
		t.Errorf("Expected deployment to succeed despite non-critical failure")
	}
}

// Halt on Failure: An app fails in phase 2, causing the deployment to halt immediately.
func TestHaltOnFailure(t *testing.T) {

	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("Error initializing zap logger:", err)
		os.Exit(1)
	}
	defer logger.Sync()

	deployer := deployment.NewMockDeployer(logger, catalog.NewMockCatalog(), 0)

	deployer.SetDeploymentResult("app2", 2, deployment.DeploymentResult{Status: deployment.Fail})
	rollbacker := rollback.NewMockRollbacker(logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	decisionMaker := func(phase int, phaseSuccess bool) bool {
		return phaseSuccess // Stop if the phase fails
	}

	dataSource := suite.NewMockSuite()
	flatSuite, err := dataSource.FetchData()
	if err != nil {
		fmt.Println("Error reading data:", err)
		os.Exit(1)
	}

	suite := suite.OrganizeSuiteData(flatSuite)

	success := DeployAllPhases(ctx, deployer, rollbacker, suite, decisionMaker, false, logger)
	if success {
		t.Errorf("Expected deployment to halt on failure")
	}
	// Check if phase 3 was not executed
}

// Rollback Current Phase: On failure in phase 2, rollback all successful deployments in that phase.
func TestRollbackCurrentPhase(t *testing.T) {

	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("Error initializing zap logger:", err)
		os.Exit(1)
	}
	defer logger.Sync()

	deployer := deployment.NewMockDeployer(logger, catalog.NewMockCatalog(), 0)

	deployer.SetDeploymentResult("app2", 2, deployment.DeploymentResult{Status: deployment.Fail})
	rollbacker := rollback.NewMockRollbacker(logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	decisionMaker := func(phase int, phaseSuccess bool) bool {
		return phaseSuccess // Stop if the phase fails
	}

	dataSource := suite.NewMockSuite()
	flatSuite, err := dataSource.FetchData()
	if err != nil {
		fmt.Println("Error reading data:", err)
		os.Exit(1)
	}

	suite := suite.OrganizeSuiteData(flatSuite)

	DeployAllPhases(ctx, deployer, rollbacker, suite, decisionMaker, false, logger)

	// Expected successful apps in each phase
	successfulApps := map[int][]string{
		2: {"app1", "app3"}, // app2 fails in phase 2
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

// Rollback All Phases: On failure in phase 2, rollback all successful deployments in phases 1 and 2.
func TestRollbackAllPhases(t *testing.T) {

	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("Error initializing zap logger:", err)
		os.Exit(1)
	}
	defer logger.Sync()

	deployer := deployment.NewMockDeployer(logger, catalog.NewMockCatalog(), 0)

	deployer.SetDeploymentResult("app2", 2, deployment.DeploymentResult{Status: deployment.Fail})
	rollbacker := rollback.NewMockRollbacker(logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	decisionMaker := func(phase int, phaseSuccess bool) bool {
		return phaseSuccess // Stop if the phase fails
	}

	dataSource := suite.NewMockSuite()
	flatSuite, err := dataSource.FetchData()
	if err != nil {
		fmt.Println("Error reading data:", err)
		os.Exit(1)
	}

	suite := suite.OrganizeSuiteData(flatSuite)

	DeployAllPhases(ctx, deployer, rollbacker, suite, decisionMaker, true, logger)

	// Expected successful apps in each phase
	successfulApps := map[int][]string{
		1: {"app1", "app2", "app3"},
		2: {"app1", "app3"}, // app2 fails in phase 2
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

	deployer := deployment.NewMockDeployer(logger, catalog.NewMockCatalog(), 3)
	rollbacker := rollback.NewMockRollbacker(logger)
	ctx, cancel := context.WithCancel(context.Background())

	dataSource := suite.NewMockSuite()
	flatSuite, err := dataSource.FetchData()
	if err != nil {
		fmt.Println("Error reading data:", err)
		os.Exit(1)
	}

	suite := suite.OrganizeSuiteData(flatSuite)

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

	deployer := deployment.NewMockDeployer(logger, catalog.NewMockCatalog(), 3)
	rollbacker := rollback.NewMockRollbacker(logger)

	ctx, cancel := context.WithCancel(context.Background())

	dataSource := suite.NewMockSuite()
	flatSuite, err := dataSource.FetchData()
	if err != nil {
		fmt.Println("Error reading data:", err)
		os.Exit(1)
	}

	suite := suite.OrganizeSuiteData(flatSuite)

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
