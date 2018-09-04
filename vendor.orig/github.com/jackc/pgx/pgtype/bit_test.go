package pgtype_test

import (
	"testing"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestBitTranscode(t *testing.T) {
	testutil.TestSuccessfulTranscode(t, "bit(40)", []interface{}{
		&pgtype.Varbit{Bytes: []byte{0, 0, 0, 0, 0}, Len: 40, Status: pgtype.Present},
		&pgtype.Varbit{Bytes: []byte{0, 1, 128, 254, 255}, Len: 40, Status: pgtype.Present},
		&pgtype.Varbit{Status: pgtype.Null},
	})
}

func TestBitNormalize(t *testing.T) {
	testutil.TestSuccessfulNormalize(t, []testutil.NormalizeTest{
		{
			SQL:   "select B'111111111'",
			Value: &pgtype.Bit{Bytes: []byte{255, 128}, Len: 9, Status: pgtype.Present},
		},
	})
}
