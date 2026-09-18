package sqltemplate

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/outputs/postgresql/utils"
)

func TestRenderWithoutTable(t *testing.T) {
	// NewTable returns nil when the table name is empty, and Render used to
	// dereference it while building "allColumns", taking the agent down with
	// a segfault instead of reporting the problem.
	tmpl := &Template{}
	require.NoError(t, tmpl.UnmarshalText([]byte(`CREATE TABLE {{ .table }} ({{ .columns }})`)))

	columns := []utils.Column{{Name: "time", Type: "timestamptz", Role: utils.TimeColType}}
	metricTable := NewTable("public", "metrics", columns)

	_, err := tmpl.Render(NewTable("", "", nil), columns, metricTable, nil)
	require.ErrorContains(t, err, "cannot render template without a table")

	sql, err := tmpl.Render(metricTable, columns, metricTable, nil)
	require.NoError(t, err)
	require.Equal(t, `CREATE TABLE "public"."metrics" ("time" timestamptz)`, string(sql))
}
