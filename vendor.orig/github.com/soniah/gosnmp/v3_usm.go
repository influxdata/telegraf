// Copyright 2012-2018 The GoSNMP Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosnmp

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/md5"
	crand "crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"hash"
	"sync"
	"sync/atomic"
)

// SnmpV3AuthProtocol describes the authentication protocol in use by an authenticated SnmpV3 connection.
type SnmpV3AuthProtocol uint8

// NoAuth, MD5, and SHA are implemented
const (
	NoAuth SnmpV3AuthProtocol = 1
	MD5    SnmpV3AuthProtocol = 2
	SHA    SnmpV3AuthProtocol = 3
)

// SnmpV3PrivProtocol is the privacy protocol in use by an private SnmpV3 connection.
type SnmpV3PrivProtocol uint8

// NoPriv, DES implemented, AES planned
const (
	NoPriv SnmpV3PrivProtocol = 1
	DES    SnmpV3PrivProtocol = 2
	AES    SnmpV3PrivProtocol = 3
)

// UsmSecurityParameters is an implementation of SnmpV3SecurityParameters for the UserSecurityModel
type UsmSecurityParameters struct {
	// localAESSalt must be 64bit aligned to use with atomic operations.
	localAESSalt uint64
	localDESSalt uint32

	AuthoritativeEngineID    string
	AuthoritativeEngineBoots uint32
	AuthoritativeEngineTime  uint32
	UserName                 string
	AuthenticationParameters string
	PrivacyParameters        []byte

	AuthenticationProtocol SnmpV3AuthProtocol
	PrivacyProtocol        SnmpV3PrivProtocol

	AuthenticationPassphrase string
	PrivacyPassphrase        string

	secretKey  []byte
	privacyKey []byte

	Logger Logger
}

func (sp *UsmSecurityParameters) Log() {
	sp.Logger.Printf("SECURITY PARAMETERS:%+v", sp)
}

// Copy method for UsmSecurityParameters used to copy a SnmpV3SecurityParameters without knowing it's implementation
func (sp *UsmSecurityParameters) Copy() SnmpV3SecurityParameters {
	return &UsmSecurityParameters{AuthoritativeEngineID: sp.AuthoritativeEngineID,
		AuthoritativeEngineBoots: sp.AuthoritativeEngineBoots,
		AuthoritativeEngineTime:  sp.AuthoritativeEngineTime,
		UserName:                 sp.UserName,
		AuthenticationParameters: sp.AuthenticationParameters,
		PrivacyParameters:        sp.PrivacyParameters,
		AuthenticationProtocol:   sp.AuthenticationProtocol,
		PrivacyProtocol:          sp.PrivacyProtocol,
		AuthenticationPassphrase: sp.AuthenticationPassphrase,
		PrivacyPassphrase:        sp.PrivacyPassphrase,
		secretKey:                sp.secretKey,
		privacyKey:               sp.privacyKey,
		localDESSalt:             sp.localDESSalt,
		localAESSalt:             sp.localAESSalt,
		Logger:                   sp.Logger,
	}
}

func (sp *UsmSecurityParameters) getDefaultContextEngineID() string {
	return sp.AuthoritativeEngineID
}

func (sp *UsmSecurityParameters) setSecurityParameters(in SnmpV3SecurityParameters) error {
	var insp *UsmSecurityParameters
	var err error

	if insp, err = castUsmSecParams(in); err != nil {
		return err
	}

	if sp.AuthoritativeEngineID != insp.AuthoritativeEngineID {
		sp.AuthoritativeEngineID = insp.AuthoritativeEngineID
		if sp.AuthenticationProtocol > NoAuth {
			sp.secretKey = genlocalkey(sp.AuthenticationProtocol,
				sp.AuthenticationPassphrase,
				sp.AuthoritativeEngineID)
		}
		if sp.PrivacyProtocol > NoPriv {
			sp.privacyKey = genlocalkey(sp.AuthenticationProtocol,
				sp.PrivacyPassphrase,
				sp.AuthoritativeEngineID)
		}
	}
	sp.AuthoritativeEngineBoots = insp.AuthoritativeEngineBoots
	sp.AuthoritativeEngineTime = insp.AuthoritativeEngineTime

	return nil
}

