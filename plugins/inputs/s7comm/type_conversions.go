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
	case "C":
		return func(buf []byte) interface{} {
			return string(buf[0])
		}
	case "S":
		return func(buf []byte) interface{} {
			if len(buf) <= 2 {
				return ""
			}
			// Get the length of the encoded string
			length := int(buf[0])
			// Clip the string if we do not fill the whole buffer
			if length < len(buf)-2 {
				return string(buf[2 : 2+length])
			}
			return string(buf[2:])
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

	panic("Unknown type! Please file an issue on https://github.com/influxdata/telegraf including your config.")
}
