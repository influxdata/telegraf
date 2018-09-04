package pgtype_test

import (
	"math/big"
	"testing"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestNumrangeTranscode(t *testing.T) {
	testutil.TestSuccessfulTranscode(t, "numrange", []interface{}{
		&pgtype.Numrange{
			LowerType: pgtype.Empty,
			UpperType: pgtype.Empty,
			Status:    pgtype.Present,
		},
		&pgtype.Numrange{
			Lower:     pgtype.Numeric{Int: big.NewInt(-543), Exp: 3, Status: pgtype.Present},
			Upper:     pgtype.Numeric{Int: big.NewInt(342), Exp: 1, Status: pgtype.Present},
			LowerType: pgtype.Inclusive,
			UpperType: pgtype.Exclusive,
			Status:    pgtype.Present,
		},
		&pgtype.Numrange{
			Lower:     pgtype.Numeric{Int: big.NewInt(-42), Exp: 1, Status: pgtype.Present},
			Upper:     pgtype.Numeric{Int: big.NewInt(-5), Exp: 0, Status: pgtype.Present},
			LowerType: pgtype.Inclusive,
			UpperType: pgtype.Exclusive,
			Status:    pgtype.Present,
		},
		&pgtype.Numrange{
			Lower:     pgtype.Numeric{Int: big.NewInt(-42), Exp: 1, Status: pgtype.Present},
			LowerType: pgtype.Inclusive,
			UpperType: pgtype.Unbounded,
			Status:    pgtype.Present,
		},
		&pgtype.Numrange{
			Upper:     pgtype.Numeric{Int: big.NewInt(-42), Exp: 1, Status: pgtype.Present},
			LowerType: pgtype.Unbounded,
			UpperType: pgtype.Exclusive,
			Status:    pgtype.Present,
		},
		&pgtype.Numrange{Status: pgtype.Null},
	})
}
