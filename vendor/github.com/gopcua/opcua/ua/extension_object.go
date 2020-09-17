// Copyright 2018-2020 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package ua

import (
	"github.com/gopcua/opcua/errors"
	"github.com/gopcua/opcua/id"
)

// eotypes contains all known extension objects.
var eotypes = NewTypeRegistry()

// RegisterExtensionObject registers a new extension object type.
// It panics if the type or the id is already registered.
func RegisterExtensionObject(typeID *NodeID, v interface{}) {
	if err := eotypes.Register(typeID.String(), v); err != nil {
		panic("Extension object " + err.Error())
	}
}

// These flags define the value type of an ExtensionObject.
// They cannot be combined.
const (
	ExtensionObjectEmpty  = 0
	ExtensionObjectBinary = 1
	ExtensionObjectXML    = 2
)

// ExtensionObject is encoded as sequence of bytes prefixed by the NodeId of its DataTypeEncoding
// and the number of bytes encoded.
//
// Specification: Part 6, 5.2.2.15
type ExtensionObject struct {
	EncodingMask uint8
	TypeID       *ExpandedNodeID
	Value        interface{}
}

func NewExtensionObject(value interface{}) *ExtensionObject {
	e := &ExtensionObject{
		TypeID: ExtensionObjectTypeID(value),
		Value:  value,
	}
	e.UpdateMask()
	return e
}

func (e *ExtensionObject) Decode(b []byte) (int, error) {
	buf := NewBuffer(b)
	e.TypeID = new(ExpandedNodeID)
	buf.ReadStruct(e.TypeID)

	e.EncodingMask = buf.ReadByte()
	if e.EncodingMask == ExtensionObjectEmpty {
		return buf.Pos(), buf.Error()
	}

	length := buf.ReadUint32()
	if length == 0 || length == 0xffffffff || buf.Error() != nil {
		return buf.Pos(), buf.Error()
	}

	body := NewBuffer(buf.ReadN(int(length)))
	if buf.Error() != nil {
		return buf.Pos(), buf.Error()
	}

	if e.EncodingMask == ExtensionObjectXML {
		e.Value = new(XMLElement)
		body.ReadStruct(e.Value)
		return buf.Pos(), body.Error()
	}

	typeID := e.TypeID.NodeID.String()
	e.Value = eotypes.New(typeID)
	if e.Value == nil {
		return buf.Pos(), errors.Errorf("invalid extension object with id %s", typeID)
	}

	body.ReadStruct(e.Value)
	return buf.Pos(), body.Error()
}

func (e *ExtensionObject) Encode() ([]byte, error) {
	buf := NewBuffer(nil)
	if e == nil {
		e = &ExtensionObject{TypeID: NewTwoByteExpandedNodeID(0), EncodingMask: ExtensionObjectEmpty}
	}
	buf.WriteStruct(e.TypeID)
	buf.WriteByte(e.EncodingMask)
	if e.EncodingMask == ExtensionObjectEmpty {
		return buf.Bytes(), buf.Error()
	}

	body := NewBuffer(nil)
	body.WriteStruct(e.Value)
	if body.Error() != nil {
		return nil, body.Error()
	}
	buf.WriteUint32(uint32(body.Len()))
	buf.Write(body.Bytes())
	return buf.Bytes(), buf.Error()
}

func (e *ExtensionObject) UpdateMask() {
	if e.Value == nil {
		e.EncodingMask = ExtensionObjectEmpty
	} else if _, ok := e.Value.(*XMLElement); ok {
		e.EncodingMask = ExtensionObjectXML
	} else {
		e.EncodingMask = ExtensionObjectBinary
	}
}

func ExtensionObjectTypeID(v interface{}) *ExpandedNodeID {
	switch v.(type) {
	case *AnonymousIdentityToken:
		return NewFourByteExpandedNodeID(0, id.AnonymousIdentityToken_Encoding_DefaultBinary)
	case *UserNameIdentityToken:
		return NewFourByteExpandedNodeID(0, id.UserNameIdentityToken_Encoding_DefaultBinary)
	case *X509IdentityToken:
		return NewFourByteExpandedNodeID(0, id.X509IdentityToken_Encoding_DefaultBinary)
	case *IssuedIdentityToken:
		return NewFourByteExpandedNodeID(0, id.IssuedIdentityToken_Encoding_DefaultBinary)
	case *ServerStatusDataType:
		return NewFourByteExpandedNodeID(0, id.ServerStatusDataType_Encoding_DefaultBinary)
	default:
		return NewTwoByteExpandedNodeID(0)
	}
}
