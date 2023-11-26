package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"qtm/pkg/catalog"
	"qtm/pkg/deployment"
	"qtm/pkg/lifecycle"
	"qtm/pkg/rollback"
	"qtm/pkg/suite"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("Error initializing zap logger:", err)
		os.Exit(1)
	}
	defer logger.Sync()

	//catalog, err := catalog.NewDataFileCatalog("test/data/catalog.yaml")
	//if err != nil {
	//	log.Fatalf("Error loading catalog: %v", err)
	//}
	catalog := catalog.NewMockCatalog()

	deployer := deployment.NewMockDeployer(logger, catalog, 0)
	rollbacker := rollback.NewMockRollbacker(logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up a channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start a goroutine to listen for interrupt signals
	go func() {
		<-sigChan
		fmt.Println("\nReceived an interrupt, cancelling deployments...")
		cancel()
	}()

	// Collect Data
	dataSource := suite.NewMockSuite()
	s, err := dataSource.FetchData()
	if err != nil {
		fmt.Println("Error reading data:", err)
		os.Exit(1)
	}
	logger.Info("Suite data", zap.Any("suite", s))

	suiteData := suite.OrganizeSuiteData(s)
	logger.Info("Organized suite data", zap.Any("suiteData", suiteData))

	// Deploy phases
	success := lifecycle.DeployAllPhases(ctx, deployer, rollbacker, suiteData, lifecycle.DefaultDecisionMaker, false, logger)

	if success {
		fmt.Println("Deployment completed successfully")
	} else {
		fmt.Println("Deployment failed or was cancelled")
	}
}
