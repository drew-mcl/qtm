package catalog

// Catalog represents a collection of app catalog entries.
type CatalogItem struct {
	Name      string `yaml:"name"`
	Version   string `yaml:"version"`
	HelmChart string `yaml:"helmChart"`
}

type Catalog struct {
	items map[string]CatalogItem // map app names to their catalog items
}

// CatalogSource defines an interface for types that can read catalog data.
type CatalogSource interface {
	FetchData(appName, appGroup string) (*CatalogItem, error)
}
