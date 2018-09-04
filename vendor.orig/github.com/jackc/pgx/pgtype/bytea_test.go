package pgtype_test

import (
	"reflect"
	"testing"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestByteaTranscode(t *testing.T) {
	testutil.TestSuccessfulTranscode(t, "bytea", []interface{}{
		&pgtype.Bytea{Bytes: []byte{1, 2, 3}, Status: pgtype.Present},
		&pgtype.Bytea{Bytes: []byte{}, Status: pgtype.Present},
		&pgtype.Bytea{Bytes: nil, Status: pgtype.Null},
	})
}

func TestByteaSet(t *testing.T) {
	successfulTests := []struct {
		source interface{}
		result pgtype.Bytea
	}{
		{source: []byte{1, 2, 3}, result: pgtype.Bytea{Bytes: []byte{1, 2, 3}, Status: pgtype.Present}},
		{source: []byte{}, result: pgtype.Bytea{Bytes: []byte{}, Status: pgtype.Present}},
		{source: []byte(nil), result: pgtype.Bytea{Status: pgtype.Null}},
		{source: _byteSlice{1, 2, 3}, result: pgtype.Bytea{Bytes: []byte{1, 2, 3}, Status: pgtype.Present}},
		{source: _byteSlice(nil), result: pgtype.Bytea{Status: pgtype.Null}},
	}

	for i, tt := range successfulTests {
		var r pgtype.Bytea
		err := r.Set(tt.source)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if !reflect.DeepEqual(r, tt.result) {
			t.Errorf("%d: expected %v to convert to %v, but it was %v", i, tt.source, tt.result, r)
		}
	}
}

func TestByteaAssignTo(t *testing.T) {
	var buf []byte
	var _buf _byteSlice
	var pbuf *[]byte
	var _pbuf *_byteSlice

	simpleTests := []struct {
		src      pgtype.Bytea
		dst      interface{}
		expected interface{}
	}{
		{src: pgtype.Bytea{Bytes: []byte{1, 2, 3}, Status: pgtype.Present}, dst: &buf, expected: []byte{1, 2, 3}},
		{src: pgtype.Bytea{Bytes: []byte{1, 2, 3}, Status: pgtype.Present}, dst: &_buf, expected: _byteSlice{1, 2, 3}},
		{src: pgtype.Bytea{Bytes: []byte{1, 2, 3}, Status: pgtype.Present}, dst: &pbuf, expected: &[]byte{1, 2, 3}},
		{src: pgtype.Bytea{Bytes: []byte{1, 2, 3}, Status: pgtype.Present}, dst: &_pbuf, expected: &_byteSlice{1, 2, 3}},
		{src: pgtype.Bytea{Status: pgtype.Null}, dst: &pbuf, expected: ((*[]byte)(nil))},
		{src: pgtype.Bytea{Status: pgtype.Null}, dst: &_pbuf, expected: ((*_byteSlice)(nil))},
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
}
