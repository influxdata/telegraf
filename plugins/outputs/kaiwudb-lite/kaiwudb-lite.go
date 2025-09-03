//go:generate ../../../tools/readme_config_includer/generator
package kaiwudb

import (
	"cmp"
	gosql "database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"iter"
	"slices"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

var defaultConvert = ConvertStruct{
	Integer:         "INT",
	UInteger:        "UINTEGER",
	Bigint:          "BIGINT",
	UBigint:         "UBIGINT",
	Real:            "REAL",
	Double:          "DOUBLE",
	Text:            "TEXT",
	Timestamp:       "TIMESTAMP",
	Timestamptz:     "TIMESTAMP WITH TIME ZONE",
	Defaultvalue:    "TEXT",
	Unsigned:        "UNSIGNED",
	Bool:            "BOOL",
	Json:            "JSON",
	Blob:            "BLOB",
	ConversionStyle: "unsigned_suffix",
}

type ConvertStruct struct {
	Integer         string `toml:"integer"`
	UInteger        string `toml:"uinteger"`
	Bigint          string `toml:"bigint"`
	UBigint         string `toml:"ubigint"`
	Real            string `toml:"real"`
	Double          string `toml:"double"`
	Text            string `toml:"text"`
	Timestamp       string `toml:"timestamp"`
	Timestamptz     string `toml:"timestamptz"`
	Defaultvalue    string `toml:"defaultvalue"`
	Unsigned        string `toml:"unsigned"`
	Bool            string `toml:"bool"`
	Json            string `toml:"json"`
	Blob            string `toml:"blob"`
	ConversionStyle string `toml:"conversion_style"`
}

const (
	Individual        string = "individual"
	BatchTx           string = "batch_transactions"
	MultiValuesPerSQL string = "multiple_values_per_sql"
)

type Kaiwudb struct {
	Driver                string          `toml:"driver"`
	DataSourceName        config.Secret   `toml:"data_source_name"`
	EnableCompactSchema   bool            `toml:"enable_compact_schema"`
	TimestampColumnName   string          `toml:"timestamp_column_name"`
	TimestampWithTZ       bool            `toml:"timestamp_with_time_zone"`
	TimeZone              string          `toml:"timezone"`
	TagsColumnName        string          `toml:"tags_column_name"`
	FieldsColumnName      string          `toml:"fileds_column_name"`
	TableTemplate         string          `toml:"table_template"`
	TableExistsTemplate   string          `toml:"table_exists_template"`
	TableUpdateTemplate   string          `toml:"table_update_template"`
	InitSQL               string          `toml:"init_sql"`
	Mode                  string          `toml:"mode"`
	Convert               ConvertStruct   `toml:"convert"`
	ConnectionMaxIdleTime config.Duration `toml:"connection_max_idle_time"`
	ConnectionMaxLifetime config.Duration `toml:"connection_max_lifetime"`
	ConnectionMaxIdle     int             `toml:"connection_max_idle"`
	ConnectionMaxOpen     int             `toml:"connection_max_open"`
	Log                   telegraf.Logger `toml:"-"`

	db                       *gosql.DB
	queryCache               map[string]string
	tables                   map[string]map[string]bool
	tableListColumnsTemplate string
}

func FieldListToJSON(fieldList []*telegraf.Field) ([]byte, error) {
	fields := make(map[string]interface{}, len(fieldList))
	for _, field := range fieldList {
		fields[field.Key] = field.Value
	}
	return json.Marshal(fields)
}

func MapToKaiwuDBDBMapExprSafe(tagList []*telegraf.Tag) string {
	var builder strings.Builder
	builder.WriteString("MAP{")

	for i, tag := range tagList {
		if i != 0 {
			builder.WriteString(", ")
		}
		valStrkey := fmt.Sprint(tag.Key)
		valStr := fmt.Sprint(tag.Value)
		escapedKey := strings.ReplaceAll(valStrkey, `'`, `''`)
		escapedVal := strings.ReplaceAll(valStr, `'`, `''`)
		builder.WriteString(fmt.Sprintf("'%s':'%s'", escapedKey, escapedVal))
	}

	builder.WriteString("}")
	return builder.String()
}

func (*Kaiwudb) SampleConfig() string {
	return sampleConfig
}

func (p *Kaiwudb) Init() error {
	// Set defaults
	if p.TableExistsTemplate == "" {
		p.TableExistsTemplate = "SELECT 1 FROM {TABLE} LIMIT 1"
	}

	if p.TableTemplate == "" {
		p.TableTemplate = "CREATE TABLE {TABLE}({COLUMNS})"
	}

	p.tableListColumnsTemplate = "SELECT column_name FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME={TABLE}"
	if p.Driver == "sqlite" {
		p.tableListColumnsTemplate = "SELECT name AS column_name FROM pragma_table_info({TABLE})"
	}

	// Check for a valid driver
	switch p.Driver {
	case "kaiwudb":
		return fmt.Errorf("Kaiwudb is currently not supported, only Kaiwudb-lite")
	case "kaiwudb-lite":
		// Do nothing, those are valid
	default:
		return fmt.Errorf("unknown driver %q", p.Driver)
	}

	// check column name
	if p.EnableCompactSchema {
		if p.TagsColumnName == "" {
			p.TagsColumnName = "tags"
		}
		if p.FieldsColumnName == "" {
			p.FieldsColumnName = "fields"
		}
	}

	p.Log.Debugf("driver                      : %s", p.Driver)
	p.Log.Debugf("data_source_name            : %s", p.DataSourceName)
	p.Log.Debugf("enable_compact_schema       : %t", p.EnableCompactSchema)
	p.Log.Debugf("timestamp_column            : %s", p.TimestampColumnName)
	p.Log.Debugf("timestamp_with_time_zone    : %t", p.TimestampWithTZ)
	p.Log.Debugf("tags_column                 : %s", p.TagsColumnName)
	p.Log.Debugf("fileds_column_name          : %s", p.FieldsColumnName)
	p.Log.Debugf("mode                        : %s", p.Mode)
	return nil
}

func (p *Kaiwudb) Connect() error {
	dsnBuffer, err := p.DataSourceName.Get()
	if err != nil {
		return fmt.Errorf("loading data source name secret failed: %w", err)
	}
	dsn := dsnBuffer.String()
	dsnBuffer.Destroy()

	// TODO: kaiwudb should use different driver
	driver_name := p.Driver
	if p.Driver == "kaiwudb" || p.Driver == "kaiwudb-lite" {
		driver_name = "pgx"
	}
	db, err := gosql.Open(driver_name, dsn)
	if err != nil {
		return fmt.Errorf("creating database client failed: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("pinging database failed: %w", err)
	}

	db.SetConnMaxIdleTime(time.Duration(p.ConnectionMaxIdleTime))
	db.SetConnMaxLifetime(time.Duration(p.ConnectionMaxLifetime))
	db.SetMaxIdleConns(p.ConnectionMaxIdle)
	db.SetMaxOpenConns(p.ConnectionMaxOpen)

	if p.InitSQL != "" {
		p.Log.Debugf("Executing InitSQL: %s", p.InitSQL)
		if _, err = db.Exec(p.InitSQL); err != nil {
			return fmt.Errorf("initializing database failed: %w", err)
		}
	}

	p.db = db
	p.tables = make(map[string]map[string]bool)
	p.queryCache = make(map[string]string)

	return nil
}

func (p *Kaiwudb) Close() error {
	return p.db.Close()
}

// Quote an identifier (table or column name)
func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(sanitizeQuoted(name), `"`, `""`) + `"`
}

// Quote a string literal
func quoteStr(name string) string {
	return "'" + strings.ReplaceAll(name, "'", "''") + "'"
}

func sanitizeQuoted(in string) string {
	// https://dev.mysql.com/doc/refman/8.0/en/identifiers.html
	// https://www.postgresql.org/docs/13/sql-syntax-lexical.html#SQL-SYNTAX-IDENTIFIERS

	// Whitelist allowed characters
	return strings.Map(func(r rune) rune {
		switch {
		case r >= '\u0001' && r <= '\uFFFF':
			return r
		default:
			return '_'
		}
	}, in)
}

func (p *Kaiwudb) deriveDatatype(value interface{}) string {
	var datatype string

	switch value.(type) {
	case int8, int16, int32:
		datatype = p.Convert.Integer
	case int64:
		datatype = p.Convert.Bigint
	case uint8, uint16, uint32:
		datatype = p.Convert.UInteger
	case uint64:
		datatype = p.Convert.UBigint
	case float32:
		datatype = p.Convert.Real
	case float64:
		datatype = p.Convert.Double
	case string:
		datatype = p.Convert.Text
	case bool:
		datatype = p.Convert.Bool
	case time.Time:
		datatype = p.Convert.Timestamp
	case []byte:
		datatype = p.Convert.Blob
	default:
		datatype = p.Convert.Defaultvalue
		p.Log.Errorf("Unknown datatype: '%T' %v", value, value)
	}
	return datatype
}

func (p *Kaiwudb) generateCreateTable(metric telegraf.Metric) string {
	columns_len := 0
	columns_len_tag := 0
	columns_len_fields := 0

	if p.TimestampColumnName != "" {
		columns_len += 1
	}
	if p.EnableCompactSchema {
		columns_len_tag = 1
		columns_len_fields = 1
	} else {
		columns_len_tag = len(metric.TagList())
		columns_len_fields = len(metric.FieldList())
	}
	columns_len += (columns_len_tag + columns_len_fields)
	columns := make([]string, 0, columns_len)
	tagColumnNames := make([]string, 0, columns_len_tag)

	if p.TimestampColumnName != "" {
		if !p.TimestampWithTZ {
			columns = append(columns, fmt.Sprintf("%s %s", quoteIdent(p.TimestampColumnName), p.Convert.Timestamp))
		} else {
			columns = append(columns, fmt.Sprintf("%s %s", quoteIdent(p.TimestampColumnName), p.Convert.Timestamptz))
		}
	}

	if !p.EnableCompactSchema {
		for _, tag := range metric.TagList() {
			columns = append(columns, fmt.Sprintf("%s %s", quoteIdent(tag.Key), p.Convert.Text))
			tagColumnNames = append(tagColumnNames, quoteIdent(tag.Key))
		}
	} else {
		columns = append(columns, fmt.Sprintf("%s MAP(%s, %s)", quoteIdent(p.TagsColumnName), p.Convert.Text, p.Convert.Text))
	}

	if !p.EnableCompactSchema {
		var datatype string
		for _, field := range metric.FieldList() {
			datatype = p.deriveDatatype(field.Value)
			columns = append(columns, fmt.Sprintf("%s %s", quoteIdent(field.Key), datatype))
		}
	} else {
		columns = append(columns, fmt.Sprintf("%s %s", quoteIdent(p.FieldsColumnName), p.Convert.Json))
	}

	query := p.TableTemplate
	query = strings.ReplaceAll(query, "{TABLE}", quoteIdent(metric.Name()))
	query = strings.ReplaceAll(query, "{TABLELITERAL}", quoteStr(metric.Name()))
	query = strings.ReplaceAll(query, "{COLUMNS}", strings.Join(columns, ","))
	query = strings.ReplaceAll(query, "{TAG_COLUMN_NAMES}", strings.Join(tagColumnNames, ","))
	query = strings.ReplaceAll(query, "{TIMESTAMP_COLUMN_NAME}", quoteIdent(p.TimestampColumnName))

	p.Log.Debugf("Creating table with str: %s", query)
	return query
}

func (p *Kaiwudb) generateAddColumn(tablename, column, columnType string) string {
	query := p.TableUpdateTemplate
	query = strings.ReplaceAll(query, "{TABLE}", quoteIdent(tablename))
	query = strings.ReplaceAll(query, "{COLUMN}", quoteIdent(column)+" "+columnType)
	p.Log.Debugf("Add column query: %s", query)
	return query
}

func (p *Kaiwudb) generateInsert(tablename string, columns []string) string {
	placeholders := make([]string, 0, len(columns))
	quotedColumns := make([]string, 0, len(columns))

	for _, column := range columns {
		quotedColumns = append(quotedColumns, quoteIdent(column))
		if p.EnableCompactSchema && p.Mode != MultiValuesPerSQL && column == p.TagsColumnName {
			placeholders = append(placeholders, "%s")
		} else {
			placeholders = append(placeholders, "?")
		}
	}

	if p.Mode == Individual || p.Mode == BatchTx {
		return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			quoteIdent(tablename),
			strings.Join(quotedColumns, ","),
			strings.Join(placeholders, ","))
	} else {
		return fmt.Sprintf("INSERT INTO %s (%s) VALUES",
			quoteIdent(tablename),
			strings.Join(quotedColumns, ","))
	}
}

