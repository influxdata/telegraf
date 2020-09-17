// Copyright 2018-2020 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package uacp

import (
	"github.com/gopcua/opcua/errors"
	"github.com/gopcua/opcua/ua"
)

// MessageType definitions.
//
// Specification: Part 6, 7.1.2.2
const (
	MessageTypeHello        = "HEL"
	MessageTypeAcknowledge  = "ACK"
	MessageTypeError        = "ERR"
	MessageTypeReverseHello = "RHE"
)

// ChunkType definitions.
//
// Specification: Part 6, 6.7.2.2
const (
	ChunkTypeIntermediate = 'C'
	ChunkTypeFinal        = 'F'
	ChunkTypeAbort        = 'A'
)

// Header represents a OPC UA Connection Header.
//
// Specification: Part 6, 7.1.2.2
type Header struct {
	MessageType string
	ChunkType   byte
	MessageSize uint32
}

func (h *Header) Decode(b []byte) (int, error) {
	buf := ua.NewBuffer(b)
	h.MessageType = string(buf.ReadN(3))
	h.ChunkType = buf.ReadByte()
	h.MessageSize = buf.ReadUint32()
	return buf.Pos(), buf.Error()
}

func (h *Header) Encode() ([]byte, error) {
	buf := ua.NewBuffer(nil)
	if len(h.MessageType) != 3 {
		return nil, errors.Errorf("invalid message type: %q", h.MessageType)
	}
	buf.Write([]byte(h.MessageType))
	buf.WriteByte(h.ChunkType)
	buf.WriteUint32(h.MessageSize)
	return buf.Bytes(), buf.Error()
}

// Hello represents a OPC UA Hello.
//
// Specification: Part6, 7.1.2.3
type Hello struct {
	Version        uint32
	ReceiveBufSize uint32
	SendBufSize    uint32
	MaxMessageSize uint32
	MaxChunkCount  uint32
	EndpointURL    string
}

func (h *Hello) Decode(b []byte) (int, error) {
	buf := ua.NewBuffer(b)
	h.Version = buf.ReadUint32()
	h.ReceiveBufSize = buf.ReadUint32()
	h.SendBufSize = buf.ReadUint32()
	h.MaxMessageSize = buf.ReadUint32()
	h.MaxChunkCount = buf.ReadUint32()
	h.EndpointURL = buf.ReadString()
	return buf.Pos(), buf.Error()
}

func (h *Hello) Encode() ([]byte, error) {
	buf := ua.NewBuffer(nil)
	buf.WriteUint32(h.Version)
	buf.WriteUint32(h.ReceiveBufSize)
	buf.WriteUint32(h.SendBufSize)
	buf.WriteUint32(h.MaxMessageSize)
	buf.WriteUint32(h.MaxChunkCount)
	buf.WriteString(h.EndpointURL)
	return buf.Bytes(), buf.Error()
}

// Acknowledge represents a OPC UA Acknowledge.
//
// Specification: Part6, 7.1.2.4
type Acknowledge struct {
	Version        uint32
	ReceiveBufSize uint32
	SendBufSize    uint32
	MaxMessageSize uint32
	MaxChunkCount  uint32
}

func (a *Acknowledge) Decode(b []byte) (int, error) {
	buf := ua.NewBuffer(b)
	a.Version = buf.ReadUint32()
	a.ReceiveBufSize = buf.ReadUint32()
	a.SendBufSize = buf.ReadUint32()
	a.MaxMessageSize = buf.ReadUint32()
	a.MaxChunkCount = buf.ReadUint32()
	return buf.Pos(), buf.Error()
}

func (a *Acknowledge) Encode() ([]byte, error) {
	buf := ua.NewBuffer(nil)
	buf.WriteUint32(a.Version)
	buf.WriteUint32(a.ReceiveBufSize)
	buf.WriteUint32(a.SendBufSize)
	buf.WriteUint32(a.MaxMessageSize)
	buf.WriteUint32(a.MaxChunkCount)
	return buf.Bytes(), buf.Error()
}

// ReverseHello represents a OPC UA ReverseHello.
//
// Specification: Part6, 7.1.2.6
type ReverseHello struct {
	ServerURI   string
	EndpointURL string
}

func (r *ReverseHello) Decode(b []byte) (int, error) {
	buf := ua.NewBuffer(b)
	r.ServerURI = buf.ReadString()
	r.EndpointURL = buf.ReadString()
	return buf.Pos(), buf.Error()
}

func (r *ReverseHello) Encode() ([]byte, error) {
	buf := ua.NewBuffer(nil)
	buf.WriteString(r.ServerURI)
	buf.WriteString(r.EndpointURL)
	return buf.Bytes(), buf.Error()
}

// Error represents a OPC UA Error.
//
// Specification: Part6, 7.1.2.5
type Error struct {
	ErrorCode uint32
	Reason    string
}

func (e *Error) Decode(b []byte) (int, error) {
	buf := ua.NewBuffer(b)
	e.ErrorCode = buf.ReadUint32()
	e.Reason = buf.ReadString()
	return buf.Pos(), buf.Error()
}

func (e *Error) Encode() ([]byte, error) {
	buf := ua.NewBuffer(nil)
	buf.WriteUint32(e.ErrorCode)
	buf.WriteString(e.Reason)
	return buf.Bytes(), buf.Error()
}

func (e *Error) Error() string {
	return ua.StatusCode(e.ErrorCode).Error()
}

type Message struct {
	Data []byte
}

func (m *Message) Decode(b []byte) (int, error) {
	m.Data = b
	return len(b), nil
}

func (m *Message) Encode() ([]byte, error) {
	return m.Data, nil
}
