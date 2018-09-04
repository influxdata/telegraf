package couchbase

import (
	"testing"
)

func TestCleanupHost(t *testing.T) {
	tests := []struct {
		name, full, suffix, exp string
	}{
		{"empty", "", "", ""},
		{"empty suffix", "aprefix", "", "aprefix"},
		{"empty host", "", "asuffix", ""},
		{"matched suffix", "server1.example.com:11210", ".example.com:11210", "server1"},
	}

	for _, test := range tests {
		got := CleanupHost(test.full, test.suffix)
		if got != test.exp {
			t.Errorf("Error on %v: got %q, expected %q",
				test.name, got, test.exp)
		}
	}
}

func TestFindCommonSuffix(t *testing.T) {
	tests := []struct {
		name, exp string
		strings   []string
	}{
		{"empty", "", nil},
		{"one", "", []string{"blah"}},
		{"two", ".com", []string{"blah.com", "foo.com"}},
	}

	for _, test := range tests {
		got := FindCommonSuffix(test.strings)
		if got != test.exp {
			t.Errorf("Error on %v: got %q, expected %q",
				test.name, got, test.exp)
		}
	}
}

func TestParseURL(t *testing.T) {
	tests := []struct {
		in    string
		works bool
	}{
		{"", false},
		{"http://whatever/", true},
		{"http://%/", false},
	}

	for _, test := range tests {
		got, err := ParseURL(test.in)
		switch {
		case err == nil && test.works,
			!(err == nil || test.works):
		case err == nil && !test.works:
			t.Errorf("Expected failure on %v, got %v", test.in, got)
		case test.works && err != nil:
			t.Errorf("Expected success on %v, got %v", test.in, err)
		}
	}
}
