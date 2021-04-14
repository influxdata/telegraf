package modbus

import (
	"encoding/binary"
	"math"
)

func getBitValue(n byte, pos uint) uint16 {
	return uint16((n >> pos) & 0x01)
}

func convertDataType(t fieldContainer, bytes []byte) interface{} {
	switch t.DataType {
	case "UINT16":
		e16 := convertEndianness16(t.ByteOrder, bytes)
		return scaleUint16(t.Scale, e16)
	case "INT16":
		e16 := convertEndianness16(t.ByteOrder, bytes)
		f16 := int16(e16)
		return scaleInt16(t.Scale, f16)
	case "UINT32":
		e32 := convertEndianness32(t.ByteOrder, bytes)
		return scaleUint32(t.Scale, e32)
	case "INT32":
		e32 := convertEndianness32(t.ByteOrder, bytes)
		f32 := int32(e32)
		return scaleInt32(t.Scale, f32)
	case "UINT64":
		e64 := convertEndianness64(t.ByteOrder, bytes)
		f64 := format64(t.DataType, e64).(uint64)
		return scaleUint64(t.Scale, f64)
	case "INT64":
		e64 := convertEndianness64(t.ByteOrder, bytes)
		f64 := format64(t.DataType, e64).(int64)
		return scaleInt64(t.Scale, f64)
	case "FLOAT32-IEEE":
		e32 := convertEndianness32(t.ByteOrder, bytes)
		f32 := math.Float32frombits(e32)
		return scaleFloat32(t.Scale, f32)
	case "FLOAT64-IEEE":
		e64 := convertEndianness64(t.ByteOrder, bytes)
		f64 := math.Float64frombits(e64)
		return scaleFloat64(t.Scale, f64)
	case "FIXED":
		if len(bytes) == 2 {
			e16 := convertEndianness16(t.ByteOrder, bytes)
			f16 := int16(e16)
			return scale16toFloat(t.Scale, f16)
		} else if len(bytes) == 4 {
			e32 := convertEndianness32(t.ByteOrder, bytes)
			f32 := int32(e32)
			return scale32toFloat(t.Scale, f32)
		} else {
			e64 := convertEndianness64(t.ByteOrder, bytes)
			f64 := int64(e64)
			return scale64toFloat(t.Scale, f64)
		}
	case "FLOAT32", "UFIXED":
		if len(bytes) == 2 {
			e16 := convertEndianness16(t.ByteOrder, bytes)
			return scale16UtoFloat(t.Scale, e16)
		} else if len(bytes) == 4 {
			e32 := convertEndianness32(t.ByteOrder, bytes)
			return scale32UtoFloat(t.Scale, e32)
		} else {
			e64 := convertEndianness64(t.ByteOrder, bytes)
			return scale64UtoFloat(t.Scale, e64)
		}
	default:
		return 0
	}
}

func convertEndianness16(o string, b []byte) uint16 {
	switch o {
	case "AB":
		return binary.BigEndian.Uint16(b)
	case "BA":
		return binary.LittleEndian.Uint16(b)
	default:
		return 0
	}
}

func convertEndianness32(o string, b []byte) uint32 {
	switch o {
	case "ABCD":
		return binary.BigEndian.Uint32(b)
	case "DCBA":
		return binary.LittleEndian.Uint32(b)
	case "BADC":
		return uint32(binary.LittleEndian.Uint16(b[0:]))<<16 | uint32(binary.LittleEndian.Uint16(b[2:]))
	case "CDAB":
		return uint32(binary.BigEndian.Uint16(b[2:]))<<16 | uint32(binary.BigEndian.Uint16(b[0:]))
	default:
		return 0
	}
}

func convertEndianness64(o string, b []byte) uint64 {
	switch o {
	case "ABCDEFGH":
		return binary.BigEndian.Uint64(b)
	case "HGFEDCBA":
		return binary.LittleEndian.Uint64(b)
	case "BADCFEHG":
		return uint64(binary.LittleEndian.Uint16(b[0:]))<<48 | uint64(binary.LittleEndian.Uint16(b[2:]))<<32 | uint64(binary.LittleEndian.Uint16(b[4:]))<<16 | uint64(binary.LittleEndian.Uint16(b[6:]))
	case "GHEFCDAB":
		return uint64(binary.BigEndian.Uint16(b[6:]))<<48 | uint64(binary.BigEndian.Uint16(b[4:]))<<32 | uint64(binary.BigEndian.Uint16(b[2:]))<<16 | uint64(binary.BigEndian.Uint16(b[0:]))
	default:
		return 0
	}
}

func format64(f string, r uint64) interface{} {
	switch f {
	case "UINT64":
		return r
	case "INT64":
		return int64(r)
	default:
		return r
	}
}

func scale16toFloat(s float64, v int16) float64 {
	return float64(v) * s
}

func scale32toFloat(s float64, v int32) float64 {
	return float64(v) * s
}

func scale64toFloat(s float64, v int64) float64 {
	return float64(v) * s
}

func scale16UtoFloat(s float64, v uint16) float64 {
	return float64(v) * s
}

func scale32UtoFloat(s float64, v uint32) float64 {
	return float64(v) * s
}

func scale64UtoFloat(s float64, v uint64) float64 {
	return float64(v) * s
}

func scaleInt16(s float64, v int16) int16 {
	return int16(float64(v) * s)
}

func scaleUint16(s float64, v uint16) uint16 {
	return uint16(float64(v) * s)
}

func scaleUint32(s float64, v uint32) uint32 {
	return uint32(float64(v) * s)
}

func scaleInt32(s float64, v int32) int32 {
	return int32(float64(v) * s)
}

func scaleFloat32(s float64, v float32) float32 {
	return float32(float64(v) * s)
}

func scaleFloat64(s float64, v float64) float64 {
	return v * s
}

func scaleUint64(s float64, v uint64) uint64 {
	return uint64(float64(v) * s)
}

func scaleInt64(s float64, v int64) int64 {
	return int64(float64(v) * s)
}
