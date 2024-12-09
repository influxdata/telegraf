package binary

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/stretchr/testify/require"
)

func TestEntryExtract(t *testing.T) {
	testdata := []byte{0x01, 0x02, 0x03, 0x04}

	e := &Entry{Type: "uint64"}
	_, _, err := e.extract(testdata, 0)
	require.EqualError(t, err, `unexpected entry: &{ uint64 0 false    [] <nil>}`)
}

func TestEntryConvertType(t *testing.T) {
	testdata := []byte{0x01, 0x02, 0x03, 0x04}

	e := &Entry{Type: "garbage"}
	_, err := e.convertType(testdata, internal.HostEndianness)
	require.EqualError(t, err, `cannot handle type "garbage"`)
}

func TestEntryConvertTimeType(t *testing.T) {
	testdata := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}

	e := &Entry{Type: "unix_ns", location: time.UTC}
	_, err := e.convertTimeType(testdata, internal.HostEndianness)
	require.EqualError(t, err, `too many bytes 9 vs 8`)
}

func TestConvertNumericType(t *testing.T) {
	testdata := []byte{0x01, 0x02, 0x03, 0x04}

	_, err := convertNumericType(testdata, "garbage", internal.HostEndianness)
	require.EqualError(t, err, `cannot determine length for type "garbage"`)

	_, err = convertNumericType(testdata, "uint8", internal.HostEndianness)
	require.EqualError(t, err, `too many bytes 4 vs 1`)
}
