package pgx_test

import (
	"strconv"
	"testing"

	"github.com/jackc/pgx"
)

func TestQueryArgs(t *testing.T) {
	var qa pgx.QueryArgs

	for i := 1; i < 512; i++ {
		expectedPlaceholder := "$" + strconv.Itoa(i)
		placeholder := qa.Append(i)
		if placeholder != expectedPlaceholder {
			t.Errorf(`Expected qa.Append to return "%s", but it returned "%s"`, expectedPlaceholder, placeholder)
		}
	}
}

func BenchmarkQueryArgs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		qa := pgx.QueryArgs(make([]interface{}, 0, 16))
		qa.Append("foo1")
		qa.Append("foo2")
		qa.Append("foo3")
		qa.Append("foo4")
		qa.Append("foo5")
		qa.Append("foo6")
		qa.Append("foo7")
		qa.Append("foo8")
		qa.Append("foo9")
		qa.Append("foo10")
	}
}
