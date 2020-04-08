package iotdevice

import (
	"crypto/tls"
)

// ModuleSharedAccessKeyCredentials is a SharedAccessKeyCredentials struct adapted for module connections
type ModuleSharedAccessKeyCredentials struct {
	SharedAccessKeyCredentials        //embedded SharedAccessKeyCredentials struct
	ModuleID                   string // moduleID
	Gateway                    string // name of host gateway
	GenerationID               string // module generation ID
	WorkloadURI                string // IoT Edge workload API URI
	EdgeGateway                bool   // connect via edgeHub
}

// GetModuleID returns ModuleID
func (c *ModuleSharedAccessKeyCredentials) GetModuleID() string {
	return c.ModuleID
}

// GetGenerationID returns GenerationID
func (c *ModuleSharedAccessKeyCredentials) GetGenerationID() string {
	return c.GenerationID
}

// GetGateway returns Gateway Host Name
func (c *ModuleSharedAccessKeyCredentials) GetGateway() string {
	return c.Gateway
}

// UseEdgeGateway returns bool to connect via edgeHub or directly to IoT Hub
func (c *ModuleSharedAccessKeyCredentials) UseEdgeGateway() bool {
	return c.EdgeGateway
}

// GetBroker returns gateway host name if UseEdgeGateway is true, else returns IoT Hub host name
func (c *ModuleSharedAccessKeyCredentials) GetBroker() string {
	output := c.GetHostName()
	gw := c.GetGateway()
	usegw := c.UseEdgeGateway()
	if usegw && len(gw) > 0 {
		output = gw
	}
	return output
}

// GetCertificate returns nil. Only here to satisfy Credentials interface
func (c *ModuleSharedAccessKeyCredentials) GetCertificate() *tls.Certificate {
	return nil
}

// GetSAK returns SharedAccessKey
func (c *ModuleSharedAccessKeyCredentials) GetSAK() string {
	return c.SharedAccessKeyCredentials.SharedAccessKey.SharedAccessKey
}

// GetWorkloadURI returns the URI of the IoT Edge workload API
func (c *ModuleSharedAccessKeyCredentials) GetWorkloadURI() string {
	return c.WorkloadURI
}