func (sp *UsmSecurityParameters) validate(flags SnmpV3MsgFlags) error {

	securityLevel := flags & AuthPriv // isolate flags that determine security level

	switch securityLevel {
	case AuthPriv:
		if sp.PrivacyProtocol <= NoPriv {
			return fmt.Errorf("SecurityParameters.PrivacyProtocol is required")
		}
		fallthrough
	case AuthNoPriv:
		if sp.AuthenticationProtocol <= NoAuth {
			return fmt.Errorf("SecurityParameters.AuthenticationProtocol is required")
		}
		fallthrough
	case NoAuthNoPriv:
		if sp.UserName == "" {
			return fmt.Errorf("SecurityParameters.UserName is required")
		}
	default:
		return fmt.Errorf("MsgFlags must be populated with an appropriate security level")
	}

	if sp.PrivacyProtocol > NoPriv {
		if sp.PrivacyPassphrase == "" {
			return fmt.Errorf("SecurityParameters.PrivacyPassphrase is required when a privacy protocol is specified.")
		}
	}

	if sp.AuthenticationProtocol > NoAuth {
		if sp.AuthenticationPassphrase == "" {
			return fmt.Errorf("SecurityParameters.AuthenticationPassphrase is required when an authentication protocol is specified.")
		}
	}

	return nil
}

func (sp *UsmSecurityParameters) init(log Logger) error {
	var err error

	sp.Logger = log

	switch sp.PrivacyProtocol {
	case AES:
		salt := make([]byte, 8)
		_, err = crand.Read(salt)
		if err != nil {
			return fmt.Errorf("Error creating a cryptographically secure salt: %s\n", err.Error())
		}
		sp.localAESSalt = binary.BigEndian.Uint64(salt)
	case DES:
		salt := make([]byte, 4)
		_, err = crand.Read(salt)
		if err != nil {
			return fmt.Errorf("Error creating a cryptographically secure salt: %s\n", err.Error())
		}
		sp.localDESSalt = binary.BigEndian.Uint32(salt)
	}

	return nil
}

func castUsmSecParams(secParams SnmpV3SecurityParameters) (*UsmSecurityParameters, error) {
	s, ok := secParams.(*UsmSecurityParameters)
	if !ok || s == nil {
		return nil, fmt.Errorf("SecurityParameters is not of type *UsmSecurityParameters")
	}
	return s, nil
}

var (
	passwordKeyHashCache = make(map[string][]byte)
	passwordKeyHashMutex sync.RWMutex
)

// Common passwordToKey algorithm, "caches" the result to avoid extra computation each reuse
func cachedPasswordToKey(hash hash.Hash, hashType string, password string) []byte {
	cacheKey := hashType + ":" + password

	passwordKeyHashMutex.RLock()
	value := passwordKeyHashCache[cacheKey]
	passwordKeyHashMutex.RUnlock()

	if value != nil {
		return value
	}
	var pi int // password index
	for i := 0; i < 1048576; i += 64 {
		var chunk []byte
		for e := 0; e < 64; e++ {
			chunk = append(chunk, password[pi%len(password)])
			pi++
		}
		hash.Write(chunk)
	}
	hashed := hash.Sum(nil)

	passwordKeyHashMutex.Lock()
	passwordKeyHashCache[cacheKey] = hashed
	passwordKeyHashMutex.Unlock()

	return hashed
}

// MD5 HMAC key calculation algorithm
func md5HMAC(password string, engineID string) []byte {
	var compressed []byte

	compressed = cachedPasswordToKey(md5.New(), "MD5", password)

	local := md5.New()
	local.Write(compressed)
	local.Write([]byte(engineID))
	local.Write(compressed)
	final := local.Sum(nil)
	return final
}

// SHA HMAC key calculation algorithm
func shaHMAC(password string, engineID string) []byte {
	var hashed []byte

	hashed = cachedPasswordToKey(sha1.New(), "SHA1", password)

	local := sha1.New()
	local.Write(hashed)
	local.Write([]byte(engineID))
	local.Write(hashed)
	final := local.Sum(nil)
	return final
}

