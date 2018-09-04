package pgtype_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestRecordTranscode(t *testing.T) {
	conn := testutil.MustConnectPgx(t)
	defer testutil.MustClose(t, conn)

	tests := []struct {
		sql      string
		expected pgtype.Record
	}{
		{
			sql: `select row()`,
			expected: pgtype.Record{
				Fields: []pgtype.Value{},
				Status: pgtype.Present,
			},
		},
		{
			sql: `select row('foo'::text, 42::int4)`,
			expected: pgtype.Record{
				Fields: []pgtype.Value{
					&pgtype.Text{String: "foo", Status: pgtype.Present},
					&pgtype.Int4{Int: 42, Status: pgtype.Present},
				},
				Status: pgtype.Present,
			},
		},
		{
			sql: `select row(100.0::float4, 1.09::float4)`,
			expected: pgtype.Record{
				Fields: []pgtype.Value{
					&pgtype.Float4{Float: 100, Status: pgtype.Present},
					&pgtype.Float4{Float: 1.09, Status: pgtype.Present},
				},
				Status: pgtype.Present,
			},
		},
		{
			sql: `select row('foo'::text, array[1, 2, null, 4]::int4[], 42::int4)`,
			expected: pgtype.Record{
				Fields: []pgtype.Value{
					&pgtype.Text{String: "foo", Status: pgtype.Present},
					&pgtype.Int4Array{
						Elements: []pgtype.Int4{
							{Int: 1, Status: pgtype.Present},
							{Int: 2, Status: pgtype.Present},
							{Status: pgtype.Null},
							{Int: 4, Status: pgtype.Present},
						},
						Dimensions: []pgtype.ArrayDimension{{Length: 4, LowerBound: 1}},
						Status:     pgtype.Present,
					},
					&pgtype.Int4{Int: 42, Status: pgtype.Present},
				},
				Status: pgtype.Present,
			},
		},
		{
			sql: `select row(null)`,
			expected: pgtype.Record{
				Fields: []pgtype.Value{
					&pgtype.Unknown{Status: pgtype.Null},
				},
				Status: pgtype.Present,
			},
		},
		{
			sql: `select null::record`,
			expected: pgtype.Record{
				Status: pgtype.Null,
			},
		},
	}

	for i, tt := range tests {
		psName := fmt.Sprintf("test%d", i)
		ps, err := conn.Prepare(psName, tt.sql)
		if err != nil {
			t.Fatal(err)
		}
		ps.FieldDescriptions[0].FormatCode = pgx.BinaryFormatCode

		var result pgtype.Record
		if err := conn.QueryRow(psName).Scan(&result); err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}

		if !reflect.DeepEqual(tt.expected, result) {
			t.Errorf("%d: expected %#v, got %#v", i, tt.expected, result)
		}
	}
}

func TestRecordWithUnknownOID(t *testing.T) {
	conn := testutil.MustConnectPgx(t)
	defer testutil.MustClose(t, conn)

	_, err := conn.Exec(`drop type if exists floatrange;

create type floatrange as range (
  subtype = float8,
  subtype_diff = float8mi
);`)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Exec("drop type floatrange")

	var result pgtype.Record
	err = conn.QueryRow("select row('foo'::text, floatrange(1, 10), 'bar'::text)").Scan(&result)
	if err == nil {
		t.Errorf("expected error but none")
	}
}

func TestRecordAssignTo(t *testing.T) {
	var valueSlice []pgtype.Value
	var interfaceSlice []interface{}

	simpleTests := []struct {
		src      pgtype.Record
		dst      interface{}
		expected interface{}
	}{
		{
			src: pgtype.Record{
				Fields: []pgtype.Value{
					&pgtype.Text{String: "foo", Status: pgtype.Present},
					&pgtype.Int4{Int: 42, Status: pgtype.Present},
				},
				Status: pgtype.Present,
			},
			dst: &valueSlice,
			expected: []pgtype.Value{
				&pgtype.Text{String: "foo", Status: pgtype.Present},
				&pgtype.Int4{Int: 42, Status: pgtype.Present},
			},
		},
		{
			src: pgtype.Record{
				Fields: []pgtype.Value{
					&pgtype.Text{String: "foo", Status: pgtype.Present},
					&pgtype.Int4{Int: 42, Status: pgtype.Present},
				},
				Status: pgtype.Present,
			},
			dst:      &interfaceSlice,
			expected: []interface{}{"foo", int32(42)},
		},
		{
			src:      pgtype.Record{Status: pgtype.Null},
			dst:      &valueSlice,
			expected: (([]pgtype.Value)(nil)),
		},
		{
			src:      pgtype.Record{Status: pgtype.Null},
			dst:      &interfaceSlice,
			expected: (([]interface{})(nil)),
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
