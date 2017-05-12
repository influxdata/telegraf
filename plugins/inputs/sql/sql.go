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
	"reflect"
	"strconv"
	"strings"
	"time"
	// database drivers here:
	_ "github.com/mattn/go-sqlite3"
	//	_ "github.com/a-palchikov/sqlago"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq" // pure go
	//	_ "github.com/denisenkom/go-mssqldb" // pure go
	_ "github.com/zensqlmonitor/go-mssqldb" // pure go
	// oracle commented because of the external proprietary libraries dependencies
	//	_ "github.com/mattn/go-oci8"
	//	_ "gopkg.in/rana/ora.v4"
	// the following commented because of the external proprietary libraries dependencies
	//	_ "bitbucket.org/phiggins/db2cli" //
	//	_ "github.com/SAP/go-hdb"
)

const TYPE_STRING = 1
const TYPE_BOOL = 2
const TYPE_INT = 3
const TYPE_FLOAT = 4
const TYPE_TIME = 5
const TYPE_AUTO = 0

var Debug = false
var qindex = 0

type Query struct {
	Query       string
	Measurement string
	//
	FieldTimestamp string
	TimestampUnit  string
	//
	TagCols   []string
	FieldCols []string
	//
	IntFields   []string
	FloatFields []string
	BoolFields  []string
	TimeFields  []string
	//
	FieldName  string
	FieldValue string
	//
	NullAsZero        bool
	IgnoreOtherFields bool
	Sanitize          bool
	//
	QueryScript string

	// internal data
	statements []*sql.Stmt
	//	Parameters []string

	column_name []string
	cell_refs   []interface{}
	cells       []interface{}

	field_count int
	field_idx   []int //Column indexes of fields
	field_type  []int //Column types of fields

	tag_count int
	tag_idx   []int //Column indexes of tags (strings)

	field_name_idx      int
	field_value_idx     int
	field_timestamp_idx int

	index int
}

