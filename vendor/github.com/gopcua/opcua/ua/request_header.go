// Copyright 2018-2020 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package ua

// These flags define the options for the ReturnDiagnostics
// field of the RequestHeader.
// Bits are or'ed together if multiple fields are set.
const (
	ServiceLevelSymbolicID         = 0x1
	ServiceLevelLocalizedText      = 0x2
	ServiceLevelAdditionalInfo     = 0x4
	ServiceLevelInnerStatusCode    = 0x8
	ServiceLevelInnerDiagnostics   = 0x10
	OperationLevelSymbolicID       = 0x20
	OperationLevelLocalizedText    = 0x40
	OperationLevelAdditionalInfo   = 0x80
	OperationLevelInnerStatusCode  = 0x100
	OperationLevelInnerDiagnostics = 0x200

	ServiceLevelAll = ServiceLevelSymbolicID |
		ServiceLevelLocalizedText |
		ServiceLevelAdditionalInfo |
		ServiceLevelInnerStatusCode |
		ServiceLevelInnerDiagnostics

	OperationLevelAll = OperationLevelSymbolicID |
		OperationLevelLocalizedText |
		OperationLevelAdditionalInfo |
		OperationLevelInnerStatusCode |
		OperationLevelInnerDiagnostics

	ReturnDiagnosticsAll = ServiceLevelAll | OperationLevelAll
)

func (r *RequestHeader) HasReturnDiagnostics(mask uint32) bool {
	return r.ReturnDiagnostics&mask == mask
}
