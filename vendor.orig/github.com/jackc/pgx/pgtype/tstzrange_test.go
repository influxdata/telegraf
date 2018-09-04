package pgtype_test

import (
	"testing"
	"time"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestTstzrangeTranscode(t *testing.T) {
	testutil.TestSuccessfulTranscodeEqFunc(t, "tstzrange", []interface{}{
		&pgtype.Tstzrange{LowerType: pgtype.Empty, UpperType: pgtype.Empty, Status: pgtype.Present},
		&pgtype.Tstzrange{
			Lower:     pgtype.Timestamptz{Time: time.Date(1990, 12, 31, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
			Upper:     pgtype.Timestamptz{Time: time.Date(2028, 1, 1, 0, 23, 12, 0, time.UTC), Status: pgtype.Present},
			LowerType: pgtype.Inclusive,
			UpperType: pgtype.Exclusive,
			Status:    pgtype.Present,
		},
		&pgtype.Tstzrange{
			Lower:     pgtype.Timestamptz{Time: time.Date(1800, 12, 31, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
			Upper:     pgtype.Timestamptz{Time: time.Date(2200, 1, 1, 0, 23, 12, 0, time.UTC), Status: pgtype.Present},
			LowerType: pgtype.Inclusive,
			UpperType: pgtype.Exclusive,
			Status:    pgtype.Present,
		},
		&pgtype.Tstzrange{Status: pgtype.Null},
	}, func(aa, bb interface{}) bool {
		a := aa.(pgtype.Tstzrange)
		b := bb.(pgtype.Tstzrange)

		return a.Status == b.Status &&
			a.Lower.Time.Equal(b.Lower.Time) &&
			a.Lower.Status == b.Lower.Status &&
			a.Lower.InfinityModifier == b.Lower.InfinityModifier &&
			a.Upper.Time.Equal(b.Upper.Time) &&
			a.Upper.Status == b.Upper.Status &&
			a.Upper.InfinityModifier == b.Upper.InfinityModifier
	})
}
