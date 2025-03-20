package modbus_server

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/x448/float16"
)

func TestGetBounds(t *testing.T) {
	tests := []struct {
		entry         MemoryEntry
		expectedStart uint16
		expectedEnd   uint16
		expectError   bool
	}{
		{MemoryEntry{Address: 0, Type: "BIT"}, 0, 1, false},
		{MemoryEntry{Address: 100, Type: "UINT16"}, 100, 101, false},
		{MemoryEntry{Address: 200, Type: "FLOAT32"}, 200, 202, false},
		{MemoryEntry{Address: 300, Type: "INT64"}, 300, 304, false},
		{MemoryEntry{Address: 400, Type: "INVALID"}, 0, 0, true},
	}

	for _, test := range tests {
		start, end, _ := test.entry.getBounds()
		require.Equal(t, test.expectedStart, start)
		require.Equal(t, test.expectedEnd, end)
		// check for unsupported type
		_, _, err := MemoryLayout{test.entry}.HasOverlap()
		if test.expectError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}

	// Test for unsupported type
	start, end, bit := MemoryEntry{Address: 1, Type: "INVALID"}.getBounds()
	require.Equal(t, uint16(0), start)
	require.Equal(t, uint16(0), end)
	require.Equal(t, uint8(0), bit)

	_, _, err := MemoryLayout{MemoryEntry{Address: 0, Type: "INVALID"}}.HasOverlap()
	require.Error(t, err)
}

