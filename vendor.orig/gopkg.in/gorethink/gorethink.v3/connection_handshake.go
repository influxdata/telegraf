package gorethink

import (
	"bufio"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"

	p "gopkg.in/gorethink/gorethink.v3/ql2"
)

type HandshakeVersion int

const (
	HandshakeV1_0 HandshakeVersion = iota
	HandshakeV0_4
)

type connectionHandshake interface {
	Send() error
}

func (c *Connection) handshake(version HandshakeVersion) (connectionHandshake, error) {
	switch version {
	case HandshakeV0_4:
		return &connectionHandshakeV0_4{conn: c}, nil
	case HandshakeV1_0:
		return &connectionHandshakeV1_0{conn: c}, nil
	default:
		return nil, fmt.Errorf("Unrecognised handshake version")
	}
}

type connectionHandshakeV0_4 struct {
	conn *Connection
}

func (c *connectionHandshakeV0_4) Send() error {
	// Send handshake request
	if err := c.writeHandshakeReq(); err != nil {
		c.conn.Close()
		return RQLConnectionError{rqlError(err.Error())}
	}
	// Read handshake response
	if err := c.readHandshakeSuccess(); err != nil {
		c.conn.Close()
		return RQLConnectionError{rqlError(err.Error())}
	}

	return nil
}

func (c *connectionHandshakeV0_4) writeHandshakeReq() error {
	pos := 0
	dataLen := 4 + 4 + len(c.conn.opts.AuthKey) + 4
	data := make([]byte, dataLen)

	// Send the protocol version to the server as a 4-byte little-endian-encoded integer
	binary.LittleEndian.PutUint32(data[pos:], uint32(p.VersionDummy_V0_4))
	pos += 4

	// Send the length of the auth key to the server as a 4-byte little-endian-encoded integer
	binary.LittleEndian.PutUint32(data[pos:], uint32(len(c.conn.opts.AuthKey)))
	pos += 4

	// Send the auth key as an ASCII string
	if len(c.conn.opts.AuthKey) > 0 {
		pos += copy(data[pos:], c.conn.opts.AuthKey)
	}

	// Send the protocol type as a 4-byte little-endian-encoded integer
	binary.LittleEndian.PutUint32(data[pos:], uint32(p.VersionDummy_JSON))
	pos += 4

	return c.conn.writeData(data)
}

func (c *connectionHandshakeV0_4) readHandshakeSuccess() error {
	reader := bufio.NewReader(c.conn.Conn)
	line, err := reader.ReadBytes('\x00')
	if err != nil {
		if err == io.EOF {
			return fmt.Errorf("Unexpected EOF: %s", string(line))
		}
		return err
	}
	// convert to string and remove trailing NUL byte
	response := string(line[:len(line)-1])
	if response != "SUCCESS" {
		response = strings.TrimSpace(response)
		// we failed authorization or something else terrible happened
		return RQLDriverError{rqlError(fmt.Sprintf("Server dropped connection with message: \"%s\"", response))}
	}

	return nil
}

const (
	handshakeV1_0_protocolVersionNumber = 0
	handshakeV1_0_authenticationMethod  = "SCRAM-SHA-256"
)

type connectionHandshakeV1_0 struct {
	conn   *Connection
	reader *bufio.Reader

	authMsg string
}

func (c *connectionHandshakeV1_0) Send() error {
	c.reader = bufio.NewReader(c.conn.Conn)

	// Generate client nonce
	clientNonce, err := c.generateNonce()
	if err != nil {
		c.conn.Close()
		return RQLDriverError{rqlError(fmt.Sprintf("Failed to generate client nonce: %s", err))}
	}
	// Send client first message
	if err := c.writeFirstMessage(clientNonce); err != nil {
		c.conn.Close()
		return err
	}
	// Read status
	if err := c.checkServerVersions(); err != nil {
		c.conn.Close()
		return err
	}

	// Read server first message
	i, salt, serverNonce, err := c.readFirstMessage()
	if err != nil {
		c.conn.Close()
		return err
	}

	// Check server nonce
	if !strings.HasPrefix(serverNonce, clientNonce) {
		return RQLAuthError{RQLDriverError{rqlError("Invalid nonce from server")}}
	}

	// Generate proof
	saltedPass := c.saltPassword(i, salt)
	clientProof := c.calculateProof(saltedPass, clientNonce, serverNonce)
	serverSignature := c.serverSignature(saltedPass)

	// Send client final message
	if err := c.writeFinalMessage(serverNonce, clientProof); err != nil {
		c.conn.Close()
		return err
	}
	// Read server final message
	if err := c.readFinalMessage(serverSignature); err != nil {
		c.conn.Close()
		return err
	}

	return nil
}

