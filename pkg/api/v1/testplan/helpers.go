package testplan

import "github.com/curious-kitten/scratch-post/internal/decoder"

// Validate checks the integrity of the TestPlan
func (s *TestPlan) Validate() error {
	if s.Name == "" {
		return decoder.NewValidationError("name is a mandatory parameter")
	}
	if s.ProjectId == "" {
		return decoder.NewValidationError("projectId is a mandatory parameter")
	}
	return nil
}
