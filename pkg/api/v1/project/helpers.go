package project

import "github.com/curious-kitten/scratch-post/pkg/errors"

// Validate check whether the constraints on Project have been met
func (p *Project) Validate() error {
	if p.Name == "" {
		return errors.NewValidationError("name is a mandatory parameter")
	}
	return nil
}
