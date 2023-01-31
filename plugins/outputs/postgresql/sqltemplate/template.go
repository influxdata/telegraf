// Package sqltemplate
/*
Templates are used for creation of the SQL used when creating and modifying tables. These templates are specified within
the configuration as the parameters 'create_templates', 'add_column_templates', 'tag_table_create_templates', and
'tag_table_add_column_templates'.

The templating functionality behaves the same in all cases. However, the variables will differ.

# Variables

The following variables are available within all template executions:

  - table - A Table object referring to the current table being
    created/modified.

  - columns - A Columns object of the new columns being added to the
    table (all columns in the case of a new table, and new columns in the case
    of existing table).

  - allColumns - A Columns object of all the columns (both old and new)
    of the table. In the case of a new table, this is the same as `columns`.

  - metricTable - A Table object referring to the table containing the
    fields. In the case of TagsAsForeignKeys and `table` is the tag table, then
    `metricTable` is the table using this one for its tags.

  - tagTable - A Table object referring to the table containing the
    tags. In the case of TagsAsForeignKeys and `table` is the metrics table,
    then `tagTable` is the table containing the tags for it.

Each object has helper methods that may be used within the template. See the documentation for the appropriate type.

When the object is interpolated without a helper, it is automatically converted to a string through its String() method.

# Functions

All the functions provided by the Sprig library (http://masterminds.github.io/sprig/) are available within template executions.

In addition, the following functions are also available:

  - quoteIdentifier - Quotes the input string as a Postgres identifier.

  - quoteLiteral - Quotes the input string as a Postgres literal.

# Examples

The default templates show basic usage. When left unconfigured, it is the equivalent of:

	[outputs.postgresql]
	  create_templates = [
	    '''CREATE TABLE {{.table}} ({{.columns}})''',
	  ]
	  add_column_templates = [
	    '''ALTER TABLE {{.table}} ADD COLUMN IF NOT EXISTS {{.columns|join ", ADD COLUMN IF NOT EXISTS "}}''',
	  ]
	  tag_table_create_templates = [
	    '''CREATE TABLE {{.table}} ({{.columns}}, PRIMARY KEY (tag_id))'''
	  ]
	  tag_table_add_column_templates = [
	    '''ALTER TABLE {{.table}} ADD COLUMN IF NOT EXISTS {{.columns|join ", ADD COLUMN IF NOT EXISTS "}}''',
	  ]

A simple example for usage with TimescaleDB would be:

	[outputs.postgresql]
	  create_templates = [
	    '''CREATE TABLE {{ .table }} ({{ .allColumns }})''',
	    '''SELECT create_hypertable({{ .table|quoteLiteral }}, 'time', chunk_time_interval => INTERVAL '1d')''',
	    '''ALTER TABLE {{ .table }} SET (timescaledb.compress, timescaledb.compress_segmentby = 'tag_id')''',
	    '''SELECT add_compression_policy({{ .table|quoteLiteral }}, INTERVAL '2h')''',
	  ]

...where the defaults for the other templates would be automatically applied.

A very complex example for versions of TimescaleDB which don't support adding columns to compressed hypertables (v<2.1.0),
using views and unions to emulate the functionality, would be:

	[outputs.postgresql]
	  schema = "telegraf"
	  create_templates = [
	    '''CREATE TABLE {{ .table }} ({{ .allColumns }})''',
	    '''SELECT create_hypertable({{ .table|quoteLiteral }}, 'time', chunk_time_interval => INTERVAL '1d')''',
	    '''ALTER TABLE {{ .table }} SET (timescaledb.compress, timescaledb.compress_segmentby = 'tag_id')''',
	    '''SELECT add_compression_policy({{ .table|quoteLiteral }}, INTERVAL '2d')''',
	    '''CREATE VIEW {{ .table.WithSuffix "_data" }} AS
	         SELECT {{ .allColumns.Selectors | join "," }} FROM {{ .table }}''',
	    '''CREATE VIEW {{ .table.WithSchema "public" }} AS
	         SELECT time, {{ (.tagTable.Columns.Tags.Concat .allColumns.Fields).Identifiers | join "," }}
	         FROM {{ .table.WithSuffix "_data" }} t, {{ .tagTable }} tt
	         WHERE t.tag_id = tt.tag_id''',
	  ]
	  add_column_templates = [
	    '''ALTER TABLE {{ .table }} RENAME TO {{ (.table.WithSuffix "_" .table.Columns.Hash).WithSchema "" }}''',
	    '''ALTER VIEW {{ .table.WithSuffix "_data" }} RENAME TO {{ (.table.WithSuffix "_" .table.Columns.Hash "_data").WithSchema "" }}''',
	    '''DROP VIEW {{ .table.WithSchema "public" }}''',

	    '''CREATE TABLE {{ .table }} ({{ .allColumns }})''',
	    '''SELECT create_hypertable({{ .table|quoteLiteral }}, 'time', chunk_time_interval => INTERVAL '1d')''',
	    '''ALTER TABLE {{ .table }} SET (timescaledb.compress, timescaledb.compress_segmentby = 'tag_id')''',
	    '''SELECT add_compression_policy({{ .table|quoteLiteral }}, INTERVAL '2d')''',
	    '''CREATE VIEW {{ .table.WithSuffix "_data" }} AS
	         SELECT {{ .allColumns.Selectors | join "," }}
	         FROM {{ .table }}
	         UNION ALL
	         SELECT {{ (.allColumns.Union .table.Columns).Selectors | join "," }}
	         FROM {{ .table.WithSuffix "_" .table.Columns.Hash "_data" }}''',
	    '''CREATE VIEW {{ .table.WithSchema "public" }}
	         AS SELECT time, {{ (.tagTable.Columns.Tags.Concat .allColumns.Fields).Identifiers | join "," }}
	         FROM {{ .table.WithSuffix "_data" }} t, {{ .tagTable }} tt
	         WHERE t.tag_id = tt.tag_id''',
	  ]
*/
package sqltemplate

