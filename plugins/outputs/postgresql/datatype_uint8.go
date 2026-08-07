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
	"database/sql/driver"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/jackc/pgio"
	"github.com/jackc/pgx/v5/pgtype"
)

type uint64Scanner interface {
	ScanUint64(Uint8) error
}

type uint64Valuer interface {
	Uint64Value() (Uint8, error)
}

type Uint8 struct {
	Uint64 uint64
	Valid  bool
}

// ScanUint64 implements the [uint64Scanner] interface.
func (dst *Uint8) ScanUint64(n Uint8) error {
	if !n.Valid {
		*dst = Uint8{}
		return nil
	}

	if n.Uint64 > math.MaxUint8 {
		return fmt.Errorf("%d is greater than maximum value for Uint8", n.Uint64)
	}
	*dst = Uint8{Uint64: n.Uint64, Valid: true}

	return nil
}

// Uint64Value implements the [uint64Valuer] interface.
func (dst Uint8) Uint64Value() (Uint8, error) {
	return Uint8{Uint64: dst.Uint64, Valid: dst.Valid}, nil
}

// Scan implements the [database/sql.Scanner] interface.
func (dst *Uint8) Scan(src any) error {
	if src == nil {
		*dst = Uint8{}
		return nil
	}

	var n uint64

	switch src := src.(type) {
	case uint64:
		n = src
	case string:
		var err error
		n, err = strconv.ParseUint(src, 10, 64)
		if err != nil {
			return err
		}
	case []byte:
		var err error
		n, err = strconv.ParseUint(string(src), 10, 64)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("cannot scan %T", src)
	}

	if n > math.MaxUint8 {
		return fmt.Errorf("%d is greater than maximum value for Uint8", n)
	}
	*dst = Uint8{Uint64: n, Valid: true}

	return nil
}

// Value implements the [database/sql/driver.Valuer] interface.
func (dst Uint8) Value() (driver.Value, error) {
	if !dst.Valid {
		return nil, nil
	}
	return dst.Uint64, nil
}

// MarshalJSON implements the [encoding/json.Marshaler] interface.
func (dst Uint8) MarshalJSON() ([]byte, error) {
	if !dst.Valid {
		return []byte("null"), nil
	}
	return []byte(strconv.FormatUint(dst.Uint64, 10)), nil
}

// UnmarshalJSON implements the [encoding/json.Unmarshaler] interface.
func (dst *Uint8) UnmarshalJSON(b []byte) error {
	var n *uint64
	err := json.Unmarshal(b, &n)
	if err != nil {
		return err
	}

	if n == nil {
		*dst = Uint8{}
	} else {
		*dst = Uint8{Uint64: *n, Valid: true}
	}

	return nil
}

type Uint8Codec struct{}

func (Uint8Codec) FormatSupported(format int16) bool {
	return format == pgtype.TextFormatCode || format == pgtype.BinaryFormatCode
}

func (Uint8Codec) PreferredFormat() int16 {
	return pgtype.BinaryFormatCode
}

func (Uint8Codec) PlanEncode(_ *pgtype.Map, _ uint32, format int16, value any) pgtype.EncodePlan {
	switch format {
	case pgtype.BinaryFormatCode:
		switch value.(type) {
		case uint64:
			return encodePlanUint8CodecBinaryUint64{}
		case uint64Valuer:
			return encodePlanUint8CodecBinaryUint64Valuer{}
		}
	case pgtype.TextFormatCode:
		switch value.(type) {
		case uint64:
			return encodePlanUint8CodecTextUint64{}
		case uint64Valuer:
			return encodePlanUint8CodecTextUint64Valuer{}
		}
	}

	return nil
}

type encodePlanUint8CodecBinaryUint64 struct{}

func (encodePlanUint8CodecBinaryUint64) Encode(value any, buf []byte) (newBuf []byte, err error) {
	return pgio.AppendUint64(buf, value.(uint64)), nil
}

type encodePlanUint8CodecTextUint64 struct{}

func (encodePlanUint8CodecTextUint64) Encode(value any, buf []byte) (newBuf []byte, err error) {
	return append(buf, strconv.FormatUint(value.(uint64), 10)...), nil
}

type encodePlanUint8CodecBinaryUint64Valuer struct{}

func (encodePlanUint8CodecBinaryUint64Valuer) Encode(value any, buf []byte) (newBuf []byte, err error) {
	n, err := value.(uint64Valuer).Uint64Value()
	if err != nil {
		return nil, err
	}

	if !n.Valid {
		return nil, nil
	}

	if n.Uint64 > math.MaxUint8 {
		return nil, fmt.Errorf("%d is greater than maximum value for uint8", n.Uint64)
	}

	return pgio.AppendUint64(buf, n.Uint64), nil
}

