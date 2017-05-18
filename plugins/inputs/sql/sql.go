package sql

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"log"
	"regexp"
	//	"net/url"
	"os"
	"plugin"
	"reflect"
	"strconv"
	"strings"
	"time"
	// database drivers here:
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq" // pure go
	_ "github.com/mattn/go-sqlite3"
	//	_ "github.com/denisenkom/go-mssqldb" // pure go
	_ "github.com/zensqlmonitor/go-mssqldb" // pure go
	// the following commented because of the external proprietary libraries dependencies
	//	_ "github.com/mattn/go-oci8"
	//	_ "gopkg.in/rana/ora.v4"
	//	_ "bitbucket.org/phiggins/db2cli" //
	//	_ "github.com/SAP/go-hdb"
	//	_ "github.com/a-palchikov/sqlago"
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
	FieldHost        string
	FieldName        string
	FieldValue       string
	FieldDatabase    string
	FieldMeasurement string
	//
	NullAsZero        bool
	IgnoreOtherFields bool
	Sanitize          bool
	IgnoreRowErrors   bool
	//
	QueryScript string
	Parameters  []string //TODO

	// -------- internal data -----------
	statements []*sql.Stmt

	column_name []string
	cell_refs   []interface{}
	cells       []interface{}

	field_count int
	field_idx   []int //Column indexes of fields
	field_type  []int //Column types of fields

	tag_count int
	tag_idx   []int //Column indexes of tags (strings)

	field_host_idx        int
	field_database_idx    int
	field_measurement_idx int
	field_name_idx        int
	field_value_idx       int
	field_value_type      int
	field_timestamp_idx   int

	index int
}

type Sql struct {
	Driver    string
	SharedLib string

	KeepConnection bool

	Servers []string

	Hosts   []string
	DbNames []string

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
	var sampleConfig = `
	[[inputs.sql]]
		# debug=false						# Enables very verbose output

		## Database Driver
		driver = "mysql" 					# required. Valid options: mssql (SQLServer), mysql (MySQL), postgres (Postgres), sqlite3 (SQLite), [oci8 ora.v4 (Oracle)]
		# shared_lib = "/home/luca/.gocode/lib/oci8_go.so"		# optional: path to the golang 1.8 plugin shared lib
		# keep_connection = false 			# true: keeps the connection with database instead to reconnect at each poll and uses prepared statements (false: reconnection at each poll, no prepared statements)

		## Server DSNs
		servers  = ["readuser:sEcReT@tcp(neteye.wp.lan:3307)/rue", "readuser:sEcReT@tcp(hostmysql.wp.lan:3307)/monitoring"] # required. Connection DSN to pass to the DB driver
		#hosts=["neteye", "hostmysql"]	# optional: for each server a relative host entry should be specified and will be added as host tag
		#db_names=["rue", "monitoring"]	# optional: for each server a relative db name entry should be specified and will be added as dbname tag

		## Queries to perform (block below can be repeated)
		[[inputs.sql.query]]
			# query has precedence on query_script, if both query and query_script are defined only query is executed
			query="SELECT avg_application_latency,avg_bytes,act_throughput FROM Baselines WHERE application>0"
			# query_script = "/path/to/sql/script.sql" # if query is empty and a valid file is provided, the query will be read from file
			#
			measurement="connection_errors"	# destination measurement
			tag_cols=["application"]		# colums used as tags
			field_cols=["avg_application_latency","avg_bytes","act_throughput"]	# select fields and use the database driver automatic datatype conversion
			#
			# bool_fields=["ON"]				# adds fields and forces his value as bool
			# int_fields=["MEMBERS",".*BYTES"]	# adds fields and forces his value as integer
			# float_fields=["TEMPERATURE"]	# adds fields and forces his value as float
			# time_fields=[".*_TIME"]		# adds fields and forces his value as time
			#
			# field_measurement = "CLASS"		# the column that contains the name of the measurement
			# field_host = "DBHOST"				# the column that contains the name of the database host used for host tag value
			# field_database = "DBHOST"			# the column that contains the name of the database used for dbname tag value
			# field_name = "counter_name"		# the column that contains the name of the counter
			# field_value = "counter_value"		# the column that contains the value of the counter
			#
			# field_timestamp = "sample_time"	# the column where is to find the time of sample (should be a date datatype)
			#
			ignore_other_fields = false 	# false: if query returns columns not defined, they are automatically added (true: ignore columns)
			null_as_zero = false			# true: converts null values into zero or empty strings (false: ignore fields)
			sanitize = false				# true: will perform some chars substitutions (false: use value as is)
			ignore_row_errors				# true: if an error in row parse is raised then the row will be skipped and the parse continue on next row (false: fatal error)
	`
	return sampleConfig
}

