package scenarios

import (
	"context"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"

	"github.com/curious-kitten/scratch-post/internal/decoder"
	metadatav1 "github.com/curious-kitten/scratch-post/pkg/api/v1/metadata"
	scenariov1 "github.com/curious-kitten/scratch-post/pkg/api/v1/scenario"
)

//go:generate mockgen -source ./scenarios.go -destination mocks/scenarios.go

type projectRetriever func(ctx context.Context, id string) (interface{}, error)

// MetaHandler handles metadata information
type MetaHandler interface {
	NewMeta(author string, objType string) (*metadatav1.Identity, error)
	UpdateMeta(author string, identity *metadatav1.Identity)
}

// Adder is used to add items to the store
type Adder interface {
	AddOne(ctx context.Context, item interface{}) error
}

// Getter is used to retrieve items from the store
type Getter interface {
	Get(ctx context.Context, id string, item interface{}) error
	GetAll(ctx context.Context, items interface{}, filterMap map[string][]string, sortBy string, reverse bool, count int, previousLastValue string) error
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

// New returns a function used to create a scenario
func New(meta MetaHandler, collection Adder, getProject projectRetriever) func(ctx context.Context, author string, data io.Reader) (interface{}, error) {
	return func(ctx context.Context, author string, data io.Reader) (interface{}, error) {
		scenario := &scenariov1.Scenario{}
		if err := decoder.Decode(scenario, data); err != nil {
			return nil, err
		}
		if _, err := getProject(ctx, scenario.ProjectId); err != nil {
			return nil, err
		}
		identity, err := meta.NewMeta(author, "scenario")
		if err != nil {
			return nil, err
		}
		scenario.Identity = identity

		if err := collection.AddOne(ctx, scenario); err != nil {
			return nil, err
		}

		return scenario, nil
	}
}

// List returns a function used to return the scenarios
func List(collection Getter) func(ctx context.Context, filter map[string][]string, sortBy string, reverse bool, count int, previousLastValue string) ([]interface{}, error) {
	return func(ctx context.Context, filter map[string][]string, sortBy string, reverse bool, count int, previousLastValue string) ([]interface{}, error) {
		scenarioList := []scenariov1.Scenario{}
		err := collection.GetAll(ctx, &scenarioList, filter, sortBy, reverse, count, previousLastValue)
		if err != nil {
			return nil, err
		}
		items := make([]interface{}, len(scenarioList))
		for i := range scenarioList {
			items[i] = proto.Clone(&scenarioList[i]).(*scenariov1.Scenario)
		}
		return items, nil
	}
}

// Get returns a function to retrieve a scenario based on the passed ID
func Get(collectiom Getter) func(ctx context.Context, id string) (interface{}, error) {
	return func(ctx context.Context, id string) (interface{}, error) {
		scenario := &scenariov1.Scenario{}
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
func Update(meta MetaHandler, collection ReaderUpdater, getProject projectRetriever) func(ctx context.Context, user string, id string, data io.Reader) (interface{}, error) {
	return func(ctx context.Context, user string, id string, data io.Reader) (interface{}, error) {
		scenario := &scenariov1.Scenario{}
		if err := decoder.Decode(scenario, data); err != nil {
			return nil, err
		}
		if _, err := getProject(ctx, scenario.ProjectId); err != nil {
			return nil, err
		}
		foundScenario, err := Get(collection)(ctx, id)
		if err != nil {
			return nil, err
		}
		var s *scenariov1.Scenario
		var ok bool
		if s, ok = foundScenario.(*scenariov1.Scenario); !ok {
			return nil, fmt.Errorf("invalid data structure in DB")
		}
		scenario.Identity = s.Identity
		meta.UpdateMeta(user, scenario.Identity)
		if err := collection.Update(ctx, id, scenario); err != nil {
			return nil, err
		}
		return scenario, nil
	}
}

// func mangeFilters(filters map[string][]string) {

// }