import (
	"bytes"
	"encoding/base32"
	"fmt"
	"hash/fnv"
	"strings"
	"text/template"
	"unsafe"

	"github.com/Masterminds/sprig"

	"github.com/influxdata/telegraf/plugins/outputs/postgresql/utils"
)

var templateFuncs = map[string]interface{}{
	"quoteIdentifier": QuoteIdentifier,
	"quoteLiteral":    QuoteLiteral,
}

func asString(obj interface{}) string {
	switch obj := obj.(type) {
	case string:
		return obj
	case []byte:
		return string(obj)
	case fmt.Stringer:
		return obj.String()
	default:
		return fmt.Sprintf("%v", obj)
	}
}

// QuoteIdentifier quotes the given string as a Postgres identifier (double-quotes the value).
//
// QuoteIdentifier is accessible within templates as 'quoteIdentifier'.
func QuoteIdentifier(name interface{}) string {
	return utils.QuoteIdentifier(asString(name))
}

// QuoteLiteral quotes the given string as a Postgres literal (single-quotes the value).
//
// QuoteLiteral is accessible within templates as 'quoteLiteral'.
func QuoteLiteral(str interface{}) string {
	return utils.QuoteLiteral(asString(str))
}

// Table is an object which represents a Postgres table.
type Table struct {
	Schema  string
	Name    string
	Columns Columns
}

func NewTable(schemaName, tableName string, columns []utils.Column) *Table {
	if tableName == "" {
		return nil
	}
	return &Table{
		Schema:  schemaName,
		Name:    tableName,
		Columns: NewColumns(columns),
	}
}

// String returns the table's fully qualified & quoted identifier (schema+table).
func (tbl *Table) String() string {
	return tbl.Identifier()
}

// Identifier returns the table's fully qualified & quoted identifier (schema+table).
//
// If schema is empty, it is omitted from the result.
func (tbl *Table) Identifier() string {
	if tbl.Schema == "" {
		return QuoteIdentifier(tbl.Name)
	}
	return QuoteIdentifier(tbl.Schema) + "." + QuoteIdentifier(tbl.Name)
}

// WithSchema returns a copy of the Table object, but with the schema replaced by the given value.
func (tbl *Table) WithSchema(name string) *Table {
	tblNew := &Table{}
	*tblNew = *tbl
	tblNew.Schema = name
	return tblNew
}

// WithName returns a copy of the Table object, but with the name replaced by the given value.
func (tbl *Table) WithName(name string) *Table {
	tblNew := &Table{}
	*tblNew = *tbl
	tblNew.Name = name
	return tblNew
}

// WithSuffix returns a copy of the Table object, but with the name suffixed with the given value.
func (tbl *Table) WithSuffix(suffixes ...string) *Table {
	tblNew := &Table{}
	*tblNew = *tbl
	tblNew.Name += strings.Join(suffixes, "")
	return tblNew
}

// A Column is an object which represents a Postgres column.
type Column utils.Column

// String returns the column's definition (as used in a CREATE TABLE statement). E.G:
//
//	"my_column" bigint
func (tc Column) String() string {
	return tc.Definition()
}

// Definition returns the column's definition (as used in a CREATE TABLE statement). E.G:
//
//	"my_column" bigint
func (tc Column) Definition() string {
	return tc.Identifier() + " " + tc.Type
}

// Identifier returns the column's quoted identifier.
func (tc Column) Identifier() string {
	return QuoteIdentifier(tc.Name)
}

// Selector returns the selector for the column. For most cases this is the same as Identifier.
// However, in some cases, such as a UNION, this may return a statement such as `NULL AS "foo"`.
func (tc Column) Selector() string {
	if tc.Type != "" {
		return tc.Identifier()
	}
	return "NULL AS " + tc.Identifier()
}

// IsTag returns true if the column is a tag column. Otherwise, false.
func (tc Column) IsTag() bool {
	return tc.Role == utils.TagColType
}

