package testplans

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/curious-kitten/scratch-post/internal/decoder"
	"github.com/curious-kitten/scratch-post/internal/store"
	"github.com/curious-kitten/scratch-post/pkg/metadata"
)

//go:generate mockgen -source ./testplan.go -destination mocks/testplan.go

type projectRetriever func(ctx context.Context, id string) (interface{}, error)

// TestPlan is used to define a test case
type TestPlan struct {
	Identity    *metadata.Identity `json:"identity,omitempty"`
	ProjectID   string             `json:"projectId,omitempty"`
	Name        string             `json:"name,omitempty"`
	Description string             `json:"description,omitempty"`
}

// AddIdentity sets the identity of the project
func (s *TestPlan) AddIdentity(identity *metadata.Identity) {
	s.Identity = identity
}

// GetIdentity retruns the identity of the project
func (s *TestPlan) GetIdentity() *metadata.Identity {
	return s.Identity
}

// Validate checks the integrity of the TestPlan 
func (s *TestPlan) Validate() error {
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

// IdentityGenerator created and identity to be set on the testplan
type IdentityGenerator interface {
	AddMeta(author string, objType string, identifiable metadata.Identifiable) error
}

// New returns a function used to create a testplan
func New(ig IdentityGenerator, collection Adder, getProject projectRetriever) func(ctx context.Context, author string, data io.Reader) (interface{}, error) {
	return func(ctx context.Context, author string, data io.Reader) (interface{}, error) {
		testplan := &TestPlan{}
		if err := decoder.Decode(testplan, data); err != nil {
			return nil, err
		}
		if _, err := getProject(ctx, testplan.ProjectID); err != nil {
			if store.IsNotFoundError(err) {
				return nil, metadata.NewValidationError("project with the provided ID does not exist")
			}
			return nil, err
		}
		if err := ig.AddMeta(author, "testplan", testplan); err != nil {
			return nil, err
		}
		if err := collection.AddOne(ctx, testplan); err != nil {
			return nil, err
		}
		return testplan, nil
	}
}

// List returns a function used to return the testplans
func List(collection Getter) func(ctx context.Context) ([]interface{}, error) {
	return func(ctx context.Context) ([]interface{}, error) {
		testplans := []TestPlan{}
		err := collection.GetAll(ctx, &testplans)
		if err != nil {
			return nil, err
		}
		items := make([]interface{}, len(testplans))
		for i, v := range testplans {
			items[i] = v
		}
		return items, nil
	}
}

// Get returns a function to retrieve a testplan based on the passed ID
func Get(collectiom Getter) func(ctx context.Context, id string) (interface{}, error) {
	return func(ctx context.Context, id string) (interface{}, error) {
		testplan := &TestPlan{}
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
func Update(collection ReaderUpdater, getProject projectRetriever) func(ctx context.Context, user string, id string, data io.Reader) (interface{}, error) {
	return func(ctx context.Context, user string, id string, data io.Reader) (interface{}, error) {
		testplan := &TestPlan{}
		if err := decoder.Decode(testplan, data); err != nil {
			return nil, err
		}
		if _, err := getProject(ctx, testplan.ProjectID); err != nil {
			if store.IsNotFoundError(err) {
				return nil, metadata.NewValidationError("project with the provided ID does not exist")
			}
			return nil, err
		}
		foundTestplan, err := Get(collection)(ctx, id)
		if err != nil {
			return nil, err
		}
		var s *TestPlan
		var ok bool
		if s, ok = foundTestplan.(*TestPlan); !ok {
			return nil, fmt.Errorf("invalid data structure in DB")
		}
		testplan.Identity = s.Identity
		testplan.Identity.UpdateTime = time.Now()
		testplan.Identity.UpdatedBy = user
		if err := collection.Update(ctx, id, testplan); err != nil {
			return nil, err
		}
		return testplan, nil
	}
}
