package decoder

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/curious-kitten/scratch-post/pkg/metadata"
)

// Validatable represents an item that has constraints on what a correct structure is an imposes these constraints through the Validate method
type Validatable interface {
	Validate() error
}

// Decode is used to unmarshall the given data into an object
func Decode(item Validatable, data io.Reader) error {
	decoder := json.NewDecoder(data)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(item)
	if err != nil {
		return metadata.NewValidationError(fmt.Sprintf("invalid scenario body: %s", err.Error()))
	}
	if err = item.Validate(); err != nil {
		return metadata.NewValidationError(err.Error())
	}
	return nil
}
