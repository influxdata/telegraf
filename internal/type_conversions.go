package internal

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

var ErrOutOfRange = strconv.ErrRange

func ToFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case string:
		return strconv.ParseFloat(v, 64)
	case []byte:
		return strconv.ParseFloat(string(v), 64)
	case fmt.Stringer:
		return strconv.ParseFloat(v.String(), 64)
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case nil:
		return 0, nil
	}
	return 0, fmt.Errorf("type \"%T\" unsupported", value)
}

func ToFloat32(value interface{}) (float32, error) {
	switch v := value.(type) {
	case string:
		x, err := strconv.ParseFloat(v, 32)
		return float32(x), err
	case []byte:
		x, err := strconv.ParseFloat(string(v), 32)
		return float32(x), err
	case fmt.Stringer:
		x, err := strconv.ParseFloat(v.String(), 32)
		return float32(x), err
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case int:
		return float32(v), nil
	case int8:
		return float32(v), nil
	case int16:
		return float32(v), nil
	case int32:
		return float32(v), nil
	case int64:
		return float32(v), nil
	case uint:
		return float32(v), nil
	case uint8:
		return float32(v), nil
	case uint16:
		return float32(v), nil
	case uint32:
		return float32(v), nil
	case uint64:
		return float32(v), nil
	case float32:
		return v, nil
	case float64:
		if v < -math.MaxFloat32 || v > math.MaxFloat32 {
			return float32(v), ErrOutOfRange
		}
		return float32(v), nil
	case nil:
		return 0, nil
	}
	return 0, fmt.Errorf("type \"%T\" unsupported", value)
}

func ToUint64(value interface{}) (uint64, error) {
	switch v := value.(type) {
	case string:
		if strings.HasPrefix(v, "0x") {
			return strconv.ParseUint(strings.TrimPrefix(v, "0x"), 16, 64)
		}
		return strconv.ParseUint(v, 10, 64)
	case []byte:
		return strconv.ParseUint(string(v), 10, 64)
	case fmt.Stringer:
		return strconv.ParseUint(v.String(), 10, 64)
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case int:
		if v < 0 {
			return uint64(v), ErrOutOfRange
		}
		return uint64(v), nil
	case int8:
		if v < 0 {
			return uint64(v), ErrOutOfRange
		}
		return uint64(v), nil
	case int16:
		if v < 0 {
			return uint64(v), ErrOutOfRange
		}
		return uint64(v), nil
	case int32:
		if v < 0 {
			return uint64(v), ErrOutOfRange
		}
		return uint64(v), nil
	case int64:
		if v < 0 {
			return uint64(v), ErrOutOfRange
		}
		return uint64(v), nil
	case uint:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case uint64:
		return v, nil
	case float32:
		if v < 0 || v > math.MaxUint64 {
			return uint64(v), ErrOutOfRange
		}
		return uint64(v), nil
	case float64:
		if v < 0 || v > math.MaxUint64 {
			return uint64(v), ErrOutOfRange
		}
		return uint64(v), nil
	case nil:
		return 0, nil
	}
	return 0, fmt.Errorf("type \"%T\" unsupported", value)
}

func ToUint32(value interface{}) (uint32, error) {
	switch v := value.(type) {
	case string:
		if strings.HasPrefix(v, "0x") {
			x, err := strconv.ParseUint(strings.TrimPrefix(v, "0x"), 16, 32)
			return uint32(x), err
		}
		x, err := strconv.ParseUint(v, 10, 32)
		return uint32(x), err
	case []byte:
		x, err := strconv.ParseUint(string(v), 10, 32)
		return uint32(x), err
	case fmt.Stringer:
		x, err := strconv.ParseUint(v.String(), 10, 32)
		return uint32(x), err
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case int:
		if v < 0 || uint64(v) > math.MaxUint32 {
			return uint32(v), ErrOutOfRange
		}
		return uint32(v), nil
	case int8:
		if v < 0 {
			return uint32(v), ErrOutOfRange
		}
		return uint32(v), nil
	case int16:
		if v < 0 {
			return uint32(v), ErrOutOfRange
		}
		return uint32(v), nil
	case int32:
		if v < 0 {
			return uint32(v), ErrOutOfRange
		}
		return uint32(v), nil
	case int64:
		if v < 0 || v > math.MaxUint32 {
			return uint32(v), ErrOutOfRange
		}
		return uint32(v), nil
	case uint:
		return uint32(v), nil
	case uint8:
		return uint32(v), nil
	case uint16:
		return uint32(v), nil
	case uint32:
		return v, nil
	case uint64:
		if v > math.MaxUint32 {
			return uint32(v), ErrOutOfRange
		}
		return uint32(v), nil
	case float32:
		if v < 0 || v > math.MaxUint32 {
			return uint32(v), ErrOutOfRange
		}
		return uint32(v), nil
	case float64:
		if v < 0 || v > math.MaxUint32 {
			return uint32(v), ErrOutOfRange
		}
		return uint32(v), nil
	case nil:
		return 0, nil
	}
	return 0, fmt.Errorf("type \"%T\" unsupported", value)
}

