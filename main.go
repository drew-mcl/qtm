package main

import (
	"qtm/cmd"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

func main() {
	// Initialize zap logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	// Initialize etcd client
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"}, // Replace with your etcd endpoints
	})
	if err != nil {
		logger.Fatal("Failed to create etcd client", zap.Error(err))
	}

	rootCmd := cmd.NewRootCmd(logger, etcdClient)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		logger.Fatal("Failed to execute root command", zap.Error(err))
	}
}
