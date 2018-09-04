package pgtype_test

import (
	"reflect"
	"testing"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestNameTranscode(t *testing.T) {
	testutil.TestSuccessfulTranscode(t, "name", []interface{}{
		&pgtype.Name{String: "", Status: pgtype.Present},
		&pgtype.Name{String: "foo", Status: pgtype.Present},
		&pgtype.Name{Status: pgtype.Null},
	})
}

func TestNameSet(t *testing.T) {
	successfulTests := []struct {
		source interface{}
		result pgtype.Name
	}{
		{source: "foo", result: pgtype.Name{String: "foo", Status: pgtype.Present}},
		{source: _string("bar"), result: pgtype.Name{String: "bar", Status: pgtype.Present}},
		{source: (*string)(nil), result: pgtype.Name{Status: pgtype.Null}},
	}

	for i, tt := range successfulTests {
		var d pgtype.Name
		err := d.Set(tt.source)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if d != tt.result {
			t.Errorf("%d: expected %v to convert to %v, but it was %v", i, tt.source, tt.result, d)
		}
	}
}

func TestNameAssignTo(t *testing.T) {
	var s string
	var ps *string

	simpleTests := []struct {
		src      pgtype.Name
		dst      interface{}
		expected interface{}
	}{
		{src: pgtype.Name{String: "foo", Status: pgtype.Present}, dst: &s, expected: "foo"},
		{src: pgtype.Name{Status: pgtype.Null}, dst: &ps, expected: ((*string)(nil))},
	}

	for i, tt := range simpleTests {
		err := tt.src.AssignTo(tt.dst)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if dst := reflect.ValueOf(tt.dst).Elem().Interface(); dst != tt.expected {
			t.Errorf("%d: expected %v to assign %v, but result was %v", i, tt.src, tt.expected, dst)
		}
	}

	pointerAllocTests := []struct {
		src      pgtype.Name
		dst      interface{}
		expected interface{}
	}{
		{src: pgtype.Name{String: "foo", Status: pgtype.Present}, dst: &ps, expected: "foo"},
	}

	for i, tt := range pointerAllocTests {
		err := tt.src.AssignTo(tt.dst)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if dst := reflect.ValueOf(tt.dst).Elem().Elem().Interface(); dst != tt.expected {
			t.Errorf("%d: expected %v to assign %v, but result was %v", i, tt.src, tt.expected, dst)
		}
	}

	errorTests := []struct {
		src pgtype.Name
		dst interface{}
	}{
		{src: pgtype.Name{Status: pgtype.Null}, dst: &s},
	}

	for i, tt := range errorTests {
		err := tt.src.AssignTo(tt.dst)
		if err == nil {
			t.Errorf("%d: expected error but none was returned (%v -> %v)", i, tt.src, tt.dst)
		}
	}
}