func TestHasOverlap(t *testing.T) {
	tests := []struct {
		layout        MemoryLayout
		expectOverlap bool
		expectedError bool
	}{
		{
			MemoryLayout{
				{Address: 0, Type: "BIT"},
				{Address: 1, Type: "UINT16"},
				{Address: 3, Type: "FLOAT32"},
			},
			false,
			false,
		},
		{
			MemoryLayout{
				{Address: 0, Type: "BIT"},
				{Address: 1, Type: "UINT16"},
				{Address: 2, Type: "FLOAT32"},
			},
			false,
			false,
		},
		{
			MemoryLayout{
				{Address: 0, Type: "BIT"},
				{Address: 1, Type: "FLOAT32"},
				{Address: 2, Type: "UINT16"},
			},
			true,
			false,
		},
		{
			MemoryLayout{
				{Address: 0, Type: "BIT", Register: "register", Bit: 0},
				{Address: 0, Type: "BIT", Register: "register", Bit: 1},
				{Address: 0, Type: "BIT", Register: "register", Bit: 2},
				{Address: 0, Type: "BIT", Register: "register", Bit: 3},
				{Address: 0, Type: "BIT", Register: "register", Bit: 4},
				{Address: 0, Type: "BIT", Register: "register", Bit: 5},
				{Address: 0, Type: "BIT", Register: "register", Bit: 6},
				{Address: 0, Type: "BIT", Register: "register", Bit: 7},
				{Address: 0, Type: "BIT", Register: "register", Bit: 8},
				{Address: 0, Type: "BIT", Register: "register", Bit: 9},
				{Address: 0, Type: "BIT", Register: "register", Bit: 10},
				{Address: 0, Type: "BIT", Register: "register", Bit: 11},
				{Address: 0, Type: "BIT", Register: "register", Bit: 12},
				{Address: 0, Type: "BIT", Register: "register", Bit: 13},
				{Address: 0, Type: "BIT", Register: "register", Bit: 14},
				{Address: 0, Type: "BIT", Register: "register", Bit: 15},
				{Address: 0, Type: "BIT", Register: "register", Bit: 16},
			},
			true,
			true,
		},
		{
			MemoryLayout{
				{Address: 0, Type: "BIT", Register: "register", Bit: 0},
				{Address: 0, Type: "BIT", Register: "register", Bit: 1},
				{Address: 0, Type: "BIT", Register: "register", Bit: 2},
				{Address: 0, Type: "BIT", Register: "register", Bit: 3},
				{Address: 0, Type: "BIT", Register: "register", Bit: 4},
				{Address: 0, Type: "BIT", Register: "register", Bit: 5},
				{Address: 0, Type: "BIT", Register: "register", Bit: 6},
				{Address: 0, Type: "BIT", Register: "register", Bit: 7},
				{Address: 0, Type: "BIT", Register: "register", Bit: 8},
				{Address: 0, Type: "BIT", Register: "register", Bit: 9},
				{Address: 0, Type: "BIT", Register: "register", Bit: 10},
				{Address: 0, Type: "BIT", Register: "register", Bit: 11},
				{Address: 0, Type: "BIT", Register: "register", Bit: 12},
				{Address: 0, Type: "BIT", Register: "register", Bit: 13},
				{Address: 0, Type: "BIT", Register: "register", Bit: 14},
				{Address: 0, Type: "BIT", Register: "register", Bit: 15},
			},
			false,
			false,
		},
	}

	for _, test := range tests {
		hasOverlap, _, err := test.layout.HasOverlap()
		require.Equal(t, test.expectOverlap, hasOverlap)
		if test.expectedError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestGetCoilsAndRegisters(t *testing.T) {
	layout := MemoryLayout{
		{Address: 1, Register: "coil"},
		{Address: 3, Register: "coil"},
		{Address: 40000, Type: "BIT", Register: "register", Bit: 0},
		{Address: 40000, Type: "BIT", Register: "register", Bit: 1},
		{Address: 40000, Type: "BIT", Register: "register", Bit: 2},
		{Address: 40000, Type: "BIT", Register: "register", Bit: 3},
		{Address: 40000, Type: "BIT", Register: "register", Bit: 4},
		{Address: 40000, Type: "BIT", Register: "register", Bit: 5},
		{Address: 40000, Type: "BIT", Register: "register", Bit: 6},
		{Address: 40000, Type: "BIT", Register: "register", Bit: 7},
		{Address: 40000, Type: "BIT", Register: "register", Bit: 8},
		{Address: 40000, Type: "BIT", Register: "register", Bit: 9},
		{Address: 40000, Type: "BIT", Register: "register", Bit: 10},
		{Address: 40000, Type: "BIT", Register: "register", Bit: 11},
		{Address: 40000, Type: "BIT", Register: "register", Bit: 12},
		{Address: 40000, Type: "BIT", Register: "register", Bit: 13},
		{Address: 40000, Type: "BIT", Register: "register", Bit: 14},
		{Address: 40000, Type: "BIT", Register: "register", Bit: 15},
		{Address: 40001, Type: "UINT32", Register: "register"},
		{Address: 40003, Type: "UINT32", Register: "register"},
	}
	expectedCoils := []bool{false, false, false}
	expectedRegisters := []uint16{0, 0, 0, 0, 0}
	expectedCoilOffset := uint16(1)
	expectedRegisterOffset := uint16(40000)

	coils, registers := layout.GetCoilsAndRegisters()
	coilOffset, registerOffset := layout.GetMemoryOffsets()
	require.Equal(t, expectedCoils, coils)
	require.Equal(t, expectedRegisters, registers)
	require.Equal(t, expectedCoilOffset, coilOffset)
	require.Equal(t, expectedRegisterOffset, registerOffset)
}

func TestParseMemoryBigEndian(t *testing.T) {
	var emptyRegisters []uint16
	var emptyCoils []bool

	tests := []struct {
		byteOrder      string
		entry          MemoryEntry
		coilOffset     uint16
		registerOffset uint16
		coils          []bool
		registers      []uint16
		expected       any
		expectError    bool
	}{
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Register: "coil"},
			coilOffset: 0, registerOffset: 0, coils: []bool{true}, registers: emptyRegisters, expected: true, expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "UINT16", Register: "register"},
			coilOffset: 0, registerOffset: 0, coils: emptyCoils, registers: []uint16{12345}, expected: uint16(12345), expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "FLOAT32", Register: "register"},
			coilOffset: 0, registerOffset: 0, coils: emptyCoils, registers: []uint16{0x3f80, 0x0000}, expected: float32(1.0), expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "INT32", Register: "register"},
			coilOffset: 0, registerOffset: 0, coils: emptyCoils, registers: []uint16{0xffff, 0xffff}, expected: int32(-1), expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "UINT32", Register: "register"},
			coilOffset: 0, registerOffset: 0, coils: emptyCoils, registers: []uint16{0x0000, 0x0001}, expected: uint32(1), expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "INT64", Register: "register"},
			coilOffset: 0, registerOffset: 0, coils: emptyCoils, registers: []uint16{0xffff, 0xffff, 0xffff, 0xffff}, expected: int64(-1), expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "UINT64", Register: "register"},
			coilOffset: 0, registerOffset: 0, coils: emptyCoils, registers: []uint16{0x0000, 0x0000, 0x0000, 0x0001}, expected: uint64(1), expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "FLOAT64", Register: "register"},
			coilOffset: 0, registerOffset: 0, coils: emptyCoils, registers: []uint16{0x3ff0, 0x0000, 0x0000, 0x0000}, expected: float64(1.0),
			expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "INT8L", Register: "register"},
			coilOffset: 0, registerOffset: 0, coils: emptyCoils, registers: []uint16{0x007f}, expected: int8(127), expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "INT8H", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x7f00}, expected: int8(127), expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "UINT8L", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x00ff}, expected: uint8(255), expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "UINT8H", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0xff00}, expected: uint8(255), expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "FLOAT16", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x3c00}, expected: float32(1.0), expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "INVALID", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: emptyRegisters, expected: nil, expectError: true,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "BIT", Register: "register", Bit: 0}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x0001}, expected: true, expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "BIT", Register: "register", Bit: 0}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x1110}, expected: false, expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "STRING", Register: "register", Length: 3}, coilOffset: 0, registerOffset: 0,
			coils: emptyCoils, registers: []uint16{0x4865, 0x6c6c, 0x6f00}, expected: "Hello", expectError: false,
		},
	}
	for _, test := range tests {
		value, err := ParseMemory(test.byteOrder, test.entry, test.coilOffset, test.registerOffset, test.coils, test.registers)
		if test.expectError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, test.expected, value)
		}
	}
}

