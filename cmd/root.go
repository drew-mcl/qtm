package cmd

import (
	"qtm/cmd/rollback"
	"qtm/cmd/rollout"

	"github.com/spf13/cobra"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

var (
	session string
)

func NewRootCmd(logger *zap.Logger, etcdClient *clientv3.Client) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "qtm",
		Short: "qtm is a tool to manage, deploy, and rollback distributed systems",
	}

	rootCmd.AddCommand(rollout.NewRolloutCmd(logger, etcdClient))
	rootCmd.AddCommand(rollback.NewRollbackCmd(logger, etcdClient))

	rootCmd.Flags().StringVar(&session, "session", "", "String ID to overwrite dynamically made session")

	return rootCmd
}
