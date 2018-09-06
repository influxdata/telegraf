// The MIT License (MIT)
//
// Copyright (c) 2016 Luca Di Stefano (luca@distefano.bz.it)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package sql

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"log"
	"os"
	"plugin"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	// database drivers here:
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/stdlib"
	//	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3" // builds only on linux
	_ "github.com/zensqlmonitor/go-mssqldb"
)

const TYPE_STRING = 1
const TYPE_BOOL = 2
const TYPE_INT = 3
const TYPE_FLOAT = 4
const TYPE_TIME = 5
const TYPE_AUTO = 0

type Query struct {
	Query       string
	QueryScript string
	Measurement string
	//
	FieldTimestamp string
	TimestampUnit  string
	//
	TagCols []string

	// Vertical structure
	FieldHost        string
	FieldName        string
	FieldValue       string
	FieldDatabase    string
	FieldMeasurement string
	//
	Sanitize bool
	//

	// -------- internal data -----------
	statement *sql.Stmt

	column_name []string
	cell_refs   []interface{}
	cells       []interface{}

	// Horizontal structure
	field_count int
	field_idx   []int // Column indexes of fields
	field_type  []int // Column types of fields

	tag_count int
	tag_idx   []int // Column indexes of tags (strings)

	// Vertical structure
	field_host_idx        int
	field_database_idx    int
	field_measurement_idx int
	field_name_idx        int
	field_timestamp_idx   int

	// Data Conversion
	field_value_idx  int
	field_value_type int
}

type Sql struct {
	Driver    string
	SharedLib string

	//KeepConnection bool
	MaxLifetime time.Duration

	Source struct {
		Dsn string
	}

	Query []Query

	// internal
	connection_ts time.Time
	connection    *sql.DB
	initialized   bool
}

