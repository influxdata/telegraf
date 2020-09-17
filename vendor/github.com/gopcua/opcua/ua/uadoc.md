# FindServersOnNetworkResponse

// FindServersOnNetworkResponse returns the Servers known to a Server or Discovery Server. The behaviour of
// Discovery Servers is described in detail in Part 12.
//
// Specification: Part4, 5.4.2

# UserIdentityToken

// UserIdentityToken structure used in the Server Service Set allows Clients to specify the
// identity of the user they are acting on behalf of. The exact mechanism used to identify users
// depends on the system configuration.
//
// Specification: Part 4, 7.36.1

# WriteRequest

// WriteRequest is used to write values to one or more Attributes of one or more Nodes. For
// constructed Attribute values whose elements are indexed, such as an array, this Service allows
// Clients to write the entire set of indexed values as a composite, to write individual elements or to
// write ranges of elements of the composite.
//
// Specification: Part 4, 5.10.4

# CancelRequest

// CancelRequest is used to cancel outstanding Service requests. Successfully cancelled service
// requests shall respond with Bad_RequestCancelledByClient.
//
// Specification: Part4, 5.6.5

# ChannelSecurityToken

// ChannelSecurityToken represents a ChannelSecurityToken.
// It describes the new SecurityToken issued by the Server.
//
// Specification: Part 4, 5.5.2.2

# ReadRequest

// ReadRequest is used to read one or more Attributes of one or more Nodes.
// For constructed Attribute values whose elements are indexed, such as an array,
// this Service allows Clients to read the entire set of indexed values as a composite,
// to read individual elements or to read ranges of elements of the composite.
//
// Specification: Part 4, 5.10.2.2
// type ReadRequest struct {
// 	*RequestHeader

// 	// Maximum age of the value to be read in milliseconds.
// 	// The age of the value is based on the difference between the ServerTimestamp
// 	// and the time when the Server starts processing the request.
// 	// For example if the Client specifies a maxAge of 500 milliseconds
// 	// and it takes 100 milliseconds until the Server starts processing the request,
// 	// the age of the returned value could be 600 milliseconds prior to the time it was requested.
// 	//
// 	// If the Server has one or more values of an Attribute that are within the maximum age,
// 	// it can return any one of the values or it can read a new value from the data source.
// 	// The number of values of an Attribute that a Server has depends on the
// 	// number of MonitoredItems that are defined for the Attribute.
// 	// In any case, the Client can make no assumption about which copy of the data will be returned.
// 	// If the Server does not have a value that is within the maximum age,
// 	// it shall attempt to read a new value from the data source.
// 	//
// 	// If the Server cannot meet the requested maxAge, it returns its “best effort” value
// 	// rather than rejecting the request. This may occur when the time it takes the
// 	// Server to process and return the new data value after it has been accessed is
// 	// greater than the specified maximum age.
// 	//
// 	// If maxAge is set to 0, the Server shall attempt to read a new value from the data source.
// 	//
// 	// If maxAge is set to the max Int32 value or greater, the Server shall attempt to get
// 	// a cached value.
// 	//
// 	// Negative values are invalid for maxAge.
// 	MaxAge uint64

// 	// An enumeration that specifies the Timestamps to be returned for each requested
// 	// Variable Value Attribute.
// 	TimestampsToReturn TimestampsToReturn

// 	// List of Nodes and their Attributes to read. For each entry in this list,
// 	// a StatusCode is returned, and if it indicates success, the Attribute Value is also returned.
// 	NodesToRead []*ReadValueID
// }

# CreateSubscriptionRequest

// CreateSubscriptionRequest is used to create a Subscription. Subscriptions monitor a set of MonitoredItems for
// Notifications and return them to the Client in response to Publish requests.
// Illegal request values for parameters that can be revised do not generate errors. Instead the
// Server will choose default values and indicate them in the corresponding revised parameter.
//
// Specification: Part 4, 5.13.2

# FindServersResponse

