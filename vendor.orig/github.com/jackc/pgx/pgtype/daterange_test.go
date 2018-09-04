package pgtype_test

import (
	"testing"
	"time"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestDaterangeTranscode(t *testing.T) {
	testutil.TestSuccessfulTranscodeEqFunc(t, "daterange", []interface{}{
		&pgtype.Daterange{LowerType: pgtype.Empty, UpperType: pgtype.Empty, Status: pgtype.Present},
		&pgtype.Daterange{
			Lower:     pgtype.Date{Time: time.Date(1990, 12, 31, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
			Upper:     pgtype.Date{Time: time.Date(2028, 1, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
			LowerType: pgtype.Inclusive,
			UpperType: pgtype.Exclusive,
			Status:    pgtype.Present,
		},
		&pgtype.Daterange{
			Lower:     pgtype.Date{Time: time.Date(1800, 12, 31, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
			Upper:     pgtype.Date{Time: time.Date(2200, 1, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
			LowerType: pgtype.Inclusive,
			UpperType: pgtype.Exclusive,
			Status:    pgtype.Present,
		},
		&pgtype.Daterange{Status: pgtype.Null},
	}, func(aa, bb interface{}) bool {
		a := aa.(pgtype.Daterange)
		b := bb.(pgtype.Daterange)

		return a.Status == b.Status &&
			a.Lower.Time.Equal(b.Lower.Time) &&
			a.Lower.Status == b.Lower.Status &&
			a.Lower.InfinityModifier == b.Lower.InfinityModifier &&
			a.Upper.Time.Equal(b.Upper.Time) &&
			a.Upper.Status == b.Upper.Status &&
			a.Upper.InfinityModifier == b.Upper.InfinityModifier
	})
}

func TestDaterangeNormalize(t *testing.T) {
	testutil.TestSuccessfulNormalizeEqFunc(t, []testutil.NormalizeTest{
		{
			SQL: "select daterange('2010-01-01', '2010-01-11', '(]')",
			Value: pgtype.Daterange{
				Lower:     pgtype.Date{Time: time.Date(2010, 1, 2, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
				Upper:     pgtype.Date{Time: time.Date(2010, 1, 12, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
				LowerType: pgtype.Inclusive,
				UpperType: pgtype.Exclusive,
				Status:    pgtype.Present,
			},
		},
	}, func(aa, bb interface{}) bool {
		a := aa.(pgtype.Daterange)
		b := bb.(pgtype.Daterange)

		return a.Status == b.Status &&
			a.Lower.Time.Equal(b.Lower.Time) &&
			a.Lower.Status == b.Lower.Status &&
			a.Lower.InfinityModifier == b.Lower.InfinityModifier &&
			a.Upper.Time.Equal(b.Upper.Time) &&
			a.Upper.Status == b.Upper.Status &&
			a.Upper.InfinityModifier == b.Upper.InfinityModifier
	})
}
