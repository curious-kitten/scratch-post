package metadata

import (
	"fmt"
	"time"

	"github.com/sony/sonyflake"

	metadatav1 "github.com/curious-kitten/scratch-post/pkg/api/v1/metadata"
)

type severity string

type issueType string

const (
	// Low severity issues should be tagged with this
	Low severity = "LOW"
	// Medium severity issues should be tagged with this
	Medium severity = "MEDIUM"
	// High severity issues should be tagged with this
	High severity = "HIGH"

	// Epic is used when the linked issue type is an Epic
	Epic issueType = "Epic"
	// Story is used when the linked issue is of type Story
	Story issueType = "Story"
	// Defect is used when the linked issue is of type Defect/Bug
	Defect issueType = "Defect"
)

// Identifiable represents an object that has/needs an Identity
type Identifiable interface {
	AddIdentity(identity *metadatav1.Identity)
	GetIdentity() *metadatav1.Identity
}

// NewMetaManager creates a new meta data manager
func NewMetaManager() *MetaManager {
	return &MetaManager{
		generator: sonyflake.NewSonyflake(sonyflake.Settings{}),
	}
}

// MetaManager is used to create and update meta information for
type MetaManager struct {
	generator *sonyflake.Sonyflake
}

// NewMeta creates a new identity
func (p *MetaManager) NewMeta(author string, objType string) (*metadatav1.Identity, error) {
	id, err := p.generator.NextID()
	if err != nil {
		return nil, err
	}
	return &metadatav1.Identity{
		Id:           fmt.Sprintf("%x", id),
		Type:         objType,
		Version:      1,
		CreatedBy:    author,
		UpdatedBy:    author,
		CreationTime: time.Now().Unix(),
		UpdateTime:   time.Now().Unix(),
	}, nil
}

// UpdateMeta updated the meta information
func (p *MetaManager) UpdateMeta(author string, identity *metadatav1.Identity) {
	identity.UpdatedBy = author
	identity.UpdateTime = time.Now().Unix()
}
