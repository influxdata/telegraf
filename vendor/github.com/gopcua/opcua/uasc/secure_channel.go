// Copyright 2018-2020 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package uasc

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"io"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gopcua/opcua/debug"
	"github.com/gopcua/opcua/errors"
	"github.com/gopcua/opcua/id"
	"github.com/gopcua/opcua/ua"
	"github.com/gopcua/opcua/uacp"
	"github.com/gopcua/opcua/uapolicy"
)

const (
	secureChannelCreated int32 = iota
	secureChannelOpen
	secureChannelClosed
	timeoutLeniency = 250 * time.Millisecond
	MaxTimeout      = math.MaxUint32 * time.Millisecond
)

type Response struct {
	ReqID uint32
	SCID  uint32
	V     interface{}
	Err   error
}

type SecureChannel struct {
	EndpointURL string

	// c is the uacp connection.
	c *uacp.Conn

	// cfg is the configuration for the secure channel.
	cfg *Config

	// reqhdr is the header for the next request.
	reqhdr *ua.RequestHeader

	// state is the state of the secure channel.
	// Must be accessed with atomic.LoadInt32/StoreInt32
	state int32

	// mu guards handler which contains the response channels
	// for the outstanding requests. The key is the request
	// handle which is part of the Request and Response headers.
	mu      sync.Mutex
	handler map[uint32]chan Response

	chunks map[uint32][]*MessageChunk

	enc *uapolicy.EncryptionAlgorithm

	// time returns the current time. When not set it defaults to time.Now().
	time func() time.Time

	// The lifetime of the SecurityToken in milliseconds. The UTC expiration time for the token
	// may be calculated by adding the lifetime to the createdAt time.
	lifetime uint32

	// secureChannelID is a unique identifier for the SecureChannel assigned by the Server.
	// If a Server receives a SecureChannelId which it does not recognize it shall return an
	// appropriate transport layer error.
	//
	// When a Server starts the first SecureChannelId used should be a value that is likely to
	// be unique after each restart. This ensures that a Server restart does not cause
	// previously connected Clients to accidentally ‘reuse’ SecureChannels that did not belong
	// to them.
	secureChannelID uint32

	// sequenceNumber is a monotonically increasing sequence number assigned by the sender to each
	// MessageChunk sent over the SecureChannel.
	sequenceNumber uint32

	// requestID is an identifier assigned by the Client to OPC UA request Message. All MessageChunks
	// for the request and the associated response use the same identifier
	requestID uint32

	// securityTokenID is a unique identifier for the SecureChannel SecurityToken used to secure the Message.
	// This identifier is returned by the Server in an OpenSecureChannel response Message.
	// If a Server receives a TokenId which it does not recognize it shall return an appropriate
	// transport layer error.
	securityTokenID uint32
}

func NewSecureChannel(endpoint string, c *uacp.Conn, cfg *Config) (*SecureChannel, error) {
	if c == nil {
		return nil, errors.Errorf("no connection")
	}
	if cfg == nil {
		return nil, errors.Errorf("no secure channel config")
	}

	if cfg.SecurityPolicyURI != ua.SecurityPolicyURINone {
		if cfg.SecurityMode == ua.MessageSecurityModeNone {
			return nil, errors.Errorf("invalid channel config: Security policy '%s' cannot be used with '%s'", cfg.SecurityPolicyURI, cfg.SecurityMode)
		}
		if cfg.LocalKey == nil {
			return nil, errors.Errorf("invalid channel config: Security policy '%s' requires a private key", cfg.SecurityPolicyURI)
		}
	}

	// Force the security mode to None if the policy is also None
	if cfg.SecurityPolicyURI == ua.SecurityPolicyURINone {
		cfg.SecurityMode = ua.MessageSecurityModeNone
	}

	return &SecureChannel{
		EndpointURL: endpoint,
		c:           c,
		cfg:         cfg,
		reqhdr: &ua.RequestHeader{
			TimeoutHint:      uint32(cfg.RequestTimeout / time.Millisecond),
			AdditionalHeader: ua.NewExtensionObject(nil),
		},
		state:     secureChannelCreated,
		handler:   make(map[uint32]chan Response),
		chunks:    make(map[uint32][]*MessageChunk),
		requestID: cfg.RequestIDSeed,
	}, nil
}

