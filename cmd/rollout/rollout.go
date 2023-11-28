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
	endpoint    string
	NewSession  bool
}

func NewRolloutCmd(ctx context.Context, etcdClient *clientv3.Client, logger *zap.Logger) *cobra.Command {
	var rolloutOpts RolloutOptions

	rolloutCmd := &cobra.Command{
		Use:   "rollout",
		Short: "Rollout a suite of deployments",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			rolloutOpts.Suite = args[0]
			runRollout(ctx, rolloutOpts, etcdClient, logger)
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
	rolloutCmd.Flags().StringVar(&rolloutOpts.suiteFile, "suite-file", "", "Use local file to upload suite data")
	rolloutCmd.Flags().StringVar(&rolloutOpts.catalogFile, "catalog-file", "", "Use local file to upload catalog data")
	rolloutCmd.Flags().StringVar(&rolloutOpts.endpoint, "endpoint", "localhost:2379", "Etcd endpoint")
	rolloutCmd.Flags().BoolVar(&rolloutOpts.NewSession, "new", false, "Indicates a new session should be created")

	return rolloutCmd
}

func runRollout(ctx context.Context, opts RolloutOptions, etcdClient *clientv3.Client, logger *zap.Logger) {
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

	sm, err := session.NewEtcdSessionManager([]string{opts.endpoint}, "qtm", "user")
	if err != nil {
		logger.Error("Error creating session manager", zap.Error(err))
		os.Exit(1)
	}

	if sm == nil {
		logger.Fatal("Session manager is nil, this should not happen")
	}

	deployer, err := initializeDeployer(opts, etcdClient, sm, logger)
	if err != nil {
		fmt.Println("Error initializing deployer:", err)
		os.Exit(1)
	}

	// Create or fetch session
	sessionManager := deployer.GetSessionManager()

	sessionOpts := session.SessionOptions{
		Session:    opts.Session,
		NewSession: opts.NewSession,
	}

	sessionID, err := session.CreateOrFetchSession(logger, sessionManager, sessionOpts)
	if err != nil {
		fmt.Println("Error creating or fetching session:", err)
		os.Exit(1)
	}

	sessionManager.SetSessionID(sessionID)
	sessionManager.RegisterNewSession(sessionID)
	logger.Info("Session created", zap.String("sessionID", sessionID))

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
		rollbacker, err = initializeRollback(opts, etcdClient, sm, logger)
		if err != nil {
			fmt.Println("Error initializing rollbacker:", err)
			os.Exit(1)
		}
	} else {
		rollbacker = nil
	}

	// Deploy phases
	success := lifecycle.DeployAllPhases(ctx, deployer, rollbacker, suiteData, lifecycle.DefaultDecisionMaker, false, logger)

	if success {
		fmt.Println("Deployment completed successfully")
	} else {
		fmt.Println("Deployment failed or was cancelled")
	}
}

func initializeDeployer(opts RolloutOptions, etcdClient *clientv3.Client, sm session.SessionManager, logger *zap.Logger) (deployment.Deployer, error) {

	var catalogSource catalog.CatalogSource
	var suiteSource suite.SuiteSource
	var err error

	if opts.UseMockData {
		catalogSource = catalog.NewMockCatalogSource()
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
		deployer = deployment.NewMockDeployer(logger, 5)
	} else {
		return nil, nil
	}

	deployer.SetCatalogSource(catalogSource)
	deployer.SetSuiteSource(suiteSource)
	deployer.SetSessionManager(sm)

	return deployer, nil
}

func initializeRollback(opts RolloutOptions, etcdClient *clientv3.Client, sm session.SessionManager, logger *zap.Logger) (rollback.Rollbacker, error) {
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

	var rollbacker rollback.Rollbacker
	if opts.DryRun {
		rollbacker = rollback.NewMockRollbacker(logger)
	} else {
		return nil, nil
	}

	rollbacker.SetSuiteSource(suiteSource)
	rollbacker.SetSessionManager(sm)

	return rollbacker, nil
}
