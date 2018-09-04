package pgtype_test

import (
	"reflect"
	"testing"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestInt4ArrayTranscode(t *testing.T) {
	testutil.TestSuccessfulTranscode(t, "int4[]", []interface{}{
		&pgtype.Int4Array{
			Elements:   nil,
			Dimensions: nil,
			Status:     pgtype.Present,
		},
		&pgtype.Int4Array{
			Elements: []pgtype.Int4{
				{Int: 1, Status: pgtype.Present},
				{Status: pgtype.Null},
			},
			Dimensions: []pgtype.ArrayDimension{{Length: 2, LowerBound: 1}},
			Status:     pgtype.Present,
		},
		&pgtype.Int4Array{Status: pgtype.Null},
		&pgtype.Int4Array{
			Elements: []pgtype.Int4{
				{Int: 1, Status: pgtype.Present},
				{Int: 2, Status: pgtype.Present},
				{Int: 3, Status: pgtype.Present},
				{Int: 4, Status: pgtype.Present},
				{Status: pgtype.Null},
				{Int: 6, Status: pgtype.Present},
			},
			Dimensions: []pgtype.ArrayDimension{{Length: 3, LowerBound: 1}, {Length: 2, LowerBound: 1}},
			Status:     pgtype.Present,
		},
		&pgtype.Int4Array{
			Elements: []pgtype.Int4{
				{Int: 1, Status: pgtype.Present},
				{Int: 2, Status: pgtype.Present},
				{Int: 3, Status: pgtype.Present},
				{Int: 4, Status: pgtype.Present},
			},
			Dimensions: []pgtype.ArrayDimension{
				{Length: 2, LowerBound: 4},
				{Length: 2, LowerBound: 2},
			},
			Status: pgtype.Present,
		},
	})
}

func TestInt4ArraySet(t *testing.T) {
	successfulTests := []struct {
		source interface{}
		result pgtype.Int4Array
	}{
		{
			source: []int32{1},
			result: pgtype.Int4Array{
				Elements:   []pgtype.Int4{{Int: 1, Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present},
		},
		{
			source: []uint32{1},
			result: pgtype.Int4Array{
				Elements:   []pgtype.Int4{{Int: 1, Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present},
		},
		{
			source: (([]int32)(nil)),
			result: pgtype.Int4Array{Status: pgtype.Null},
		},
	}

	for i, tt := range successfulTests {
		var r pgtype.Int4Array
		err := r.Set(tt.source)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if !reflect.DeepEqual(r, tt.result) {
			t.Errorf("%d: expected %v to convert to %v, but it was %v", i, tt.source, tt.result, r)
		}
	}
}

func TestInt4ArrayAssignTo(t *testing.T) {
	var int32Slice []int32
	var uint32Slice []uint32
	var namedInt32Slice _int32Slice

	simpleTests := []struct {
		src      pgtype.Int4Array
		dst      interface{}
		expected interface{}
	}{
		{
			src: pgtype.Int4Array{
				Elements:   []pgtype.Int4{{Int: 1, Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst:      &int32Slice,
			expected: []int32{1},
		},
		{
			src: pgtype.Int4Array{
				Elements:   []pgtype.Int4{{Int: 1, Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst:      &uint32Slice,
			expected: []uint32{1},
		},
		{
			src: pgtype.Int4Array{
				Elements:   []pgtype.Int4{{Int: 1, Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst:      &namedInt32Slice,
			expected: _int32Slice{1},
		},
		{
			src:      pgtype.Int4Array{Status: pgtype.Null},
			dst:      &int32Slice,
			expected: (([]int32)(nil)),
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
		src pgtype.Int4Array
		dst interface{}
	}{
		{
			src: pgtype.Int4Array{
				Elements:   []pgtype.Int4{{Status: pgtype.Null}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst: &int32Slice,
		},
		{
			src: pgtype.Int4Array{
				Elements:   []pgtype.Int4{{Int: -1, Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst: &uint32Slice,
		},
	}

	for i, tt := range errorTests {
		err := tt.src.AssignTo(tt.dst)
		if err == nil {
			t.Errorf("%d: expected error but none was returned (%v -> %v)", i, tt.src, tt.dst)
		}
	}

}
