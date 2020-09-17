// Copyright 2018-2020 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package uapolicy

/*

"SecurityPolicy - PubSub-Aes256-CTR" Profile
http://opcfoundation.org/UA/SecurityPolicy#PubSub-Aes256-CTR

Name 	Opt. 	 Description 	 From Profile
	Security Encryption Required 		Encryption is required using the algorithms provided in the security algorithm suite.
	Security Signing Required 		Signing is required using the algorithms provided in the security algorithm suite.
	SymmetricSignatureAlgorithm_HMAC-SHA2-256 		A keyed hash used for message authentication which is defined in https://tools.ietf.org/html/rfc2104.
The hash algorithm is SHA2 with 256 bits and described in https://tools.ietf.org/html/rfc4634

	SymmetricEncryptionAlgorithm_AES256-CTR 		The AES encryption algorithm which is defined in http://nvlpubs.nist.gov/nistpubs/FIPS/NIST.FIPS.197.pdf.
Multiple blocks encrypted using the CTR mode described in http://nvlpubs.nist.gov/nistpubs/Legacy/SP/nistspecialpublication800-38a.pdf.
The counter block format is defined in https://tools.ietf.org/html/rfc3686.
The key size is 256 bits. The block size is 16 bytes. The input nonce length is 4 bytes.
The URI is http://opcfoundation.org/UA/security/aes256-ctr.
	AsymmetricSignatureAlgorithm_None 		This algorithm does not apply.
	AsymmetricEncryptionAlgorithm_None 		This algorithm does not apply.
	KeyDerivationAlgorithm_P-SHA2-256 		The P_SHA256 pseudo-random function defined in https://tools.ietf.org/html/rfc5246.
The URI is http://docs.oasis-open.org/ws-sx/ws-secureconversation/200512/dk/p_sha256.
	CertificateSignatureAlgorithm_None 		This algorithm does not apply.
	PubSub-Aes256-CTR_Limits 		-> DerivedSignatureKeyLength: 256 bits
-> MinAsymmetricKeyLength: n/a
-> MaxAsymmetricKeyLength: n/a
-> SecureChannelNonceLength: n/a


*/
