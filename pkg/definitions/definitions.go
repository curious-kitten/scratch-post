package definitions

import (
	"fmt"
	"sync"
	"time"
)

type severity string

type issueType string

const (
	// Low severity issues should be tagged with this
	Low severity = "LOW"
	// Medium severity issues should be tagged with this
	Medium severity = "MEDIUM"
	// High severity issues should be tagged with this
	High severity = "HIGH"

	// Epic is used when the linked issue type is an Epic
	Epic issueType = "Epic"
	// Story is used when the linked issue is of type Story
	Story issueType = "Story"
	// Defect is used when the linked issue is of type Defect/Bug
	Defect issueType = "Defect"
)

// LinkedIssue are used to identify what external resorce the currect test item refers to
type LinkedIssue struct {
	Link     string    `json:"link"`
	Severity severity  `json:"severity"`
	Type     issueType `json:"type"`
	State    string    `json:"state"`
}

// Identity represents information to identify the given item
type Identity struct {
	ID           string    `json:"id"`
	Version      int       `json:"version"`
	CreatedBy    string    `json:"createdBy"`
	UpdatedBy    string    `json:"updatedBy"`
	CreationTime time.Time `json:"creationTime"`
	UpdateTime   time.Time `json:"updateTime"`
}

// Project represents a umbrella for tests that refer to the same product
type Project struct {
	Identity    Identity `json:"identity"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	counter     counter
}

// MakeIdentity creates an identity to be used inside the project
func (p *Project) MakeIdentity(author string) *Identity {
	return &Identity{
		ID:           fmt.Sprintf("%s-%d", p.Identity.ID, p.counter.up()),
		Version:      1,
		CreatedBy:    author,
		CreationTime: time.Now(),
	}
}

type counter struct {
	sync.RWMutex
	index int
}

func (c *counter) up() int {
	c.Lock()
	defer c.Unlock()
	c.index++
	return c.index
}

// Step is one of the actions that need to be performed in order to verify a behavior
type Step struct {
	Identity        Identity `json:"identity"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Action          string   `json:"action"`
	ExpectedOutcome string   `json:"expectedOutcome"`
}

// Scenario is used to define a test case
type Scenario struct {
	Identity      Identity            `json:"identity"`
	Name          string              `json:"name"`
	Description   string              `json:"description"`
	Prerequisites string              `json:"prerequisites"`
	Steps         map[string]Identity `json:"steps"`
	Issues        []LinkedIssue       `json:"issues"`
	Labels        []string            `json:"labels"`
	sync.RWMutex
}

// NewScenario creates a new Scenario
func NewScenario(name, description, prerequisites string, identity Identity) *Scenario {
	return &Scenario{
		Identity:      identity,
		Name:          name,
		Description:   description,
		Prerequisites: prerequisites,
		Steps:         map[string]Identity{},
	}
}

// UpdateSteps adds steps to the scenario
func (s *Scenario) UpdateSteps(steps ...Step) {
	s.Lock()
	defer s.Unlock()
	for _, step := range steps {
		s.Steps[step.Identity.ID] = step.Identity
	}
}

// RemoveSteps removes steps from the scenario
func (s *Scenario) RemoveSteps(steps ...Step) {
	s.Lock()
	defer s.Unlock()
	for _, step := range steps {
		_, ok := s.Steps[step.Identity.ID]
		if ok {
			delete(s.Steps, step.Identity.ID)
		}
	}
}
