package binmetric

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var binaryDataBigEndian = []byte{
	0x00, 0x01, 0x02, 0x03, 0x00, 0x04, 0x00, 0x05, 0x00, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00, 0x07,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x09,
	0x41, 0x29, 0x00, 0x00, 0x40, 0x26, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40, 0x41, 0x42, 0x43,
	0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4A, 0x4B, 0x4C, 0x4D, 0x4E, 0x4F, 0x50, 0x51, 0x52, 0x53,
	0x5D, 0xA8, 0x6C, 0x4C,
}

var binaryDataLittleEndian = []byte{
	0x00, 0x01, 0x02, 0x03, 0x04, 0x00, 0x05, 0x00, 0x06, 0x00, 0x00, 0x00, 0x07, 0x00, 0x00, 0x00,
	0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x09, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x29, 0x41, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x26, 0x40, 0x40, 0x41, 0x42, 0x43,
	0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4A, 0x4B, 0x4C, 0x4D, 0x4E, 0x4F, 0x50, 0x51, 0x52, 0x53,
	0x4C, 0x6C, 0xA8, 0x5D,
}

var binaryDataBigEndianTimeMs = []byte{
	0x00,
	0x00, 0x00, 0x01, 0x6D, 0xD9, 0xE7, 0x0B, 0x17,
}

func TestAllFieldsOrderedBigEndian(t *testing.T) {

	var allTypes = BinMetric{
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
		"fieldInt8":    int64(3),
		"fieldUint16":  uint64(4),
		"fieldInt16":   int64(5),
		"fieldUint32":  uint64(6),
		"fieldInt32":   int64(7),
		"fieldUint64":  uint64(8),
		"fieldInt64":   int64(9),
		"fieldFloat32": 10.5625,
		"fieldFloat64": 11.0,
		"fieldString":  "@ABCDEFGHIJKLMNOPQRS",
	}, metrics[0].Fields())

	// fmt.Println("time", metrics[0].Time())
	// for key, value := range metrics[0].Fields() {
	// 	fmt.Println(key, value)
	// }
}

func TestAllFieldsLittleEndian(t *testing.T) {

	var allTypes = BinMetric{
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
		"fieldInt8":    int64(3),
		"fieldUint16":  uint64(4),
		"fieldInt16":   int64(5),
		"fieldUint32":  uint64(6),
		"fieldInt32":   int64(7),
		"fieldUint64":  uint64(8),
		"fieldInt64":   int64(9),
		"fieldFloat32": 10.5625,
		"fieldFloat64": 11.0,
		"fieldString":  "@ABCDEFGHIJKLMNOPQRS",
	}, metrics[0].Fields())
}

func TestFieldsOutOfOrder(t *testing.T) {

	var outOfOrder = BinMetric{
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
		"fieldInt8":    int64(3),
		"fieldUint16":  uint64(4),
		"fieldInt16":   int64(5),
		"fieldUint32":  uint64(6),
		"fieldInt32":   int64(7),
		"fieldUint64":  uint64(8),
		"fieldInt64":   int64(9),
		"fieldFloat32": 10.5625,
		"fieldFloat64": 11.0,
		"fieldString":  "@ABCDEFGHIJKLMNOPQRS",
	}, metrics[0].Fields())
}

func TestWithGaps(t *testing.T) {

	var withGaps = BinMetric{
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
		"fieldInt8":    int64(3),
		"fieldUint32":  uint64(6),
		"fieldInt32":   int64(7),
		"fieldUint64":  uint64(8),
		"fieldInt64":   int64(9),
		"fieldFloat32": 10.5625,
		"fieldString":  "@ABCDEFGHIJKLMNOPQRS",
	}, metrics[0].Fields())
}

func TestTimeAddedByParser(t *testing.T) {

	var noTimeHere = BinMetric{
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

	var invalidType = BinMetric{
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

func TestTimeMs(t *testing.T) {

	var timeMs = BinMetric{
		MetricName: "time_ms",
		Endiannes:  "be",
		TimeFormat: "unix_ms",
		Fields: []Field{
			Field{Name: "fieldBool0", Type: "bool", Offset: 0, Size: 1},
			Field{Name: "time", Type: "int64", Offset: 1, Size: 8},
		},
	}

	metrics, err := timeMs.Parse(binaryDataBigEndianTimeMs)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	require.Equal(t, timeMs.MetricName, metrics[0].Name())
	assert.Equal(t, int64(0x0000016DD9E70B17), metrics[0].Time().UnixNano()/1000000)
	assert.Equal(t, map[string]interface{}{
		"fieldBool0": false,
	}, metrics[0].Fields())

}
