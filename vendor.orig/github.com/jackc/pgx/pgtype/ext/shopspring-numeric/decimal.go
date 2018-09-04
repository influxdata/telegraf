package numeric

import (
	"database/sql/driver"
	"strconv"

	"github.com/pkg/errors"

	"github.com/jackc/pgx/pgtype"
	"github.com/shopspring/decimal"
)

var errUndefined = errors.New("cannot encode status undefined")

type Numeric struct {
	Decimal decimal.Decimal
	Status  pgtype.Status
}

func (dst *Numeric) Set(src interface{}) error {
	if src == nil {
		*dst = Numeric{Status: pgtype.Null}
		return nil
	}

	switch value := src.(type) {
	case decimal.Decimal:
		*dst = Numeric{Decimal: value, Status: pgtype.Present}
	case float32:
		*dst = Numeric{Decimal: decimal.NewFromFloat(float64(value)), Status: pgtype.Present}
	case float64:
		*dst = Numeric{Decimal: decimal.NewFromFloat(value), Status: pgtype.Present}
	case int8:
		*dst = Numeric{Decimal: decimal.New(int64(value), 0), Status: pgtype.Present}
	case uint8:
		*dst = Numeric{Decimal: decimal.New(int64(value), 0), Status: pgtype.Present}
	case int16:
		*dst = Numeric{Decimal: decimal.New(int64(value), 0), Status: pgtype.Present}
	case uint16:
		*dst = Numeric{Decimal: decimal.New(int64(value), 0), Status: pgtype.Present}
	case int32:
		*dst = Numeric{Decimal: decimal.New(int64(value), 0), Status: pgtype.Present}
	case uint32:
		*dst = Numeric{Decimal: decimal.New(int64(value), 0), Status: pgtype.Present}
	case int64:
		*dst = Numeric{Decimal: decimal.New(int64(value), 0), Status: pgtype.Present}
	case uint64:
		// uint64 could be greater than int64 so convert to string then to decimal
		dec, err := decimal.NewFromString(strconv.FormatUint(value, 10))
		if err != nil {
			return err
		}
		*dst = Numeric{Decimal: dec, Status: pgtype.Present}
	case int:
		*dst = Numeric{Decimal: decimal.New(int64(value), 0), Status: pgtype.Present}
	case uint:
		// uint could be greater than int64 so convert to string then to decimal
		dec, err := decimal.NewFromString(strconv.FormatUint(uint64(value), 10))
		if err != nil {
			return err
		}
		*dst = Numeric{Decimal: dec, Status: pgtype.Present}
	case string:
		dec, err := decimal.NewFromString(value)
		if err != nil {
			return err
		}
		*dst = Numeric{Decimal: dec, Status: pgtype.Present}
	default:
		// If all else fails see if pgtype.Numeric can handle it. If so, translate through that.
		num := &pgtype.Numeric{}
		if err := num.Set(value); err != nil {
			return errors.Errorf("cannot convert %v to Numeric", value)
		}

		buf, err := num.EncodeText(nil, nil)
		if err != nil {
			return errors.Errorf("cannot convert %v to Numeric", value)
		}

		dec, err := decimal.NewFromString(string(buf))
		if err != nil {
			return errors.Errorf("cannot convert %v to Numeric", value)
		}
		*dst = Numeric{Decimal: dec, Status: pgtype.Present}
	}

	return nil
}

func (dst *Numeric) Get() interface{} {
	switch dst.Status {
	case pgtype.Present:
		return dst.Decimal
	case pgtype.Null:
		return nil
	default:
		return dst.Status
	}
}

