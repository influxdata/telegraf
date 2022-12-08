package netflow

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeInt32(t *testing.T) {
	buf := []byte{0x82, 0xad, 0x80, 0x86}
	out, ok := decodeInt32(buf).(int64)
	require.True(t, ok)
	require.Equal(t, int64(-2102558586), out)
}
