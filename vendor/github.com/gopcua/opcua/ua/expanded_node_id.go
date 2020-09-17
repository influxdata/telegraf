// Copyright 2018-2020 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package ua

// ExpandedNodeID extends the NodeID structure by allowing the NamespaceURI to be
// explicitly specified instead of using the NamespaceIndex. The NamespaceURI is optional.
// If it is specified, then the NamespaceIndex inside the NodeID shall be ignored.
//
// Specification: Part 6, 5.2.2.10
type ExpandedNodeID struct {
	NodeID       *NodeID
	NamespaceURI string
	ServerIndex  uint32
}

func (a ExpandedNodeID) String() string {
	return a.NodeID.String()
}

// NewExpandedNodeID creates a new ExpandedNodeID.
func NewExpandedNodeID(hasURI, hasIndex bool, nodeID *NodeID, uri string, idx uint32) *ExpandedNodeID {
	e := &ExpandedNodeID{
		NodeID:      nodeID,
		ServerIndex: idx,
	}

	if hasURI {
		e.NodeID.SetURIFlag()
		e.NamespaceURI = uri
	}
	if hasIndex {
		e.NodeID.SetIndexFlag()
	}

	return e
}

// NewTwoByteExpandedNodeID creates a two byte numeric expanded node id.
func NewTwoByteExpandedNodeID(id uint8) *ExpandedNodeID {
	return &ExpandedNodeID{
		NodeID: NewTwoByteNodeID(id),
	}
}

// NewFourByteExpandedNodeID creates a four byte numeric expanded node id.
func NewFourByteExpandedNodeID(ns uint8, id uint16) *ExpandedNodeID {
	return &ExpandedNodeID{
		NodeID: NewFourByteNodeID(ns, id),
	}
}

func (e *ExpandedNodeID) Decode(b []byte) (int, error) {
	buf := NewBuffer(b)
	e.NodeID = new(NodeID)
	buf.ReadStruct(e.NodeID)
	if e.HasNamespaceURI() {
		e.NamespaceURI = buf.ReadString()
	}
	if e.HasServerIndex() {
		e.ServerIndex = buf.ReadUint32()
	}
	return buf.Pos(), buf.Error()
}

func (e *ExpandedNodeID) Encode() ([]byte, error) {
	buf := NewBuffer(nil)
	buf.WriteStruct(e.NodeID)
	if e.HasNamespaceURI() {
		buf.WriteString(e.NamespaceURI)
	}
	if e.HasServerIndex() {
		buf.WriteUint32(e.ServerIndex)
	}
	return buf.Bytes(), buf.Error()

}

// HasNamespaceURI checks if an ExpandedNodeID has NamespaceURI Flag.
func (e *ExpandedNodeID) HasNamespaceURI() bool {
	return e.NodeID.EncodingMask()>>7&0x1 == 1
}

// HasServerIndex checks if an ExpandedNodeID has ServerIndex Flag.
func (e *ExpandedNodeID) HasServerIndex() bool {
	return e.NodeID.EncodingMask()>>6&0x1 == 1
}