func ToUint16(value interface{}) (uint16, error) {
	switch v := value.(type) {
	case string:
		if strings.HasPrefix(v, "0x") {
			x, err := strconv.ParseUint(strings.TrimPrefix(v, "0x"), 16, 16)
			return uint16(x), err
		}
		x, err := strconv.ParseUint(v, 10, 32)
		return uint16(x), err
	case []byte:
		x, err := strconv.ParseUint(string(v), 10, 32)
		return uint16(x), err
	case fmt.Stringer:
		x, err := strconv.ParseUint(v.String(), 10, 32)
		return uint16(x), err
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case int:
		if v < 0 || v > math.MaxUint16 {
			return uint16(v), ErrOutOfRange
		}
		return uint16(v), nil
	case int8:
		if v < 0 {
			return uint16(v), ErrOutOfRange
		}
		return uint16(v), nil
	case int16:
		if v < 0 {
			return uint16(v), ErrOutOfRange
		}
		return uint16(v), nil
	case int32:
		if v < 0 || v > math.MaxUint16 {
			return uint16(v), ErrOutOfRange
		}
		return uint16(v), nil
	case int64:
		if v < 0 || v > math.MaxUint16 {
			return uint16(v), ErrOutOfRange
		}
		return uint16(v), nil
	case uint:
		return uint16(v), nil
	case uint8:
		return uint16(v), nil
	case uint16:
		return v, nil
	case uint32:
		if v > math.MaxUint16 {
			return uint16(v), ErrOutOfRange
		}
		return uint16(v), nil
	case uint64:
		if v > math.MaxUint16 {
			return uint16(v), ErrOutOfRange
		}
		return uint16(v), nil
	case float32:
		if v < 0 || v > math.MaxUint16 {
			return uint16(v), ErrOutOfRange
		}
		return uint16(v), nil
	case float64:
		if v < 0 || v > math.MaxUint16 {
			return uint16(v), ErrOutOfRange
		}
		return uint16(v), nil
	case nil:
		return 0, nil
	}
	return 0, fmt.Errorf("type \"%T\" unsupported", value)
}

func ToUint8(value interface{}) (uint8, error) {
	switch v := value.(type) {
	case string:
		if strings.HasPrefix(v, "0x") {
			x, err := strconv.ParseUint(strings.TrimPrefix(v, "0x"), 16, 8)
			return uint8(x), err
		}
		x, err := strconv.ParseUint(v, 10, 32)
		return uint8(x), err
	case []byte:
		x, err := strconv.ParseUint(string(v), 10, 32)
		return uint8(x), err
	case fmt.Stringer:
		x, err := strconv.ParseUint(v.String(), 10, 32)
		return uint8(x), err
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case int:
		if v < 0 || v > math.MaxUint8 {
			return uint8(v), ErrOutOfRange
		}
		return uint8(v), nil
	case int8:
		if v < 0 {
			return uint8(v), ErrOutOfRange
		}
		return uint8(v), nil
	case int16:
		if v < 0 || v > math.MaxUint8 {
			return uint8(v), ErrOutOfRange
		}
		return uint8(v), nil
	case int32:
		if v < 0 || v > math.MaxUint8 {
			return uint8(v), ErrOutOfRange
		}
		return uint8(v), nil
	case int64:
		if v < 0 || v > math.MaxUint8 {
			return uint8(v), ErrOutOfRange
		}
		return uint8(v), nil
	case uint:
		return uint8(v), nil
	case uint8:
		return v, nil
	case uint16:
		if v > math.MaxUint8 {
			return uint8(v), ErrOutOfRange
		}
		return uint8(v), nil
	case uint32:
		if v > math.MaxUint8 {
			return uint8(v), ErrOutOfRange
		}
		return uint8(v), nil
	case uint64:
		if v > math.MaxUint8 {
			return uint8(v), ErrOutOfRange
		}
		return uint8(v), nil
	case float32:
		if v < 0 || v > math.MaxUint8 {
			return uint8(v), ErrOutOfRange
		}
		return uint8(v), nil
	case float64:
		if v < 0 || v > math.MaxUint8 {
			return uint8(v), ErrOutOfRange
		}
		return uint8(v), nil
	case nil:
		return 0, nil
	}
	return 0, fmt.Errorf("type \"%T\" unsupported", value)
}

func ToInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case string:
		if strings.HasPrefix(v, "0x") {
			return strconv.ParseInt(strings.TrimPrefix(v, "0x"), 16, 64)
		}
		return strconv.ParseInt(v, 10, 64)
	case []byte:
		return strconv.ParseInt(string(v), 10, 64)
	case fmt.Stringer:
		return strconv.ParseInt(v.String(), 10, 64)
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		if uint64(v) > math.MaxInt64 {
			return int64(v), ErrOutOfRange
		}
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		if v > math.MaxInt64 {
			return int64(v), ErrOutOfRange
		}
		return int64(v), nil
	case float32:
		if v < math.MinInt64 || v > math.MaxInt64 {
			return int64(v), ErrOutOfRange
		}
		return int64(v), nil
	case float64:
		if v < math.MinInt64 || v > math.MaxInt64 {
			return int64(v), ErrOutOfRange
		}
		return int64(v), nil
	case nil:
		return 0, nil
	}
	return 0, fmt.Errorf("type \"%T\" unsupported", value)
}

func ToInt32(value interface{}) (int32, error) {
	switch v := value.(type) {
	case string:
		if strings.HasPrefix(v, "0x") {
			x, err := strconv.ParseInt(strings.TrimPrefix(v, "0x"), 16, 32)
			return int32(x), err
		}
		x, err := strconv.ParseInt(v, 10, 32)
		return int32(x), err
	case []byte:
		x, err := strconv.ParseInt(string(v), 10, 32)
		return int32(x), err
	case fmt.Stringer:
		x, err := strconv.ParseInt(v.String(), 10, 32)
		return int32(x), err
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case int:
		if int64(v) < math.MinInt32 || int64(v) > math.MaxInt32 {
			return int32(v), ErrOutOfRange
		}
		return int32(v), nil
	case int8:
		return int32(v), nil
	case int16:
		return int32(v), nil
	case int32:
		return v, nil
	case int64:
		if v < math.MinInt32 || v > math.MaxInt32 {
			return int32(v), ErrOutOfRange
		}
		return int32(v), nil
	case uint:
		if v > math.MaxInt32 {
			return int32(v), ErrOutOfRange
		}
		return int32(v), nil
	case uint8:
		return int32(v), nil
	case uint16:
		return int32(v), nil
	case uint32:
		if v > math.MaxInt32 {
			return int32(v), ErrOutOfRange
		}
		return int32(v), nil
	case uint64:
		if v > math.MaxInt32 {
			return int32(v), ErrOutOfRange
		}
		return int32(v), nil
	case float32:
		if v < math.MinInt32 || v > math.MaxInt32 {
			return int32(v), ErrOutOfRange
		}
		return int32(v), nil
	case float64:
		if v < math.MinInt32 || v > math.MaxInt32 {
			return int32(v), ErrOutOfRange
		}
		return int32(v), nil
	case nil:
		return 0, nil
	}
	return 0, fmt.Errorf("type \"%T\" unsupported", value)
}

func ToInt16(value interface{}) (int16, error) {
	switch v := value.(type) {
	case string:
		if strings.HasPrefix(v, "0x") {
			x, err := strconv.ParseInt(strings.TrimPrefix(v, "0x"), 16, 16)
			return int16(x), err
		}
		x, err := strconv.ParseInt(v, 10, 32)
		return int16(x), err
	case []byte:
		x, err := strconv.ParseInt(string(v), 10, 32)
		return int16(x), err
	case fmt.Stringer:
		x, err := strconv.ParseInt(v.String(), 10, 32)
		return int16(x), err
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case int:
		return int16(v), nil
	case int8:
		return int16(v), nil
	case int16:
		return v, nil
	case int32:
		if v < math.MinInt16 || v > math.MaxInt16 {
			return int16(v), ErrOutOfRange
		}
		return int16(v), nil
	case int64:
		if v < math.MinInt16 || v > math.MaxInt16 {
			return int16(v), ErrOutOfRange
		}
		return int16(v), nil
	case uint:
		if v > math.MaxInt16 {
			return int16(v), ErrOutOfRange
		}
		return int16(v), nil
	case uint8:
		return int16(v), nil
	case uint16:
		if v > math.MaxInt16 {
			return int16(v), ErrOutOfRange
		}
		return int16(v), nil
	case uint32:
		if v > math.MaxInt16 {
			return int16(v), ErrOutOfRange
		}
		return int16(v), nil
	case uint64:
		if v > math.MaxInt16 {
			return int16(v), ErrOutOfRange
		}
		return int16(v), nil
	case float32:
		if v < math.MinInt16 || v > math.MaxInt16 {
			return int16(v), ErrOutOfRange
		}
		return int16(v), nil
	case float64:
		if v < math.MinInt16 || v > math.MaxInt16 {
			return int16(v), ErrOutOfRange
		}
		return int16(v), nil
	case nil:
		return 0, nil
	}
	return 0, fmt.Errorf("type \"%T\" unsupported", value)
}

