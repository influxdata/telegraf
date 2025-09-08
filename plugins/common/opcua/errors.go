package opcua

import (
	"errors"
	"fmt"
)

// Common OPC UA error types for better error handling and classification
var (
	// ErrInvalidEndpoint indicates an invalid or malformed endpoint URL
	ErrInvalidEndpoint = errors.New("invalid endpoint")

	// ErrInvalidSecurityPolicy indicates an unsupported security policy
	ErrInvalidSecurityPolicy = errors.New("invalid security policy")

	// ErrInvalidSecurityMode indicates an unsupported security mode
	ErrInvalidSecurityMode = errors.New("invalid security mode")

	// ErrInvalidAuthMethod indicates an unsupported authentication method
	ErrInvalidAuthMethod = errors.New("invalid authentication method")

	// ErrCertificateGeneration indicates a failure in certificate generation
	ErrCertificateGeneration = errors.New("certificate generation failed")

	// ErrConnectionFailed indicates a connection failure
	ErrConnectionFailed = errors.New("connection failed")

	// ErrEndpointNotFound indicates no suitable endpoint was found
	ErrEndpointNotFound = errors.New("no suitable endpoint found")

	// ErrInvalidConfiguration indicates invalid configuration parameters
	ErrInvalidConfiguration = errors.New("invalid configuration")

	// ErrStatusCodeParsing indicates failure to parse status codes
	ErrStatusCodeParsing = errors.New("status code parsing failed")
)

// EndpointError represents an error related to endpoint configuration
type EndpointError struct {
	Endpoint string
	Err      error
}

func (e *EndpointError) Error() string {
	return fmt.Sprintf("endpoint %q: %v", e.Endpoint, e.Err)
}

func (e *EndpointError) Unwrap() error {
	return e.Err
}

// SecurityError represents an error related to security configuration
type SecurityError struct {
	Policy string
	Mode   string
	Err    error
}

func (e *SecurityError) Error() string {
	if e.Policy != "" && e.Mode != "" {
		return fmt.Sprintf("security policy %q, mode %q: %v", e.Policy, e.Mode, e.Err)
	} else if e.Policy != "" {
		return fmt.Sprintf("security policy %q: %v", e.Policy, e.Err)
	} else if e.Mode != "" {
		return fmt.Sprintf("security mode %q: %v", e.Mode, e.Err)
	}
	return fmt.Sprintf("security configuration: %v", e.Err)
}

func (e *SecurityError) Unwrap() error {
	return e.Err
}

// AuthenticationError represents an error related to authentication
type AuthenticationError struct {
	Method string
	Err    error
}

func (e *AuthenticationError) Error() string {
	return fmt.Sprintf("authentication method %q: %v", e.Method, e.Err)
}

func (e *AuthenticationError) Unwrap() error {
	return e.Err
}

// CertificateError represents an error related to certificate operations
type CertificateError struct {
	Operation string
	Path      string
	Err       error
}

func (e *CertificateError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("certificate %s at %q: %v", e.Operation, e.Path, e.Err)
	}
	return fmt.Sprintf("certificate %s: %v", e.Operation, e.Err)
}

func (e *CertificateError) Unwrap() error {
	return e.Err
}
