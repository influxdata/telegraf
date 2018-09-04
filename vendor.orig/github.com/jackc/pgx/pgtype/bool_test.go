package pgtype_test

import (
	"reflect"
	"testing"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestBoolTranscode(t *testing.T) {
	testutil.TestSuccessfulTranscode(t, "bool", []interface{}{
		&pgtype.Bool{Bool: false, Status: pgtype.Present},
		&pgtype.Bool{Bool: true, Status: pgtype.Present},
		&pgtype.Bool{Bool: false, Status: pgtype.Null},
	})
}

func TestBoolSet(t *testing.T) {
	successfulTests := []struct {
		source interface{}
		result pgtype.Bool
	}{
		{source: true, result: pgtype.Bool{Bool: true, Status: pgtype.Present}},
		{source: false, result: pgtype.Bool{Bool: false, Status: pgtype.Present}},
		{source: "true", result: pgtype.Bool{Bool: true, Status: pgtype.Present}},
		{source: "false", result: pgtype.Bool{Bool: false, Status: pgtype.Present}},
		{source: "t", result: pgtype.Bool{Bool: true, Status: pgtype.Present}},
		{source: "f", result: pgtype.Bool{Bool: false, Status: pgtype.Present}},
		{source: _bool(true), result: pgtype.Bool{Bool: true, Status: pgtype.Present}},
		{source: _bool(false), result: pgtype.Bool{Bool: false, Status: pgtype.Present}},
	}

	for i, tt := range successfulTests {
		var r pgtype.Bool
		err := r.Set(tt.source)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if r != tt.result {
			t.Errorf("%d: expected %v to convert to %v, but it was %v", i, tt.source, tt.result, r)
		}
	}
}

func TestBoolAssignTo(t *testing.T) {
	var b bool
	var _b _bool
	var pb *bool
	var _pb *_bool

	simpleTests := []struct {
		src      pgtype.Bool
		dst      interface{}
		expected interface{}
	}{
		{src: pgtype.Bool{Bool: false, Status: pgtype.Present}, dst: &b, expected: false},
		{src: pgtype.Bool{Bool: true, Status: pgtype.Present}, dst: &b, expected: true},
		{src: pgtype.Bool{Bool: false, Status: pgtype.Present}, dst: &_b, expected: _bool(false)},
		{src: pgtype.Bool{Bool: true, Status: pgtype.Present}, dst: &_b, expected: _bool(true)},
		{src: pgtype.Bool{Bool: false, Status: pgtype.Null}, dst: &pb, expected: ((*bool)(nil))},
		{src: pgtype.Bool{Bool: false, Status: pgtype.Null}, dst: &_pb, expected: ((*_bool)(nil))},
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
		src      pgtype.Bool
		dst      interface{}
		expected interface{}
	}{
		{src: pgtype.Bool{Bool: true, Status: pgtype.Present}, dst: &pb, expected: true},
		{src: pgtype.Bool{Bool: true, Status: pgtype.Present}, dst: &_pb, expected: _bool(true)},
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
}
