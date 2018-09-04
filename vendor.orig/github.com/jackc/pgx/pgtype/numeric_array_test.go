package pgtype_test

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestNumericArrayTranscode(t *testing.T) {
	testutil.TestSuccessfulTranscode(t, "numeric[]", []interface{}{
		&pgtype.NumericArray{
			Elements:   nil,
			Dimensions: nil,
			Status:     pgtype.Present,
		},
		&pgtype.NumericArray{
			Elements: []pgtype.Numeric{
				{Int: big.NewInt(1), Status: pgtype.Present},
				{Status: pgtype.Null},
			},
			Dimensions: []pgtype.ArrayDimension{{Length: 2, LowerBound: 1}},
			Status:     pgtype.Present,
		},
		&pgtype.NumericArray{Status: pgtype.Null},
		&pgtype.NumericArray{
			Elements: []pgtype.Numeric{
				{Int: big.NewInt(1), Status: pgtype.Present},
				{Int: big.NewInt(2), Status: pgtype.Present},
				{Int: big.NewInt(3), Status: pgtype.Present},
				{Int: big.NewInt(4), Status: pgtype.Present},
				{Status: pgtype.Null},
				{Int: big.NewInt(6), Status: pgtype.Present},
			},
			Dimensions: []pgtype.ArrayDimension{{Length: 3, LowerBound: 1}, {Length: 2, LowerBound: 1}},
			Status:     pgtype.Present,
		},
		&pgtype.NumericArray{
			Elements: []pgtype.Numeric{
				{Int: big.NewInt(1), Status: pgtype.Present},
				{Int: big.NewInt(2), Status: pgtype.Present},
				{Int: big.NewInt(3), Status: pgtype.Present},
				{Int: big.NewInt(4), Status: pgtype.Present},
			},
			Dimensions: []pgtype.ArrayDimension{
				{Length: 2, LowerBound: 4},
				{Length: 2, LowerBound: 2},
			},
			Status: pgtype.Present,
		},
	})
}

func TestNumericArraySet(t *testing.T) {
	successfulTests := []struct {
		source interface{}
		result pgtype.NumericArray
	}{
		{
			source: []float32{1},
			result: pgtype.NumericArray{
				Elements:   []pgtype.Numeric{{Int: big.NewInt(1), Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present},
		},
		{
			source: []float64{1},
			result: pgtype.NumericArray{
				Elements:   []pgtype.Numeric{{Int: big.NewInt(1), Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present},
		},
		{
			source: (([]float32)(nil)),
			result: pgtype.NumericArray{Status: pgtype.Null},
		},
	}

	for i, tt := range successfulTests {
		var r pgtype.NumericArray
		err := r.Set(tt.source)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if !reflect.DeepEqual(r, tt.result) {
			t.Errorf("%d: expected %v to convert to %v, but it was %v", i, tt.source, tt.result, r)
		}
	}
}

func TestNumericArrayAssignTo(t *testing.T) {
	var float32Slice []float32
	var float64Slice []float64

	simpleTests := []struct {
		src      pgtype.NumericArray
		dst      interface{}
		expected interface{}
	}{
		{
			src: pgtype.NumericArray{
				Elements:   []pgtype.Numeric{{Int: big.NewInt(1), Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst:      &float32Slice,
			expected: []float32{1},
		},
		{
			src: pgtype.NumericArray{
				Elements:   []pgtype.Numeric{{Int: big.NewInt(1), Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst:      &float64Slice,
			expected: []float64{1},
		},
		{
			src:      pgtype.NumericArray{Status: pgtype.Null},
			dst:      &float32Slice,
			expected: (([]float32)(nil)),
		},
	}

	for i, tt := range simpleTests {
		err := tt.src.AssignTo(tt.dst)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if dst := reflect.ValueOf(tt.dst).Elem().Interface(); !reflect.DeepEqual(dst, tt.expected) {
			t.Errorf("%d: expected %v to assign %v, but result was %v", i, tt.src, tt.expected, dst)
		}
	}

	errorTests := []struct {
		src pgtype.NumericArray
		dst interface{}
	}{
		{
			src: pgtype.NumericArray{
				Elements:   []pgtype.Numeric{{Status: pgtype.Null}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst: &float32Slice,
		},
	}

	for i, tt := range errorTests {
		err := tt.src.AssignTo(tt.dst)
		if err == nil {
			t.Errorf("%d: expected error but none was returned (%v -> %v)", i, tt.src, tt.dst)
		}
	}

}
