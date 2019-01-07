package sqlserver

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/influxdata/telegraf"
	opc "github.com/influxdata/telegraf/lib"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type MetricGroup struct {
	Metrics   []telegraf.Metric
	Timestamp time.Time
}

type SqlServer struct {
	db                *sql.DB
	Address           string
	TagsAsForeignkeys bool
	TagsAsJsonb       bool
	FieldsAsJsonb     bool
	Timestamp         string
	TableTemplate     string
	TagTableSuffix    string
	Tables            map[string]bool
	QueueDataDir      string
	Connections       int
	MaxItems          uint64
	TagKey            []string
	Strategy          string
	Exclude           []string

	inputMutex  sync.RWMutex
	inputQueue  chan []telegraf.Metric
	insertTypes map[string]map[string]string
	exclude     []*regexp.Regexp
}

type InsertKey struct {
	timestamp time.Time
	tags      map[string]string
}

type InsertItem struct {
	TableName string
	Columns   []string
	Values    []interface{}
	Types     map[string]string
}

func (s *SqlServer) Connect() error {
	var err error
	var db *sql.DB

	s.insertTypes = make(map[string]map[string]string)
	s.inputQueue = make(chan []telegraf.Metric, s.MaxItems*uint64(s.Connections))
	maxconn := int(s.MaxItems * uint64(s.Connections) * 3)

	db, err = sql.Open("sqlserver", s.Address)
	if err != nil {
		log.Printf("E! [Open Connection]: %v", err)
		return err
	}
	s.db = db
	s.db.SetMaxOpenConns(maxconn)
	s.Tables = make(map[string]bool)

	for i := 0; i < s.Connections; i++ {
		go s.WriteMetrics(i)
	}

	return nil
}