func (c *connectionHandshakeV1_0) writeFirstMessage(clientNonce string) error {
	// Default username to admin if not set
	username := "admin"
	if c.conn.opts.Username != "" {
		username = c.conn.opts.Username
	}

	c.authMsg = fmt.Sprintf("n=%s,r=%s", username, clientNonce)
	msg := fmt.Sprintf(
		`{"protocol_version": %d,"authentication": "n,,%s","authentication_method": "%s"}`,
		handshakeV1_0_protocolVersionNumber, c.authMsg, handshakeV1_0_authenticationMethod,
	)

	pos := 0
	dataLen := 4 + len(msg) + 1
	data := make([]byte, dataLen)

	// Send the protocol version to the server as a 4-byte little-endian-encoded integer
	binary.LittleEndian.PutUint32(data[pos:], uint32(p.VersionDummy_V1_0))
	pos += 4

	// Send the auth message as an ASCII string
	pos += copy(data[pos:], msg)

	// Add null terminating byte
	data[pos] = '\x00'

	return c.writeData(data)
}

func (c *connectionHandshakeV1_0) checkServerVersions() error {
	b, err := c.readResponse()
	if err != nil {
		return err
	}

	// Read status
	type versionsResponse struct {
		Success            bool   `json:"success"`
		MinProtocolVersion int    `json:"min_protocol_version"`
		MaxProtocolVersion int    `json:"max_protocol_version"`
		ServerVersion      string `json:"server_version"`
		ErrorCode          int    `json:"error_code"`
		Error              string `json:"error"`
	}
	var rsp *versionsResponse
	statusStr := string(b)

	if err := json.Unmarshal(b, &rsp); err != nil {
		if strings.HasPrefix(statusStr, "ERROR: ") {
			statusStr = strings.TrimPrefix(statusStr, "ERROR: ")
			return RQLConnectionError{rqlError(statusStr)}
		}

		return RQLDriverError{rqlError(fmt.Sprintf("Error reading versions: %s", err))}
	}

	if !rsp.Success {
		return c.handshakeError(rsp.ErrorCode, rsp.Error)
	}
	if rsp.MinProtocolVersion > handshakeV1_0_protocolVersionNumber ||
		rsp.MaxProtocolVersion < handshakeV1_0_protocolVersionNumber {
		return RQLDriverError{rqlError(
			fmt.Sprintf(
				"Unsupported protocol version %d, expected between %d and %d.",
				handshakeV1_0_protocolVersionNumber,
				rsp.MinProtocolVersion,
				rsp.MaxProtocolVersion,
			),
		)}
	}

	return nil
}

func (c *connectionHandshakeV1_0) readFirstMessage() (i int64, salt []byte, serverNonce string, err error) {
	b, err2 := c.readResponse()
	if err2 != nil {
		err = err2
		return
	}

	// Read server message
	type firstMessageResponse struct {
		Success        bool   `json:"success"`
		Authentication string `json:"authentication"`
		ErrorCode      int    `json:"error_code"`
		Error          string `json:"error"`
	}
	var rsp *firstMessageResponse

	if err2 := json.Unmarshal(b, &rsp); err2 != nil {
		err = RQLDriverError{rqlError(fmt.Sprintf("Error parsing auth response: %s", err2))}
		return
	}
	if !rsp.Success {
		err = c.handshakeError(rsp.ErrorCode, rsp.Error)
		return
	}

	c.authMsg += ","
	c.authMsg += rsp.Authentication

	// Parse authentication field
	auth := map[string]string{}
	parts := strings.Split(rsp.Authentication, ",")
	for _, part := range parts {
		i := strings.Index(part, "=")
		if i != -1 {
			auth[part[:i]] = part[i+1:]
		}
	}

	// Extract return values
	if v, ok := auth["i"]; ok {
		i, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return
		}
	}
	if v, ok := auth["s"]; ok {
		salt, err = base64.StdEncoding.DecodeString(v)
		if err != nil {
			return
		}
	}
	if v, ok := auth["r"]; ok {
		serverNonce = v
	}

	return
}