func genlocalkey(authProtocol SnmpV3AuthProtocol, passphrase string, engineID string) []byte {
	var secretKey []byte

	switch authProtocol {
	default:
		secretKey = md5HMAC(passphrase, engineID)
	case SHA:
		secretKey = shaHMAC(passphrase, engineID)
	}

	return secretKey
}

// http://tools.ietf.org/html/rfc2574#section-8.1.1.1
// localDESSalt needs to be incremented on every packet.
func (sp *UsmSecurityParameters) usmAllocateNewSalt() (interface{}, error) {
	var newSalt interface{}

	switch sp.PrivacyProtocol {
	case AES:
		newSalt = atomic.AddUint64(&(sp.localAESSalt), 1)
	default:
		newSalt = atomic.AddUint32(&(sp.localDESSalt), 1)
	}
	return newSalt, nil
}

func (sp *UsmSecurityParameters) usmSetSalt(newSalt interface{}) error {

	switch sp.PrivacyProtocol {
	case AES:
		aesSalt, ok := newSalt.(uint64)
		if !ok {
			return fmt.Errorf("salt provided to usmSetSalt is not the correct type for the AES privacy protocol")
		}
		var salt = make([]byte, 8)
		binary.BigEndian.PutUint64(salt, aesSalt)
		sp.PrivacyParameters = salt
	default:
		desSalt, ok := newSalt.(uint32)
		if !ok {
			return fmt.Errorf("salt provided to usmSetSalt is not the correct type for the DES privacy protocol")
		}
		var salt = make([]byte, 8)
		binary.BigEndian.PutUint32(salt, sp.AuthoritativeEngineBoots)
		binary.BigEndian.PutUint32(salt[4:], desSalt)
		sp.PrivacyParameters = salt
	}
	return nil
}

func (sp *UsmSecurityParameters) initPacket(packet *SnmpPacket) error {
	// http://tools.ietf.org/html/rfc2574#section-8.1.1.1
	// localDESSalt needs to be incremented on every packet.
	newSalt, err := sp.usmAllocateNewSalt()
	if err != nil {
		return err
	}
	if packet.MsgFlags&AuthPriv > AuthNoPriv {
		var s *UsmSecurityParameters
		if s, err = castUsmSecParams(packet.SecurityParameters); err != nil {
			return err
		}
		return s.usmSetSalt(newSalt)
	}

	return nil
}

func (sp *UsmSecurityParameters) discoveryRequired() *SnmpPacket {

	if sp.AuthoritativeEngineID == "" {
		var emptyPdus []SnmpPDU

		// send blank packet to discover authoriative engine ID/boots/time
		blankPacket := &SnmpPacket{
			Version:            Version3,
			MsgFlags:           Reportable | NoAuthNoPriv,
			SecurityModel:      UserSecurityModel,
			SecurityParameters: &UsmSecurityParameters{Logger: sp.Logger},
			PDUType:            GetRequest,
			Logger:             sp.Logger,
			Variables:          emptyPdus,
		}

		return blankPacket
	}
	return nil
}

func usmFindAuthParamStart(packet []byte) (uint32, error) {
	idx := bytes.Index(packet, []byte{byte(OctetString), 12,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0})

	if idx < 0 {
		return 0, fmt.Errorf("Unable to locate the position in packet to write authentication key")
	}

	return uint32(idx + 2), nil
}

func (sp *UsmSecurityParameters) authenticate(packet []byte) error {

	var extkey [64]byte

	copy(extkey[:], sp.secretKey)

	var k1, k2 [64]byte

	for i := 0; i < 64; i++ {
		k1[i] = extkey[i] ^ 0x36
		k2[i] = extkey[i] ^ 0x5c
	}

	var h, h2 hash.Hash

	switch sp.AuthenticationProtocol {
	default:
		h = md5.New()
		h2 = md5.New()
	case SHA:
		h = sha1.New()
		h2 = sha1.New()
	}

	h.Write(k1[:])
	h.Write(packet)
	d1 := h.Sum(nil)
	h2.Write(k2[:])
	h2.Write(d1)
	authParamStart, err := usmFindAuthParamStart(packet)
	if err != nil {
		return err
	}

	copy(packet[authParamStart:authParamStart+12], h2.Sum(nil)[:12])

	return nil
}

