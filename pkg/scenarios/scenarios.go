package scenarios

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/curious-kitten/scratch-post/internal/store"

	"github.com/curious-kitten/scratch-post/pkg/metadata"
)

//go:generate mockgen -source ./scenarios.go -destination mocks/scenarios.go

// Step represents an action that need to be performed in order to complete a scenario
type Step struct {
	Position        int    `json:"position"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Action          string `json:"action"`
	ExpectedOutcome string `json:"expectedOutcome"`
}

// Scenario is used to define a test case
type Scenario struct {
	Identity      *metadata.Identity     `json:"identity"`
	ProjectID     string                 `json:"projectId"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Prerequisites string                 `json:"prerequisites"`
	Steps         []Step                 `json:"steps"`
	Issues        []metadata.LinkedIssue `json:"issues"`
	Labels        []string               `json:"labels"`
}

// AddIdentity sets the identity of the project
func (s *Scenario) AddIdentity(identity *metadata.Identity) {
	s.Identity = identity
}

// GetIdentity retruns the identity of the project
func (s *Scenario) GetIdentity() *metadata.Identity {
	return s.Identity
}

func (s *Scenario) Validate() error {
	if s.Name == "" {
		return metadata.NewValidationError("name is a mandatory parameter")
	}
	if s.ProjectID == "" {
		return metadata.NewValidationError("projectId is a mandatory parameter")
	}
	return nil
}

// Adder is used to add items to the store
type Adder interface {
	AddOne(ctx context.Context, item interface{}) error
}

// Getter is used to retrieve items from the store
type Getter interface {
	Get(ctx context.Context, id string, item interface{}) error
	GetAll(ctx context.Context, items interface{}) error
}

// Deleter deletes an entry from the collection
type Deleter interface {
	Delete(ctx context.Context, id string) error
}

// IdentityGenerator created and identity to be set on the scenario
type IdentityGenerator interface {
	AddMeta(author string, objType string, identifiable metadata.Identifiable) error
}

// New returns a function used to create a scenario
func New(ig IdentityGenerator, collection Adder, getProject func(ctx context.Context, id string) (interface{}, error)) func(ctx context.Context, author string, scenarioData io.ReadCloser) (interface{}, error) {
	return func(ctx context.Context, author string, scenarioData io.ReadCloser) (interface{}, error) {
		scenario := &Scenario{}
		decoder := json.NewDecoder(scenarioData)
		decoder.DisallowUnknownFields()
		err := decoder.Decode(scenario)
		if err != nil {
			return nil, metadata.NewValidationError(fmt.Sprintf("invalid scenario body: %s", err.Error()))
		}
		if err = scenario.Validate(); err != nil {
			return nil, err
		}
		if _, err = getProject(ctx, scenario.ProjectID); err != nil {
			if store.IsNotFoundError(err) {
				return nil, metadata.NewValidationError("a project with the provided ID does not exist")
			}
			return nil, err
		}
		err = ig.AddMeta(author, "scenario", scenario)
		if err != nil {
			return nil, err
		}
		err = collection.AddOne(ctx, scenario)
		if err != nil {
			return nil, err
		}
		return scenario, nil
	}
}

// List returns a function used to return the scenarios
func List(collection Getter) func(ctx context.Context) ([]interface{}, error) {
	return func(ctx context.Context) ([]interface{}, error) {
		scenarios := []Scenario{}
		err := collection.GetAll(ctx, &scenarios)
		if err != nil {
			return nil, err
		}
		items := make([]interface{}, len(scenarios))
		fmt.Println(len(items))
		for i, v := range scenarios {
			items[i] = v
		}
		return items, nil
	}
}

// Get returns a function to retrieve a scenario based on the passed ID
func Get(collectiom Getter) func(ctx context.Context, id string) (interface{}, error) {
	return func(ctx context.Context, id string) (interface{}, error) {
		scenario := &Scenario{}
		if err := collectiom.Get(ctx, id, scenario); err != nil {
			return nil, err
		}
		return scenario, nil
	}
}

// Delete returns a function to delete a scenario based on the passed ID
func Delete(collection Deleter) func(ctx context.Context, id string) error {
	return func(ctx context.Context, id string) error {
		if err := collection.Delete(ctx, id); err != nil {
			return err
		}
		return nil
	}
}
