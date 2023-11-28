package lifecycle

import (
	"context"
	"qtm/pkg/deployment"
	"qtm/pkg/rollback"
	"qtm/pkg/suite"
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type PhaseInfo struct {
	SuccessfulApps []string // List of app IDs that were successfully deployed in this phase
	IsSuccessful   bool     // Indicates whether the phase was overall successful
}

// DeployAllPhases is a modified function to handle the deployment of all phases
func DeployAllPhases(ctx context.Context, deployer deployment.Deployer, rollbacker rollback.Rollbacker, suiteData map[int][]suite.SuiteItem, decisionMaker func(int, bool) bool, rollbackEverything bool, logger *zap.Logger) bool {
	logger.Info("Starting deployment", zap.Bool("rollbackEverything", rollbackEverything))

	phaseInfos := make(map[int]PhaseInfo)

	for phase, apps := range suiteData {
		logger.Info("Starting phase", zap.Int("phase", phase), zap.Any("apps", apps))

		results := make(chan deployment.DeploymentResult, len(apps))

		var wg sync.WaitGroup
		for _, app := range apps {
			wg.Add(1)
			go func(app suite.SuiteItem) {
				defer wg.Done()
				deployment.DeployApp(ctx, deployer, app, phase, results)
			}(app)
		}

		wg.Wait()
		close(results)

		phaseSuccess, successfulApps := processPhaseResults(results, logger)
		phaseInfos[phase] = PhaseInfo{SuccessfulApps: successfulApps, IsSuccessful: phaseSuccess}

		if ctx.Err() != nil {
			// Context is canceled - perform rollback
			rolbackCtx := context.Background()
			RollbackPhase(rolbackCtx, rollbacker, phase, successfulApps, logger)
			return false
		}

		logger.Info("Phase ended", zap.Int("phase", phase), zap.Any("overall", phaseInfos), zap.Bool("phaseSuccess", phaseSuccess))

		if !decisionMaker(phase, phaseSuccess) {
			if !phaseSuccess {
				logger.Info("Initiating rollback due to phase failure", zap.Int("phase", phase))
				if rollbackEverything {
					logger.Info("Rolling back all phases", zap.Int("phase", phase))
					RollbackAllPhases(ctx, rollbacker, phaseInfos, phase, logger)
				} else {
					logger.Info("Rolling back single phase", zap.Int("phase", phase))
					RollbackPhase(ctx, rollbacker, phase, successfulApps, logger)
				}
			}
			return false
		}

	}
	return true
}

// Implement rollbackPhase and rollbackAllPhases functions
func RollbackPhase(ctx context.Context, rollbacker rollback.Rollbacker, phase int, apps []string, logger *zap.Logger) {
	logger.Info("Rolling back phase", zap.Int("phase", phase), zap.Any("apps", apps))
	var wg sync.WaitGroup

	for _, appID := range apps {
		wg.Add(1)
		go func(appID string) {
			defer wg.Done()

			// Generate a unique ID for this goroutine
			goroutineID := uuid.New().String()

			// Creating a new logger instance for this specific goroutine
			goroutineLogger := logger.With(
				zap.String("appID", appID),
				zap.String("goroutineID", goroutineID),
			)
			goroutineLogger.Info("Starting rollback goroutine")
			rollback.RollbackApp(ctx, rollbacker, appID, phase, goroutineLogger)
		}(appID)
	}

	wg.Wait()
}

// rollbackAllPhases rolls back all phases up to and including the specified phase
func RollbackAllPhases(ctx context.Context, rollbacker rollback.Rollbacker, phaseInfos map[int]PhaseInfo, upToPhase int, logger *zap.Logger) {
	logger.Info("Rolling back all phases", zap.Int("upToPhase", upToPhase))
	for phase := upToPhase; phase >= 0; phase-- {
		info, exists := phaseInfos[phase]
		if !exists {
			continue // Skip if no information about the phase
		}
		logger.Info("Rolling back phase", zap.Int("phase", phase), zap.Any("apps", info.SuccessfulApps))
		var wg sync.WaitGroup
		for _, appID := range info.SuccessfulApps {
			wg.Add(1)
			go func(appID string) {
				defer wg.Done()

				// Generate a unique ID for this goroutine
				goroutineID := uuid.New().String()

				// Creating a new logger instance for this specific goroutine
				goroutineLogger := logger.With(
					zap.String("appID", appID),
					zap.String("goroutineID", goroutineID),
					zap.Int("phase", phase),
				)
				goroutineLogger.Info("Starting rollback goroutine")
				rollback.RollbackApp(ctx, rollbacker, appID, phase, goroutineLogger)
			}(appID)
		}
		wg.Wait()
		logger.Info("Phase rollback completed", zap.Int("phase", phase))
	}
	logger.Info("Rollback of all phases completed", zap.Int("upToPhase", upToPhase))
}

// processPhaseResults processes the results of a deployment phase
func processPhaseResults(results chan deployment.DeploymentResult, logger *zap.Logger) (bool, []string) {
	phaseSuccess := true
	var successfulApps []string

	for res := range results {
		if res.Status == deployment.Fail {
			logger.Error("Deployment failed", zap.String("appID", res.AppID), zap.Int("phase", res.Phase), zap.String("errorMsg", res.ErrorMsg))
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

func CreatePhaseInfoFromSuite(s suite.Suite) map[int]PhaseInfo {
	phaseData := suite.OrganizeSuiteData(s)
	phaseInfos := make(map[int]PhaseInfo)

	for phase, items := range phaseData {
		var appNames []string
		for _, item := range items {
			appNames = append(appNames, item.Name)
		}
		phaseInfos[phase] = PhaseInfo{
			SuccessfulApps: appNames,
			// Assumption: Marking all as successful, needs adjustment based on actual deployment results
			IsSuccessful: true,
		}
	}

	return phaseInfos
}
