package opcua

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
)

// ValidationError represents a structured validation error with field context
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("validation failed for field %q (value: %v): %s", e.Field, e.Value, e.Message)
	}
	return fmt.Sprintf("validation failed for field %q: %s", e.Field, e.Message)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// MultiValidationError represents multiple validation errors
type MultiValidationError struct {
	Errors []error
}

func (e *MultiValidationError) Error() string {
	if len(e.Errors) == 0 {
		return "no validation errors"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("validation failed with %d errors:", len(e.Errors)))
	for i, err := range e.Errors {
		sb.WriteString(fmt.Sprintf("\n  %d. %s", i+1, err.Error()))
	}
	return sb.String()
}

func (e *MultiValidationError) Add(err error) {
	if err != nil {
		e.Errors = append(e.Errors, err)
	}
}

func (e *MultiValidationError) HasErrors() bool {
	return len(e.Errors) > 0
}

// Unwrap returns the errors for use with errors.Is and errors.As
func (e *MultiValidationError) Unwrap() []error {
	return e.Errors
}

// ErrorOrNil returns nil if there are no errors, otherwise returns the error collection
func (e *MultiValidationError) ErrorOrNil() error {
	if e.HasErrors() {
		return e
	}
	return nil
}

// ValidateEndpoint validates an OPC UA endpoint URL
func ValidateEndpoint(endpoint string) error {
	if endpoint == "" {
		return &ValidationError{
			Field:   "endpoint",
			Value:   endpoint,
			Message: "URL cannot be empty",
			Err:     ErrInvalidEndpoint,
		}
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return &ValidationError{
			Field:   "endpoint",
			Value:   endpoint,
			Message: fmt.Sprintf("invalid URL format: %v", err),
			Err:     ErrInvalidEndpoint,
		}
	}

	if u.Scheme != "opc.tcp" {
		return &ValidationError{
			Field:   "endpoint",
			Value:   endpoint,
			Message: fmt.Sprintf("invalid scheme %q, expected opc.tcp", u.Scheme),
			Err:     ErrInvalidEndpoint,
		}
	}

	return nil
}

// ValidateChoice validates that a value is one of the allowed choices
func ValidateChoice(fieldName, value string, validChoices []string) error {
	if !slices.Contains(validChoices, value) {
		return &ValidationError{
			Field:   fieldName,
			Value:   value,
			Message: fmt.Sprintf("unknown choice %q, expected one of: %v", value, validChoices),
			Err:     ErrInvalidConfiguration,
		}
	}
	return nil
}

// ValidateOptionalFields validates the optional fields configuration
func ValidateOptionalFields(fields []string) error {
	validFields := []string{"DataType"}
	for i, field := range fields {
		if !slices.Contains(validFields, field) {
			return &ValidationError{
				Field:   fmt.Sprintf("optional_fields[%d]", i),
				Value:   field,
				Message: fmt.Sprintf("unknown field %q, expected one of: %v", field, validFields),
				Err:     ErrInvalidConfiguration,
			}
		}
	}
	return nil
}
