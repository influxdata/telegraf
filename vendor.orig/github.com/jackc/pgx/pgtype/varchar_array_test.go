package pgtype_test

import (
	"reflect"
	"testing"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestVarcharArrayTranscode(t *testing.T) {
	testutil.TestSuccessfulTranscode(t, "varchar[]", []interface{}{
		&pgtype.VarcharArray{
			Elements:   nil,
			Dimensions: nil,
			Status:     pgtype.Present,
		},
		&pgtype.VarcharArray{
			Elements: []pgtype.Varchar{
				{String: "foo", Status: pgtype.Present},
				{Status: pgtype.Null},
			},
			Dimensions: []pgtype.ArrayDimension{{Length: 2, LowerBound: 1}},
			Status:     pgtype.Present,
		},
		&pgtype.VarcharArray{Status: pgtype.Null},
		&pgtype.VarcharArray{
			Elements: []pgtype.Varchar{
				{String: "bar ", Status: pgtype.Present},
				{String: "NuLL", Status: pgtype.Present},
				{String: `wow"quz\`, Status: pgtype.Present},
				{String: "", Status: pgtype.Present},
				{Status: pgtype.Null},
				{String: "null", Status: pgtype.Present},
			},
			Dimensions: []pgtype.ArrayDimension{{Length: 3, LowerBound: 1}, {Length: 2, LowerBound: 1}},
			Status:     pgtype.Present,
		},
		&pgtype.VarcharArray{
			Elements: []pgtype.Varchar{
				{String: "bar", Status: pgtype.Present},
				{String: "baz", Status: pgtype.Present},
				{String: "quz", Status: pgtype.Present},
				{String: "foo", Status: pgtype.Present},
			},
			Dimensions: []pgtype.ArrayDimension{
				{Length: 2, LowerBound: 4},
				{Length: 2, LowerBound: 2},
			},
			Status: pgtype.Present,
		},
	})
}

func TestVarcharArraySet(t *testing.T) {
	successfulTests := []struct {
		source interface{}
		result pgtype.VarcharArray
	}{
		{
			source: []string{"foo"},
			result: pgtype.VarcharArray{
				Elements:   []pgtype.Varchar{{String: "foo", Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present},
		},
		{
			source: (([]string)(nil)),
			result: pgtype.VarcharArray{Status: pgtype.Null},
		},
	}

	for i, tt := range successfulTests {
		var r pgtype.VarcharArray
		err := r.Set(tt.source)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if !reflect.DeepEqual(r, tt.result) {
			t.Errorf("%d: expected %v to convert to %v, but it was %v", i, tt.source, tt.result, r)
		}
	}
}

func TestVarcharArrayAssignTo(t *testing.T) {
	var stringSlice []string
	type _stringSlice []string
	var namedStringSlice _stringSlice

	simpleTests := []struct {
		src      pgtype.VarcharArray
		dst      interface{}
		expected interface{}
	}{
		{
			src: pgtype.VarcharArray{
				Elements:   []pgtype.Varchar{{String: "foo", Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst:      &stringSlice,
			expected: []string{"foo"},
		},
		{
			src: pgtype.VarcharArray{
				Elements:   []pgtype.Varchar{{String: "bar", Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst:      &namedStringSlice,
			expected: _stringSlice{"bar"},
		},
		{
			src:      pgtype.VarcharArray{Status: pgtype.Null},
			dst:      &stringSlice,
			expected: (([]string)(nil)),
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
		src pgtype.VarcharArray
		dst interface{}
	}{
		{
			src: pgtype.VarcharArray{
				Elements:   []pgtype.Varchar{{Status: pgtype.Null}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst: &stringSlice,
		},
	}

	for i, tt := range errorTests {
		err := tt.src.AssignTo(tt.dst)
		if err == nil {
			t.Errorf("%d: expected error but none was returned (%v -> %v)", i, tt.src, tt.dst)
		}
	}
}
