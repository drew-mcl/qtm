package cmd

import (
	"context"
	"qtm/cmd/rollback"
	"qtm/cmd/rollout"

	"github.com/spf13/cobra"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

var (
	session string
)

func NewRootCmd(ctx context.Context, etcdClient *clientv3.Client, logger *zap.Logger) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "qtm",
		Short: "qtm is a tool to manage, deploy, and rollback distributed systems",
	}

	rootCmd.AddCommand(rollout.NewRolloutCmd(ctx, etcdClient, logger))
	rootCmd.AddCommand(rollback.NewRollbackCmd(ctx, etcdClient, logger))

	rootCmd.Flags().StringVar(&session, "session", "", "String ID to overwrite dynamically made session")

	return rootCmd
}
