package lifecycle

import (
	"context"
	"fmt"
	"qtm/pkg/deployment"
	"qtm/pkg/rollback"
	"qtm/pkg/suite"
	"sync"

	"go.uber.org/zap"
)

type PhaseInfo struct {
	SuccessfulApps []string // List of app IDs that were successfully deployed in this phase
	IsSuccessful   bool     // Indicates whether the phase was overall successful
}

// deployApp is a modified function to handle the deployment of a single app
func deployApp(ctx context.Context, deployer deployment.Deployer, appName string, appGroup string, phase int, results chan<- deployment.DeploymentResult) {
	// Check for cancellation before starting deployment
	if ctx.Err() != nil {
		return
	}

	result := deployer.Deploy(ctx, appName, appGroup, phase)
	//if result.Status == deployment.Success {
	// Assuming version and other details are part of the result
	//	deployer.GetSessionManager().AddApp(appName, "1.0.0") // Add the app to the session
	//}
	results <- result
}

// DeployAllPhases is a modified function to handle the deployment of all phases
func DeployAllPhases(ctx context.Context, deployer deployment.Deployer, rollbacker rollback.Rollbacker, suiteData map[int][]suite.SuiteItem, decisionMaker func(int, bool) bool, rollbackEverything bool, logger *zap.Logger) bool {
	logger.Info("Starting deployment", zap.Bool("rollbackEverything", rollbackEverything))

	phaseInfos := make(map[int]PhaseInfo)

	for phase, apps := range suiteData {
		logger.Info("Starting phase", zap.Int("phase", phase), zap.Any("apps", apps))
		results := make(chan deployment.DeploymentResult, len(apps))
		var wg sync.WaitGroup

		allDeployedApps := make([]string, 0, len(apps))
		for _, app := range apps {

			appName := app.Name
			appGroup := app.Group

			logger.Info("Deploying app", zap.String("app_name", appName), zap.String("group", appGroup), zap.Int("phase", phase))
			allDeployedApps = append(allDeployedApps, appName)

			wg.Add(1)
			go func(appName, appGroup string) {
				defer wg.Done()
				deployApp(ctx, deployer, appName, appGroup, phase, results)
			}(appName, appGroup)
		}

		wg.Wait()
		close(results)

		if ctx.Err() != nil {
			fmt.Println("Deployment cancelled")
			rollbackPhase(rollbacker, phase, allDeployedApps) // Rollback the current phase only
			return false
		}

		phaseSuccess, successfulApps := processPhaseResults(results)
		phaseInfos[phase] = PhaseInfo{SuccessfulApps: successfulApps, IsSuccessful: phaseSuccess}

		if !decisionMaker(phase, phaseSuccess) {
			if !phaseSuccess {
				if rollbackEverything {
					rollbackAllPhases(rollbacker, phaseInfos, phase)
				} else {
					rollbackPhase(rollbacker, phase, successfulApps)
				}
			}
			return false
		}
		logger.Info("Phase completed", zap.Int("phase", phase), zap.Bool("phaseSuccess", phaseSuccess))
	}
	logger.Info("Deployment completed successfully")
	return true
}

// Implement rollbackPhase and rollbackAllPhases functions
func rollbackPhase(rollbacker rollback.Rollbacker, phase int, apps []string) {
	fmt.Println("Rolling back phase:", phase)
	for _, appID := range apps {
		rollbacker.Rollback(appID, phase)
	}
}

// rollbackAllPhases rolls back all phases up to and including the specified phase
func rollbackAllPhases(rollbacker rollback.Rollbacker, phaseInfos map[int]PhaseInfo, upToPhase int) {
	for phase := 1; phase <= upToPhase; phase++ {
		info, exists := phaseInfos[phase]
		if !exists {
			continue // Skip if no information about the phase
		}
		fmt.Println("Rolling back phase:", phase)
		for _, appID := range info.SuccessfulApps {
			rollbacker.Rollback(appID, phase)
		}
	}
}

// processPhaseResults processes the results of a deployment phase
func processPhaseResults(results chan deployment.DeploymentResult) (bool, []string) {
	phaseSuccess := true
	var successfulApps []string

	for res := range results {
		if res.Status == deployment.Fail {
			fmt.Printf("App %s failed in phase %d: %s\n", res.AppID, res.Phase, res.ErrorMsg)
			phaseSuccess = false
		} else {
			successfulApps = append(successfulApps, res.AppID)
		}
	}
	return phaseSuccess, successfulApps
}

// defaultDecisionMaker is a default implementation of the decisionMaker function
func DefaultDecisionMaker(phase int, phaseSuccess bool) bool {
	return phaseSuccess // Continue only if the phase is successful
}