type Sql struct {
	Hosts []string

	Driver         string
	Servers        []string
	KeepConnection bool

	Query []Query

	// internal
	Debug bool

	connections []*sql.DB
	initialized bool
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

func contains_str(key string, str_array []string) bool {
	for _, b := range str_array {
		if b == key {
			return true
		}
	}
	return false
}

func (s *Sql) SampleConfig() string {
	var sampleConfig = `
	[[inputs.sql]]
		# debug=false						# Enables very verbose output

		## Database Driver
		driver = "oci8" 					# required. Valid options: go-mssqldb (sqlserver) , oci8 ora.v4 (Oracle), mysql, pq (Postgres)
		# keep_connection = false 			# true: keeps the connection with database instead to reconnect at each poll and uses prepared statements (false: reconnection at each poll, no prepared statements)

		## Server DSNs
		servers  = ["telegraf/monitor@10.0.0.5:1521/thesid", "telegraf/monitor@orahost:1521/anothersid"] # required. Connection DSN to pass to the DB driver
		hosts=["oraserver1", "oraserver2"]	# for each server a relative host entry should be specified and will be added as host tag

		## Queries to perform (block below can be repeated)
		[[inputs.sql.query]]
			# query has precedence on query_script, if both query and query_script are defined only query is executed
			query="select GROUP#,MEMBERS,STATUS,FIRST_TIME,FIRST_CHANGE#,BYTES,ARCHIVED from v$log"
			# query_script = "/path/to/sql/script.sql" # if query is empty and a valid file is provided, the query will be read from file
			#
			measurement="log"				# destination measurement
			tag_cols=["GROUP#","NAME"]		# colums used as tags
			field_cols=["UNIT"]				# select fields and use the database driver automatic datatype conversion
			#
			# bool_fields=["ON"]				# adds fields and forces his value as bool
			# int_fields=["MEMBERS","BYTES"]	# adds fields and forces his value as integer
			# float_fields=["TEMPERATURE"]	# adds fields and forces his value as float
			# time_fields=["FIRST_TIME"]		# adds fields and forces his value as time
			#
			# field_name = "counter_name"		# the column that contains the name of the counter
			# field_value = "counter_value"		# the column that contains the value of the counter
			#
			# field_timestamp = "sample_time"	# the column where is to find the time of sample (should be a date datatype)

			ignore_other_fields = false 	# false: if query returns columns not defined, they are automatically added (true: ignore columns)
			null_as_zero = false			# true: converts null values into zero or empty strings (false: ignore fields)
			sanitize = false				# true: will perform some chars substitutions (false: use value as is)
	`
	return sampleConfig
}

func (_ *Sql) Description() string {
	return "SQL Plugin"
}

func (s *Sql) Init() error {
	Debug = s.Debug

	if Debug {
		log.Printf("I! Init %d servers %d queries, driver %s", len(s.Servers), len(s.Query), s.Driver)
	}
	if s.KeepConnection {
		s.connections = make([]*sql.DB, len(s.Servers))
		for i := 0; i < len(s.Query); i++ {
			s.Query[i].statements = make([]*sql.Stmt, len(s.Servers))
		}
	}
	for i := 0; i < len(s.Servers); i++ {
		//TODO get host from server
		//		match, _ := regexp.MatchString(".*@([0-9.a-zA-Z]*)[:]?[0-9]*/.*", "peach")
		//    fmt.Println(match)
		//				addr, err := net.LookupHost("198.252.206.16")

	}
	if len(s.Servers) > 0 && len(s.Servers) != len(s.Hosts) {
		return errors.New("For each server a host should be specified")

	}
	return nil
}

func (s *Query) Init(cols []string) error {
	qindex++
	s.index = qindex

	if Debug {
		log.Printf("I! Init Query %d with %d columns", s.index, len(cols))
	}
	s.column_name = cols
	//Define index of tags and fields and keep it for reuse
	col_count := len(s.column_name)

	expected_tag_count := len(s.TagCols)
	var expected_field_count int
	if !s.IgnoreOtherFields {
		expected_field_count = col_count // - expected_tag_count
	} else {
		expected_field_count = len(s.FieldCols) + len(s.BoolFields) + len(s.IntFields) + len(s.FloatFields) + len(s.TimeFields)
	}

	if Debug {
		log.Printf("I! Extpected %d tags and %d fields", expected_tag_count, expected_field_count)
	}

	s.tag_idx = make([]int, expected_tag_count)
	s.field_idx = make([]int, expected_field_count)
	s.field_type = make([]int, expected_field_count)
	s.tag_count = 0
	s.field_count = 0

	s.cells = make([]interface{}, col_count)
	s.cell_refs = make([]interface{}, col_count)

	s.field_name_idx = -1
	s.field_value_idx = -1
	s.field_timestamp_idx = -1

	if len(s.FieldTimestamp) > 0 && !contains_str(s.FieldTimestamp, s.column_name) {
		log.Printf("E! Missing given field_timestamp in columns: %s", s.FieldTimestamp)
		return errors.New("Missing given field_timestamp in columns")
	}
	if len(s.FieldName) > 0 && !contains_str(s.FieldName, s.column_name) {
		log.Printf("E! Missing given field_name in columns: %s", s.FieldName)
		return errors.New("Missing given field_name in columns")
	}
	if len(s.FieldValue) > 0 && !contains_str(s.FieldValue, s.column_name) {
		log.Printf("E! Missing given field_value in columns: %s", s.FieldValue)
		return errors.New("Missing given field_value in columns")
	}
	if (len(s.FieldValue) > 0 && len(s.FieldName) == 0) || (len(s.FieldName) > 0 && len(s.FieldValue) == 0) {
		return errors.New("Both field_name and field_value should be set")
	}

	var cell interface{}
	for i := 0; i < col_count; i++ {
		if Debug {
			log.Printf("I! Field %s %d", s.column_name[i], i)
		}
		field_matched := true

		if s.column_name[i] == s.FieldName {
			s.field_name_idx = i
		}
		if s.column_name[i] == s.FieldValue {
			s.field_value_idx = i
			//			TODO force datatype?
		}
		if s.column_name[i] == s.FieldTimestamp {
			s.field_timestamp_idx = i
			//			TODO force datatype?
		}

		if contains_str(s.column_name[i], s.TagCols) {
			field_matched = false
			s.tag_idx[s.tag_count] = i
			s.tag_count++
			//			cell = new(sql.RawBytes)
			cell = new(string)
		} else if contains_str(s.column_name[i], s.IntFields) {
			s.field_type[s.field_count] = TYPE_INT
			cell = new(sql.RawBytes)
			//				cell = new(int);
		} else if contains_str(s.column_name[i], s.FloatFields) {
			s.field_type[s.field_count] = TYPE_FLOAT
			//				cell = new(float64);
			cell = new(sql.RawBytes)
		} else if contains_str(s.column_name[i], s.TimeFields) {
			//TODO as number?
			s.field_type[s.field_count] = TYPE_TIME
			cell = new(string)
			//				cell = new(sql.RawBytes)
		} else if contains_str(s.column_name[i], s.BoolFields) {
			s.field_type[s.field_count] = TYPE_BOOL
			//				cell = new(bool);
			cell = new(sql.RawBytes)
		} else if contains_str(s.column_name[i], s.FieldCols) {
			s.field_type[s.field_count] = TYPE_AUTO
			cell = new(sql.RawBytes)
		} else if !s.IgnoreOtherFields {
			s.field_type[s.field_count] = TYPE_AUTO
			cell = new(sql.RawBytes)
			//				cell = new(string);
		} else {
			field_matched = false
			cell = new(sql.RawBytes)
			if Debug {
				log.Printf("I! Skipped field %s", s.column_name[i])
			}
		}
		if field_matched {
			s.field_idx[s.field_count] = i
			s.field_count++
		}
		s.cells[i] = cell
		s.cell_refs[i] = &s.cells[i]
	}

	if Debug {
		log.Printf("I! Query received %d tags and %d fields on %d columns...", s.tag_count, s.field_count, col_count)
	}

	return nil
}

func ConvertString(name string, cell interface{}) (string, bool) {
	value, ok := cell.(string)
	if !ok {
		var barr []byte
		barr, ok = cell.([]byte)
		value = string(barr)
		if !ok {
			value = fmt.Sprintf("%v", cell)
			ok = true
			if Debug {
				log.Printf("W! converting '%s' type %s raw data '%s'", name, reflect.TypeOf(cell).Kind(), fmt.Sprintf("%v", cell))
			}
		}
	}
	return value, ok
}

func (s *Query) ConvertField(name string, cell interface{}, field_type int, NullAsZero bool) (interface{}, error) {
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
			break
		case TYPE_FLOAT:
			str, ok = cell.(string)
			if ok {
				value, err = strconv.ParseFloat(str, 64)
			}
			break
		case TYPE_BOOL:
			str, ok = cell.(string)
			if ok {
				value, err = strconv.ParseBool(str)
			}
			break
		case TYPE_TIME:
			value, ok = cell.(time.Time)
			// TODO convert to ns/ms/s/us??
			if !ok {
				var intvalue int64
				intvalue, ok = value.(int64)
				if !ok {
					log.Printf("E! Unable to convert timestamp '%s' type %s", s.column_name[s.field_timestamp_idx], reflect.TypeOf(cell).Kind())
				} else {
					value = time.Unix(intvalue, 0)
				}
			}
			break
		case TYPE_STRING:
			value, ok = ConvertString(name, cell)
			break
		default:
			value = cell
		}
	} else if NullAsZero {
		switch field_type {
		case TYPE_AUTO:
		case TYPE_STRING:
			value = ""
			break
		default:
			value = 0
		}

		if Debug {
			log.Printf("I! forcing to 0 field name '%s' type %d", name, field_type)
		}
	} else {
		value = nil
		if Debug {
			log.Printf("I! nil value for field name '%s' type %d", name, field_type)
		}
	}
	if !ok {
		err = errors.New("Error converting field into string")
	}
	if err != nil {
		log.Printf("E! converting name '%s' type %s into type %d, raw data '%s'", name, reflect.TypeOf(cell).Kind(), field_type, fmt.Sprintf("%v", cell))
		return nil, err
	}
	return value, nil
}

