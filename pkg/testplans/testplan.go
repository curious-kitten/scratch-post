package testplans

import (
	"context"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"

	"github.com/curious-kitten/scratch-post/internal/decoder"
	"github.com/curious-kitten/scratch-post/internal/store"
	metadatav1 "github.com/curious-kitten/scratch-post/pkg/api/v1/metadata"
	testplanv1 "github.com/curious-kitten/scratch-post/pkg/api/v1/testplan"
	"github.com/curious-kitten/scratch-post/pkg/errors"
)

//go:generate mockgen -source ./testplan.go -destination mocks/testplan.go

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

// New returns a function used to create a testplan
func New(meta MetaHandler, collection Adder, getProject projectRetriever) func(ctx context.Context, author string, data io.Reader) (interface{}, error) {
	return func(ctx context.Context, author string, data io.Reader) (interface{}, error) {
		testplan := &testplanv1.TestPlan{}
		if err := decoder.Decode(testplan, data); err != nil {
			return nil, err
		}
		if _, err := getProject(ctx, testplan.ProjectId); err != nil {
			if store.IsNotFoundError(err) {
				return nil, errors.NewValidationError("project with the provided ID does not exist")
			}
			return nil, err
		}
		identity, err := meta.NewMeta(author, "testplan")
		if err != nil {
			return nil, err
		}
		testplan.Identity = identity
		if err := collection.AddOne(ctx, testplan); err != nil {
			return nil, err
		}
		return testplan, nil
	}
}

// List returns a function used to return the testplans
func List(collection Getter) func(ctx context.Context, filter map[string][]string, sortBy string, reverse bool, count int, previousLastValue string) ([]interface{}, error) {
	return func(ctx context.Context, filter map[string][]string, sortBy string, reverse bool, count int, previousLastValue string) ([]interface{}, error) {
		testplans := []testplanv1.TestPlan{}
		err := collection.GetAll(ctx, &testplans, filter, sortBy, reverse, count, previousLastValue)
		if err != nil {
			return nil, err
		}
		items := make([]interface{}, len(testplans))
		for i := range testplans {
			items[i] = proto.Clone(&testplans[i]).(*testplanv1.TestPlan)
		}
		return items, nil
	}
}

// Get returns a function to retrieve a testplan based on the passed ID
func Get(collectiom Getter) func(ctx context.Context, id string) (interface{}, error) {
	return func(ctx context.Context, id string) (interface{}, error) {
		testplan := &testplanv1.TestPlan{}
		if err := collectiom.Get(ctx, id, testplan); err != nil {
			return nil, err
		}
		return testplan, nil
	}
}

// Delete returns a function to delete a testplan based on the passed ID
func Delete(collection Deleter) func(ctx context.Context, id string) error {
	return func(ctx context.Context, id string) error {
		if err := collection.Delete(ctx, id); err != nil {
			return err
		}
		return nil
	}
}

// Update is used to replace a testplan with the provided testplan
func Update(meta MetaHandler, collection ReaderUpdater, getProject projectRetriever) func(ctx context.Context, user string, id string, data io.Reader) (interface{}, error) {
	return func(ctx context.Context, user string, id string, data io.Reader) (interface{}, error) {
		testplan := &testplanv1.TestPlan{}
		if err := decoder.Decode(testplan, data); err != nil {
			return nil, err
		}
		if _, err := getProject(ctx, testplan.ProjectId); err != nil {
			if store.IsNotFoundError(err) {
				return nil, errors.NewValidationError("project with the provided ID does not exist")
			}
			return nil, err
		}
		foundTestplan, err := Get(collection)(ctx, id)
		if err != nil {
			return nil, err
		}
		var t *testplanv1.TestPlan
		var ok bool
		if t, ok = foundTestplan.(*testplanv1.TestPlan); !ok {
			return nil, fmt.Errorf("invalid data structure in DB")
		}
		testplan.Identity = t.Identity
		meta.UpdateMeta(user, testplan.Identity)
		if err := collection.Update(ctx, id, testplan); err != nil {
			return nil, err
		}
		return testplan, nil
	}
}