func ToInt8(value interface{}) (int8, error) {
	switch v := value.(type) {
	case string:
		if strings.HasPrefix(v, "0x") {
			x, err := strconv.ParseInt(strings.TrimPrefix(v, "0x"), 16, 8)
			return int8(x), err
		}
		x, err := strconv.ParseInt(v, 10, 32)
		return int8(x), err
	case []byte:
		x, err := strconv.ParseInt(string(v), 10, 32)
		return int8(x), err
	case fmt.Stringer:
		x, err := strconv.ParseInt(v.String(), 10, 32)
		return int8(x), err
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case int:
		return int8(v), nil
	case int8:
		return v, nil
	case int16:
		if v < math.MinInt8 || v > math.MaxInt8 {
			return int8(v), ErrOutOfRange
		}
		return int8(v), nil
	case int32:
		if v < math.MinInt8 || v > math.MaxInt8 {
			return int8(v), ErrOutOfRange
		}
		return int8(v), nil
	case int64:
		if v < math.MinInt8 || v > math.MaxInt8 {
			return int8(v), ErrOutOfRange
		}
		return int8(v), nil
	case uint:
		if v > math.MaxInt8 {
			return int8(v), ErrOutOfRange
		}
		return int8(v), nil
	case uint8:
		if v > math.MaxInt8 {
			return int8(v), ErrOutOfRange
		}
		return int8(v), nil
	case uint16:
		if v > math.MaxInt8 {
			return int8(v), ErrOutOfRange
		}
		return int8(v), nil
	case uint32:
		if v > math.MaxInt8 {
			return int8(v), ErrOutOfRange
		}
		return int8(v), nil
	case uint64:
		if v > math.MaxInt8 {
			return int8(v), ErrOutOfRange
		}
		return int8(v), nil
	case float32:
		if v < math.MinInt8 || v > math.MaxInt8 {
			return int8(v), ErrOutOfRange
		}
		return int8(v), nil
	case float64:
		if v < math.MinInt8 || v > math.MaxInt8 {
			return int8(v), ErrOutOfRange
		}
		return int8(v), nil
	case nil:
		return 0, nil
	}
	return 0, fmt.Errorf("type \"%T\" unsupported", value)
}

func ToString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case int:
		return strconv.FormatInt(int64(v), 10), nil
	case int8:
		return strconv.FormatInt(int64(v), 10), nil
	case int16:
		return strconv.FormatInt(int64(v), 10), nil
	case int32:
		return strconv.FormatInt(int64(v), 10), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case uint:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(v), nil
	case fmt.Stringer:
		return v.String(), nil
	case nil:
		return "", nil
	}
	return "", fmt.Errorf("type \"%T\" unsupported", value)
}

func ToBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case string:
		return strconv.ParseBool(v)
	case []byte:
		return strconv.ParseBool(string(v))
	case fmt.Stringer:
		return strconv.ParseBool(v.String())
	case int:
		return v > 0, nil
	case int8:
		return v > 0, nil
	case int16:
		return v > 0, nil
	case int32:
		return v > 0, nil
	case int64:
		return v > 0, nil
	case uint:
		return v > 0, nil
	case uint8:
		return v > 0, nil
	case uint16:
		return v > 0, nil
	case uint32:
		return v > 0, nil
	case uint64:
		return v > 0, nil
	case float32:
		return v > 0, nil
	case float64:
		return v > 0, nil
	case bool:
		return v, nil
	case nil:
		return false, nil
	}
	return false, fmt.Errorf("type \"%T\" unsupported", value)
}
