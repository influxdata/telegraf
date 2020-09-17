// Copyright 2018-2019 gopcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package uapolicy

import (
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
)

// Thumbprint returns the thumbprint of a DER-encoded certificate
func Thumbprint(c []byte) []byte {
	thumbprint := sha1.Sum(c)

	return thumbprint[:]
}

// PublicKey returns the RSA PublicKey from a DER-encoded certificate
func PublicKey(c []byte) (*rsa.PublicKey, error) {
	cert, err := x509.ParseCertificate(c)
	if err != nil {
		return nil, err
	}

	return cert.PublicKey.(*rsa.PublicKey), nil
}
