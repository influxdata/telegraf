// Copyright 2013-2017 Aerospike, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aerospike

import (
	"crypto/tls"
	"time"
)

const defaultIdleTimeout = 14 * time.Second

// ClientPolicy encapsulates parameters for client policy command.
type ClientPolicy struct {
	// User authentication to cluster. Leave empty for clusters running without restricted access.
	User string

	// Password authentication to cluster. The password will be stored by the client and sent to server
	// in hashed format. Leave empty for clusters running without restricted access.
	Password string

	// ClusterName sets the expected cluster ID.  If not null, server nodes must return this cluster ID in order to
	// join the client's view of the cluster. Should only be set when connecting to servers that
	// support the "cluster-name" info command. (v3.10+)
	ClusterName string //=""

	// Initial host connection timeout duration.  The timeout when opening a connection
	// to the server host for the first time.
	Timeout time.Duration //= 30 seconds

	// Connection idle timeout. Every time a connection is used, its idle
	// deadline will be extended by this duration. When this deadline is reached,
	// the connection will be closed and discarded from the connection pool.
	IdleTimeout time.Duration //= 14 seconds

	// ConnectionQueueCache specifies the size of the Connection Queue cache PER NODE.
	ConnectionQueueSize int //= 256

	// If set to true, will not create a new connection
	// to the node if there are already `ConnectionQueueSize` active connections.
	LimitConnectionsToQueueSize bool //= true

	// Throw exception if host connection fails during addHost().
	FailIfNotConnected bool //= true

	// TendInterval determines interval for checking for cluster state changes.
	// Minimum possible interval is 10 Milliseconds.
	TendInterval time.Duration //= 1 second

	// A IP translation table is used in cases where different clients
	// use different server IP addresses.  This may be necessary when
	// using clients from both inside and outside a local area
	// network. Default is no translation.
	// The key is the IP address returned from friend info requests to other servers.
	// The value is the real IP address used to connect to the server.
	IpMap map[string]string

	// UseServicesAlternate determines if the client should use "services-alternate" instead of "services"
	// in info request during cluster tending.
	//"services-alternate" returns server configured external IP addresses that client
	// uses to talk to nodes.  "services-alternate" can be used in place of providing a client "ipMap".
	// This feature is recommended instead of using the client-side IpMap above.
	//
	// "services-alternate" is available with Aerospike Server versions >= 3.7.1.
	UseServicesAlternate bool // false

	// RequestProleReplicas determines if prole replicas should be requested from each server node in the cluster tend goroutine.
	// This option is required if there is a need to distribute reads across proles.
	// If RequestProleReplicas is enabled, all prole partition maps will be cached on the client which results in
	// extra storage multiplied by the replication factor.
	// The default is false (only request master replicas and never prole replicas).
	RequestProleReplicas bool // false

	// TlsConfig specifies TLS secure connection policy for TLS enabled servers.
	// For better performance, we suggest prefering the server-side ciphers by
	// setting PreferServerCipherSuites = true.
	TlsConfig *tls.Config //= nil

	// IgnoreOtherSubnetAliases helps to ignore aliases that are outside main subnet
	IgnoreOtherSubnetAliases bool //= false
}

// NewClientPolicy generates a new ClientPolicy with default values.
func NewClientPolicy() *ClientPolicy {
	return &ClientPolicy{
		Timeout:                     30 * time.Second,
		IdleTimeout:                 defaultIdleTimeout,
		ConnectionQueueSize:         256,
		FailIfNotConnected:          true,
		TendInterval:                time.Second,
		LimitConnectionsToQueueSize: true,
		RequestProleReplicas:        false,
		IgnoreOtherSubnetAliases:    false,
	}
}

// RequiresAuthentication returns true if a USer or Password is set for ClientPolicy.
func (cp *ClientPolicy) RequiresAuthentication() bool {
	return (cp.User != "") || (cp.Password != "")
}

func (cp *ClientPolicy) serviceString() string {
	if cp.UseServicesAlternate {
		return "services-alternate"
	}
	return "services"
}