func (s *Query) ParseRow(tags map[string]string, fields map[string]interface{}, timestamp time.Time) (time.Time, error) {
	if s.field_timestamp_idx >= 0 {
		// get the value of timestamp field
		cell := s.cells[s.field_timestamp_idx]
		value, err := s.ConvertField(s.column_name[s.field_timestamp_idx], cell, TYPE_TIME, false)
		if err != nil {
			return timestamp, errors.New("Cannot convert timestamp")
		}
		timestamp, _ = value.(time.Time)

		//		var ok bool
		//		value, err := s.ConvertField(s.column_name[s.field_timestamp_idx], cell, TYPE_AUTO, s.NullAsZero)
		//		if err != nil {
		//			return timestamp, err
		//		}
		//		timestamp, ok = value.(time.Time)
		//		// TODO convert to ns/ms/s/us??
		//		if !ok {
		//			log.Printf("W! Unable to convert timestamp '%s' to time.Time", s.column_name[s.field_timestamp_idx])
		//			var intvalue int64
		//			intvalue, ok = value.(int64)
		//			if !ok {
		//				cell_type := reflect.TypeOf(cell).Kind()
		//				log.Printf("E! Unable to convert timestamp '%s' type %s", s.column_name[s.field_timestamp_idx], cell_type)
		//				return timestamp, errors.New("Cannot convert timestamp")
		//			}
		//			timestamp = time.Unix(intvalue, 0)
		//		}
	}

	// fill tags
	for i := 0; i < s.tag_count; i++ {
		cell := s.cells[s.tag_idx[i]]
		if cell != nil {
			// tags should be always strings
			name := s.column_name[s.tag_idx[i]]

			//			//TODO set flag for force tag data conversion?
			//			value, ok := cell.(string)
			//			if !ok {
			//				barr, ok := cell.([]byte)
			//				value = string(barr)
			//				if !ok {
			//					value = fmt.Sprintf("%v", cell)
			//					log.Printf("W! converting tag %d '%s' type %s", s.tag_idx[i], name, reflect.TypeOf(cell).Kind())
			//					//				return nil	// skips the row
			//					//				return errors.New("Cannot convert tag")	// break the run
			//				}
			//			}

			value, ok := ConvertString(name, cell)
			if !ok {
				log.Printf("W! ignored tag %s", name)
				// ignoring tag is correct?
				//				return nil	// skips the row
				//				return errors.New("Cannot convert tag")	// break the run
			} else {
				if s.Sanitize {
					tags[name] = sanitize(value)
				} else {
					tags[name] = value
				}
			}
		}
	}

	if s.field_name_idx >= 0 {
		// get the name of the field from value on column
		cell := s.cells[s.field_name_idx]
		name, ok := ConvertString(s.column_name[s.field_name_idx], cell)
		//		name, ok := cell.(string)
		if !ok {
			log.Printf("W! converting field name '%s'", s.column_name[s.field_name_idx])
			return timestamp, nil
			//			return errors.New("Cannot convert tag")
		}

		if s.Sanitize {
			name = sanitize(name)
		}

		// get the value of field
		cell = s.cells[s.field_value_idx]
		value, err := s.ConvertField(s.column_name[s.field_value_idx], cell, TYPE_AUTO, s.NullAsZero) // TODO set forced field type
		if err != nil {
			return timestamp, err
		}
		fields[name] = value
	} else {
		// fill fields from column values
		for i := 0; i < s.field_count; i++ {
			cell := s.cells[s.field_idx[i]]
			name := s.column_name[s.field_idx[i]]
			value, err := s.ConvertField(name, cell, s.field_type[i], s.NullAsZero)
			if err != nil {
				return timestamp, err
			}
			fields[name] = value
		}
	}
	return timestamp, nil
}

