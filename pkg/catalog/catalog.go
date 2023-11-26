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

type CatalogSourceHolder struct {
	Source CatalogSource
}

func (h *CatalogSourceHolder) SetCatalogSource(cs CatalogSource) {
	h.Source = cs
}

func (h *CatalogSourceHolder) GetCatalogSource() CatalogSource {
	return h.Source
}