func (s *SecureChannel) LocalEndpoint() string {
	return s.EndpointURL
}

func (s *SecureChannel) Lifetime() uint32 {
	return s.lifetime
}

func (s *SecureChannel) setState(n int32) {
	atomic.StoreInt32(&s.state, n)
}

func (s *SecureChannel) hasState(n int32) bool {
	return atomic.LoadInt32(&s.state) == n
}

// SendRequest sends the service request and calls h with the response.
func (s *SecureChannel) SendRequest(req ua.Request, authToken *ua.NodeID, h func(interface{}) error) error {
	return s.SendRequestWithTimeout(req, authToken, s.cfg.RequestTimeout, h)
}

// SendRequestWithTimeout sends the service request and calls h with the response with a specific timeout.
func (s *SecureChannel) SendRequestWithTimeout(req ua.Request, authToken *ua.NodeID, timeout time.Duration, h func(interface{}) error) error {
	respRequired := h != nil

	ch, reqid, err := s.sendAsyncWithTimeout(req, authToken, respRequired, timeout)
	if err != nil {
		return err
	}

	if !respRequired {
		return nil
	}

	// `+ timeoutLeniency` to give the server a chance to respond to TimeoutHint
	timer := time.NewTimer(timeout + timeoutLeniency)
	defer timer.Stop()

	select {
	case resp := <-ch:
		if resp.Err != nil {
			if resp.V != nil {
				_ = h(resp.V) // ignore result because resp.Err takes precedence
			}
			return resp.Err
		}
		return h(resp.V)
	case <-timer.C:
		s.mu.Lock()
		s.popHandlerLock(reqid)
		s.mu.Unlock()
		return ua.StatusBadTimeout
	}
}

// sendAsyncWithTimeout sends the service request with a specific timeout and
// returns a channel which will receive the response when it arrives.
func (s *SecureChannel) sendAsyncWithTimeout(req ua.Request, authToken *ua.NodeID, respReq bool, timeout time.Duration) (resp chan Response, reqID uint32, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// encode the message
	m, err := s.newRequestMessage(req, authToken, timeout)
	if err != nil {
		return nil, 0, err
	}
	reqid := m.SequenceHeader.RequestID
	b, err := m.Encode()
	if err != nil {
		return nil, reqid, err
	}

	// encrypt the message prior to sending it
	// if SecurityMode == None, this returns the byte stream untouched
	b, err = s.signAndEncrypt(m, b)
	if err != nil {
		return nil, reqid, err
	}

	// send the message
	if _, err := s.c.Write(b); err != nil {
		return nil, reqid, err
	}
	debug.Printf("uasc %d/%d: send %T with %d bytes", s.c.ID(), reqid, req, len(b))

	// register the handler if a callback was passed
	if !respReq {
		return nil, 0, nil
	}
	resp = make(chan Response)
	if s.handler[reqid] != nil {
		return nil, reqid, errors.Errorf("error: duplicate handler registration for request id %d", reqid)
	}
	s.handler[reqid] = resp
	return resp, reqid, nil
}

