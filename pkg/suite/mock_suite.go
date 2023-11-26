package suite

import "errors"

type MockSuiteOption func(*MockSuite)

func WithNormalBehavior() MockSuiteOption {
	return func(ms *MockSuite) {
		// Set up normal behavior
		ms.fetchDataFunc = fetchDataNormal
	}
}

func WithErrorBehavior() MockSuiteOption {
	return func(ms *MockSuite) {
		// Set up error behavior
		ms.fetchDataFunc = fetchDataError
	}
}

type MockSuite struct {
	fetchDataFunc func() (Suite, error)
}

func NewMockSuiteSource(opts ...MockSuiteOption) *MockSuite {
	ms := &MockSuite{
		fetchDataFunc: fetchDataNormal,
	}
	for _, opt := range opts {
		opt(ms)
	}
	return ms
}

func (ms *MockSuite) FetchSuite() (Suite, error) {
	return ms.fetchDataFunc()
}

func fetchDataNormal() (Suite, error) {
	return Suite{
		Name: "mock",
		Items: []SuiteItem{
			{Name: "app1", Group: "test", RolloutPhase: 1},
			{Name: "app2", Group: "test", RolloutPhase: 1},
			{Name: "app3", Group: "test", RolloutPhase: 1},
			{Name: "app1", Group: "test", RolloutPhase: 2},
			{Name: "app2", Group: "test", RolloutPhase: 2},
			{Name: "app3", Group: "test", RolloutPhase: 2},
			{Name: "app1", Group: "test", RolloutPhase: 3},
			{Name: "app2", Group: "test", RolloutPhase: 3},
			{Name: "app3", Group: "test", RolloutPhase: 3},
		},
	}, nil
}

func fetchDataError() (Suite, error) {
	return Suite{}, errors.New("Mock suite is operating in error mode: Signalling not found")
}