func TestParseMemoryBigEndianByteSwap(t *testing.T) {
	var emptyRegisters []uint16
	var emptyCoils []bool

	tests := []struct {
		byteOrder      string
		entry          MemoryEntry
		coilOffset     uint16
		registerOffset uint16
		coils          []bool
		registers      []uint16
		expected       any
		expectError    bool
	}{
		{
			byteOrder: "BADC", entry: MemoryEntry{Address: 0, Register: "coil"}, coilOffset: 0, registerOffset: 0, coils: []bool{true},
			registers: emptyRegisters,
			expected:  true, expectError: false,
		},
		{
			byteOrder: "BADC", entry: MemoryEntry{Address: 0, Type: "UINT16", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{12345}, expected: uint16(12345), expectError: false,
		},
		{
			byteOrder: "BADC", entry: MemoryEntry{Address: 0, Type: "FLOAT32", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x3f80, 0x0000}, expected: float32(1.0), expectError: false,
		},
		{
			byteOrder: "BADC", entry: MemoryEntry{Address: 0, Type: "INT32", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0xffff, 0xffff}, expected: int32(-1), expectError: false,
		},
		{
			byteOrder: "BADC", entry: MemoryEntry{Address: 0, Type: "UINT32", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x0000, 0x0001}, expected: uint32(1), expectError: false,
		},
		{
			byteOrder: "BADC", entry: MemoryEntry{Address: 0, Type: "INT64", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0xffff, 0xffff, 0xffff, 0xffff}, expected: int64(-1), expectError: false,
		},
		{
			byteOrder: "BADC", entry: MemoryEntry{Address: 0, Type: "UINT64", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x0000, 0x0000, 0x0000, 0x0001}, expected: uint64(1), expectError: false,
		},
		{
			byteOrder: "BADC", entry: MemoryEntry{Address: 0, Type: "FLOAT64", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x3ff0, 0x0000, 0x0000, 0x0000}, expected: float64(1.0), expectError: false,
		},
		{
			byteOrder: "BADC", entry: MemoryEntry{Address: 0, Type: "INT8L", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x007f}, expected: nil, expectError: true,
		},
		{
			byteOrder: "BADC", entry: MemoryEntry{Address: 0, Type: "INT8H", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x7f00}, expected: nil, expectError: true,
		},
		{
			byteOrder: "BADC", entry: MemoryEntry{Address: 0, Type: "UINT8L", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x00ff}, expected: nil, expectError: true,
		},
		{
			byteOrder: "BADC", entry: MemoryEntry{Address: 0, Type: "UINT8H", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0xff00}, expected: nil, expectError: true,
		},
		{
			byteOrder: "BADC", entry: MemoryEntry{Address: 0, Type: "FLOAT16", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x3c00}, expected: float32(1.0), expectError: false,
		},
		{
			byteOrder: "BADC", entry: MemoryEntry{Address: 0, Type: "INVALID", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: emptyRegisters, expected: nil, expectError: true,
		},
		{
			byteOrder: "BADC", entry: MemoryEntry{Address: 0, Type: "BIT", Register: "register", Bit: 0}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x0001}, expected: true, expectError: false,
		},
		{
			byteOrder: "BADC", entry: MemoryEntry{Address: 0, Type: "BIT", Register: "register", Bit: 0}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x1110}, expected: false, expectError: false,
		},
		{
			byteOrder: "BADC", entry: MemoryEntry{Address: 0, Type: "STRING", Register: "register", Length: 3}, coilOffset: 0, registerOffset: 0,
			coils: emptyCoils, registers: []uint16{0x4865, 0x6c6c, 0x6f00}, expected: "Hello", expectError: false,
		},
	}

	for _, test := range tests {
		value, err := ParseMemory(test.byteOrder, test.entry, test.coilOffset, test.registerOffset, test.coils, test.registers)
		if test.expectError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, test.expected, value)
		}
	}
}