// New creates a OPC UA Secure Conversation message.New
// MessageType of UASC is determined depending on the type of service given as below.
//
// Service type: OpenSecureChannel => Message type: OPN.
//
// Service type: CloseSecureChannel => Message type: CLO.
//
// Service type: Others => Message type: MSG.
//
func (s *SecureChannel) newMessage(srv interface{}, typeID uint16) *Message {
	switch typeID {
	case id.OpenSecureChannelRequest_Encoding_DefaultBinary, id.OpenSecureChannelResponse_Encoding_DefaultBinary:
		// Do not send the thumbprint for security mode None
		// even if we have a certificate.
		//
		// See https://github.com/gopcua/opcua/issues/259
		thumbprint := s.cfg.Thumbprint
		if s.cfg.SecurityMode == ua.MessageSecurityModeNone {
			thumbprint = nil
		}

		return &Message{
			MessageHeader: &MessageHeader{
				Header:                   NewHeader(MessageTypeOpenSecureChannel, ChunkTypeFinal, s.secureChannelID),
				AsymmetricSecurityHeader: NewAsymmetricSecurityHeader(s.cfg.SecurityPolicyURI, s.cfg.Certificate, thumbprint),
				SequenceHeader:           NewSequenceHeader(s.sequenceNumber, s.requestID),
			},
			TypeID:  ua.NewFourByteExpandedNodeID(0, typeID),
			Service: srv,
		}

	case id.CloseSecureChannelRequest_Encoding_DefaultBinary, id.CloseSecureChannelResponse_Encoding_DefaultBinary:
		return &Message{
			MessageHeader: &MessageHeader{
				Header:                  NewHeader(MessageTypeCloseSecureChannel, ChunkTypeFinal, s.secureChannelID),
				SymmetricSecurityHeader: NewSymmetricSecurityHeader(s.securityTokenID),
				SequenceHeader:          NewSequenceHeader(s.sequenceNumber, s.requestID),
			},
			TypeID:  ua.NewFourByteExpandedNodeID(0, typeID),
			Service: srv,
		}

	default:
		return &Message{
			MessageHeader: &MessageHeader{
				Header:                  NewHeader(MessageTypeMessage, ChunkTypeFinal, s.secureChannelID),
				SymmetricSecurityHeader: NewSymmetricSecurityHeader(s.securityTokenID),
				SequenceHeader:          NewSequenceHeader(s.sequenceNumber, s.requestID),
			},
			TypeID:  ua.NewFourByteExpandedNodeID(0, typeID),
			Service: srv,
		}
	}
}

func (s *SecureChannel) newRequestMessage(req ua.Request, authToken *ua.NodeID, timeout time.Duration) (*Message, error) {
	typeID := ua.ServiceTypeID(req)
	if typeID == 0 {
		return nil, errors.Errorf("unknown service %T. Did you call register?", req)
	}
	if authToken == nil {
		authToken = ua.NewTwoByteNodeID(0)
	}

	s.sequenceNumber++
	if s.sequenceNumber > math.MaxUint32-1023 {
		s.sequenceNumber = 1
	}
	s.requestID++
	if s.requestID == 0 {
		s.requestID = 1
	}
	s.reqhdr.RequestHandle++
	if s.reqhdr.RequestHandle == 0 {
		s.reqhdr.RequestHandle = 1
	}
	s.reqhdr.AuthenticationToken = authToken
	s.reqhdr.Timestamp = s.timeNow()
	if timeout > 0 && timeout < s.cfg.RequestTimeout {
		timeout = s.cfg.RequestTimeout
	}
	s.reqhdr.TimeoutHint = uint32(timeout / time.Millisecond)
	req.SetHeader(s.reqhdr)

	// encode the message
	return s.newMessage(req, typeID), nil
}

// SendResponse sends a service response.
// todo(fs): this method is most likely needed for the server and we haven't tested it yet.
// todo(fs): it exists to implement the handleOpenSecureChannelRequest() method during the
// todo(fs): refactor to remove the reflect code. It will likely change.
func (s *SecureChannel) SendResponse(req ua.Response) error {
	typeID := ua.ServiceTypeID(req)
	if typeID == 0 {
		return errors.Errorf("unknown service %T. Did you call register?", req)
	}

	// encode the message
	m := s.newMessage(req, typeID)
	reqid := m.SequenceHeader.RequestID
	b, err := m.Encode()
	if err != nil {
		return err
	}

	// encrypt the message prior to sending it
	// if SecurityMode == None, this returns the byte stream untouched
	b, err = s.signAndEncrypt(m, b)
	if err != nil {
		return err
	}

	// send the message
	if _, err := s.c.Write(b); err != nil {
		return err
	}
	debug.Printf("uasc %d/%d: send %T with %d bytes", s.c.ID(), reqid, req, len(b))

	return nil
}

