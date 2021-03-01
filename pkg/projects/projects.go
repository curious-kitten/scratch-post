package projects

import (
	"context"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"

	"github.com/curious-kitten/scratch-post/internal/decoder"
	metadatav1 "github.com/curious-kitten/scratch-post/pkg/api/v1/metadata"
	projectv1 "github.com/curious-kitten/scratch-post/pkg/api/v1/project"
)

//go:generate mockgen -source ./projects.go -destination mocks/projects.go

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

// New creates a new project
func New(meta MetaHandler, store Adder) func(ctx context.Context, author string, data io.Reader) (interface{}, error) {
	return func(ctx context.Context, author string, data io.Reader) (interface{}, error) {
		project := &projectv1.Project{}
		if err := decoder.Decode(project, data); err != nil {
			return nil, err
		}
		identity, err := meta.NewMeta(author, "project")
		if err != nil {
			return nil, err
		}
		project.Identity = identity
		if err := store.AddOne(ctx, project); err != nil {
			return nil, err
		}
		return project, nil
	}
}

// List returns a function used to return the projects
func List(collection Getter) func(ctx context.Context, filter map[string][]string, sortBy string, reverse bool, count int, previousLastValue string) ([]interface{}, error) {
	return func(ctx context.Context, filter map[string][]string, sortBy string, reverse bool, count int, previousLastValue string) ([]interface{}, error) {
		projects := []projectv1.Project{}
		err := collection.GetAll(ctx, &projects, filter, sortBy, reverse, count, previousLastValue)
		if err != nil {
			return nil, err
		}
		items := make([]interface{}, len(projects))
		for i := range projects {
			items[i] = proto.Clone(&projects[i]).(*projectv1.Project)
		}
		return items, nil
	}
}

// Get returns a scenario based on the passed ID
func Get(collectiom Getter) func(ctx context.Context, id string) (interface{}, error) {
	return func(ctx context.Context, id string) (interface{}, error) {
		project := &projectv1.Project{}
		if err := collectiom.Get(ctx, id, project); err != nil {
			return nil, err
		}
		return project, nil
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
func Update(meta MetaHandler, collection ReaderUpdater) func(ctx context.Context, user string, id string, data io.Reader) (interface{}, error) {
	return func(ctx context.Context, user string, id string, data io.Reader) (interface{}, error) {
		project := &projectv1.Project{}
		if err := decoder.Decode(project, data); err != nil {
			return nil, err
		}
		foundProject, err := Get(collection)(ctx, id)
		if err != nil {
			return nil, err
		}
		var p *projectv1.Project
		var ok bool
		if p, ok = foundProject.(*projectv1.Project); !ok {
			return nil, fmt.Errorf("invalid data structure in DB")
		}
		project.Identity = p.Identity
		meta.UpdateMeta(user, project.Identity)
		if err := collection.Update(ctx, id, project); err != nil {
			return nil, err
		}
		return project, nil
	}
}
