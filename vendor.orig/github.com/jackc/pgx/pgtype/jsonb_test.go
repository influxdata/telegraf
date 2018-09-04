package pgtype_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestJSONBTranscode(t *testing.T) {
	conn := testutil.MustConnectPgx(t)
	defer testutil.MustClose(t, conn)
	if _, ok := conn.ConnInfo.DataTypeForName("jsonb"); !ok {
		t.Skip("Skipping due to no jsonb type")
	}

	testutil.TestSuccessfulTranscode(t, "jsonb", []interface{}{
		&pgtype.JSONB{Bytes: []byte("{}"), Status: pgtype.Present},
		&pgtype.JSONB{Bytes: []byte("null"), Status: pgtype.Present},
		&pgtype.JSONB{Bytes: []byte("42"), Status: pgtype.Present},
		&pgtype.JSONB{Bytes: []byte(`"hello"`), Status: pgtype.Present},
		&pgtype.JSONB{Status: pgtype.Null},
	})
}

func TestJSONBSet(t *testing.T) {
	successfulTests := []struct {
		source interface{}
		result pgtype.JSONB
	}{
		{source: "{}", result: pgtype.JSONB{Bytes: []byte("{}"), Status: pgtype.Present}},
		{source: []byte("{}"), result: pgtype.JSONB{Bytes: []byte("{}"), Status: pgtype.Present}},
		{source: ([]byte)(nil), result: pgtype.JSONB{Status: pgtype.Null}},
		{source: (*string)(nil), result: pgtype.JSONB{Status: pgtype.Null}},
		{source: []int{1, 2, 3}, result: pgtype.JSONB{Bytes: []byte("[1,2,3]"), Status: pgtype.Present}},
		{source: map[string]interface{}{"foo": "bar"}, result: pgtype.JSONB{Bytes: []byte(`{"foo":"bar"}`), Status: pgtype.Present}},
	}

	for i, tt := range successfulTests {
		var d pgtype.JSONB
		err := d.Set(tt.source)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if !reflect.DeepEqual(d, tt.result) {
			t.Errorf("%d: expected %v to convert to %v, but it was %v", i, tt.source, tt.result, d)
		}
	}
}

func TestJSONBAssignTo(t *testing.T) {
	var s string
	var ps *string
	var b []byte

	rawStringTests := []struct {
		src      pgtype.JSONB
		dst      *string
		expected string
	}{
		{src: pgtype.JSONB{Bytes: []byte("{}"), Status: pgtype.Present}, dst: &s, expected: "{}"},
	}

	for i, tt := range rawStringTests {
		err := tt.src.AssignTo(tt.dst)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if *tt.dst != tt.expected {
			t.Errorf("%d: expected %v to assign %v, but result was %v", i, tt.src, tt.expected, *tt.dst)
		}
	}

	rawBytesTests := []struct {
		src      pgtype.JSONB
		dst      *[]byte
		expected []byte
	}{
		{src: pgtype.JSONB{Bytes: []byte("{}"), Status: pgtype.Present}, dst: &b, expected: []byte("{}")},
		{src: pgtype.JSONB{Status: pgtype.Null}, dst: &b, expected: (([]byte)(nil))},
	}

	for i, tt := range rawBytesTests {
		err := tt.src.AssignTo(tt.dst)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if bytes.Compare(tt.expected, *tt.dst) != 0 {
			t.Errorf("%d: expected %v to assign %v, but result was %v", i, tt.src, tt.expected, *tt.dst)
		}
	}

	var mapDst map[string]interface{}
	type structDst struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var strDst structDst

	unmarshalTests := []struct {
		src      pgtype.JSONB
		dst      interface{}
		expected interface{}
	}{
		{src: pgtype.JSONB{Bytes: []byte(`{"foo":"bar"}`), Status: pgtype.Present}, dst: &mapDst, expected: map[string]interface{}{"foo": "bar"}},
		{src: pgtype.JSONB{Bytes: []byte(`{"name":"John","age":42}`), Status: pgtype.Present}, dst: &strDst, expected: structDst{Name: "John", Age: 42}},
	}
	for i, tt := range unmarshalTests {
		err := tt.src.AssignTo(tt.dst)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if dst := reflect.ValueOf(tt.dst).Elem().Interface(); !reflect.DeepEqual(dst, tt.expected) {
			t.Errorf("%d: expected %v to assign %v, but result was %v", i, tt.src, tt.expected, dst)
		}
	}

	pointerAllocTests := []struct {
		src      pgtype.JSONB
		dst      **string
		expected *string
	}{
		{src: pgtype.JSONB{Status: pgtype.Null}, dst: &ps, expected: ((*string)(nil))},
	}

	for i, tt := range pointerAllocTests {
		err := tt.src.AssignTo(tt.dst)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if *tt.dst == tt.expected {
			t.Errorf("%d: expected %v to assign %v, but result was %v", i, tt.src, tt.expected, *tt.dst)
		}
	}
}