type encodePlanUint8CodecTextUint64Valuer struct{}

func (encodePlanUint8CodecTextUint64Valuer) Encode(value any, buf []byte) (newBuf []byte, err error) {
	n, err := value.(uint64Valuer).Uint64Value()
	if err != nil {
		return nil, err
	}

	if !n.Valid {
		return nil, nil
	}

	if n.Uint64 > math.MaxUint8 {
		return nil, fmt.Errorf("%d is greater than maximum value for uint8", n.Uint64)
	}

	return append(buf, strconv.FormatUint(n.Uint64, 10)...), nil
}

func (Uint8Codec) PlanScan(_ *pgtype.Map, _ uint32, format int16, target any) pgtype.ScanPlan {
	switch format {
	case pgtype.BinaryFormatCode:
		switch target.(type) {
		case *int8:
			return scanPlanBinaryInt8ToInt8{}
		case *int16:
			return scanPlanBinaryInt8ToInt16{}
		case *int32:
			return scanPlanBinaryInt8ToInt32{}
		case *int64:
			return scanPlanBinaryInt8ToInt64{}
		case *int:
			return scanPlanBinaryInt8ToInt{}
		case *uint8:
			return scanPlanBinaryInt8ToUint8{}
		case *uint16:
			return scanPlanBinaryInt8ToUint16{}
		case *uint32:
			return scanPlanBinaryInt8ToUint32{}
		case *uint64:
			return scanPlanBinaryInt8ToUint64{}
		case *uint:
			return scanPlanBinaryInt8ToUint{}
		case uint64Scanner:
			return scanPlanBinaryInt8ToInt64Scanner{}
		case pgtype.TextScanner:
			return scanPlanBinaryInt8ToTextScanner{}
		}
	case pgtype.TextFormatCode:
		switch target.(type) {
		case *int8:
			return scanPlanTextAnyToInt8{}
		case *int16:
			return scanPlanTextAnyToInt16{}
		case *int32:
			return scanPlanTextAnyToInt32{}
		case *int64:
			return scanPlanTextAnyToInt64{}
		case *int:
			return scanPlanTextAnyToInt{}
		case *uint8:
			return scanPlanTextAnyToUint8{}
		case *uint16:
			return scanPlanTextAnyToUint16{}
		case *uint32:
			return scanPlanTextAnyToUint32{}
		case *uint64:
			return scanPlanTextAnyToUint64{}
		case *uint:
			return scanPlanTextAnyToUint{}
		case uint64Scanner:
			return scanPlanTextAnyToInt64Scanner{}
		}
	}

	return nil
}

