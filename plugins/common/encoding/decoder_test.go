package encoding

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecoder(t *testing.T) {
	tests := []struct {
		name        string
		encoding    string
		input       []byte
		expected    []byte
		expectedErr bool
	}{
		{
			name:     "no decoder utf-8",
			encoding: "",
			input:    []byte("howdy"),
			expected: []byte("howdy"),
		},
		{
			name:     "utf-8 decoder",
			encoding: "utf-8",
			input:    []byte("howdy"),
			expected: []byte("howdy"),
		},
		{
			name:     "utf-8 decoder invalid bytes replaced with replacement char",
			encoding: "utf-8",
			input:    []byte("\xff\xfe"),
			expected: []byte("\uFFFD\uFFFD"),
		},
		{
			name:     "utf-16le decoder no BOM",
			encoding: "utf-16le",
			input:    []byte("h\x00o\x00w\x00d\x00y\x00"),
			expected: []byte("howdy"),
		},
		{
			name:     "utf-16le decoder with BOM",
			encoding: "utf-16le",
			input:    []byte("\xff\xfeh\x00o\x00w\x00d\x00y\x00"),
			expected: []byte("\xef\xbb\xbfhowdy"),
		},
		{
			name:     "utf-16be decoder no BOM",
			encoding: "utf-16be",
			input:    []byte("\x00h\x00o\x00w\x00d\x00y"),
			expected: []byte("howdy"),
		},
		{
			name:     "utf-16be decoder with BOM",
			encoding: "utf-16be",
			input:    []byte("\xfe\xff\x00h\x00o\x00w\x00d\x00y"),
			expected: []byte("\xef\xbb\xbfhowdy"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder, err := NewDecoder(tt.encoding)
			require.NoError(t, err)
			buf := bytes.NewBuffer(tt.input)
			r := decoder.Reader(buf)
			actual, err := ioutil.ReadAll(r)
			if tt.expectedErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)
		})
	}
}
