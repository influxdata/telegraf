package pgtype_test

import (
	"reflect"
	"testing"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestFloat4ArrayTranscode(t *testing.T) {
	testutil.TestSuccessfulTranscode(t, "float4[]", []interface{}{
		&pgtype.Float4Array{
			Elements:   nil,
			Dimensions: nil,
			Status:     pgtype.Present,
		},
		&pgtype.Float4Array{
			Elements: []pgtype.Float4{
				{Float: 1, Status: pgtype.Present},
				{Status: pgtype.Null},
			},
			Dimensions: []pgtype.ArrayDimension{{Length: 2, LowerBound: 1}},
			Status:     pgtype.Present,
		},
		&pgtype.Float4Array{Status: pgtype.Null},
		&pgtype.Float4Array{
			Elements: []pgtype.Float4{
				{Float: 1, Status: pgtype.Present},
				{Float: 2, Status: pgtype.Present},
				{Float: 3, Status: pgtype.Present},
				{Float: 4, Status: pgtype.Present},
				{Status: pgtype.Null},
				{Float: 6, Status: pgtype.Present},
			},
			Dimensions: []pgtype.ArrayDimension{{Length: 3, LowerBound: 1}, {Length: 2, LowerBound: 1}},
			Status:     pgtype.Present,
		},
		&pgtype.Float4Array{
			Elements: []pgtype.Float4{
				{Float: 1, Status: pgtype.Present},
				{Float: 2, Status: pgtype.Present},
				{Float: 3, Status: pgtype.Present},
				{Float: 4, Status: pgtype.Present},
			},
			Dimensions: []pgtype.ArrayDimension{
				{Length: 2, LowerBound: 4},
				{Length: 2, LowerBound: 2},
			},
			Status: pgtype.Present,
		},
	})
}

func TestFloat4ArraySet(t *testing.T) {
	successfulTests := []struct {
		source interface{}
		result pgtype.Float4Array
	}{
		{
			source: []float32{1},
			result: pgtype.Float4Array{
				Elements:   []pgtype.Float4{{Float: 1, Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present},
		},
		{
			source: (([]float32)(nil)),
			result: pgtype.Float4Array{Status: pgtype.Null},
		},
	}

	for i, tt := range successfulTests {
		var r pgtype.Float4Array
		err := r.Set(tt.source)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if !reflect.DeepEqual(r, tt.result) {
			t.Errorf("%d: expected %v to convert to %v, but it was %v", i, tt.source, tt.result, r)
		}
	}
}

func TestFloat4ArrayAssignTo(t *testing.T) {
	var float32Slice []float32
	var namedFloat32Slice _float32Slice

	simpleTests := []struct {
		src      pgtype.Float4Array
		dst      interface{}
		expected interface{}
	}{
		{
			src: pgtype.Float4Array{
				Elements:   []pgtype.Float4{{Float: 1.23, Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst:      &float32Slice,
			expected: []float32{1.23},
		},
		{
			src: pgtype.Float4Array{
				Elements:   []pgtype.Float4{{Float: 1.23, Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst:      &namedFloat32Slice,
			expected: _float32Slice{1.23},
		},
		{
			src:      pgtype.Float4Array{Status: pgtype.Null},
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
		src pgtype.Float4Array
		dst interface{}
	}{
		{
			src: pgtype.Float4Array{
				Elements:   []pgtype.Float4{{Status: pgtype.Null}},
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
