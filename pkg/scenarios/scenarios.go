package scenarios

import (
	"sync"

	"github.com/curious-kitten/scratch-post/pkg/definitions"
)

type Step struct {
	Position        int    `json:"position"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Action          string `json:"action"`
	ExpectedOutcome string `json:"expectedOutcome"`
}

// Scenario is used to define a test case
type Scenario struct {
	Identity      definitions.Identity      `json:"identity"`
	Name          string                    `json:"name"`
	Description   string                    `json:"description"`
	Prerequisites string                    `json:"prerequisites"`
	Steps         []Step                    `json:"steps"`
	Issues        []definitions.LinkedIssue `json:"issues"`
	Labels        []string                  `json:"labels"`
	sync.RWMutex
}

// NewScenario creates a new Scenario
func NewScenario(name, description, prerequisites string, identity definitions.Identity) *Scenario {
	return &Scenario{
		Identity:      identity,
		Name:          name,
		Description:   description,
		Prerequisites: prerequisites,
		Steps:         []Step{},
	}
}

type Store interface {
	Add(scenario interface{})
	Get(id *definitions.Identity) (*Scenario, error)
}

type ScenarioStore struct {
	store Store
}

func (sm *ScenarioStore) AddScenario(scenario *Scenario) {
	sm.store.Add(scenario)
}

func (sm *ScenarioStore) GetScenario(id *definitions.Identity) (*Scenario, error) {
	s, err := sm.store.Get(id)
	if err != nil {
		return nil, err
	}
	return s, nil
}
