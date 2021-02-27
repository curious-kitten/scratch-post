package testplan

import "github.com/curious-kitten/scratch-post/pkg/errors"

// Validate checks the integrity of the TestPlan
func (s *TestPlan) Validate() error {
	if s.Name == "" {
		return errors.NewValidationError("name is a mandatory parameter")
	}
	if s.ProjectId == "" {
		return errors.NewValidationError("projectId is a mandatory parameter")
	}
	return nil
}