var sanitizedChars = strings.NewReplacer("/sec", "_persec", "/Sec", "_persec",
	" ", "_", "%", "Percent", `\`, "")

func trimSuffix(s, suffix string) string {
	for strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

func sanitize(text string) string {
	text = sanitizedChars.Replace(text)
	text = trimSuffix(text, "_")
	return text
}

func match_str(key string, str_array []string) bool {
	for _, pattern := range str_array {
		if pattern == key {
			return true
		}
		matched, _ := regexp.MatchString(pattern, key)
		if matched {
			return true
		}
	}
	return false
}

func (s *Sql) SampleConfig() string {
	return `
[[inputs.sql]]
  ## Database Driver, required. 
  ## Valid options: mssql (SQLServer), mysql (MySQL), postgres (Postgres), sqlite3 (SQLite), [oci8 ora.v4 (Oracle)]
  driver = "mysql"
  
  ## optional: path to the golang 1.8 plugin shared lib where additional sql drivers are linked
  # shared_lib = "/home/luca/.gocode/lib/oci8_go.so"
  
  ## optional: 
  ##    if true keeps the connection with database instead to reconnect at each poll and uses prepared statements
  ##    if false reconnection at each poll, no prepared statements 
  # keep_connection = false
  
  ## Maximum lifetime of a connection.
  max_lifetime = "0s"
  
  ## Connection information for data source.  Table can be repeated to define multiple sources.
  [[inputs.sql.source]]
    ## Data source name for connecting.  Syntax depends on selected driver.
    dsn = "readuser:sEcReT@tcp(neteye.wp.lan:3307)/rue"
    
  ## Queries to perform (block below can be repeated)
  [[inputs.sql.query]]
    ## query has precedence on query_script, if both query and query_script are defined only query is executed
    query="SELECT avg_application_latency,avg_bytes,act_throughput FROM Baselines WHERE application>0"
    # query_script = "/path/to/sql/script.sql" # if query is empty and a valid file is provided, the query will be read from file
    ## destination measurement
    measurement="connection_errors"
    
    ## Horizontal srtucture
    ## colums used as tags
    tag_cols=["application"]
    ## select fields and use the database driver automatic datatype conversion
    field_cols=["avg_application_latency","avg_bytes","act_throughput"]
    
    ## Vertical srtucture
    ## optional: the column that contains the name of the measurement, if not specified the value of the option measurement is used
    # field_measurement = "CLASS"
    ## the column that contains the name of the database host used for host tag value
    # field_host = "DBHOST"
    ## the column that contains the name of the database used for dbname tag value
    # field_database = "DBHOST"
    ## required if vertical: the column that contains the name of the counter
    # field_name = "counter_name"
    ## required if vertical: the column that contains the value of the counter
    # field_value = "counter_value"
    ## optional: the column where is to find the time of sample (should be a date datatype)
    # field_timestamp = "sample_time"  
`
}

func (_ *Sql) Description() string {
	return "SQL Plugin"
}

func (s *Sql) Init() error {
	log.Printf("D! Init %s servers %d queries, driver %s", s.Source.Dsn, len(s.Query), s.Driver)

	if len(s.SharedLib) > 0 {
		_, err := plugin.Open(s.SharedLib)
		if err != nil {
			panic(err)
		}
		log.Printf("D! Loaded shared lib '%s'", s.SharedLib)
	}

	return nil
}

func (s *Query) Init(cols []string) error {
	log.Printf("D! Init Query with %d columns", len(cols))

	// Define index of tags and fields and keep it for reuse
	s.column_name = cols

	// init the arrays for store row data
	col_count := len(s.column_name)
	s.cells = make([]interface{}, col_count)
	s.cell_refs = make([]interface{}, col_count)

	// because of regex, now we must assume the max cols
	expected_field_count := col_count
	expected_tag_count := col_count

	s.tag_idx = make([]int, expected_tag_count)
	s.field_idx = make([]int, expected_field_count)
	s.field_type = make([]int, expected_field_count)
	s.tag_count = 0
	s.field_count = 0

	// Vertical structure
	// prepare vars for vertical counter parsing
	s.field_name_idx = -1
	s.field_value_idx = -1
	s.field_timestamp_idx = -1
	s.field_measurement_idx = -1
	s.field_database_idx = -1
	s.field_host_idx = -1

	if len(s.FieldHost) > 0 && !match_str(s.FieldHost, s.column_name) {
		return fmt.Errorf("Missing column %s for given field_host", s.FieldHost)
	}
	if len(s.FieldDatabase) > 0 && !match_str(s.FieldDatabase, s.column_name) {
		return fmt.Errorf("Missing column %s for given field_database", s.FieldDatabase)
	}
	if len(s.FieldMeasurement) > 0 && !match_str(s.FieldMeasurement, s.column_name) {
		return fmt.Errorf("Missing column %s for given field_measurement", s.FieldMeasurement)
	}
	if len(s.FieldTimestamp) > 0 && !match_str(s.FieldTimestamp, s.column_name) {
		return fmt.Errorf("Missing column %s for given field_timestamp", s.FieldTimestamp)
	}
	if len(s.FieldName) > 0 && !match_str(s.FieldName, s.column_name) {
		return fmt.Errorf("Missing column %s for given field_name", s.FieldName)
	}
	if len(s.FieldValue) > 0 && !match_str(s.FieldValue, s.column_name) {
		return fmt.Errorf("Missing column %s for given field_value", s.FieldValue)
	}
	if (len(s.FieldValue) > 0 && len(s.FieldName) == 0) || (len(s.FieldName) > 0 && len(s.FieldValue) == 0) {
		return fmt.Errorf("Both field_name and field_value should be set")
	}
	//------------

	// fill columns info
	var cell interface{}
	for i := 0; i < col_count; i++ {
		dest_type := TYPE_AUTO
		field_matched := true // is horizontal field

		if match_str(s.column_name[i], s.TagCols) {
			field_matched = false
			s.tag_idx[s.tag_count] = i
			s.tag_count++
			cell = new(string)
		} else {
			dest_type = TYPE_AUTO
			cell = new(sql.RawBytes)
		}

		// Vertical structure
		if s.column_name[i] == s.FieldHost {
			s.field_host_idx = i
			field_matched = false
		} else if s.column_name[i] == s.FieldDatabase {
			s.field_database_idx = i
			field_matched = false
		} else if s.column_name[i] == s.FieldMeasurement {
			s.field_measurement_idx = i
			field_matched = false
		} else if s.column_name[i] == s.FieldName {
			s.field_name_idx = i
			field_matched = false
		} else if s.column_name[i] == s.FieldValue {
			s.field_value_idx = i
			s.field_value_type = dest_type
			field_matched = false
		} else if s.column_name[i] == s.FieldTimestamp {
			s.field_timestamp_idx = i
			field_matched = false
		}

		// Horizontal
		if field_matched {
			s.field_type[s.field_count] = dest_type
			s.field_idx[s.field_count] = i
			s.field_count++
		}

		//
		s.cells[i] = cell
		s.cell_refs[i] = &s.cells[i]
	}

	log.Printf("D! Query structure with %d tags and %d fields on %d columns...", s.tag_count, s.field_count, col_count)

	return nil
}

func ConvertString(name string, cell interface{}) (string, bool) {
	value, ok := cell.(string)
	if !ok {
		var barr []byte
		barr, ok = cell.([]byte)
		if !ok {
			var ivalue int64
			ivalue, ok = cell.(int64)
			if !ok {
				value = fmt.Sprintf("%v", cell)
				ok = true
				log.Printf("W! converting '%s' type %s raw data '%s'", name, reflect.TypeOf(cell).Kind(), fmt.Sprintf("%v", cell))
			} else {
				value = strconv.FormatInt(ivalue, 10)
			}
		} else {
			value = string(barr)
		}
	}
	return value, ok
}

func (s *Query) ConvertField(name string, cell interface{}, field_type int) (interface{}, error) {
	var value interface{}
	var ok bool
	var str string
	var err error

	ok = true
	if cell != nil {
		switch field_type {
		case TYPE_INT:
			str, ok = cell.(string)
			if ok {
				value, err = strconv.ParseInt(str, 10, 64)
			}
		case TYPE_FLOAT:
			str, ok = cell.(string)
			if ok {
				value, err = strconv.ParseFloat(str, 64)
			}
		case TYPE_BOOL:
			str, ok = cell.(string)
			if ok {
				value, err = strconv.ParseBool(str)
			}
		case TYPE_TIME:
			value, ok = cell.(time.Time)
			if !ok {
				var intvalue int64
				intvalue, ok = value.(int64)
				if ok {
					// TODO convert to s/ms/us/ns??
					value = time.Unix(intvalue, 0)
				}
			}
		case TYPE_STRING:
			value, ok = ConvertString(name, cell)
		default:
			value = cell
		}
	} else {
		value = nil
	}
	if !ok {
		err = fmt.Errorf("Error by converting field %s", name)
	}
	if err != nil {
		log.Printf("E! converting name '%s' type %s into type %d, raw data '%s'", name, reflect.TypeOf(cell).Kind(), field_type, fmt.Sprintf("%v", cell))
		return nil, err
	}
	return value, nil
}

func (s *Query) GetStringFieldValue(index int) (string, error) {
	cell := s.cells[index]
	if cell == nil {
		return "", fmt.Errorf("Error converting name '%s' is nil", s.column_name[index])
	}

	value, ok := ConvertString(s.column_name[index], cell)
	if !ok {
		return "", fmt.Errorf("Error converting name '%s' type %s, raw data '%s'", s.column_name[index], reflect.TypeOf(cell).Kind(), fmt.Sprintf("%v", cell))
	}

	//	if s.Sanitize {
	//		value = sanitize(value)
	//	}
	return value, nil
}

func (s *Query) ParseRow(timestamp time.Time, measurement string, tags map[string]string, fields map[string]interface{}) (time.Time, string, error) {
	// Vertical structure

	// get timestamp from row
	if s.field_timestamp_idx >= 0 {
		// get the value of timestamp field
		value, err := s.ConvertField(s.column_name[s.field_timestamp_idx], s.cells[s.field_timestamp_idx], TYPE_TIME)
		if err != nil {
			return timestamp, measurement, errors.New("Cannot convert timestamp")
		}
		timestamp, _ = value.(time.Time)
	}
	// get measurement from row
	if s.field_measurement_idx >= 0 {
		var err error
		measurement, err = s.GetStringFieldValue(s.field_measurement_idx)
		if err != nil {
			log.Printf("E! converting field measurement '%s'", s.column_name[s.field_measurement_idx])
			//cannot put data in correct measurement, skip line
			return timestamp, measurement, err
		}
	}
	// get dbname from row
	if s.field_database_idx >= 0 {
		dbname, err := s.GetStringFieldValue(s.field_database_idx)
		if err != nil {
			log.Printf("E! converting field dbname '%s'", s.column_name[s.field_database_idx])
			//cannot put data in correct, skip line
			return timestamp, measurement, err
		} else {
			tags["dbname"] = dbname
		}
	}
	// get server from row
	if s.field_host_idx >= 0 {
		server, err := s.GetStringFieldValue(s.field_host_idx)
		if err != nil {
			log.Printf("E! converting field host '%s'", s.column_name[s.field_host_idx])
			//cannot put data in correct, skip line
			return timestamp, measurement, err
		} else {
			tags["server"] = server
		}
	}
	// vertical counter
	if s.field_name_idx >= 0 {
		// get the name of the field
		name, err := s.GetStringFieldValue(s.field_name_idx)
		if err != nil {
			log.Printf("E! converting field name '%s'", s.column_name[s.field_name_idx])
			// cannot get name of field, skip line
			return timestamp, measurement, err
		}

		// get the value of field
		var value interface{}
		value, err = s.ConvertField(s.column_name[s.field_value_idx], s.cells[s.field_value_idx], s.field_value_type)
		if err != nil {
			// cannot get value of column with expected datatype, skip line
			return timestamp, measurement, err
		}

		// fill the field
		fields[name] = value
	}
	// ---------------

	// fill tags
	for i := 0; i < s.tag_count; i++ {
		index := s.tag_idx[i]
		name := s.column_name[index]
		value, err := s.GetStringFieldValue(index)
		if err != nil {
			log.Printf("E! ignored tag %s", name)
			// cannot put data in correct series, skip line
			return timestamp, measurement, err
		} else {
			tags[name] = value
		}
	}

	// horizontal counters
	// fill fields from column values
	for i := 0; i < s.field_count; i++ {
		name := s.column_name[s.field_idx[i]]
		// get the value of field
		value, err := s.ConvertField(name, s.cells[s.field_idx[i]], s.field_type[i])
		if err != nil {
			// cannot get value of column with expected datatype, warning and continue
			log.Printf("W! converting value of field '%s'", name)
		} else {
			fields[name] = value
		}
	}

	return timestamp, measurement, nil
}

func (p *Sql) Connect() (*sql.DB, error) {
	var err error

	// create connection to db server if not already done
	var db *sql.DB
	if p.MaxLifetime > 0 && time.Since(p.connection_ts) < p.MaxLifetime {
		db = p.connection
	} else {
		db = nil
	}

	if db == nil {
		log.Printf("D! Setting up DB %s %s ...", p.Driver, p.Source.Dsn)
		db, err = sql.Open(p.Driver, p.Source.Dsn)
		if err != nil {
			return nil, err
		}
		p.connection_ts = time.Now()
	} else {
		log.Printf("D! Reusing connection to %s ...", p.Source.Dsn)
	}

	log.Printf("D! Connecting to DB %s ...", p.Source.Dsn)
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	if p.MaxLifetime > 0 {
		p.connection = db
	}
	return db, nil
}

func (q *Query) Execute(db *sql.DB, KeepConnection bool) (*sql.Rows, error) {
	var err error
	var rows *sql.Rows
	// read query from sql script and put it in query string
	if len(q.QueryScript) > 0 && len(q.Query) == 0 {
		if _, err := os.Stat(q.QueryScript); os.IsNotExist(err) {
			log.Printf("E! SQL script file not exists '%s'...", q.QueryScript)
			return nil, err
		}
		filerc, err := os.Open(q.QueryScript)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		defer filerc.Close()

		buf := new(bytes.Buffer)
		buf.ReadFrom(filerc)
		q.Query = buf.String()
		log.Printf("D! Read %d bytes SQL script from '%s' for query ...", len(q.Query), q.QueryScript)
	}
	if len(q.Query) > 0 {
		if KeepConnection {
			// prepare statement if not already done
			if q.statement == nil {
				log.Printf("D! Preparing statement query ...")
				q.statement, err = db.Prepare(q.Query)
				if err != nil {
					return nil, err
				}
			}

			// execute prepared statement
			log.Printf("D! Performing query:\n\t\t%s\n...", q.Query)
			rows, err = q.statement.Query()
		} else {
			// execute query
			log.Printf("D! Performing query '%s'...", q.Query)
			rows, err = db.Query(q.Query)
		}
	} else {
		log.Printf("W! No query to execute")
		return nil, nil
	}

	return rows, err
}

func (p *Sql) Gather(acc telegraf.Accumulator) error {
	var err error

	start_time := time.Now()

	if !p.initialized {
		err = p.Init()
		if err != nil {
			return err
		}
		p.initialized = true
	}

	log.Printf("D! Starting poll")

	var db *sql.DB
	var query_time time.Time

	db, err = p.Connect()
	query_time = time.Now()
	duration := time.Since(query_time)
	log.Printf("D! Server %s connection time: %s", p.Source.Dsn, duration)

	if err != nil {
		return err
	}
	if p.MaxLifetime == 0 {
		defer db.Close()
	}

	// execute queries
	for qi := 0; qi < len(p.Query); qi++ {
		var rows *sql.Rows
		q := &p.Query[qi]

		query_time = time.Now()
		rows, err = q.Execute(db, p.MaxLifetime > 0)
		log.Printf("D! Query exectution time: %s", time.Since(query_time))

		query_time = time.Now()

		if err != nil {
			return err
		}
		if rows == nil {
			continue
		}
		defer rows.Close()

		if q.field_count == 0 {
			// initialize once the structure of query
			var cols []string
			cols, err = rows.Columns()
			if err != nil {
				return err
			}
			err = q.Init(cols)
			if err != nil {
				return err
			}
		}

		row_count := 0

		for rows.Next() {
			var timestamp time.Time

			if err = rows.Err(); err != nil {
				return err
			}
			// database driver datatype conversion
			err := rows.Scan(q.cell_refs...)
			if err != nil {
				return err
			}

			// collect tags and fields
			tags := map[string]string{}
			fields := map[string]interface{}{}
			var measurement string

			timestamp, measurement, err = q.ParseRow(query_time, q.Measurement, tags, fields)
			if err != nil {
				log.Printf("W! Ignored error on row %d: %s", row_count, err)
			} else {
				acc.AddFields(measurement, fields, tags, timestamp)
			}
			row_count += 1
		}
		log.Printf("D! Query found %d rows written, processing duration %s", row_count, time.Since(query_time))
	}

	log.Printf("D! Poll done, duration %s", time.Since(start_time))

	return nil
}

func init() {
	inputs.Add("sql", func() telegraf.Input {
		return &Sql{}
	})
}
