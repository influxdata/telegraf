package bindata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestAllFieldsOrderedBigEndian(t *testing.T) {

	var allTypes = BinData{
		MetricName: "all_types_be",
		Endiannes:  "be",
		TimeFormat: "unix",
		Fields: []Field{
			Field{Name: "fieldBool0", Type: "bool", Offset: 0, Size: 1},
			Field{Name: "fieldBool1", Type: "bool", Offset: 1, Size: 1},
			Field{Name: "fieldUint8", Type: "uint8", Offset: 2, Size: 1},
			Field{Name: "fieldInt8", Type: "int8", Offset: 3, Size: 1},
			Field{Name: "fieldUint16", Type: "uint16", Offset: 4, Size: 2},
			Field{Name: "fieldInt16", Type: "int16", Offset: 6, Size: 2},
			Field{Name: "fieldUint32", Type: "uint32", Offset: 8, Size: 4},
			Field{Name: "fieldInt32", Type: "int32", Offset: 12, Size: 4},
			Field{Name: "fieldUint64", Type: "uint64", Offset: 16, Size: 8},
			Field{Name: "fieldInt64", Type: "int64", Offset: 24, Size: 8},
			Field{Name: "fieldFloat32", Type: "float32", Offset: 32, Size: 4},
			Field{Name: "fieldFloat64", Type: "float64", Offset: 36, Size: 8},
			Field{Name: "fieldString", Type: "string", Offset: 44, Size: 20},
			Field{Name: "time", Type: "int32", Offset: 64, Size: 4},
		},
	}

	metrics, err := allTypes.Parse(binaryDataBigEndian)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	require.Equal(t, allTypes.MetricName, metrics[0].Name())
	assert.Equal(t, int64(0x5DA86C4C), metrics[0].Time().Unix())
	assert.Equal(t, map[string]interface{}{
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
	}, metrics[0].Fields())
}

func TestFieldsNoSize(t *testing.T) {

	var noSize = BinData{
		MetricName: "no_size",
		Endiannes:  "be",
		TimeFormat: "unix",
		Fields: []Field{
			Field{Name: "fieldBool0", Type: "bool", Offset: 0},
			Field{Name: "fieldBool1", Type: "bool", Offset: 1},
			Field{Name: "fieldUint8", Type: "uint8", Offset: 2},
			Field{Name: "fieldInt8", Type: "int8", Offset: 3},
			Field{Name: "fieldUint16", Type: "uint16", Offset: 4},
			Field{Name: "fieldInt16", Type: "int16", Offset: 6},
			Field{Name: "fieldUint32", Type: "uint32", Offset: 8},
			Field{Name: "fieldInt32", Type: "int32", Offset: 12},
			Field{Name: "fieldUint64", Type: "uint64", Offset: 16},
			Field{Name: "fieldInt64", Type: "int64", Offset: 24},
			Field{Name: "fieldFloat32", Type: "float32", Offset: 32},
			Field{Name: "fieldFloat64", Type: "float64", Offset: 36},
			Field{Name: "fieldString", Type: "string", Offset: 44, Size: 20},
			Field{Name: "time", Type: "int32", Offset: 64},
		},
	}

	metrics, err := noSize.Parse(binaryDataBigEndian)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	require.Equal(t, noSize.MetricName, metrics[0].Name())
	assert.Equal(t, int64(0x5DA86C4C), metrics[0].Time().Unix())
	assert.Equal(t, map[string]interface{}{
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
	}, metrics[0].Fields())
}

func TestAllFieldsLittleEndian(t *testing.T) {

	var allTypes = BinData{
		MetricName: "all_types_le",
		Endiannes:  "le",
		TimeFormat: "unix",
		Fields: []Field{
			Field{Name: "fieldBool0", Type: "bool", Offset: 0, Size: 1},
			Field{Name: "fieldBool1", Type: "bool", Offset: 1, Size: 1},
			Field{Name: "fieldUint8", Type: "uint8", Offset: 2, Size: 1},
			Field{Name: "fieldInt8", Type: "int8", Offset: 3, Size: 1},
			Field{Name: "fieldUint16", Type: "uint16", Offset: 4, Size: 2},
			Field{Name: "fieldInt16", Type: "int16", Offset: 6, Size: 2},
			Field{Name: "fieldUint32", Type: "uint32", Offset: 8, Size: 4},
			Field{Name: "fieldInt32", Type: "int32", Offset: 12, Size: 4},
			Field{Name: "fieldUint64", Type: "uint64", Offset: 16, Size: 8},
			Field{Name: "fieldInt64", Type: "int64", Offset: 24, Size: 8},
			Field{Name: "fieldFloat32", Type: "float32", Offset: 32, Size: 4},
			Field{Name: "fieldFloat64", Type: "float64", Offset: 36, Size: 8},
			Field{Name: "fieldString", Type: "string", Offset: 44, Size: 20},
			Field{Name: "time", Type: "int32", Offset: 64, Size: 4},
		},
	}

	metrics, err := allTypes.Parse(binaryDataLittleEndian)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	require.Equal(t, allTypes.MetricName, metrics[0].Name())
	assert.Equal(t, int64(0x5DA86C4C), metrics[0].Time().Unix())
	assert.Equal(t, map[string]interface{}{
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
	}, metrics[0].Fields())
}

