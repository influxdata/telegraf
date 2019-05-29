package postgresql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostgresqlQuote(t *testing.T) {
	assert.Equal(t, `"foo"`, quoteIdent("foo"))
	assert.Equal(t, `"fo'o"`, quoteIdent("fo'o"))
	assert.Equal(t, `"fo""o"`, quoteIdent("fo\"o"))

	assert.Equal(t, "'foo'", quoteLiteral("foo"))
	assert.Equal(t, "'fo''o'", quoteLiteral("fo'o"))
	assert.Equal(t, "'fo\"o'", quoteLiteral("fo\"o"))
}

func TestPostgresqlInsertStatement(t *testing.T) {
	p := newPostgresql()

	p.TagsAsJsonb = false
	p.FieldsAsJsonb = false

	sql := p.generateInsert("m", []string{"time", "f"})
	assert.Equal(t, `INSERT INTO "public"."m"("time","f") VALUES($1,$2)`, sql)

	sql = p.generateInsert("m", []string{"time", "i"})
	assert.Equal(t, `INSERT INTO "public"."m"("time","i") VALUES($1,$2)`, sql)

	sql = p.generateInsert("m", []string{"time", "f", "i"})
	assert.Equal(t, `INSERT INTO "public"."m"("time","f","i") VALUES($1,$2,$3)`, sql)

	sql = p.generateInsert("m", []string{"time", "k", "i"})
	assert.Equal(t, `INSERT INTO "public"."m"("time","k","i") VALUES($1,$2,$3)`, sql)

	sql = p.generateInsert("m", []string{"time", "k1", "k2", "i"})
	assert.Equal(t, `INSERT INTO "public"."m"("time","k1","k2","i") VALUES($1,$2,$3,$4)`, sql)
}
