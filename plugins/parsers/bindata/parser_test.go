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

var expectedParseResultWithPadding = map[string]interface{}{
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

var expectedParseResultStringUTF8 = map[string]interface{}{
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
	"fieldString":  "2H₂ + O₂ ⇌ 2H₂O, R = 4.7 kΩ, ⌀ 200 mm",
}

var fields = []Field{
	{Name: "fieldBool0", Type: "bool"},
	{Name: "fieldBool1", Type: "bool"},
	{Name: "fieldUint8", Type: "uint8"},
	{Name: "fieldInt8", Type: "int8"},
	{Name: "fieldUint16", Type: "uint16"},
	{Name: "fieldInt16", Type: "int16"},
	{Name: "fieldUint32", Type: "uint32"},
	{Name: "fieldInt32", Type: "int32"},
	{Name: "fieldUint64", Type: "uint64"},
	{Name: "fieldInt64", Type: "int64"},
	{Name: "fieldFloat32", Type: "float32"},
	{Name: "fieldFloat64", Type: "float64"},
	{Name: "fieldString", Type: "string", Size: 20},
	{Name: "time", Type: "int32"},
}

var fieldsWithPadding = []Field{
	{Type: "padding", Size: 1},
	{Name: "fieldBool1", Type: "bool"},
	{Name: "fieldUint8", Type: "uint8"},
	{Name: "fieldInt8", Type: "int8"},
	{Type: "padding", Size: 4},
	{Name: "fieldUint32", Type: "uint32"},
	{Name: "fieldInt32", Type: "int32"},
	{Name: "fieldUint64", Type: "uint64"},
	{Name: "fieldInt64", Type: "int64"},
	{Name: "fieldFloat32", Type: "float32"},
	{Type: "padding", Size: 8},
	{Name: "fieldString", Type: "string", Size: 20},
	{Name: "time", Type: "int32"},
}

var fieldsWithStringUTF8 = []Field{
	{Name: "fieldBool0", Type: "bool"},
	{Name: "fieldBool1", Type: "bool"},
	{Name: "fieldUint8", Type: "uint8"},
	{Name: "fieldInt8", Type: "int8"},
	{Name: "fieldUint16", Type: "uint16"},
	{Name: "fieldInt16", Type: "int16"},
	{Name: "fieldUint32", Type: "uint32"},
	{Name: "fieldInt32", Type: "int32"},
	{Name: "fieldUint64", Type: "uint64"},
	{Name: "fieldInt64", Type: "int64"},
	{Name: "fieldFloat32", Type: "float32"},
	{Name: "fieldFloat64", Type: "float64"},
	{Name: "fieldString", Type: "string", Size: 48},
	{Name: "time", Type: "int32"},
}

var fieldsWithDuplicateNames = []Field{
	{Name: "fieldFoo", Type: "bool"},
	{Name: "fieldBool1", Type: "bool"},
	{Name: "fieldUint8", Type: "uint8"},
	{Name: "fieldInt8", Type: "int8"},
	{Name: "fieldUint16", Type: "uint16"},
	{Name: "fieldInt16", Type: "int16"},
	{Name: "fieldUint32", Type: "uint32"},
	{Name: "fieldInt32", Type: "int32"},
	{Name: "fieldUint64", Type: "uint64"},
	{Name: "fieldInt64", Type: "int64"},
	{Name: "fieldFloat32", Type: "float32"},
	{Name: "fieldFloat64", Type: "float64"},
	{Name: "fieldFoo", Type: "string", Size: 48},
	{Name: "time", Type: "int32"},
}
var defaultTags = map[string]string{
	"tag0": "value0",
	"tag1": "value1",
	"tag2": "value2",
	"tag3": "value3",
}

var binaryDataBigEndian = []byte{
	0x00,
	0x01,
	0x02,
	0xFE,
	0x00, 0x04,
	0xFF, 0xFC,
	0x00, 0x00, 0x00, 0x06,
	0xFF, 0xFF, 0xFF, 0xFA,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x08,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xF8,
	0x41, 0x29, 0x00, 0x00,
	0x40, 0x26, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47,
	0x48, 0x49, 0x4A, 0x4B, 0x4C, 0x4D, 0x4E, 0x4F,
	0x50, 0x51, 0x52, 0x53,
	0x5D, 0xA8, 0x6C, 0x4C,
}

var binaryDataLittleEndian = []byte{
	0x00,
	0x01,
	0x02,
	0xFE,
	0x04, 0x00,
	0xFC, 0xFF,
	0x06, 0x00, 0x00, 0x00,
	0xFA, 0xFF, 0xFF, 0xFF,
	0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0xF8, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0x00, 0x00, 0x29, 0x41,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x26, 0x40,
	0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47,
	0x48, 0x49, 0x4A, 0x4B, 0x4C, 0x4D, 0x4E, 0x4F,
	0x50, 0x51, 0x52, 0x53,
	0x4C, 0x6C, 0xA8, 0x5D,
}

var binaryDataStringUTF8 = []byte{
	0x00,
	0x01,
	0x02,
	0xFE,
	0x04, 0x00,
	0xFC, 0xFF,
	0x06, 0x00, 0x00, 0x00,
	0xFA, 0xFF, 0xFF, 0xFF,
	0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0xF8, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0x00, 0x00, 0x29, 0x41,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x26, 0x40,
	0x32, 0x48, 0xe2, 0x82, 0x82, 0x20, 0x2b, 0x20,
	0x4f, 0xe2, 0x82, 0x82, 0x20, 0xe2, 0x87, 0x8c,
	0x20, 0x32, 0x48, 0xe2, 0x82, 0x82, 0x4f, 0x2c,
	0x20, 0x52, 0x20, 0x3d, 0x20, 0x34, 0x2e, 0x37,
	0x20, 0x6b, 0xce, 0xa9, 0x2c, 0x20, 0xe2, 0x8c,
	0x80, 0x20, 0x32, 0x30, 0x30, 0x20, 0x6d, 0x6d,
	0x4C, 0x6C, 0xA8, 0x5D,
}

var binaryDataBigEndianTimeXs = []byte{
	0x00, 0x00, 0x01, 0x6D, 0xD9, 0xE7, 0x0B, 0x17,
}

func TestBigEndian(t *testing.T) {

	var parser, err = NewBinDataParser(
		"big_endian",
		"unix",
		"be",
		"utf-8",
		fields,
		nil,
	)
	require.NoError(t, err)
	assert.NotNil(t, parser)

	metrics, err := parser.Parse(binaryDataBigEndian)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	require.Equal(t, parser.MetricName, metrics[0].Name())
	assert.Equal(t, int64(0x5DA86C4C), metrics[0].Time().Unix())
	assert.Equal(t, expectedParseResult, metrics[0].Fields())
}

func TestLittleEndian(t *testing.T) {

	var parser, err = NewBinDataParser(
		"little_endian",
		"unix",
		"le",
		"utf-8",
		fields,
		nil,
	)
	require.NoError(t, err)
	assert.NotNil(t, parser)

	metrics, err := parser.Parse(binaryDataLittleEndian)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	require.Equal(t, parser.MetricName, metrics[0].Name())
	assert.Equal(t, int64(0x5DA86C4C), metrics[0].Time().Unix())
	assert.Equal(t, expectedParseResult, metrics[0].Fields())
}

func TestInvalidEndianness(t *testing.T) {

	var parser, err = NewBinDataParser(
		"invalid_endiannes",
		"unix",
		"FOO",
		"utf-8",
		nil,
		nil,
	)

	assert.Nil(t, parser)
	require.Error(t, err)
}

func TestStringEncoding(t *testing.T) {

	var parser, err = NewBinDataParser(
		"default_string_encoding",
		"unix",
		"be",
		"utf-8",
		fields,
		nil,
	)
	require.NoError(t, err)
	assert.NotNil(t, parser)

	metrics, err := parser.Parse(binaryDataBigEndian)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
}

func TestInvalidStringEncoding(t *testing.T) {

	var parser, err = NewBinDataParser(
		"invalid_string_encoding",
		"unix",
		"be",
		"utf-16",
		nil,
		nil,
	)

	assert.Nil(t, parser)
	require.Error(t, err)
}

func TestStringUTF8(t *testing.T) {

	var parser, err = NewBinDataParser(
		"string_utf8",
		"unix",
		"le",
		"utf-8",
		fieldsWithStringUTF8,
		nil,
	)
	require.NoError(t, err)
	assert.NotNil(t, parser)

	metrics, err := parser.Parse(binaryDataStringUTF8)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	require.Equal(t, parser.MetricName, metrics[0].Name())
	assert.Equal(t, int64(0x5DA86C4C), metrics[0].Time().Unix())
	assert.Equal(t, expectedParseResultStringUTF8, metrics[0].Fields())
}

func TestPadding(t *testing.T) {

	var parser, err = NewBinDataParser(
		"padding",
		"unix",
		"be",
		"utf-8",
		fieldsWithPadding,
		nil,
	)
	require.NoError(t, err)
	assert.NotNil(t, parser)

	metrics, err := parser.Parse(binaryDataBigEndian)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	require.Equal(t, parser.MetricName, metrics[0].Name())
	assert.Equal(t, int64(0x5DA86C4C), metrics[0].Time().Unix())
	assert.Equal(t, expectedParseResultWithPadding, metrics[0].Fields())
}

func TestTimeAddedByParser(t *testing.T) {

	var parser, err = NewBinDataParser(
		"no_time",
		"unix",
		"be",
		"utf-8",
		[]Field{
			{Name: "fieldBool0", Type: "bool", Size: 1},
		},
		nil,
	)
	require.NoError(t, err)
	assert.NotNil(t, parser)

	metrics, err := parser.Parse(binaryDataBigEndian)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	require.Equal(t, parser.MetricName, metrics[0].Name())
	assert.True(t, int64(0x5DA86C4C) < metrics[0].Time().Unix())
	assert.Equal(t, map[string]interface{}{
		"fieldBool0": false,
	}, metrics[0].Fields())
}

func TestInvalidFieldType(t *testing.T) {

	var parser, err = NewBinDataParser(
		"no_time",
		"unix",
		"be",
		"utf-8",
		[]Field{
			{Name: "fieldBool0", Type: "FOO", Size: 1},
		},
		nil,
	)
	assert.Nil(t, parser)
	require.Error(t, err)
}

func TestTimeXs(t *testing.T) {

	var parser, err = NewBinDataParser(
		"time_unix_ms",
		"unix_ms",
		"be",
		"utf-8",
		[]Field{
			{Name: "time", Type: "int64", Size: 8},
		},
		nil,
	)
	require.NoError(t, err)
	assert.NotNil(t, parser)

	metrics, err := parser.Parse(binaryDataBigEndianTimeXs)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	require.Equal(t, parser.MetricName, metrics[0].Name())
	assert.Equal(t, int64(0x0000016DD9E70B17), metrics[0].Time().UnixNano()/1000000)

	parser, err = NewBinDataParser(
		"time_unix_us",
		"unix_us",
		"be",
		"utf-8",
		[]Field{
			{Name: "time", Type: "int64", Size: 8},
		},
		nil,
	)
	require.NoError(t, err)
	assert.NotNil(t, parser)

	metrics, err = parser.Parse(binaryDataBigEndianTimeXs)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	require.Equal(t, parser.MetricName, metrics[0].Name())
	assert.Equal(t, int64(0x0000016DD9E70B17), metrics[0].Time().UnixNano()/1000)

	parser, err = NewBinDataParser(
		"time_unix_ns",
		"unix_ns",
		"be",
		"utf-8",
		[]Field{
			{Name: "time", Type: "int64", Size: 8},
		},
		nil,
	)
	require.NoError(t, err)
	assert.NotNil(t, parser)

	metrics, err = parser.Parse(binaryDataBigEndianTimeXs)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	require.Equal(t, parser.MetricName, metrics[0].Name())
	assert.Equal(t, int64(0x0000016DD9E70B17), metrics[0].Time().UnixNano())

	parser, err = NewBinDataParser(
		"time_invalid_format",
		"FOO",
		"be",
		"utf-8",
		[]Field{
			{Name: "time", Type: "int64", Size: 8},
		},
		nil,
	)
	assert.Nil(t, parser)
	require.Error(t, err)
}

func TestDefaultTags(t *testing.T) {

	parser, err := NewBinDataParser(
		"default_tags",
		"unix",
		"be",
		"utf-8",
		fields,
		defaultTags,
	)
	require.NoError(t, err)
	assert.NotNil(t, parser)

	metrics, err := parser.Parse(binaryDataBigEndian)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	require.Equal(t, parser.MetricName, metrics[0].Name())
	assert.Equal(t, defaultTags, metrics[0].Tags())
}

func TestDuplicateNames(t *testing.T) {

	var parser, err = NewBinDataParser(
		"duplicate_names",
		"unix",
		"be",
		"utf-8",
		fieldsWithDuplicateNames,
		nil,
	)

	assert.Nil(t, parser)
	require.Error(t, err)
}
