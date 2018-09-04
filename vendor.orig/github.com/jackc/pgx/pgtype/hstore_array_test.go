package pgtype_test

import (
	"reflect"
	"testing"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestHstoreArrayTranscode(t *testing.T) {
	conn := testutil.MustConnectPgx(t)
	defer testutil.MustClose(t, conn)

	text := func(s string) pgtype.Text {
		return pgtype.Text{String: s, Status: pgtype.Present}
	}

	values := []pgtype.Hstore{
		{Map: map[string]pgtype.Text{}, Status: pgtype.Present},
		{Map: map[string]pgtype.Text{"foo": text("bar")}, Status: pgtype.Present},
		{Map: map[string]pgtype.Text{"foo": text("bar"), "baz": text("quz")}, Status: pgtype.Present},
		{Map: map[string]pgtype.Text{"NULL": text("bar")}, Status: pgtype.Present},
		{Map: map[string]pgtype.Text{"foo": text("NULL")}, Status: pgtype.Present},
		{Status: pgtype.Null},
	}

	specialStrings := []string{
		`"`,
		`'`,
		`\`,
		`\\`,
		`=>`,
		` `,
		`\ / / \\ => " ' " '`,
	}
	for _, s := range specialStrings {
		// Special key values
		values = append(values, pgtype.Hstore{Map: map[string]pgtype.Text{s + "foo": text("bar")}, Status: pgtype.Present})         // at beginning
		values = append(values, pgtype.Hstore{Map: map[string]pgtype.Text{"foo" + s + "bar": text("bar")}, Status: pgtype.Present}) // in middle
		values = append(values, pgtype.Hstore{Map: map[string]pgtype.Text{"foo" + s: text("bar")}, Status: pgtype.Present})         // at end
		values = append(values, pgtype.Hstore{Map: map[string]pgtype.Text{s: text("bar")}, Status: pgtype.Present})                 // is key

		// Special value values
		values = append(values, pgtype.Hstore{Map: map[string]pgtype.Text{"foo": text(s + "bar")}, Status: pgtype.Present})         // at beginning
		values = append(values, pgtype.Hstore{Map: map[string]pgtype.Text{"foo": text("foo" + s + "bar")}, Status: pgtype.Present}) // in middle
		values = append(values, pgtype.Hstore{Map: map[string]pgtype.Text{"foo": text("foo" + s)}, Status: pgtype.Present})         // at end
		values = append(values, pgtype.Hstore{Map: map[string]pgtype.Text{"foo": text(s)}, Status: pgtype.Present})                 // is key
	}

	src := &pgtype.HstoreArray{
		Elements:   values,
		Dimensions: []pgtype.ArrayDimension{{Length: int32(len(values)), LowerBound: 1}},
		Status:     pgtype.Present,
	}

	ps, err := conn.Prepare("test", "select $1::hstore[]")
	if err != nil {
		t.Fatal(err)
	}

	formats := []struct {
		name       string
		formatCode int16
	}{
		{name: "TextFormat", formatCode: pgx.TextFormatCode},
		{name: "BinaryFormat", formatCode: pgx.BinaryFormatCode},
	}

	for _, fc := range formats {
		ps.FieldDescriptions[0].FormatCode = fc.formatCode
		vEncoder := testutil.ForceEncoder(src, fc.formatCode)
		if vEncoder == nil {
			t.Logf("%#v does not implement %v", src, fc.name)
			continue
		}

		var result pgtype.HstoreArray
		err := conn.QueryRow("test", vEncoder).Scan(&result)
		if err != nil {
			t.Errorf("%v: %v", fc.name, err)
			continue
		}

		if result.Status != src.Status {
			t.Errorf("%v: expected Status %v, got %v", fc.formatCode, src.Status, result.Status)
			continue
		}

		if len(result.Elements) != len(src.Elements) {
			t.Errorf("%v: expected %v elements, got %v", fc.formatCode, len(src.Elements), len(result.Elements))
			continue
		}

		for i := range result.Elements {
			a := src.Elements[i]
			b := result.Elements[i]

			if a.Status != b.Status {
				t.Errorf("%v element idx %d: expected status %v, got %v", fc.formatCode, i, a.Status, b.Status)
			}

			if len(a.Map) != len(b.Map) {
				t.Errorf("%v element idx %d: expected %v pairs, got %v", fc.formatCode, i, len(a.Map), len(b.Map))
			}

			for k := range a.Map {
				if a.Map[k] != b.Map[k] {
					t.Errorf("%v element idx %d: expected key %v to be %v, got %v", fc.formatCode, i, k, a.Map[k], b.Map[k])
				}
			}
		}
	}
}

func TestHstoreArraySet(t *testing.T) {
	successfulTests := []struct {
		src    []map[string]string
		result pgtype.HstoreArray
	}{
		{
			src: []map[string]string{{"foo": "bar"}},
			result: pgtype.HstoreArray{
				Elements: []pgtype.Hstore{
					{
						Map:    map[string]pgtype.Text{"foo": {String: "bar", Status: pgtype.Present}},
						Status: pgtype.Present,
					},
				},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
		},
	}

	for i, tt := range successfulTests {
		var dst pgtype.HstoreArray
		err := dst.Set(tt.src)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if !reflect.DeepEqual(dst, tt.result) {
			t.Errorf("%d: expected %v to convert to %v, but it was %v", i, tt.src, tt.result, dst)
		}
	}
}

func TestHstoreArrayAssignTo(t *testing.T) {
	var m []map[string]string

	simpleTests := []struct {
		src      pgtype.HstoreArray
		dst      *[]map[string]string
		expected []map[string]string
	}{
		{
			src: pgtype.HstoreArray{
				Elements: []pgtype.Hstore{
					{
						Map:    map[string]pgtype.Text{"foo": {String: "bar", Status: pgtype.Present}},
						Status: pgtype.Present,
					},
				},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst:      &m,
			expected: []map[string]string{{"foo": "bar"}}},
		{src: pgtype.HstoreArray{Status: pgtype.Null}, dst: &m, expected: (([]map[string]string)(nil))},
	}

	for i, tt := range simpleTests {
		err := tt.src.AssignTo(tt.dst)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if !reflect.DeepEqual(*tt.dst, tt.expected) {
			t.Errorf("%d: expected %v to assign %v, but result was %v", i, tt.src, tt.expected, *tt.dst)
		}
	}
}
