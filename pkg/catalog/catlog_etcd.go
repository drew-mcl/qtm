package catalog

import (
	"context"
	"fmt"
	"path/filepath"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func GetCatalogItem(appName, appGroup string) (*CatalogItem, error) {
	// Connect to etcd
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"}, // Replace with your etcd endpoints
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to etcd: %v", err)
	}
	defer cli.Close()

	// Build the key
	key := filepath.Join(appGroup, appName)

	// Query /version
	respVersion, err := cli.Get(context.Background(), filepath.Join(key, "version"))
	if err != nil {
		return nil, fmt.Errorf("failed to query /version: %v", err)
	}
	if len(respVersion.Kvs) == 0 {
		return nil, fmt.Errorf("no version found for %s", key)
	}
	version := string(respVersion.Kvs[0].Value)

	// Query /chart
	respChart, err := cli.Get(context.Background(), filepath.Join(key, "chart"))
	if err != nil {
		return nil, fmt.Errorf("failed to query /chart: %v", err)
	}
	if len(respChart.Kvs) == 0 {
		return nil, fmt.Errorf("no chart found for %s", key)
	}
	chart := string(respChart.Kvs[0].Value)

	// Build the catalog item structure
	catalogItem := &CatalogItem{
		Name:      appName,
		Version:   version,
		HelmChart: chart,
	}

	return catalogItem, nil
}
