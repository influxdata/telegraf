package uasc

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"

	"github.com/gopcua/opcua/ua"
	"github.com/gopcua/opcua/uapolicy"
)

// signAndEncrypt encrypts the message bytes stored in b and returns the
// data signed and encrypted per the security policy information from the
// secure channel.
func (s *SecureChannel) signAndEncrypt(m *Message, b []byte) ([]byte, error) {
	// Nothing to do
	if s.cfg.SecurityMode == ua.MessageSecurityModeNone {
		return b, nil
	}

	var isAsymmetric bool
	if s.hasState(secureChannelCreated) {
		isAsymmetric = true
	}

	var headerLength int
	if isAsymmetric {
		headerLength = 12 + m.AsymmetricSecurityHeader.Len()
	} else {
		headerLength = 12 + m.SymmetricSecurityHeader.Len()
	}

	var encryptedLength int
	if s.cfg.SecurityMode == ua.MessageSecurityModeSignAndEncrypt || isAsymmetric {
		plaintextBlockSize := s.enc.PlaintextBlockSize()
		paddingLength := plaintextBlockSize - ((len(b[headerLength:]) + s.enc.SignatureLength() + 1) % plaintextBlockSize)

		for i := 0; i <= paddingLength; i++ {
			b = append(b, byte(paddingLength))
		}
		encryptedLength = ((len(b[headerLength:]) + s.enc.SignatureLength()) / plaintextBlockSize) * s.enc.BlockSize()
	} else { // MessageSecurityModeSign
		encryptedLength = len(b[headerLength:]) + s.enc.SignatureLength()
	}

	// Fix header size to account for signing / encryption
	binary.LittleEndian.PutUint32(b[4:], uint32(headerLength+encryptedLength))
	m.Header.MessageSize = uint32(headerLength + encryptedLength)

	signature, err := s.enc.Signature(b)
	if err != nil {
		return nil, ua.StatusBadSecurityChecksFailed
	}

	b = append(b, signature...)
	c := b[headerLength:]
	if s.cfg.SecurityMode == ua.MessageSecurityModeSignAndEncrypt || isAsymmetric {
		c, err = s.enc.Encrypt(c)
		if err != nil {
			return nil, ua.StatusBadSecurityChecksFailed
		}
	}
	return append(b[:headerLength], c...), nil
}

// verifyAndDecrypt decrypts an incoming message stored in b and returns the
// data in plaintext.  After decryption, the message signature is also verified.
// Any error in decryption or verification of the signature will return an error
// The result is stored in m.Data
func (s *SecureChannel) verifyAndDecrypt(m *MessageChunk, b []byte) ([]byte, error) {
	var err error

	var isAsymmetric bool
	if s.hasState(secureChannelCreated) {
		isAsymmetric = true
	}

	var headerLength int
	if isAsymmetric {
		headerLength = 12 + m.AsymmetricSecurityHeader.Len()
	} else {
		headerLength = 12 + m.SymmetricSecurityHeader.Len()
	}

	// Nothing to do
	if s.cfg.SecurityMode == ua.MessageSecurityModeNone {
		return m.Data, nil
	}

	if s.cfg.SecurityMode == ua.MessageSecurityModeSignAndEncrypt || isAsymmetric {
		p, err := s.enc.Decrypt(b[headerLength:])
		if err != nil {
			return nil, ua.StatusBadSecurityChecksFailed
		}
		b = append(b[:headerLength], p...)
	}

	signature := b[len(b)-s.enc.RemoteSignatureLength():]
	messageToVerify := b[:len(b)-s.enc.RemoteSignatureLength()]

	if err = s.enc.VerifySignature(messageToVerify, signature); err != nil {
		return nil, ua.StatusBadSecurityChecksFailed
	}

	var paddingLength int
	if s.cfg.SecurityMode == ua.MessageSecurityModeSignAndEncrypt || isAsymmetric {
		paddingLength = int(messageToVerify[len(messageToVerify)-1]) + 1
	}

	b = messageToVerify[headerLength : len(messageToVerify)-paddingLength]

	return b, nil
}