func (p *Sql) Connect(si int) (*sql.DB, error) {
	var err error

	// create connection to db server if not already done
	var db *sql.DB
	if p.KeepConnection {
		db = p.connections[si]
	} else {
		db = nil
	}

	if db == nil {
		if Debug {
			log.Printf("I! Setting up DB %s %s ...", p.Driver, p.Servers[si])
		}
		db, err = sql.Open(p.Driver, p.Servers[si])
		if err != nil {
			return nil, err
		}
	} else {
		if Debug {
			log.Printf("I! Reusing connection to %s ...", p.Servers[si])
		}
	}

	if Debug {
		log.Printf("I! Connecting to DB %s ...", p.Servers[si])
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	if p.KeepConnection {
		p.connections[si] = db
	}
	return db, nil
}

func (q *Query) Execute(db *sql.DB, si int, KeepConnection bool) (*sql.Rows, error) {
	var err error
	var rows *sql.Rows
	// read query from sql script and put it in query string
	if len(q.QueryScript) > 0 && len(q.Query) == 0 {
		if _, err := os.Stat(q.QueryScript); os.IsNotExist(err) {
			log.Printf("E! SQL script not exists '%s'...", q.QueryScript)
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
		if Debug {
			log.Printf("I! Read %d bytes SQL script from '%s' for query %d ...", len(q.Query), q.QueryScript, q.index)
		}
	}
	if len(q.Query) > 0 {
		if KeepConnection {
			// prepare statement if not already done
			if q.statements[si] == nil {
				if Debug {
					log.Printf("I! Preparing statement query %d...", q.index)
				}
				q.statements[si], err = db.Prepare(q.Query)
				if err != nil {
					return nil, err
				}
				//defer stmt.Close()
			}

			// execute prepared statement
			if Debug {
				log.Printf("I! Performing query '%s'...", q.Query)
			}
			rows, err = q.statements[si].Query()
		} else {
			// execute query
			if Debug {
				log.Printf("I! Performing query '%s'...", q.Query)
			}
			rows, err = db.Query(q.Query)
		}
	} else {
		log.Printf("W! No query to execute %d", q.index)
		//				err = errors.New("No query to execute")
		//				return nil, err
		return nil, nil
	}

	return rows, err
}

func (p *Sql) Gather(acc telegraf.Accumulator) error {
	var err error

	if !p.initialized {
		err = p.Init()
		if err != nil {
			return err
		}
		p.initialized = true
	}

	if Debug {
		log.Printf("I! Starting poll")
	}
	for si := 0; si < len(p.Servers); si++ {
		var db *sql.DB
		var query_time time.Time

		db, err = p.Connect(si)
		query_time = time.Now()
		if Debug {
			duration := time.Since(query_time)
			log.Printf("I! Server %d connection time: %s", si, duration)
		}

		if err != nil {
			return err
		}
		if !p.KeepConnection {
			defer db.Close()
		}

		// execute queries
		for qi := 0; qi < len(p.Query); qi++ {
			var rows *sql.Rows
			q := &p.Query[qi]

			query_time = time.Now()
			rows, err = q.Execute(db, si, p.KeepConnection)
			if Debug {
				duration := time.Since(query_time)
				log.Printf("I! Query %d exectution time: %s", q.index, duration)
			}
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

				// use database server as host, not the local host
				tags["host"] = p.Hosts[si]

				timestamp, err = q.ParseRow(tags, fields, query_time)
				if err != nil {
					return err
				}

				acc.AddFields(q.Measurement, fields, tags, timestamp)

				//		fieldsG := map[string]interface{}{
				//			"usage_user":       100 * (cts.User - lastCts.User - (cts.Guest - lastCts.Guest)) / totalDelta,
				//		}
				//		acc.AddGauge("cpu", fieldsG, tags, now)

				row_count += 1
			}
			if Debug {
				duration := time.Since(query_time)
				log.Printf("I! Query %d on %s found %d rows written in %s... processing duration %s", q.index, p.Hosts[si], row_count, q.Measurement, duration)
			}
		}
	}
	if Debug {
		log.Printf("I! Poll done")
	}
	return nil
}

func init() {
	inputs.Add("sql", func() telegraf.Input {
		return &Sql{}
	})
}
