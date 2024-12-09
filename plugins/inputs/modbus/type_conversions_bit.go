package modbus

func determineConverterBit(byteOrder string, bit uint8) (fieldConverterFunc, error) {
	tohost, err := endiannessConverter16(byteOrder)
	if err != nil {
		return nil, err
	}

	return func(b []byte) interface{} {
		// Swap the bytes according to endianness
		v := tohost(b)
		return uint8(v >> bit & 0x01)
	}, nil
}
