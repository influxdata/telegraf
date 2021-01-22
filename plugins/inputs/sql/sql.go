package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
  ## Database Driver
	## See https://github.com/influxdata/telegraf/blob/master/docs/SQL_DRIVERS_INPUT.md for
  ## a list of supported drivers.
  driver = "mysql"

  ## Data source name for connecting
  ## The syntax and supported options depends on selected driver.
  dsn = "username:password@mysqlserver:3307/dbname?param=value"

	## Timeout for any operation
  # timeout = "5s"

	## Connection time limits
	## By default the maximum idle time and maximum lifetime of a connection is unlimited, i.e. the connections
	## will not be closed automatically. If you specify a positive time, the connections will be closed after
	## idleing or existing for at least that amount of time, respectively.
	# connection_max_idle_time = "0s"
	# connection_max_life_time = "0s"

	## Connection count limits
	## By default the number of open connections is not limited and the number of maximum idle connections
	## will be inferred from the number of queries specified. If you specify a positive number for any of the
	## two options, connections will be closed when reaching the specified limit. The number of idle connections
	## will be clipped to the maximum number of connections limit if any.
	# connection_max_open = 0
	# connection_max_idle = auto

  [[inputs.sql.query]]
    ## Query to perform on the server
    query="SELECT user,state,latency,score FROM Scoreboard WHERE application > 0"
    ## Alternatively to specifying the query directly you can select a file here containing the SQL query.
    ## Only one of 'query' and 'query_script' can be specified!
    # query_script = "/path/to/sql/script.sql"

    ## Name of the measurement
    ## In case both measurement and 'measurement_col' are given, the latter takes precedence.
    # measurement = "sql"

    ## Column name containing the name of the measurement
    ## If given, this will take precedence over the 'measurement' setting. In case a query result
		## does not contain the specified column, we fall-back to the 'measurement' setting.
    # measurement_col = ""

		## Column name containing the time of the measurement
    ## If ommited, the time of the query will be used.
    # time_col = ""

		## Format of the time contained in 'time_col'
		## The time must be 'unix', 'unix_ms', 'unix_us', 'unix_ns', or a golang time format.
  	## See https://golang.org/pkg/time/#Time.Format for details.
		# time_format = "unix"

    ## Column names containing tags
		## An empty include list will reject all columns and an empty exclude list will not exclude any column.
		## I.e. by default no columns will be returned as tag and the tags are empty.
    # tag_cols_include = []
		# tag_cols_exclude = []

    ## Column names containing fields
		## An empty include list is equivalent to '[*]' and all returned columns will be accepted. An empty
		## exclude list will not exclude any column. I.e. by default all columns will be returned as fields.
		## NOTE: We rely on the database driver to perform automatic datatype conversion.
    # field_cols_include = []
		# field_cols_exclude = []
