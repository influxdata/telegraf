package pgtype_test

import (
	"testing"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestBPCharArrayTranscode(t *testing.T) {
	testutil.TestSuccessfulTranscode(t, "char(8)[]", []interface{}{
		&pgtype.BPCharArray{
			Elements:   nil,
			Dimensions: nil,
			Status:     pgtype.Present,
		},
		&pgtype.BPCharArray{
			Elements: []pgtype.BPChar{
				pgtype.BPChar{String: "foo     ", Status: pgtype.Present},
				pgtype.BPChar{Status: pgtype.Null},
			},
			Dimensions: []pgtype.ArrayDimension{{Length: 2, LowerBound: 1}},
			Status:     pgtype.Present,
		},
		&pgtype.BPCharArray{Status: pgtype.Null},
		&pgtype.BPCharArray{
			Elements: []pgtype.BPChar{
				pgtype.BPChar{String: "bar     ", Status: pgtype.Present},
				pgtype.BPChar{String: "NuLL    ", Status: pgtype.Present},
				pgtype.BPChar{String: `wow"quz\`, Status: pgtype.Present},
				pgtype.BPChar{String: "1       ", Status: pgtype.Present},
				pgtype.BPChar{String: "1       ", Status: pgtype.Present},
				pgtype.BPChar{String: "null    ", Status: pgtype.Present},
			},
			Dimensions: []pgtype.ArrayDimension{
				{Length: 3, LowerBound: 1},
				{Length: 2, LowerBound: 1},
			},
			Status: pgtype.Present,
		},
		&pgtype.BPCharArray{
			Elements: []pgtype.BPChar{
				pgtype.BPChar{String: " bar    ", Status: pgtype.Present},
				pgtype.BPChar{String: "    baz ", Status: pgtype.Present},
				pgtype.BPChar{String: "    quz ", Status: pgtype.Present},
				pgtype.BPChar{String: "foo     ", Status: pgtype.Present},
			},
			Dimensions: []pgtype.ArrayDimension{
				{Length: 2, LowerBound: 4},
				{Length: 2, LowerBound: 2},
			},
			Status: pgtype.Present,
		},
	})
}
