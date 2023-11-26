package suite

import "sort"

type SuiteItem struct {
	Name         string `yaml:"name"`
	Group        string `yaml:"group"`
	RolloutPhase int    `yaml:"rolloutPhase"`
}

type Suite struct {
	Name  string
	Items []SuiteItem `yaml:",inline"`
}

type SuiteSource interface {
	FetchSuite() (Suite, error)
}

// organizeSuiteData organizes the suite data by phase
func OrganizeSuiteData(s Suite) map[int][]SuiteItem {
	phaseData := make(map[int][]SuiteItem)
	var phases []int

	// Collect suite items into the map
	for _, item := range s.Items {
		phaseData[item.RolloutPhase] = append(phaseData[item.RolloutPhase], item)
		if !contains(phases, item.RolloutPhase) {
			phases = append(phases, item.RolloutPhase)
		}
	}

	// Sort phases
	sort.Ints(phases)

	// Reorder the map based on sorted phases
	sortedPhaseData := make(map[int][]SuiteItem)
	for _, phase := range phases {
		sortedPhaseData[phase] = phaseData[phase]
	}

	return sortedPhaseData
}

// contains checks if an int slice contains a specific int
func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