func TestFieldsOutOfOrder(t *testing.T) {

	var outOfOrder = BinData{
		MetricName: "out_of_order",
		Endiannes:  "be",
		TimeFormat: "unix",
		Fields: []Field{
			Field{Name: "fieldString", Type: "string", Offset: 44, Size: 20},
			Field{Name: "fieldBool1", Type: "bool", Offset: 1, Size: 1},
			Field{Name: "fieldBool0", Type: "bool", Offset: 0, Size: 1},
			Field{Name: "fieldInt8", Type: "int8", Offset: 3, Size: 1},
			Field{Name: "fieldUint8", Type: "uint8", Offset: 2, Size: 1},
			Field{Name: "fieldInt16", Type: "int16", Offset: 6, Size: 2},
			Field{Name: "fieldUint16", Type: "uint16", Offset: 4, Size: 2},
			Field{Name: "time", Type: "int32", Offset: 64, Size: 4},
			Field{Name: "fieldInt32", Type: "int32", Offset: 12, Size: 4},
			Field{Name: "fieldUint32", Type: "uint32", Offset: 8, Size: 4},
			Field{Name: "fieldInt64", Type: "int64", Offset: 24, Size: 8},
			Field{Name: "fieldUint64", Type: "uint64", Offset: 16, Size: 8},
			Field{Name: "fieldFloat64", Type: "float64", Offset: 36, Size: 8},
			Field{Name: "fieldFloat32", Type: "float32", Offset: 32, Size: 4},
		},
	}

	metrics, err := outOfOrder.Parse(binaryDataBigEndian)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	require.Equal(t, outOfOrder.MetricName, metrics[0].Name())
	assert.Equal(t, int64(0x5DA86C4C), metrics[0].Time().Unix())
	assert.Equal(t, map[string]interface{}{
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
	}, metrics[0].Fields())
}

func TestWithGaps(t *testing.T) {

	var withGaps = BinData{
		MetricName: "with_gaps",
		Endiannes:  "be",
		TimeFormat: "unix",
		Fields: []Field{
			Field{Name: "fieldString", Type: "string", Offset: 44, Size: 20},
			Field{Name: "fieldBool1", Type: "bool", Offset: 1, Size: 1},
			Field{Name: "fieldInt8", Type: "int8", Offset: 3, Size: 1},
			Field{Name: "fieldUint8", Type: "uint8", Offset: 2, Size: 1},
			Field{Name: "time", Type: "int32", Offset: 64, Size: 4},
			Field{Name: "fieldInt32", Type: "int32", Offset: 12, Size: 4},
			Field{Name: "fieldUint32", Type: "uint32", Offset: 8, Size: 4},
			Field{Name: "fieldInt64", Type: "int64", Offset: 24, Size: 8},
			Field{Name: "fieldUint64", Type: "uint64", Offset: 16, Size: 8},
			Field{Name: "fieldFloat32", Type: "float32", Offset: 32, Size: 4},
		},
	}

	metrics, err := withGaps.Parse(binaryDataBigEndian)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	require.Equal(t, withGaps.MetricName, metrics[0].Name())
	assert.Equal(t, int64(0x5DA86C4C), metrics[0].Time().Unix())
	assert.Equal(t, map[string]interface{}{
		"fieldBool1":   true,
		"fieldUint8":   uint64(2),
		"fieldInt8":    int64(-2),
		"fieldUint32":  uint64(6),
		"fieldInt32":   int64(-6),
		"fieldUint64":  uint64(8),
		"fieldInt64":   int64(-8),
		"fieldFloat32": 10.5625,
		"fieldString":  "@ABCDEFGHIJKLMNOPQRS",
	}, metrics[0].Fields())
}

func TestTimeAddedByParser(t *testing.T) {

	var noTimeHere = BinData{
		MetricName: "no_time_here",
		Endiannes:  "be",
		TimeFormat: "unix",
		Fields: []Field{
			Field{Name: "fieldBool0", Type: "bool", Offset: 0, Size: 1},
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
			Field{Name: "fieldBool0", Type: "boo", Offset: 0, Size: 1},
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
			Field{Name: "fieldBool0", Type: "bool", Offset: 0, Size: 1},
			Field{Name: "time", Type: "int64", Offset: 1, Size: 8},
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
			Field{Name: "fieldBool0", Type: "bool", Offset: 0, Size: 1},
			Field{Name: "time", Type: "int64", Offset: 1, Size: 8},
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
			Field{Name: "fieldBool0", Type: "bool", Offset: 0, Size: 1},
			Field{Name: "time", Type: "int64", Offset: 1, Size: 8},
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
			Field{Name: "fieldBool0", Type: "bool", Offset: 0, Size: 1},
			Field{Name: "time", Type: "int64", Offset: 1, Size: 8},
		},
	}

	metrics, err = timeFormatInvalid.Parse(binaryDataBigEndianTimeXs)
	require.Error(t, err)
	assert.Len(t, metrics, 0)
}

func TestBinaryProtocol(t *testing.T) {

	var invalidProtocol = BinData{
		MetricName: "invalid_protocol",
		Protocol:   "invalid",
		Endiannes:  "be",
		TimeFormat: "unix",
		Fields: []Field{
			Field{Name: "fieldBool0", Type: "bool", Offset: 0, Size: 1},
		},
	}

	metrics, err := invalidProtocol.Parse(binaryDataBigEndian)
	require.Error(t, err)
	assert.Len(t, metrics, 0)

	var validProtocol = BinData{
		MetricName: "valid_protocol",
		Protocol:   "raw",
		Endiannes:  "be",
		TimeFormat: "unix",
		Fields: []Field{
			Field{Name: "fieldBool0", Type: "bool", Offset: 0, Size: 1},
		},
	}

	metrics, err = validProtocol.Parse(binaryDataBigEndian)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)

	var noProtocol = BinData{
		MetricName: "no_protocol",
		Endiannes:  "be",
		TimeFormat: "unix",
		Fields: []Field{
			Field{Name: "fieldBool0", Type: "bool", Offset: 0, Size: 1},
		},
	}

	metrics, err = noProtocol.Parse(binaryDataBigEndian)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
}