// determine whether a message is authentic
func (sp *UsmSecurityParameters) isAuthentic(packetBytes []byte, packet *SnmpPacket) (bool, error) {

	var packetSecParams *UsmSecurityParameters
	var err error

	if packetSecParams, err = castUsmSecParams(packet.SecurityParameters); err != nil {
		return false, err
	}
	// TODO: investigate call chain to determine if this is really the best spot for this

	var extkey [64]byte

	copy(extkey[:], sp.secretKey)

	var k1, k2 [64]byte

	for i := 0; i < 64; i++ {
		k1[i] = extkey[i] ^ 0x36
		k2[i] = extkey[i] ^ 0x5c
	}

	var h, h2 hash.Hash

	switch sp.AuthenticationProtocol {
	default:
		h = md5.New()
		h2 = md5.New()
	case SHA:
		h = sha1.New()
		h2 = sha1.New()
	}

	h.Write(k1[:])
	h.Write(packetBytes)
	d1 := h.Sum(nil)
	h2.Write(k2[:])
	h2.Write(d1)

	result := h2.Sum(nil)[:12]
	for k, v := range []byte(packetSecParams.AuthenticationParameters) {
		if result[k] != v {
			return false, nil
		}
	}
	return true, nil
}

func (sp *UsmSecurityParameters) encryptPacket(scopedPdu []byte) ([]byte, error) {
	var b []byte

	switch sp.PrivacyProtocol {
	case AES:
		var iv [16]byte
		binary.BigEndian.PutUint32(iv[:], sp.AuthoritativeEngineBoots)
		binary.BigEndian.PutUint32(iv[4:], sp.AuthoritativeEngineTime)
		copy(iv[8:], sp.PrivacyParameters)

		block, err := aes.NewCipher(sp.privacyKey[:16])
		if err != nil {
			return nil, err
		}
		stream := cipher.NewCFBEncrypter(block, iv[:])
		ciphertext := make([]byte, len(scopedPdu))
		stream.XORKeyStream(ciphertext, scopedPdu)
		pduLen, err := marshalLength(len(ciphertext))
		if err != nil {
			return nil, err
		}
		b = append([]byte{byte(OctetString)}, pduLen...)
		scopedPdu = append(b, ciphertext...)
	default:
		preiv := sp.privacyKey[8:]
		var iv [8]byte
		for i := 0; i < len(iv); i++ {
			iv[i] = preiv[i] ^ sp.PrivacyParameters[i]
		}
		block, err := des.NewCipher(sp.privacyKey[:8])
		if err != nil {
			return nil, err
		}
		mode := cipher.NewCBCEncrypter(block, iv[:])

		pad := make([]byte, des.BlockSize-len(scopedPdu)%des.BlockSize)
		scopedPdu = append(scopedPdu, pad...)

		ciphertext := make([]byte, len(scopedPdu))
		mode.CryptBlocks(ciphertext, scopedPdu)
		pduLen, err := marshalLength(len(ciphertext))
		if err != nil {
			return nil, err
		}
		b = append([]byte{byte(OctetString)}, pduLen...)
		scopedPdu = append(b, ciphertext...)
	}

	return scopedPdu, nil
}

