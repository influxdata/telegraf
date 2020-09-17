// Copyright 2018-2020 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package uasc

import (
	"crypto/rsa"
	"time"

	"github.com/gopcua/opcua/ua"
)

// Config represents a configuration which UASC client/server has in common.
type Config struct {

	// SecurityPolicyURI is the URI of the Security Policy used to secure the Message.
	// This field is encoded as a UTF-8 string without a null terminator.
	SecurityPolicyURI string

	// Certificate is the X.509 v3 Certificate assigned to the sending application Instance.
	// This is a DER encoded blob.
	// The structure of an X.509 v3 Certificate is defined in X.509 v3.
	// The DER format for a Certificate is defined in X690.
	// This indicates what Private Key was used to sign the MessageChunk.
	// The Stack shall close the channel and report an error to the application if
	// the Certificate is too large for the buffer size supported by the
	// transport layer.
	// This field shall be null if the Message is not signed.
	Certificate []byte

	// LocalKey is a RSA Private Key which will be used to encrypt the OpenSecureChannel
	// messages.  It is the key associated with Certificate
	LocalKey *rsa.PrivateKey

	// Thumbprint is the thumbprint of the X.509 v3 Certificate assigned to the receiving
	// application Instance.
	// The thumbprint is the CertificateDigest of the DER encoded form of the
	// Certificate.
	// This indicates what public key was used to encrypt the MessageChunk.
	// This field shall be null if the Message is not encrypted.
	Thumbprint []byte

	// RemoteCertificate is the X.509 Certificate for the receiving instance.
	// Used to encrypt the message chunks in the OpenSecureChannel phase.
	RemoteCertificate []byte

	// RequestIDSeed is the initial value for RequestID counter in each new SecureChannel
	RequestIDSeed uint32

	// SecurityMode is The type of security to apply to the messages. The type MessageSecurityMode
	// is defined in 7.15.
	// A SecureChannel may have to be created even if the securityMode is NONE. The exact behaviour
	// depends on the mapping used and is described in the Part 6.
	SecurityMode ua.MessageSecurityMode

	// Lifetime is the requested lifetime, in milliseconds, for the new SecurityToken when the
	// SecureChannel works as client. It specifies when the Client expects to renew the SecureChannel
	// by calling the OpenSecureChannel Service again. If a SecureChannel is not renewed, then all
	// Messages sent using the current SecurityTokens shall be rejected by the receiver.
	// Lifetime can also be the revised lifetime, the lifetime of the SecurityToken in milliseconds.
	// The UTC expiration time for the token may be calculated by adding the lifetime to the createdAt time.
	Lifetime uint32

	// RequestTimeout is timeout duration for all synchronous requests over SecureChannel.
	// If the Server doesn't respond within RequestTimeout time, Client returns StatusBadTimeout
	RequestTimeout time.Duration
}

// SessionConfig is a set of common configurations used in Session.
type SessionConfig struct {
	// AuthenticationToken is the secret Session identifier used to verify that the request is
	// associated with the Session. The SessionAuthenticationToken type is defined in 7.31.
	AuthenticationToken *ua.NodeID

	// ClientDescription is the information that describes the Client application.
	// The type ApplicationDescription is defined in 7.1.
	ClientDescription *ua.ApplicationDescription

	// ServerEndpoints is the list of Endpoints that the Server supports.
	// The Server shall return a set of EndpointDescriptions available for the serverUri
	// specified in the request. The EndpointDescription type is defined in 7.10. The Client
	// shall verify this list with the list from a DiscoveryEndpoint if it used a
	// DiscoveryEndpoint to fetch the EndpointDescriptions.
	// It is recommended that Servers only include the server.applicationUri, endpointUrl,
	// securityMode, securityPolicyUri, userIdentityTokens, transportProfileUri and
	// securityLevel with all other parameters set to null. Only the recommended
	// parameters shall be verified by the client.
	ServerEndpoints []*ua.EndpointDescription

	// LocaleIDs is the list of locale ids in priority order for localized strings. The first
	// LocaleId in the list has the highest priority. If the Server returns a localized string
	// to the Client, the Server shall return the translation with the highest priority that
	// it can. If it does not have a translation for any of the locales identified in this list,
	// then it shall return the string value that it has and include the locale id with the
	// string. See Part 3 for more detail on locale ids. If the Client fails to specify at least
	// one locale id, the Server shall use any that it has.
	// This parameter only needs to be specified during the first call to ActivateSession during
	// a single application Session. If it is not specified the Server shall keep using the
	// current localeIds for the Session.
	LocaleIDs []string

	// UserIdentityToken is the credentials of the user associated with the Client application.
	// The Server uses these credentials to determine whether the Client should be allowed to
	// activate a Session and what resources the Client has access to during this Session.
	// The UserIdentityToken is an extensible parameter type defined in 7.36.
	// The EndpointDescription specifies what UserIdentityTokens the Server shall accept.
	// Null or empty user token shall always be interpreted as anonymous.
	UserIdentityToken interface{}

	// If the Client specified a user identity token that supports digital signatures, then it
	// shall create a signature and pass it as this parameter. Otherwise the parameter is null.
	// The SignatureAlgorithm depends on the identity token type.
	// The SignatureData type is defined in 7.32.
	UserTokenSignature *ua.SignatureData

	// SessionName is an optional name of the session.
	// The default is a unique value for every new session.
	SessionName string

	// If Session works as a client, SessionTimeout is the requested maximum number of milliseconds
	// that a Session should remain open without activity. If the Client fails to issue a Service
	// request within this interval, then the Server shall automatically terminate the Client Session.
	// If Session works as a server, SessionTimeout is an actual maximum number of milliseconds
	// that a Session shall remain open without activity. The Server should attempt to honour the
	// Client request for this parameter,but may negotiate this value up or down to meet its own constraints.
	SessionTimeout time.Duration

	// Stored version of the password to authenticate against a server
	// todo: storing passwords in memory seems wrong
	AuthPassword string

	// PolicyURI to use when encrypting secrets for the User Identity Token
	// Could be different from the secure channel's policy
	AuthPolicyURI string
}
