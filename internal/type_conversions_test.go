package internal

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

// stringer is a helper implementing fmt.Stringer for the conversion tests.
type stringer struct {
	s string
}

func (s stringer) String() string {
	return s.s
}

func TestToUint16OutOfRange(t *testing.T) {
	// Values above math.MaxUint16 must be reported as out of range instead of
	// being silently truncated to a 16-bit value.
	for _, input := range []any{"70000", []byte("70000"), stringer{"70000"}} {
		v, err := ToUint16(input)
		require.ErrorIsf(t, err, ErrOutOfRange, "input %v (%T)", input, input)
		require.Equalf(t, uint16(math.MaxUint16), v, "input %v (%T)", input, input)
	}

	// In-range values must still convert without error.
	v, err := ToUint16("65535")
	require.NoError(t, err)
	require.Equal(t, uint16(math.MaxUint16), v)
}

func TestToUint8OutOfRange(t *testing.T) {
	for _, input := range []any{"300", []byte("300"), stringer{"300"}} {
		v, err := ToUint8(input)
		require.ErrorIsf(t, err, ErrOutOfRange, "input %v (%T)", input, input)
		require.Equalf(t, uint8(math.MaxUint8), v, "input %v (%T)", input, input)
	}

	v, err := ToUint8("255")
	require.NoError(t, err)
	require.Equal(t, uint8(math.MaxUint8), v)
}

func TestToInt16OutOfRange(t *testing.T) {
	for _, input := range []any{"40000", []byte("40000"), stringer{"40000"}} {
		v, err := ToInt16(input)
		require.ErrorIsf(t, err, ErrOutOfRange, "input %v (%T)", input, input)
		require.Equalf(t, int16(math.MaxInt16), v, "input %v (%T)", input, input)
	}

	v, err := ToInt16("32767")
	require.NoError(t, err)
	require.Equal(t, int16(math.MaxInt16), v)
}

func TestToInt8OutOfRange(t *testing.T) {
	for _, input := range []any{"200", []byte("200"), stringer{"200"}} {
		v, err := ToInt8(input)
		require.ErrorIsf(t, err, ErrOutOfRange, "input %v (%T)", input, input)
		require.Equalf(t, int8(math.MaxInt8), v, "input %v (%T)", input, input)
	}

	v, err := ToInt8("127")
	require.NoError(t, err)
	require.Equal(t, int8(math.MaxInt8), v)
}