func (_ *Sql) Description() string {
	return "SQL Plugin"
}

type DSN struct {
	host   string
	dbname string
}

//
//func ParseDSN(dsn string) (*DSN, error) {
//
//	url, err := url.Parse(dsn)
//	if err != nil {
//		return nil, err
//	}
//	pdsn := &DSN{}
//	pdsn.host = url.Host
//	pdsn.dbname = url.Path
//	return pdsn, err
//
//	res = map[string]string{}
//	parts := strings.Split(dsn, ";")
//	for _, part := range parts {
//		if len(part) == 0 {
//			continue
//		}
//		lst := strings.SplitN(part, "=", 2)
//		name := strings.TrimSpace(strings.ToLower(lst[0]))
//		if len(name) == 0 {
//			continue
//		}
//		var value string = ""
//		if len(lst) > 1 {
//			value = strings.TrimSpace(lst[1])
//		}
//		res[name] = value
//	}
//	return res
//	//	prm := &p.SessionPrm{Host: url.Host}
//	//
//	//	if url.User != nil {
//	//		pdsn.Username = url.User.Username()
//	//		prm.Password, _ = url.User.Password()
//	//	}
//}

func (s *Sql) Init() error {
	Debug = s.Debug

	if Debug {
		log.Printf("I! Init %d servers %d queries, driver %s", len(s.Servers), len(s.Query), s.Driver)
	}

	if len(s.SharedLib) > 0 {
		_, err := plugin.Open(s.SharedLib)
		if err != nil {
			panic(err)
		}
		if Debug {
			log.Printf("I! Loaded shared lib '%s'", s.SharedLib)
		}
	}

	if s.KeepConnection {
		s.connections = make([]*sql.DB, len(s.Servers))
		for i := 0; i < len(s.Query); i++ {
			s.Query[i].statements = make([]*sql.Stmt, len(s.Servers))
		}
	}
	//	for i := 0; i < len(s.Servers); i++ {
	//		c, err := ParseDSN(s.Servers[i])
	//		if err == nil {
	//			log.Printf("Host %s Database %s", c.host, c.dbname)
	//		} else {
	//			panic(err)
	//		}
	//
	//		//TODO get host from server
	//		// mysql servers  = ["nprobe:nprobe@tcp(neteye.wp.lan:3307)/nprobe"]
	//		// "postgres://nprobe:nprobe@rue-test/nprobe?sslmode=disable"
	//		// oracle telegraf/monitor@10.62.6.1:1522/tunapit
	//		//		match, _ := regexp.MatchString(".*@([0-9.a-zA-Z]*)[:]?[0-9]*/.*", "peach")
	//		//    fmt.Println(match)
	//		//				addr, err := net.LookupHost("198.252.206.16")
	//
	//	}
	if len(s.Servers) > 0 && len(s.Hosts) > 0 && len(s.Hosts) != len(s.Servers) {
		return fmt.Errorf("For each server a host should be specified (%d/%d)", len(s.Hosts), len(s.Servers))
	}
	if len(s.Servers) > 0 && len(s.DbNames) > 0 && len(s.DbNames) != len(s.Servers) {
		return fmt.Errorf("For each server a db name should be specified (%d/%d)", len(s.DbNames), len(s.Servers))
	}
	return nil
}

