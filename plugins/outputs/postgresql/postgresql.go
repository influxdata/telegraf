package postgresql

import (
	"log"
	"sort"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

const (
	tagsJSONColumn   = "tags"
	fieldsJSONColumn = "fields"
)

type Postgresql struct {
	db                dbWrapper
	Address           string
	Schema            string
	TagsAsForeignkeys bool
	TagsAsJsonb       bool
	FieldsAsJsonb     bool
	TableTemplate     string
	TagTableSuffix    string
	tables            tableKeeper
}

func init() {
	outputs.Add("postgresql", func() telegraf.Output { return newPostgresql() })
}

func newPostgresql() *Postgresql {
	return &Postgresql{
		Schema:         "public",
		TableTemplate:  "CREATE TABLE IF NOT EXISTS {TABLE}({COLUMNS})",
		TagTableSuffix: "_tag",
	}
}

// Connect establishes a connection to the target database and prepares the cache
func (p *Postgresql) Connect() error {
	db, err := newDbWrapper(p.Address)
	if err != nil {
		return err
	}
	p.db = db
	p.tables = newTableKeeper(db)
	return nil
}

// Close closes the connection to the database
func (p *Postgresql) Close() error {
	return p.db.Close()
}

func (p *Postgresql) fullTableName(name string) string {
	return quoteIdent(p.Schema) + "." + quoteIdent(name)
}

var sampleConfig = `
  ## specify address via a url matching:
  ##   postgres://[pqgotest[:password]]@localhost[/dbname]\
  ##       ?sslmode=[disable|verify-ca|verify-full]
  ## or a simple string:
  ##   host=localhost user=pqotest password=... sslmode=... dbname=app_production
  ##
  ## All connection parameters are optional.
  ##
  ## Without the dbname parameter, the driver will default to a database
  ## with the same name as the user. This dbname is just for instantiating a
  ## connection with the server and doesn't restrict the databases we are trying
  ## to grab metrics for.
  ##
  address = "host=localhost user=postgres sslmode=verify-full"

  ## Store tags as foreign keys in the metrics table. Default is false.
  # tags_as_foreignkeys = false

  ## Template to use for generating tables
  ## Available Variables:
  ##   {TABLE} - tablename as identifier
  ##   {TABLELITERAL} - tablename as string literal
  ##   {COLUMNS} - column definitions
  ##   {KEY_COLUMNS} - comma-separated list of key columns (time + tags)

  ## Default template
  # table_template = "CREATE TABLE IF NOT EXISTS {TABLE}({COLUMNS})"
  ## Example for timescaledb
  # table_template = "CREATE TABLE IF NOT EXISTS {TABLE}({COLUMNS}); SELECT create_hypertable({TABLELITERAL},'time',chunk_time_interval := '1 week'::interval,if_not_exists := true);"

  ## Schema to create the tables into
  # schema = "public"

  ## Use jsonb datatype for tags
  # tags_as_jsonb = false

  ## Use jsonb datatype for fields
  # fields_as_jsonb = false

`

func (p *Postgresql) SampleConfig() string { return sampleConfig }
func (p *Postgresql) Description() string  { return "Send metrics to PostgreSQL" }

func (p *Postgresql) Write(metrics []telegraf.Metric) error {
	toInsert := make(map[string][]*colsAndValues)
	for _, metric := range metrics {
		tablename := metric.Name()

		// create table if needed
		if p.tables.exists(p.Schema, tablename) == false {
			createStmt := p.generateCreateTable(metric)
			_, err := p.db.Exec(createStmt)
			if err != nil {
				log.Printf("E! Creating table failed: statement: %v, error: %v", createStmt, err)
				return err
			}
			p.tables.add(tablename)
		}

		columns := []string{"time"}
		values := []interface{}{metric.Time()}
		tagColumns, tagValues, err := p.prepareTags(metric)
		if err != nil {
			return err
		}
		if tagColumns != nil {
			columns = append(columns, tagColumns...)
			values = append(values, tagValues...)
		}

		if p.FieldsAsJsonb {
			d, err := buildJsonb(metric.Fields())
			if err != nil {
				return err
			}

			columns = append(columns, fieldsJSONColumn)
			values = append(values, d)
		} else {
			var keys []string
			fields := metric.Fields()
			for column := range fields {
				keys = append(keys, column)
			}
			sort.Strings(keys)
			for _, column := range keys {
				columns = append(columns, column)
				values = append(values, fields[column])
			}
		}

		newValues := &colsAndValues{
			cols: columns,
			vals: values,
		}
		toInsert[tablename] = append(toInsert[tablename], newValues)
	}

	return p.insertBatches(toInsert)
}

func (p *Postgresql) prepareTags(metric telegraf.Metric) ([]string, []interface{}, error) {
	if len(metric.Tags()) == 0 {
		return nil, nil, nil
	}

	if p.TagsAsForeignkeys {
		// tags in separate table
		tagID, err := p.getTagID(metric)
		if err != nil {
			return nil, nil, err
		}
		return []string{tagIDColumn}, []interface{}{tagID}, nil
	}
	// tags in measurement table
	if p.TagsAsJsonb {
		d, err := buildJsonbTags(metric.Tags())
		if err != nil {
			return nil, nil, err
		}

		if d != nil {
			return []string{tagsJSONColumn}, []interface{}{d}, nil
		}
		return nil, nil, nil

	}

	var keys []string
	tags := metric.Tags()
	for column := range tags {
		keys = append(keys, column)
	}
	sort.Strings(keys)
	numColumns := len(keys)
	var columns = make([]string, numColumns)
	var values = make([]interface{}, numColumns)
	for i, column := range keys {
		columns[i] = column
		values[i] = tags[column]
	}
	return columns, values, nil
}

type colsAndValues struct {
	cols []string
	vals []interface{}
}

// insertBatches takes batches of data to be inserted. The batches are mapped
// by the target table, and each batch contains the columns and values for those
// columns that will generate the INSERT statement.
// On column mismatch an attempt is made to create the column and try to reinsert.
func (p *Postgresql) insertBatches(batches map[string][]*colsAndValues) error {
	for tableName, colsAndValues := range batches {
		for _, row := range colsAndValues {
			sql := p.generateInsert(tableName, row.cols)
			_, err := p.db.Exec(sql, row.vals...)
			if err == nil {
				continue
			}

			// check if insert error was caused by column mismatch
			if p.FieldsAsJsonb {
				return err
			}

			log.Printf("W! Possible column mismatch while inserting new metrics: %v", err)

			retry := false
			retry, err = p.addMissingColumns(tableName, row.cols, row.vals)
			if err != nil {
				log.Printf("E! Could not fix column mismatch: %v", err)
				return err
			}

			// We added some columns and insert might work now. Try again immediately to
			// avoid long lead time in getting metrics when there are several columns missing
			// from the original create statement and they get added in small drops.
			if retry {
				_, err = p.db.Exec(sql, row.vals...)
			}
			if err != nil {
				return err
			}
		}
	}

	return nil
}
