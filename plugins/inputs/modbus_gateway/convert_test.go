package modbus_gateway

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"
)

func TestModbusConvertFrom(t *testing.T) {
	testInput := []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x00, 0x11}

	var tests = []struct {
		format   string
		expected []interface{}
	}{
		{
			format:   "AB", /* implies ABCD, ABCDEFGH */
			expected: []interface{}{uint16(0xAABB), uint32(0xAABBCCDD), uint64(0xAABBCCDDEEFF0011)},
		},
		{
			format:   "ABCD", /* implies AB, ABCDEFGH */
			expected: []interface{}{uint16(0xAABB), uint32(0xAABBCCDD), uint64(0xAABBCCDDEEFF0011)},
		},
		{
			format:   "BA", /* implies BADC, BADCFEHG */
			expected: []interface{}{uint16(0xBBAA), uint32(0xBBAADDCC), uint64(0xBBAADDCCFFEE1100)}},
		{
			format:   "CDAB", /* implies AB, CDABGHEF */
			expected: []interface{}{uint16(0xAABB), uint32(0xCCDDAABB)}},
		{
			format:   "DCBA", /* implies BA, DCBAHGFE */
			expected: []interface{}{uint16(0xBBAA), uint32(0xDDCCBBAA)},
		},
	}

	for _, test := range tests {
		for _, expected := range test.expected {
			order, err := CreateCustomByteOrder(test.format)
			if err != nil {
				t.Errorf("Cound not create order %s: %s", test.format, err)
			}

			reader := bytes.NewReader(testInput)

			switch expected.(type) {
			case uint16:
				var result uint16
				err = binary.Read(reader, order, &result)
				_assert(t, order, err, result == expected, expected, result)
			case uint32:
				var result uint32
				err = binary.Read(reader, order, &result)
				_assert(t, order, err, result == expected, expected, result)

			case uint64:
				var result uint64
				err = binary.Read(reader, order, &result)
				_assert(t, order, err, result == expected, expected, result)
			}
		}
	}
}

func TestModbusConvertTo(t *testing.T) {
	{
		/* uint16 ABCD */
		order, err := CreateCustomByteOrder("ABCD")
		if err != nil {
			t.Errorf("Could not create converter")
		} else {
			got := make([]byte, 2)
			order.PutUint16(got, 1)

			expected := []byte{0x00, 0x01}
			if !bytes.Equal(got, expected) {
				t.Errorf("convert to uint16 failed for format ABCD, expected %v, got %v", expected, got)
			}
		}
	}

	{
		/* uint16 DCBA */
		order, err := CreateCustomByteOrder("DCBA")
		if err != nil {
			t.Errorf("Could not create converter")
		} else {
			got := make([]byte, 2)
			order.PutUint16(got, 1)

			expected := []byte{0x01, 0x00}
			if !bytes.Equal(got, expected) {
				t.Errorf("convert to uint16 failed for format DCBA, expected %v, got %v", expected, got)
			}
		}
	}

	{
		/* uint32 ABCD */
		order, err := CreateCustomByteOrder("ABCD")
		if err != nil {
			t.Errorf("Could not create converter")
		} else {
			got := make([]byte, 4)
			order.PutUint32(got, 1)

			expected := []byte{0x00, 0x00, 0x00, 0x01}
			if !bytes.Equal(got, expected) {
				t.Errorf("convert to uint32 failed for format ABCD, expected %v, got %v", expected, got)
			}
		}
	}

	{
		/* uint32 DCBA */
		order, err := CreateCustomByteOrder("DCBA")
		if err != nil {
			t.Errorf("Could not create converter")
		} else {
			got := make([]byte, 4)
			order.PutUint32(got, 1)

			expected := []byte{0x01, 0x00, 0x00, 0x00}
			if !bytes.Equal(got, expected) {
				t.Errorf("convert to uint32 failed for format DCBA, expected %v, got %v", expected, got)
			}
		}
	}

	{
		/* uint64 ABCD */
		order, err := CreateCustomByteOrder("ABCDEFGH")
		if err != nil {
			t.Errorf("Could not create converter")
		} else {
			got := make([]byte, 8)
			order.PutUint64(got, 1)

			expected := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}
			if !bytes.Equal(got, expected) {
				t.Errorf("convert to uint64 failed for format ABCDEFGH, expected %v, got %v", expected, got)
			}
		}
	}

	{
		/* uint64 DCBA */
		order, err := CreateCustomByteOrder("HGFEDCBA")
		if err != nil {
			t.Errorf("Could not create converter")
		} else {
			got := make([]byte, 8)
			order.PutUint64(got, 0xAABBCCDD00112233)

			expected := []byte{0x33, 0x22, 0x11, 0x00, 0xDD, 0xCC, 0xBB, 0xAA}
			if !bytes.Equal(got, expected) {
				t.Errorf("convert to uint64 failed for format HGFEDCBA, expected %v, got %v", expected, got)
			}
		}
	}

}

func _assert(t *testing.T, order *CustomByteOrder, err error, success bool, expected interface{}, got interface{}) {
	if err != nil {
		t.Errorf("Test %s (%s) Error reading from stream: %s", order.order, reflect.TypeOf(expected), err)
	} else if !success {
		t.Errorf("Test %s (%s) expected 0x%08X, got 0x%08X", order.order, reflect.TypeOf(expected), expected, got)
		t.Logf("  order should have been %v", order.positions)
	} else {
		//t.Logf("Test %s (%s) PASSED", order.order, reflect.TypeOf(expected))
	}

}