func (s *SqlServer) WriteDB_Batch(txn *sql.Tx, insertItems []InsertItem) error {
	for _, insert := range insertItems {
		sql := s.generateInsertWithValues(insert.TableName, insert.Columns, insert.Values)
		_, err := txn.Exec(sql)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SqlServer) HandleInserts_Batch(i int, insertItems []InsertItem) error {
	txn, err := s.db.Begin()
	if err != nil {
		log.Println("ERROR: [s.db.Begin]: ", err)
		return err
	}

	err = s.WriteDB_Batch(txn, insertItems)
	if err != nil {
		err2 := txn.Rollback()
		if err2 != nil {
			log.Println("ERROR: [txn.Rollback]: ", err)
		}

		exists, table, column := s.ColumnExists(err)
		if !exists {
			err = s.AddColumn(table, column)
			if err != nil {
				log.Println("ERROR [batch.AddColumn]: ", err)
			}
		}
	} else {
		err = txn.Commit()
		if err != nil {
			log.Println("ERROR: [txn.Commit]: ", err)
		}
	}

	return err
}

func (s *SqlServer) ColumnExists(err error) (bool, string, string) {
	if s.FieldsAsJsonb == false {
		missingColumnRegex := regexp.MustCompile("pq: column \"(.*?)\" of relation \"(.*?)\" does not exist.*$")
		//dpInvalidInput := regexp.MustCompile("ERROR: invalid input syntax for type (.*?): \"(.*?)\".*$")
		matches := missingColumnRegex.FindStringSubmatch(err.Error())
		if matches != nil && len(matches) > 2 {
			table := matches[2]
			column := matches[1]

			return false, table, column
		}

		return true, "", ""
	}

	return true, "", ""
}

func (s *SqlServer) AddColumn(table string, column string) error {
	query := "ALTER TABLE %s.%s ADD COLUMN %s %s;"
	dbquery := fmt.Sprintf(query, quoteIdent("public"), quoteIdent(table), quoteIdent(column), "double precision")
	log.Println(dbquery)
	_, err := s.db.Exec(dbquery)
	if err != nil {
		return err
	}

	log.Println("Added Column", column, "to table", table)
	return nil
}

func (s *SqlServer) Close() error {
	return s.db.Close()
}

func contains(haystack []string, needle string) bool {
	for _, key := range haystack {
		if found, _ := regexp.MatchString(needle, key); found {
			return true
		}
	}
	return false
}

func quoteIdent(name string) string {
	return name //pgx.Identifier{name}.Sanitize()
}

func quoteLiteral(name string) string {
	return "'" + strings.Replace(name, "'", "''", -1) + "'"
}

func deriveDatatype(value interface{}) string {
	var datatype string

	switch value.(type) {
	case int64:
		datatype = "int8"
	case float64:
		datatype = "float8"
	case string:
		datatype = "text"
	default:
		datatype = "text"
		log.Printf("E! Unknown datatype %v", value)
	}
	return datatype
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
  # table_template = "CREATE TABLE {TABLE}({COLUMNS})"
  ## Example for timescale
  # table_template = "CREATE TABLE {TABLE}({COLUMNS}); SELECT create_hypertable({TABLELITERAL},'time',chunk_time_interval := '1 week'::interval);"

  ## Use jsonb datatype for tags
  # tags_as_jsonb = true

  ## Use jsonb datatype for fields
  # fields_as_jsonb = true

`

func (s *SqlServer) SampleConfig() string { return sampleConfig }
func (s *SqlServer) Description() string  { return "Send metrics to PostgreSQL" }

func (s *SqlServer) generateCreateTable(metric telegraf.Metric) string {
	var columns []string
	var pk []string
	var sql []string

	pk = append(pk, quoteIdent("time"))
	columns = append(columns, "time timestamp")

	// handle tags if necessary
	if len(metric.Tags()) > 0 {
		if s.TagsAsForeignkeys {
			// tags in separate table
			var tag_columns []string
			var tag_columndefs []string
			columns = append(columns, "tag_id int")

			if s.TagsAsJsonb {
				tag_columns = append(tag_columns, "tags")
				tag_columndefs = append(tag_columndefs, "tags jsonb")
			} else {
				for column, _ := range metric.Tags() {
					tag_columns = append(tag_columns, quoteIdent(column))
					tag_columndefs = append(tag_columndefs, fmt.Sprintf("%s text", quoteIdent(column)))
				}
			}
			table := quoteIdent(metric.Name() + s.TagTableSuffix)
			sql = append(sql, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(tag_id serial primary key,%s,UNIQUE(%s))", table, strings.Join(tag_columndefs, ","), strings.Join(tag_columns, ",")))
		} else {
			// tags in measurement table
			if s.TagsAsJsonb {
				columns = append(columns, "tags jsonb")
			} else {
				for column, _ := range metric.Tags() {
					pk = append(pk, quoteIdent(column))
					columns = append(columns, fmt.Sprintf("%s text", quoteIdent(column)))
				}
			}
		}
	}

	if s.FieldsAsJsonb {
		columns = append(columns, "fields jsonb")
	} else {
		var datatype string
		for column, v := range metric.Fields() {
			datatype = deriveDatatype(v)
			columns = append(columns, fmt.Sprintf("%s %s", quoteIdent(column), datatype))
		}
	}

	query := strings.Replace(s.TableTemplate, "{TABLE}", quoteIdent(metric.Name()), -1)
	query = strings.Replace(query, "{TABLELITERAL}", quoteLiteral("\""+metric.Name()+"\""), -1)
	query = strings.Replace(query, "{COLUMNS}", strings.Join(columns, ","), -1)
	query = strings.Replace(query, "{KEY_COLUMNS}", strings.Join(pk, ","), -1)

	sql = append(sql, query)
	return strings.Join(sql, ";")
}

func (s *SqlServer) generateInsert(tablename string, columns []string) string {

	var placeholder, quoted []string
	for i, column := range columns {
		placeholder = append(placeholder, fmt.Sprintf("$%d", i+1))
		quoted = append(quoted, quoteIdent(column))
	}

	sql := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", quoteIdent(tablename), strings.Join(quoted, ","), strings.Join(placeholder, ","))
	return sql
}

func (s *SqlServer) generateInsertWithValues(tablename string, columns []string, values []interface{}) string {
	var qvals, quoted []string
	for i, column := range columns {
		qval := values[i]
		switch qval.(type) {
		case int64:
			if val, ok := qval.(int64); ok {
				sval := fmt.Sprintf("%d", val)
				qvals = append(qvals, sval)
			} else {
				fmt.Println("Could not convert: ", qval)
			}
			break
		case float64:
			if val, ok := qval.(float64); ok {
				sval := fmt.Sprintf("%f", val)
				qvals = append(qvals, sval)
			} else {
				fmt.Println("Could not convert: ", qval)
			}
			break
		case time.Time:
			sval := fmt.Sprintf("'%s'", qval.(time.Time).Format("2006-01-02 15:04:05"))
			qvals = append(qvals, sval)
			break
		case []uint8:
			var sval []string
			for _, v := range qval.([]uint8) {
				sval = append(sval, fmt.Sprintf("%d", v))
			}
			qvals = append(qvals, strings.Join(sval, ","))
		case string:
			qvals = append(qvals, quoteLiteral(qval.(string)))
			break
		default:
			break
		}
		quoted = append(quoted, quoteIdent(column))
	}

	sql := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", quoteIdent(tablename), strings.Join(quoted, ","), strings.Join(qvals, ","))
	return sql
}

func (s *SqlServer) tableExists(tableName string) bool {
	stmt := "SELECT tablename FROM pg_tables WHERE tablename = $1 AND schemaname NOT IN ('information_schema','pg_catalog');"
	result, err := s.db.Exec(stmt, tableName)
	if err != nil {
		log.Printf("E! Error checking for existence of metric table %s: %v", tableName, err)
		return false
	}
	if count, _ := result.RowsAffected(); count == 1 {
		s.inputMutex.Lock()
		s.Tables[tableName] = true
		s.inputMutex.Unlock()
		return true
	}
	return false
}

func (s *SqlServer) getInsertKey(timestamp time.Time, tags map[string]string) string {
	ret := fmt.Sprintf("%d", timestamp.UTC().UnixNano())
	tagArray := make([]string, len(s.TagKey))
	i := 0
	for _, key := range s.TagKey {
		kvStr := key + "=" + tags[key]
		tagArray[i] = kvStr
		i++
	}
	sort.Strings(tagArray)
	for _, val := range tagArray {
		ret += "," + val
	}
	return ret
}

func (s *SqlServer) WriteMetrics(id int) {
	for true {
		var i, maxItems uint64
		var metrics []telegraf.Metric
		var insertItems []InsertItem

		queueLength := uint64(len(s.inputQueue))
		maxItems = queueLength

		if maxItems > s.MaxItems {
			maxItems = s.MaxItems
		}

		for i = 0; i < maxItems; i++ {
			m := <-s.inputQueue
			metrics = append(metrics, m...)
		}

		// First transpose the metrics into something we understand. Then
		// Write them out.
		for _, metric := range metrics {
			tags := metric.Tags()

			if name, ok := tags["measurement"]; ok {
				for fieldKey, fieldVal := range metric.Fields() {
					fullname := fmt.Sprintf("%s.%s", name, fieldKey)
					if len(fullname) >= 50 {
						fullname = fullname[:50]
					}

					var quality int
					if val, ok := metric.Tags()["Quality"]; ok {
						quality = opc.ParseQualityString(val)
					}
					if val, ok := metric.Tags()["quality"]; ok {
						quality = opc.ParseQualityString(val)
					}

					insertItem := InsertItem{
						TableName: "[dbo].[db_TeckData]",
						Columns: []string{
							"txt_TagName",
							"int_Quality",
							"dt_Timestamp",
							"txt_Value",
						},
						Values: []interface{}{
							fullname,
							fmt.Sprintf("%d", quality),
							fmt.Sprintf("%v", metric.Time().Format("2006-01-02T15:04:05.000")),
							fmt.Sprintf("%v", fieldVal),
						},
						Types: map[string]string{
							"txt_TagName":  "string",
							"int_Quality":  "int64",
							"dt_Timestamp": "datetime",
							"txt_Value":    "string",
						},
					}

					if len(insertItem.Columns) != len(insertItem.Values) {
						fmt.Println("error", insertItem)
					}
					insertItems = append(insertItems, insertItem)
				}
			}
		}

		if len(insertItems) > 0 {
			err := s.HandleInserts_Batch(id, insertItems)
			if err != nil {
				log.Printf("E! Database error: %v", err)
			}
		}

		// Clear the memory
		insertItems = nil
		time.Sleep(100 * time.Millisecond)
	}
}

func (s *SqlServer) Write(metrics []telegraf.Metric) error {
	s.inputQueue <- metrics
	return nil
}

func init() {
	outputs.Add("sqlserver", func() telegraf.Output { return newSqlServer() })
}

func newSqlServer() *SqlServer {
	return &SqlServer{
		TableTemplate:  "CREATE TABLE {TABLE}({COLUMNS})",
		TagsAsJsonb:    true,
		TagTableSuffix: "_tag",
		FieldsAsJsonb:  true,
	}
}
