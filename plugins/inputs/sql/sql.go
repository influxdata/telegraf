package sql

import (
	"context"
	dbsql "database/sql"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const magicIdleCount int = (-int(^uint(0) >> 1))

type Query struct {
	Query               string   `toml:"query"`
	Script              string   `toml:"query_script"`
	Measurement         string   `toml:"measurement"`
	MeasurementColumn   string   `toml:"measurement_column"`
	TimeColumn          string   `toml:"time_column"`
	TimeFormat          string   `toml:"time_format"`
	TagColumnsInclude   []string `toml:"tag_columns_include"`
	TagColumnsExclude   []string `toml:"tag_columns_exclude"`
	FieldColumnsInclude []string `toml:"field_columns_include"`
	FieldColumnsExclude []string `toml:"field_columns_exclude"`
	FieldColumnsFloat   []string `toml:"field_columns_float"`
	FieldColumnsInt     []string `toml:"field_columns_int"`
	FieldColumnsUint    []string `toml:"field_columns_uint"`
	FieldColumnsBool    []string `toml:"field_columns_bool"`
	FieldColumnsString  []string `toml:"field_columns_string"`

	statement         *dbsql.Stmt
	tagFilter         filter.Filter
	fieldFilter       filter.Filter
	fieldFilterFloat  filter.Filter
	fieldFilterInt    filter.Filter
	fieldFilterUint   filter.Filter
	fieldFilterBool   filter.Filter
	fieldFilterString filter.Filter
}

func (q *Query) parse(ctx context.Context, acc telegraf.Accumulator, rows *dbsql.Rows, t time.Time) (int, error) {
	columnNames, err := rows.Columns()
	if err != nil {
		return 0, err
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
		tags := make(map[string]string)
		fields := make(map[string]interface{}, len(columnNames))

		// Do the parsing with (hopefully) automatic type conversion
		if err := rows.Scan(columnDataPtr...); err != nil {
			return 0, err
		}

		for i, name := range columnNames {
			if q.MeasurementColumn != "" && name == q.MeasurementColumn {
				var ok bool
				if measurement, ok = columnData[i].(string); !ok {
					return 0, fmt.Errorf("measurement column type \"%T\" unsupported", columnData[i])
				}
			}

			if q.TimeColumn != "" && name == q.TimeColumn {
				var fieldvalue interface{}
				var skipParsing bool

				switch v := columnData[i].(type) {
				case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
					fieldvalue = v
				case []byte:
					fieldvalue = string(v)
				case time.Time:
					timestamp = v
					skipParsing = true
				case fmt.Stringer:
					fieldvalue = v.String()
				default:
					return 0, fmt.Errorf("time column %q of type \"%T\" unsupported", name, columnData[i])
				}
				if !skipParsing {
					if timestamp, err = internal.ParseTimestamp(q.TimeFormat, fieldvalue, ""); err != nil {
						return 0, fmt.Errorf("parsing time failed: %v", err)
					}
				}
			}

			if q.tagFilter.Match(name) {
				tagvalue, err := internal.ToString(columnData[i])
				if err != nil {
					return 0, fmt.Errorf("converting tag column %q failed: %v", name, err)
				}
				if v := strings.TrimSpace(tagvalue); v != "" {
					tags[name] = v
				}
			}

			// Explicit type conversions take precedence
			if q.fieldFilterFloat.Match(name) {
				v, err := internal.ToFloat64(columnData[i])
				if err != nil {
					return 0, fmt.Errorf("converting field column %q to float failed: %v", name, err)
				}
				fields[name] = v
				continue
			}

			if q.fieldFilterInt.Match(name) {
				v, err := internal.ToInt64(columnData[i])
				if err != nil {
					return 0, fmt.Errorf("converting field column %q to int failed: %v", name, err)
				}
				fields[name] = v
				continue
			}

			if q.fieldFilterUint.Match(name) {
				v, err := internal.ToUint64(columnData[i])
				if err != nil {
					return 0, fmt.Errorf("converting field column %q to uint failed: %v", name, err)
				}
				fields[name] = v
				continue
			}

			if q.fieldFilterBool.Match(name) {
				v, err := internal.ToBool(columnData[i])
				if err != nil {
					return 0, fmt.Errorf("converting field column %q to bool failed: %v", name, err)
				}
				fields[name] = v
				continue
			}

			if q.fieldFilterString.Match(name) {
				v, err := internal.ToString(columnData[i])
				if err != nil {
					return 0, fmt.Errorf("converting field column %q to string failed: %v", name, err)
				}
				fields[name] = v
				continue
			}

			// Try automatic conversion for all remaining fields
			if q.fieldFilter.Match(name) {
				var fieldvalue interface{}
				switch v := columnData[i].(type) {
				case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
					fieldvalue = v
				case []byte:
					fieldvalue = string(v)
				case time.Time:
					fieldvalue = v.UnixNano()
				case nil:
					fieldvalue = nil
				case fmt.Stringer:
					fieldvalue = v.String()
				default:
					return 0, fmt.Errorf("field column %q of type \"%T\" unsupported", name, columnData[i])
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
		return rowCount, err
	}

	return rowCount, nil
}

type SQL struct {
	Driver             string          `toml:"driver"`
	Dsn                string          `toml:"dsn"`
	Timeout            config.Duration `toml:"timeout"`
	MaxIdleTime        config.Duration `toml:"connection_max_idle_time"`
	MaxLifetime        config.Duration `toml:"connection_max_life_time"`
	MaxOpenConnections int             `toml:"connection_max_open"`
	MaxIdleConnections int             `toml:"connection_max_idle"`
	Queries            []Query         `toml:"query"`
	Log                telegraf.Logger `toml:"-"`

	driverName string
	db         *dbsql.DB
}

func (s *SQL) Init() error {
	// Option handling
	if s.Driver == "" {
		return errors.New("missing SQL driver option")
	}

	if s.Dsn == "" {
		return errors.New("missing data source name (DSN) option")
	}

	if s.Timeout <= 0 {
		s.Timeout = config.Duration(5 * time.Second)
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
			query, err := os.ReadFile(q.Script)
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

		// Compile the explicit type field-filter
		fieldfilterFloat, err := filter.NewIncludeExcludeFilterDefaults(q.FieldColumnsFloat, nil, false, false)
		if err != nil {
			return fmt.Errorf("creating field filter for float failed: %v", err)
		}
		s.Queries[i].fieldFilterFloat = fieldfilterFloat

		fieldfilterInt, err := filter.NewIncludeExcludeFilterDefaults(q.FieldColumnsInt, nil, false, false)
		if err != nil {
			return fmt.Errorf("creating field filter for int failed: %v", err)
		}
		s.Queries[i].fieldFilterInt = fieldfilterInt

		fieldfilterUint, err := filter.NewIncludeExcludeFilterDefaults(q.FieldColumnsUint, nil, false, false)
		if err != nil {
			return fmt.Errorf("creating field filter for uint failed: %v", err)
		}
		s.Queries[i].fieldFilterUint = fieldfilterUint

		fieldfilterBool, err := filter.NewIncludeExcludeFilterDefaults(q.FieldColumnsBool, nil, false, false)
		if err != nil {
			return fmt.Errorf("creating field filter for bool failed: %v", err)
		}
		s.Queries[i].fieldFilterBool = fieldfilterBool

		fieldfilterString, err := filter.NewIncludeExcludeFilterDefaults(q.FieldColumnsString, nil, false, false)
		if err != nil {
			return fmt.Errorf("creating field filter for string failed: %v", err)
		}
		s.Queries[i].fieldFilterString = fieldfilterString

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

	availDrivers := dbsql.Drivers()
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
	s.db, err = dbsql.Open(s.driverName, s.Dsn)
	if err != nil {
		return err
	}

	// Set the connection limits
	// s.db.SetConnMaxIdleTime(time.Duration(s.MaxIdleTime)) // Requires go >= 1.15
	s.db.SetConnMaxLifetime(time.Duration(s.MaxLifetime))
	s.db.SetMaxOpenConns(s.MaxOpenConnections)
	s.db.SetMaxIdleConns(s.MaxIdleConnections)

	// Test if the connection can be established
	s.Log.Debugf("Testing connectivity...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.Timeout))
	err = s.db.PingContext(ctx)
	cancel()
	if err != nil {
		return fmt.Errorf("connecting to database failed: %v", err)
	}

	// Prepare the statements
	for i, q := range s.Queries {
		s.Log.Debugf("Preparing statement %q...", q.Query)
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.Timeout))
		stmt, err := s.db.PrepareContext(ctx, q.Query) //nolint:sqlclosecheck // Closed in Stop()
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
			if err := q.statement.Close(); err != nil {
				s.Log.Errorf("closing statement for query %q failed: %v", q.Query, err)
			}
		}
	}

	// Close the connection to the server
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			s.Log.Errorf("closing database connection failed: %v", err)
		}
	}
}

func (s *SQL) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	tstart := time.Now()
	for _, query := range s.Queries {
		wg.Add(1)
		go func(q Query) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.Timeout))
			defer cancel()
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
			MaxIdleTime:        config.Duration(0), // unlimited
			MaxLifetime:        config.Duration(0), // unlimited
			MaxOpenConnections: 0,                  // unlimited
			MaxIdleConnections: magicIdleCount,     // will trigger auto calculation
		}
	})
}

func (s *SQL) executeQuery(ctx context.Context, acc telegraf.Accumulator, q Query, tquery time.Time) error {
	if q.statement == nil {
		return fmt.Errorf("statement is nil for query %q", q.Query)
	}

	// Execute the query
	rows, err := q.statement.QueryContext(ctx)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Handle the rows
	columnNames, err := rows.Columns()
	if err != nil {
		return err
	}
	rowCount, err := q.parse(ctx, acc, rows, tquery)
	s.Log.Debugf("Received %d rows and %d columns for query %q", rowCount, len(columnNames), q.Query)

	return err
}