`

const magicIdleCount int = (-int(^uint(0) >> 1))

type Query struct {
	Query               string   `toml:"query"`
	Script              string   `toml:"query_script"`
	Measurement         string   `toml:"measurement"`
	MeasurementColumn   string   `toml:"measurement_col"`
	TimeColumn          string   `toml:"time_col"`
	TimeFormat          string   `toml:"time_format"`
	TagColumnsInclude   []string `toml:"tag_cols_include"`
	TagColumnsExclude   []string `toml:"tag_cols_exclude"`
	FieldColumnsInclude []string `toml:"field_cols_include"`
	FieldColumnsExclude []string `toml:"field_cols_exclude"`

	statement   *sql.Stmt
	tagFilter   filter.Filter
	fieldFilter filter.Filter
}

func (q *Query) parse(ctx context.Context, acc telegraf.Accumulator, rows *sql.Rows, t time.Time) (error, int, int) {
	columnNames, err := rows.Columns()
	if err != nil {
		return err, 0, 0
	}

	// Prepare the list of datapoints according to the received row
	columnData := make([]interface{}, len(columnNames))
	columnDataPtr := make([]interface{}, len(columnNames))

	for i := range columnData {
		columnDataPtr[i] = &columnData[i]
	}

	rowCount := 0
	for rows.Next() {
		measurement := q.Measurement
		timestamp := t
		tags := make(map[string]string, 0)
		fields := make(map[string]interface{}, len(columnNames))

		// Do the parsing with (hopefully) automatic type conversion
		if err := rows.Scan(columnDataPtr...); err != nil {
			return err, len(columnNames), 0
		}

		for i, name := range columnNames {
			if q.MeasurementColumn != "" && name == q.MeasurementColumn {
				var ok bool
				if measurement, ok = columnData[i].(string); !ok {
					err := fmt.Errorf("measurement column type \"%T\" unsupported", columnData[i])
					return err, len(columnNames), 0
				}
			}

			if q.TimeColumn != "" && name == q.TimeColumn {
				var e error

				switch v := columnData[i].(type) {
				case time.Time:
					timestamp, e = v, nil
				case int:
					timestamp, e = internal.ParseTimestamp(q.TimeFormat, int64(v), "")
				case int32:
					timestamp, e = internal.ParseTimestamp(q.TimeFormat, int64(v), "")
				case int64:
					timestamp, e = internal.ParseTimestamp(q.TimeFormat, v, "")
				case float32:
					timestamp, e = internal.ParseTimestamp(q.TimeFormat, float64(v), "")
				case float64:
					timestamp, e = internal.ParseTimestamp(q.TimeFormat, v, "")
				case string:
					timestamp, e = internal.ParseTimestamp(q.TimeFormat, v, "")
				default:
					e = fmt.Errorf("column type \"%T\" unsupported", columnData[i])
				}
				if e != nil {
					err := fmt.Errorf("parsing time failed: %v", e)
					return err, len(columnNames), 0
				}
			}

			if q.tagFilter.Match(name) {
				var tagvalue string
				switch v := columnData[i].(type) {
				case string:
					tagvalue = v
				case []byte:
					tagvalue = string(v)
				case int:
					tagvalue = strconv.FormatInt(int64(v), 10)
				case int8:
					tagvalue = strconv.FormatInt(int64(v), 10)
				case int16:
					tagvalue = strconv.FormatInt(int64(v), 10)
				case int32:
					tagvalue = strconv.FormatInt(int64(v), 10)
				case int64:
					tagvalue = strconv.FormatInt(int64(v), 10)
				case uint:
					tagvalue = strconv.FormatUint(uint64(v), 10)
				case uint8:
					tagvalue = strconv.FormatUint(uint64(v), 10)
				case uint16:
					tagvalue = strconv.FormatUint(uint64(v), 10)
				case uint32:
					tagvalue = strconv.FormatUint(uint64(v), 10)
				case uint64:
					tagvalue = strconv.FormatUint(uint64(v), 10)
				case float32:
					tagvalue = strconv.FormatFloat(float64(v), 'f', -1, 32)
				case float64:
					tagvalue = strconv.FormatFloat(float64(v), 'f', -1, 64)
				case bool:
					tagvalue = strconv.FormatBool(v)
				case time.Time:
					tagvalue = v.String()
				case nil:
					tagvalue = ""
				default:
					err := fmt.Errorf("tag column %q of type \"%T\" unsupported", name, columnData[i])
					return err, len(columnNames), 0
				}
				if v := strings.TrimSpace(tagvalue); v != "" {
					tags[name] = v
				}
			}

			if q.fieldFilter.Match(name) {
				var fieldvalue interface{}
				switch v := columnData[i].(type) {
				case string:
					fieldvalue = v
				case []byte:
					fieldvalue = string(v)
				case int:
					fieldvalue = int64(v)
				case int8:
					fieldvalue = int64(v)
				case int16:
					fieldvalue = int64(v)
				case int32:
					fieldvalue = int64(v)
				case int64:
					fieldvalue = int64(v)
				case uint:
					fieldvalue = uint64(v)
				case uint8:
					fieldvalue = uint64(v)
				case uint16:
					fieldvalue = uint64(v)
				case uint32:
					fieldvalue = uint64(v)
				case uint64:
					fieldvalue = uint64(v)
				case float32:
					fieldvalue = float64(v)
				case float64:
					fieldvalue = v
				case bool:
					fieldvalue = v
				case time.Time:
					fieldvalue = v.UnixNano()
				case nil:
					fieldvalue = nil
				default:
					err := fmt.Errorf("field column %q of type \"%T\" unsupported", name, columnData[i])
					return err, len(columnNames), 0
				}
				if fieldvalue != nil {
					fields[name] = fieldvalue
				}
			}
		}
		acc.AddFields(measurement, fields, tags, timestamp)
		rowCount++
	}

	if err := rows.Err(); err != nil {
		return err, len(columnNames), rowCount
	}

	return nil, len(columnNames), rowCount
}

type SQL struct {
	Driver             string            `toml:"driver"`
	Dsn                string            `toml:"dsn"`
	Timeout            internal.Duration `toml:"timeout"`
	MaxIdleTime        internal.Duration `toml:"connection_max_idle_time"`
	MaxLifetime        internal.Duration `toml:"connection_max_life_time"`
	MaxOpenConnections int               `toml:"connection_max_open"`
	MaxIdleConnections int               `toml:"connection_max_idle"`
	Queries            []Query           `toml:"query"`
	Log                telegraf.Logger   `toml:"-"`

	driverName string
	db         *sql.DB
}

func (s *SQL) Description() string {
	return `Read metrics from SQL queries`
}

func (s *SQL) SampleConfig() string {
	return sampleConfig
}

func (s *SQL) Init() error {
	// Option handling
	if s.Driver == "" {
		return errors.New("missing SQL driver option")
	}

	if s.Dsn == "" {
		return errors.New("missing data source name (DSN) option")
	}

	if s.Timeout.Duration <= 0 {
		s.Timeout = internal.Duration{Duration: 5 * time.Second}
	}

	if s.MaxIdleConnections == magicIdleCount {
		// Determine the number by the number of queries + the golang default value
		s.MaxIdleConnections = len(s.Queries) + 2
	}

	for i, q := range s.Queries {
		if q.Query == "" && q.Script == "" {
			return errors.New("neither 'query' nor 'query_script' specified")
		}

		if q.Query != "" && q.Script != "" {
			return errors.New("only one of 'query' and 'query_script' can be specified")
		}

		// In case we got a script, we should read the query now.
		if q.Script != "" {
			query, err := ioutil.ReadFile(q.Script)
			if err != nil {
				return fmt.Errorf("reading script %q failed: %v", q.Script, err)
			}
			s.Queries[i].Query = string(query)
		}

		// Time format
		if q.TimeFormat == "" {
			s.Queries[i].TimeFormat = "unix"
		}

		// Compile the tag-filter
		tagfilter, err := filter.NewIncludeExcludeFilterDefaults(q.TagColumnsInclude, q.TagColumnsExclude, false, false)
		if err != nil {
			return fmt.Errorf("creating tag filter failed: %v", err)
		}
		s.Queries[i].tagFilter = tagfilter

		// Compile the field-filter
		fieldfilter, err := filter.NewIncludeExcludeFilter(q.FieldColumnsInclude, q.FieldColumnsExclude)
		if err != nil {
			return fmt.Errorf("creating field filter failed: %v", err)
		}
		s.Queries[i].fieldFilter = fieldfilter

		if q.Measurement == "" {
			s.Queries[i].Measurement = "sql"
		}
	}

	// Derive the sql-framework driver name from our config name. This abstracts the actual driver
	// from the database-type the user wants.
	aliases := map[string]string{
		"cockroach": "pgx",
		"tidb":      "mysql",
		"mssql":     "sqlserver",
		"maria":     "mysql",
		"postgres":  "pgx",
	}
	s.driverName = s.Driver
	if driver, ok := aliases[s.Driver]; ok {
		s.driverName = driver
	}

	availDrivers := sql.Drivers()
	if !choice.Contains(s.driverName, availDrivers) {
		for d, r := range aliases {
			if choice.Contains(r, availDrivers) {
				availDrivers = append(availDrivers, d)
			}
		}

		// Sort the list of drivers and make them unique
		sort.Strings(availDrivers)
		last := 0
		for _, d := range availDrivers {
			if d != availDrivers[last] {
				last++
				availDrivers[last] = d
			}
		}
		availDrivers = availDrivers[:last+1]

		return fmt.Errorf("driver %q not supported use one of %v", s.Driver, availDrivers)
	}

	return nil
}

func (s *SQL) Start(_ telegraf.Accumulator) error {
	var err error

	// Connect to the database server
	s.Log.Debugf("Connecting to %q...", s.Dsn)
	s.db, err = sql.Open(s.driverName, s.Dsn)
	if err != nil {
		return err
	}

	// Set the connection limits
	// s.db.SetConnMaxIdleTime(s.MaxIdleTime.Duration) // Requires go >= 1.15
	s.db.SetConnMaxLifetime(s.MaxLifetime.Duration)
	s.db.SetMaxOpenConns(s.MaxOpenConnections)
	s.db.SetMaxIdleConns(s.MaxIdleConnections)

	// Test if the connection can be established
	s.Log.Debugf("Testing connectivity...")
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout.Duration)
	err = s.db.PingContext(ctx)
	cancel()
	if err != nil {
		return fmt.Errorf("connecting to database failed: %v", err)
	}

	// Prepare the statements
	for i, q := range s.Queries {
		s.Log.Debugf("Preparing statement %q...", q.Query)
		ctx, cancel := context.WithTimeout(context.Background(), s.Timeout.Duration)
		stmt, err := s.db.PrepareContext(ctx, q.Query)
		cancel()
		if err != nil {
			return fmt.Errorf("preparing query %q failed: %v", q.Query, err)
		}
		s.Queries[i].statement = stmt
	}

	return nil
}

func (s *SQL) Stop() {
	// Free the statements
	for _, q := range s.Queries {
		if q.statement != nil {
			q.statement.Close()
		}
	}

	// Close the connection to the server
	if s.db != nil {
		s.db.Close()
	}
}

func (s *SQL) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout.Duration)
	defer cancel()

	tstart := time.Now()
	for _, query := range s.Queries {
		wg.Add(1)

		go func(q Query) {
			defer wg.Done()
			if err := s.executeQuery(ctx, acc, q, tstart); err != nil {
				acc.AddError(err)
			}
		}(query)
	}
	wg.Wait()
	s.Log.Debugf("Executed %d queries in %s", len(s.Queries), time.Since(tstart).String())

	return nil
}

func init() {
	inputs.Add("sql", func() telegraf.Input {
		return &SQL{
			MaxIdleTime:        internal.Duration{Duration: 0 * time.Second}, // unlimited
			MaxLifetime:        internal.Duration{Duration: 0 * time.Second}, // unlimited
			MaxOpenConnections: 0,                                            // unlimited
			MaxIdleConnections: magicIdleCount,                               // will trigger auto calculation
		}
	})
}

func (s *SQL) executeQuery(ctx context.Context, acc telegraf.Accumulator, q Query, tquery time.Time) error {
	if q.statement == nil {
		return fmt.Errorf("statement is nil for query %q", q.Query)
	}

	// Execute the query
	rows, err := q.statement.QueryContext(ctx)
	defer rows.Close()
	if err != nil {
		return err
	}

	// Handle the rows
	err, colCount, rowCount := q.parse(ctx, acc, rows, tquery)
	s.Log.Debugf("Received %d rows and %d columns for query %q", rowCount, colCount, q.Query)

	return err
}