func (p *Kaiwudb) createTable(metric telegraf.Metric) error {
	tablename := metric.Name()
	stmt := p.generateCreateTable(metric)
	if _, err := p.db.Exec(stmt); err != nil {
		return fmt.Errorf("creating table failed: %w", err)
	}
	// Ensure compatibility: set the table cache to an empty map
	p.tables[tablename] = make(map[string]bool)
	// Modifying the table schema is opt-in
	if p.TableUpdateTemplate != "" {
		if err := p.updateTableCache(tablename); err != nil {
			return fmt.Errorf("updating table cache failed: %w", err)
		}
	}
	return nil
}

func (p *Kaiwudb) createColumn(tablename, column, columnType string) error {
	// Ensure table exists in cache before accessing columns
	if _, tableExists := p.tables[tablename]; !tableExists {
		if err := p.updateTableCache(tablename); err != nil {
			return fmt.Errorf("updating table cache failed: %w", err)
		}
	}
	// Ensure column existence check doesn't panic
	if _, tableExists := p.tables[tablename]; !tableExists {
		return fmt.Errorf("table %s does not exist in cache", tablename)
	}
	// Column already exists, nothing to do
	if exists, colExists := p.tables[tablename][column]; colExists && exists {
		return nil
	}
	// Generate and execute column addition statement
	createColumn := p.generateAddColumn(tablename, column, columnType)
	if _, err := p.db.Exec(createColumn); err != nil {
		return fmt.Errorf("creating column failed: %w", err)
	}
	// Update cache after adding the column
	if err := p.updateTableCache(tablename); err != nil {
		return fmt.Errorf("updating table cache failed: %w", err)
	}
	return nil
}