func (sp *UsmSecurityParameters) decryptPacket(packet []byte, cursor int) ([]byte, error) {
	_, cursorTmp := parseLength(packet[cursor:])
	cursorTmp += cursor

	switch sp.PrivacyProtocol {
	case AES:
		var iv [16]byte
		binary.BigEndian.PutUint32(iv[:], sp.AuthoritativeEngineBoots)
		binary.BigEndian.PutUint32(iv[4:], sp.AuthoritativeEngineTime)
		copy(iv[8:], sp.PrivacyParameters)

		block, err := aes.NewCipher(sp.privacyKey[:16])
		if err != nil {
			return nil, err
		}
		stream := cipher.NewCFBDecrypter(block, iv[:])
		plaintext := make([]byte, len(packet[cursorTmp:]))
		stream.XORKeyStream(plaintext, packet[cursorTmp:])
		copy(packet[cursor:], plaintext)
		packet = packet[:cursor+len(plaintext)]
	default:
		if len(packet[cursorTmp:])%des.BlockSize != 0 {
			return nil, fmt.Errorf("Error decrypting ScopedPDU: not multiple of des block size.")
		}
		preiv := sp.privacyKey[8:]
		var iv [8]byte
		for i := 0; i < len(iv); i++ {
			iv[i] = preiv[i] ^ sp.PrivacyParameters[i]
		}
		block, err := des.NewCipher(sp.privacyKey[:8])
		if err != nil {
			return nil, err
		}
		mode := cipher.NewCBCDecrypter(block, iv[:])

		plaintext := make([]byte, len(packet[cursorTmp:]))
		mode.CryptBlocks(plaintext, packet[cursorTmp:])
		copy(packet[cursor:], plaintext)
		// truncate packet to remove extra space caused by the
		// octetstring/length header that was just replaced
		packet = packet[:cursor+len(plaintext)]
	}
	return packet, nil
}

// marshal a snmp version 3 security parameters field for the User Security Model
func (sp *UsmSecurityParameters) marshal(flags SnmpV3MsgFlags) ([]byte, error) {
	var buf bytes.Buffer
	var err error

	// msgAuthoritativeEngineID
	buf.Write([]byte{byte(OctetString), byte(len(sp.AuthoritativeEngineID))})
	buf.WriteString(sp.AuthoritativeEngineID)

	// msgAuthoritativeEngineBoots
	msgAuthoritativeEngineBoots := marshalUvarInt(sp.AuthoritativeEngineBoots)
	buf.Write([]byte{byte(Integer), byte(len(msgAuthoritativeEngineBoots))})
	buf.Write(msgAuthoritativeEngineBoots)

	// msgAuthoritativeEngineTime
	msgAuthoritativeEngineTime := marshalUvarInt(sp.AuthoritativeEngineTime)
	buf.Write([]byte{byte(Integer), byte(len(msgAuthoritativeEngineTime))})
	buf.Write(msgAuthoritativeEngineTime)

	// msgUserName
	buf.Write([]byte{byte(OctetString), byte(len(sp.UserName))})
	buf.WriteString(sp.UserName)

	// msgAuthenticationParameters
	if flags&AuthNoPriv > 0 {
		buf.Write([]byte{byte(OctetString), 12,
			0, 0, 0, 0,
			0, 0, 0, 0,
			0, 0, 0, 0})
	} else {
		buf.Write([]byte{byte(OctetString), 0})
	}
	// msgPrivacyParameters
	if flags&AuthPriv > AuthNoPriv {
		privlen, err := marshalLength(len(sp.PrivacyParameters))
		if err != nil {
			return nil, err
		}
		buf.Write([]byte{byte(OctetString)})
		buf.Write(privlen)
		buf.Write(sp.PrivacyParameters)
	} else {
		buf.Write([]byte{byte(OctetString), 0})
	}

	// wrap security parameters in a sequence
	paramLen, err := marshalLength(buf.Len())
	if err != nil {
		return nil, err
	}
	tmpseq := append([]byte{byte(Sequence)}, paramLen...)
	tmpseq = append(tmpseq, buf.Bytes()...)

	return tmpseq, nil
}

