package pgtype_test

import (
	"reflect"
	"testing"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestInt8ArrayTranscode(t *testing.T) {
	testutil.TestSuccessfulTranscode(t, "int8[]", []interface{}{
		&pgtype.Int8Array{
			Elements:   nil,
			Dimensions: nil,
			Status:     pgtype.Present,
		},
		&pgtype.Int8Array{
			Elements: []pgtype.Int8{
				{Int: 1, Status: pgtype.Present},
				{Status: pgtype.Null},
			},
			Dimensions: []pgtype.ArrayDimension{{Length: 2, LowerBound: 1}},
			Status:     pgtype.Present,
		},
		&pgtype.Int8Array{Status: pgtype.Null},
		&pgtype.Int8Array{
			Elements: []pgtype.Int8{
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
		&pgtype.Int8Array{
			Elements: []pgtype.Int8{
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

func TestInt8ArraySet(t *testing.T) {
	successfulTests := []struct {
		source interface{}
		result pgtype.Int8Array
	}{
		{
			source: []int64{1},
			result: pgtype.Int8Array{
				Elements:   []pgtype.Int8{{Int: 1, Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present},
		},
		{
			source: []uint64{1},
			result: pgtype.Int8Array{
				Elements:   []pgtype.Int8{{Int: 1, Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present},
		},
		{
			source: (([]int64)(nil)),
			result: pgtype.Int8Array{Status: pgtype.Null},
		},
	}

	for i, tt := range successfulTests {
		var r pgtype.Int8Array
		err := r.Set(tt.source)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if !reflect.DeepEqual(r, tt.result) {
			t.Errorf("%d: expected %v to convert to %v, but it was %v", i, tt.source, tt.result, r)
		}
	}
}

func TestInt8ArrayAssignTo(t *testing.T) {
	var int64Slice []int64
	var uint64Slice []uint64
	var namedInt64Slice _int64Slice

	simpleTests := []struct {
		src      pgtype.Int8Array
		dst      interface{}
		expected interface{}
	}{
		{
			src: pgtype.Int8Array{
				Elements:   []pgtype.Int8{{Int: 1, Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst:      &int64Slice,
			expected: []int64{1},
		},
		{
			src: pgtype.Int8Array{
				Elements:   []pgtype.Int8{{Int: 1, Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst:      &uint64Slice,
			expected: []uint64{1},
		},
		{
			src: pgtype.Int8Array{
				Elements:   []pgtype.Int8{{Int: 1, Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst:      &namedInt64Slice,
			expected: _int64Slice{1},
		},
		{
			src:      pgtype.Int8Array{Status: pgtype.Null},
			dst:      &int64Slice,
			expected: (([]int64)(nil)),
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
		src pgtype.Int8Array
		dst interface{}
	}{
		{
			src: pgtype.Int8Array{
				Elements:   []pgtype.Int8{{Status: pgtype.Null}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst: &int64Slice,
		},
		{
			src: pgtype.Int8Array{
				Elements:   []pgtype.Int8{{Int: -1, Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst: &uint64Slice,
		},
	}

	for i, tt := range errorTests {
		err := tt.src.AssignTo(tt.dst)
		if err == nil {
			t.Errorf("%d: expected error but none was returned (%v -> %v)", i, tt.src, tt.dst)
		}
	}

}