func (c *connectionHandshakeV1_0) writeFinalMessage(serverNonce, clientProof string) error {
	authMsg := "c=biws,r="
	authMsg += serverNonce
	authMsg += ",p="
	authMsg += clientProof

	msg := fmt.Sprintf(`{"authentication": "%s"}`, authMsg)

	pos := 0
	dataLen := len(msg) + 1
	data := make([]byte, dataLen)

	// Send the auth message as an ASCII string
	pos += copy(data[pos:], msg)

	// Add null terminating byte
	data[pos] = '\x00'

	return c.writeData(data)
}

func (c *connectionHandshakeV1_0) readFinalMessage(serverSignature string) error {
	b, err := c.readResponse()
	if err != nil {
		return err
	}

	// Read server message
	type finalMessageResponse struct {
		Success        bool   `json:"success"`
		Authentication string `json:"authentication"`
		ErrorCode      int    `json:"error_code"`
		Error          string `json:"error"`
	}
	var rsp *finalMessageResponse

	if err := json.Unmarshal(b, &rsp); err != nil {
		return RQLDriverError{rqlError(fmt.Sprintf("Error parsing auth response: %s", err))}
	}
	if !rsp.Success {
		return c.handshakeError(rsp.ErrorCode, rsp.Error)
	}

	// Parse authentication field
	auth := map[string]string{}
	parts := strings.Split(rsp.Authentication, ",")
	for _, part := range parts {
		i := strings.Index(part, "=")
		if i != -1 {
			auth[part[:i]] = part[i+1:]
		}
	}

	// Validate server response
	if serverSignature != auth["v"] {
		return RQLAuthError{RQLDriverError{rqlError("Invalid server signature")}}
	}

	return nil
}

func (c *connectionHandshakeV1_0) writeData(data []byte) error {

	if err := c.conn.writeData(data); err != nil {
		return RQLConnectionError{rqlError(err.Error())}
	}

	return nil
}

func (c *connectionHandshakeV1_0) readResponse() ([]byte, error) {
	line, err := c.reader.ReadBytes('\x00')
	if err != nil {
		if err == io.EOF {
			return nil, RQLConnectionError{rqlError(fmt.Sprintf("Unexpected EOF: %s", string(line)))}
		}
		return nil, RQLConnectionError{rqlError(err.Error())}
	}

	// Strip null byte and return
	return line[:len(line)-1], nil
}

func (c *connectionHandshakeV1_0) generateNonce() (string, error) {
	const nonceSize = 24

	b := make([]byte, nonceSize)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

func (c *connectionHandshakeV1_0) saltPassword(iter int64, salt []byte) []byte {
	pass := []byte(c.conn.opts.Password)

	return pbkdf2.Key(pass, salt, int(iter), sha256.Size, sha256.New)
}

func (c *connectionHandshakeV1_0) calculateProof(saltedPass []byte, clientNonce, serverNonce string) string {
	// Generate proof
	c.authMsg += ",c=biws,r=" + serverNonce

	mac := hmac.New(c.hashFunc(), saltedPass)
	mac.Write([]byte("Client Key"))
	clientKey := mac.Sum(nil)

	hash := c.hashFunc()()
	hash.Write(clientKey)
	storedKey := hash.Sum(nil)

	mac = hmac.New(c.hashFunc(), storedKey)
	mac.Write([]byte(c.authMsg))
	clientSignature := mac.Sum(nil)
	clientProof := make([]byte, len(clientKey))
	for i, _ := range clientKey {
		clientProof[i] = clientKey[i] ^ clientSignature[i]
	}

	return base64.StdEncoding.EncodeToString(clientProof)
}

func (c *connectionHandshakeV1_0) serverSignature(saltedPass []byte) string {
	mac := hmac.New(c.hashFunc(), saltedPass)
	mac.Write([]byte("Server Key"))
	serverKey := mac.Sum(nil)

	mac = hmac.New(c.hashFunc(), serverKey)
	mac.Write([]byte(c.authMsg))
	serverSignature := mac.Sum(nil)

	return base64.StdEncoding.EncodeToString(serverSignature)
}

func (c *connectionHandshakeV1_0) handshakeError(code int, message string) error {
	if code >= 10 || code <= 20 {
		return RQLAuthError{RQLDriverError{rqlError(message)}}
	}

	return RQLDriverError{rqlError(message)}
}

func (c *connectionHandshakeV1_0) hashFunc() func() hash.Hash {
	return sha256.New
}
