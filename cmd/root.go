package cmd

import (
	"fmt"
	"os"

	"qtm/cmd/rollback"
	"qtm/cmd/rollout"

	"github.com/spf13/cobra"
)

var (
	session string
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "qtm",
		Short: "qtm is a tool to manage, deploy, and rollback distributed systems",
	}

	rootCmd.AddCommand(rollout.NewRolloutCmd())
	rootCmd.AddCommand(rollback.NewRollbackCmd())

	rootCmd.Flags().StringVar(&session, "session", "", "String ID to overwrite dynamically made session")

	return rootCmd
}

// Execute executes the root command.
func Execute() {
	rootCmd := NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
