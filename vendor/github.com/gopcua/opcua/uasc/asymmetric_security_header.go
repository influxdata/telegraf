// Copyright 2018-2020 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package uasc

import (
	"fmt"

	"github.com/gopcua/opcua/ua"
)

// AsymmetricSecurityHeader represents a Asymmetric Algorithm Security Header in OPC UA Secure Conversation.
type AsymmetricSecurityHeader struct {
	SecurityPolicyURI             string
	SenderCertificate             []byte
	ReceiverCertificateThumbprint []byte
}

// NewAsymmetricSecurityHeader creates a new OPC UA Secure Conversation Asymmetric Algorithm Security Header.
func NewAsymmetricSecurityHeader(uri string, cert, thumbprint []byte) *AsymmetricSecurityHeader {
	return &AsymmetricSecurityHeader{
		SecurityPolicyURI:             uri,
		SenderCertificate:             cert,
		ReceiverCertificateThumbprint: thumbprint,
	}
}

func (h *AsymmetricSecurityHeader) Decode(b []byte) (int, error) {
	buf := ua.NewBuffer(b)
	h.SecurityPolicyURI = buf.ReadString()
	h.SenderCertificate = buf.ReadBytes()
	h.ReceiverCertificateThumbprint = buf.ReadBytes()
	return buf.Pos(), buf.Error()
}

func (h *AsymmetricSecurityHeader) Encode() ([]byte, error) {
	buf := ua.NewBuffer(nil)
	buf.WriteString(h.SecurityPolicyURI)
	buf.WriteByteString(h.SenderCertificate)
	buf.WriteByteString(h.ReceiverCertificateThumbprint)
	return buf.Bytes(), buf.Error()
}

// String returns Header in string.
func (a *AsymmetricSecurityHeader) String() string {
	return fmt.Sprintf(
		"SecurityPolicyURI: %v, SenderCertificate: %v, ReceiverCertificateThumbprint: %v",
		a.SecurityPolicyURI,
		a.SenderCertificate,
		a.ReceiverCertificateThumbprint,
	)
}

// Len returns the Header Length in bytes.
func (h *AsymmetricSecurityHeader) Len() int {
	var l int
	l += 12
	l += len(h.SecurityPolicyURI)
	l += len(h.SenderCertificate)
	l += len(h.ReceiverCertificateThumbprint)

	return l
}
