package sql

import (
	"bytes"
	"database/sql"
	"errors"
	"github.com/gchaincl/dotsql"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"log"
	"os"
	"reflect"
	"strconv"
	"time"
	// database drivers here:
	//	_ "bitbucket.org/phiggins/db2cli" //
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq" // pure go
	_ "github.com/mattn/go-oci8"
	_ "gopkg.in/rana/ora.v4"
	//	_ "github.com/denisenkom/go-mssqldb" // pure go
	_ "github.com/zensqlmonitor/go-mssqldb" // pure go
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

	TagCols   []string
	FieldCols []string
	//
	IntFields   []string
	FloatFields []string
	BoolFields  []string
	TimeFields  []string
	//
	FieldsName  []string
	FieldsValue []string
	//
	NullAsZero        bool
	IgnoreOtherFields bool
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

	index int
}

//type Database struct {
//	Hosts          []string
//	Driver         string
//	Servers        []string
//	KeepConnection bool
//
//	Query []Query
//
//	// internal
//	connections  []*sql.DB
//	_initialized bool
//}
//
//type Sql struct {
//	Instance []Database
//	// internal
//	Debug bool
//}
type Sql struct {
	Hosts []string

	Driver         string
	Servers        []string
	KeepConnection bool

	Query []Query

	// internal
	Debug bool

	connections  []*sql.DB
	_initialized bool
}