func TestParseMemoryLittleEndian(t *testing.T) {
	var emptyRegisters []uint16
	var emptyCoils []bool

	tests := []struct {
		byteOrder      string
		entry          MemoryEntry
		coilOffset     uint16
		registerOffset uint16
		coils          []bool
		registers      []uint16
		expected       any
		expectError    bool
	}{
		{
			byteOrder: "DCBA", entry: MemoryEntry{Address: 0, Register: "coil"}, coilOffset: 0, registerOffset: 0, coils: []bool{true},
			registers: emptyRegisters,
			expected:  true, expectError: false,
		},
		{
			byteOrder: "DCBA", entry: MemoryEntry{Address: 0, Type: "UINT16", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{12345}, expected: uint16(12345), expectError: false,
		},
		{
			byteOrder: "DCBA", entry: MemoryEntry{Address: 0, Type: "FLOAT32", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x0000, 0x3f80}, expected: float32(1.0), expectError: false,
		},
		{
			byteOrder: "DCBA", entry: MemoryEntry{Address: 0, Type: "INT32", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0xffff, 0xffff}, expected: int32(-1), expectError: false,
		},
		{
			byteOrder: "DCBA", entry: MemoryEntry{Address: 0, Type: "UINT32", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x0001, 0x0000}, expected: uint32(1), expectError: false,
		},
		{
			byteOrder: "DCBA", entry: MemoryEntry{Address: 0, Type: "INT64", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0xffff, 0xffff, 0xffff, 0xffff}, expected: int64(-1), expectError: false,
		},
		{
			byteOrder: "DCBA", entry: MemoryEntry{Address: 0, Type: "UINT64", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x0001, 0x0000, 0x0000, 0x0000}, expected: uint64(1), expectError: false,
		},
		{
			byteOrder: "DCBA", entry: MemoryEntry{Address: 0, Type: "FLOAT64", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x0000, 0x0000, 0x0000, 0x3ff0}, expected: float64(1.0), expectError: false,
		},
		{
			byteOrder: "DCBA", entry: MemoryEntry{Address: 0, Type: "INT8L", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x007f}, expected: int8(127), expectError: false,
		},
		{
			byteOrder: "DCBA", entry: MemoryEntry{Address: 0, Type: "INT8H", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x7f00}, expected: int8(127), expectError: false,
		},
		{
			byteOrder: "DCBA", entry: MemoryEntry{Address: 0, Type: "UINT8L", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x00ff}, expected: uint8(255), expectError: false,
		},
		{
			byteOrder: "DCBA", entry: MemoryEntry{Address: 0, Type: "UINT8H", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0xff00}, expected: uint8(255), expectError: false,
		},
		{
			byteOrder: "DCBA", entry: MemoryEntry{Address: 0, Type: "FLOAT16", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x3c00}, expected: float32(1.0), expectError: false,
		},
		{
			byteOrder: "DCBA", entry: MemoryEntry{Address: 0, Type: "INVALID", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: emptyRegisters, expected: nil, expectError: true,
		},
		{
			byteOrder: "DCBA", entry: MemoryEntry{Address: 0, Type: "BIT", Register: "register", Bit: 0}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x0001}, expected: true, expectError: false,
		},
		{
			byteOrder: "DCBA", entry: MemoryEntry{Address: 0, Type: "BIT", Register: "register", Bit: 0}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x1110}, expected: false, expectError: false,
		},
		{
			byteOrder: "DCBA", entry: MemoryEntry{Address: 0, Type: "STRING", Register: "register", Length: 3}, coilOffset: 0, registerOffset: 0,
			coils: emptyCoils, registers: []uint16{0x4865, 0x6c6c, 0x6f00}, expected: "Hello", expectError: false,
		},
	}

	for _, test := range tests {
		value, err := ParseMemory(test.byteOrder, test.entry, test.coilOffset, test.registerOffset, test.coils, test.registers)
		if test.expectError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, test.expected, value)
		}
	}
}

