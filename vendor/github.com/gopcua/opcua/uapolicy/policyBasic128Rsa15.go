// Copyright 2018-2020 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package uapolicy

import (
	"crypto"
	"crypto/rsa"
	"fmt"

	// Force compilation of required hashing algorithms, although we don't directly use the packages
	_ "crypto/sha1"
	_ "crypto/sha256"

	"github.com/gopcua/opcua/errors"
)

/*
OLD SecurityPolicy – Basic128Rsa15" Profile (DEPRECATED IN 1.04)
http://opcfoundation.org/UA/SecurityPolicy#Basic128Rsa15

Name 	Opt. 	 Description 	 From Profile
	Security Certificate Validation 		A certificate will be validated as specified in Part 4. This includes among others structure and signature examination. Allowing for some validation errors to be suppressed by administration directive.
	Security Basic 128Rsa15 		A suite of algorithms that uses RSA15 as Key-Wrap-algorithm and 128-Bit for encryption algorithms.
-> SymmetricSignatureAlgorithm – HmacSha1 – (http://www.w3.org/2000/09/xmldsig#hmac-sha1).
-> SymmetricEncryptionAlgorithm – Aes128 – (http://www.w3.org/2001/04/xmlenc#aes128-cbc).
-> AsymmetricSignatureAlgorithm – RsaSha1 – (http://www.w3.org/2000/09/xmldsig#rsa-sha1).
-> AsymmetricKeyWrapAlgorithm – KwRsa15 – (http://www.w3.org/2001/04/xmlenc#rsa-1_5).
-> AsymmetricEncryptionAlgorithm – Rsa15 – (http://www.w3.org/2001/04/xmlenc#rsa-1_5).
-> KeyDerivationAlgorithm – PSha1 – (http://docs.oasis-open.org/ws-sx/ws-secureconversation/200512/dk/p_sha1).
-> DerivedSignatureKeyLength – 128.
-> MinAsymmetricKeyLength – 1024
-> MaxAsymmetricKeyLength – 2048
-> CertificateSignatureAlgorithm – Sha1

If a certificate or any certificate in the chain is not signed with a hash that is Sha1 or stronger then the certificate shall be rejected.
	Security Encryption Required 		Encryption is required using the algorithms provided in the security algorithm suite.
	Security Signing Required 			Signing is required using the algorithms provided in the security algorithm suite.


*/

func newBasic128Rsa15Symmetric(localNonce []byte, remoteNonce []byte) (*EncryptionAlgorithm, error) {
	const (
		signatureKeyLength  = 16
		encryptionKeyLength = 16
		encryptionBlockSize = AESBlockSize
	)

	localHmac := &HMAC{Hash: crypto.SHA1, Secret: localNonce}
	remoteHmac := &HMAC{Hash: crypto.SHA1, Secret: remoteNonce}

	localKeys := generateKeys(localHmac, remoteNonce, signatureKeyLength, encryptionKeyLength, encryptionBlockSize)
	remoteKeys := generateKeys(remoteHmac, localNonce, signatureKeyLength, encryptionKeyLength, encryptionBlockSize)

	return &EncryptionAlgorithm{
		blockSize:             AESBlockSize,
		plainttextBlockSize:   AESBlockSize - AESMinPadding,
		encrypt:               &AES{KeyLength: 128, IV: remoteKeys.iv, Secret: remoteKeys.encryption}, // AES128-CBC
		decrypt:               &AES{KeyLength: 128, IV: localKeys.iv, Secret: localKeys.encryption},   // AES128-CBC
		signature:             &HMAC{Hash: crypto.SHA1, Secret: remoteKeys.signing},                   // HMAC-SHA1
		verifySignature:       &HMAC{Hash: crypto.SHA1, Secret: localKeys.signing},                    // HMAC-SHA1
		signatureLength:       160 / 8,
		remoteSignatureLength: 160 / 8,
		encryptionURI:         "http://www.w3.org/2001/04/xmlenc#aes128-cbc",
		signatureURI:          "http://www.w3.org/2000/09/xmldsig#hmac-sha1",
	}, nil
}

func newBasic128Rsa15Asymmetric(localKey *rsa.PrivateKey, remoteKey *rsa.PublicKey) (*EncryptionAlgorithm, error) {
	const (
		minAsymmetricKeyLength = 128 // 1024 bits
		maxAsymmetricKeyLength = 256 // 2048 bits
		nonceLength            = 16
	)

	if localKey != nil && (localKey.PublicKey.Size() < minAsymmetricKeyLength || localKey.PublicKey.Size() > maxAsymmetricKeyLength) {
		msg := fmt.Sprintf("local key size should be %d-%d bytes, got %d bytes", minAsymmetricKeyLength, maxAsymmetricKeyLength, localKey.PublicKey.Size())
		return nil, errors.New(msg)
	}

	if remoteKey != nil && (remoteKey.Size() < minAsymmetricKeyLength || remoteKey.Size() > maxAsymmetricKeyLength) {
		msg := fmt.Sprintf("remote key size should be %d-%d bytes, got %d bytes", minAsymmetricKeyLength, maxAsymmetricKeyLength, remoteKey.Size())
		return nil, errors.New(msg)
	}

	var localKeySize, remoteKeySize int
	if localKey != nil {
		localKeySize = localKey.PublicKey.Size()
	}

	if remoteKey != nil {
		remoteKeySize = remoteKey.Size()
	}

	return &EncryptionAlgorithm{
		blockSize:             remoteKeySize,
		plainttextBlockSize:   remoteKeySize - PKCS1v15MinPadding,
		encrypt:               &PKCS1v15{PublicKey: remoteKey},                    // RSA-SHA15+KWRSA15
		decrypt:               &PKCS1v15{PrivateKey: localKey},                    // RSA-SHA15+KWRSA15
		signature:             &PKCS1v15{Hash: crypto.SHA1, PrivateKey: localKey}, // RSA-SHA1
		verifySignature:       &PKCS1v15{Hash: crypto.SHA1, PublicKey: remoteKey}, // RSA-SHA1
		nonceLength:           nonceLength,
		signatureLength:       localKeySize,
		remoteSignatureLength: remoteKeySize,
		encryptionURI:         "http://www.w3.org/2001/04/xmlenc#rsa-1_5",
		signatureURI:          "http://www.w3.org/2000/09/xmldsig#rsa-sha1",
	}, nil
}
