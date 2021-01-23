package scenarios

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/curious-kitten/scratch-post/internal/decoder"
	"github.com/curious-kitten/scratch-post/internal/store"
	"github.com/curious-kitten/scratch-post/pkg/metadata"
)

//go:generate mockgen -source ./scenarios.go -destination mocks/scenarios.go

type projectRetriever func(ctx context.Context, id string) (interface{}, error)

// Step represents an action that need to be performed in order to complete a scenario
type Step struct {
	Position        int    `json:"position,omitempty"`
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	Action          string `json:"action,omitempty"`
	ExpectedOutcome string `json:"expectedOutcome,omitempty"`
}

// Scenario is used to define a test case
type Scenario struct {
	Identity      *metadata.Identity     `json:"identity,omitempty"`
	ProjectID     string                 `json:"projectId,omitempty"`
	Name          string                 `json:"name,omitempty"`
	Description   string                 `json:"description,omitempty"`
	Prerequisites string                 `json:"prerequisites,omitempty"`
	Steps         []Step                 `json:"steps,omitempty"`
	Issues        []metadata.LinkedIssue `json:"issues,omitempty"`
	Labels        []string               `json:"labels,omitempty"`
}

// AddIdentity sets the identity of the project
func (s *Scenario) AddIdentity(identity *metadata.Identity) {
	s.Identity = identity
}

// GetIdentity retruns the identity of the project
func (s *Scenario) GetIdentity() *metadata.Identity {
	return s.Identity
}

// Validate is used to check the integrity of the scenario object
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

// Updater is used to replace information into the Data Base
type Updater interface {
	Update(ctx context.Context, id string, item interface{}) error
}

// ReaderUpdater is used to read and update objects in the Data Base
type ReaderUpdater interface {
	Getter
	Updater
}

// IdentityGenerator created and identity to be set on the scenario
type IdentityGenerator interface {
	AddMeta(author string, objType string, identifiable metadata.Identifiable) error
}

// New returns a function used to create a scenario
func New(ig IdentityGenerator, collection Adder, getProject projectRetriever) func(ctx context.Context, author string, data io.Reader) (interface{}, error) {
	return func(ctx context.Context, author string, data io.Reader) (interface{}, error) {
		scenario := &Scenario{}
		if err := decoder.Decode(scenario, data); err != nil {
			return nil, err
		}
		if _, err := getProject(ctx, scenario.ProjectID); err != nil {
			if store.IsNotFoundError(err) {
				return nil, metadata.NewValidationError("project with the provided ID does not exist")
			}
			return nil, err
		}
		if err := ig.AddMeta(author, "scenario", scenario); err != nil {
			return nil, err
		}

		if err := collection.AddOne(ctx, scenario); err != nil {
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

// Update is used to replace a scenario with the provided scenario
func Update(collection ReaderUpdater, getProject projectRetriever) func(ctx context.Context, user string, id string, data io.Reader) (interface{}, error) {
	return func(ctx context.Context, user string, id string, data io.Reader) (interface{}, error) {
		scenario := &Scenario{}
		if err := decoder.Decode(scenario, data); err != nil {
			return nil, err
		}
		if _, err := getProject(ctx, scenario.ProjectID); err != nil {
			if store.IsNotFoundError(err) {
				return nil, metadata.NewValidationError("project with the provided ID does not exist")
			}
			return nil, err
		}
		foundScenario, err := Get(collection)(ctx, id)
		if err != nil {
			return nil, err
		}
		var s *Scenario
		var ok bool
		if s, ok = foundScenario.(*Scenario); !ok {
			return nil, fmt.Errorf("invalid data structure in DB")
		}
		scenario.Identity = s.Identity
		scenario.Identity.UpdateTime = time.Now()
		scenario.Identity.UpdatedBy = user
		if err := collection.Update(ctx, id, scenario); err != nil {
			return nil, err
		}
		return scenario, nil
	}
}
