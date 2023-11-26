package suite

import (
	"errors"
	"os"

	"gopkg.in/yaml.v2"
)

type FileSource struct {
	SuiteName string
	Filename  string
	suite     Suite
}

func NewFileSuiteSource(filePath string) (*FileSource, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var suite Suite
	if err := yaml.Unmarshal(data, &suite.Items); err != nil {
		return nil, err
	}

	return &FileSource{
		suite: suite,
	}, nil

}

func (fds *FileSource) FetchSuite() (Suite, error) {
	if fds.suite.Name == fds.SuiteName {
		return fds.suite, nil
	}
	return Suite{}, errors.New("suite not found")
}