func TestParseMemoryLittleEndianByteSwap(t *testing.T) {
	var emptyRegisters []uint16
	var emptyCoils []bool

	tests := []struct {
		byteOrder      string
		entry          MemoryEntry
		coilOffset     uint16
		registerOffset uint16
		coils          []bool
		registers      []uint16
		expected       any
		expectError    bool
	}{
		{
			byteOrder: "CDAB", entry: MemoryEntry{Address: 0, Register: "coil"}, coilOffset: 0, registerOffset: 0, coils: []bool{true},
			registers: emptyRegisters,
			expected:  true, expectError: false,
		},
		{
			byteOrder: "CDAB", entry: MemoryEntry{Address: 0, Type: "UINT16", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{12345}, expected: uint16(12345), expectError: false,
		},
		{
			byteOrder: "CDAB", entry: MemoryEntry{Address: 0, Type: "FLOAT32", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x0000, 0x3f80}, expected: float32(1.0), expectError: false,
		},
		{
			byteOrder: "CDAB", entry: MemoryEntry{Address: 0, Type: "INT32", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0xffff, 0xffff}, expected: int32(-1), expectError: false,
		},
		{
			byteOrder: "CDAB", entry: MemoryEntry{Address: 0, Type: "UINT32", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x0001, 0x0000}, expected: uint32(1), expectError: false,
		},
		{
			byteOrder: "CDAB", entry: MemoryEntry{Address: 0, Type: "INT64", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0xffff, 0xffff, 0xffff, 0xffff}, expected: int64(-1), expectError: false,
		},
		{
			byteOrder: "CDAB", entry: MemoryEntry{Address: 0, Type: "UINT64", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x0001, 0x0000, 0x0000, 0x0000}, expected: uint64(1), expectError: false,
		},
		{
			byteOrder: "CDAB", entry: MemoryEntry{Address: 0, Type: "FLOAT64", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x0000, 0x0000, 0x0000, 0x3ff0}, expected: float64(1.0), expectError: false,
		},
		{
			byteOrder: "CDAB", entry: MemoryEntry{Address: 0, Type: "INT8L", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x007f}, expected: nil, expectError: true,
		},
		{
			byteOrder: "CDAB", entry: MemoryEntry{Address: 0, Type: "INT8H", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x7f00}, expected: nil, expectError: true,
		},
		{
			byteOrder: "CDAB", entry: MemoryEntry{Address: 0, Type: "UINT8L", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x00ff}, expected: nil, expectError: true,
		},
		{
			byteOrder: "CDAB", entry: MemoryEntry{Address: 0, Type: "UINT8H", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0xff00}, expected: nil, expectError: true,
		},
		{
			byteOrder: "CDAB", entry: MemoryEntry{Address: 0, Type: "FLOAT16", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x3c00}, expected: float32(1.0), expectError: false,
		},
		{
			byteOrder: "CDAB", entry: MemoryEntry{Address: 0, Type: "INVALID", Register: "register"}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: emptyRegisters, expected: nil, expectError: true,
		},
		{
			byteOrder: "CDAB", entry: MemoryEntry{Address: 0, Type: "BIT", Register: "register", Bit: 0}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x0001}, expected: true, expectError: false,
		},
		{
			byteOrder: "CDAB", entry: MemoryEntry{Address: 0, Type: "BIT", Register: "register", Bit: 0}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{0x1110}, expected: false, expectError: false,
		},
		{
			byteOrder: "CDAB", entry: MemoryEntry{Address: 0, Type: "STRING", Register: "register", Length: 3}, coilOffset: 0, registerOffset: 0,
			coils: emptyCoils, registers: []uint16{0x4865, 0x6c6c, 0x6f00}, expected: "Hello", expectError: false,
		},
	}

	for _, test := range tests {
		value, err := ParseMemory(test.byteOrder, test.entry, test.coilOffset, test.registerOffset, test.coils, test.registers)
		if test.expectError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, test.expected, value)
		}
	}
}

