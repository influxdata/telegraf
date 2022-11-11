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
	"math"
	"strconv"

	"github.com/jackc/pgio"
	"github.com/jackc/pgtype"
)

var errUndefined = errors.New("cannot encode status undefined")
var errBadStatus = errors.New("invalid status")

type Uint8 struct {
	Int    uint64
	Status pgtype.Status
}

func (u *Uint8) Set(src interface{}) error {
	if src == nil {
		*u = Uint8{Status: pgtype.Null}
		return nil
	}

	if value, ok := src.(interface{ Get() interface{} }); ok {
		value2 := value.Get()
		if value2 != value {
			return u.Set(value2)
		}
	}

	switch value := src.(type) {
	case int8:
		*u = Uint8{Int: uint64(value), Status: pgtype.Present}
	case uint8:
		*u = Uint8{Int: uint64(value), Status: pgtype.Present}
	case int16:
		*u = Uint8{Int: uint64(value), Status: pgtype.Present}
	case uint16:
		*u = Uint8{Int: uint64(value), Status: pgtype.Present}
	case int32:
		*u = Uint8{Int: uint64(value), Status: pgtype.Present}
	case uint32:
		*u = Uint8{Int: uint64(value), Status: pgtype.Present}
	case int64:
		*u = Uint8{Int: uint64(value), Status: pgtype.Present}
	case uint64:
		*u = Uint8{Int: value, Status: pgtype.Present}
	case int:
		if value < 0 {
			return fmt.Errorf("%d is less than maximum value for Uint8", value)
		}
		*u = Uint8{Int: uint64(value), Status: pgtype.Present}
	case uint:
		if uint64(value) > math.MaxInt64 {
			return fmt.Errorf("%d is greater than maximum value for Uint8", value)
		}
		*u = Uint8{Int: uint64(value), Status: pgtype.Present}
	case string:
		num, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		*u = Uint8{Int: num, Status: pgtype.Present}
	case float32:
		if value > math.MaxInt64 {
			return fmt.Errorf("%f is greater than maximum value for Uint8", value)
		}
		*u = Uint8{Int: uint64(value), Status: pgtype.Present}
	case float64:
		if value > math.MaxInt64 {
			return fmt.Errorf("%f is greater than maximum value for Uint8", value)
		}
		*u = Uint8{Int: uint64(value), Status: pgtype.Present}
	case *int8:
		if value != nil {
			return u.Set(*value)
		}
		*u = Uint8{Status: pgtype.Null}
	case *uint8:
		if value != nil {
			return u.Set(*value)
		}
		*u = Uint8{Status: pgtype.Null}
	case *int16:
		if value != nil {
			return u.Set(*value)
		}
		*u = Uint8{Status: pgtype.Null}
	case *uint16:
		if value != nil {
			return u.Set(*value)
		}
		*u = Uint8{Status: pgtype.Null}
	case *int32:
		if value != nil {
			return u.Set(*value)
		}
		*u = Uint8{Status: pgtype.Null}
	case *uint32:
		if value != nil {
			return u.Set(*value)
		}
		*u = Uint8{Status: pgtype.Null}
	case *int64:
		if value != nil {
			return u.Set(*value)
		}
		*u = Uint8{Status: pgtype.Null}
	case *uint64:
		if value != nil {
			return u.Set(*value)
		}
		*u = Uint8{Status: pgtype.Null}
	case *int:
		if value != nil {
			return u.Set(*value)
		}
		*u = Uint8{Status: pgtype.Null}
	case *uint:
		if value != nil {
			return u.Set(*value)
		}
		*u = Uint8{Status: pgtype.Null}
	case *string:
		if value != nil {
			return u.Set(*value)
		}
		*u = Uint8{Status: pgtype.Null}
	case *float32:
		if value != nil {
			return u.Set(*value)
		}
		*u = Uint8{Status: pgtype.Null}
	case *float64:
		if value != nil {
			return u.Set(*value)
		}
		*u = Uint8{Status: pgtype.Null}
	default:
		return fmt.Errorf("cannot convert %v to Uint8", value)
	}

	return nil
}

