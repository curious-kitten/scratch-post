package steps

import (
	"github.com/curious-kitten/scratch-post/pkg/definitions"
)

// Step is one of the actions that need to be performed in order to verify a behavior
type Step struct {
	Identity definitions.Identity `json:"identity"`
}
