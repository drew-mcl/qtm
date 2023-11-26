package suite

import (
	"reflect"
	"testing"
)

// TestOrganizeSuiteData tests the OrganizeSuiteData function
func TestOrganizeSuiteData(t *testing.T) {
	tests := []struct {
		name     string
		suite    Suite
		expected map[int][]SuiteItem
	}{
		{
			name: "Empty suite",
			suite: Suite{
				Items: []SuiteItem{},
			},
			expected: make(map[int][]SuiteItem),
		},
		{
			name: "Single phase",
			suite: Suite{
				Items: []SuiteItem{{Name: "app1", Group: "group1", RolloutPhase: 1}},
			},
			expected: map[int][]SuiteItem{
				1: {{Name: "app1", Group: "group1", RolloutPhase: 1}},
			},
		},
		{
			name: "Multiple phases",
			suite: Suite{
				Items: []SuiteItem{
					{Name: "app1", Group: "group1", RolloutPhase: 2},
					{Name: "app2", Group: "group2", RolloutPhase: 1},
					{Name: "app3", Group: "group3", RolloutPhase: 2},
				},
			},
			expected: map[int][]SuiteItem{
				1: {{Name: "app2", Group: "group2", RolloutPhase: 1}},
				2: {
					{Name: "app1", Group: "group1", RolloutPhase: 2},
					{Name: "app3", Group: "group3", RolloutPhase: 2},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OrganizeSuiteData(tt.suite)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("OrganizeSuiteData() = %v, want %v", got, tt.expected)
			}
		})
	}
}