func (sp *UsmSecurityParameters) unmarshal(flags SnmpV3MsgFlags, packet []byte, cursor int) (int, error) {

	var err error

	if PDUType(packet[cursor]) != Sequence {
		return 0, fmt.Errorf("Error parsing SNMPV3 User Security Model parameters\n")
	}
	_, cursorTmp := parseLength(packet[cursor:])
	cursor += cursorTmp

	rawMsgAuthoritativeEngineID, count, err := parseRawField(packet[cursor:], "msgAuthoritativeEngineID")
	if err != nil {
		return 0, fmt.Errorf("Error parsing SNMPV3 User Security Model msgAuthoritativeEngineID: %s", err.Error())
	}
	cursor += count
	if AuthoritativeEngineID, ok := rawMsgAuthoritativeEngineID.(string); ok {
		if sp.AuthoritativeEngineID != AuthoritativeEngineID {
			sp.AuthoritativeEngineID = AuthoritativeEngineID
			sp.Logger.Printf("Parsed authoritativeEngineID %s", AuthoritativeEngineID)
			if sp.AuthenticationProtocol > NoAuth {
				sp.secretKey = genlocalkey(sp.AuthenticationProtocol,
					sp.AuthenticationPassphrase,
					sp.AuthoritativeEngineID)
			}
			if sp.PrivacyProtocol > NoPriv {
				sp.privacyKey = genlocalkey(sp.AuthenticationProtocol,
					sp.PrivacyPassphrase,
					sp.AuthoritativeEngineID)
			}
		}
	}

	rawMsgAuthoritativeEngineBoots, count, err := parseRawField(packet[cursor:], "msgAuthoritativeEngineBoots")
	if err != nil {
		return 0, fmt.Errorf("Error parsing SNMPV3 User Security Model msgAuthoritativeEngineBoots: %s", err.Error())
	}
	cursor += count
	if AuthoritativeEngineBoots, ok := rawMsgAuthoritativeEngineBoots.(int); ok {
		sp.AuthoritativeEngineBoots = uint32(AuthoritativeEngineBoots)
		sp.Logger.Printf("Parsed authoritativeEngineBoots %d", AuthoritativeEngineBoots)
	}

	rawMsgAuthoritativeEngineTime, count, err := parseRawField(packet[cursor:], "msgAuthoritativeEngineTime")
	if err != nil {
		return 0, fmt.Errorf("Error parsing SNMPV3 User Security Model msgAuthoritativeEngineTime: %s", err.Error())
	}
	cursor += count
	if AuthoritativeEngineTime, ok := rawMsgAuthoritativeEngineTime.(int); ok {
		sp.AuthoritativeEngineTime = uint32(AuthoritativeEngineTime)
		sp.Logger.Printf("Parsed authoritativeEngineTime %d", AuthoritativeEngineTime)
	}

	rawMsgUserName, count, err := parseRawField(packet[cursor:], "msgUserName")
	if err != nil {
		return 0, fmt.Errorf("Error parsing SNMPV3 User Security Model msgUserName: %s", err.Error())
	}
	cursor += count
	if msgUserName, ok := rawMsgUserName.(string); ok {
		sp.UserName = msgUserName
		sp.Logger.Printf("Parsed userName %s", msgUserName)
	}

	rawMsgAuthParameters, count, err := parseRawField(packet[cursor:], "msgAuthenticationParameters")
	if err != nil {
		return 0, fmt.Errorf("Error parsing SNMPV3 User Security Model msgAuthenticationParameters: %s", err.Error())
	}
	if msgAuthenticationParameters, ok := rawMsgAuthParameters.(string); ok {
		sp.AuthenticationParameters = msgAuthenticationParameters
		sp.Logger.Printf("Parsed authenticationParameters %s", msgAuthenticationParameters)
	}
	// blank msgAuthenticationParameters to prepare for authentication check later
	if flags&AuthNoPriv > 0 {
		blank := make([]byte, 12)
		copy(packet[cursor+2:cursor+14], blank)
	}
	cursor += count

	rawMsgPrivacyParameters, count, err := parseRawField(packet[cursor:], "msgPrivacyParameters")
	if err != nil {
		return 0, fmt.Errorf("Error parsing SNMPV3 User Security Model msgPrivacyParameters: %s", err.Error())
	}
	cursor += count
	if msgPrivacyParameters, ok := rawMsgPrivacyParameters.(string); ok {
		sp.PrivacyParameters = []byte(msgPrivacyParameters)
		sp.Logger.Printf("Parsed privacyParameters %s", msgPrivacyParameters)
	}

	return cursor, nil
}
