package bindata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedParseResult = map[string]interface{}{
	"fieldBool0":   false,
	"fieldBool1":   true,
	"fieldUint8":   uint64(2),
	"fieldInt8":    int64(-2),
	"fieldUint16":  uint64(4),
	"fieldInt16":   int64(-4),
	"fieldUint32":  uint64(6),
	"fieldInt32":   int64(-6),
	"fieldUint64":  uint64(8),
	"fieldInt64":   int64(-8),
	"fieldFloat32": 10.5625,
	"fieldFloat64": 11.0,
	"fieldString":  "@ABCDEFGHIJKLMNOPQRS",
}

var expectedParseResultWithOmittedFields = map[string]interface{}{
	"fieldBool1":   true,
	"fieldUint8":   uint64(2),
	"fieldInt8":    int64(-2),
	"fieldUint32":  uint64(6),
	"fieldInt32":   int64(-6),
	"fieldUint64":  uint64(8),
	"fieldInt64":   int64(-8),
	"fieldFloat32": 10.5625,
	"fieldString":  "@ABCDEFGHIJKLMNOPQRS",
}

var fields = []Field{
	Field{Name: "fieldBool0", Type: "bool"},
	Field{Name: "fieldBool1", Type: "bool"},
	Field{Name: "fieldUint8", Type: "uint8"},
	Field{Name: "fieldInt8", Type: "int8"},
	Field{Name: "fieldUint16", Type: "uint16"},
	Field{Name: "fieldInt16", Type: "int16"},
	Field{Name: "fieldUint32", Type: "uint32"},
	Field{Name: "fieldInt32", Type: "int32"},
	Field{Name: "fieldUint64", Type: "uint64"},
	Field{Name: "fieldInt64", Type: "int64"},
	Field{Name: "fieldFloat32", Type: "float32"},
	Field{Name: "fieldFloat64", Type: "float64"},
	Field{Name: "fieldString", Type: "string", Size: 20},
	Field{Name: "time", Type: "int32"},
}

var withOmittedFields = []Field{
	Field{Type: "padding", Size: 1},
	Field{Name: "fieldBool1", Type: "bool"},
	Field{Name: "fieldUint8", Type: "uint8"},
	Field{Name: "fieldInt8", Type: "int8"},
	Field{Type: "padding", Size: 4},
	Field{Name: "fieldUint32", Type: "uint32"},
	Field{Name: "fieldInt32", Type: "int32"},
	Field{Name: "fieldUint64", Type: "uint64"},
	Field{Name: "fieldInt64", Type: "int64"},
	Field{Name: "fieldFloat32", Type: "float32"},
	Field{Type: "padding", Size: 8},
	Field{Name: "fieldString", Type: "string", Size: 20},
	Field{Name: "time", Type: "int32"},
}

var binaryDataBigEndian = []byte{
	0x00, 0x01, 0x02, 0xFE, 0x00, 0x04, 0xFF, 0xFC, 0x00, 0x00, 0x00, 0x06, 0xFF, 0xFF, 0xFF, 0xFA,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x08, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xF8,
	0x41, 0x29, 0x00, 0x00, 0x40, 0x26, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40, 0x41, 0x42, 0x43,
	0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4A, 0x4B, 0x4C, 0x4D, 0x4E, 0x4F, 0x50, 0x51, 0x52, 0x53,
	0x5D, 0xA8, 0x6C, 0x4C,
}

var binaryDataLittleEndian = []byte{
	0x00, 0x01, 0x02, 0xFE, 0x04, 0x00, 0xFC, 0xFF, 0x06, 0x00, 0x00, 0x00, 0xFA, 0xFF, 0xFF, 0xFF,
	0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF8, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0x00, 0x00, 0x29, 0x41, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x26, 0x40, 0x40, 0x41, 0x42, 0x43,
	0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4A, 0x4B, 0x4C, 0x4D, 0x4E, 0x4F, 0x50, 0x51, 0x52, 0x53,
	0x4C, 0x6C, 0xA8, 0x5D,
}

var binaryDataBigEndianTimeXs = []byte{
	0x00,
	0x00, 0x00, 0x01, 0x6D, 0xD9, 0xE7, 0x0B, 0x17,
}

func TestBigEndian(t *testing.T) {

	var bigEndian = BinData{
		MetricName: "big_endian",
		Endiannes:  "be",
		TimeFormat: "unix",
		Fields:     fields,
	}

	metrics, err := bigEndian.Parse(binaryDataBigEndian)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	require.Equal(t, bigEndian.MetricName, metrics[0].Name())
	assert.Equal(t, int64(0x5DA86C4C), metrics[0].Time().Unix())
	assert.Equal(t, expectedParseResult, metrics[0].Fields())
}

func TestLittleEndian(t *testing.T) {

	var littleEndian = BinData{
		MetricName: "little_endian",
		Endiannes:  "le",
		TimeFormat: "unix",
		Fields:     fields,
	}

	metrics, err := littleEndian.Parse(binaryDataLittleEndian)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	require.Equal(t, littleEndian.MetricName, metrics[0].Name())
	assert.Equal(t, int64(0x5DA86C4C), metrics[0].Time().Unix())
	assert.Equal(t, expectedParseResult, metrics[0].Fields())
}

