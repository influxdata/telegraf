package pgtype_test

import (
	"testing"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestInt8rangeTranscode(t *testing.T) {
	testutil.TestSuccessfulTranscode(t, "Int8range", []interface{}{
		&pgtype.Int8range{LowerType: pgtype.Empty, UpperType: pgtype.Empty, Status: pgtype.Present},
		&pgtype.Int8range{Lower: pgtype.Int8{Int: 1, Status: pgtype.Present}, Upper: pgtype.Int8{Int: 10, Status: pgtype.Present}, LowerType: pgtype.Inclusive, UpperType: pgtype.Exclusive, Status: pgtype.Present},
		&pgtype.Int8range{Lower: pgtype.Int8{Int: -42, Status: pgtype.Present}, Upper: pgtype.Int8{Int: -5, Status: pgtype.Present}, LowerType: pgtype.Inclusive, UpperType: pgtype.Exclusive, Status: pgtype.Present},
		&pgtype.Int8range{Lower: pgtype.Int8{Int: 1, Status: pgtype.Present}, LowerType: pgtype.Inclusive, UpperType: pgtype.Unbounded, Status: pgtype.Present},
		&pgtype.Int8range{Upper: pgtype.Int8{Int: 1, Status: pgtype.Present}, LowerType: pgtype.Unbounded, UpperType: pgtype.Exclusive, Status: pgtype.Present},
		&pgtype.Int8range{Status: pgtype.Null},
	})
}

func TestInt8rangeNormalize(t *testing.T) {
	testutil.TestSuccessfulNormalize(t, []testutil.NormalizeTest{
		{
			SQL:   "select Int8range(1, 10, '(]')",
			Value: pgtype.Int8range{Lower: pgtype.Int8{Int: 2, Status: pgtype.Present}, Upper: pgtype.Int8{Int: 11, Status: pgtype.Present}, LowerType: pgtype.Inclusive, UpperType: pgtype.Exclusive, Status: pgtype.Present},
		},
	})
}
