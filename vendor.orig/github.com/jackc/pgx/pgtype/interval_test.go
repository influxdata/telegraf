package pgtype_test

import (
	"testing"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestIntervalTranscode(t *testing.T) {
	testutil.TestSuccessfulTranscode(t, "interval", []interface{}{
		&pgtype.Interval{Microseconds: 1, Status: pgtype.Present},
		&pgtype.Interval{Microseconds: 1000000, Status: pgtype.Present},
		&pgtype.Interval{Microseconds: 1000001, Status: pgtype.Present},
		&pgtype.Interval{Microseconds: 123202800000000, Status: pgtype.Present},
		&pgtype.Interval{Days: 1, Status: pgtype.Present},
		&pgtype.Interval{Months: 1, Status: pgtype.Present},
		&pgtype.Interval{Months: 12, Status: pgtype.Present},
		&pgtype.Interval{Months: 13, Days: 15, Microseconds: 1000001, Status: pgtype.Present},
		&pgtype.Interval{Microseconds: -1, Status: pgtype.Present},
		&pgtype.Interval{Microseconds: -1000000, Status: pgtype.Present},
		&pgtype.Interval{Microseconds: -1000001, Status: pgtype.Present},
		&pgtype.Interval{Microseconds: -123202800000000, Status: pgtype.Present},
		&pgtype.Interval{Days: -1, Status: pgtype.Present},
		&pgtype.Interval{Months: -1, Status: pgtype.Present},
		&pgtype.Interval{Months: -12, Status: pgtype.Present},
		&pgtype.Interval{Months: -13, Days: -15, Microseconds: -1000001, Status: pgtype.Present},
		&pgtype.Interval{Status: pgtype.Null},
	})
}

func TestIntervalNormalize(t *testing.T) {
	testutil.TestSuccessfulNormalize(t, []testutil.NormalizeTest{
		{
			SQL:   "select '1 second'::interval",
			Value: &pgtype.Interval{Microseconds: 1000000, Status: pgtype.Present},
		},
		{
			SQL:   "select '1.000001 second'::interval",
			Value: &pgtype.Interval{Microseconds: 1000001, Status: pgtype.Present},
		},
		{
			SQL:   "select '34223 hours'::interval",
			Value: &pgtype.Interval{Microseconds: 123202800000000, Status: pgtype.Present},
		},
		{
			SQL:   "select '1 day'::interval",
			Value: &pgtype.Interval{Days: 1, Status: pgtype.Present},
		},
		{
			SQL:   "select '1 month'::interval",
			Value: &pgtype.Interval{Months: 1, Status: pgtype.Present},
		},
		{
			SQL:   "select '1 year'::interval",
			Value: &pgtype.Interval{Months: 12, Status: pgtype.Present},
		},
		{
			SQL:   "select '-13 mon'::interval",
			Value: &pgtype.Interval{Months: -13, Status: pgtype.Present},
		},
	})
}
