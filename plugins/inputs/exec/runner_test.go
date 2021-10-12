//go:build !windows
// +build !windows

// TODO: Windows - should be enabled for Windows when super asterisk is fixed on Windows
// https://github.com/influxdata/telegraf/issues/6248

package exec

import (
	"bytes"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type CarriageReturnTest struct {
	input  []byte
	output []byte
}

var crTests = []CarriageReturnTest{
	{[]byte{0x4c, 0x69, 0x6e, 0x65, 0x20, 0x31, 0x0d, 0x0a, 0x4c, 0x69,
		0x6e, 0x65, 0x20, 0x32, 0x0d, 0x0a, 0x4c, 0x69, 0x6e, 0x65,
		0x20, 0x33},
		[]byte{0x4c, 0x69, 0x6e, 0x65, 0x20, 0x31, 0x0a, 0x4c, 0x69, 0x6e,
			0x65, 0x20, 0x32, 0x0a, 0x4c, 0x69, 0x6e, 0x65, 0x20, 0x33}},
	{[]byte{0x4c, 0x69, 0x6e, 0x65, 0x20, 0x31, 0x0a, 0x4c, 0x69, 0x6e,
		0x65, 0x20, 0x32, 0x0a, 0x4c, 0x69, 0x6e, 0x65, 0x20, 0x33},
		[]byte{0x4c, 0x69, 0x6e, 0x65, 0x20, 0x31, 0x0a, 0x4c, 0x69, 0x6e,
			0x65, 0x20, 0x32, 0x0a, 0x4c, 0x69, 0x6e, 0x65, 0x20, 0x33}},
	{[]byte{0x54, 0x68, 0x69, 0x73, 0x20, 0x69, 0x73, 0x20, 0x61, 0x6c,
		0x6c, 0x20, 0x6f, 0x6e, 0x65, 0x20, 0x62, 0x69, 0x67, 0x20,
		0x6c, 0x69, 0x6e, 0x65},
		[]byte{0x54, 0x68, 0x69, 0x73, 0x20, 0x69, 0x73, 0x20, 0x61, 0x6c,
			0x6c, 0x20, 0x6f, 0x6e, 0x65, 0x20, 0x62, 0x69, 0x67, 0x20,
			0x6c, 0x69, 0x6e, 0x65}},
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name string
		bufF func() *bytes.Buffer
		expF func() *bytes.Buffer
	}{
		{
			name: "should not truncate",
			bufF: func() *bytes.Buffer {
				var b bytes.Buffer
				_, err := b.WriteString("hello world")
				require.NoError(t, err)
				return &b
			},
			expF: func() *bytes.Buffer {
				var b bytes.Buffer
				_, err := b.WriteString("hello world")
				require.NoError(t, err)
				return &b
			},
		},
		{
			name: "should truncate up to the new line",
			bufF: func() *bytes.Buffer {
				var b bytes.Buffer
				_, err := b.WriteString("hello world\nand all the people")
				require.NoError(t, err)
				return &b
			},
			expF: func() *bytes.Buffer {
				var b bytes.Buffer
				_, err := b.WriteString("hello world...")
				require.NoError(t, err)
				return &b
			},
		},
		{
			name: "should truncate to the MaxStderrBytes",
			bufF: func() *bytes.Buffer {
				var b bytes.Buffer
				for i := 0; i < 2*MaxStderrBytes; i++ {
					require.NoError(t, b.WriteByte('b'))
				}
				return &b
			},
			expF: func() *bytes.Buffer {
				var b bytes.Buffer
				for i := 0; i < MaxStderrBytes; i++ {
					require.NoError(t, b.WriteByte('b'))
				}
				_, err := b.WriteString("...")
				require.NoError(t, err)
				return &b
			},
		},
	}

	c := CommandRunner{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := c.truncate(*tt.bufF())
			require.Equal(t, tt.expF().Bytes(), res.Bytes())
		})
	}
}

func TestRemoveCarriageReturns(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Test that all carriage returns are removed
		for _, test := range crTests {
			b := bytes.NewBuffer(test.input)
			out := removeWindowsCarriageReturns(*b)
			assert.True(t, bytes.Equal(test.output, out.Bytes()))
		}
	} else {
		// Test that the buffer is returned unaltered
		for _, test := range crTests {
			b := bytes.NewBuffer(test.input)
			out := removeWindowsCarriageReturns(*b)
			assert.True(t, bytes.Equal(test.input, out.Bytes()))
		}
	}
}
