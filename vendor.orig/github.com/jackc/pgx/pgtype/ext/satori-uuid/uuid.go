package uuid

import (
	"database/sql/driver"

	"github.com/pkg/errors"

	"github.com/jackc/pgx/pgtype"
	uuid "github.com/satori/go.uuid"
)

var errUndefined = errors.New("cannot encode status undefined")

type UUID struct {
	UUID   uuid.UUID
	Status pgtype.Status
}

func (dst *UUID) Set(src interface{}) error {
	switch value := src.(type) {
	case uuid.UUID:
		*dst = UUID{UUID: value, Status: pgtype.Present}
	case [16]byte:
		*dst = UUID{UUID: uuid.UUID(value), Status: pgtype.Present}
	case []byte:
		if len(value) != 16 {
			return errors.Errorf("[]byte must be 16 bytes to convert to UUID: %d", len(value))
		}
		*dst = UUID{Status: pgtype.Present}
		copy(dst.UUID[:], value)
	case string:
		uuid, err := uuid.FromString(value)
		if err != nil {
			return err
		}
		*dst = UUID{UUID: uuid, Status: pgtype.Present}
	default:
		// If all else fails see if pgtype.UUID can handle it. If so, translate through that.
		pgUUID := &pgtype.UUID{}
		if err := pgUUID.Set(value); err != nil {
			return errors.Errorf("cannot convert %v to UUID", value)
		}

		*dst = UUID{UUID: uuid.UUID(pgUUID.Bytes), Status: pgUUID.Status}
	}

	return nil
}

func (dst *UUID) Get() interface{} {
	switch dst.Status {
	case pgtype.Present:
		return dst.UUID
	case pgtype.Null:
		return nil
	default:
		return dst.Status
	}
}

func (src *UUID) AssignTo(dst interface{}) error {
	switch src.Status {
	case pgtype.Present:
		switch v := dst.(type) {
		case *uuid.UUID:
			*v = src.UUID
		case *[16]byte:
			*v = [16]byte(src.UUID)
			return nil
		case *[]byte:
			*v = make([]byte, 16)
			copy(*v, src.UUID[:])
			return nil
		case *string:
			*v = src.UUID.String()
			return nil
		default:
			if nextDst, retry := pgtype.GetAssignToDstType(v); retry {
				return src.AssignTo(nextDst)
			}
		}
	case pgtype.Null:
		return pgtype.NullAssignTo(dst)
	}

	return errors.Errorf("cannot assign %v into %T", src, dst)
}

func (dst *UUID) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = UUID{Status: pgtype.Null}
		return nil
	}

	u, err := uuid.FromString(string(src))
	if err != nil {
		return err
	}

	*dst = UUID{UUID: u, Status: pgtype.Present}
	return nil
}

func (dst *UUID) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = UUID{Status: pgtype.Null}
		return nil
	}

	if len(src) != 16 {
		return errors.Errorf("invalid length for UUID: %v", len(src))
	}

	*dst = UUID{Status: pgtype.Present}
	copy(dst.UUID[:], src)
	return nil
}

func (src *UUID) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	return append(buf, src.UUID.String()...), nil
}

func (src *UUID) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	return append(buf, src.UUID[:]...), nil
}

// Scan implements the database/sql Scanner interface.
func (dst *UUID) Scan(src interface{}) error {
	if src == nil {
		*dst = UUID{Status: pgtype.Null}
		return nil
	}

	switch src := src.(type) {
	case string:
		return dst.DecodeText(nil, []byte(src))
	case []byte:
		return dst.DecodeText(nil, src)
	}

	return errors.Errorf("cannot scan %T", src)
}

// Value implements the database/sql/driver Valuer interface.
func (src *UUID) Value() (driver.Value, error) {
	return pgtype.EncodeValueText(src)
}
