package postgresql

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs/postgresql/template"
	"github.com/influxdata/telegraf/plugins/outputs/postgresql/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTableManager_EnsureStructure(t *testing.T) {
	p := newPostgresqlTest(t)
	require.NoError(t, p.Connect())

	cols := []utils.Column{
		ColumnFromTag("foo", ""),
		ColumnFromField("baz", 0),
	}
	missingCols, err := p.tableManager.EnsureStructure(
		ctx,
		p.db,
		p.tableManager.table(t.Name()),
		cols,
		p.CreateTemplates,
		p.AddColumnTemplates,
		p.tableManager.table(t.Name()),
		nil,
		)
	require.NoError(t, err)
	require.Empty(t, missingCols)

	assert.EqualValues(t, cols[0], p.tableManager.table(t.Name()).Columns()["foo"])
	assert.EqualValues(t, cols[1], p.tableManager.table(t.Name()).Columns()["baz"])
}

func TestTableManager_refreshTableStructure(t *testing.T) {
	p := newPostgresqlTest(t)
	require.NoError(t, p.Connect())

	cols := []utils.Column{
		ColumnFromTag("foo", ""),
		ColumnFromField("baz", 0),
	}
	_, err := p.tableManager.EnsureStructure(
		ctx,
		p.db,
		p.tableManager.table(t.Name()),
		cols,
		p.CreateTemplates,
		p.AddColumnTemplates,
		p.tableManager.table(t.Name()),
		nil,
	)
	require.NoError(t, err)

	p.tableManager.ClearTableCache()
	require.Empty(t, p.tableManager.table(t.Name()).Columns())

	require.NoError(t, p.tableManager.refreshTableStructure(ctx, p.db, p.tableManager.table(t.Name())))

	assert.EqualValues(t, cols[0], p.tableManager.table(t.Name()).Columns()["foo"])
	assert.EqualValues(t, cols[1], p.tableManager.table(t.Name()).Columns()["baz"])
}

func TestTableManager_MatchSource(t *testing.T) {
	p := newPostgresqlTest(t)
	p.TagsAsForeignKeys = true
	require.NoError(t, p.Connect())

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"tag": "foo"}, MSI{"a": 1}),
	}
	tsrc := NewTableSources(&p.Postgresql, metrics)[t.Name()]

	require.NoError(t, p.tableManager.MatchSource(ctx, p.db, tsrc))
	assert.Contains(t, p.tableManager.table(t.Name() + p.TagTableSuffix).Columns(), "tag")
	assert.Contains(t, p.tableManager.table(t.Name()).Columns(), "a")
}

func TestTableManager_noCreateTable(t *testing.T) {
	p := newPostgresqlTest(t)
	p.CreateTemplates = nil
	require.NoError(t, p.Connect())

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"tag": "foo"}, MSI{"a": 1}),
	}
	tsrc := NewTableSources(&p.Postgresql, metrics)[t.Name()]

	require.Error(t, p.tableManager.MatchSource(ctx, p.db, tsrc))
}

func TestTableManager_noCreateTagTable(t *testing.T) {
	p := newPostgresqlTest(t)
	p.TagTableCreateTemplates = nil
	p.TagsAsForeignKeys = true
	require.NoError(t, p.Connect())

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"tag": "foo"}, MSI{"a": 1}),
	}
	tsrc := NewTableSources(&p.Postgresql, metrics)[t.Name()]

	require.Error(t, p.tableManager.MatchSource(ctx, p.db, tsrc))
}

// verify that TableManager updates & caches the DB table structure unless the incoming metric can't fit.
func TestTableManager_cache(t *testing.T) {
	p := newPostgresqlTest(t)
	p.TagsAsForeignKeys = true
	require.NoError(t, p.Connect())

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"tag": "foo"}, MSI{"a": 1}),
	}
	tsrc := NewTableSources(&p.Postgresql, metrics)[t.Name()]

	require.NoError(t, p.tableManager.MatchSource(ctx, p.db, tsrc))
}

// Verify that when alter statements are disabled and a metric comes in with a new tag key, that the tag is omitted.
func TestTableManager_noAlterMissingTag(t *testing.T) {
	p := newPostgresqlTest(t)
	p.AddColumnTemplates = []*template.Template{}
	require.NoError(t, p.Connect())

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"tag": "foo"}, MSI{"a": 1}),
	}
	tsrc := NewTableSources(&p.Postgresql, metrics)[t.Name()]
	require.NoError(t, p.tableManager.MatchSource(ctx, p.db, tsrc))

	metrics = []telegraf.Metric{
		newMetric(t, "", MSS{"tag": "foo"}, MSI{"a": 2}),
		newMetric(t, "", MSS{"tag": "foo", "bar": "baz"}, MSI{"a": 3}),
	}
	tsrc = NewTableSources(&p.Postgresql, metrics)[t.Name()]
	require.NoError(t, p.tableManager.MatchSource(ctx, p.db, tsrc))
	assert.NotContains(t, tsrc.ColumnNames(), "bar")
}

// Verify that when alter statements are disabled with foreign tags and a metric comes in with a new tag key, that the
// field is omitted.
func TestTableManager_noAlterMissingTagTableTag(t *testing.T) {
	p := newPostgresqlTest(t)
	p.TagsAsForeignKeys = true
	p.TagTableAddColumnTemplates = []*template.Template{}
	require.NoError(t, p.Connect())

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"tag": "foo"}, MSI{"a": 1}),
	}
	tsrc := NewTableSources(&p.Postgresql, metrics)[t.Name()]
	require.NoError(t, p.tableManager.MatchSource(ctx, p.db, tsrc))

	metrics = []telegraf.Metric{
		newMetric(t, "", MSS{"tag": "foo"}, MSI{"a": 2}),
		newMetric(t, "", MSS{"tag": "foo", "bar": "baz"}, MSI{"a": 3}),
	}
	tsrc = NewTableSources(&p.Postgresql, metrics)[t.Name()]
	ttsrc := NewTagTableSource(tsrc)
	require.NoError(t, p.tableManager.MatchSource(ctx, p.db, tsrc))
	assert.NotContains(t, ttsrc.ColumnNames(), "bar")
}

// verify that when alter statements are disabled and a metric comes in with a new field key, that the field is omitted.
func TestTableManager_noAlterMissingField(t *testing.T) {
	p := newPostgresqlTest(t)
	p.AddColumnTemplates = []*template.Template{}
	require.NoError(t, p.Connect())

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"tag": "foo"}, MSI{"a": 1}),
	}
	tsrc := NewTableSources(&p.Postgresql, metrics)[t.Name()]
	require.NoError(t, p.tableManager.MatchSource(ctx, p.db, tsrc))

	metrics = []telegraf.Metric{
		newMetric(t, "", MSS{"tag": "foo"}, MSI{"a": 2}),
		newMetric(t, "", MSS{"tag": "foo"}, MSI{"a": 3, "b":3}),
	}
	tsrc = NewTableSources(&p.Postgresql, metrics)[t.Name()]
	require.NoError(t, p.tableManager.MatchSource(ctx, p.db, tsrc))
	assert.NotContains(t, tsrc.ColumnNames(), "b")
}