func (p *Kaiwudb) tableExists(tableName string) bool {
	stmt := strings.ReplaceAll(p.TableExistsTemplate, "{TABLE}", quoteIdent(tableName))

	_, err := p.db.Exec(stmt)
	return err == nil
}

func (p *Kaiwudb) updateTableCache(tablename string) error {
	stmt := strings.ReplaceAll(p.tableListColumnsTemplate, "{TABLE}", quoteStr(tablename))

	columns, err := p.db.Query(stmt)
	if err != nil {
		return fmt.Errorf("fetching columns for table(%s) failed: %w", tablename, err)
	}
	defer columns.Close()

	if p.tables[tablename] == nil {
		p.tables[tablename] = make(map[string]bool)
	}

	for columns.Next() {
		var columnName string
		if err := columns.Scan(&columnName); err != nil {
			return err
		}

		if !p.tables[tablename][columnName] {
			p.tables[tablename][columnName] = true
		}
	}

	return nil
}

func (p *Kaiwudb) processMetric(metric telegraf.Metric) (string, []string, []interface{}) {
	// Preallocate the columns and values. Note we always allocate for the
	// timestamp column even if we don't need it but that's not an issue.
	entries := len(metric.TagList()) + len(metric.FieldList()) + 1
	columns := make([]string, 0, entries)
	values := make([]interface{}, 0, entries)
	if p.TimestampColumnName != "" {
		columns = append(columns, p.TimestampColumnName)
		values = append(values, metric.Time())
	}
	// Tags are already sorted so we can add them without modification
	for _, tag := range metric.TagList() {
		columns = append(columns, tag.Key)
		values = append(values, tag.Value)
	}
	// Fields are not sorted so sort them
	fields := slices.SortedFunc(
		iterSlice(metric.FieldList()),
		func(a, b *telegraf.Field) int { return cmp.Compare(a.Key, b.Key) },
	)
	for _, field := range fields {
		columns = append(columns, field.Key)
		values = append(values, field.Value)
	}
	return strings.Join(append([]string{metric.Name()}, columns...), "\n"), columns, values
}

