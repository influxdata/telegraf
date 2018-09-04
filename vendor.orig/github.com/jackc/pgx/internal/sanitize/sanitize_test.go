package sanitize_test

import (
	"testing"

	"github.com/jackc/pgx/internal/sanitize"
)

func TestNewQuery(t *testing.T) {
	successTests := []struct {
		sql      string
		expected sanitize.Query
	}{
		{
			sql:      "select 42",
			expected: sanitize.Query{Parts: []sanitize.Part{"select 42"}},
		},
		{
			sql:      "select $1",
			expected: sanitize.Query{Parts: []sanitize.Part{"select ", 1}},
		},
		{
			sql:      "select 'quoted $42', $1",
			expected: sanitize.Query{Parts: []sanitize.Part{"select 'quoted $42', ", 1}},
		},
		{
			sql:      `select "doubled quoted $42", $1`,
			expected: sanitize.Query{Parts: []sanitize.Part{`select "doubled quoted $42", `, 1}},
		},
		{
			sql:      "select 'foo''bar', $1",
			expected: sanitize.Query{Parts: []sanitize.Part{"select 'foo''bar', ", 1}},
		},
		{
			sql:      `select "foo""bar", $1`,
			expected: sanitize.Query{Parts: []sanitize.Part{`select "foo""bar", `, 1}},
		},
		{
			sql:      "select '''', $1",
			expected: sanitize.Query{Parts: []sanitize.Part{"select '''', ", 1}},
		},
		{
			sql:      `select """", $1`,
			expected: sanitize.Query{Parts: []sanitize.Part{`select """", `, 1}},
		},
		{
			sql:      "select $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11",
			expected: sanitize.Query{Parts: []sanitize.Part{"select ", 1, ", ", 2, ", ", 3, ", ", 4, ", ", 5, ", ", 6, ", ", 7, ", ", 8, ", ", 9, ", ", 10, ", ", 11}},
		},
		{
			sql:      `select "adsf""$1""adsf", $1, 'foo''$$12bar', $2, '$3'`,
			expected: sanitize.Query{Parts: []sanitize.Part{`select "adsf""$1""adsf", `, 1, `, 'foo''$$12bar', `, 2, `, '$3'`}},
		},
		{
			sql:      `select E'escape string\' $42', $1`,
			expected: sanitize.Query{Parts: []sanitize.Part{`select E'escape string\' $42', `, 1}},
		},
		{
			sql:      `select e'escape string\' $42', $1`,
			expected: sanitize.Query{Parts: []sanitize.Part{`select e'escape string\' $42', `, 1}},
		},
	}

	for i, tt := range successTests {
		query, err := sanitize.NewQuery(tt.sql)
		if err != nil {
			t.Errorf("%d. %v", i, err)
		}

		if len(query.Parts) == len(tt.expected.Parts) {
			for j := range query.Parts {
				if query.Parts[j] != tt.expected.Parts[j] {
					t.Errorf("%d. expected part %d to be %v but it was %v", i, j, tt.expected.Parts[j], query.Parts[j])
				}
			}
		} else {
			t.Errorf("%d. expected query parts to be %v but it was %v", i, tt.expected.Parts, query.Parts)
		}
	}
}

func TestQuerySanitize(t *testing.T) {
	successfulTests := []struct {
		query    sanitize.Query
		args     []interface{}
		expected string
	}{
		{
			query:    sanitize.Query{Parts: []sanitize.Part{"select 42"}},
			args:     []interface{}{},
			expected: `select 42`,
		},
		{
			query:    sanitize.Query{Parts: []sanitize.Part{"select ", 1}},
			args:     []interface{}{int64(42)},
			expected: `select 42`,
		},
		{
			query:    sanitize.Query{Parts: []sanitize.Part{"select ", 1}},
			args:     []interface{}{float64(1.23)},
			expected: `select 1.23`,
		},
		{
			query:    sanitize.Query{Parts: []sanitize.Part{"select ", 1}},
			args:     []interface{}{true},
			expected: `select true`,
		},
		{
			query:    sanitize.Query{Parts: []sanitize.Part{"select ", 1}},
			args:     []interface{}{[]byte{0, 1, 2, 3, 255}},
			expected: `select '\x00010203ff'`,
		},
		{
			query:    sanitize.Query{Parts: []sanitize.Part{"select ", 1}},
			args:     []interface{}{nil},
			expected: `select null`,
		},
		{
			query:    sanitize.Query{Parts: []sanitize.Part{"select ", 1}},
			args:     []interface{}{"foobar"},
			expected: `select 'foobar'`,
		},
		{
			query:    sanitize.Query{Parts: []sanitize.Part{"select ", 1}},
			args:     []interface{}{"foo'bar"},
			expected: `select 'foo''bar'`,
		},
		{
			query:    sanitize.Query{Parts: []sanitize.Part{"select ", 1}},
			args:     []interface{}{`foo\'bar`},
			expected: `select 'foo\''bar'`,
		},
	}

	for i, tt := range successfulTests {
		actual, err := tt.query.Sanitize(tt.args...)
		if err != nil {
			t.Errorf("%d. %v", i, err)
			continue
		}

		if tt.expected != actual {
			t.Errorf("%d. expected %s, but got %s", i, tt.expected, actual)
		}
	}

	errorTests := []struct {
		query    sanitize.Query
		args     []interface{}
		expected string
	}{
		{
			query:    sanitize.Query{Parts: []sanitize.Part{"select ", 1, ", ", 2}},
			args:     []interface{}{int64(42)},
			expected: `insufficient arguments`,
		},
		{
			query:    sanitize.Query{Parts: []sanitize.Part{"select 'foo'"}},
			args:     []interface{}{int64(42)},
			expected: `unused argument: 0`,
		},
		{
			query:    sanitize.Query{Parts: []sanitize.Part{"select ", 1}},
			args:     []interface{}{42},
			expected: `invalid arg type: int`,
		},
	}

	for i, tt := range errorTests {
		_, err := tt.query.Sanitize(tt.args...)
		if err == nil || err.Error() != tt.expected {
			t.Errorf("%d. expected error %v, got %v", i, tt.expected, err)
		}
	}
}
