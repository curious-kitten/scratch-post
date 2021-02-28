package executions

import (
	"context"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"

	"github.com/curious-kitten/scratch-post/internal/decoder"
	"github.com/curious-kitten/scratch-post/internal/store"
	executionv1 "github.com/curious-kitten/scratch-post/pkg/api/v1/execution"
	metadatav1 "github.com/curious-kitten/scratch-post/pkg/api/v1/metadata"
	scenariov1 "github.com/curious-kitten/scratch-post/pkg/api/v1/scenario"
	"github.com/curious-kitten/scratch-post/pkg/errors"
)

//go:generate mockgen -source ./executions.go -destination mocks/executions.go

type getItem func(ctx context.Context, id string) (interface{}, error)

// Adder is used to add items to the store
type Adder interface {
	AddOne(ctx context.Context, item interface{}) error
}

// Getter is used to retrieve items from the store
type Getter interface {
	Get(ctx context.Context, id string, item interface{}) error
	GetAll(ctx context.Context, items interface{}) error
}

// MetaHandler handles metadata information
type MetaHandler interface {
	NewMeta(author string, objType string) (*metadatav1.Identity, error)
	UpdateMeta(author string, identity *metadatav1.Identity)
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
func New(meta MetaHandler, collection Adder, getProject getItem, getScenario getItem, getTestPlan getItem) func(ctx context.Context, author string, data io.Reader) (interface{}, error) {
	return func(ctx context.Context, author string, data io.Reader) (interface{}, error) {
		execution := &executionv1.Execution{}
		if err := decoder.Decode(execution, data); err != nil {
			return nil, err
		}
		if _, err := getProject(ctx, execution.ProjectId); err != nil {
			if store.IsNotFoundError(err) {
				return nil, errors.NewValidationError("project with the provided ID does not exist")
			}
			return nil, err
		}
		_, err := getTestPlan(ctx, execution.TestPlanId)
		if err != nil {
			if store.IsNotFoundError(err) {
				return nil, errors.NewValidationError("scenario with the provided ID does not exist")
			}
			return nil, err
		}
		raw, err := getScenario(ctx, execution.ScenarioId)
		if err != nil {
			if store.IsNotFoundError(err) {
				return nil, errors.NewValidationError("test plan with the provided ID does not exist")
			}
			return nil, err
		}

		scenario, ok := raw.(*scenariov1.Scenario)
		if !ok {
			return nil, fmt.Errorf("invalid DB entry for scenario %s", execution.ScenarioId)
		}

		identity, err := meta.NewMeta(author, "execution")
		if err != nil {
			return nil, err
		}
		execution.Identity = identity

		execution.PopulateSteps(scenario.Steps)
		execution.Status = executionv1.Status_Pending
		fmt.Println(execution.Identity)
		if err := collection.AddOne(ctx, execution); err != nil {
			return nil, err
		}

		return execution, nil
	}
}

// List returns a function used to return the executions
func List(collection Getter) func(ctx context.Context) ([]interface{}, error) {
	return func(ctx context.Context) ([]interface{}, error) {
		executions := []executionv1.Execution{}
		err := collection.GetAll(ctx, &executions)
		if err != nil {
			return nil, err
		}
		items := make([]interface{}, len(executions))
		fmt.Println(len(items))
		for i := range executions {
			items[i] = proto.Clone(&executions[i]).(*executionv1.Execution)
		}
		return items, nil
	}
}

// Get returns a function to retrieve a execution based on the passed ID
func Get(collectiom Getter) func(ctx context.Context, id string) (interface{}, error) {
	return func(ctx context.Context, id string) (interface{}, error) {
		execution := &executionv1.Execution{}
		if err := collectiom.Get(ctx, id, execution); err != nil {
			return nil, err
		}
		return execution, nil
	}
}

// Update is used to replace a scenario with the provided scenario
func Update(meta MetaHandler, collection ReaderUpdater, getProject getItem, getScenario getItem, getTestPlan getItem) func(ctx context.Context, user string, id string, data io.Reader) (interface{}, error) {
	return func(ctx context.Context, user string, id string, data io.Reader) (interface{}, error) {
		execution := &executionv1.Execution{}
		if err := decoder.Decode(execution, data); err != nil {
			return nil, err
		}
		if _, err := getProject(ctx, execution.ProjectId); err != nil {
			if store.IsNotFoundError(err) {
				return nil, errors.NewValidationError("project with the provided ID does not exist")
			}
			return nil, err
		}
		if _, err := getScenario(ctx, execution.ScenarioId); err != nil {
			if store.IsNotFoundError(err) {
				return nil, errors.NewValidationError("scenario with the provided ID does not exist")
			}
			return nil, err
		}
		if _, err := getTestPlan(ctx, execution.TestPlanId); err != nil {
			if store.IsNotFoundError(err) {
				return nil, errors.NewValidationError("test plan with the provided ID does not exist")
			}
			return nil, err
		}
		rawExecution, err := Get(collection)(ctx, id)
		if err != nil {
			return nil, err
		}

		foundExecution, ok := rawExecution.(*executionv1.Execution)
		if !ok {
			return nil, fmt.Errorf("invalid data sructure in DB")
		}

		meta.UpdateMeta(user, foundExecution.Identity)
		foundExecution.Status = execution.Status

		for _, v := range execution.Steps {
			found := false
			for _, step := range foundExecution.Steps {
				if v.Definition.Name == step.Definition.Name && step.Definition.Position == v.Definition.Position {
					found = true
					step.Status = v.Status
					step.ActualResult = v.ActualResult
					if v.Status == executionv1.Status_Fail {
						foundExecution.Status = executionv1.Status_Fail
					}
				}
			}
			if !found {
				return nil, errors.NewValidationError(fmt.Sprintf("step '%s' is not part of the current scenario", v.Definition.Name))
			}
		}

		if err := collection.Update(ctx, id, foundExecution); err != nil {
			return nil, err
		}
		return foundExecution, nil
	}
}
