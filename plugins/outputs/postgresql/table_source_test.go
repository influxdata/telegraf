package postgresql

import (
	"encoding/json"
	"slices"
	"testing"
	"time"

	"github.com/coocood/freecache"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/outputs/postgresql/utils"
)

func TestTableSource(_ *testing.T) {
}

type source interface {
	pgx.CopyFromSource
	ColumnNames() []string
}

func nextSrcRow(src source) MSI {
	if !src.Next() {
		return nil
	}
	row := MSI{}
	vals, err := src.Values()
	if err != nil {
		panic(err)
	}
	for i, name := range src.ColumnNames() {
		row[name] = vals[i]
	}
	return row
}

func TestTableSourceIntegration_tagJSONB(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p, err := newPostgresqlTest(t)
	require.NoError(t, err)
	p.TagsAsJsonb = true

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"a": "one", "b": "two"}, MSI{"v": 1}),
	}

	tsrc := NewTableSources(p.Postgresql, metrics)[tableKey{schema: p.Schema, name: t.Name()}]
	row := nextSrcRow(tsrc)
	require.NoError(t, tsrc.Err())

	require.IsType(t, time.Time{}, row["time"])
	var tags MSI
	require.NoError(t, json.Unmarshal(row["tags"].([]byte), &tags))
	require.EqualValues(t, MSI{"a": "one", "b": "two"}, tags)
	require.EqualValues(t, 1, row["v"])
}

func TestTableSourceIntegration_tagTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p, err := newPostgresqlTest(t)
	require.NoError(t, err)
	p.TagsAsForeignKeys = true
	p.tagsCache = freecache.NewCache(5 * 1024 * 1024)

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"a": "one", "b": "two"}, MSI{"v": 1}),
	}

	tsrc := NewTableSources(p.Postgresql, metrics)[tableKey{schema: p.Schema, name: t.Name()}]
	ttsrc := NewTagTableSource(tsrc)
	ttrow := nextSrcRow(ttsrc)
	require.EqualValues(t, "one", ttrow["a"])
	require.EqualValues(t, "two", ttrow["b"])

	row := nextSrcRow(tsrc)
	require.Equal(t, row["tag_id"], ttrow["tag_id"])
}

func TestTableSourceIntegration_tagTableJSONB(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p, err := newPostgresqlTest(t)
	require.NoError(t, err)
	p.TagsAsForeignKeys = true
	p.TagsAsJsonb = true
	p.tagsCache = freecache.NewCache(5 * 1024 * 1024)

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"a": "one", "b": "two"}, MSI{"v": 1}),
	}

	tsrc := NewTableSources(p.Postgresql, metrics)[tableKey{schema: p.Schema, name: t.Name()}]
	ttsrc := NewTagTableSource(tsrc)
	ttrow := nextSrcRow(ttsrc)
	var tags MSI
	require.NoError(t, json.Unmarshal(ttrow["tags"].([]byte), &tags))
	require.EqualValues(t, MSI{"a": "one", "b": "two"}, tags)
}

func TestTableSourceIntegration_fieldsJSONB(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p, err := newPostgresqlTest(t)
	require.NoError(t, err)
	p.FieldsAsJsonb = true

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"tag": "foo"}, MSI{"a": 1, "b": 2}),
	}

	tsrc := NewTableSources(p.Postgresql, metrics)[tableKey{schema: p.Schema, name: t.Name()}]
	row := nextSrcRow(tsrc)
	var fields MSI
	require.NoError(t, json.Unmarshal(row["fields"].([]byte), &fields))
	// json unmarshals numbers as floats
	require.EqualValues(t, MSI{"a": 1.0, "b": 2.0}, fields)
}

// TagsAsForeignKeys=false
// Test that when a tag column is dropped, all metrics containing that tag are dropped.
func TestTableSourceIntegration_DropColumn_tag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p, err := newPostgresqlTest(t)
	require.NoError(t, err)

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"a": "one", "b": "two"}, MSI{"v": 1}),
		newMetric(t, "", MSS{"a": "one"}, MSI{"v": 2}),
	}
	tsrc := NewTableSources(p.Postgresql, metrics)[tableKey{schema: p.Schema, name: t.Name()}]

	// Drop column "b"
	var col utils.Column
	for _, c := range tsrc.TagColumns() {
		if c.Name == "b" {
			col = c
			break
		}
	}
	require.NoError(t, tsrc.DropColumn(col))

	row := nextSrcRow(tsrc)
	require.EqualValues(t, "one", row["a"])
	require.EqualValues(t, 2, row["v"])
	require.False(t, tsrc.Next())
}

