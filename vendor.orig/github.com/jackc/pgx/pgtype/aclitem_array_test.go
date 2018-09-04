package pgtype_test

import (
	"reflect"
	"testing"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestACLItemArrayTranscode(t *testing.T) {
	testutil.TestSuccessfulTranscode(t, "aclitem[]", []interface{}{
		&pgtype.ACLItemArray{
			Elements:   nil,
			Dimensions: nil,
			Status:     pgtype.Present,
		},
		&pgtype.ACLItemArray{
			Elements: []pgtype.ACLItem{
				{String: "=r/postgres", Status: pgtype.Present},
				{Status: pgtype.Null},
			},
			Dimensions: []pgtype.ArrayDimension{{Length: 2, LowerBound: 1}},
			Status:     pgtype.Present,
		},
		&pgtype.ACLItemArray{Status: pgtype.Null},
		&pgtype.ACLItemArray{
			Elements: []pgtype.ACLItem{
				{String: "=r/postgres", Status: pgtype.Present},
				{String: "postgres=arwdDxt/postgres", Status: pgtype.Present},
				{String: `postgres=arwdDxt/" tricky, ' } "" \ test user "`, Status: pgtype.Present},
				{String: "=r/postgres", Status: pgtype.Present},
				{Status: pgtype.Null},
				{String: "=r/postgres", Status: pgtype.Present},
			},
			Dimensions: []pgtype.ArrayDimension{{Length: 3, LowerBound: 1}, {Length: 2, LowerBound: 1}},
			Status:     pgtype.Present,
		},
		&pgtype.ACLItemArray{
			Elements: []pgtype.ACLItem{
				{String: "=r/postgres", Status: pgtype.Present},
				{String: "postgres=arwdDxt/postgres", Status: pgtype.Present},
				{String: "=r/postgres", Status: pgtype.Present},
				{String: "postgres=arwdDxt/postgres", Status: pgtype.Present},
			},
			Dimensions: []pgtype.ArrayDimension{
				{Length: 2, LowerBound: 4},
				{Length: 2, LowerBound: 2},
			},
			Status: pgtype.Present,
		},
	})
}

func TestACLItemArraySet(t *testing.T) {
	successfulTests := []struct {
		source interface{}
		result pgtype.ACLItemArray
	}{
		{
			source: []string{"=r/postgres"},
			result: pgtype.ACLItemArray{
				Elements:   []pgtype.ACLItem{{String: "=r/postgres", Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present},
		},
		{
			source: (([]string)(nil)),
			result: pgtype.ACLItemArray{Status: pgtype.Null},
		},
	}

	for i, tt := range successfulTests {
		var r pgtype.ACLItemArray
		err := r.Set(tt.source)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if !reflect.DeepEqual(r, tt.result) {
			t.Errorf("%d: expected %v to convert to %v, but it was %v", i, tt.source, tt.result, r)
		}
	}
}

func TestACLItemArrayAssignTo(t *testing.T) {
	var stringSlice []string
	type _stringSlice []string
	var namedStringSlice _stringSlice

	simpleTests := []struct {
		src      pgtype.ACLItemArray
		dst      interface{}
		expected interface{}
	}{
		{
			src: pgtype.ACLItemArray{
				Elements:   []pgtype.ACLItem{{String: "=r/postgres", Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst:      &stringSlice,
			expected: []string{"=r/postgres"},
		},
		{
			src: pgtype.ACLItemArray{
				Elements:   []pgtype.ACLItem{{String: "=r/postgres", Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst:      &namedStringSlice,
			expected: _stringSlice{"=r/postgres"},
		},
		{
			src:      pgtype.ACLItemArray{Status: pgtype.Null},
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
		src pgtype.ACLItemArray
		dst interface{}
	}{
		{
			src: pgtype.ACLItemArray{
				Elements:   []pgtype.ACLItem{{Status: pgtype.Null}},
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
