package errors

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
