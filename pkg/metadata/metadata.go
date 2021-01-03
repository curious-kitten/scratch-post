package metadata

import (
	"fmt"
	"time"

	"github.com/sony/sonyflake"
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

// ValidationError represents an issue with value setting on a struct
type ValidationError struct {
	message string
}

func (v *ValidationError) Error() string {
	return v.message
}

// IsValidationError checkIfAnError is a validation error
func IsValidationError(err error) bool {
	switch err.(type) {
	case *ValidationError:
		return true
	default:
		return false
	}
}

// NewValidationError creates a new ValidationError
func NewValidationError(message string) error {
	return &ValidationError{
		message: message,
	}
}

// LinkedIssue are used to identify what external resorce the currect test item refers to
type LinkedIssue struct {
	Link     string    `json:"link"`
	Severity severity  `json:"severity"`
	Type     issueType `json:"type"`
	State    string    `json:"state"`
}

// Identity represents information to identify the given item
type Identity struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Version      int       `json:"version"`
	CreatedBy    string    `json:"createdBy"`
	UpdatedBy    string    `json:"updatedBy"`
	CreationTime time.Time `json:"creationTime"`
	UpdateTime   time.Time `json:"updateTime"`
}

// Identifiable represents an object that has/needs an Identity
type Identifiable interface {
	AddIdentity(identity *Identity)
	GetIdentity() *Identity
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

// AddMeta adds identity to an Identifiable item
func (p *MetaManager) AddMeta(author string, objType string, identifiable Identifiable) error {
	id, err := p.generator.NextID()
	if err != nil {
		return err
	}
	identifiable.AddIdentity(&Identity{
		ID:           fmt.Sprintf("%x", id),
		Type:         objType,
		Version:      1,
		CreatedBy:    author,
		CreationTime: time.Now(),
	})
	return nil
}

// UpdateMeta updated the meta information of an object
func (p *MetaManager) UpdateMeta(author string, identifiable Identifiable) {
	identity := identifiable.GetIdentity()
	identity.UpdatedBy = author
	identity.UpdateTime = time.Now()
}
