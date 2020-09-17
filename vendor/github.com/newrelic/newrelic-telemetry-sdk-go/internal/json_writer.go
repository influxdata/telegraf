// Copyright 2019 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"bytes"
	"encoding/json"

	"github.com/newrelic/newrelic-telemetry-sdk-go/internal/jsonx"
)

// JSONWriter is something that can write JSON to a buffer.
type JSONWriter interface {
	WriteJSON(buf *bytes.Buffer)
}

// JSONFieldsWriter helps write JSON objects to a buffer.
type JSONFieldsWriter struct {
	Buf        *bytes.Buffer
	needsComma bool
}

// AddKey adds the key for a new object field.  A comma is prefixed if another
// field has previously been added.
func (w *JSONFieldsWriter) AddKey(key string) {
	if w.needsComma {
		w.Buf.WriteByte(',')
	} else {
		w.needsComma = true
	}
	// defensively assume that the key needs escaping:
	jsonx.AppendString(w.Buf, key)
	w.Buf.WriteByte(':')
}

// StringField adds a string field to the object.
func (w *JSONFieldsWriter) StringField(key string, val string) {
	w.AddKey(key)
	jsonx.AppendString(w.Buf, val)
}

// IntField adds an int field to the object.
func (w *JSONFieldsWriter) IntField(key string, val int64) {
	w.AddKey(key)
	jsonx.AppendInt(w.Buf, val)
}

// FloatField adds a float field to the object.
func (w *JSONFieldsWriter) FloatField(key string, val float64) {
	w.AddKey(key)
	jsonx.AppendFloat(w.Buf, val)
}

// BoolField adds a bool field to the object.
func (w *JSONFieldsWriter) BoolField(key string, val bool) {
	w.AddKey(key)
	if val {
		w.Buf.WriteString("true")
	} else {
		w.Buf.WriteString("false")
	}
}

// RawField adds a raw JSON field to the object.
func (w *JSONFieldsWriter) RawField(key string, val json.RawMessage) {
	w.AddKey(key)
	w.Buf.Write(val)
}

// WriterField adds a JSONWriter field to the object.
func (w *JSONFieldsWriter) WriterField(key string, val JSONWriter) {
	w.AddKey(key)
	val.WriteJSON(w.Buf)
}