func (u *Uint8) Get() interface{} {
	switch u.Status {
	case pgtype.Present:
		return u.Int
	case pgtype.Null:
		return nil
	default:
		return u.Status
	}
}

func (u *Uint8) AssignTo(dst interface{}) error {
	switch v := dst.(type) {
	case *int:
		*v = int(u.Int)
	case *int8:
		*v = int8(u.Int)
	case *int16:
		*v = int16(u.Int)
	case *int32:
		*v = int32(u.Int)
	case *int64:
		*v = int64(u.Int)
	case *uint:
		*v = uint(u.Int)
	case *uint8:
		*v = uint8(u.Int)
	case *uint16:
		*v = uint16(u.Int)
	case *uint32:
		*v = uint32(u.Int)
	case *uint64:
		*v = u.Int
	case *float32:
		*v = float32(u.Int)
	case *float64:
		*v = float64(u.Int)
	case *string:
		*v = strconv.FormatUint(u.Int, 10)
	case sql.Scanner:
		return v.Scan(u.Int)
	case interface{ Set(interface{}) error }:
		return v.Set(u.Int)
	default:
		return fmt.Errorf("cannot assign %v into %T", u.Int, dst)
	}
	return nil
}

func (u *Uint8) DecodeText(_, src []byte) error {
	if src == nil {
		*u = Uint8{Status: pgtype.Null}
		return nil
	}

	n, err := strconv.ParseUint(string(src), 10, 64)
	if err != nil {
		return err
	}

	*u = Uint8{Int: n, Status: pgtype.Present}
	return nil
}

func (u *Uint8) DecodeBinary(_, src []byte) error {
	if src == nil {
		*u = Uint8{Status: pgtype.Null}
		return nil
	}

	if len(src) != 8 {
		return fmt.Errorf("invalid length for int8: %v", len(src))
	}

	n := binary.BigEndian.Uint64(src)

	*u = Uint8{Int: n, Status: pgtype.Present}
	return nil
}

func (u *Uint8) EncodeText(_, buf []byte) ([]byte, error) {
	switch u.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	return append(buf, strconv.FormatUint(u.Int, 10)...), nil
}

func (u *Uint8) EncodeBinary(_, buf []byte) ([]byte, error) {
	switch u.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	return pgio.AppendUint64(buf, u.Int), nil
}

// Scan implements the database/sql Scanner interface.
func (u *Uint8) Scan(src interface{}) error {
	if src == nil {
		*u = Uint8{Status: pgtype.Null}
		return nil
	}

	switch src := src.(type) {
	case uint64:
		*u = Uint8{Int: src, Status: pgtype.Present}
		return nil
	case string:
		return u.DecodeText(nil, []byte(src))
	case []byte:
		srcCopy := make([]byte, len(src))
		copy(srcCopy, src)
		return u.DecodeText(nil, srcCopy)
	}

	return fmt.Errorf("cannot scan %T", src)
}

// Value implements the database/sql/driver Valuer interface.
func (u *Uint8) Value() (driver.Value, error) {
	switch u.Status {
	case pgtype.Present:
		return int64(u.Int), nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, errUndefined
	}
}

func (u *Uint8) MarshalJSON() ([]byte, error) {
	switch u.Status {
	case pgtype.Present:
		return []byte(strconv.FormatUint(u.Int, 10)), nil
	case pgtype.Null:
		return []byte("null"), nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	return nil, errBadStatus
}

func (u *Uint8) UnmarshalJSON(b []byte) error {
	var n *uint64
	err := json.Unmarshal(b, &n)
	if err != nil {
		return err
	}

	if n == nil {
		*u = Uint8{Status: pgtype.Null}
	} else {
		*u = Uint8{Int: *n, Status: pgtype.Present}
	}

	return nil
}
