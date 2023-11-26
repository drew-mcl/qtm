package rollback

import (
	"fmt"

	"github.com/spf13/cobra"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type RollbackOptions struct {
	Session   string
	Namespace string
	StopAt    int
}

func NewRollbackCmd(logger *zap.Logger, etcdClient *clientv3.Client) *cobra.Command {
	var rollbackOpts RollbackOptions

	rollbackCmd := &cobra.Command{
		Use:   "rollback",
		Short: "Rollback suites from a session",
		Run: func(cmd *cobra.Command, args []string) {
			runRollback(rollbackOpts, args)
		},
	}

	rollbackCmd.Flags().StringVar(&rollbackOpts.Namespace, "namespace", "", "defines namespace to operate in")
	rollbackCmd.Flags().IntVar(&rollbackOpts.StopAt, "stop-at", 0, "defines a phase for the rollback to stop at")

	return rollbackCmd
}

func runRollback(opts RollbackOptions, args []string) {
	argument := args[0]
	fmt.Printf("Rolling back '%s' in namespace '%s' with stop-at phase %d\n", argument, opts.Namespace, opts.StopAt)
	// TODO: Implement rollback logic here
}
