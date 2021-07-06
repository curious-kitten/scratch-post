package project

import "github.com/curious-kitten/scratch-post/internal/decoder"

// Validate check whether the constraints on Project have been met
func (p *Project) Validate() error {
	if p.Name == "" {
		return decoder.NewValidationError("name is a mandatory parameter")
	}
	return nil
}