//TODO
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
		debug=false
	
		## DB Driver
		driver = "oci8" 					# required. Options: go-mssqldb (sqlserver) , oci8 ora.v4 (Oracle), mysql, pq (Postgres)
		# keep_connection = false 			# keeps the connection with database instead to reconnect at each poll
		
		## Server URLs
		servers  = ["telegraf/monitor@10.0.0.5:1521/thesid", "telegraf/monitor@orahost:1521/anothersid"] # required. Connection URL to pass to the DB driver
		hosts=["oraserver1", "oraserver2"]	# for each server a relative host entry should be specified and will be added as host tag
	
		## Queries to perform (block below can be repeated)
		[[inputs.sql.query]]
			query="select GROUP#,MEMBERS,STATUS,FIRST_TIME,FIRST_CHANGE#,BYTES,ARCHIVED from v$log"
			query_script = "/path/to/sql/script.sql" # if query is empty and a valid file is provided, the query will be read from file
			measurement="log"				# destination measurement
			tag_cols=["GROUP#","NAME"]		# colums used as tags
			field_cols=["UNIT"]				# select fields and use the database driver automatic datatype conversion
			#bool_fields=["ON"]				# adds fields and forces his value as bool
			#int_fields=["MEMBERS","BYTES"]	# adds fields and forces his value as integer
			#float_fields=["TEMPERATURE"]	# adds fields and forces his value as float
			#time_fields=["FIRST_TIME"]		# adds fields and forces his value as time
			ignore_other_fields = false 	# false: if query returns columns not defined, they are automatically added (true: ignore columns)
			null_as_zero = false			# true: Push null results as zeros/empty strings (false: ignore fields)
	`
	return sampleConfig
}

func (_ *Sql) Description() string {
	return "SQL Plugin"
}

func (s *Sql) Init() {
	Debug = s.Debug

	if Debug {
		log.Printf("I! Init %d servers %d queries", len(s.Servers), len(s.Query))
	}
	if s.KeepConnection {
		s.connections = make([]*sql.DB, len(s.Servers))
		//		for _, q := range s.Query {
		//			q.statements = make([]*sql.Stmt, len(s.Servers))
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

	var cell interface{}
	for i := 0; i < col_count; i++ {
		if Debug {
			log.Printf("I! Field %s %d", s.column_name[i], i)
		}
		field_matched := true
		if contains_str(s.column_name[i], s.TagCols) {
			field_matched = false
			s.tag_idx[s.tag_count] = i
			s.tag_count++
			cell = new(sql.RawBytes)
			//				cell = new(string);
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
			//					case TYPE_TIME:
			//						value = cell
			//						break
			//					case TYPE_STRING:
			//						value = cell
			//						break
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
		cell_type := reflect.TypeOf(cell).Kind()

		log.Printf("E! converting field name '%s' type %d %s into string", name, field_type, cell_type)
		err = errors.New("Error converting field into string")
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (s *Query) ParseRow(query_time time.Time, host string, acc telegraf.Accumulator) error {
	tags := map[string]string{}
	fields := map[string]interface{}{}

	//	if host != nil {
	//Use database server as host, not the local host
	tags["host"] = host
	//	}

	//Fill tags
	for i := 0; i < s.tag_count; i++ {
		cell := s.cells[s.tag_idx[i]]
		if cell != nil {
			//Tags are always strings
			name := s.column_name[s.tag_idx[i]]
			value, ok := cell.(string)
			if !ok {
				log.Printf("E! converting tag %d '%s' type %d", s.field_idx[i], name, s.field_type[i])
				return nil
			}
			tags[name] = value
		}
	}

	//Fill fields
	for i := 0; i < s.field_count; i++ {
		cell := s.cells[s.field_idx[i]]
		name := s.column_name[s.field_idx[i]]
		value, err := s.ConvertField(name, cell, s.field_type[i], s.NullAsZero)
		if err != nil {
			return err
		}
		fields[name] = value
	}

	acc.AddFields(s.Measurement, fields, tags, query_time)
	return nil
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
	} else if len(q.QueryScript) > 0 {
		// Loads queries from file
		var dot *dotsql.DotSql
		dot, err = dotsql.LoadFromFile(q.QueryScript)
		if err != nil {
			return nil, err
		}
		rows, err = dot.Query(db, "find-users-by-email")
	} else {
		log.Printf("E! No query to execute %d", q.index)
		//				err = errors.New("No query to execute")
		//				return err
		return nil, nil
	}

	return rows, err
}

func (p *Sql) Gather(acc telegraf.Accumulator) error {
	if !p._initialized {
		//	if len(p.connections) == 0 {
		p.Init()
		p._initialized = true
	}

	if Debug {
		log.Printf("I! Starting poll")
	}
	for si := 0; si < len(p.Servers); si++ {
		var err error
		var db *sql.DB
		db, err = p.Connect(si)
		if err != nil {
			return err
		}
		if !p.KeepConnection {
			defer db.Close()
		}

		// Execute queries
		for qi := 0; qi < len(p.Query); qi++ {
			var rows *sql.Rows
			var query_time time.Time
			q := &p.Query[qi]

			query_time = time.Now()
			rows, err = q.Execute(db, si, p.KeepConnection)

			//			// read query from sql script and put it in query string
			//			if len(q.QueryScript) > 0 && len(q.Query) == 0 {
			//				if _, err := os.Stat(q.QueryScript); os.IsNotExist(err) {
			//					log.Printf("E! SQL script not exists '%s'...", q.QueryScript)
			//					return err
			//				}
			//				filerc, err := os.Open(q.QueryScript)
			//				if err != nil {
			//					log.Fatal(err)
			//					return err
			//				}
			//				defer filerc.Close()
			//
			//				buf := new(bytes.Buffer)
			//				buf.ReadFrom(filerc)
			//				q.Query = buf.String()
			//				if Debug {
			//					log.Printf("I! Read %d bytes SQL script from '%s' for query %d ...", len(q.Query), q.QueryScript, q.index)
			//				}
			//			}
			//			if len(q.Query) > 0 {
			//				if p.KeepConnection {
			//					// prepare statement if not already done
			//					if q.statements[si] == nil {
			//						if Debug {
			//							log.Printf("I! Preparing statement query %d...", q.index)
			//						}
			//						q.statements[si], err = db.Prepare(q.Query)
			//						if err != nil {
			//							return err
			//						}
			//						//					defer stmt.Close()
			//					}
			//
			//					// execute prepared statement
			//					if Debug {
			//						log.Printf("I! Performing query '%s'...", q.Query)
			//					}
			//					query_time = time.Now()
			//					rows, err = q.statements[si].Query()
			//					//			err = stmt.QueryRow(1)
			//				} else {
			//					// execute query
			//					if Debug {
			//						log.Printf("I! Performing query '%s'...", q.Query)
			//					}
			//					query_time = time.Now()
			//					rows, err = db.Query(q.Query)
			//				}
			//			} else if len(q.QueryScript) > 0 {
			//				// Loads queries from file
			//				var dot *dotsql.DotSql
			//				dot, err = dotsql.LoadFromFile(q.QueryScript)
			//				if err != nil {
			//					return err
			//				}
			//				rows, err = dot.Query(db, "find-users-by-email")
			//			} else {
			//				log.Printf("E! No query to execute %d", q.index)
			//				//				err = errors.New("No query to execute")
			//				//				return err
			//				continue
			//			}

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
				q.Init(cols)
			}

			row_count := 0

			for rows.Next() {
				if err = rows.Err(); err != nil {
					return err
				}

				err := rows.Scan(q.cell_refs...)
				if err != nil {
					return err
				}
				err = q.ParseRow(query_time, p.Hosts[si], acc)
				if err != nil {
					return err
				}
				row_count += 1
			}
			//			if Debug {
			log.Printf("I! Query %d on %s found %d rows written in %s...", q.index, p.Hosts[si], row_count, q.Measurement)
			//			}
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