func (p *Kaiwudb) processMetricCompact(metric telegraf.Metric) (string, []string, string, []interface{}) {
	if !p.EnableCompactSchema {
		p.Log.Errorf("When call this function, p.EnableCompactSchema should be true. Otherwise, the function processMetric should be called.")
		return "", nil, "", nil
	}

	// Preallocate the columns and values. Note we always allocate for the
	// timestamp column even if we don't need it but that's not an issue.
	entries := 2 + 1
	columns := make([]string, 0, entries)
	values := make([]interface{}, 0, entries)
	var tag_values string

	if p.TimestampColumnName != "" {
		columns = append(columns, p.TimestampColumnName)
		values = append(values, metric.Time())
	}
	// Tags are already sorted so we can add them without modification
	columns = append(columns, p.TagsColumnName)

	if p.EnableCompactSchema && p.Mode != MultiValuesPerSQL {
		tag_values = MapToKaiwuDBDBMapExprSafe(metric.TagList())
	} else {
		values = append(values, MapToKaiwuDBDBMapExprSafe(metric.TagList()))
	}
	// Fields are not sorted so sort them
	fields := slices.SortedFunc(
		iterSlice(metric.FieldList()),
		func(a, b *telegraf.Field) int { return cmp.Compare(a.Key, b.Key) },
	)
	columns = append(columns, p.FieldsColumnName)
	value_fields, _ := FieldListToJSON(fields)
	values = append(values, value_fields)
	return strings.Join(append(append([]string{metric.Name()}, columns...), tag_values), "\n"), columns, tag_values, values
}