func (src *Numeric) AssignTo(dst interface{}) error {
	switch src.Status {
	case pgtype.Present:
		switch v := dst.(type) {
		case *decimal.Decimal:
			*v = src.Decimal
		case *float32:
			f, _ := src.Decimal.Float64()
			*v = float32(f)
		case *float64:
			f, _ := src.Decimal.Float64()
			*v = f
		case *int:
			if src.Decimal.Exponent() < 0 {
				return errors.Errorf("cannot convert %v to %T", dst, *v)
			}
			n, err := strconv.ParseInt(src.Decimal.String(), 10, strconv.IntSize)
			if err != nil {
				return errors.Errorf("cannot convert %v to %T", dst, *v)
			}
			*v = int(n)
		case *int8:
			if src.Decimal.Exponent() < 0 {
				return errors.Errorf("cannot convert %v to %T", dst, *v)
			}
			n, err := strconv.ParseInt(src.Decimal.String(), 10, 8)
			if err != nil {
				return errors.Errorf("cannot convert %v to %T", dst, *v)
			}
			*v = int8(n)
		case *int16:
			if src.Decimal.Exponent() < 0 {
				return errors.Errorf("cannot convert %v to %T", dst, *v)
			}
			n, err := strconv.ParseInt(src.Decimal.String(), 10, 16)
			if err != nil {
				return errors.Errorf("cannot convert %v to %T", dst, *v)
			}
			*v = int16(n)
		case *int32:
			if src.Decimal.Exponent() < 0 {
				return errors.Errorf("cannot convert %v to %T", dst, *v)
			}
			n, err := strconv.ParseInt(src.Decimal.String(), 10, 32)
			if err != nil {
				return errors.Errorf("cannot convert %v to %T", dst, *v)
			}
			*v = int32(n)
		case *int64:
			if src.Decimal.Exponent() < 0 {
				return errors.Errorf("cannot convert %v to %T", dst, *v)
			}
			n, err := strconv.ParseInt(src.Decimal.String(), 10, 64)
			if err != nil {
				return errors.Errorf("cannot convert %v to %T", dst, *v)
			}
			*v = int64(n)
		case *uint:
			if src.Decimal.Exponent() < 0 || src.Decimal.Sign() < 0 {
				return errors.Errorf("cannot convert %v to %T", dst, *v)
			}
			n, err := strconv.ParseUint(src.Decimal.String(), 10, strconv.IntSize)
			if err != nil {
				return errors.Errorf("cannot convert %v to %T", dst, *v)
			}
			*v = uint(n)
		case *uint8:
			if src.Decimal.Exponent() < 0 || src.Decimal.Sign() < 0 {
				return errors.Errorf("cannot convert %v to %T", dst, *v)
			}
			n, err := strconv.ParseUint(src.Decimal.String(), 10, 8)
			if err != nil {
				return errors.Errorf("cannot convert %v to %T", dst, *v)
			}
			*v = uint8(n)
		case *uint16:
			if src.Decimal.Exponent() < 0 || src.Decimal.Sign() < 0 {
				return errors.Errorf("cannot convert %v to %T", dst, *v)
			}
			n, err := strconv.ParseUint(src.Decimal.String(), 10, 16)
			if err != nil {
				return errors.Errorf("cannot convert %v to %T", dst, *v)
			}
			*v = uint16(n)
		case *uint32:
			if src.Decimal.Exponent() < 0 || src.Decimal.Sign() < 0 {
				return errors.Errorf("cannot convert %v to %T", dst, *v)
			}
			n, err := strconv.ParseUint(src.Decimal.String(), 10, 32)
			if err != nil {
				return errors.Errorf("cannot convert %v to %T", dst, *v)
			}
			*v = uint32(n)
		case *uint64:
			if src.Decimal.Exponent() < 0 || src.Decimal.Sign() < 0 {
				return errors.Errorf("cannot convert %v to %T", dst, *v)
			}
			n, err := strconv.ParseUint(src.Decimal.String(), 10, 64)
			if err != nil {
				return errors.Errorf("cannot convert %v to %T", dst, *v)
			}
			*v = uint64(n)
		default:
			if nextDst, retry := pgtype.GetAssignToDstType(dst); retry {
				return src.AssignTo(nextDst)
			}
		}
	case pgtype.Null:
		return pgtype.NullAssignTo(dst)
	}

	return nil
}

func (dst *Numeric) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = Numeric{Status: pgtype.Null}
		return nil
	}

	dec, err := decimal.NewFromString(string(src))
	if err != nil {
		return err
	}

	*dst = Numeric{Decimal: dec, Status: pgtype.Present}
	return nil
}

func (dst *Numeric) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = Numeric{Status: pgtype.Null}
		return nil
	}

	// For now at least, implement this in terms of pgtype.Numeric

	num := &pgtype.Numeric{}
	if err := num.DecodeBinary(ci, src); err != nil {
		return err
	}

	buf, err := num.EncodeText(ci, nil)
	if err != nil {
		return err
	}

	dec, err := decimal.NewFromString(string(buf))
	if err != nil {
		return err
	}

	*dst = Numeric{Decimal: dec, Status: pgtype.Present}

	return nil
}

func (src *Numeric) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	return append(buf, src.Decimal.String()...), nil
}

func (src *Numeric) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	// For now at least, implement this in terms of pgtype.Numeric
	num := &pgtype.Numeric{}
	if err := num.DecodeText(ci, []byte(src.Decimal.String())); err != nil {
		return nil, err
	}

	return num.EncodeBinary(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (dst *Numeric) Scan(src interface{}) error {
	if src == nil {
		*dst = Numeric{Status: pgtype.Null}
		return nil
	}

	switch src := src.(type) {
	case float64:
		*dst = Numeric{Decimal: decimal.NewFromFloat(src), Status: pgtype.Present}
		return nil
	case string:
		return dst.DecodeText(nil, []byte(src))
	case []byte:
		return dst.DecodeText(nil, src)
	}

	return errors.Errorf("cannot scan %T", src)
}

// Value implements the database/sql/driver Valuer interface.
func (src *Numeric) Value() (driver.Value, error) {
	switch src.Status {
	case pgtype.Present:
		return src.Decimal.Value()
	case pgtype.Null:
		return nil, nil
	default:
		return nil, errUndefined
	}
}