// IsField returns true if the column is a field column. Otherwise, false.
func (tc Column) IsField() bool {
	return tc.Role == utils.FieldColType
}

// Columns represents an ordered list of Column objects, with convenience methods for operating on the
// list.
type Columns []Column

func NewColumns(cols []utils.Column) Columns {
	tcols := make(Columns, len(cols))
	for i, col := range cols {
		tcols[i] = Column(col)
	}
	return tcols
}

// List returns the Columns object as a slice of Column.
func (cols Columns) List() []Column {
	return cols
}

// Definitions returns the list of column definitions.
func (cols Columns) Definitions() []string {
	defs := make([]string, len(cols))
	for i, tc := range cols {
		defs[i] = tc.Definition()
	}
	return defs
}

// Identifiers returns the list of quoted column identifiers.
func (cols Columns) Identifiers() []string {
	idents := make([]string, len(cols))
	for i, tc := range cols {
		idents[i] = tc.Identifier()
	}
	return idents
}

// Selectors returns the list of column selectors.
func (cols Columns) Selectors() []string {
	selectors := make([]string, len(cols))
	for i, tc := range cols {
		selectors[i] = tc.Selector()
	}
	return selectors
}

// String returns the comma delimited list of column identifiers.
func (cols Columns) String() string {
	colStrs := make([]string, len(cols))
	for i, tc := range cols {
		colStrs[i] = tc.String()
	}
	return strings.Join(colStrs, ", ")
}

// Keys returns a Columns list of the columns which are not fields (e.g. time, tag_id, & tags).
func (cols Columns) Keys() Columns {
	var newCols []Column
	for _, tc := range cols {
		if tc.Role != utils.FieldColType {
			newCols = append(newCols, tc)
		}
	}
	return newCols
}

// Sorted returns a sorted copy of Columns.
//
// Columns are sorted so that they are in order as: [Time, Tags, Fields], with the columns within each group sorted
// alphabetically.
func (cols Columns) Sorted() Columns {
	newCols := append([]Column{}, cols...)
	(*utils.ColumnList)(unsafe.Pointer(&newCols)).Sort()
	return newCols
}

// Concat returns a copy of Columns with the given tcsList appended to the end.
func (cols Columns) Concat(tcsList ...Columns) Columns {
	tcsNew := append(Columns{}, cols...)
	for _, tcs := range tcsList {
		tcsNew = append(tcsNew, tcs...)
	}
	return tcsNew
}

// Union generates a list of SQL selectors against the given columns.
//
// For each column in tcs, if the column also exist in tcsFrom, it will be selected. If the column does not exist NULL will be selected.
func (cols Columns) Union(tcsFrom Columns) Columns {
	tcsNew := append(Columns{}, cols...)
TCS:
	for i, tc := range cols {
		for _, tcFrom := range tcsFrom {
			if tc.Name == tcFrom.Name {
				continue TCS
			}
		}
		tcsNew[i].Type = ""
	}
	return tcsNew
}

// Tags returns a Columns list of the columns which are tags.
func (cols Columns) Tags() Columns {
	var newCols []Column
	for _, tc := range cols {
		if tc.Role == utils.TagColType {
			newCols = append(newCols, tc)
		}
	}
	return newCols
}

// Fields returns a Columns list of the columns which are fields.
func (cols Columns) Fields() Columns {
	var newCols []Column
	for _, tc := range cols {
		if tc.Role == utils.FieldColType {
			newCols = append(newCols, tc)
		}
	}
	return newCols
}

// Hash returns a hash of the column names. The hash is base-32 encoded string, up to 7 characters long with no padding.
//
// This can be useful as an identifier for supporting table renaming + unions in the case of non-modifiable tables.
func (cols Columns) Hash() string {
	hash := fnv.New32a()
	for _, tc := range cols.Sorted() {
		hash.Write([]byte(tc.Name)) //nolint:revive // all Write() methods for hash in fnv.go returns nil err
		hash.Write([]byte{0})       //nolint:revive // all Write() methods for hash in fnv.go returns nil err
	}
	return strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(hash.Sum(nil)))
}

type Template template.Template

func (t *Template) UnmarshalText(text []byte) error {
	tmpl := template.New("")
	tmpl.Option("missingkey=error")
	tmpl.Funcs(templateFuncs)
	tmpl.Funcs(sprig.TxtFuncMap())
	tt, err := tmpl.Parse(string(text))
	if err != nil {
		return err
	}
	*t = Template(*tt)
	return nil
}

func (t *Template) Render(table *Table, newColumns []utils.Column, metricTable *Table, tagTable *Table) ([]byte, error) {
	tcs := NewColumns(newColumns).Sorted()
	data := map[string]interface{}{
		"table":       table,
		"columns":     tcs,
		"allColumns":  tcs.Concat(table.Columns).Sorted(),
		"metricTable": metricTable,
		"tagTable":    tagTable,
	}

	buf := bytes.NewBuffer(nil)
	err := (*template.Template)(t).Execute(buf, data)
	return buf.Bytes(), err
}