// NewSessionSignature issues a new signature for the client to send on the next ActivateSessionRequest
func (s *SecureChannel) NewSessionSignature(cert, nonce []byte) ([]byte, string, error) {

	if s.cfg.SecurityMode == ua.MessageSecurityModeNone {
		return nil, "", nil
	}

	remoteX509Cert, err := x509.ParseCertificate(cert)
	if err != nil {
		return nil, "", err
	}
	remoteKey := remoteX509Cert.PublicKey.(*rsa.PublicKey)

	enc, err := uapolicy.Asymmetric(s.cfg.SecurityPolicyURI, s.cfg.LocalKey, remoteKey)
	if err != nil {
		return nil, "", err
	}

	sig, err := enc.Signature(append(cert, nonce...))
	if err != nil {
		return nil, "", err
	}
	sigAlg := enc.SignatureURI()

	return sig, sigAlg, nil
}

// VerifySessionSignature checks the integrity of a Create/Activate Session Response's signature
func (s *SecureChannel) VerifySessionSignature(cert, nonce, signature []byte) error {

	if s.cfg.SecurityMode == ua.MessageSecurityModeNone {
		return nil
	}

	remoteX509Cert, err := x509.ParseCertificate(cert)
	if err != nil {
		return err
	}
	remoteKey := remoteX509Cert.PublicKey.(*rsa.PublicKey)

	enc, err := uapolicy.Asymmetric(s.cfg.SecurityPolicyURI, s.cfg.LocalKey, remoteKey)
	if err != nil {
		return err
	}
	err = enc.VerifySignature(append(s.cfg.Certificate, nonce...), signature)
	if err != nil {
		return err
	}

	return nil
}

// EncryptUserPassword issues a new signature for the client to send in ActivateSessionRequest
func (s *SecureChannel) EncryptUserPassword(policyURI, password string, cert, nonce []byte) ([]byte, string, error) {

	if policyURI == ua.SecurityPolicyURINone {
		return []byte(password), "", nil
	}

	// If the User ID Token's policy was null, then default to the secure channel's policy
	if policyURI == "" {
		policyURI = s.cfg.SecurityPolicyURI
	}

	remoteX509Cert, err := x509.ParseCertificate(cert)
	if err != nil {
		return nil, "", err
	}
	remoteKey := remoteX509Cert.PublicKey.(*rsa.PublicKey)

	enc, err := uapolicy.Asymmetric(policyURI, s.cfg.LocalKey, remoteKey)
	if err != nil {
		return nil, "", err
	}

	l := len(password) + len(nonce)
	secret := make([]byte, 4)
	binary.LittleEndian.PutUint32(secret, uint32(l))
	secret = append(secret, []byte(password)...)
	secret = append(secret, nonce...)
	pass, err := enc.Encrypt(secret)
	if err != nil {
		return nil, "", err
	}
	passAlg := enc.EncryptionURI()

	return pass, passAlg, nil
}

// NewUserTokenSignature issues a new signature for the client to send in ActivateSessionRequest
func (s *SecureChannel) NewUserTokenSignature(policyURI string, cert, nonce []byte) ([]byte, string, error) {

	if policyURI == ua.SecurityPolicyURINone {
		return nil, "", nil
	}

	remoteX509Cert, err := x509.ParseCertificate(cert)
	if err != nil {
		return nil, "", err
	}
	remoteKey := remoteX509Cert.PublicKey.(*rsa.PublicKey)

	enc, err := uapolicy.Asymmetric(policyURI, s.cfg.LocalKey, remoteKey)
	if err != nil {
		return nil, "", err
	}

	sig, err := enc.Signature(append(cert, nonce...))
	if err != nil {
		return nil, "", err
	}
	sigAlg := enc.SignatureURI()

	return sig, sigAlg, nil
}