func (s *SecureChannel) readChunk() (*MessageChunk, error) {
	// read a full message from the underlying conn.
	b, err := s.c.Receive()
	if err == io.EOF || s.hasState(secureChannelClosed) {
		return nil, io.EOF
	}
	if errf, ok := err.(*uacp.Error); ok {
		return nil, errf
	}
	if err != nil {
		return nil, errors.Errorf("sechan: read header failed: %s %#v", err, err)
	}

	const hdrlen = 12
	h := new(Header)
	if _, err := h.Decode(b[:hdrlen]); err != nil {
		return nil, errors.Errorf("sechan: decode header failed: %s", err)
	}

	// decode the other headers
	m := new(MessageChunk)
	if _, err := m.Decode(b); err != nil {
		return nil, errors.Errorf("sechan: decode chunk failed: %s", err)
	}

	// OPN Request, initialize encryption
	// todo(dh): How to account for renew requests?
	switch m.MessageType {
	case "OPN":
		debug.Printf("uasc: OPN Request")
		// Make sure we have a valid security header
		if m.AsymmetricSecurityHeader == nil {
			return nil, ua.StatusBadDecodingError // todo(dh): check if this is the correct error
		}

		// Load the remote certificates from the security header, if present
		var remoteKey *rsa.PublicKey
		if m.SecurityPolicyURI != ua.SecurityPolicyURINone {
			remoteKey, err = uapolicy.PublicKey(m.AsymmetricSecurityHeader.SenderCertificate)
			if err != nil {
				return nil, err
			}

			s.cfg.RemoteCertificate = m.AsymmetricSecurityHeader.SenderCertificate
			debug.Printf("Setting securityPolicy to %s", m.SecurityPolicyURI)
		}

		s.cfg.SecurityPolicyURI = m.SecurityPolicyURI
		s.requestID = m.RequestID

		s.enc, err = uapolicy.Asymmetric(m.SecurityPolicyURI, s.cfg.LocalKey, remoteKey)
		if err != nil {
			return nil, err
		}

	case "CLO":
		if !s.hasState(secureChannelOpen) {
			return nil, ua.StatusBadSecureChannelIDInvalid
		}

		// We received the close request so no response is necessary.
		// Returning io.EOF signals to the calling methods that the channel is to be shut down
		s.setState(secureChannelClosed)

		return nil, io.EOF

	case "MSG":
	}

	// Decrypts the block and returns data back into m.Data
	m.Data, err = s.verifyAndDecrypt(m, b)
	if err != nil {
		return nil, err
	}

	n, err := m.SequenceHeader.Decode(m.Data)
	if err != nil {
		return nil, errors.Errorf("sechan: decode sequence header failed: %s", err)
	}
	m.Data = m.Data[n:]

	if s.secureChannelID == 0 {
		s.secureChannelID = h.SecureChannelID
		debug.Printf("uasc %d/%d: set secure channel id to %d", s.c.ID(), m.SequenceHeader.RequestID, s.secureChannelID)
	}

	return m, nil
}

