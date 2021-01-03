package projects

import (
	"context"
	"encoding/json"
	"io"

	"github.com/curious-kitten/scratch-post/pkg/metadata"
)

// Project represents a umbrella for tests that refer to the same product
type Project struct {
	Identity    *metadata.Identity `json:"identity"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
}

// AddIdentity sets the identity of the project
func (p *Project) AddIdentity(identity *metadata.Identity) {
	p.Identity = identity
}

// GetIdentity retruns the identity of the project
func (p *Project) GetIdentity() *metadata.Identity {
	return p.Identity
}

type Adder interface {
	AddOne(ctx context.Context, item interface{}) error
}

// Getter is used to retrieve items from the store
type Getter interface {
	Get(ctx context.Context, id string, item interface{}) error
	GetAll(ctx context.Context, items interface{}) error
}

type identityGenerator interface {
	AddMeta(author string, objType string, identifiable metadata.Identifiable) error
}

// Creator creates a new project
func Creator(ig identityGenerator, store Adder) func(ctx context.Context, author string, data io.ReadCloser) (interface{}, error) {
	return func(ctx context.Context, author string, data io.ReadCloser) (interface{}, error) {
		project := &Project{}
		err := json.NewDecoder(data).Decode(project)
		if err != nil {
			return nil, err
		}
		err = ig.AddMeta(author, "project", project)
		if err != nil {
			return nil, err
		}
		err = store.AddOne(ctx, project)
		if err != nil {
			return nil, err
		}
		return project, nil
	}
}

// List returns a function used to return the projects
func List(collection Getter) func(ctx context.Context) ([]interface{}, error) {
	return func(ctx context.Context) ([]interface{}, error) {
		projects := []Project{}
		err := collection.GetAll(ctx, &projects)
		if err != nil {
			return nil, err
		}
		items := make([]interface{}, len(projects))
		for i, v := range projects {
			items[i] = v
		}
		return items, nil
	}
}

// Get returns a scenario based on the passed ID
func Get(collectiom Getter) func(ctx context.Context, id string) (interface{}, error) {
	return func(ctx context.Context, id string) (interface{}, error) {
		project := &Project{}
		if err := collectiom.Get(ctx, id, project); err != nil {
			return nil, err
		}
		return project, nil
	}
}
