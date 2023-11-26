package catalog

import (
	"errors"
	"os"

	"gopkg.in/yaml.v2"
)

// FileCatalogSource reads catalog data from a file.
type FileCatalog struct {
	Filename string
	items    map[string]CatalogItem
}

func NewDataFileCatalog(filePath string) (*FileCatalog, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var items []CatalogItem
	if err := yaml.Unmarshal(data, &items); err != nil {
		return nil, err
	}

	catalog := &Catalog{
		items: make(map[string]CatalogItem),
	}

	for _, item := range items {
		catalog.items[item.Name] = item
	}

	return &FileCatalog{
		items: catalog.items,
	}, nil
}

func (fc *FileCatalog) FetchData(appName, appGroup string) (CatalogItem, error) {
	if item, ok := fc.items[appName]; ok {
		return item, nil
	}
	return CatalogItem{}, errors.New("app not found")
}
