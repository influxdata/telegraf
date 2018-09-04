//  Copyright (c) 2016 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the
//  License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing,
//  software distributed under the License is distributed on an "AS
//  IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
//  express or implied. See the License for the specific language
//  governing permissions and limitations under the License.

// Package trace provides a ring buffer utility to trace events.
package trace

import (
	"bytes"
	"fmt"
	"sync"
)

// A Msg is a trace message, which might be repeated.
type Msg struct {
	Title string
	Body  []byte `json:"Body,omitempty"`

	// Repeats will be >1 when there was a "run" of consolidated,
	// repeated trace messages.
	Repeats uint64
}

// A RingBuffer provides a ring buffer to capture trace messages,
// along with an optional consolidation func that can merge similar,
// consecutive trace messages.
type RingBuffer struct {
	consolidateFunc MsgConsolidateFunc

	m sync.Mutex // Protects the fields that follow.

	next int // Index in msgs where next entry will be written.
	msgs []Msg
}

// MsgConsolidateFunc is the func signature of an optional merge
// function, allowing for similar trace messages to be consolidated.
// For example, instead of using 216 individual slots in the ring
// buffer, in order to save space, the consolidation func can arrange
// for just a single entry of "216 repeated mutations" to be used
// instead.
//
// The next and prev Msg parameters may be modified by the
// consolidate func.  The consolidate func should return nil if it
// performed a consolidation and doesn't want a new entry written.
type MsgConsolidateFunc func(next *Msg, prev *Msg) *Msg

// ConsolidateByTitle implements the MsgConsolidateFunc signature
// by consolidating trace message when their titles are the same.
func ConsolidateByTitle(next *Msg, prev *Msg) *Msg {
	if prev == nil || prev.Title != next.Title {
		return next
	}

	prev.Repeats++
	return nil
}

// NewRingBuffer returns a RingBuffer initialized with the
// given capacity and optional consolidateFunc.
func NewRingBuffer(
	capacity int,
	consolidateFunc MsgConsolidateFunc) *RingBuffer {
	return &RingBuffer{
		consolidateFunc: consolidateFunc,
		next:            0,
		msgs:            make([]Msg, capacity),
	}
}

// Add appens a trace message to the ring buffer, consolidating trace
// messages based on the optional consolidation function.
func (trb *RingBuffer) Add(title string, body []byte) {
	if len(trb.msgs) <= 0 {
		return
	}

	msg := &Msg{
		Title:   title,
		Body:    body,
		Repeats: 1,
	}

	trb.m.Lock()

	if trb.consolidateFunc != nil {
		msg = trb.consolidateFunc(msg, trb.lastUNLOCKED())
		if msg == nil {
			trb.m.Unlock()

			return
		}
	}

	trb.msgs[trb.next] = *msg

	trb.next++
	if trb.next >= len(trb.msgs) {
		trb.next = 0
	}

	trb.m.Unlock()
}

// Cap returns the capacity of the ring buffer.
func (trb *RingBuffer) Cap() int {
	return len(trb.msgs)
}

// Last returns the last trace in the ring buffer.
func (trb *RingBuffer) Last() *Msg {
	trb.m.Lock()
	last := trb.lastUNLOCKED()
	trb.m.Unlock()
	return last
}

func (trb *RingBuffer) lastUNLOCKED() *Msg {
	if len(trb.msgs) <= 0 {
		return nil
	}
	last := trb.next - 1
	if last < 0 {
		last = len(trb.msgs) - 1
	}
	return &trb.msgs[last]
}

// Msgs returns a copy of all the trace messages, as an array with the
// oldest trace message first.
func (trb *RingBuffer) Msgs() []Msg {
	rv := make([]Msg, 0, len(trb.msgs))

	trb.m.Lock()

	i := trb.next
	for {
		if trb.msgs[i].Title != "" {
			rv = append(rv, trb.msgs[i])
		}

		i++
		if i >= len(trb.msgs) {
			i = 0
		}

		if i == trb.next { // We've returned to the beginning.
			break
		}
	}

	trb.m.Unlock()

	return rv
}

// MsgsToString formats a []Msg into a pretty string.
// lineSep is usually something like "\n".
// linePrefix is usually something like "  ".
func MsgsToString(msgs []Msg, lineSep, linePrefix string) string {
	linePrefixRest := lineSep + linePrefix

	var buf bytes.Buffer

	for i := range msgs {
		msg := &msgs[i]

		body := ""
		bodySep := ""
		if msg.Body != nil {
			body = string(msg.Body)
			bodySep = " "
		}

		linePrefixCur := ""
		if i > 0 {
			linePrefixCur = linePrefixRest
		}

		if msg.Repeats > 1 {
			fmt.Fprintf(&buf, "%s%s (%dx)%s%s",
				linePrefixCur, msg.Title, msg.Repeats, bodySep, body)
		} else {
			fmt.Fprintf(&buf, "%s%s%s%s",
				linePrefixCur, msg.Title, bodySep, body)
		}
	}

	return buf.String()
}