func TestParseBits(t *testing.T) {
	var emptyCoils []bool

	tests := []struct {
		byteOrder      string
		entry          MemoryEntry
		coilOffset     uint16
		registerOffset uint16
		coils          []bool
		registers      []uint16
		expected       any
		expectError    bool
	}{
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "BIT", Register: "register", Bit: 0}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{11784}, expected: false, expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "BIT", Register: "register", Bit: 1}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{11784}, expected: false, expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "BIT", Register: "register", Bit: 2}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{11784}, expected: false, expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "BIT", Register: "register", Bit: 3}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{11784}, expected: true, expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "BIT", Register: "register", Bit: 4}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{11784}, expected: false, expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "BIT", Register: "register", Bit: 5}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{11784}, expected: false, expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "BIT", Register: "register", Bit: 6}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{11784}, expected: false, expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "BIT", Register: "register", Bit: 7}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{11784}, expected: false, expectError: false,
		},

		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "BIT", Register: "register", Bit: 8}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{11784}, expected: false, expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "BIT", Register: "register", Bit: 9}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{11784}, expected: true, expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "BIT", Register: "register", Bit: 10}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{11784}, expected: true, expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "BIT", Register: "register", Bit: 11}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{11784}, expected: true, expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "BIT", Register: "register", Bit: 12}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{11784}, expected: false, expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "BIT", Register: "register", Bit: 13}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{11784}, expected: true, expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "BIT", Register: "register", Bit: 14}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{11784}, expected: false, expectError: false,
		},
		{
			byteOrder: "ABCD", entry: MemoryEntry{Address: 0, Type: "BIT", Register: "register", Bit: 15}, coilOffset: 0, registerOffset: 0, coils: emptyCoils,
			registers: []uint16{11784}, expected: false, expectError: false,
		},
	}

	// 11784 = 00101110 00001000

	for _, test := range tests {
		value, err := ParseMemory(test.byteOrder, test.entry, test.coilOffset, test.registerOffset, test.coils, test.registers)
		if test.expectError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, test.expected, value)
		}
	}
}

func TestGetMemoryMappedByName(t *testing.T) {
	entries := MemoryLayout{
		{HashID: 0, Measurement: "measurement1", Field: "field1", Address: 0, Type: "UINT16"},
		{HashID: 0, Measurement: "measurement1", Field: "field2", Address: 1, Type: "FLOAT32"},
		{HashID: 1, Measurement: "measurement2", Field: "field1", Address: 2, Type: "INT32"},
	}

	expected := map[uint64]map[string]MemoryEntry{
		0: {
			"field1": {HashID: 0, Measurement: "measurement1", Field: "field1", Address: 0, Type: "UINT16"},
			"field2": {HashID: 0, Measurement: "measurement1", Field: "field2", Address: 1, Type: "FLOAT32"},
		},
		1: {
			"field1": {HashID: 1, Measurement: "measurement2", Field: "field1", Address: 2, Type: "INT32"},
		},
	}

	memoryMap, err := entries.GetMemoryMappedByHashID()
	require.NoError(t, err)
	require.Equal(t, expected, memoryMap)
}

func TestCastToType(t *testing.T) {
	tests := []struct {
		value     any
		valueType string
		expected  any
	}{
		{value: int64(127), valueType: "INT8L", expected: int8(127)},
		{value: uint64(255), valueType: "UINT8L", expected: uint8(255)},
		{value: int64(127), valueType: "INT8H", expected: int8(127)},
		{value: uint64(255), valueType: "UINT8H", expected: uint8(255)},
		{value: float64(1.0), valueType: "FLOAT16", expected: float16.Fromfloat32(1.0)},
		{value: int64(123), valueType: "INT16", expected: int16(123)},
		{value: uint64(123), valueType: "UINT16", expected: uint16(123)},
		{value: float64(1.23), valueType: "FLOAT32", expected: float32(1.23)},
		{value: int64(123), valueType: "INT32", expected: int32(123)},
		{value: uint64(123), valueType: "UINT32", expected: uint32(123)},
		{value: int64(123), valueType: "INT64", expected: int64(123)},
		{value: uint64(123), valueType: "UINT64", expected: uint64(123)},
		{value: float64(1.23), valueType: "FLOAT64", expected: float64(1.23)},
		{value: "test", valueType: "STRING", expected: "test"},
	}

	for _, test := range tests {
		result := castToType(test.value, test.valueType)
		require.Equal(t, test.expected, result)
	}
}

