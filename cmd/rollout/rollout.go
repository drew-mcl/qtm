package rollout

import (
	"fmt"
	"qtm/pkg/catalog"
	"qtm/pkg/deployment"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type RolloutOptions struct {
	Session        string
	Config         string
	Local          bool
	Atomic         bool
	Nuclear        bool
	Namespace      string
	StartAt        int
	DryRun         bool
	UseMockCatalog bool
	UseMockSuite   bool
}

func NewRolloutCmd() *cobra.Command {
	var rolloutOpts RolloutOptions

	rolloutCmd := &cobra.Command{
		Use:   "rollout",
		Short: "Rollout a suite of deployments",
		Run: func(cmd *cobra.Command, args []string) {
			runRollout(rolloutOpts)
		},
	}

	rolloutCmd.Flags().StringVar(&rolloutOpts.Config, "config", "", "Use local file to upload session data")
	rolloutCmd.Flags().BoolVar(&rolloutOpts.Local, "local", false, "Indicates remote session fetch data should not be used")
	rolloutCmd.Flags().BoolVar(&rolloutOpts.Atomic, "atomic", false, "Indicates if any aspect of the deployment fails, everything should rollback to last safe space")
	rolloutCmd.Flags().BoolVar(&rolloutOpts.Nuclear, "nuclear", false, "Indicates the entire deployment should be rolled back if one app fails")
	rolloutCmd.Flags().StringVar(&rolloutOpts.Namespace, "namespace", "", "Namespace to perform the operations")
	rolloutCmd.Flags().IntVar(&rolloutOpts.StartAt, "start-at", 0, "Defines which phase to start at")
	rolloutCmd.Flags().BoolVar(&rolloutOpts.DryRun, "dry-run", false, "Perform a mock deployment without any real changes")
	rolloutCmd.Flags().BoolVar(&rolloutOpts.UseMockCatalog, "use-mock-catalog", false, "Use mock catalog for testing")
	rolloutCmd.Flags().BoolVar(&rolloutOpts.UseMockSuite, "use-mock-suite", false, "Use mock suite for testing")

	return rolloutCmd
}

func runRollout(opts RolloutOptions) {
	// Your rollout logic here
	fmt.Println("Rollout command executed")
	fmt.Println("Session:", opts.Session)
	fmt.Println("Config:", opts.Config)
	fmt.Println("Local:", opts.Local)
	fmt.Println("Atomic:", opts.Atomic)
	fmt.Println("Nuclear:", opts.Nuclear)
	fmt.Println("Namespace:", opts.Namespace)
	fmt.Println("StartAt:", opts.StartAt)
}

func initializeDeployer(opts RolloutOptions, logger *zap.Logger, catalog catalog.Catalog) deployment.Deployer {
	if opts.MockDeploy {
		return deployment.NewMockDeployer(logger, catalog, 0)
	}
	// Initialize and return your real deployer here
	// return ...
}
