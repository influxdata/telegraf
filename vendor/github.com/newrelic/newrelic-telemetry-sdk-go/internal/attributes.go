// Copyright 2019 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// MarshalAttributes turns attributes into JSON.
func MarshalAttributes(ats map[string]interface{}) []byte {
	attrs := Attributes(ats)
	buf := &bytes.Buffer{}
	attrs.WriteJSON(buf)
	return buf.Bytes()
}

// Attributes is used for marshalling attributes to JSON.
type Attributes map[string]interface{}

// WriteJSON writes the attributes in JSON.
func (attrs Attributes) WriteJSON(buf *bytes.Buffer) {
	w := JSONFieldsWriter{Buf: buf}
	w.Buf.WriteByte('{')
	AddAttributes(&w, attrs)
	w.Buf.WriteByte('}')
}

// AddAttributes writes the attributes to the fields writer.
func AddAttributes(w *JSONFieldsWriter, attrs map[string]interface{}) {
	for key, val := range attrs {
		writeAttribute(w, key, val)
	}
}

// MarshalOrderedAttributes marshals the given attributes into JSON in
// alphabetical order.
func MarshalOrderedAttributes(attrs map[string]interface{}) []byte {
	buf := &bytes.Buffer{}
	OrderedAttributes(attrs).WriteJSON(buf)
	return buf.Bytes()
}

// OrderedAttributes turns attributes into JSON in a fixed order.
type OrderedAttributes map[string]interface{}

// WriteJSON writes the attributes in JSON in a fixed order.
func (attrs OrderedAttributes) WriteJSON(buf *bytes.Buffer) {
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	w := JSONFieldsWriter{Buf: buf}
	w.Buf.WriteByte('{')
	for _, k := range keys {
		writeAttribute(&w, k, attrs[k])
	}
	w.Buf.WriteByte('}')
}

func writeAttribute(w *JSONFieldsWriter, key string, val interface{}) {
	switch v := val.(type) {
	case string:
		w.StringField(key, v)
	case bool:
		if v {
			w.RawField(key, json.RawMessage(`true`))
		} else {
			w.RawField(key, json.RawMessage(`false`))
		}
	case uint8:
		w.IntField(key, int64(v))
	case uint16:
		w.IntField(key, int64(v))
	case uint32:
		w.IntField(key, int64(v))
	case uint64:
		w.IntField(key, int64(v))
	case uint:
		w.IntField(key, int64(v))
	case uintptr:
		w.IntField(key, int64(v))
	case int8:
		w.IntField(key, int64(v))
	case int16:
		w.IntField(key, int64(v))
	case int32:
		w.IntField(key, int64(v))
	case int64:
		w.IntField(key, v)
	case int:
		w.IntField(key, int64(v))
	case float32:
		w.FloatField(key, float64(v))
	case float64:
		w.FloatField(key, v)
	case nil:
		// nil gets dropped.
	default:
		w.StringField(key, fmt.Sprintf("%T", v))
	}
}