// Receive waits for a complete message to be read from the channel and sends
// it back to the caller. If the caller was initiated from a SendRequest(), the
// message is directed to the registered callback function and Receive() does
// not return. Otherwise, if no handler is detected, the Receive returns with
// the message as a return value.
//
// This behaviour means that anticipated results are automatically directed
// back to their callers but unsolicited messages are sent to the caller of
// Receive() to handle.
func (s *SecureChannel) Receive(ctx context.Context) Response {
	for {
		select {
		case <-ctx.Done():
			return Response{Err: io.EOF}
		default:
			reqid, svc, err := s.receive(ctx)
			if _, ok := err.(*uacp.Error); ok || err == io.EOF {
				s.notifyCallers(ctx, err)
				return Response{
					ReqID: reqid,
					SCID:  s.secureChannelID,
					V:     svc,
					Err:   err,
				}
			}
			if err != nil {
				debug.Printf("uasc %d/%d: err: %v", s.c.ID(), reqid, err)
			} else {
				debug.Printf("uasc %d/%d: recv %T", s.c.ID(), reqid, svc)
			}

			// Revert data race fix from #232 with an additional type check
			if _, ok := svc.(ua.Request); ok {
				s.requestID = reqid
			}

			switch svc.(type) {
			case *ua.OpenSecureChannelRequest:
				err := s.handleOpenSecureChannelRequest(svc)
				if err != nil {
					return Response{
						Err: err,
					}
				}
				continue
			}

			// check if we have a pending request handler for this response.
			s.mu.Lock()
			ch, ok := s.handler[reqid]
			delete(s.handler, reqid)
			s.mu.Unlock()
			if !ok {
				debug.Printf("uasc %d/%d: no handler for %T, returning result to caller", s.c.ID(), reqid, svc)
				return Response{
					ReqID: reqid,
					SCID:  s.secureChannelID,
					V:     svc,
					Err:   err,
				}
			}

			// send response to caller
			go func() {
				debug.Printf("sending %T to handler\n", svc)
				r := Response{
					ReqID: reqid,
					SCID:  s.secureChannelID,
					V:     svc,
					Err:   err,
				}
				select {
				case <-ctx.Done():
				case ch <- r:
				}
			}()
		}
	}
}

// receive receives message chunks from the secure channel, decodes and forwards
// them to the registered callback channel, if there is one. Otherwise,
// the message is dropped.
func (s *SecureChannel) receive(ctx context.Context) (uint32, interface{}, error) {
	for {
		select {
		case <-ctx.Done():
			return 0, nil, nil

		default:
			chunk, err := s.readChunk()
			if err == io.EOF {
				return 0, nil, err
			}
			if errf, ok := err.(*uacp.Error); ok {
				s.notifyCallers(ctx, errf)
				return 0, nil, errf
			}
			if err != nil {
				debug.Printf("error received while receiving chunk: %s", err)
				continue
			}

			hdr := chunk.Header
			reqid := chunk.SequenceHeader.RequestID
			debug.Printf("uasc %d/%d: recv %s%c with %d bytes", s.c.ID(), reqid, hdr.MessageType, hdr.ChunkType, hdr.MessageSize)

			switch hdr.ChunkType {
			case 'A':
				delete(s.chunks, reqid)

				msga := new(MessageAbort)
				if _, err := msga.Decode(chunk.Data); err != nil {
					debug.Printf("conn %d/%d: invalid MSGA chunk. %s", s.c.ID(), reqid, err)
					return reqid, nil, ua.StatusBadDecodingError
				}

				return reqid, nil, ua.StatusCode(msga.ErrorCode)

			case 'C':
				s.chunks[reqid] = append(s.chunks[reqid], chunk)
				if n := len(s.chunks[reqid]); uint32(n) > s.c.MaxChunkCount() {
					delete(s.chunks, reqid)
					return reqid, nil, errors.Errorf("too many chunks: %d > %d", n, s.c.MaxChunkCount())
				}
				continue
			}

			// merge chunks
			all := append(s.chunks[reqid], chunk)
			delete(s.chunks, reqid)
			b, err := mergeChunks(all)
			if err != nil {
				return reqid, nil, errors.Errorf("chunk merge error: %v", err)
			}

			if uint32(len(b)) > s.c.MaxMessageSize() {
				return reqid, nil, errors.Errorf("message too large: %d > %d", uint32(len(b)), s.c.MaxMessageSize())
			}

			// since we are not decoding the ResponseHeader separately
			// we need to drop every message that has an error since we
			// cannot get to the RequestHandle in the ResponseHeader.
			// To fix this we must a) decode the ResponseHeader separately
			// and subsequently remove it and the TypeID from all service
			// structs and tests. We also need to add a deadline to all
			// handlers and check them periodically to time them out.
			_, svc, err := ua.DecodeService(b)
			if err != nil {
				return reqid, nil, err
			}

			// If the service status is not OK then bubble
			// that error up to the caller.
			if resp, ok := svc.(ua.Response); ok {
				status := resp.Header().ServiceResult
				debug.Printf("uasc %d/%d: res:%v", s.c.ID(), reqid, status)
				if status != ua.StatusOK {
					return reqid, svc, status
				}
			}
			return reqid, svc, err
		}
	}
}

