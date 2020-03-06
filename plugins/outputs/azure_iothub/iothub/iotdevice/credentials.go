package iotdevice

import (
	"crypto/tls"
	"errors"
	"time"

	"github.com/amenzhinsky/iothub/common"
)

type X509Credentials struct {
	HostName    string
	DeviceID    string
	Certificate *tls.Certificate
}

func (c *X509Credentials) GetDeviceID() string {
	return c.DeviceID
}

func (c *X509Credentials) GetHostName() string {
	return c.HostName
}

func (c *X509Credentials) GetCertificate() *tls.Certificate {
	return c.Certificate
}

func (c *X509Credentials) Token(
	resource string, lifetime time.Duration,
) (*common.SharedAccessSignature, error) {
	return nil, errors.New("cannot generate SAS tokens with x509 credentials")
}

type SharedAccessKeyCredentials struct {
	DeviceID string
	common.SharedAccessKey
}

func (c *SharedAccessKeyCredentials) GetDeviceID() string {
	return c.DeviceID
}

func (c *SharedAccessKeyCredentials) GetHostName() string {
	return c.SharedAccessKey.HostName
}

// NOT IMPLEMENTED

// GetCertificate not implemented for SharedAccessKeyCredentials
func (c *SharedAccessKeyCredentials) GetCertificate() *tls.Certificate {
	return nil
}

// GetModuleID not implemented for SharedAccessKeyCredentials
func (c *SharedAccessKeyCredentials) GetModuleID() string {
	return ""
}

// GetGenerationID not implemented for SharedAccessKeyCredentials
func (c *SharedAccessKeyCredentials) GetGenerationID() string {
	return ""
}

// GetGateway not implemented for SharedAccessKeyCredentials
func (c *SharedAccessKeyCredentials) GetGateway() string {
	return ""
}

// GetBroker not implemented for SharedAccessKeyCredentials
func (c *SharedAccessKeyCredentials) GetBroker() string {
	return ""
}

// GetWorkloadURI not implemented for SharedAccessKeyCredentials
func (c *SharedAccessKeyCredentials) GetWorkloadURI() string {
	return ""
}

// UseEdgeGateway not implemented for SharedAccessKeyCredentials
func (c *SharedAccessKeyCredentials) UseEdgeGateway() bool {
	return false
}

// GetModuleID not implemented for X509Credentials
func (c *X509Credentials) GetModuleID() string {
	return ""
}

// GetGenerationID not implemented for X509Credentials
func (c *X509Credentials) GetGenerationID() string {
	return ""
}

// GetGateway not implemented for X509Credentials
func (c *X509Credentials) GetGateway() string {
	return ""
}

// GetBroker not implemented for X509Credentials
func (c *X509Credentials) GetBroker() string {
	return ""
}

// GetWorkloadURI not implemented for X509Credentials
func (c *X509Credentials) GetWorkloadURI() string {
	return ""
}

// UseEdgeGateway not implemented for X509Credentials
func (c *X509Credentials) UseEdgeGateway() bool {
	return false
}