func TestParseMetricBigEndian(t *testing.T) {
	tests := []struct {
		byteOrder   string
		value       any
		valueType   string
		expected    []uint16
		expectError bool
	}{
		{byteOrder: "ABCD", value: uint64(12345), valueType: "UINT16", expected: []uint16{12345}, expectError: false},
		{byteOrder: "ABCD", value: float64(1.0), valueType: "FLOAT32", expected: []uint16{0x3f80, 0x0000}, expectError: false},
		{byteOrder: "ABCD", value: int64(-1), valueType: "INT32", expected: []uint16{0xffff, 0xffff}, expectError: false},
		{byteOrder: "ABCD", value: uint64(1), valueType: "UINT32", expected: []uint16{0x0000, 0x0001}, expectError: false},
		{byteOrder: "ABCD", value: int64(-1), valueType: "INT64", expected: []uint16{0xffff, 0xffff, 0xffff, 0xffff}, expectError: false},
		{byteOrder: "ABCD", value: uint64(1), valueType: "UINT64", expected: []uint16{0x0000, 0x0000, 0x0000, 0x0001}, expectError: false},
		{byteOrder: "ABCD", value: float64(1.0), valueType: "FLOAT64", expected: []uint16{0x3ff0, 0x0000, 0x0000, 0x0000}, expectError: false},
		{byteOrder: "ABCD", value: "invalid", valueType: "INVALID", expected: nil, expectError: true},
		{byteOrder: "ABCD", value: int64(127), valueType: "INT8L", expected: []uint16{0x007f}, expectError: false},
		{byteOrder: "ABCD", value: int64(127), valueType: "INT8H", expected: []uint16{0x7f00}, expectError: false},
		{byteOrder: "ABCD", value: uint64(255), valueType: "UINT8L", expected: []uint16{0x00ff}, expectError: false},
		{byteOrder: "ABCD", value: uint64(255), valueType: "UINT8H", expected: []uint16{0xff00}, expectError: false},
		{byteOrder: "ABCD", value: float64(1.0), valueType: "FLOAT16", expected: []uint16{0x3c00}, expectError: false},
		{byteOrder: "ABCD", value: "Hello", valueType: "STRING", expected: []uint16{0x4865, 0x6c6c, 0x6f00}, expectError: false},
	}

	for _, test := range tests {
		result, err := ParseMetric(test.byteOrder, test.value, test.valueType, 0)
		if test.expectError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, test.expected, result)
		}
	}
}

func TestParseMetricBigEndianByteSwap(t *testing.T) {
	tests := []struct {
		byteOrder   string
		value       any
		valueType   string
		expected    []uint16
		expectError bool
	}{
		{byteOrder: "BADC", value: uint64(12345), valueType: "UINT16", expected: []uint16{12345}, expectError: false},
		{byteOrder: "BADC", value: float64(1.0), valueType: "FLOAT32", expected: []uint16{0x0000, 0x3f80}, expectError: false},
		{byteOrder: "BADC", value: int64(-1), valueType: "INT32", expected: []uint16{0xffff, 0xffff}, expectError: false},
		{byteOrder: "BADC", value: uint64(1), valueType: "UINT32", expected: []uint16{0x0001, 0x0000}, expectError: false},
		{byteOrder: "BADC", value: int64(-1), valueType: "INT64", expected: []uint16{0xffff, 0xffff, 0xffff, 0xffff}, expectError: false},
		{byteOrder: "BADC", value: uint64(1), valueType: "UINT64", expected: []uint16{0x0001, 0x0000, 0x0000, 0x0000}, expectError: false},
		{byteOrder: "BADC", value: float64(1.0), valueType: "FLOAT64", expected: []uint16{0x0000, 0x0000, 0x0000, 0x3ff0}, expectError: false},
		{byteOrder: "BADC", value: "invalid", valueType: "INVALID", expected: nil, expectError: true},
		{byteOrder: "BADC", value: int64(127), valueType: "INT8L", expected: []uint16{0x007f}, expectError: false},
		{byteOrder: "BADC", value: int64(127), valueType: "INT8H", expected: []uint16{0x7f00}, expectError: false},
		{byteOrder: "BADC", value: uint64(255), valueType: "UINT8L", expected: []uint16{0x00ff}, expectError: false},
		{byteOrder: "BADC", value: uint64(255), valueType: "UINT8H", expected: []uint16{0xff00}, expectError: false},
		{byteOrder: "BADC", value: float64(1.0), valueType: "FLOAT16", expected: []uint16{0x3c00}, expectError: false},
		{byteOrder: "BADC", value: "Hello", valueType: "STRING", expected: []uint16{0x4865, 0x6c6c, 0x6f00}, expectError: false},
	}

	for _, test := range tests {
		result, err := ParseMetric(test.byteOrder, test.value, test.valueType, 0)
		if test.expectError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, test.expected, result)
		}
	}
}

