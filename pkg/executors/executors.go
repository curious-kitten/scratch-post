package executors

import (
	"context"
	"fmt"
	"io"

	"github.com/curious-kitten/scratch-post/internal/decoder"
	"github.com/curious-kitten/scratch-post/internal/store"
	"github.com/curious-kitten/scratch-post/pkg/metadata"
)

//go:generate mockgen -source ./executors.go -destination mocks/executors.go

type getItem func(ctx context.Context, id string) (interface{}, error)

// Executor is used to define a test case
type Executor struct {
	Identity   *metadata.Identity `json:"identity,omitempty"`
	ProjectID  string             `json:"projectId,omitempty"`
	ScenarioID string             `json:"scenarioId,omitempty"`
	TestPlanID string             `json:"testplanId,omitempty"`
}

// AddIdentity sets the identity of the project
func (e *Executor) AddIdentity(identity *metadata.Identity) {
	e.Identity = identity
}

// GetIdentity retruns the identity of the project
func (e *Executor) GetIdentity() *metadata.Identity {
	return e.Identity
}

// Validate is used to check the integrity of the executor object
func (e *Executor) Validate() error {
	if e.TestPlanID == "" {
		return metadata.NewValidationError("testplanId is a mandatory parameter")
	}
	if e.ProjectID == "" {
		return metadata.NewValidationError("projectId is a mandatory parameter")
	}
	if e.ScenarioID == "" {
		return metadata.NewValidationError("scenarioId is a mandatory parameter")
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

// IdentityGenerator created and identity to be set on the executor
type IdentityGenerator interface {
	AddMeta(author string, objType string, identifiable metadata.Identifiable) error
}

// New returns a function used to create an executor
func New(ig IdentityGenerator, collection Adder, getProject getItem, getScenario getItem, getTestPlan getItem) func(ctx context.Context, author string, data io.Reader) (interface{}, error) {
	return func(ctx context.Context, author string, data io.Reader) (interface{}, error) {
		executor := &Executor{}
		if err := decoder.Decode(executor, data); err != nil {
			return nil, err
		}
		if _, err := getProject(ctx, executor.ProjectID); err != nil {
			if store.IsNotFoundError(err) {
				return nil, metadata.NewValidationError("project with the provided ID does not exist")
			}
			return nil, err
		}
		if _, err := getScenario(ctx, executor.ScenarioID); err != nil {
			if store.IsNotFoundError(err) {
				return nil, metadata.NewValidationError("scenario with the provided ID does not exist")
			}
			return nil, err
		}
		if _, err := getTestPlan(ctx, executor.TestPlanID); err != nil {
			if store.IsNotFoundError(err) {
				return nil, metadata.NewValidationError("test plan with the provided ID does not exist")
			}
			return nil, err
		}
		if err := ig.AddMeta(author, "executor", executor); err != nil {
			return nil, err
		}

		if err := collection.AddOne(ctx, executor); err != nil {
			return nil, err
		}

		return executor, nil
	}
}

// List returns a function used to return the executors
func List(collection Getter) func(ctx context.Context) ([]interface{}, error) {
	return func(ctx context.Context) ([]interface{}, error) {
		executors := []Executor{}
		err := collection.GetAll(ctx, &executors)
		if err != nil {
			return nil, err
		}
		items := make([]interface{}, len(executors))
		fmt.Println(len(items))
		for i, v := range executors {
			items[i] = v
		}
		return items, nil
	}
}

// Get returns a function to retrieve a executor based on the passed ID
func Get(collectiom Getter) func(ctx context.Context, id string) (interface{}, error) {
	return func(ctx context.Context, id string) (interface{}, error) {
		executor := &Executor{}
		if err := collectiom.Get(ctx, id, executor); err != nil {
			return nil, err
		}
		return executor, nil
	}
}