func (s *SecureChannel) notifyCallers(ctx context.Context, err error) {
	s.mu.Lock()
	var reqids []uint32
	for rid := range s.handler {
		reqids = append(reqids, rid)
	}
	for _, rid := range reqids {
		s.notifyCallerLock(ctx, rid, nil, err)
	}
	s.mu.Unlock()
}

func (s *SecureChannel) notifyCallerLock(ctx context.Context, reqid uint32, svc interface{}, err error) {
	if err != nil {
		debug.Printf("uasc %d/%d: %v", s.c.ID(), reqid, err)
	} else {
		debug.Printf("uasc %d/%d: recv %T", s.c.ID(), reqid, svc)
	}

	// check if we have a pending request handler for this response.
	ch := s.popHandlerLock(reqid)

	// no handler -> next response
	if ch == nil {
		debug.Printf("uasc %d/%d: no handler for %T", s.c.ID(), reqid, svc)
		return
	}

	// send response to caller
	go func() {
		r := Response{
			ReqID: reqid,
			SCID:  s.secureChannelID,
			V:     svc,
			Err:   err,
		}
		select {
		case <-ctx.Done():
		case ch <- r:
		}
		close(ch)
	}()
}

// Open opens a new secure channel with a server
func (s *SecureChannel) Open() error {
	return s.openSecureChannel(ua.SecurityTokenRequestTypeIssue)
}

func (s *SecureChannel) Renew() error {
	return s.openSecureChannel(ua.SecurityTokenRequestTypeRenew)
}

// Close closes an existing secure channel
func (s *SecureChannel) Close() error {
	if err := s.closeSecureChannel(); err != nil && err != io.EOF {
		debug.Printf("failed to send close secure channel request: %s", err)
	}

	if err := s.c.Close(); err != nil && err != io.EOF {
		debug.Printf("failed to close transport connection: %s", err)
	}

	return io.EOF
}

func (s *SecureChannel) openSecureChannel(requestType ua.SecurityTokenRequestType) error {
	var err error
	var localKey *rsa.PrivateKey
	var remoteKey *rsa.PublicKey

	// Set the encryption methods to Asymmetric with the appropriate
	// public keys.  OpenSecureChannel is always encrypted with the
	// asymmetric algorithms.
	// The default value of the encryption algorithm method is the
	// SecurityModeNone so no additional work is required for that case
	if s.cfg.SecurityMode != ua.MessageSecurityModeNone {
		localKey = s.cfg.LocalKey
		// todo(dh): move this into the uapolicy package proper or
		// adjust the Asymmetric method to receive a certificate instead
		remoteCert, err := x509.ParseCertificate(s.cfg.RemoteCertificate)
		if err != nil {
			return err
		}
		var ok bool
		remoteKey, ok = remoteCert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return ua.StatusBadCertificateInvalid
		}
	}

	s.enc, err = uapolicy.Asymmetric(s.cfg.SecurityPolicyURI, localKey, remoteKey)
	if err != nil {
		return err
	}

	nonce := make([]byte, s.enc.NonceLength())
	if _, err := rand.Read(nonce); err != nil {
		return err
	}

	req := &ua.OpenSecureChannelRequest{
		ClientProtocolVersion: 0,
		RequestType:           requestType,
		SecurityMode:          s.cfg.SecurityMode,
		ClientNonce:           nonce,
		RequestedLifetime:     s.cfg.Lifetime,
	}

	return s.SendRequest(req, nil, func(v interface{}) error {
		resp, ok := v.(*ua.OpenSecureChannelResponse)
		if !ok {
			return errors.Errorf("got %T, want OpenSecureChannelResponse", req)
		}
		s.securityTokenID = resp.SecurityToken.TokenID
		s.lifetime = resp.SecurityToken.RevisedLifetime
		debug.Printf("received security token tokenID: %v, createdAt: %v, lifetime %v", resp.SecurityToken.TokenID, resp.SecurityToken.CreatedAt, resp.SecurityToken.RevisedLifetime)

		s.enc, err = uapolicy.Symmetric(s.cfg.SecurityPolicyURI, nonce, resp.ServerNonce)
		if err != nil {
			return err
		}

		s.setState(secureChannelOpen)
		return nil
	})
}

