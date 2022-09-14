// nolint
package postgresql

// Copied from https://github.com/jackc/pgtype/blob/master/int8.go and tweaked for uint64
/*
Copyright (c) 2013-2021 Jack Christensen

MIT License

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

import (
	"database/sql"
	"database/sql/driver"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/jackc/pgtype"
	"math"
	"strconv"

	"github.com/jackc/pgio"
)

var errUndefined = errors.New("cannot encode status undefined")
var errBadStatus = errors.New("invalid status")

type Uint8 struct {
	Int    uint64
	Status Status
}

func (dst *Uint8) Set(src interface{}) error {
	if src == nil {
		*dst = Uint8{Status: Null}
		return nil
	}

	if value, ok := src.(interface{ Get() interface{} }); ok {
		value2 := value.Get()
		if value2 != value {
			return dst.Set(value2)
		}
	}

	switch value := src.(type) {
	case int8:
		*dst = Uint8{Int: uint64(value), Status: Present}
	case uint8:
		*dst = Uint8{Int: uint64(value), Status: Present}
	case int16:
		*dst = Uint8{Int: uint64(value), Status: Present}
	case uint16:
		*dst = Uint8{Int: uint64(value), Status: Present}
	case int32:
		*dst = Uint8{Int: uint64(value), Status: Present}
	case uint32:
		*dst = Uint8{Int: uint64(value), Status: Present}
	case int64:
		*dst = Uint8{Int: uint64(value), Status: Present}
	case uint64:
		*dst = Uint8{Int: value, Status: Present}
	case int:
		if value < 0 {
			return fmt.Errorf("%d is less than maximum value for Uint8", value)
		}
		*dst = Uint8{Int: uint64(value), Status: Present}
	case uint:
		if uint64(value) > math.MaxInt64 {
			return fmt.Errorf("%d is greater than maximum value for Uint8", value)
		}
		*dst = Uint8{Int: uint64(value), Status: Present}
	case string:
		num, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		*dst = Uint8{Int: num, Status: Present}
	case float32:
		if value > math.MaxInt64 {
			return fmt.Errorf("%f is greater than maximum value for Uint8", value)
		}
		*dst = Uint8{Int: uint64(value), Status: Present}
	case float64:
		if value > math.MaxInt64 {
			return fmt.Errorf("%f is greater than maximum value for Uint8", value)
		}
		*dst = Uint8{Int: uint64(value), Status: Present}
	case *int8:
		if value == nil {
			*dst = Uint8{Status: Null}
		} else {
			return dst.Set(*value)
		}
	case *uint8:
		if value == nil {
			*dst = Uint8{Status: Null}
		} else {
			return dst.Set(*value)
		}
	case *int16:
		if value == nil {
			*dst = Uint8{Status: Null}
		} else {
			return dst.Set(*value)
		}
	case *uint16:
		if value == nil {
			*dst = Uint8{Status: Null}
		} else {
			return dst.Set(*value)
		}
	case *int32:
		if value == nil {
			*dst = Uint8{Status: Null}
		} else {
			return dst.Set(*value)
		}
	case *uint32:
		if value == nil {
			*dst = Uint8{Status: Null}
		} else {
			return dst.Set(*value)
		}
	case *int64:
		if value == nil {
			*dst = Uint8{Status: Null}
		} else {
			return dst.Set(*value)
		}
	case *uint64:
		if value == nil {
			*dst = Uint8{Status: Null}
		} else {
			return dst.Set(*value)
		}
	case *int:
		if value == nil {
			*dst = Uint8{Status: Null}
		} else {
			return dst.Set(*value)
		}
	case *uint:
		if value == nil {
			*dst = Uint8{Status: Null}
		} else {
			return dst.Set(*value)
		}
	case *string:
		if value == nil {
			*dst = Uint8{Status: Null}
		} else {
			return dst.Set(*value)
		}
	case *float32:
		if value == nil {
			*dst = Uint8{Status: Null}
		} else {
			return dst.Set(*value)
		}
	case *float64:
		if value == nil {
			*dst = Uint8{Status: Null}
		} else {
			return dst.Set(*value)
		}
	default:
		return fmt.Errorf("cannot convert %v to Uint8", value)
	}

	return nil
}

func (dst Uint8) Get() interface{} {
	switch dst.Status {
	case Present:
		return dst.Int
	case Null:
		return nil
	default:
		return dst.Status
	}
}

func (src *Uint8) AssignTo(dst interface{}) error {
	switch v := dst.(type) {
	case *int:
		*v = int(src.Int)
	case *int8:
		*v = int8(src.Int)
	case *int16:
		*v = int16(src.Int)
	case *int32:
		*v = int32(src.Int)
	case *int64:
		*v = int64(src.Int)
	case *uint:
		*v = uint(src.Int)
	case *uint8:
		*v = uint8(src.Int)
	case *uint16:
		*v = uint16(src.Int)
	case *uint32:
		*v = uint32(src.Int)
	case *uint64:
		*v = src.Int
	case *float32:
		*v = float32(src.Int)
	case *float64:
		*v = float64(src.Int)
	case *string:
		*v = strconv.FormatUint(src.Int, 10)
	case sql.Scanner:
		return v.Scan(src.Int)
	case interface{ Set(interface{}) error }:
		return v.Set(src.Int)
	default:
		return fmt.Errorf("cannot assign %v into %T", src.Int, dst)
	}
	return nil
}

func (dst *Uint8) DecodeText(ci *ConnInfo, src []byte) error {
	if src == nil {
		*dst = Uint8{Status: Null}
		return nil
	}

	n, err := strconv.ParseUint(string(src), 10, 64)
	if err != nil {
		return err
	}

	*dst = Uint8{Int: n, Status: Present}
	return nil
}

func (dst *Uint8) DecodeBinary(ci *ConnInfo, src []byte) error {
	if src == nil {
		*dst = Uint8{Status: Null}
		return nil
	}

	if len(src) != 8 {
		return fmt.Errorf("invalid length for int8: %v", len(src))
	}

	n := binary.BigEndian.Uint64(src)

	*dst = Uint8{Int: n, Status: Present}
	return nil
}

func (src Uint8) EncodeText(ci *ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case Null:
		return nil, nil
	case Undefined:
		return nil, errUndefined
	}

	return append(buf, strconv.FormatUint(src.Int, 10)...), nil
}

func (src Uint8) EncodeBinary(ci *ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case Null:
		return nil, nil
	case Undefined:
		return nil, errUndefined
	}

	return pgio.AppendUint64(buf, src.Int), nil
}

// Scan implements the database/sql Scanner interface.
func (dst *Uint8) Scan(src interface{}) error {
	if src == nil {
		*dst = Uint8{Status: Null}
		return nil
	}

	switch src := src.(type) {
	case uint64:
		*dst = Uint8{Int: src, Status: Present}
		return nil
	case string:
		return dst.DecodeText(nil, []byte(src))
	case []byte:
		srcCopy := make([]byte, len(src))
		copy(srcCopy, src)
		return dst.DecodeText(nil, srcCopy)
	}

	return fmt.Errorf("cannot scan %T", src)
}

// Value implements the database/sql/driver Valuer interface.
func (src Uint8) Value() (driver.Value, error) {
	switch src.Status {
	case Present:
		return int64(src.Int), nil
	case Null:
		return nil, nil
	default:
		return nil, errUndefined
	}
}

func (src Uint8) MarshalJSON() ([]byte, error) {
	switch src.Status {
	case Present:
		return []byte(strconv.FormatUint(src.Int, 10)), nil
	case Null:
		return []byte("null"), nil
	case Undefined:
		return nil, errUndefined
	}

	return nil, errBadStatus
}

func (dst *Uint8) UnmarshalJSON(b []byte) error {
	var n *uint64
	err := json.Unmarshal(b, &n)
	if err != nil {
		return err
	}

	if n == nil {
		*dst = Uint8{Status: Null}
	} else {
		*dst = Uint8{Int: *n, Status: Present}
	}

	return nil
}
