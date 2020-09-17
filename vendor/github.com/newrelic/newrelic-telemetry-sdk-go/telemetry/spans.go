// Copyright 2019 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/newrelic/newrelic-telemetry-sdk-go/internal"
)

// Span is a distributed tracing span.
type Span struct {
	// Required Fields:
	//
	// ID is a unique identifier for this span.
	ID string
	// TraceID is a unique identifier shared by all spans within a single
	// trace.
	TraceID string
	// Timestamp is when this span started.  If Timestamp is not set, it
	// will be assigned to time.Now() in Harvester.RecordSpan.
	Timestamp time.Time

	// Recommended Fields:
	//
	// Name is the name of this span.
	Name string
	// ParentID is the span id of the previous caller of this span.  This
	// can be empty if this is the first span.
	ParentID string
	// Duration is the duration of this span.  This field will be reported
	// in milliseconds.
	Duration time.Duration
	// ServiceName is the name of the service that created this span.
	ServiceName string

	// Additional Fields:
	//
	// Attributes is a map of user specified tags on this span.  The map
	// values can be any of bool, number, or string.
	Attributes map[string]interface{}
	// AttributesJSON is a json.RawMessage of attributes for this metric. It
	// will only be sent if Attributes is nil.
	AttributesJSON json.RawMessage
}

func (s *Span) writeJSON(buf *bytes.Buffer) {
	w := internal.JSONFieldsWriter{Buf: buf}
	buf.WriteByte('{')

	w.StringField("id", s.ID)
	w.StringField("trace.id", s.TraceID)
	w.IntField("timestamp", s.Timestamp.UnixNano()/(1000*1000))

	w.AddKey("attributes")
	buf.WriteByte('{')
	ww := internal.JSONFieldsWriter{Buf: buf}

	if "" != s.Name {
		ww.StringField("name", s.Name)
	}
	if "" != s.ParentID {
		ww.StringField("parent.id", s.ParentID)
	}
	if 0 != s.Duration {
		ww.FloatField("duration.ms", s.Duration.Seconds()*1000.0)
	}
	if "" != s.ServiceName {
		ww.StringField("service.name", s.ServiceName)
	}

	internal.AddAttributes(&ww, s.Attributes)

	buf.WriteByte('}')
	buf.WriteByte('}')
}

// spanBatch represents a single batch of spans to report to New Relic.
type spanBatch struct {
	// AttributesJSON is a json.RawMessage of attributes to apply to all
	// spans in this spanBatch. It will only be sent if the Attributes field
	// on this spanBatch is nil. These attributes are included in addition
	// to any attributes on any particular span.
	AttributesJSON json.RawMessage
	Spans          []Span
}

// split will split the spanBatch into 2 equally sized batches.
// If the number of spans in the original is 0 or 1 then nil is returned.
func (batch *spanBatch) split() []requestsBuilder {
	if len(batch.Spans) < 2 {
		return nil
	}

	half := len(batch.Spans) / 2
	b1 := *batch
	b1.Spans = batch.Spans[:half]
	b2 := *batch
	b2.Spans = batch.Spans[half:]

	return []requestsBuilder{
		requestsBuilder(&b1),
		requestsBuilder(&b2),
	}
}

func (batch *spanBatch) writeJSON(buf *bytes.Buffer) {
	buf.WriteByte('[')
	buf.WriteByte('{')
	w := internal.JSONFieldsWriter{Buf: buf}

	w.AddKey("common")
	buf.WriteByte('{')
	ww := internal.JSONFieldsWriter{Buf: buf}
	if nil != batch.AttributesJSON {
		ww.RawField("attributes", batch.AttributesJSON)
	}
	buf.WriteByte('}')

	w.AddKey("spans")
	buf.WriteByte('[')
	for idx, s := range batch.Spans {
		if idx > 0 {
			buf.WriteByte(',')
		}
		s.writeJSON(buf)
	}
	buf.WriteByte(']')
	buf.WriteByte('}')
	buf.WriteByte(']')
}

func (batch *spanBatch) makeBody() json.RawMessage {
	buf := &bytes.Buffer{}
	batch.writeJSON(buf)
	return buf.Bytes()
}
