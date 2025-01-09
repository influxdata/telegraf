package binary

import (
	"encoding/binary"
	"math"

	"github.com/influxdata/telegraf/internal"
)

func (e *Entry) convertToString(value interface{}, _ binary.ByteOrder) ([]byte, error) {
	v, err := internal.ToString(value)
	if err != nil {
		return nil, err
	}

	buf := []byte(v)

	// If string is longer than target length, truncate it and append terminator.
	// Thus, there is one less place for the data so that the terminator can be placed.
	if len(buf) >= int(e.StringLength) {
		dataLength := int(e.StringLength) - 1
		return append(buf[:dataLength], e.termination), nil
	}
	for i := len(buf); i < int(e.StringLength); i++ {
		buf = append(buf, e.termination)
	}
	return buf, nil
}

func convertToUint64(value interface{}, order binary.ByteOrder) ([]byte, error) {
	buf := make([]byte, 8)

	v, err := internal.ToUint64(value)
	order.PutUint64(buf, v)
	return buf, err
}

func convertToUint32(value interface{}, order binary.ByteOrder) ([]byte, error) {
	buf := make([]byte, 4)

	v, err := internal.ToUint32(value)
	order.PutUint32(buf, v)
	return buf, err
}

func convertToUint16(value interface{}, order binary.ByteOrder) ([]byte, error) {
	buf := make([]byte, 2)

	v, err := internal.ToUint16(value)
	order.PutUint16(buf, v)
	return buf, err
}

func convertToUint8(value interface{}, _ binary.ByteOrder) ([]byte, error) {
	v, err := internal.ToUint8(value)
	return []byte{v}, err
}

func convertToInt64(value interface{}, order binary.ByteOrder) ([]byte, error) {
	buf := make([]byte, 8)

	v, err := internal.ToInt64(value)
	order.PutUint64(buf, uint64(v))
	return buf, err
}

func convertToInt32(value interface{}, order binary.ByteOrder) ([]byte, error) {
	buf := make([]byte, 4)

	v, err := internal.ToInt32(value)
	order.PutUint32(buf, uint32(v))
	return buf, err
}

func convertToInt16(value interface{}, order binary.ByteOrder) ([]byte, error) {
	buf := make([]byte, 2)

	v, err := internal.ToInt16(value)
	order.PutUint16(buf, uint16(v))
	return buf, err
}

func convertToInt8(value interface{}, _ binary.ByteOrder) ([]byte, error) {
	v, err := internal.ToInt8(value)
	return []byte{uint8(v)}, err
}

func convertToFloat64(value interface{}, order binary.ByteOrder) ([]byte, error) {
	v, err := internal.ToFloat64(value)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 8)
	x := math.Float64bits(v)
	order.PutUint64(buf, x)
	return buf, nil
}

func convertToFloat32(value interface{}, order binary.ByteOrder) ([]byte, error) {
	v, err := internal.ToFloat32(value)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 4)
	x := math.Float32bits(v)
	order.PutUint32(buf, x)
	return buf, nil
}
