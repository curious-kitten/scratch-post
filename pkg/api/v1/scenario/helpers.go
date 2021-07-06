package scenario

import "github.com/curious-kitten/scratch-post/internal/decoder"

// Validate is used to check the integrity of the scenario object
func (s *Scenario) Validate() error {
	if s.Name == "" {
		return decoder.NewValidationError("name is a mandatory parameter")
	}
	if s.ProjectId == "" {
		return decoder.NewValidationError("projectId is a mandatory parameter")
	}
	return nil
}

// Validate is used to check the integrity of a scenario step
func (s *Step) Validate() error {
	if s.Name == "" {
		return decoder.NewValidationError("name is a mandatory parameter for a step")
	}
	return nil
}