func (s *Query) Init(cols []string) error {
	qindex++
	s.index = qindex

	if Debug {
		log.Printf("I! Init Query %d with %d columns", s.index, len(cols))
	}

	//Define index of tags and fields and keep it for reuse
	s.column_name = cols

	// init the arrays for store row data
	col_count := len(s.column_name)
	s.cells = make([]interface{}, col_count)
	s.cell_refs = make([]interface{}, col_count)

	// init the arrays for store field/tag infos
	//	expected_tag_count := len(s.TagCols)
	//	var expected_field_count int
	//	if !s.IgnoreOtherFields {
	//		expected_field_count = col_count // - expected_tag_count
	//	} else {
	//		expected_field_count = len(s.FieldCols) + len(s.BoolFields) + len(s.IntFields) + len(s.FloatFields) + len(s.TimeFields)
	//	}
	// because of regex, now we must assume the max cols
	expected_field_count := col_count
	expected_tag_count := col_count

	s.tag_idx = make([]int, expected_tag_count)
	s.field_idx = make([]int, expected_field_count)
	s.field_type = make([]int, expected_field_count)
	s.tag_count = 0
	s.field_count = 0

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
		return fmt.Errorf("Missing column %s for given field_measurement", s.FieldTimestamp)
	}
	if len(s.FieldName) > 0 && !match_str(s.FieldName, s.column_name) {
		return fmt.Errorf("Missing column %s for given field_measurement", s.FieldName)
	}
	if len(s.FieldValue) > 0 && !match_str(s.FieldValue, s.column_name) {
		return fmt.Errorf("Missing column %s for given field_measurement", s.FieldValue)
	}
	if (len(s.FieldValue) > 0 && len(s.FieldName) == 0) || (len(s.FieldName) > 0 && len(s.FieldValue) == 0) {
		return fmt.Errorf("Both field_name and field_value should be set")
	}

	// fill columns info
	var cell interface{}
	for i := 0; i < col_count; i++ {
		dest_type := TYPE_AUTO
		field_matched := true

		if match_str(s.column_name[i], s.TagCols) {
			field_matched = false
			s.tag_idx[s.tag_count] = i
			s.tag_count++
			//			cell = new(sql.RawBytes)
			cell = new(string)
		} else if match_str(s.column_name[i], s.IntFields) {
			dest_type = TYPE_INT
			cell = new(sql.RawBytes)
			//				cell = new(int);
		} else if match_str(s.column_name[i], s.FloatFields) {
			dest_type = TYPE_FLOAT
			//				cell = new(float64);
			cell = new(sql.RawBytes)
		} else if match_str(s.column_name[i], s.TimeFields) {
			dest_type = TYPE_TIME
			//			cell = new(string)
			cell = new(sql.RawBytes)
		} else if match_str(s.column_name[i], s.BoolFields) {
			dest_type = TYPE_BOOL
			//				cell = new(bool);
			cell = new(sql.RawBytes)
		} else if match_str(s.column_name[i], s.FieldCols) {
			dest_type = TYPE_AUTO
			cell = new(sql.RawBytes)
		} else if !s.IgnoreOtherFields {
			dest_type = TYPE_AUTO
			cell = new(sql.RawBytes)
			//				cell = new(string);
		} else {
			field_matched = false
			cell = new(sql.RawBytes)
			if Debug {
				log.Printf("I! Skipped field %s", s.column_name[i])
			}
		}

		if Debug && !field_matched {
			log.Printf("I! Column %d '%s' dest type  %d", i, s.column_name[i], dest_type)
		}

		if s.column_name[i] == s.FieldHost {
			s.field_host_idx = i
			field_matched = false
		}
		if s.column_name[i] == s.FieldDatabase {
			s.field_database_idx = i
			field_matched = false
		}
		if s.column_name[i] == s.FieldMeasurement {
			s.field_measurement_idx = i
			field_matched = false
		}
		if s.column_name[i] == s.FieldName {
			s.field_name_idx = i
			field_matched = false
		}
		if s.column_name[i] == s.FieldValue {
			s.field_value_idx = i
			s.field_value_type = dest_type
			field_matched = false
		}
		if s.column_name[i] == s.FieldTimestamp {
			s.field_timestamp_idx = i
			field_matched = false
		}

		if field_matched {
			s.field_type[s.field_count] = dest_type
			s.field_idx[s.field_count] = i
			s.field_count++
		}
		s.cells[i] = cell
		s.cell_refs[i] = &s.cells[i]
	}

	if Debug {
		log.Printf("I! Query structure with %d tags and %d fields on %d columns...", s.tag_count, s.field_count, col_count)
	}

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
				if Debug {
					log.Printf("W! converting '%s' type %s raw data '%s'", name, reflect.TypeOf(cell).Kind(), fmt.Sprintf("%v", cell))
				}
			} else {
				value = strconv.FormatInt(ivalue, 10)
			}
		} else {
			value = string(barr)
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
			if !ok {
				var intvalue int64
				intvalue, ok = value.(int64)
				// TODO convert to s/ms/us/ns??
				if ok {
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
		case TYPE_INT:
			value = 0i
		case TYPE_FLOAT:
			value = 0.0
		case TYPE_BOOL:
			value = false
		case TYPE_TIME:
			value = time.Unix(0, 0)
			break
		default:
			value = 0
		}

		if Debug {
			log.Printf("I! forcing nil value of field '%s' type %d to %s", name, field_type, fmt.Sprintf("%v", value))
		}
	} else {
		value = nil
		if Debug {
			//			log.Printf("I! nil value for field name '%s' type %d", name, field_type)
		}
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
		if s.NullAsZero {
			return "", nil
		} else {
			return "", fmt.Errorf("Error converting name '%s' is nil", s.column_name[index])
		}
	}

	value, ok := ConvertString(s.column_name[index], cell)
	if !ok {
		return "", fmt.Errorf("Error converting name '%s' type %s, raw data '%s'", s.column_name[index], reflect.TypeOf(cell).Kind(), fmt.Sprintf("%v", cell))
	}

	if s.Sanitize {
		value = sanitize(value)
	}
	return value, nil
}

