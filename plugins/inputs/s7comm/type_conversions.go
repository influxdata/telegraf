package s7comm

import (
	"encoding/binary"
	"math"

	"github.com/robinson/gos7"
)

var helper = &gos7.Helper{}

func determineConversion(dtype string, extra int) converterFunc {
	switch dtype {
	case "X":
		return func(buf []byte) interface{} {
			return (buf[0] & (1 << extra)) != 0
		}
	case "B":
		return func(buf []byte) interface{} {
			return buf[0]
		}
	case "S":
		return func(buf []byte) interface{} {
			return helper.GetStringAt(buf, 0)
		}
	case "W":
		return func(buf []byte) interface{} {
			return binary.BigEndian.Uint16(buf)
		}
	case "I":
		return func(buf []byte) interface{} {
			return int16(binary.BigEndian.Uint16(buf))
		}
	case "DW":
		return func(buf []byte) interface{} {
			return binary.BigEndian.Uint32(buf)
		}
	case "DI":
		return func(buf []byte) interface{} {
			return int32(binary.BigEndian.Uint32(buf))
		}
	case "R":
		return func(buf []byte) interface{} {
			x := binary.BigEndian.Uint32(buf)
			return math.Float32frombits(x)
		}
	case "DT":
		return func(buf []byte) interface{} {
			return helper.GetDateTimeAt(buf, 0).UnixNano()
		}
	}

	panic("unknown type")
}
