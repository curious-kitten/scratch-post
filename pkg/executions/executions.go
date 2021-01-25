package executions

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/curious-kitten/scratch-post/internal/decoder"
	"github.com/curious-kitten/scratch-post/internal/store"

	// "github.com/curious-kitten/scratch-post/pkg/executors"
	"github.com/curious-kitten/scratch-post/pkg/metadata"
	"github.com/curious-kitten/scratch-post/pkg/scenarios"
)

type Status string

const (
	Pass    Status = "pass"
	Fail    Status = "fail"
	Pending Status = "pending"
)

//go:generate mockgen -source ./executions.go -destination mocks/executions.go

type getItem func(ctx context.Context, id string) (interface{}, error)

// Execution is used to define a test case
type Execution struct {
	Identity     *metadata.Identity `json:"identity,omitempty"`
	ProjectID    string             `json:"projectId,omitempty"`
	ScenarioID   string             `json:"scenarioId,omitempty"`
	TestPlanID   string             `json:"testplanId,omitempty"`
	Name         string
	Steps        []*Step `json:"steps,omitempty"`
	Status       Status  `json:"status,omitempty"`
	ActualResult string  `json:"actualResult,omitempty"`
}

type Step struct {
	scenarios.Step `json:",inline"`
	Status         Status `json:"status,omitempty"`
	ActualResult   string `json:"actualResult,omitempty"`
}

// AddIdentity sets the identity of the project
func (e *Execution) AddIdentity(identity *metadata.Identity) {
	e.Identity = identity
}

// GetIdentity retruns the identity of the project
func (e *Execution) GetIdentity() *metadata.Identity {
	return e.Identity
}

// Validate is used to check the integrity of the execution object
func (e *Execution) Validate() error {
	if e.TestPlanID == "" {
		return metadata.NewValidationError("testplanID is a mandatory parameter")
	}
	if e.ScenarioID == "" {
		return metadata.NewValidationError("scenarioID is a mandatory parameter")
	}
	if e.ProjectID == "" {
		return metadata.NewValidationError("projectId is a mandatory parameter")
	}
	return nil
}

// PopulateSteps the Execution stepts given scenario steps
func (e *Execution) PopulateSteps(s []scenarios.Step) {
	e.Steps = make([]*Step, len(s))
	for i, v := range s {
		e.Steps[i] = &Step{Step: v, Status: Pending}
	}
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

// IdentityGenerator created and identity to be set on the execution
type IdentityGenerator interface {
	AddMeta(author string, objType string, identifiable metadata.Identifiable) error
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

// New returns a function used to create an execution
func New(ig IdentityGenerator, collection Adder, getProject getItem, getScenario getItem, getTestPlan getItem) func(ctx context.Context, author string, data io.Reader) (interface{}, error) {
	return func(ctx context.Context, author string, data io.Reader) (interface{}, error) {
		execution := &Execution{}
		if err := decoder.Decode(execution, data); err != nil {
			return nil, err
		}
		if _, err := getProject(ctx, execution.ProjectID); err != nil {
			if store.IsNotFoundError(err) {
				return nil, metadata.NewValidationError("project with the provided ID does not exist")
			}
			return nil, err
		}
		_, err := getTestPlan(ctx, execution.TestPlanID)
		if err != nil {
			if store.IsNotFoundError(err) {
				return nil, metadata.NewValidationError("scenario with the provided ID does not exist")
			}
			return nil, err
		}
		raw, err := getScenario(ctx, execution.ScenarioID)
		if err != nil {
			if store.IsNotFoundError(err) {
				return nil, metadata.NewValidationError("test plan with the provided ID does not exist")
			}
			return nil, err
		}

		scenario, ok := raw.(*scenarios.Scenario)
		if !ok {
			return nil, fmt.Errorf("invalid DB entry for scenario %s", execution.ScenarioID)
		}

		if err := ig.AddMeta(author, "execution", execution); err != nil {
			return nil, err
		}

		execution.PopulateSteps(scenario.Steps)
		execution.Status = Pending
		if err := collection.AddOne(ctx, execution); err != nil {
			return nil, err
		}

		return execution, nil
	}
}

// List returns a function used to return the executions
func List(collection Getter) func(ctx context.Context) ([]interface{}, error) {
	return func(ctx context.Context) ([]interface{}, error) {
		executions := []Execution{}
		err := collection.GetAll(ctx, &executions)
		if err != nil {
			return nil, err
		}
		items := make([]interface{}, len(executions))
		fmt.Println(len(items))
		for i, v := range executions {
			items[i] = v
		}
		return items, nil
	}
}

// Get returns a function to retrieve a execution based on the passed ID
func Get(collectiom Getter) func(ctx context.Context, id string) (interface{}, error) {
	return func(ctx context.Context, id string) (interface{}, error) {
		execution := &Execution{}
		if err := collectiom.Get(ctx, id, execution); err != nil {
			return nil, err
		}
		return execution, nil
	}
}

// Update is used to replace a scenario with the provided scenario
func Update(collection ReaderUpdater, getProject getItem, getScenario getItem, getTestPlan getItem) func(ctx context.Context, user string, id string, data io.Reader) (interface{}, error) {
	return func(ctx context.Context, user string, id string, data io.Reader) (interface{}, error) {
		execution := &Execution{}
		if err := decoder.Decode(execution, data); err != nil {
			return nil, err
		}
		if _, err := getProject(ctx, execution.ProjectID); err != nil {
			if store.IsNotFoundError(err) {
				return nil, metadata.NewValidationError("project with the provided ID does not exist")
			}
			return nil, err
		}
		if _, err := getScenario(ctx, execution.ScenarioID); err != nil {
			if store.IsNotFoundError(err) {
				return nil, metadata.NewValidationError("scenario with the provided ID does not exist")
			}
			return nil, err
		}
		if _, err := getTestPlan(ctx, execution.TestPlanID); err != nil {
			if store.IsNotFoundError(err) {
				return nil, metadata.NewValidationError("test plan with the provided ID does not exist")
			}
			return nil, err
		}
		rawExecution, err := Get(collection)(ctx, id)
		if err != nil {
			return nil, err
		}

		foundExecution, ok := rawExecution.(*Execution)
		if !ok {
			return nil, fmt.Errorf("invalid data sructure in DB")
		}

		foundExecution.Identity.UpdateTime = time.Now()
		foundExecution.Identity.UpdatedBy = user
		foundExecution.Status = execution.Status

		for _, v := range execution.Steps {
			for _, step := range foundExecution.Steps {
				if v.Name == step.Name && step.Position == v.Position {
					step.Status = v.Status
					step.ActualResult = v.ActualResult
					if v.Status == Fail {
						foundExecution.Status = Fail
					}
				}
			}
		}

		if err := collection.Update(ctx, id, foundExecution); err != nil {
			return nil, err
		}
		return foundExecution, nil
	}
}
