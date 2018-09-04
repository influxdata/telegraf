package couchbase

import "testing"

func TestWriteOptionsString(t *testing.T) {
	tests := []struct {
		opts WriteOptions
		exp  string
	}{
		{Raw, "raw"},
		{AddOnly, "addonly"},
		{Persist, "persist"},
		{Indexable, "indexable"},
		{Append, "append"},
		{AddOnly | Raw, "raw|addonly"},
		{0, "0x0"},
		{Raw | AddOnly | Persist | Indexable | Append,
			"raw|addonly|persist|indexable|append"},
		{Raw | 8192, "raw|0x2000"},
	}

	for _, test := range tests {
		got := test.opts.String()
		if got != test.exp {
			t.Errorf("Expected %v, got %v", test.exp, got)
		}
	}
}
