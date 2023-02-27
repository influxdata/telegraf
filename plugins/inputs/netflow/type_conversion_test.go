package netflow

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeInt32(t *testing.T) {
	buf := []byte{0x82, 0xad, 0x80, 0x86}
	out, ok := decodeInt32(buf).(int64)
	require.True(t, ok)
	require.Equal(t, int64(-2102558586), out)
}

func TestDecodeUint(t *testing.T) {
	tests := []struct {
		name     string
		in       []byte
		expected uint64
	}{
		{
			name:     "uint8",
			in:       []byte{0x42},
			expected: 66,
		},
		{
			name:     "uint16",
			in:       []byte{0x0A, 0x42},
			expected: 2626,
		},
		{
			name:     "uint32",
			in:       []byte{0x82, 0xad, 0x80, 0x86},
			expected: 2192408710,
		},
		{
			name:     "uint64",
			in:       []byte{0x00, 0x00, 0x23, 0x42, 0x8f, 0xad, 0x80, 0x86},
			expected: 38768785326214,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, ok := decodeUint(tt.in).(uint64)
			require.True(t, ok)
			require.Equal(t, tt.expected, out)
		})
	}
}

func TestDecodeUintInvalid(t *testing.T) {
	require.Panics(t, func() { decodeUint([]byte{0x00, 0x00, 0x00}) })
}

func TestDecodeFloat64(t *testing.T) {
	buf := []byte{0x40, 0x09, 0x21, 0xfb, 0x54, 0x44, 0x2e, 0xea}
	out, ok := decodeFloat64(buf).(float64)
	require.True(t, ok)
	require.Equal(t, float64(3.14159265359), out)
}

func TestDecodeBool(t *testing.T) {
	tests := []struct {
		name     string
		in       []byte
		expected interface{}
	}{
		{
			name:     "zero",
			in:       []byte{0x00},
			expected: uint8(0),
		},
		{
			name:     "true",
			in:       []byte{0x01},
			expected: true,
		},
		{
			name:     "false",
			in:       []byte{0x02},
			expected: false,
		},
		{
			name:     "other",
			in:       []byte{0x23},
			expected: uint8(35),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := decodeBool(tt.in)
			require.Equal(t, tt.expected, out)
		})
	}
}

func TestDecodeHex(t *testing.T) {
	buf := []byte{0x40, 0x09, 0x21, 0xfb, 0x54, 0x44, 0x2e, 0xea}
	out, ok := decodeHex(buf).(string)
	require.True(t, ok)
	require.Equal(t, "0x400921fb54442eea", out)
}

func TestDecodeString(t *testing.T) {
	buf := []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x74, 0x65, 0x6c, 0x65, 0x67, 0x72, 0x61, 0x66}
	out, ok := decodeString(buf).(string)
	require.True(t, ok)
	require.Equal(t, "hello telegraf", out)
}

func TestDecodeMAC(t *testing.T) {
	buf := []byte{0x2c, 0xf0, 0x5d, 0xe9, 0x04, 0x42}
	out, ok := decodeMAC(buf).(string)
	require.True(t, ok)
	require.Equal(t, "2c:f0:5d:e9:04:42", out)
}

func TestDecodeIP(t *testing.T) {
	tests := []struct {
		name     string
		in       []byte
		expected string
	}{
		{
			name:     "localhost IPv4",
			in:       []byte{0x7f, 0x00, 0x00, 0x01},
			expected: "127.0.0.1",
		},
		{
			name:     "unrouted IPv4",
			in:       []byte{0xc0, 0xa8, 0x04, 0x42},
			expected: "192.168.4.66",
		},
		{
			name:     "localhost IPv6",
			in:       []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			expected: "::1",
		},
		{
			name:     "local network IPv6",
			in:       []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xfe, 0x80, 0xd6, 0x8e, 0x07, 0x7f, 0x59, 0x5a, 0x23, 0xf1},
			expected: "::fe80:d68e:77f:595a:23f1",
		},
		{
			name:     "google.com IPv6",
			in:       []byte{0x00, 0x00, 0x00, 0x00, 0x2a, 0x00, 0x14, 0x50, 0x40, 0x01, 0x08, 0x11, 0x00, 0x00, 0x20, 0x0e},
			expected: "::2a00:1450:4001:811:0:200e",
		},
		{
			name:     "stripped in between IPv6",
			in:       []byte{0x2a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x14, 0x50, 0x40, 0x01, 0x08, 0x11, 0x00, 0x01, 0x20, 0x0e},
			expected: "2a00::1450:4001:811:1:200e",
		},
		{
			name:     "IPv6 not enough bytes",
			in:       []byte{0x00, 0x00, 0x00, 0xff, 0x00, 0x01},
			expected: "?000000ff0001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, ok := decodeIP(tt.in).(string)
			require.True(t, ok)
			require.Equal(t, tt.expected, out)
		})
	}
}

func TestDecodeIPFromUint32(t *testing.T) {
	in := uint32(0x7f000001)
	out, ok := decodeIPFromUint32(in).(string)
	require.True(t, ok)
	require.Equal(t, "127.0.0.1", out)
}

