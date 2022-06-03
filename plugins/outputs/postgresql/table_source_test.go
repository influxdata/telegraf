package postgresql

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/coocood/freecache"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs/postgresql/utils"
)

func TestTableSource(t *testing.T) {
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

func TestTableSource_tagJSONB(t *testing.T) {
	p := newPostgresqlTest(t)
	p.TagsAsJsonb = true

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"a": "one", "b": "two"}, MSI{"v": 1}),
	}

	tsrc := NewTableSources(p.Postgresql, metrics)[t.Name()]
	row := nextSrcRow(tsrc)
	require.NoError(t, tsrc.Err())

	assert.IsType(t, time.Time{}, row["time"])
	var tags MSI
	require.NoError(t, json.Unmarshal(row["tags"].([]byte), &tags))
	assert.EqualValues(t, MSI{"a": "one", "b": "two"}, tags)
	assert.EqualValues(t, 1, row["v"])
}

func TestTableSource_tagTable(t *testing.T) {
	p := newPostgresqlTest(t)
	p.TagsAsForeignKeys = true
	p.tagsCache = freecache.NewCache(5 * 1024 * 1024)

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"a": "one", "b": "two"}, MSI{"v": 1}),
	}

	tsrc := NewTableSources(p.Postgresql, metrics)[t.Name()]
	ttsrc := NewTagTableSource(tsrc)
	ttrow := nextSrcRow(ttsrc)
	assert.EqualValues(t, "one", ttrow["a"])
	assert.EqualValues(t, "two", ttrow["b"])

	row := nextSrcRow(tsrc)
	assert.Equal(t, row["tag_id"], ttrow["tag_id"])
}

func TestTableSource_tagTableJSONB(t *testing.T) {
	p := newPostgresqlTest(t)
	p.TagsAsForeignKeys = true
	p.TagsAsJsonb = true
	p.tagsCache = freecache.NewCache(5 * 1024 * 1024)

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"a": "one", "b": "two"}, MSI{"v": 1}),
	}

	tsrc := NewTableSources(p.Postgresql, metrics)[t.Name()]
	ttsrc := NewTagTableSource(tsrc)
	ttrow := nextSrcRow(ttsrc)
	var tags MSI
	require.NoError(t, json.Unmarshal(ttrow["tags"].([]byte), &tags))
	assert.EqualValues(t, MSI{"a": "one", "b": "two"}, tags)
}

func TestTableSource_fieldsJSONB(t *testing.T) {
	p := newPostgresqlTest(t)
	p.FieldsAsJsonb = true

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"tag": "foo"}, MSI{"a": 1, "b": 2}),
	}

	tsrc := NewTableSources(p.Postgresql, metrics)[t.Name()]
	row := nextSrcRow(tsrc)
	var fields MSI
	require.NoError(t, json.Unmarshal(row["fields"].([]byte), &fields))
	// json unmarshals numbers as floats
	assert.EqualValues(t, MSI{"a": 1.0, "b": 2.0}, fields)
}

// TagsAsForeignKeys=false
// Test that when a tag column is dropped, all metrics containing that tag are dropped.
func TestTableSource_DropColumn_tag(t *testing.T) {
	p := newPostgresqlTest(t)

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"a": "one", "b": "two"}, MSI{"v": 1}),
		newMetric(t, "", MSS{"a": "one"}, MSI{"v": 2}),
	}
	tsrc := NewTableSources(p.Postgresql, metrics)[t.Name()]

	// Drop column "b"
	var col utils.Column
	for _, c := range tsrc.TagColumns() {
		if c.Name == "b" {
			col = c
			break
		}
	}
	_ = tsrc.DropColumn(col)

	row := nextSrcRow(tsrc)
	assert.EqualValues(t, "one", row["a"])
	assert.EqualValues(t, 2, row["v"])
	assert.False(t, tsrc.Next())
}

