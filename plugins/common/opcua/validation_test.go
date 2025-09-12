package opcua

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		expected string
	}{
		{
			name: "error with value",
			err: &ValidationError{
				Field:   "security_policy",
				Value:   "InvalidPolicy",
				Message: "unknown choice",
			},
			expected: `validation failed for field "security_policy" (value: InvalidPolicy): unknown choice`,
		},
		{
			name: "error without value",
			err: &ValidationError{
				Field:   "endpoint",
				Message: "cannot be empty",
			},
			expected: `validation failed for field "endpoint": cannot be empty`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestMultiValidationError(t *testing.T) {
	tests := []struct {
		name     string
		errors   []error
		expected string
	}{
		{
			name:     "no errors",
			errors:   nil,
			expected: "no validation errors",
		},
		{
			name: "single error",
			errors: []error{
				&ValidationError{Field: "test", Message: "test error"},
			},
			expected: `validation failed for field "test": test error`,
		},
		{
			name: "multiple errors",
			errors: []error{
				&ValidationError{Field: "field1", Message: "error1"},
				&ValidationError{Field: "field2", Message: "error2"},
			},
			expected: "validation failed with 2 errors:\n" +
				"  1. validation failed for field \"field1\": error1\n" +
				"  2. validation failed for field \"field2\": error2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationErrors := &MultiValidationError{Errors: tt.errors}
			require.Equal(t, tt.expected, validationErrors.Error())
			require.Equal(t, len(tt.errors) > 0, validationErrors.HasErrors())
		})
	}
}

func TestMultiValidationErrorUnwrap(t *testing.T) {
	testErr := ErrInvalidConfiguration
	validationError := &ValidationError{
		Field:   "test_field",
		Message: "test error",
		Err:     testErr,
	}

	validationErrors := &MultiValidationError{}
	validationErrors.Add(validationError)

	// This should work with errors.Is due to our Unwrap implementation
	require.ErrorIs(t, validationErrors, testErr)
}

func TestMultiValidationErrorErrorOrNil(t *testing.T) {
	emptyErrors := &MultiValidationError{}
	require.NoError(t, emptyErrors.ErrorOrNil())

	// Test that ErrorOrNil returns the error collection when there are errors
	nonEmptyErrors := &MultiValidationError{}
	nonEmptyErrors.Add(&ValidationError{Field: "test", Message: "test error"})
	require.Error(t, nonEmptyErrors.ErrorOrNil())
	require.Equal(t, nonEmptyErrors, nonEmptyErrors.ErrorOrNil())
}

func TestValidateEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		wantErr  bool
	}{
		{"valid endpoint", "opc.tcp://localhost:4840", false},
		{"empty endpoint", "", true},
		{"invalid URL", "://invalid", true},
		{"wrong scheme", "http://localhost:4840", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEndpoint(tt.endpoint)
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrInvalidEndpoint)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateChoice(t *testing.T) {
	validChoices := []string{"None", "Sign", "SignAndEncrypt"}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid choice", "None", false},
		{"invalid choice", "Invalid", true},
		{"empty choice", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateChoice("test_field", tt.value, validChoices)
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrInvalidConfiguration)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateOptionalFields(t *testing.T) {
	tests := []struct {
		name    string
		fields  []string
		wantErr bool
	}{
		{"valid field", []string{"DataType"}, false},
		{"empty fields", nil, false},
		{"invalid field", []string{"InvalidField"}, true},
		{"mixed valid/invalid", []string{"DataType", "InvalidField"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOptionalFields(tt.fields)
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrInvalidConfiguration)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
