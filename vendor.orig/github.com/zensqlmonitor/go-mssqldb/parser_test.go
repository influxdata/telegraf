package mssql

import (
	"testing"
)

func TestParseParams(t *testing.T) {
	values := []struct {
		s string
		d string
		n int
	}{
		{"select ?", "select @p1", 1},
		{"select ?, ?", "select @p1, @p2", 2},
		{"select ? -- ?", "select @p1 -- ?", 1},
		{"select ? -- ?\n, ?", "select @p1 -- ?\n, @p2", 2},
		{"select ? - ?", "select @p1 - @p2", 2},
		{"select ? /* ? */, ?", "select @p1 /* ? */, @p2", 2},
		{"select ? /* ? * ? */, ?", "select @p1 /* ? * ? */, @p2", 2},
		{"select \"foo?\", [foo?], 'foo?', ?", "select \"foo?\", [foo?], 'foo?', @p1", 1},
		{"select \"x\"\"y\", [x]]y], 'x''y', ?", "select \"x\"\"y\", [x]]y], 'x''y', @p1", 1},
		{"select \"foo?\", ?", "select \"foo?\", @p1", 1},
		{"select 'foo?', ?", "select 'foo?', @p1", 1},
		{"select [foo?], ?", "select [foo?], @p1", 1},
		{"select $1", "select @p1", 1},
		{"select $1, $2", "select @p1, @p2", 2},
		{"select $1, $1", "select @p1, @p1", 1},
		{"select :1", "select @p1", 1},
		{"select :1, :2", "select @p1, @p2", 2},
		{"select :1, :1", "select @p1, @p1", 1},
		{"select ?1", "select @p1", 1},
		{"select ?1, ?2", "select @p1, @p2", 2},
		{"select ?1, ?1", "select @p1, @p1", 1},
		{"select $12", "select @p12", 12},
		{"select ? /* ? /* ? */ ? */ ?", "select @p1 /* ? /* ? */ ? */ @p2", 2},
		{"select ? /* ? / ? */ ?", "select @p1 /* ? / ? */ @p2", 2},
		{"select $", "select $", 0},
		{"select x::y", "select x:@y", 1},
		{"select '", "select '", 0},
		{"select \"", "select \"", 0},
		{"select [", "select [", 0},
		{"select []", "select []", 0},
		{"select -", "select -", 0},
		{"select /", "select /", 0},
		{"select 1/1", "select 1/1", 0},
		{"select /*", "select /*", 0},
		{"select /**", "select /**", 0},
		{"select /*/", "select /*/", 0},
	}

	for _, v := range values {
		d, n := parseParams(v.s)
		if d != v.d {
			t.Errorf("Parse params don't match for %s, got %s but expected %s", v.s, d, v.d)
		}
		if n != v.n {
			t.Errorf("Parse number of params don't match for %s, got %d but expected %d", v.s, n, v.n)
		}
	}
}
