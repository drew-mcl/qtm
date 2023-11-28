package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"qtm/cmd"
	"syscall"

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

	// Initialize etcd client
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"}, // Replace with your etcd endpoints
	})
	if err != nil {
		logger.Fatal("Failed to create etcd client", zap.Error(err))
	}

	rootCmd := cmd.NewRootCmd(ctx, etcdClient, logger)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		logger.Fatal("Failed to execute root command", zap.Error(err))
	}
}
