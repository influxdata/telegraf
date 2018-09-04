package pgtype_test

import (
	"testing"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestVarbitTranscode(t *testing.T) {
	testutil.TestSuccessfulTranscode(t, "varbit", []interface{}{
		&pgtype.Varbit{Bytes: []byte{}, Len: 0, Status: pgtype.Present},
		&pgtype.Varbit{Bytes: []byte{0, 1, 128, 254, 255}, Len: 40, Status: pgtype.Present},
		&pgtype.Varbit{Bytes: []byte{0, 1, 128, 254, 128}, Len: 33, Status: pgtype.Present},
		&pgtype.Varbit{Status: pgtype.Null},
	})
}

func TestVarbitNormalize(t *testing.T) {
	testutil.TestSuccessfulNormalize(t, []testutil.NormalizeTest{
		{
			SQL:   "select B'111111111'",
			Value: &pgtype.Varbit{Bytes: []byte{255, 128}, Len: 9, Status: pgtype.Present},
		},
	})
}