func TestParseMetricLittleEndian(t *testing.T) {
	tests := []struct {
		byteOrder   string
		value       any
		valueType   string
		expected    []uint16
		expectError bool
	}{
		{byteOrder: "DCBA", value: uint64(12345), valueType: "UINT16", expected: []uint16{12345}, expectError: false},
		{byteOrder: "DCBA", value: float64(1.0), valueType: "FLOAT32", expected: []uint16{0x0000, 0x3f80}, expectError: false},
		{byteOrder: "DCBA", value: int64(-1), valueType: "INT32", expected: []uint16{0xffff, 0xffff}, expectError: false},
		{byteOrder: "DCBA", value: uint64(1), valueType: "UINT32", expected: []uint16{0x0001, 0x0000}, expectError: false},
		{byteOrder: "DCBA", value: int64(-1), valueType: "INT64", expected: []uint16{0xffff, 0xffff, 0xffff, 0xffff}, expectError: false},
		{byteOrder: "DCBA", value: uint64(1), valueType: "UINT64", expected: []uint16{0x0001, 0x0000, 0x0000, 0x0000}, expectError: false},
		{byteOrder: "DCBA", value: float64(1.0), valueType: "FLOAT64", expected: []uint16{0x0000, 0x0000, 0x0000, 0x3ff0}, expectError: false},
		{byteOrder: "DCBA", value: "invalid", valueType: "INVALID", expected: nil, expectError: true},
		{byteOrder: "DCBA", value: int64(127), valueType: "INT8L", expected: []uint16{0x007f}, expectError: false},
		{byteOrder: "DCBA", value: int64(127), valueType: "INT8H", expected: []uint16{0x7f00}, expectError: false},
		{byteOrder: "DCBA", value: uint64(255), valueType: "UINT8L", expected: []uint16{0x00ff}, expectError: false},
		{byteOrder: "DCBA", value: uint64(255), valueType: "UINT8H", expected: []uint16{0xff00}, expectError: false},
		{byteOrder: "DCBA", value: float64(1.0), valueType: "FLOAT16", expected: []uint16{0x3c00}, expectError: false},
		{byteOrder: "DCBA", value: "Hello", valueType: "STRING", expected: []uint16{0x4865, 0x6c6c, 0x6f00}, expectError: false},
	}

	for _, test := range tests {
		result, err := ParseMetric(test.byteOrder, test.value, test.valueType, 0)
		if test.expectError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, test.expected, result)
		}
	}
}

func TestParseMetricLittleEndianByteSwap(t *testing.T) {
	tests := []struct {
		byteOrder   string
		value       any
		valueType   string
		expected    []uint16
		expectError bool
	}{
		{byteOrder: "CDAB", value: uint64(12345), valueType: "UINT16", expected: []uint16{12345}, expectError: false},
		{byteOrder: "CDAB", value: float64(1.0), valueType: "FLOAT32", expected: []uint16{0x3f80, 0x0000}, expectError: false},
		{byteOrder: "CDAB", value: int64(-1), valueType: "INT32", expected: []uint16{0xffff, 0xffff}, expectError: false},
		{byteOrder: "CDAB", value: uint64(1), valueType: "UINT32", expected: []uint16{0x0000, 0x0001}, expectError: false},
		{byteOrder: "CDAB", value: int64(-1), valueType: "INT64", expected: []uint16{0xffff, 0xffff, 0xffff, 0xffff}, expectError: false},
		{byteOrder: "CDAB", value: uint64(1), valueType: "UINT64", expected: []uint16{0x0000, 0x0000, 0x0000, 0x0001}, expectError: false},
		{byteOrder: "CDAB", value: float64(1.0), valueType: "FLOAT64", expected: []uint16{0x3ff0, 0x0000, 0x0000, 0x0000}, expectError: false},
		{byteOrder: "CDAB", value: "invalid", valueType: "INVALID", expected: nil, expectError: true},
		{byteOrder: "CDAB", value: int64(127), valueType: "INT8L", expected: []uint16{0x007f}, expectError: false},
		{byteOrder: "CDAB", value: int64(127), valueType: "INT8H", expected: []uint16{0x7f00}, expectError: false},
		{byteOrder: "CDAB", value: uint64(255), valueType: "UINT8L", expected: []uint16{0x00ff}, expectError: false},
		{byteOrder: "CDAB", value: uint64(255), valueType: "UINT8H", expected: []uint16{0xff00}, expectError: false},
		{byteOrder: "CDAB", value: float64(1.0), valueType: "FLOAT16", expected: []uint16{0x3c00}, expectError: false},
		{byteOrder: "CDAB", value: "Hello", valueType: "STRING", expected: []uint16{0x4865, 0x6c6c, 0x6f00}, expectError: false},
	}
	for _, test := range tests {
		result, err := ParseMetric(test.byteOrder, test.value, test.valueType, 0)
		if test.expectError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, test.expected, result)
		}
	}
}
