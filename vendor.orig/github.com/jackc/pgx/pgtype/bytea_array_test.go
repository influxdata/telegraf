package pgtype_test

import (
	"reflect"
	"testing"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestByteaArrayTranscode(t *testing.T) {
	testutil.TestSuccessfulTranscode(t, "bytea[]", []interface{}{
		&pgtype.ByteaArray{
			Elements:   nil,
			Dimensions: nil,
			Status:     pgtype.Present,
		},
		&pgtype.ByteaArray{
			Elements: []pgtype.Bytea{
				{Bytes: []byte{1, 2, 3}, Status: pgtype.Present},
				{Status: pgtype.Null},
			},
			Dimensions: []pgtype.ArrayDimension{{Length: 2, LowerBound: 1}},
			Status:     pgtype.Present,
		},
		&pgtype.ByteaArray{Status: pgtype.Null},
		&pgtype.ByteaArray{
			Elements: []pgtype.Bytea{
				{Bytes: []byte{1, 2, 3}, Status: pgtype.Present},
				{Bytes: []byte{1, 2, 3}, Status: pgtype.Present},
				{Bytes: []byte{}, Status: pgtype.Present},
				{Bytes: []byte{1, 2, 3}, Status: pgtype.Present},
				{Status: pgtype.Null},
				{Bytes: []byte{1}, Status: pgtype.Present},
			},
			Dimensions: []pgtype.ArrayDimension{{Length: 3, LowerBound: 1}, {Length: 2, LowerBound: 1}},
			Status:     pgtype.Present,
		},
		&pgtype.ByteaArray{
			Elements: []pgtype.Bytea{
				{Bytes: []byte{1, 2, 3}, Status: pgtype.Present},
				{Bytes: []byte{}, Status: pgtype.Present},
				{Bytes: []byte{1, 2, 3}, Status: pgtype.Present},
				{Bytes: []byte{1}, Status: pgtype.Present},
			},
			Dimensions: []pgtype.ArrayDimension{
				{Length: 2, LowerBound: 4},
				{Length: 2, LowerBound: 2},
			},
			Status: pgtype.Present,
		},
	})
}

func TestByteaArraySet(t *testing.T) {
	successfulTests := []struct {
		source interface{}
		result pgtype.ByteaArray
	}{
		{
			source: [][]byte{{1, 2, 3}},
			result: pgtype.ByteaArray{
				Elements:   []pgtype.Bytea{{Bytes: []byte{1, 2, 3}, Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present},
		},
		{
			source: (([][]byte)(nil)),
			result: pgtype.ByteaArray{Status: pgtype.Null},
		},
	}

	for i, tt := range successfulTests {
		var r pgtype.ByteaArray
		err := r.Set(tt.source)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if !reflect.DeepEqual(r, tt.result) {
			t.Errorf("%d: expected %v to convert to %v, but it was %v", i, tt.source, tt.result, r)
		}
	}
}

func TestByteaArrayAssignTo(t *testing.T) {
	var byteByteSlice [][]byte

	simpleTests := []struct {
		src      pgtype.ByteaArray
		dst      interface{}
		expected interface{}
	}{
		{
			src: pgtype.ByteaArray{
				Elements:   []pgtype.Bytea{{Bytes: []byte{1, 2, 3}, Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst:      &byteByteSlice,
			expected: [][]byte{{1, 2, 3}},
		},
		{
			src:      pgtype.ByteaArray{Status: pgtype.Null},
			dst:      &byteByteSlice,
			expected: (([][]byte)(nil)),
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
}
