package binary

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"

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

func (e *Entry) convertToUint64(value interface{}, order binary.ByteOrder) ([]byte, error) {
	buf := make([]byte, 8)

	v, err := internal.ToUint64(value)
	order.PutUint64(buf, v)
	return buf, err
}

func (e *Entry) convertToUint32(value interface{}, order binary.ByteOrder) ([]byte, error) {
	buf := make([]byte, 4)

	v, err := internal.ToUint32(value)
	order.PutUint32(buf, v)
	return buf, err
}

func (e *Entry) convertToUint16(value interface{}, order binary.ByteOrder) ([]byte, error) {
	buf := make([]byte, 2)

	v, err := internal.ToUint16(value)
	order.PutUint16(buf, v)
	return buf, err
}

func (e *Entry) convertToUint8(value interface{}, _ binary.ByteOrder) ([]byte, error) {
	v, err := internal.ToUint8(value)
	return []byte{v}, err
}

func (e *Entry) convertToInt64(value interface{}, order binary.ByteOrder) ([]byte, error) {
	buf := make([]byte, 8)

	v, err := internal.ToInt64(value)
	order.PutUint64(buf, uint64(v))
	return buf, err
}

func (e *Entry) convertToInt32(value interface{}, order binary.ByteOrder) ([]byte, error) {
	buf := make([]byte, 4)

	v, err := internal.ToInt32(value)
	order.PutUint32(buf, uint32(v))
	return buf, err
}

func (e *Entry) convertToInt16(value interface{}, order binary.ByteOrder) ([]byte, error) {
	buf := make([]byte, 2)

	v, err := internal.ToInt16(value)
	order.PutUint16(buf, uint16(v))
	return buf, err
}

func (e *Entry) convertToInt8(value interface{}, _ binary.ByteOrder) ([]byte, error) {
	v, err := internal.ToInt8(value)
	return []byte{uint8(v)}, err
}

func (e *Entry) convertToFloat64(value interface{}, order binary.ByteOrder) ([]byte, error) {
	v, err := internal.ToFloat64(value)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 8)
	x := math.Float64bits(v)
	order.PutUint64(buf, x)
	return buf, nil
}

func (e *Entry) convertToFloat32(value interface{}, order binary.ByteOrder) ([]byte, error) {
	var v float32
	switch raw := value.(type) {
	case string:
		x, err := strconv.ParseFloat(raw, 32)
		if err != nil {
			return nil, err
		}
		v = float32(x)
	case []byte:
		x, err := strconv.ParseFloat(string(raw), 32)
		if err != nil {
			return nil, err
		}
		v = float32(x)
	case fmt.Stringer:
		x, err := strconv.ParseFloat(raw.String(), 32)
		if err != nil {
			return nil, err
		}
		v = float32(x)
	case bool:
		if raw {
			v = 1
		} else {
			v = 0
		}
	case int:
		v = float32(raw)
	case int8:
		v = float32(raw)
	case int16:
		v = float32(raw)
	case int32:
		v = float32(raw)
	case int64:
		v = float32(raw)
	case uint:
		v = float32(raw)
	case uint8:
		v = float32(raw)
	case uint16:
		v = float32(raw)
	case uint32:
		v = float32(raw)
	case uint64:
		v = float32(raw)
	case float32:
		v = raw
	case float64:
		if raw < -math.MaxFloat32 || raw > math.MaxFloat32 {
			return nil, fmt.Errorf("overflow/underflow with %v for field %q", raw, e.ReadFrom)
		}
		v = float32(raw)
	case nil:
		v = 0
	default:
		return nil, fmt.Errorf("type \"%T\" unsupported", value)
	}

	buf := make([]byte, 4)
	x := math.Float32bits(v)
	order.PutUint32(buf, x)

	return buf, nil
}