func TestDefaultStringEncoding(t *testing.T) {

	var defaultStringEncoding = BinData{
		MetricName:     "default_string_encoding",
		Endiannes:      "be",
		TimeFormat:     "unix",
		StringEncoding: "utf-8",
		Fields:         fields,
	}

	metrics, err := defaultStringEncoding.Parse(binaryDataBigEndian)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
}

func TestInvalidStringEncoding(t *testing.T) {

	var invalidStringEncoding = BinData{
		MetricName:     "invalid_string_encoding",
		Endiannes:      "be",
		TimeFormat:     "unix",
		StringEncoding: "utf-16",
		Fields:         fields,
	}

	metrics, err := invalidStringEncoding.Parse(binaryDataBigEndian)
	require.Error(t, err)
	assert.Len(t, metrics, 0)
}
func TestWithOmittedFields(t *testing.T) {

	var withFieldsOmitted = BinData{
		MetricName: "with_omitted_fields",
		Endiannes:  "be",
		TimeFormat: "unix",
		Fields:     withOmittedFields,
	}

	metrics, err := withFieldsOmitted.Parse(binaryDataBigEndian)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	require.Equal(t, withFieldsOmitted.MetricName, metrics[0].Name())
	assert.Equal(t, int64(0x5DA86C4C), metrics[0].Time().Unix())
	assert.Equal(t, expectedParseResultWithOmittedFields, metrics[0].Fields())
}

func TestTimeAddedByParser(t *testing.T) {

	var noTimeHere = BinData{
		MetricName: "no_time_here",
		Endiannes:  "be",
		TimeFormat: "unix",
		Fields: []Field{
			Field{Name: "fieldBool0", Type: "bool", Size: 1},
		},
	}

	metrics, err := noTimeHere.Parse(binaryDataBigEndian)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	require.Equal(t, noTimeHere.MetricName, metrics[0].Name())
	assert.True(t, int64(0x5DA86C4C) < metrics[0].Time().Unix())
	assert.Equal(t, map[string]interface{}{
		"fieldBool0": false,
	}, metrics[0].Fields())
}

func TestInvalidType(t *testing.T) {

	var invalidType = BinData{
		MetricName: "invalid_type",
		Endiannes:  "be",
		TimeFormat: "unix",
		Fields: []Field{
			Field{Name: "fieldBool0", Type: "boo", Size: 1},
		},
	}

	metrics, err := invalidType.Parse(binaryDataBigEndian)
	require.Error(t, err)
	assert.Len(t, metrics, 0)
}

func TestTimeXs(t *testing.T) {

	var timeFormatUnixMs = BinData{
		MetricName: "time_format_unix_ms",
		Endiannes:  "be",
		TimeFormat: "unix_ms",
		Fields: []Field{
			Field{Name: "fieldBool0", Type: "bool", Size: 1},
			Field{Name: "time", Type: "int64", Size: 8},
		},
	}

	metrics, err := timeFormatUnixMs.Parse(binaryDataBigEndianTimeXs)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	require.Equal(t, timeFormatUnixMs.MetricName, metrics[0].Name())
	assert.Equal(t, int64(0x0000016DD9E70B17), metrics[0].Time().UnixNano()/1000000)
	assert.Equal(t, map[string]interface{}{
		"fieldBool0": false,
	}, metrics[0].Fields())

	var timeFormatUnixUs = BinData{
		MetricName: "time_format_unix_ms",
		Endiannes:  "be",
		TimeFormat: "unix_us",
		Fields: []Field{
			Field{Name: "fieldBool0", Type: "bool", Size: 1},
			Field{Name: "time", Type: "int64", Size: 8},
		},
	}

	metrics, err = timeFormatUnixUs.Parse(binaryDataBigEndianTimeXs)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	require.Equal(t, timeFormatUnixUs.MetricName, metrics[0].Name())
	assert.Equal(t, int64(0x0000016DD9E70B17), metrics[0].Time().UnixNano()/1000)
	assert.Equal(t, map[string]interface{}{
		"fieldBool0": false,
	}, metrics[0].Fields())

	var timeFromatUnixNs = BinData{
		MetricName: "time_format_unix_ns",
		Endiannes:  "be",
		TimeFormat: "unix_ns",
		Fields: []Field{
			Field{Name: "fieldBool0", Type: "bool", Size: 1},
			Field{Name: "time", Type: "int64", Size: 8},
		},
	}

	metrics, err = timeFromatUnixNs.Parse(binaryDataBigEndianTimeXs)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	require.Equal(t, timeFromatUnixNs.MetricName, metrics[0].Name())
	assert.Equal(t, int64(0x0000016DD9E70B17), metrics[0].Time().UnixNano())
	assert.Equal(t, map[string]interface{}{
		"fieldBool0": false,
	}, metrics[0].Fields())

	var timeFormatInvalid = BinData{
		MetricName: "time_format_invalid",
		Endiannes:  "be",
		TimeFormat: "foo",
		Fields: []Field{
			Field{Name: "fieldBool0", Type: "bool", Size: 1},
			Field{Name: "time", Type: "int64", Size: 8},
		},
	}

	metrics, err = timeFormatInvalid.Parse(binaryDataBigEndianTimeXs)
	require.Error(t, err)
	assert.Len(t, metrics, 0)
}
