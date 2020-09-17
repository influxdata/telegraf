// Copyright 2018-2020 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package uasc

import (
	"github.com/gopcua/opcua/errors"
	"github.com/gopcua/opcua/ua"
)

type MessageHeader struct {
	*Header
	*AsymmetricSecurityHeader
	*SymmetricSecurityHeader
	*SequenceHeader
}

func (m *MessageHeader) Decode(b []byte) (int, error) {
	buf := ua.NewBuffer(b)

	m.Header = new(Header)
	buf.ReadStruct(m.Header)

	switch m.Header.MessageType {
	case "OPN":
		m.AsymmetricSecurityHeader = new(AsymmetricSecurityHeader)
		buf.ReadStruct(m.AsymmetricSecurityHeader)

	case "MSG", "CLO":
		m.SymmetricSecurityHeader = new(SymmetricSecurityHeader)
		buf.ReadStruct(m.SymmetricSecurityHeader)

	default:
		return buf.Pos(), errors.Errorf("invalid message type %q", m.Header.MessageType)
	}

	// Sequence header could be encrypted, defer decoding until after decryption
	m.SequenceHeader = new(SequenceHeader)
	//buf.ReadStruct(m.SequenceHeader)

	return buf.Pos(), buf.Error()
}

type MessageChunk struct {
	*MessageHeader
	Data []byte
}

func (m *MessageChunk) Decode(b []byte) (int, error) {
	m.MessageHeader = new(MessageHeader)
	n, err := m.MessageHeader.Decode(b)
	if err != nil {
		return n, err
	}
	m.Data = b[n:]
	return len(b), nil
}

// MessageAbort represents a non-terminal OPC UA Secure Channel error.
//
// Specification: Part6, 7.3
type MessageAbort struct {
	ErrorCode uint32
	Reason    string
}

func (m *MessageAbort) Decode(b []byte) (int, error) {
	buf := ua.NewBuffer(b)
	m.ErrorCode = buf.ReadUint32()
	m.Reason = buf.ReadString()
	return buf.Pos(), buf.Error()
}

func (m *MessageAbort) Encode() ([]byte, error) {
	buf := ua.NewBuffer(nil)
	buf.WriteUint32(m.ErrorCode)
	buf.WriteString(m.Reason)
	return buf.Bytes(), buf.Error()
}

func (m *MessageAbort) MessageAbort() string {
	return ua.StatusCode(m.ErrorCode).Error()
}

// Message represents a OPC UA Secure Conversation message.
type Message struct {
	*MessageHeader
	TypeID  *ua.ExpandedNodeID
	Service interface{}
}

func (m *Message) Decode(b []byte) (int, error) {
	m.MessageHeader = new(MessageHeader)
	var pos int
	n, err := m.MessageHeader.Decode(b)
	if err != nil {
		return n, err
	}
	pos += n

	m.SequenceHeader = new(SequenceHeader)
	n, err = m.SequenceHeader.Decode(b[pos:])
	if err != nil {
		return n, err
	}
	pos += n

	m.TypeID, m.Service, err = ua.DecodeService(b[pos:])
	return len(b), err
}

func (m *Message) Encode() ([]byte, error) {
	body := ua.NewBuffer(nil)
	switch m.Header.MessageType {
	case "OPN":
		body.WriteStruct(m.AsymmetricSecurityHeader)
	case "CLO", "MSG":
		body.WriteStruct(m.SymmetricSecurityHeader)
	default:
		return nil, errors.Errorf("invalid message type %q", m.Header.MessageType)
	}
	body.WriteStruct(m.SequenceHeader)
	body.WriteStruct(m.TypeID)
	body.WriteStruct(m.Service)
	if body.Error() != nil {
		return nil, body.Error()
	}

	m.Header.MessageSize = uint32(12 + body.Len())
	buf := ua.NewBuffer(nil)
	buf.WriteStruct(m.Header)
	buf.Write(body.Bytes())
	return buf.Bytes(), buf.Error()
}
