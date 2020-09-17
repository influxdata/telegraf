// Copyright 2018-2020 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package ua

// These flags define which fields of a DiagnosticInfo are set.
// Bits are or'ed together if multiple fields are set.
const (
	DiagnosticInfoSymbolicID          = 0x1
	DiagnosticInfoNamespaceURI        = 0x2
	DiagnosticInfoLocalizedText       = 0x4
	DiagnosticInfoLocale              = 0x8
	DiagnosticInfoAdditionalInfo      = 0x10
	DiagnosticInfoInnerStatusCode     = 0x20
	DiagnosticInfoInnerDiagnosticInfo = 0x40
)

// DiagnosticInfo represents the DiagnosticInfo.
//
// Specification: Part 4, 7.8
type DiagnosticInfo struct {
	EncodingMask        uint8
	SymbolicID          int32
	NamespaceURI        int32
	Locale              int32
	LocalizedText       int32
	AdditionalInfo      string
	InnerStatusCode     StatusCode
	InnerDiagnosticInfo *DiagnosticInfo
}

func (d *DiagnosticInfo) Decode(b []byte) (int, error) {
	buf := NewBuffer(b)
	d.EncodingMask = buf.ReadByte()
	if d.Has(DiagnosticInfoSymbolicID) {
		d.SymbolicID = buf.ReadInt32()
	}
	if d.Has(DiagnosticInfoNamespaceURI) {
		d.NamespaceURI = buf.ReadInt32()
	}
	if d.Has(DiagnosticInfoLocale) {
		d.Locale = buf.ReadInt32()
	}
	if d.Has(DiagnosticInfoLocalizedText) {
		d.LocalizedText = buf.ReadInt32()
	}
	if d.Has(DiagnosticInfoAdditionalInfo) {
		d.AdditionalInfo = buf.ReadString()
	}
	if d.Has(DiagnosticInfoInnerStatusCode) {
		d.InnerStatusCode = StatusCode(buf.ReadUint32())
	}
	if d.Has(DiagnosticInfoInnerDiagnosticInfo) {
		d.InnerDiagnosticInfo = new(DiagnosticInfo)
		buf.ReadStruct(d.InnerDiagnosticInfo)
	}
	return buf.Pos(), buf.Error()
}

func (d *DiagnosticInfo) Encode() ([]byte, error) {
	buf := NewBuffer(nil)
	buf.WriteByte(d.EncodingMask)
	if d.Has(DiagnosticInfoSymbolicID) {
		buf.WriteInt32(d.SymbolicID)
	}
	if d.Has(DiagnosticInfoNamespaceURI) {
		buf.WriteInt32(d.NamespaceURI)
	}
	if d.Has(DiagnosticInfoLocale) {
		buf.WriteInt32(d.Locale)
	}
	if d.Has(DiagnosticInfoLocalizedText) {
		buf.WriteInt32(d.LocalizedText)
	}
	if d.Has(DiagnosticInfoAdditionalInfo) {
		buf.WriteString(d.AdditionalInfo)
	}
	if d.Has(DiagnosticInfoInnerStatusCode) {
		buf.WriteUint32(uint32(d.InnerStatusCode))
	}
	if d.Has(DiagnosticInfoInnerDiagnosticInfo) {
		buf.WriteStruct(d.InnerDiagnosticInfo)
	}
	return buf.Bytes(), buf.Error()
}

func (d *DiagnosticInfo) Has(mask byte) bool {
	return d.EncodingMask&mask == mask
}

func (d *DiagnosticInfo) UpdateMask() {
	d.EncodingMask = 0
	if d.SymbolicID != 0 {
		d.EncodingMask |= DiagnosticInfoSymbolicID
	}
	if d.NamespaceURI != 0 {
		d.EncodingMask |= DiagnosticInfoNamespaceURI
	}
	if d.Locale != 0 {
		d.EncodingMask |= DiagnosticInfoLocale
	}
	if d.LocalizedText != 0 {
		d.EncodingMask |= DiagnosticInfoLocalizedText
	}
	if d.AdditionalInfo != "" {
		d.EncodingMask |= DiagnosticInfoAdditionalInfo
	}
	if d.InnerStatusCode != 0 {
		d.EncodingMask |= DiagnosticInfoInnerStatusCode
	}
	if d.InnerDiagnosticInfo != nil {
		d.EncodingMask |= DiagnosticInfoInnerDiagnosticInfo
	}
}
