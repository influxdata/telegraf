package opcua

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEndpointError(t *testing.T) {
	tests := []struct {
		name     string
		err      *EndpointError
		expected string
	}{
		{
			name: "endpoint error with wrapped error",
			err: &EndpointError{
				Endpoint: "opc.tcp://localhost:4840",
				Err:      ErrInvalidEndpoint,
			},
			expected: `endpoint "opc.tcp://localhost:4840": invalid endpoint`,
		},
		{
			name: "empty endpoint",
			err: &EndpointError{
				Endpoint: "",
				Err:      ErrInvalidEndpoint,
			},
			expected: `endpoint "": invalid endpoint`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.err.Error())
			require.ErrorIs(t, tt.err, ErrInvalidEndpoint)
		})
	}
}

func TestSecurityError(t *testing.T) {
	tests := []struct {
		name     string
		err      *SecurityError
		expected string
	}{
		{
			name: "policy and mode error",
			err: &SecurityError{
				Policy: "Basic256",
				Mode:   "SignAndEncrypt",
				Err:    ErrInvalidSecurityPolicy,
			},
			expected: `security policy "Basic256", mode "SignAndEncrypt": invalid security policy`,
		},
		{
			name: "policy only error",
			err: &SecurityError{
				Policy: "Invalid",
				Err:    ErrInvalidSecurityPolicy,
			},
			expected: `security policy "Invalid": invalid security policy`,
		},
		{
			name: "mode only error",
			err: &SecurityError{
				Mode: "Invalid",
				Err:  ErrInvalidSecurityMode,
			},
			expected: `security mode "Invalid": invalid security mode`,
		},
		{
			name: "generic security error",
			err: &SecurityError{
				Err: ErrInvalidConfiguration,
			},
			expected: `security configuration: invalid configuration`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.err.Error())
			require.ErrorIs(t, tt.err, tt.err.Err)
		})
	}
}

func TestAuthenticationError(t *testing.T) {
	tests := []struct {
		name     string
		err      *AuthenticationError
		expected string
	}{
		{
			name: "username authentication error",
			err: &AuthenticationError{
				Method: "username",
				Err:    errors.New("missing username"),
			},
			expected: `authentication method "username": missing username`,
		},
		{
			name: "certificate authentication error",
			err: &AuthenticationError{
				Method: "certificate",
				Err:    errors.New("invalid certificate"),
			},
			expected: `authentication method "certificate": invalid certificate`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestCertificateError(t *testing.T) {
	tests := []struct {
		name     string
		err      *CertificateError
		expected string
	}{
		{
			name: "certificate generation with path",
			err: &CertificateError{
				Operation: "generation",
				Path:      "/tmp/cert.pem",
				Err:       ErrCertificateGeneration,
			},
			expected: `certificate generation at "/tmp/cert.pem": certificate generation failed`,
		},
		{
			name: "certificate operation without path",
			err: &CertificateError{
				Operation: "validation",
				Err:       ErrCertificateGeneration,
			},
			expected: `certificate validation: certificate generation failed`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.err.Error())
			require.ErrorIs(t, tt.err, ErrCertificateGeneration)
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	baseErr := errors.New("base error")

	tests := []struct {
		name    string
		err     error
		target  error
		wrapped error
	}{
		{
			name: "endpoint error wraps correctly",
			err: &EndpointError{
				Endpoint: "test",
				Err:      baseErr,
			},
			target:  baseErr,
			wrapped: baseErr,
		},
		{
			name: "security error wraps correctly",
			err: &SecurityError{
				Policy: "test",
				Err:    baseErr,
			},
			target:  baseErr,
			wrapped: baseErr,
		},
		{
			name: "authentication error wraps correctly",
			err: &AuthenticationError{
				Method: "test",
				Err:    baseErr,
			},
			target:  baseErr,
			wrapped: baseErr,
		},
		{
			name: "certificate error wraps correctly",
			err: &CertificateError{
				Operation: "test",
				Err:       baseErr,
			},
			target:  baseErr,
			wrapped: baseErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.ErrorIs(t, tt.err, tt.target)
			require.Equal(t, tt.wrapped, errors.Unwrap(tt.err))
		})
	}
}
