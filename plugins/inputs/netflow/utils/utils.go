package utils

import (
	"bytes"
	"encoding/binary"
)

func ReadIntFromByteArray(byteArray []byte, size uint16) interface{} {
	var byteBuf *bytes.Buffer
	byteBuf = bytes.NewBuffer(byteArray)
	switch size {
	case 1:
		var value uint8
		binary.Read(byteBuf, binary.BigEndian, &value)
		return value
	case 2:
		var value uint16
		binary.Read(byteBuf, binary.BigEndian, &value)
		return value
	case 3:
		var value1 uint16
		var value2 uint8
		binary.Read(byteBuf, binary.BigEndian, &value1)
		binary.Read(byteBuf, binary.BigEndian, &value2)
		var value uint32 = uint32(value1)<<8 + uint32(value2)
		return value
	case 4:
		var value uint32
		binary.Read(byteBuf, binary.BigEndian, &value)
		return value
	case 5:
		var value1 uint32
		var value2 uint8
		binary.Read(byteBuf, binary.BigEndian, &value1)
		binary.Read(byteBuf, binary.BigEndian, &value2)
		var value uint64 = uint64(value1)<<8 + uint64(value2)
		return value
	case 6:
		var value1 uint32
		var value2 uint16
		binary.Read(byteBuf, binary.BigEndian, &value1)
		binary.Read(byteBuf, binary.BigEndian, &value2)
		var value uint64 = uint64(value1)<<16 + uint64(value2)
		return value
	case 7:
		var value1 uint32
		var value2 uint16
		var value3 uint8
		binary.Read(byteBuf, binary.BigEndian, &value1)
		binary.Read(byteBuf, binary.BigEndian, &value2)
		binary.Read(byteBuf, binary.BigEndian, &value3)
		var value uint64 = uint64(value1)<<24 + uint64(value2)<<8 + uint64(value3)
		return value
	case 8:
		var value uint64
		binary.Read(byteBuf, binary.BigEndian, &value)
		return value
	default:
		return 0
	}
}