func (p *Kaiwudb) sendIndividual(sql string, values []interface{}) error {
	switch p.Driver {
	default:
		_, err := p.db.Exec(sql, values...)
		if err != nil {
			return fmt.Errorf("execution failed: %w", err)
		}
	}

	return nil
}

func (p *Kaiwudb) sendBatch(sql string, values [][]interface{}) error {
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("begin failed: %w", err)
	}

	batch, err := tx.Prepare(sql)
	if err != nil {
		return fmt.Errorf("prepare failed: %w", err)
	}
	defer batch.Close()

	for _, params := range values {
		if _, err := batch.Exec(params...); err != nil {
			if errRollback := tx.Rollback(); errRollback != nil {
				return fmt.Errorf("execution failed: %w, unable to rollback: %w", err, errRollback)
			}
			return fmt.Errorf("execution failed: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}

// TODO, deal upper bound allowed of string
func (p *Kaiwudb) sendMultiValues(sql string, multipleValues [][]interface{}) error {
	if len(multipleValues) == 0 {
		return nil
	}

	var builder strings.Builder
	builder.WriteString(sql)
	for i, row := range multipleValues {
		builder.WriteString("(")
		for j, val := range row {
			switch v := val.(type) {
			case time.Time:
				formated_v := v.Format("2006-01-02 15:04:05.000000")
				builder.WriteString(quoteStr(formated_v))
			case string:
				if p.EnableCompactSchema && p.TimestampColumnName != "" && j == 1 {
					builder.WriteString(v)
				} else {
					builder.WriteString(quoteStr(v))
				}
			case []byte:
				builder.WriteString(quoteStr(string(v)))
			default:
				builder.WriteString(fmt.Sprintf("%v", v)) // any other types
			}
			if j < len(row)-1 {
				builder.WriteString(", ")
			}
		}
		builder.WriteString(")")
		if i < len(multipleValues)-1 {
			builder.WriteString(", ")
		}
	}

	builder.WriteString(";")
	// p.Log.Debugf("sql: %s", builder.String())
	switch p.Driver {
	default:
		// _, err := p.db.Exec("INSERT INTO cpu (ts,tags,fields) VALUES('2025-08-02 02:38:29.099989', MAP{'cpu':'cpu1', 'host':'node'}, {\"usage_guest\":1})")
		_, err := p.db.Exec(builder.String())
		if err != nil {
			return fmt.Errorf("execution failed: %w", err)
		}
	}

	return nil
}

func (p *Kaiwudb) Write(metrics []telegraf.Metric) error {
	batchedQueries := make(map[string][][]interface{})

	p.Log.Debugf("Metrics size: %d", len(metrics))
	for i, metric := range metrics {
		tablename := metric.Name()
		if i == 0 {
			p.Log.Debugf("metric[0]: timestamp: %s, name: %s, tags: %s, fields: %",
				metric.Time().Format(time.RFC3339), tablename, metric.Tags(), metric.Fields())
		}

		// create table if needed
		if _, found := p.tables[tablename]; !found && !p.tableExists(tablename) {
			if err := p.createTable(metric); err != nil {
				return err
			}
		}
		var cacheKey string
		var columns []string
		var tag_values string
		var values []interface{}
		if p.EnableCompactSchema {
			cacheKey, columns, tag_values, values = p.processMetricCompact(metric)
		} else {
			cacheKey, columns, values = p.processMetric(metric)
		}

		sql, found := p.queryCache[cacheKey]
		if !found {
			sql = p.generateInsert(tablename, columns)
			if p.EnableCompactSchema && p.Mode != MultiValuesPerSQL {
				// repalce '%s' by value_tags which like 'MAP{'host':'node', 'tags':'host1'}'
				sql = fmt.Sprintf(sql, tag_values)
			}
			p.Log.Debugf("Generated prepared sql: %s", sql)
			p.queryCache[cacheKey] = sql
		} else {
			if i == 0 {
				p.Log.Debugf("metric[%d], using cached sql: %s", i, sql)
			}
		}

		// Modifying the table schema is opt-in
		if p.TableUpdateTemplate != "" {
			for i := range len(columns) {
				if err := p.createColumn(tablename, columns[i], p.deriveDatatype(values[i])); err != nil {
					return err
				}
			}
		}
		// Using BatchTx is opt-in
		if p.Mode != Individual {
			batchedQueries[sql] = append(batchedQueries[sql], values)
		} else {
			if err := p.sendIndividual(sql, values); err != nil {
				return err
			}
		}
	}

	if p.Mode == MultiValuesPerSQL {
		for query, queryParams := range batchedQueries {
			if err := p.sendMultiValues(query, queryParams); err != nil {
				return fmt.Errorf("failed to send a multiple values tx: %w", err)
			}
		}
	} else if p.Mode == BatchTx {
		for query, queryParams := range batchedQueries {
			if err := p.sendBatch(query, queryParams); err != nil {
				return fmt.Errorf("failed to send a batched tx: %w", err)
			}
		}
	}

	return nil
}

func init() {
	// We provide default values here, which can be overridden by parameters in the .conf file
	outputs.Add("kaiwudb", func() telegraf.Output {
		return &Kaiwudb{
			Driver:              "kaiwudb-lite",
			EnableCompactSchema: true,
			TimestampColumnName: "ts",
			TimestampWithTZ:     false,
			Mode:                MultiValuesPerSQL,
			Convert:             defaultConvert,

			// Defaults for the connection settings (ConnectionMaxIdleTime,
			// ConnectionMaxLifetime, ConnectionMaxIdle, and ConnectionMaxOpen)
			// mirror the golang defaults. As of go 1.18 all of them default to 0
			// except max idle connections which is 2. See
			// https://pkg.go.dev/database/sql#DB.SetMaxIdleConns
			ConnectionMaxIdle: 2,
		}
	})
}

func iterSlice[E any](slice []E) iter.Seq[E] {
	return func(yield func(E) bool) {
		for _, element := range slice {
			if ok := yield(element); !ok {
				return
			}
		}
	}
}
