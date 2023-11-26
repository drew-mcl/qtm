package suite

import (
	"os"

	"gopkg.in/yaml.v2"
)

type FileDataSource struct {
	SuiteName string
	Filename  string
}

func (fds *FileDataSource) FetchSuite() (*Suite, error) {
	data, err := os.ReadFile(fds.Filename)
	if err != nil {
		return nil, err
	}

	var suite Suite
	if err := yaml.Unmarshal(data, &suite.Items); err != nil {
		return nil, err
	}

	return &suite, nil
}
