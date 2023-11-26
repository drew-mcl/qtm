package rollout

import (
	"context"
	"fmt"
	"os"
	"qtm/pkg/catalog"
	"qtm/pkg/deployment"
	"qtm/pkg/lifecycle"
	"qtm/pkg/rollback"
	"qtm/pkg/session"
	"qtm/pkg/suite"

	"github.com/spf13/cobra"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type RolloutOptions struct {
	Suite       string
	Session     string
	Config      string
	Local       bool
	Atomic      bool
	Nuclear     bool
	Namespace   string
	StartAt     int
	DryRun      bool
	UseMockData bool
	suiteFile   string
	catalogFile string
	local       bool
}

func NewRolloutCmd(logger *zap.Logger, etcdClient *clientv3.Client) *cobra.Command {
	var rolloutOpts RolloutOptions

	rolloutCmd := &cobra.Command{
		Use:   "rollout",
		Short: "Rollout a suite of deployments",
		Run: func(cmd *cobra.Command, args []string) {
			runRollout(rolloutOpts, etcdClient, logger)
		},
	}

	rolloutCmd.Flags().StringVar(&rolloutOpts.Config, "config", "", "Use local file to upload suite data")
	rolloutCmd.Flags().BoolVar(&rolloutOpts.Local, "local", false, "Indicates remote session fetch data should not be used")
	rolloutCmd.Flags().BoolVar(&rolloutOpts.Atomic, "atomic", false, "Indicates if any aspect of the deployment fails, everything should rollback to last safe space")
	rolloutCmd.Flags().BoolVar(&rolloutOpts.Nuclear, "nuclear", false, "Indicates the entire deployment should be rolled back if one app fails")
	rolloutCmd.Flags().StringVar(&rolloutOpts.Namespace, "namespace", "", "Namespace to perform the operations")
	rolloutCmd.Flags().IntVar(&rolloutOpts.StartAt, "start-at", 0, "Defines which phase to start at")
	rolloutCmd.Flags().BoolVar(&rolloutOpts.DryRun, "dry-run", false, "Perform a mock deployment without any real changes")
	rolloutCmd.Flags().BoolVar(&rolloutOpts.UseMockData, "mock", false, "Use mock data for testing")

	return rolloutCmd
}

func runRollout(opts RolloutOptions, etcdClient *clientv3.Client, logger *zap.Logger) {
	// Your rollout logic here
	fmt.Println("Rollout command executed")
	fmt.Println("Session:", opts.Session)
	fmt.Println("Config:", opts.Config)
	fmt.Println("Local:", opts.Local)
	fmt.Println("Atomic:", opts.Atomic)
	fmt.Println("Nuclear:", opts.Nuclear)
	fmt.Println("Namespace:", opts.Namespace)
	fmt.Println("StartAt:", opts.StartAt)
	fmt.Println("DryRun:", opts.DryRun)
	fmt.Println("UseMockData:", opts.UseMockData)

	deployer, err := initializeDeployer(opts, etcdClient, logger)
	if err != nil {
		fmt.Println("Error initializing deployer:", err)
		os.Exit(1)
	}

	// Fetch data using deployer's suite source
	suiteSource := deployer.GetSuiteSource()
	s, err := suiteSource.FetchSuite()
	if err != nil {
		fmt.Println("Error reading data:", err)
		os.Exit(1)
	}
	logger.Debug("Suite data", zap.Any("suite", s))

	suiteData := suite.OrganizeSuiteData(s)
	logger.Debug("Organized suite data", zap.Any("suiteData", suiteData))

	//Determine if rollback is required by checking atomic and nuclear flags
	rollbackRequired := opts.Atomic || opts.Nuclear
	var rollbacker rollback.Rollbacker
	if rollbackRequired {
		rollbacker, err = initializeRollback(opts, etcdClient, logger)
		if err != nil {
			fmt.Println("Error initializing rollbacker:", err)
			os.Exit(1)
		}
	} else {
		rollbacker = nil
	}

	// Deploy phases
	ctx, _ := context.WithCancel(context.Background())
	success := lifecycle.DeployAllPhases(ctx, deployer, rollbacker, suiteData, lifecycle.DefaultDecisionMaker, false, logger)

	if success {
		fmt.Println("Deployment completed successfully")
	} else {
		fmt.Println("Deployment failed or was cancelled")
	}
}

func initializeDeployer(opts RolloutOptions, etcdClient *clientv3.Client, logger *zap.Logger) (deployment.Deployer, error) {

	var catalogSource catalog.CatalogSource
	var suiteSource suite.SuiteSource
	var sessionManager session.SessionManager
	var err error

	if opts.UseMockData {
		catalogSource = catalog.NewMockCatalogSource()
		suiteSource = suite.NewMockSuiteSource()
		sessionManager = session.NewMockSessionManager()
	} else {
		if opts.suiteFile != "" {
			suiteSource, err = suite.NewFileSuiteSource(opts.suiteFile)
			if err != nil {
				return nil, err
			}
		} else {
			suiteSource = suite.NewRemoteSuiteSource(etcdClient, opts.Suite, "qtm")
		}

		if opts.catalogFile != "" {
			catalogSource, err = catalog.NewFileCatalogSource(opts.catalogFile)
			if err != nil {
				return nil, err
			}
		} else {
			catalogSource = catalog.NewRemoteCatalogSource(etcdClient, "qtm")
		}
	}

	// Initialize and configure MockDeployer or real deployer with appropriate sources
	var deployer deployment.Deployer
	if opts.DryRun {
		deployer = deployment.NewMockDeployer(logger, 0)
	} else {
		return nil, nil
	}

	deployer.SetSessionManager(sessionManager)
	deployer.SetCatalogSource(catalogSource)
	deployer.SetSuiteSource(suiteSource)

	return deployer, nil
}

func initializeRollback(opts RolloutOptions, etcdClient *clientv3.Client, logger *zap.Logger) (rollback.Rollbacker, error) {
	var suiteSource suite.SuiteSource
	var err error

	if opts.UseMockData {
		suiteSource = suite.NewMockSuiteSource()
	} else {
		if opts.suiteFile != "" {
			suiteSource, err = suite.NewFileSuiteSource(opts.suiteFile)
			if err != nil {
				return nil, err
			}
		} else {
			suiteSource = suite.NewRemoteSuiteSource(etcdClient, opts.Suite, "qtm")
		}
	}

	logger.Info("Suite data", zap.Any("suite", suiteSource))

	// Initialize and configure MockDeployer or real deployer with appropriate sources
	if opts.DryRun {
		return rollback.NewMockRollbacker(logger), nil
	} else {
		return nil, nil
	}
}