// FindServersResponse returns the Servers known to a Server or Discovery Server. The behaviour of
// Discovery Servers is described in detail in Part 12.
//
// Specification: Part4, 5.4.2

# ServersOnNetwork

// ServersOnNetwork is a DNS service record that meet criteria specified in the request.
// This list is empty if no Servers meet the criteria.
//
// Specification: Part4, 5.4.3.2

# GetEndpointsRequest

// GetEndpointsRequest represents an GetEndpointsRequest.
// This Service returns the Endpoints supported by a Server and all of the configuration information
// required to establish a SecureChannel and a Session.
//
// Specification: Part 4, 5.4.4.2

# CancelRequest

// CancelResponse is used to cancel outstanding Service requests. Successfully cancelled service
// requests shall respond with Bad_RequestCancelledByClient.
//
// Specification: Part4, 5.6.5

# CloseSecureChannel

CloseSecureChannelRequest represents an CloseSecureChannelRequest.
This Service is used to terminate a SecureChannel.

Specification: Part 4, 5.5.3.2

# CloseSessionRequest

CloseSessionRequest represents an CloseSessionRequest.
This Service is used to terminate a Session.

Specification: Part 4, 5.6.4.2

# CloseSessionResponse

CloseSessionResponse represents an CloseSessionResponse.
This Service is used to terminate a Session.

Specification: Part 4, 5.6.4.2


# FindServersOnNetworkRequest

FindServersOnNetworkRequest returns the Servers known to a Discovery Server. Unlike FindServers, this Service is
only implemented by Discovery Servers.

The Client may reduce the number of results returned by specifying filter criteria. An empty list is
returned if no Server matches the criteria specified by the Client.

This Service shall not require message security but it may require transport layer security.

Each time the Discovery Server creates or updates a record in its cache it shall assign a
monotonically increasing identifier to the record. This allows Clients to request records in batches
by specifying the identifier for the last record received in the last call to FindServersOnNetwork.
To support this the Discovery Server shall return records in numerical order starting from the
lowest record identifier. The Discovery Server shall also return the last time the counter was reset
for example due to a restart of the Discovery Server. If a Client detects that this time is more
recent than the last time the Client called the Service it shall call the Service again with a
startingRecordId of 0.

This Service can be used without security and it is therefore vulnerable to denial of service (DOS)
attacks. A Server should minimize the amount of processing required to send the response for this
Service. This can be achieved by preparing the result in advance.

Specification: Part 4, 5.4.3

# AnonymousIdentityToken

AnonymousIdentityToken is used to indicate that the Client has no user credentials.

Specification: Part4, 7.36.5

# UserNameIdentityToken

UserNameIdentityToken is used to pass simple username/password credentials to the Server.

This token shall be encrypted by the Client if required by the SecurityPolicy of the
UserTokenPolicy. The Server should specify a SecurityPolicy for the UserTokenPolicy if the
SecureChannel has a SecurityPolicy of None and no transport layer encryption is available. If
None is specified for the UserTokenPolicy and SecurityPolicy is None then the password only
contains the UTF-8 encoded password. The SecurityPolicy of the SecureChannel is used if no
SecurityPolicy is specified in the UserTokenPolicy.

If the token is to be encrypted the password shall be converted to a UTF-8 ByteString, encrypted
and then serialized as shown in Table 181.
The Server shall decrypt the password and verify the ServerNonce.

If the SecurityPolicy is None then the password only contains the UTF-8 encoded password. This
configuration should not be used unless the network is encrypted in some other manner such as a
VPN. The use of this configuration without network encryption would result in a serious security
fault, in that it would cause the appearance of a secure user access, but it would make the
password visible in clear text.

Specification: Part4, 7.36.4

# X509IdentityToken

X509IdentityToken is used to pass an X.509 v3 Certificate which is issued by the user.
This token shall always be accompanied by a Signature in the userTokenSignature parameter of
ActivateSession if required by the SecurityPolicy. The Server should specify a SecurityPolicy for
the UserTokenPolicy if the SecureChannel has a SecurityPolicy of None.

Specification: Part4, 7.36.5