func TestDecodeLayer4ProtocolNumber(t *testing.T) {
	require.NoError(t, initL4ProtoMapping())

	tests := []struct {
		name     string
		in       []byte
		expected string
	}{
		{
			name:     "ICMP 1",
			in:       []byte{0x01},
			expected: "icmp",
		},
		{
			name:     "IPv4 4",
			in:       []byte{0x04},
			expected: "ipv4",
		},
		{
			name:     "IPv6 41",
			in:       []byte{0x29},
			expected: "ipv6",
		},
		{
			name:     "L2TP 115",
			in:       []byte{0x73},
			expected: "l2tp",
		},
		{
			name:     "PTP 123",
			in:       []byte{0x7b},
			expected: "ptp",
		},
		{
			name:     "unassigned 201",
			in:       []byte{0xc9},
			expected: "201",
		},
		{
			name:     "experimental 254",
			in:       []byte{0xfe},
			expected: "experimental",
		},
		{
			name:     "Reserved 255",
			in:       []byte{0xff},
			expected: "reserved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, ok := decodeL4Proto(tt.in).(string)
			require.True(t, ok)
			require.Equal(t, tt.expected, out)
		})
	}
}

func TestDecodeIPv4Options(t *testing.T) {
	require.NoError(t, initIPv4OptionMapping())

	tests := []struct {
		name     string
		bits     []int
		expected string
	}{
		{
			name:     "none",
			bits:     []int{},
			expected: "",
		},
		{
			name: "all",
			bits: []int{
				0, 1, 2, 3, 4, 5, 6, 7,
				8, 9, 10, 11, 12, 13, 14, 15,
				16, 17, 18, 19, 20, 21, 22, 23,
				24, 25, 26, 27, 28, 29, 30, 31,
			},
			expected: "EOOL,NOP,SEC,LSR,TS,E-SEC,CIPSO,RR,SID,SSR,ZSU,MTUP," +
				"MTUR,FINN,VISA,ENCODE,IMITD,EIP,TR,ADDEXT,RTRALT,SDB," +
				"UA22,DPS,UMP,QS,UA26,UA27,UA28,UA29,EXP,UA31",
		},
		{
			name:     "EOOL",
			bits:     []int{0},
			expected: "EOOL",
		},
		{
			name:     "SSR",
			bits:     []int{9},
			expected: "SSR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var options uint32
			for _, bit := range tt.bits {
				options |= 1 << bit
			}
			in := make([]byte, 4)
			binary.BigEndian.PutUint32(in, options)

			out, ok := decodeIPv4Options(in).(string)
			require.True(t, ok)
			require.Equal(t, tt.expected, out)
		})
	}
}

func TestDecodeTCPFlags(t *testing.T) {
	tests := []struct {
		name     string
		bits     []int
		expected string
		ipfix    bool
	}{
		{
			name:     "none",
			bits:     []int{},
			expected: "........",
		},
		{
			name:     "none IPFIX",
			bits:     []int{},
			expected: "................",
			ipfix:    true,
		},
		{
			name:     "all",
			bits:     []int{0, 1, 2, 3, 4, 5, 6, 7},
			expected: "CEUAPRSF",
		},
		{
			name:     "all IPFIX",
			bits:     []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			expected: "********CEUAPRSF",
			ipfix:    true,
		},
		{
			name:     "SYN",
			bits:     []int{1},
			expected: "......S.",
		},
		{
			name:     "SYN/ACK",
			bits:     []int{1, 4},
			expected: "...A..S.",
		},
		{
			name:     "ACK",
			bits:     []int{4},
			expected: "...A....",
		},
		{
			name:     "FIN",
			bits:     []int{0},
			expected: ".......F",
		},
		{
			name:     "FIN/ACK",
			bits:     []int{0, 4},
			expected: "...A...F",
		},
		{
			name:     "ACK IPFIX",
			bits:     []int{4},
			expected: "...........A....",
			ipfix:    true,
		},
		{
			name:     "FIN IPFIX",
			bits:     []int{0},
			expected: "...............F",
			ipfix:    true,
		},
		{
			name:     "ECN Nounce Sum IPFIX",
			bits:     []int{8},
			expected: ".......*........",
			ipfix:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var in []byte

			if tt.ipfix {
				var options uint16
				for _, bit := range tt.bits {
					options |= 1 << bit
				}
				in = make([]byte, 2)
				binary.BigEndian.PutUint16(in, options)
			} else {
				var options uint8
				for _, bit := range tt.bits {
					options |= 1 << bit
				}
				in = []byte{options}
			}
			out, ok := decodeTCPFlags(in).(string)
			require.True(t, ok)
			require.Equal(t, tt.expected, out)
		})
	}
}

func TestDecodeFragmentFlags(t *testing.T) {
	tests := []struct {
		name     string
		bits     []int
		expected string
	}{
		{
			name:     "none",
			bits:     []int{},
			expected: "........",
		},
		{
			name:     "all",
			bits:     []int{0, 1, 2, 3, 4, 5, 6, 7},
			expected: "RDM*****",
		},
		{
			name:     "RS",
			bits:     []int{7},
			expected: "R.......",
		},
		{
			name:     "DF",
			bits:     []int{6},
			expected: ".D......",
		},
		{
			name:     "MF",
			bits:     []int{5},
			expected: "..M.....",
		},
		{
			name:     "Bit 7 (LSB)",
			bits:     []int{0},
			expected: ".......*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var flags uint8
			for _, bit := range tt.bits {
				flags |= 1 << bit
			}
			in := []byte{flags}
			out, ok := decodeFragmentFlags(in).(string)
			require.True(t, ok)
			require.Equal(t, tt.expected, out)
		})
	}
}
