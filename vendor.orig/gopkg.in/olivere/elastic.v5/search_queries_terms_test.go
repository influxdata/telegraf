// Copyright 2012-present Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

package elastic

import (
	"encoding/json"
	"testing"
)

func TestTermsQuery(t *testing.T) {
	q := NewTermsQuery("user", "ki")
	src, err := q.Source()
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(src)
	if err != nil {
		t.Fatalf("marshaling to JSON failed: %v", err)
	}
	got := string(data)
	expected := `{"terms":{"user":["ki"]}}`
	if got != expected {
		t.Errorf("expected\n%s\n,got:\n%s", expected, got)
	}
}

func TestTermsQueryWithEmptyArray(t *testing.T) {
	included := make([]interface{}, 0)
	q := NewTermsQuery("tags", included...)
	src, err := q.Source()
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(src)
	if err != nil {
		t.Fatalf("marshaling to JSON failed: %v", err)
	}
	got := string(data)
	expected := `{"terms":{"tags":[]}}`
	if got != expected {
		t.Errorf("expected\n%s\n,got:\n%s", expected, got)
	}
}

func TestTermsQueryWithTermsLookup(t *testing.T) {
	q := NewTermsQuery("user").
		TermsLookup(NewTermsLookup().Index("users").Type("user").Id("2").Path("followers"))
	src, err := q.Source()
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(src)
	if err != nil {
		t.Fatalf("marshaling to JSON failed: %v", err)
	}
	got := string(data)
	expected := `{"terms":{"user":{"id":"2","index":"users","path":"followers","type":"user"}}}`
	if got != expected {
		t.Errorf("expected\n%s\n,got:\n%s", expected, got)
	}
}

func TestTermQuerysWithOptions(t *testing.T) {
	q := NewTermsQuery("user", "ki", "ko")
	q = q.Boost(2.79)
	q = q.QueryName("my_tq")
	src, err := q.Source()
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(src)
	if err != nil {
		t.Fatalf("marshaling to JSON failed: %v", err)
	}
	got := string(data)
	expected := `{"terms":{"_name":"my_tq","boost":2.79,"user":["ki","ko"]}}`
	if got != expected {
		t.Errorf("expected\n%s\n,got:\n%s", expected, got)
	}
}
