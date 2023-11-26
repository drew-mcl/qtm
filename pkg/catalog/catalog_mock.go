package catalog

import (
	"errors"
)

type MockCatalogOption func(*MockCatalog)

func WithNormalBehavior() MockCatalogOption {
	return func(mc *MockCatalog) {
		mc.lookupFunc = mc.normalLookup
	}
}

func WithCatalogNotFoundError() MockCatalogOption {
	return func(mc *MockCatalog) {
		mc.lookupFunc = mc.catalogNotFoundLookup
	}
}

func WithItemNotFoundError() MockCatalogOption {
	return func(mc *MockCatalog) {
		mc.lookupFunc = mc.itemNotFoundLookup
	}
}

func WithVersionNotFoundError() MockCatalogOption {
	return func(mc *MockCatalog) {
		mc.lookupFunc = mc.versionNotFoundLookup
	}
}

type MockCatalog struct {
	items      map[string]CatalogItem // Maps app names to their catalog items
	lookupFunc func(string) (*CatalogItem, error)
}

func NewMockCatalogSource(opts ...MockCatalogOption) *MockCatalog {
	mc := &MockCatalog{
		items: make(map[string]CatalogItem),
	}

	mc.lookupFunc = func(appName string) (*CatalogItem, error) {
		return mc.normalLookup(appName) // default behavior
	}

	// Add some mock data to mc.items
	mc.items["app1-phase1"] = CatalogItem{Name: "app1-phase1", Version: "1.1.1", HelmChart: "mychartfrommock-1.0.0.tgz"}
	mc.items["app2-phase1"] = CatalogItem{Name: "app2-phase1", Version: "1.1.1", HelmChart: "mychartfrommock-1.0.0.tgz"}
	mc.items["app3-phase1"] = CatalogItem{Name: "app3-phase1", Version: "1.1.1", HelmChart: "mychartfrommock-1.0.0.tgz"}
	mc.items["app1-phase2"] = CatalogItem{Name: "app1-phase2", Version: "1.1.1", HelmChart: "mychartfrommock-1.0.0.tgz"}
	mc.items["app2-phase2"] = CatalogItem{Name: "app2-phase2", Version: "2.2.2", HelmChart: "mychartfrommock-1.0.0.tgz"}
	mc.items["app3-phase2"] = CatalogItem{Name: "app3-phase2", Version: "3.3.3", HelmChart: "mychartfrommock-1.0.0.tgz"}
	mc.items["app1-phase3"] = CatalogItem{Name: "app1-phase3", Version: "1.1.1", HelmChart: "mychartfrommock-1.0.0.tgz"}
	mc.items["app2-phase3"] = CatalogItem{Name: "app2-phase3", Version: "2.2.2", HelmChart: "mychartfrommock-1.0.0.tgz"}
	mc.items["app3-phase3"] = CatalogItem{Name: "app3-phase3", Version: "3.3.3", HelmChart: "mychartfrommock-1.0.0.tgz"}

	for _, opt := range opts {
		opt(mc)
	}

	return mc
}

func (mc *MockCatalog) FetchData(appName, appGroup string) (*CatalogItem, error) {
	return mc.lookupFunc(appName)
}

func (mc *MockCatalog) normalLookup(appName string) (*CatalogItem, error) {
	if item, ok := mc.items[appName]; ok {
		return &item, nil
	}
	return nil, errors.New("item not found")
}

func (mc *MockCatalog) catalogNotFoundLookup(appName string) (*CatalogItem, error) {
	return nil, errors.New("catalog not found")
}

func (mc *MockCatalog) itemNotFoundLookup(appName string) (*CatalogItem, error) {
	return nil, errors.New("item not found")
}

func (mc *MockCatalog) versionNotFoundLookup(appName string) (*CatalogItem, error) {
	return nil, errors.New("version not found")
}