func (c Uint8Codec) DecodeDatabaseSQLValue(m *pgtype.Map, oid uint32, format int16, src []byte) (driver.Value, error) {
	if src == nil {
		return nil, nil
	}

	var n int64
	err := codecScan(c, m, oid, format, src, &n)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (c Uint8Codec) DecodeValue(m *pgtype.Map, oid uint32, format int16, src []byte) (any, error) {
	if src == nil {
		return nil, nil
	}

	var n int64
	err := codecScan(c, m, oid, format, src, &n)
	if err != nil {
		return nil, err
	}
	return n, nil
}

type scanPlanBinaryInt8ToInt8 struct{}

func (scanPlanBinaryInt8ToInt8) Scan(src []byte, dst any) error {
	if src == nil {
		return fmt.Errorf("cannot scan NULL into %T", dst)
	}

	if len(src) != 8 {
		return fmt.Errorf("invalid length for int8: %v", len(src))
	}

	p, ok := (dst).(*int8)
	if !ok {
		return pgtype.ErrScanTargetTypeChanged
	}

	n := int64(binary.BigEndian.Uint64(src))
	if n < math.MinInt8 {
		return fmt.Errorf("%d is less than minimum value for int8", n)
	} else if n > math.MaxInt8 {
		return fmt.Errorf("%d is greater than maximum value for int8", n)
	}

	*p = int8(n)

	return nil
}

type scanPlanBinaryInt8ToUint8 struct{}

func (scanPlanBinaryInt8ToUint8) Scan(src []byte, dst any) error {
	if src == nil {
		return fmt.Errorf("cannot scan NULL into %T", dst)
	}

	if len(src) != 8 {
		return fmt.Errorf("invalid length for uint8: %v", len(src))
	}

	p, ok := (dst).(*uint8)
	if !ok {
		return pgtype.ErrScanTargetTypeChanged
	}

	n := int64(binary.BigEndian.Uint64(src))
	if n < 0 {
		return fmt.Errorf("%d is less than minimum value for uint8", n)
	}

	if n > math.MaxUint8 {
		return fmt.Errorf("%d is greater than maximum value for uint8", n)
	}

	*p = uint8(n)

	return nil
}

type scanPlanBinaryInt8ToInt16 struct{}

func (scanPlanBinaryInt8ToInt16) Scan(src []byte, dst any) error {
	if src == nil {
		return fmt.Errorf("cannot scan NULL into %T", dst)
	}

	if len(src) != 8 {
		return fmt.Errorf("invalid length for int8: %v", len(src))
	}

	p, ok := (dst).(*int16)
	if !ok {
		return pgtype.ErrScanTargetTypeChanged
	}

	n := int64(binary.BigEndian.Uint64(src))
	if n < math.MinInt16 {
		return fmt.Errorf("%d is less than minimum value for int16", n)
	} else if n > math.MaxInt16 {
		return fmt.Errorf("%d is greater than maximum value for int16", n)
	}

	*p = int16(n)

	return nil
}

type scanPlanBinaryInt8ToUint16 struct{}

func (scanPlanBinaryInt8ToUint16) Scan(src []byte, dst any) error {
	if src == nil {
		return fmt.Errorf("cannot scan NULL into %T", dst)
	}

	if len(src) != 8 {
		return fmt.Errorf("invalid length for uint8: %v", len(src))
	}

	p, ok := (dst).(*uint16)
	if !ok {
		return pgtype.ErrScanTargetTypeChanged
	}

	n := int64(binary.BigEndian.Uint64(src))
	if n < 0 {
		return fmt.Errorf("%d is less than minimum value for uint16", n)
	}

	if n > math.MaxUint16 {
		return fmt.Errorf("%d is greater than maximum value for uint16", n)
	}

	*p = uint16(n)

	return nil
}

type scanPlanBinaryInt8ToInt32 struct{}

func (scanPlanBinaryInt8ToInt32) Scan(src []byte, dst any) error {
	if src == nil {
		return fmt.Errorf("cannot scan NULL into %T", dst)
	}

	if len(src) != 8 {
		return fmt.Errorf("invalid length for int8: %v", len(src))
	}

	p, ok := (dst).(*int32)
	if !ok {
		return pgtype.ErrScanTargetTypeChanged
	}

	n := int64(binary.BigEndian.Uint64(src))
	if n < math.MinInt32 {
		return fmt.Errorf("%d is less than minimum value for int32", n)
	} else if n > math.MaxInt32 {
		return fmt.Errorf("%d is greater than maximum value for int32", n)
	}

	*p = int32(n)

	return nil
}

type scanPlanBinaryInt8ToUint32 struct{}

func (scanPlanBinaryInt8ToUint32) Scan(src []byte, dst any) error {
	if src == nil {
		return fmt.Errorf("cannot scan NULL into %T", dst)
	}

	if len(src) != 8 {
		return fmt.Errorf("invalid length for uint8: %v", len(src))
	}

	p, ok := (dst).(*uint32)
	if !ok {
		return pgtype.ErrScanTargetTypeChanged
	}

	n := int64(binary.BigEndian.Uint64(src))
	if n < 0 {
		return fmt.Errorf("%d is less than minimum value for uint32", n)
	}

	if n > math.MaxUint32 {
		return fmt.Errorf("%d is greater than maximum value for uint32", n)
	}

	*p = uint32(n)

	return nil
}

type scanPlanBinaryInt8ToInt64 struct{}

func (scanPlanBinaryInt8ToInt64) Scan(src []byte, dst any) error {
	if src == nil {
		return fmt.Errorf("cannot scan NULL into %T", dst)
	}

	if len(src) != 8 {
		return fmt.Errorf("invalid length for int8: %v", len(src))
	}

	p, ok := (dst).(*int64)
	if !ok {
		return pgtype.ErrScanTargetTypeChanged
	}

	*p = int64(binary.BigEndian.Uint64(src))

	return nil
}

type scanPlanBinaryInt8ToUint64 struct{}

func (scanPlanBinaryInt8ToUint64) Scan(src []byte, dst any) error {
	if src == nil {
		return fmt.Errorf("cannot scan NULL into %T", dst)
	}

	if len(src) != 8 {
		return fmt.Errorf("invalid length for uint8: %v", len(src))
	}

	p, ok := (dst).(*uint64)
	if !ok {
		return pgtype.ErrScanTargetTypeChanged
	}

	n := int64(binary.BigEndian.Uint64(src))
	if n < 0 {
		return fmt.Errorf("%d is less than minimum value for uint64", n)
	}

	*p = uint64(n)

	return nil
}

type scanPlanBinaryInt8ToInt struct{}

func (scanPlanBinaryInt8ToInt) Scan(src []byte, dst any) error {
	if src == nil {
		return fmt.Errorf("cannot scan NULL into %T", dst)
	}

	if len(src) != 8 {
		return fmt.Errorf("invalid length for int8: %v", len(src))
	}

	p, ok := (dst).(*int)
	if !ok {
		return pgtype.ErrScanTargetTypeChanged
	}

	n := int64(binary.BigEndian.Uint64(src))
	if n < math.MinInt {
		return fmt.Errorf("%d is less than minimum value for int", n)
	} else if n > math.MaxInt {
		return fmt.Errorf("%d is greater than maximum value for int", n)
	}

	*p = int(n)

	return nil
}

type scanPlanBinaryInt8ToUint struct{}

func (scanPlanBinaryInt8ToUint) Scan(src []byte, dst any) error {
	if src == nil {
		return fmt.Errorf("cannot scan NULL into %T", dst)
	}

	if len(src) != 8 {
		return fmt.Errorf("invalid length for uint8: %v", len(src))
	}

	p, ok := (dst).(*uint)
	if !ok {
		return pgtype.ErrScanTargetTypeChanged
	}

	n := int64(binary.BigEndian.Uint64(src))
	if n < 0 {
		return fmt.Errorf("%d is less than minimum value for uint", n)
	}

	if uint64(n) > math.MaxUint {
		return fmt.Errorf("%d is greater than maximum value for uint", n)
	}

	*p = uint(n)

	return nil
}

type scanPlanBinaryInt8ToInt64Scanner struct{}

func (scanPlanBinaryInt8ToInt64Scanner) Scan(src []byte, dst any) error {
	s, ok := (dst).(uint64Scanner)
	if !ok {
		return pgtype.ErrScanTargetTypeChanged
	}

	if src == nil {
		return s.ScanUint64(Uint8{})
	}

	if len(src) != 8 {
		return fmt.Errorf("invalid length for int8: %v", len(src))
	}

	n := binary.BigEndian.Uint64(src)

	return s.ScanUint64(Uint8{Uint64: n, Valid: true})
}

type scanPlanBinaryInt8ToTextScanner struct{}

func (scanPlanBinaryInt8ToTextScanner) Scan(src []byte, dst any) error {
	s, ok := (dst).(pgtype.TextScanner)
	if !ok {
		return pgtype.ErrScanTargetTypeChanged
	}

	if src == nil {
		return s.ScanText(pgtype.Text{})
	}

	if len(src) != 8 {
		return fmt.Errorf("invalid length for int8: %v", len(src))
	}

	n := int64(binary.BigEndian.Uint64(src))

	return s.ScanText(pgtype.Text{String: strconv.FormatInt(n, 10), Valid: true})
}

type scanPlanTextAnyToInt8 struct{}

func (scanPlanTextAnyToInt8) Scan(src []byte, dst any) error {
	if src == nil {
		return fmt.Errorf("cannot scan NULL into %T", dst)
	}

	p, ok := (dst).(*int8)
	if !ok {
		return pgtype.ErrScanTargetTypeChanged
	}

	n, err := strconv.ParseInt(string(src), 10, 8)
	if err != nil {
		return err
	}

	*p = int8(n)
	return nil
}

type scanPlanTextAnyToUint8 struct{}

func (scanPlanTextAnyToUint8) Scan(src []byte, dst any) error {
	if src == nil {
		return fmt.Errorf("cannot scan NULL into %T", dst)
	}

	p, ok := (dst).(*uint8)
	if !ok {
		return pgtype.ErrScanTargetTypeChanged
	}

	n, err := strconv.ParseUint(string(src), 10, 8)
	if err != nil {
		return err
	}

	*p = uint8(n)
	return nil
}

type scanPlanTextAnyToInt16 struct{}

func (scanPlanTextAnyToInt16) Scan(src []byte, dst any) error {
	if src == nil {
		return fmt.Errorf("cannot scan NULL into %T", dst)
	}

	p, ok := (dst).(*int16)
	if !ok {
		return pgtype.ErrScanTargetTypeChanged
	}

	n, err := strconv.ParseInt(string(src), 10, 16)
	if err != nil {
		return err
	}

	*p = int16(n)
	return nil
}

type scanPlanTextAnyToUint16 struct{}

func (scanPlanTextAnyToUint16) Scan(src []byte, dst any) error {
	if src == nil {
		return fmt.Errorf("cannot scan NULL into %T", dst)
	}

	p, ok := (dst).(*uint16)
	if !ok {
		return pgtype.ErrScanTargetTypeChanged
	}

	n, err := strconv.ParseUint(string(src), 10, 16)
	if err != nil {
		return err
	}

	*p = uint16(n)
	return nil
}

type scanPlanTextAnyToInt32 struct{}

func (scanPlanTextAnyToInt32) Scan(src []byte, dst any) error {
	if src == nil {
		return fmt.Errorf("cannot scan NULL into %T", dst)
	}

	p, ok := (dst).(*int32)
	if !ok {
		return pgtype.ErrScanTargetTypeChanged
	}

	n, err := strconv.ParseInt(string(src), 10, 32)
	if err != nil {
		return err
	}

	*p = int32(n)
	return nil
}

type scanPlanTextAnyToUint32 struct{}

func (scanPlanTextAnyToUint32) Scan(src []byte, dst any) error {
	if src == nil {
		return fmt.Errorf("cannot scan NULL into %T", dst)
	}

	p, ok := (dst).(*uint32)
	if !ok {
		return pgtype.ErrScanTargetTypeChanged
	}

	n, err := strconv.ParseUint(string(src), 10, 32)
	if err != nil {
		return err
	}

	*p = uint32(n)
	return nil
}

type scanPlanTextAnyToInt64 struct{}

func (scanPlanTextAnyToInt64) Scan(src []byte, dst any) error {
	if src == nil {
		return fmt.Errorf("cannot scan NULL into %T", dst)
	}

	p, ok := (dst).(*int64)
	if !ok {
		return pgtype.ErrScanTargetTypeChanged
	}

	n, err := strconv.ParseInt(string(src), 10, 64)
	if err != nil {
		return err
	}

	*p = n
	return nil
}

type scanPlanTextAnyToUint64 struct{}

func (scanPlanTextAnyToUint64) Scan(src []byte, dst any) error {
	if src == nil {
		return fmt.Errorf("cannot scan NULL into %T", dst)
	}

	p, ok := (dst).(*uint64)
	if !ok {
		return pgtype.ErrScanTargetTypeChanged
	}

	n, err := strconv.ParseUint(string(src), 10, 64)
	if err != nil {
		return err
	}

	*p = n
	return nil
}

type scanPlanTextAnyToInt struct{}

func (scanPlanTextAnyToInt) Scan(src []byte, dst any) error {
	if src == nil {
		return fmt.Errorf("cannot scan NULL into %T", dst)
	}

	p, ok := (dst).(*int)
	if !ok {
		return pgtype.ErrScanTargetTypeChanged
	}

	n, err := strconv.ParseInt(string(src), 10, 0)
	if err != nil {
		return err
	}

	*p = int(n)
	return nil
}

type scanPlanTextAnyToUint struct{}

func (scanPlanTextAnyToUint) Scan(src []byte, dst any) error {
	if src == nil {
		return fmt.Errorf("cannot scan NULL into %T", dst)
	}

	p, ok := (dst).(*uint)
	if !ok {
		return pgtype.ErrScanTargetTypeChanged
	}

	n, err := strconv.ParseUint(string(src), 10, 0)
	if err != nil {
		return err
	}

	*p = uint(n)
	return nil
}

type scanPlanTextAnyToInt64Scanner struct{}

func (scanPlanTextAnyToInt64Scanner) Scan(src []byte, dst any) error {
	s, ok := (dst).(uint64Scanner)
	if !ok {
		return pgtype.ErrScanTargetTypeChanged
	}

	if src == nil {
		return s.ScanUint64(Uint8{})
	}

	n, err := strconv.ParseUint(string(src), 10, 64)
	if err != nil {
		return err
	}

	err = s.ScanUint64(Uint8{Uint64: n, Valid: true})
	if err != nil {
		return err
	}

	return nil
}

func codecScan(codec pgtype.Codec, m *pgtype.Map, oid uint32, format int16, src []byte, dst any) error {
	scanPlan := codec.PlanScan(m, oid, format, dst)
	if scanPlan == nil {
		return errors.New("codec PlanScan did not find a plan")
	}
	return scanPlan.Scan(src, dst)
}
