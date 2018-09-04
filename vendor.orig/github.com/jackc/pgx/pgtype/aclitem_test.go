package pgtype_test

import (
	"reflect"
	"testing"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestACLItemTranscode(t *testing.T) {
	testutil.TestSuccessfulTranscode(t, "aclitem", []interface{}{
		&pgtype.ACLItem{String: "postgres=arwdDxt/postgres", Status: pgtype.Present},
		&pgtype.ACLItem{String: `postgres=arwdDxt/" tricky, ' } "" \ test user "`, Status: pgtype.Present},
		&pgtype.ACLItem{Status: pgtype.Null},
	})
}

func TestACLItemSet(t *testing.T) {
	successfulTests := []struct {
		source interface{}
		result pgtype.ACLItem
	}{
		{source: "postgres=arwdDxt/postgres", result: pgtype.ACLItem{String: "postgres=arwdDxt/postgres", Status: pgtype.Present}},
		{source: (*string)(nil), result: pgtype.ACLItem{Status: pgtype.Null}},
	}

	for i, tt := range successfulTests {
		var d pgtype.ACLItem
		err := d.Set(tt.source)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if d != tt.result {
			t.Errorf("%d: expected %v to convert to %v, but it was %v", i, tt.source, tt.result, d)
		}
	}
}

func TestACLItemAssignTo(t *testing.T) {
	var s string
	var ps *string

	simpleTests := []struct {
		src      pgtype.ACLItem
		dst      interface{}
		expected interface{}
	}{
		{src: pgtype.ACLItem{String: "postgres=arwdDxt/postgres", Status: pgtype.Present}, dst: &s, expected: "postgres=arwdDxt/postgres"},
		{src: pgtype.ACLItem{Status: pgtype.Null}, dst: &ps, expected: ((*string)(nil))},
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
		src      pgtype.ACLItem
		dst      interface{}
		expected interface{}
	}{
		{src: pgtype.ACLItem{String: "postgres=arwdDxt/postgres", Status: pgtype.Present}, dst: &ps, expected: "postgres=arwdDxt/postgres"},
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
		src pgtype.ACLItem
		dst interface{}
	}{
		{src: pgtype.ACLItem{Status: pgtype.Null}, dst: &s},
	}

	for i, tt := range errorTests {
		err := tt.src.AssignTo(tt.dst)
		if err == nil {
			t.Errorf("%d: expected error but none was returned (%v -> %v)", i, tt.src, tt.dst)
		}
	}
}