// TagsAsForeignKeys=true, ForeignTagConstraint=true
// Test that when a tag column is dropped, all metrics containing that tag are dropped.
func TestTableSource_DropColumn_tag_fkTrue_fcTrue(t *testing.T) {
	p := newPostgresqlTest(t)
	p.TagsAsForeignKeys = true
	p.ForeignTagConstraint = true
	p.tagsCache = freecache.NewCache(5 * 1024 * 1024)

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"a": "one", "b": "two"}, MSI{"v": 1}),
		newMetric(t, "", MSS{"a": "one"}, MSI{"v": 2}),
	}
	tsrc := NewTableSources(p.Postgresql, metrics)[t.Name()]

	// Drop column "b"
	var col utils.Column
	for _, c := range tsrc.TagColumns() {
		if c.Name == "b" {
			col = c
			break
		}
	}
	_ = tsrc.DropColumn(col)

	ttsrc := NewTagTableSource(tsrc)
	row := nextSrcRow(ttsrc)
	assert.EqualValues(t, "one", row["a"])
	assert.False(t, ttsrc.Next())

	row = nextSrcRow(tsrc)
	assert.EqualValues(t, 2, row["v"])
	assert.False(t, tsrc.Next())
}

// TagsAsForeignKeys=true, ForeignTagConstraint=false
// Test that when a tag column is dropped, metrics are still added while the tag is not.
func TestTableSource_DropColumn_tag_fkTrue_fcFalse(t *testing.T) {
	p := newPostgresqlTest(t)
	p.TagsAsForeignKeys = true
	p.ForeignTagConstraint = false
	p.tagsCache = freecache.NewCache(5 * 1024 * 1024)

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"a": "one", "b": "two"}, MSI{"v": 1}),
		newMetric(t, "", MSS{"a": "one"}, MSI{"v": 2}),
	}
	tsrc := NewTableSources(p.Postgresql, metrics)[t.Name()]

	// Drop column "b"
	var col utils.Column
	for _, c := range tsrc.TagColumns() {
		if c.Name == "b" {
			col = c
			break
		}
	}
	_ = tsrc.DropColumn(col)

	ttsrc := NewTagTableSource(tsrc)
	row := nextSrcRow(ttsrc)
	assert.EqualValues(t, "one", row["a"])
	assert.False(t, ttsrc.Next())

	row = nextSrcRow(tsrc)
	assert.EqualValues(t, 1, row["v"])
	row = nextSrcRow(tsrc)
	assert.EqualValues(t, 2, row["v"])
}

// Test that when a field is dropped, only the field is dropped, and all rows remain, unless it was the only field.
func TestTableSource_DropColumn_field(t *testing.T) {
	p := newPostgresqlTest(t)

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"tag": "foo"}, MSI{"a": 1}),
		newMetric(t, "", MSS{"tag": "foo"}, MSI{"a": 2, "b": 3}),
	}
	tsrc := NewTableSources(p.Postgresql, metrics)[t.Name()]

	// Drop column "a"
	var col utils.Column
	for _, c := range tsrc.FieldColumns() {
		if c.Name == "a" {
			col = c
			break
		}
	}
	_ = tsrc.DropColumn(col)

	row := nextSrcRow(tsrc)
	assert.EqualValues(t, "foo", row["tag"])
	assert.EqualValues(t, 3, row["b"])
	assert.False(t, tsrc.Next())
}

func TestTableSource_InconsistentTags(t *testing.T) {
	p := newPostgresqlTest(t)

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"a": "1"}, MSI{"b": 2}),
		newMetric(t, "", MSS{"c": "3"}, MSI{"d": 4}),
	}
	tsrc := NewTableSources(p.Postgresql, metrics)[t.Name()]

	trow := nextSrcRow(tsrc)
	assert.EqualValues(t, "1", trow["a"])
	assert.EqualValues(t, nil, trow["c"])

	trow = nextSrcRow(tsrc)
	assert.EqualValues(t, nil, trow["a"])
	assert.EqualValues(t, "3", trow["c"])
}

func TestTagTableSource_InconsistentTags(t *testing.T) {
	p := newPostgresqlTest(t)
	p.TagsAsForeignKeys = true
	p.tagsCache = freecache.NewCache(5 * 1024 * 1024)

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"a": "1"}, MSI{"b": 2}),
		newMetric(t, "", MSS{"c": "3"}, MSI{"d": 4}),
	}
	tsrc := NewTableSources(p.Postgresql, metrics)[t.Name()]
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

	assert.ElementsMatch(t, expected, actual)
}
