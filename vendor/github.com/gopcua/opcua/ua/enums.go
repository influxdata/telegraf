// Copyright 2018-2020 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package ua

import "strings"

// additional enum values which are not generated.

const (
	NodeClassAll NodeClass = 0xff
)

// Identifiers assigned to Attributes.
//
// Specification: Part 6, A.1
type AttributeID uint32

const (
	AttributeIDInvalid                 AttributeID = 0
	AttributeIDNodeID                  AttributeID = 1
	AttributeIDNodeClass               AttributeID = 2
	AttributeIDBrowseName              AttributeID = 3
	AttributeIDDisplayName             AttributeID = 4
	AttributeIDDescription             AttributeID = 5
	AttributeIDWriteMask               AttributeID = 6
	AttributeIDUserWriteMask           AttributeID = 7
	AttributeIDIsAbstract              AttributeID = 8
	AttributeIDSymmetric               AttributeID = 9
	AttributeIDInverseName             AttributeID = 10
	AttributeIDContainsNoLoops         AttributeID = 11
	AttributeIDEventNotifier           AttributeID = 12
	AttributeIDValue                   AttributeID = 13
	AttributeIDDataType                AttributeID = 14
	AttributeIDValueRank               AttributeID = 15
	AttributeIDArrayDimensions         AttributeID = 16
	AttributeIDAccessLevel             AttributeID = 17
	AttributeIDUserAccessLevel         AttributeID = 18
	AttributeIDMinimumSamplingInterval AttributeID = 19
	AttributeIDHistorizing             AttributeID = 20
	AttributeIDExecutable              AttributeID = 21
	AttributeIDUserExecutable          AttributeID = 22
	AttributeIDDataTypeDefinition      AttributeID = 23
	AttributeIDRolePermissions         AttributeID = 24
	AttributeIDUserRolePermissions     AttributeID = 25
	AttributeIDAccessRestrictions      AttributeID = 26
	AttributeIDAccessLevelEx           AttributeID = 27
)

// Built-in type identifiers.
//
// All OPC UA DataEncodings are based on rules that are defined for a standard
// set of built-in types. These built-in types are then used to construct
// structures, arrays and messages.
//
// Specification: Part 6, 5.1.2
type TypeID byte

const (
	TypeIDNull            TypeID = 0 // not part of specification but some servers (e.g. Prosys) return it anyway
	TypeIDBoolean         TypeID = 1
	TypeIDSByte           TypeID = 2
	TypeIDByte            TypeID = 3
	TypeIDInt16           TypeID = 4
	TypeIDUint16          TypeID = 5
	TypeIDInt32           TypeID = 6
	TypeIDUint32          TypeID = 7
	TypeIDInt64           TypeID = 8
	TypeIDUint64          TypeID = 9
	TypeIDFloat           TypeID = 10
	TypeIDDouble          TypeID = 11
	TypeIDString          TypeID = 12
	TypeIDDateTime        TypeID = 13
	TypeIDGUID            TypeID = 14
	TypeIDByteString      TypeID = 15
	TypeIDXMLElement      TypeID = 16
	TypeIDNodeID          TypeID = 17
	TypeIDExpandedNodeID  TypeID = 18
	TypeIDStatusCode      TypeID = 19
	TypeIDQualifiedName   TypeID = 20
	TypeIDLocalizedText   TypeID = 21
	TypeIDExtensionObject TypeID = 22
	TypeIDDataValue       TypeID = 23
	TypeIDVariant         TypeID = 24
	TypeIDDiagnosticInfo  TypeID = 25
)

// SecurityPolicyURI is a listing of UA security policy URIs
// Specification: Part 7, 6.6.161-166

const (
	SecurityPolicyURIPrefix              = "http://opcfoundation.org/UA/SecurityPolicy#"
	SecurityPolicyURINone                = "http://opcfoundation.org/UA/SecurityPolicy#None"
	SecurityPolicyURIBasic128Rsa15       = "http://opcfoundation.org/UA/SecurityPolicy#Basic128Rsa15"
	SecurityPolicyURIBasic256            = "http://opcfoundation.org/UA/SecurityPolicy#Basic256"
	SecurityPolicyURIBasic256Sha256      = "http://opcfoundation.org/UA/SecurityPolicy#Basic256Sha256"
	SecurityPolicyURIAes128Sha256RsaOaep = "http://opcfoundation.org/UA/SecurityPolicy#Aes128_Sha256_RsaOaep"
	SecurityPolicyURIAes256Sha256RsaPss  = "http://opcfoundation.org/UA/SecurityPolicy#Aes256_Sha256_RsaPss"
)

var SecurityPolicyURIs = map[string]string{
	"None":                SecurityPolicyURINone,
	"Basic128Rsa15":       SecurityPolicyURIBasic128Rsa15,
	"Basic256":            SecurityPolicyURIBasic256,
	"Basic256Sha256":      SecurityPolicyURIBasic256Sha256,
	"Aes128Sha256RsaOaep": SecurityPolicyURIAes128Sha256RsaOaep,
	"Aes256Sha256RsaPss":  SecurityPolicyURIAes256Sha256RsaPss,
}

// FormatSecurityPolicy converts a short name for a security policy into a
// canonical policy URI.
func FormatSecurityPolicyURI(policy string) string {
	if policy == "" {
		return ""
	}
	if p, ok := SecurityPolicyURIs[policy]; ok {
		return p
	}
	if !strings.HasPrefix(policy, SecurityPolicyURIPrefix) {
		return SecurityPolicyURIPrefix + policy
	}
	return policy
}
