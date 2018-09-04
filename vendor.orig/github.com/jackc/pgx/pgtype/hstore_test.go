package pgtype_test

import (
	"reflect"
	"testing"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestHstoreTranscode(t *testing.T) {
	text := func(s string) pgtype.Text {
		return pgtype.Text{String: s, Status: pgtype.Present}
	}

	values := []interface{}{
		&pgtype.Hstore{Map: map[string]pgtype.Text{}, Status: pgtype.Present},
		&pgtype.Hstore{Map: map[string]pgtype.Text{"foo": text("bar")}, Status: pgtype.Present},
		&pgtype.Hstore{Map: map[string]pgtype.Text{"foo": text("bar"), "baz": text("quz")}, Status: pgtype.Present},
		&pgtype.Hstore{Map: map[string]pgtype.Text{"NULL": text("bar")}, Status: pgtype.Present},
		&pgtype.Hstore{Map: map[string]pgtype.Text{"foo": text("NULL")}, Status: pgtype.Present},
		&pgtype.Hstore{Status: pgtype.Null},
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
		values = append(values, &pgtype.Hstore{Map: map[string]pgtype.Text{s + "foo": text("bar")}, Status: pgtype.Present})         // at beginning
		values = append(values, &pgtype.Hstore{Map: map[string]pgtype.Text{"foo" + s + "bar": text("bar")}, Status: pgtype.Present}) // in middle
		values = append(values, &pgtype.Hstore{Map: map[string]pgtype.Text{"foo" + s: text("bar")}, Status: pgtype.Present})         // at end
		values = append(values, &pgtype.Hstore{Map: map[string]pgtype.Text{s: text("bar")}, Status: pgtype.Present})                 // is key

		// Special value values
		values = append(values, &pgtype.Hstore{Map: map[string]pgtype.Text{"foo": text(s + "bar")}, Status: pgtype.Present})         // at beginning
		values = append(values, &pgtype.Hstore{Map: map[string]pgtype.Text{"foo": text("foo" + s + "bar")}, Status: pgtype.Present}) // in middle
		values = append(values, &pgtype.Hstore{Map: map[string]pgtype.Text{"foo": text("foo" + s)}, Status: pgtype.Present})         // at end
		values = append(values, &pgtype.Hstore{Map: map[string]pgtype.Text{"foo": text(s)}, Status: pgtype.Present})                 // is key
	}

	testutil.TestSuccessfulTranscodeEqFunc(t, "hstore", values, func(ai, bi interface{}) bool {
		a := ai.(pgtype.Hstore)
		b := bi.(pgtype.Hstore)

		if len(a.Map) != len(b.Map) || a.Status != b.Status {
			return false
		}

		for k := range a.Map {
			if a.Map[k] != b.Map[k] {
				return false
			}
		}

		return true
	})
}

func TestHstoreSet(t *testing.T) {
	successfulTests := []struct {
		src    map[string]string
		result pgtype.Hstore
	}{
		{src: map[string]string{"foo": "bar"}, result: pgtype.Hstore{Map: map[string]pgtype.Text{"foo": {String: "bar", Status: pgtype.Present}}, Status: pgtype.Present}},
	}

	for i, tt := range successfulTests {
		var dst pgtype.Hstore
		err := dst.Set(tt.src)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if !reflect.DeepEqual(dst, tt.result) {
			t.Errorf("%d: expected %v to convert to %v, but it was %v", i, tt.src, tt.result, dst)
		}
	}
}

func TestHstoreAssignTo(t *testing.T) {
	var m map[string]string

	simpleTests := []struct {
		src      pgtype.Hstore
		dst      *map[string]string
		expected map[string]string
	}{
		{src: pgtype.Hstore{Map: map[string]pgtype.Text{"foo": {String: "bar", Status: pgtype.Present}}, Status: pgtype.Present}, dst: &m, expected: map[string]string{"foo": "bar"}},
		{src: pgtype.Hstore{Status: pgtype.Null}, dst: &m, expected: ((map[string]string)(nil))},
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
