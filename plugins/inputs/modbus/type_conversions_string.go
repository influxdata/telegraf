package modbus

import (
	"bytes"
)

func determineConverterString(byteOrder string) (fieldConverterFunc, error) {
	tohost, err := endiannessConverter16(byteOrder)
	if err != nil {
		return nil, err
	}

	return func(b []byte) interface{} {
		// Swap the bytes according to endianness
		var buf bytes.Buffer
		for i := 0; i < len(b); i += 2 {
			v := tohost(b[i : i+2])
			buf.WriteByte(byte(v >> 8))
			buf.WriteByte(byte(v & 0xFF))
		}
		// Remove everything after null-termination
		s, _ := bytes.CutSuffix(buf.Bytes(), []byte{0x00})
		return string(s)
	}, nil
}
