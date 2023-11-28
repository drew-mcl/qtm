package rollback

import (
	"context"
	"fmt"
	"os"
	"qtm/pkg/lifecycle"
	"qtm/pkg/rollback"
	"qtm/pkg/session"
	"qtm/pkg/suite"

	"github.com/spf13/cobra"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type RollbackOptions struct {
	Session     string
	Namespace   string
	StopAt      int
	Suite       string
	UseMockData bool
	suiteFile   string
	DryRun      bool
	endpoint    string
}

func NewRollbackCmd(ctx context.Context, etcdClient *clientv3.Client, logger *zap.Logger) *cobra.Command {
	var rollbackOpts RollbackOptions

	rollbackCmd := &cobra.Command{
		Use:   "rollback",
		Short: "Rollback suites from a session",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			rollbackOpts.Suite = args[0]
			runRollback(ctx, rollbackOpts, etcdClient, logger)
		},
	}

	rollbackCmd.Flags().StringVar(&rollbackOpts.Namespace, "namespace", "", "defines namespace to operate in")
	rollbackCmd.Flags().IntVar(&rollbackOpts.StopAt, "stop-at", 0, "defines a phase for the rollback to stop at")
	rollbackCmd.Flags().BoolVar(&rollbackOpts.UseMockData, "mock", false, "Use mock data for testing")
	rollbackCmd.Flags().StringVar(&rollbackOpts.suiteFile, "suite-file", "", "Use local file to upload suite data")
	rollbackCmd.Flags().BoolVar(&rollbackOpts.DryRun, "dry-run", false, "Perform a mock deployment without any real changes")
	rollbackCmd.Flags().StringVar(&rollbackOpts.endpoint, "endpoint", "localhost:2379", "Etcd endpoint")

	return rollbackCmd
}

func runRollback(ctx context.Context, opts RollbackOptions, etcdClient *clientv3.Client, logger *zap.Logger) {
	fmt.Printf("Rolling back '%s' in namespace '%s' with stop-at phase %d\n", opts.Suite, opts.Namespace, opts.StopAt)

	sm, err := session.NewEtcdSessionManager([]string{opts.endpoint}, "qtm", "user")
	if err != nil {
		logger.Error("Error creating session manager", zap.Error(err))
		os.Exit(1)
	}

	if sm == nil {
		logger.Fatal("Session manager is nil, this should not happen")
	}

	// Initialize and configure rollbacker with appropriate sources
	rollbacker, err := initializeRollback(opts, etcdClient, sm, logger)
	if err != nil {
		fmt.Println("Error initializing rollbacker:", err)
		os.Exit(1)
	}

	// Session selection and handling
	sessionManager := rollbacker.GetSessionManager()
	sessionID, err := session.ChooseSession(logger, sessionManager)
	if err != nil {
		fmt.Println("Error choosing session:", err)
		os.Exit(1)
	}
	sessionManager.SetSessionID(sessionID)
	logger.Info("Session registered", zap.String("sessionID", sessionID))

	// Fetch suite data from configured source
	suiteSource := rollbacker.GetSuiteSource()
	s, err := suiteSource.FetchSuite()
	if err != nil {
		fmt.Println("Error reading data:", err)
		os.Exit(1)
	}

	// Create PhaseInfo from Suite - HACK
	phaseInfos := lifecycle.CreatePhaseInfoFromSuite(s)

	// Call rollback function with the constructed PhaseInfo
	lifecycle.RollbackAllPhases(ctx, rollbacker, phaseInfos, opts.StopAt, logger)

	// Remove session
	err = sessionManager.RemoveSession()
	if err != nil {
		fmt.Println("Error removing session:", err)
		os.Exit(1)
	}
}

func initializeRollback(opts RollbackOptions, etcdClient *clientv3.Client, sm session.SessionManager, logger *zap.Logger) (rollback.Rollbacker, error) {
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