# IssuedIdentityToken

IssuedIdentityToken is used to pass SecurityTokens issued by an external Authorization
Service to the Server. These tokens may be text or binary.
OAuth2 defines a standard for Authorization Services that produce JSON Web Tokens (JWT).
These JWTs are passed as an Issued Token to an OPC UA Server which uses the signature
contained in the JWT to validate the token. Part 6 describes OAuth2 and JWTs in more detail. If
the token is encrypted, it shall use the EncryptedSecret format defined in 7.36.2.3.
This token shall be encrypted by the Client if required by the SecurityPolicy of the
UserTokenPolicy. The Server should specify a SecurityPolicy for the UserTokenPolicy if the
SecureChannel has a SecurityPolicy of None and no transport layer encryption is available. The
SecurityPolicy of the SecureChannel is used If no SecurityPolicy is specified in the
UserTokenPolicy.
If the SecurityPolicy is not None, the tokenData shall be encoded in UTF-8 (if it is not already
binary), signed and encrypted according the rules specified for the tokenType of the associated
UserTokenPolicy (see 7.37).
If the SecurityPolicy is None then the tokenData only contains the UTF-8 encoded tokenData. This
configuration should not be used unless the network is encrypted in some other manner such as a
VPN. The use of this configuration without network encryption would result in a serious security
fault, in that it would cause the appearance of a secure user access, but it would make the token
visible in clear text.

Specification: Part4, 7.36.6

# QualifiedName

// QualifiedName contains a qualified name. It is, for example, used as BrowseName.
// The name part of the QualifiedName is restricted to 512 characters.
//
// Specification: Part 3, 8.3

# WriteResponse

WriteResponse is used to write values to one or more Attributes of one or more Nodes. For
constructed Attribute values whose elements are indexed, such as an array, this Service allows
Clients to write the entire set of indexed values as a composite, to write individual elements or to
write ranges of elements of the composite.

Specification: Part 4, 5.10.4

# ActivateSessionRequest

Specification: Part 4, 5.6.3.2

ActivateSessionRequest is used by the Client to specify the identity of the user
associated with the Session. This Service request shall be issued by the Client
before it issues any Service request other than CloseSession after CreateSession.
Failure to do so shall cause the Server to close the Session.

Whenever the Client calls this Service the Client shall prove that it is the same application that
called the CreateSession Service. The Client does this by creating a signature with the private key
associated with the clientCertificate specified in the CreateSession request. This signature is
created by appending the last serverNonce provided by the Server to the serverCertificate and
calculating the signature of the resulting sequence of bytes.


# ActivateSessionResponse

Specification: Part 4, 5.6.3.2

ActivateSessionResponse is used by the Server to answer to the ActivateSessionRequest.
Once used, a serverNonce cannot be used again. For that reason, the Server returns a new
serverNonce each time the ActivateSession Service is called.

When the ActivateSession Service is called for the first time then the Server shall reject the
request if the SecureChannel is not same as the one associated with the CreateSession request.
Subsequent calls to ActivateSession may be associated with different SecureChannels. If this is
the case then the Server shall verify that the Certificate the Client used to create the new
SecureChannel is the same as the Certificate used to create the original SecureChannel. In
addition, the Server shall verify that the Client supplied a UserIdentityToken that is identical to the
token currently associated with the Session. Once the Server accepts the new SecureChannel it
shall reject requests sent via the old SecureChannel.

# ApplicationType

Specification: Part 4, 7.1

# ApplicationDescription

Specification: Part 4, 7.1

# ReadValueID

ReadValueID is an identifier for an item to read or to monitor.

Specification: Part 4, 7.24

# FindServersRequest

FindServersRequest returns the Servers known to a Server or Discovery Server. The behaviour of
Discovery Servers is described in detail in Part 12.

The Client may reduce the number of results returned by specifying filter criteria. A Discovery
Server returns an empty list if no Servers match the criteria specified by the client. The filter
criteria supported by this Service are described in 5.4.2.2.

Specification: Part 4, 5.4.2