// TagsAsForeignKeys=true, ForeignTagConstraint=true
// Test that when a tag column is dropped, all metrics containing that tag are dropped.
func TestTableSourceIntegration_DropColumn_tag_fkTrue_fcTrue(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p, err := newPostgresqlTest(t)
	require.NoError(t, err)
	p.TagsAsForeignKeys = true
	p.ForeignTagConstraint = true
	p.tagsCache = freecache.NewCache(5 * 1024 * 1024)

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"a": "one", "b": "two"}, MSI{"v": 1}),
		newMetric(t, "", MSS{"a": "one"}, MSI{"v": 2}),
	}
	tsrc := NewTableSources(p.Postgresql, metrics)[tableKey{schema: p.Schema, name: t.Name()}]

	// Drop column "b"
	var col utils.Column
	for _, c := range tsrc.TagColumns() {
		if c.Name == "b" {
			col = c
			break
		}
	}
	require.NoError(t, tsrc.DropColumn(col))

	ttsrc := NewTagTableSource(tsrc)
	row := nextSrcRow(ttsrc)
	require.EqualValues(t, "one", row["a"])
	require.False(t, ttsrc.Next())

	row = nextSrcRow(tsrc)
	require.EqualValues(t, 2, row["v"])
	require.False(t, tsrc.Next())
}

// TagsAsForeignKeys=true, ForeignTagConstraint=false
// Test that when a tag column is dropped, metrics are still added while the tag is not.
func TestTableSourceIntegration_DropColumn_tag_fkTrue_fcFalse(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p, err := newPostgresqlTest(t)
	require.NoError(t, err)
	p.TagsAsForeignKeys = true
	p.ForeignTagConstraint = false
	p.tagsCache = freecache.NewCache(5 * 1024 * 1024)

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"a": "one", "b": "two"}, MSI{"v": 1}),
		newMetric(t, "", MSS{"a": "one"}, MSI{"v": 2}),
	}
	tsrc := NewTableSources(p.Postgresql, metrics)[tableKey{schema: p.Schema, name: t.Name()}]

	// Drop column "b"
	var col utils.Column
	for _, c := range tsrc.TagColumns() {
		if c.Name == "b" {
			col = c
			break
		}
	}
	require.NoError(t, tsrc.DropColumn(col))

	ttsrc := NewTagTableSource(tsrc)
	row := nextSrcRow(ttsrc)
	require.EqualValues(t, "one", row["a"])
	require.False(t, ttsrc.Next())

	row = nextSrcRow(tsrc)
	require.EqualValues(t, 1, row["v"])
	row = nextSrcRow(tsrc)
	require.EqualValues(t, 2, row["v"])
}

// Test that when a field is dropped, only the field is dropped, and all rows remain, unless it was the only field.
func TestTableSourceIntegration_DropColumn_field(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p, err := newPostgresqlTest(t)
	require.NoError(t, err)

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"tag": "foo"}, MSI{"a": 1}),
		newMetric(t, "", MSS{"tag": "foo"}, MSI{"a": 2, "b": 3}),
	}
	tsrc := NewTableSources(p.Postgresql, metrics)[tableKey{schema: p.Schema, name: t.Name()}]

	// Drop column "a"
	var col utils.Column
	for _, c := range tsrc.FieldColumns() {
		if c.Name == "a" {
			col = c
			break
		}
	}
	require.NoError(t, tsrc.DropColumn(col))

	row := nextSrcRow(tsrc)
	require.EqualValues(t, "foo", row["tag"])
	require.EqualValues(t, 3, row["b"])
	require.False(t, tsrc.Next())
}

func TestTableSourceIntegration_InconsistentTags(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p, err := newPostgresqlTest(t)
	require.NoError(t, err)

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"a": "1"}, MSI{"b": 2}),
		newMetric(t, "", MSS{"c": "3"}, MSI{"d": 4}),
	}
	tsrc := NewTableSources(p.Postgresql, metrics)[tableKey{schema: p.Schema, name: t.Name()}]

	trow := nextSrcRow(tsrc)
	require.EqualValues(t, "1", trow["a"])
	require.Nil(t, trow["c"])

	trow = nextSrcRow(tsrc)
	require.Nil(t, trow["a"])
	require.EqualValues(t, "3", trow["c"])
}

func TestTagTableSourceIntegration_InconsistentTags(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p, err := newPostgresqlTest(t)
	require.NoError(t, err)
	p.TagsAsForeignKeys = true
	p.tagsCache = freecache.NewCache(5 * 1024 * 1024)

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"a": "1"}, MSI{"b": 2}),
		newMetric(t, "", MSS{"c": "3"}, MSI{"d": 4}),
	}
	tsrc := NewTableSources(p.Postgresql, metrics)[tableKey{schema: p.Schema, name: t.Name()}]
	ttsrc := NewTagTableSource(tsrc)

	// ttsrc is in non-deterministic order
	expected := []MSI{
		{"a": "1", "c": nil},
		{"a": nil, "c": "3"},
	}

	var actual []MSI
	for row := nextSrcRow(ttsrc); row != nil; row = nextSrcRow(ttsrc) {
		delete(row, "tag_id")
		actual = append(actual, row)
	}

	require.ElementsMatch(t, expected, actual)
}

