// Copyright 2018-2020 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package uapolicy

/*
Byte[] PRF(
	Byte[] secret,
	Byte[] seed,
	Int32 length,
	Int32 offset
)
Where length is the number of bytes to return and offset is a number of bytes from the beginning
of the sequence.

Where length is the number of bytes to return and offset is a number of bytes from the beginning
of the sequence.
The lengths of the keys that need to be generated depend on the SecurityPolicy used for the
channel. The following information is specified by the SecurityPolicy:
	a) SigningKeyLength (from the DerivedSignatureKeyLength);
	b) EncryptingKeyLength (implied by the SymmetricEncryptionAlgorithm);
	c) EncryptingBlockSize (implied by the SymmetricEncryptionAlgorithm).

Name			Derivation
ClientSecret	The value of the ClientNonce provided in the OpenSecureChannel request.
ClientSeed		The value of the ClientNonce provided in the OpenSecureChannel request.
ServerSecret	The value of the ServerNonce provided in the OpenSecureChannel response.
ServerSeed		The value of the ServerNonce provided in the OpenSecureChannel response.

 Key						Secret			Seed		Length				Offset
 ClientSigningKey			ServerSecret	ClientSeed	SigningKeyLength	0
 ClientEncryptingKey		ServerSecret	ClientSeed	EncryptingKeyLength	SigningKeyLength
 ClientInitializationVector	ServerSecret	ClientSeed	EncryptingBlockSize	SigningKeyLength+ EncryptingKeyLength
 ServerSigningKey			ClientSecret	ServerSeed	SigningKeyLength	0
 ServerEncryptingKey		ClientSecret	ServerSeed	EncryptingKeyLength	SigningKeyLength
 ServerInitializationVector	ClientSecret	ServerSeed	EncryptingBlockSize	SigningKeyLength+ EncryptingKeyLength

*/

type derivedKeys struct {
	signing, encryption, iv []byte
}

func generateKeys(hmac *HMAC, seed []byte, signingLength, encryptingLength, encryptingBlockSize int) *derivedKeys {
	var p []byte
	a, _ := hmac.Signature(seed)
	for len(p) < signingLength+encryptingLength+encryptingBlockSize {
		input := append(a, seed...)
		h, _ := hmac.Signature(input)
		p = append(p, h...)
		a, _ = hmac.Signature(a)
	}

	return &derivedKeys{
		signing:    p[:signingLength],
		encryption: p[signingLength : signingLength+encryptingLength],
		iv:         p[signingLength+encryptingLength : signingLength+encryptingLength+encryptingBlockSize],
	}
}