// closeSecureChannel sends CloseSecureChannelRequest on top of UASC to SecureChannel.
func (s *SecureChannel) closeSecureChannel() error {
	req := &ua.CloseSecureChannelRequest{}

	defer s.setState(secureChannelClosed)
	// Don't send the CloseSecureChannel message if it was never fully opened (due to ERR, etc)
	if !s.hasState(secureChannelOpen) {
		return io.EOF
	}

	err := s.SendRequest(req, nil, nil)
	if err != nil {
		return err
	}

	return io.EOF
}

func (s *SecureChannel) handleOpenSecureChannelRequest(svc interface{}) error {
	debug.Printf("handleOpenSecureChannelRequest: Got OPN Request\n")

	var err error

	req, ok := svc.(*ua.OpenSecureChannelRequest)
	if !ok {
		debug.Printf("Expected OpenSecureChannel Request, got %T\n", svc)
	}

	s.cfg.Lifetime = req.RequestedLifetime
	s.cfg.SecurityMode = req.SecurityMode

	nonce := make([]byte, s.enc.NonceLength())
	if _, err := rand.Read(nonce); err != nil {
		return err
	}
	resp := &ua.OpenSecureChannelResponse{
		ResponseHeader: &ua.ResponseHeader{
			Timestamp:          s.timeNow(),
			RequestHandle:      req.RequestHeader.RequestHandle,
			ServiceDiagnostics: &ua.DiagnosticInfo{},
			StringTable:        []string{},
			AdditionalHeader:   ua.NewExtensionObject(nil),
		},
		ServerProtocolVersion: 0,
		SecurityToken: &ua.ChannelSecurityToken{
			ChannelID:       s.secureChannelID,
			TokenID:         s.securityTokenID,
			CreatedAt:       s.timeNow(),
			RevisedLifetime: req.RequestedLifetime,
		},
		ServerNonce: nonce,
	}

	if err := s.SendResponse(resp); err != nil {
		return err
	}

	s.enc, err = uapolicy.Symmetric(s.cfg.SecurityPolicyURI, nonce, req.ClientNonce)
	if err != nil {
		return err
	}
	s.setState(secureChannelOpen)

	return nil
}

func (s *SecureChannel) popHandlerLock(reqid uint32) chan Response {
	ch := s.handler[reqid]
	delete(s.handler, reqid)
	return ch
}

func (s *SecureChannel) timeNow() time.Time {
	if s.time != nil {
		return s.time()
	}
	return time.Now()
}

func mergeChunks(chunks []*MessageChunk) ([]byte, error) {
	if len(chunks) == 0 {
		return nil, nil
	}
	if len(chunks) == 1 {
		return chunks[0].Data, nil
	}

	// todo(fs): check if this is correct and necessary
	// sort.Sort(bySequence(chunks))

	var b []byte
	var seqnr uint32
	for _, c := range chunks {
		if c.SequenceHeader.SequenceNumber == seqnr {
			continue // duplicate chunk
		}
		seqnr = c.SequenceHeader.SequenceNumber
		b = append(b, c.Data...)
	}
	return b, nil
}

// todo(fs): we only need this if we need to sort chunks. Need to check the spec
// type bySequence []*MessageChunk

// func (a bySequence) Len() int      { return len(a) }
// func (a bySequence) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
// func (a bySequence) Less(i, j int) bool {
// 	return a[i].SequenceHeader.SequenceNumber < a[j].SequenceHeader.SequenceNumber
// }