func tagColumnNames(tsrc *TableSource) []string {
	cols := tsrc.TagColumns()
	names := make([]string, 0, len(cols))
	for _, col := range cols {
		names = append(names, col.Name)
	}
	return names
}

func TestNewTableSourcesSchemaTag(t *testing.T) {
	p := newPostgresql()
	p.SchemaTag = "schema"

	now := time.Unix(0, 0)
	metrics := []telegraf.Metric{
		metric.New("cpu", MSS{"schema": "a", "host": "h1"}, MSI{"v": 1}, now),
		metric.New("cpu", MSS{"schema": "b", "host": "h1"}, MSI{"v": 2}, now),
		metric.New("cpu", MSS{"host": "h1"}, MSI{"v": 3}, now),
		metric.New("mem", MSS{"schema": "a", "host": "h1"}, MSI{"v": 4}, now),
		metric.New("cpu", MSS{"schema": "", "host": "h1"}, MSI{"v": 5}, now),
		metric.New("cpu", MSS{"schema": "a", "host": "h2"}, MSI{"v": 6}, now),
	}

	expected := map[tableKey]int{
		{schema: "a", name: "cpu"}:      2,
		{schema: "b", name: "cpu"}:      1,
		{schema: "public", name: "cpu"}: 2,
		{schema: "a", name: "mem"}:      1,
	}

	check := func(t *testing.T, metrics []telegraf.Metric) {
		tsrcs := NewTableSources(p, metrics)
		require.Len(t, tsrcs, len(expected))
		for key, count := range expected {
			tsrc, ok := tsrcs[key]
			require.Truef(t, ok, "missing table source for %v", key)
			require.Equal(t, key.schema, tsrc.Schema())
			require.Equal(t, key.name, tsrc.Name())
			require.Len(t, tsrc.metrics, count)
			require.Contains(t, tagColumnNames(tsrc), "schema")
		}
		// The configured schema must never be modified
		require.Equal(t, "public", p.Schema)
	}

	t.Run("in order", func(t *testing.T) {
		check(t, metrics)
	})

	t.Run("reversed", func(t *testing.T) {
		reversed := slices.Clone(metrics)
		slices.Reverse(reversed)
		check(t, reversed)
	})
}

func TestNewTableSourcesExcludeSchemaTag(t *testing.T) {
	p := newPostgresql()
	p.SchemaTag = "schema"
	p.ExcludeSchemaTag = true

	m := metric.New("cpu", MSS{"schema": "a", "host": "h1"}, MSI{"v": 1}, time.Unix(0, 0))
	tsrcs := NewTableSources(p, []telegraf.Metric{m})

	tsrc := tsrcs[tableKey{schema: "a", name: "cpu"}]
	require.NotNil(t, tsrc)
	require.Equal(t, []string{"host"}, tagColumnNames(tsrc))
	require.False(t, tsrc.metrics[0].HasTag("schema"))

	// The input metric must not be modified
	require.True(t, m.HasTag("schema"))
}

func TestNewTableSourcesExcludeSchemaTagTracking(t *testing.T) {
	p := newPostgresql()
	p.SchemaTag = "schema"
	p.ExcludeSchemaTag = true

	var delivered int
	m := metric.New("cpu", MSS{"schema": "a"}, MSI{"v": 1}, time.Unix(0, 0))
	tm, _ := metric.WithTracking(m, func(telegraf.DeliveryInfo) { delivered++ })

	tsrcs := NewTableSources(p, []telegraf.Metric{tm})
	require.Len(t, tsrcs, 1)
	require.Zero(t, delivered, "delivery must not happen before the original metric is accepted")

	// Accepting the original metric must complete the delivery, so the copy
	// must not hold a tracking reference.
	tm.Accept()
	require.Equal(t, 1, delivered)
}

func TestNewTableSourcesNoSchemaTag(t *testing.T) {
	p := newPostgresql()

	now := time.Unix(0, 0)
	metrics := []telegraf.Metric{
		metric.New("cpu", MSS{"schema": "a"}, MSI{"v": 1}, now),
		metric.New("cpu", MSS{"schema": "b"}, MSI{"v": 2}, now),
	}
	tsrcs := NewTableSources(p, metrics)
	require.Len(t, tsrcs, 1)

	tsrc := tsrcs[tableKey{schema: "public", name: "cpu"}]
	require.NotNil(t, tsrc)
	require.Equal(t, "public", tsrc.Schema())
	require.Len(t, tsrc.metrics, 2)
	require.Contains(t, tagColumnNames(tsrc), "schema")
}

func TestTableSourceTagHashSaltPerSchema(t *testing.T) {
	p := newPostgresql()
	a := NewTableSource(p, "a", "cpu")
	b := NewTableSource(p, "b", "cpu")
	require.NotEqual(t, a.tagHashSalt, b.tagHashSalt)
	// The separator prevents schema/table boundaries from colliding
	require.NotEqual(t, NewTableSource(p, "ab", "c").tagHashSalt, NewTableSource(p, "a", "bc").tagHashSalt)
}
