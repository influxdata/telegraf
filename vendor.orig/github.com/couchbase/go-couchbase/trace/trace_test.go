//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the
//  License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing,
//  software distributed under the License is distributed on an "AS
//  IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
//  express or implied. See the License for the specific language
//  governing permissions and limitations under the License.

package trace

import (
	"reflect"
	"testing"
)

func TestTrace(t *testing.T) {
	r := NewRingBuffer(2, nil)
	if r.Cap() != 2 {
		t.Errorf("expected 2")
	}

	m := r.Msgs()
	exp := []Msg{}
	if !reflect.DeepEqual(m, exp) {
		t.Errorf("expected %#v, got %#v", exp, m)
	}

	r.Add("hi", nil)
	m = r.Msgs()
	exp = []Msg{
		Msg{"hi", nil, 1},
	}
	if !reflect.DeepEqual(m, exp) {
		t.Errorf("expected %#v, got %#v", exp, m)
	}

	r.Add("bye", nil)
	m = r.Msgs()
	exp = []Msg{
		Msg{"hi", nil, 1},
		Msg{"bye", nil, 1},
	}
	if !reflect.DeepEqual(m, exp) {
		t.Errorf("expected %#v, got %#v", exp, m)
	}

	r.Add("buh", nil)
	m = r.Msgs()
	exp = []Msg{
		Msg{"bye", nil, 1},
		Msg{"buh", nil, 1},
	}
	if !reflect.DeepEqual(m, exp) {
		t.Errorf("expected %#v, got %#v", exp, m)
	}

	r.Add("buh", nil)
	m = r.Msgs()
	exp = []Msg{
		Msg{"buh", nil, 1},
		Msg{"buh", nil, 1},
	}
	if !reflect.DeepEqual(m, exp) {
		t.Errorf("expected %#v, got %#v", exp, m)
	}

	if !reflect.DeepEqual(r.Last(), &Msg{"buh", nil, 1}) {
		t.Errorf("expected last to be buh")
	}

	s := MsgsToString(r.Msgs(), "\n", "")
	exps := "buh\nbuh"
	if s != exps {
		t.Errorf("expected string %q, got %q", exps, s)
	}

	s = MsgsToString(r.Msgs(), "\n", "foo")
	exps = "buh\nfoobuh"
	if s != exps {
		t.Errorf("expected string %q, got %q", exps, s)
	}
}

func TestTraceConsolidateByTitle(t *testing.T) {
	r := NewRingBuffer(2, ConsolidateByTitle)
	if r.Cap() != 2 {
		t.Errorf("expected 2")
	}

	m := r.Msgs()
	exp := []Msg{}
	if !reflect.DeepEqual(m, exp) {
		t.Errorf("expected %#v, got %#v", exp, m)
	}

	r.Add("hi", nil)
	m = r.Msgs()
	exp = []Msg{
		Msg{"hi", nil, 1},
	}
	if !reflect.DeepEqual(m, exp) {
		t.Errorf("expected %#v, got %#v", exp, m)
	}

	r.Add("hi", nil)
	m = r.Msgs()
	exp = []Msg{
		Msg{"hi", nil, 2},
	}
	if !reflect.DeepEqual(m, exp) {
		t.Errorf("expected %#v, got %#v", exp, m)
	}

	r.Add("hi", nil)
	m = r.Msgs()
	exp = []Msg{
		Msg{"hi", nil, 3},
	}
	if !reflect.DeepEqual(m, exp) {
		t.Errorf("expected %#v, got %#v", exp, m)
	}

	r.Add("bye", nil)
	m = r.Msgs()
	exp = []Msg{
		Msg{"hi", nil, 3},
		Msg{"bye", nil, 1},
	}
	if !reflect.DeepEqual(m, exp) {
		t.Errorf("expected %#v, got %#v", exp, m)
	}

	r.Add("bye", nil)
	m = r.Msgs()
	exp = []Msg{
		Msg{"hi", nil, 3},
		Msg{"bye", nil, 2},
	}
	if !reflect.DeepEqual(m, exp) {
		t.Errorf("expected %#v, got %#v", exp, m)
	}

	r.Add("buh", nil)
	m = r.Msgs()
	exp = []Msg{
		Msg{"bye", nil, 2},
		Msg{"buh", nil, 1},
	}
	if !reflect.DeepEqual(m, exp) {
		t.Errorf("expected %#v, got %#v", exp, m)
	}

	r.Add("buh", nil)
	m = r.Msgs()
	exp = []Msg{
		Msg{"bye", nil, 2},
		Msg{"buh", nil, 2},
	}
	if !reflect.DeepEqual(m, exp) {
		t.Errorf("expected %#v, got %#v", exp, m)
	}

	if !reflect.DeepEqual(r.Last(), &Msg{"buh", nil, 2}) {
		t.Errorf("expected last to be buh")
	}

	s := MsgsToString(r.Msgs(), "\n", "")
	exps := "bye (2 times)\nbuh (2 times)"
	if s != exps {
		t.Errorf("expected string %q, got %q", exps, s)
	}

	s = MsgsToString(r.Msgs(), "\n", "prefix")
	exps = "bye (2 times)\nprefixbuh (2 times)"
	if s != exps {
		t.Errorf("expected string %q, got %q", exps, s)
	}
}