func (s *Query) ParseRow(timestamp time.Time, measurement string, tags map[string]string, fields map[string]interface{}) (time.Time, string, error) {
	if s.field_timestamp_idx >= 0 {
		// get the value of timestamp field
		value, err := s.ConvertField(s.column_name[s.field_timestamp_idx], s.cells[s.field_timestamp_idx], TYPE_TIME, false)
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
	// get host from row
	if s.field_host_idx >= 0 {
		host, err := s.GetStringFieldValue(s.field_host_idx)
		if err != nil {
			log.Printf("E! converting field host '%s'", s.column_name[s.field_host_idx])
			//cannot put data in correct, skip line
			return timestamp, measurement, err
		} else {
			tags["host"] = host
		}
	}
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
			//			log.Printf("******! tag %s=%s %s %s", name, value, , reflect.TypeOf(cell).Kind(), fmt.Sprintf("%v", cell))
			tags[name] = value
		}
	}
	// vertical counters
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
		value, err = s.ConvertField(s.column_name[s.field_value_idx], s.cells[s.field_value_idx], s.field_value_type, s.NullAsZero)
		if err != nil {
			// cannot get value of column with expected datatype, skip line
			return timestamp, measurement, err
		}

		// fill the field
		fields[name] = value
	}
	// horizontal counters
	// fill fields from column values
	for i := 0; i < s.field_count; i++ {
		name := s.column_name[s.field_idx[i]]
		// get the value of field
		value, err := s.ConvertField(name, s.cells[s.field_idx[i]], s.field_type[i], s.NullAsZero)
		if err != nil {
			// cannot get value of column with expected datatype, warning and continue
			log.Printf("W! converting value of field '%s'", name)
			//			return timestamp, measurement, err
		} else {
			fields[name] = value
		}
	}

	return timestamp, measurement, nil
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
				log.Printf("I! Performing query:\n\t\t%s\n...", q.Query)
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

	start_time := time.Now()

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
				log.Printf("I! Query %d exectution time: %s", q.index, time.Since(query_time))
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

				//	for debug purposes...
				//				if row_count == 0 && Debug {
				//					for ci := 0; ci < len(q.cells); ci++ {
				//						if q.cells[ci] != nil {
				//							log.Printf("I! Column '%s' type %s, raw data '%s'", q.column_name[ci], reflect.TypeOf(q.cells[ci]).Kind(), fmt.Sprintf("%v", q.cells[ci]))
				//						}
				//					}
				//				}

				// collect tags and fields
				tags := map[string]string{}
				fields := map[string]interface{}{}
				var measurement string

				// use database server as host, not the local host
				if len(p.Hosts) > 0 {
					_, ok := tags["host"]
					if !ok {
						tags["host"] = p.Hosts[si]
					}
				}
				// add dbname tag
				if len(p.DbNames) > 0 {
					_, ok := tags["dbname"]
					if !ok {
						tags["dbname"] = p.DbNames[si]
					}
				}

				timestamp, measurement, err = q.ParseRow(query_time, q.Measurement, tags, fields)
				if err != nil {
					if q.IgnoreRowErrors {
						log.Printf("W! Ignored error on row %d: %s", row_count, err)
					} else {
						return err
					}
				}

				//import "reflect"
				//// m1 and m2 are the maps we want to compare
				//eq := reflect.DeepEqual(m1, m2)

				acc.AddFields(measurement, fields, tags, timestamp)

				//		fieldsG := map[string]interface{}{
				//			"usage_user":       100 * (cts.User - lastCts.User - (cts.Guest - lastCts.Guest)) / totalDelta,
				//		}
				//		acc.AddGauge("cpu", fieldsG, tags, now) // TODO use gauge too?

				row_count += 1
			}
			if Debug {
				log.Printf("I! Query %d found %d rows written, processing duration %s", q.index, row_count, time.Since(query_time))
			}
		}
	}
	if Debug {
		log.Printf("I! Poll done, duration %s", time.Since(start_time))
	}

	return nil
}

func init() {
	inputs.Add("sql", func() telegraf.Input {
		return &Sql{}
	})
}
