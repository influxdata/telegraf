//go:generate ../../../tools/readme_config_includer/generator
package postgresql_extensible

import (
	"context"
	_ "embed"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	// Required for SQL framework driver
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/postgresql"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// Columns never reported as a field or tag
var ignoredColumns = map[string]bool{"stats_reset": true}

type Postgresql struct {
	Databases                 []string        `deprecated:"1.22.4;use the sqlquery option to specify database to use"`
	Queries                   []query         `toml:"query"`
	PreparedStatements        bool            `toml:"prepared_statements"`
	Timeout                   config.Duration `toml:"timeout"`
	Role                      string          `toml:"role"`
	DatnameInclude            []string        `toml:"datname_include"`
	DatnameExclude            []string        `toml:"datname_exclude"`
	MetadataRefreshInterval   config.Duration `toml:"metadata_refresh_interval"`
	MaxConnectionsPerDatabase int             `toml:"max_connections_per_database"`
	MaxParallelDatabases      int             `toml:"max_parallel_databases"`
	KeepIdleConnections       bool            `toml:"keep_idle_connections"`
	StringColumnsAsTags       bool            `toml:"string_columns_as_tags"`
	NumericAsFloat            bool            `toml:"numeric_as_float"`
	Log                       telegraf.Logger `toml:"-"`
	postgresql.Config

	service *postgresql.Service

	includeRegex []*regexp.Regexp
	excludeRegex []*regexp.Regexp

	dbVersion       int
	inRecovery      bool
	datnames        []string
	metadataUpdated time.Time
}

type query struct {
	Sqlquery            string          `toml:"sqlquery"`
	Script              string          `toml:"script"`
	Version             int             `deprecated:"1.28.0;use minVersion to specify minimal DB version this query supports"`
	MinVersion          int             `toml:"min_version"`
	MaxVersion          int             `toml:"max_version"`
	Tagvalue            string          `toml:"tagvalue"`
	Measurement         string          `toml:"measurement"`
	Timestamp           string          `toml:"timestamp"`
	Withdbname          bool            `deprecated:"1.22.4;use the sqlquery option to specify database to use"`
	Timeout             config.Duration `toml:"timeout"`
	Role                string          `toml:"role"`
	StringColumnsAsTags *bool           `toml:"string_columns_as_tags"`
	NumericAsFloat      *bool           `toml:"numeric_as_float"`

	additionalTags map[string]bool
	stringTags     bool
	numericFloat   bool
}

// job is a single unit of work: one query executed in one database
type job struct {
	datname string
	query   *query
}

// collection holds what all databases of one Gather have in common. It is
// read by the goroutines processing the databases and never modified after
// it is built.
type collection struct {
	queries   []*query
	timestamp time.Time
	// concurrency is the number of queries executing at the same time in a
	// single database
	concurrency int
}

func (*Postgresql) SampleConfig() string {
	return sampleConfig
}

func (p *Postgresql) Init() error {
	if p.MaxConnectionsPerDatabase <= 0 {
		p.MaxConnectionsPerDatabase = 1
	}
	if p.MaxParallelDatabases <= 0 {
		p.MaxParallelDatabases = 1
	}
	if err := validateRole(p.Role); err != nil {
		return err
	}
	if p.Role == "" {
		p.Role = "any"
	}

	// Compile the database name filters
	for _, pattern := range p.DatnameInclude {
		re, err := regexp.Compile("^(?:" + pattern + ")$")
		if err != nil {
			return fmt.Errorf("compiling datname_include %q failed: %w", pattern, err)
		}
		p.includeRegex = append(p.includeRegex, re)
	}
	for _, pattern := range p.DatnameExclude {
		re, err := regexp.Compile("^(?:" + pattern + ")$")
		if err != nil {
			return fmt.Errorf("compiling datname_exclude %q failed: %w", pattern, err)
		}
		p.excludeRegex = append(p.excludeRegex, re)
	}

	// Set defaults for the queries
	for i, q := range p.Queries {
		if q.Sqlquery == "" {
			query, err := os.ReadFile(q.Script)
			if err != nil {
				return err
			}
			q.Sqlquery = string(query)
		}
		if q.MinVersion == 0 {
			q.MinVersion = q.Version
		}
		if q.Measurement == "" {
			q.Measurement = "postgresql"
		}

		// Complete a query ending in "WHERE datname" with the deprecated
		// list of databases
		if q.Withdbname {
			if len(p.Databases) != 0 {
				q.Sqlquery += fmt.Sprintf(` IN ('%s')`, strings.Join(p.Databases, "','"))
			} else {
				q.Sqlquery += " is not null"
			}
		}
		if err := validateRole(q.Role); err != nil {
			return fmt.Errorf("query %d: %w", i, err)
		}
		if q.Role == "" {
			q.Role = p.Role
		}

		q.stringTags = p.StringColumnsAsTags
		if q.StringColumnsAsTags != nil {
			q.stringTags = *q.StringColumnsAsTags
		}
		q.numericFloat = p.NumericAsFloat
		if q.NumericAsFloat != nil {
			q.numericFloat = *q.NumericAsFloat
		}

		q.additionalTags = make(map[string]bool)
		if q.Tagvalue != "" {
			for _, tag := range strings.Split(q.Tagvalue, ",") {
				q.additionalTags[tag] = true
			}
		}
		p.Queries[i] = q
	}
	p.Config.SimpleProtocol = !p.PreparedStatements

	// Create a service to access the PostgreSQL server
	service, err := p.Config.CreateService()
	if err != nil {
		return err
	}
	p.service = service

	return nil
}

func (p *Postgresql) Start(_ telegraf.Accumulator) error {
	if err := p.service.Start(); err != nil {
		return err
	}

	// The pool of the connection database serves a single database, so size it
	// by the per-database limit. This overrides the undocumented max_open and
	// max_idle options, max_connections_per_database is the documented way to
	// size the pools of this plugin.
	if p.MaxConnectionsPerDatabase > p.MaxOpen {
		p.Log.Debugf("Raising max_open from %d to max_connections_per_database %d", p.MaxOpen, p.MaxConnectionsPerDatabase)
		p.service.DB.SetMaxOpenConns(p.MaxConnectionsPerDatabase)
	}
	if p.MaxConnectionsPerDatabase > p.MaxIdle {
		p.Log.Debugf("Raising max_idle from %d to max_connections_per_database %d", p.MaxIdle, p.MaxConnectionsPerDatabase)
		p.service.DB.SetMaxIdleConns(p.MaxConnectionsPerDatabase)
	}

	return nil
}

func (p *Postgresql) Gather(acc telegraf.Accumulator) error {
	// A timeout of zero means no limit on the duration of a collection
	ctx, cancel := context.WithCancel(context.Background())
	if p.Timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(p.Timeout))
	}
	defer cancel()

	// Refresh the cached server version, role and database list if due
	if err := p.refreshMetadata(ctx); err != nil {
		acc.AddError(err)
	}
	if p.dbVersion == 0 {
		// The server information was never read successfully
		return nil
	}

	// Select the queries to run for this server version and role
	queries := make([]*query, 0, len(p.Queries))
	for i := range p.Queries {
		q := &p.Queries[i]
		if !matchesVersion(q, p.dbVersion) {
			continue
		}
		if !matchesRole(q.Role, p.inRecovery) {
			p.Log.Debugf("Skipping query %q as the server role does not match %q", q.Measurement, q.Role)
			continue
		}
		queries = append(queries, q)
	}
	if len(queries) == 0 {
		return nil
	}

	// set default timestamp to Now and use for all generated metrics during
	// the same Gather call
	timestamp := time.Now()

	// Unless the connections should be kept, close the idle connections of
	// the connection database after the collection
	if !p.KeepIdleConnections {
		defer p.closeIdleConnections()
	}

	// Determine the databases to run the queries in
	datnames := []string{p.service.ConnectionDatabase}
	if len(p.includeRegex) > 0 || len(p.excludeRegex) > 0 {
		datnames = p.datnames
	}
	if len(datnames) == 0 {
		return nil
	}

	// Process up to max_parallel_databases databases concurrently, each running
	// up to max_connections_per_database of its queries at the same time. The
	// semaphore is acquired before spawning the goroutine to avoid a large
	// number of waiting goroutines when there are many databases.
	coll := &collection{
		queries:     queries,
		timestamp:   timestamp,
		concurrency: p.MaxConnectionsPerDatabase,
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, p.MaxParallelDatabases)
	for _, datname := range datnames {
		semaphore <- struct{}{}
		wg.Add(1)
		go func(datname string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			if err := p.gatherInDatabase(ctx, acc, datname, coll); err != nil {
				acc.AddError(fmt.Errorf("database %q: %w", datname, err))
			}
		}(datname)
	}
	wg.Wait()

	return nil
}

// gatherInDatabase runs the queries in the given database. The connection
// database uses the existing pool, any other database gets a pool of its own
// which is closed when done.
func (p *Postgresql) gatherInDatabase(ctx context.Context, acc telegraf.Accumulator, datname string, coll *collection) error {
	service := p.service
	if datname != p.service.ConnectionDatabase {
		var err error
		service, err = p.Config.CreateServiceForDatabase(datname)
		if err != nil {
			return err
		}
		if err := service.Start(); err != nil {
			return err
		}
		defer service.Stop()
		service.DB.SetMaxOpenConns(coll.concurrency)
		service.DB.SetMaxIdleConns(coll.concurrency)
	}

	p.gatherDatabase(ctx, acc, service, datname, coll)
	return nil
}

// closeIdleConnections closes the idle connections of the connection
// database by temporarily dropping the idle limit of the pool
func (p *Postgresql) closeIdleConnections() {
	p.service.DB.SetMaxIdleConns(0)
	p.service.DB.SetMaxIdleConns(max(p.MaxIdle, p.MaxConnectionsPerDatabase))
}

// gatherDatabase runs the queries in one database with coll.concurrency
// queries executing at the same time
func (p *Postgresql) gatherDatabase(ctx context.Context, acc telegraf.Accumulator, service *postgresql.Service, datname string, coll *collection) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, coll.concurrency)
	for _, q := range coll.queries {
		semaphore <- struct{}{}
		wg.Add(1)
		go func(j job) {
			defer wg.Done()
			defer func() { <-semaphore }()

			if err := p.gatherMetricsFromQuery(ctx, acc, service, j, coll.timestamp); err != nil {
				acc.AddError(fmt.Errorf("database %q: %w", j.datname, err))
			}
		}(job{datname: datname, query: q})
	}
	wg.Wait()
}

func (p *Postgresql) Stop() {
	p.service.Stop()
}

// refreshMetadata updates the cached server version, role and list of
// databases if a refresh is due. Caching them keeps the collections between
// two refreshes from connecting to the connection database at all.
func (p *Postgresql) refreshMetadata(ctx context.Context) error {
	if !refreshDue(p.metadataUpdated, time.Now(), time.Duration(p.MetadataRefreshInterval)) {
		return nil
	}

	row := p.service.DB.QueryRowContext(ctx,
		`SELECT setting::integer / 100, pg_is_in_recovery() FROM pg_settings WHERE name = 'server_version_num'`)
	if err := row.Scan(&p.dbVersion, &p.inRecovery); err != nil {
		return fmt.Errorf("querying server version and role failed: %w", err)
	}

	if len(p.includeRegex) > 0 || len(p.excludeRegex) > 0 {
		if err := p.refreshDatnames(ctx); err != nil {
			return err
		}
	}

	p.metadataUpdated = time.Now()
	return nil
}

// refreshDatnames updates the list of databases matching the filters
func (p *Postgresql) refreshDatnames(ctx context.Context) error {
	rows, err := p.service.DB.QueryContext(ctx,
		`SELECT datname FROM pg_database WHERE datallowconn AND NOT datistemplate ORDER BY datname`)
	if err != nil {
		return fmt.Errorf("querying database list failed: %w", err)
	}
	defer rows.Close()

	datnames := make([]string, 0)
	for rows.Next() {
		var datname string
		if err := rows.Scan(&datname); err != nil {
			return fmt.Errorf("scanning database list failed: %w", err)
		}
		if p.matchesDatname(datname) {
			datnames = append(datnames, datname)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("reading database list failed: %w", err)
	}

	if len(datnames) == 0 {
		p.Log.Warn("No database matches the datname_include/datname_exclude filters")
	}
	p.datnames = datnames

	return nil
}

func (p *Postgresql) matchesDatname(datname string) bool {
	if len(p.includeRegex) > 0 {
		var included bool
		for _, re := range p.includeRegex {
			if re.MatchString(datname) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}
	for _, re := range p.excludeRegex {
		if re.MatchString(datname) {
			return false
		}
	}
	return true
}

func (p *Postgresql) gatherMetricsFromQuery(
	ctx context.Context,
	acc telegraf.Accumulator,
	service *postgresql.Service,
	j job,
	timestamp time.Time,
) error {
	if j.query.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(j.query.Timeout))
		defer cancel()
	}

	rows, err := service.DB.QueryContext(ctx, j.query.Sqlquery)
	if err != nil {
		return fmt.Errorf("query %q: %w", j.query.Measurement, err)
	}
	defer rows.Close()

	// grab the column information from the result
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return err
	}
	numeric := make(map[string]bool, len(columnTypes))
	for _, ct := range columnTypes {
		if ct.DatabaseTypeName() == "NUMERIC" {
			numeric[ct.Name()] = true
		}
	}

	// Prepare the row buffers once and reuse them for every row
	values := make([]interface{}, len(columns))
	pointers := make([]interface{}, len(columns))
	for i := range values {
		pointers[i] = &values[i]
	}

	// Logging every value is only done on demand: the arguments of a
	// disabled Debugf call are still boxed and allocated for every column
	// of every row
	debug := p.Log.Level() >= telegraf.Debug

	for rows.Next() {
		if err := rows.Scan(pointers...); err != nil {
			return err
		}
		p.accRow(acc, columns, values, numeric, j, timestamp, debug)
	}
	return rows.Err()
}

func (p *Postgresql) accRow(
	acc telegraf.Accumulator,
	columns []string,
	values []interface{},
	numeric map[string]bool,
	j job,
	timestamp time.Time,
	debug bool,
) {
	q := j.query

	// extract the database name from the row if available
	dbname := j.datname
	for i, col := range columns {
		if col != "datname" {
			continue
		}
		if datname, ok := values[i].(string); ok {
			dbname = datname
		}
		break
	}

	tags := make(map[string]string, len(q.additionalTags)+2)
	tags["server"] = p.service.SanitizedAddress
	tags["db"] = dbname

	fields := make(map[string]interface{}, len(columns))
	for i, col := range columns {
		value := values[i]
		if debug {
			p.Log.Debugf("Column: %s = %T: %v\n", col, value, value)
		}
		if ignoredColumns[col] || value == nil {
			continue
		}

		if col == q.Timestamp {
			if v, ok := value.(time.Time); ok {
				timestamp = v
			}
			continue
		}

		if q.additionalTags[col] {
			v, err := internal.ToString(value)
			if err != nil {
				p.Log.Debugf("Failed to add %q as additional tag: %v", col, err)
			} else {
				tags[col] = v
			}
			continue
		}

		if q.numericFloat && numeric[col] {
			v, err := internal.ToFloat64(value)
			if err != nil {
				p.Log.Debugf("Failed to convert numeric column %q: %v", col, err)
				continue
			}
			if math.IsNaN(v) || math.IsInf(v, 0) {
				continue
			}
			fields[col] = v
			continue
		}

		switch v := value.(type) {
		case []byte:
			if q.stringTags {
				tags[col] = string(v)
			} else {
				fields[col] = string(v)
			}
		case string:
			if q.stringTags {
				tags[col] = v
			} else {
				fields[col] = v
			}
		default:
			fields[col] = value
		}
	}
	acc.AddFields(q.Measurement, fields, tags, timestamp)
}

func refreshDue(updated, now time.Time, interval time.Duration) bool {
	if updated.IsZero() {
		return true
	}
	return !now.Before(updated.Truncate(interval).Add(interval))
}

func validateRole(role string) error {
	switch role {
	case "", "any", "primary", "replica":
		return nil
	default:
		return fmt.Errorf("invalid role %q, expected one of \"any\", \"primary\" or \"replica\"", role)
	}
}

func matchesVersion(q *query, dbVersion int) bool {
	return q.MinVersion <= dbVersion && (q.MaxVersion == 0 || q.MaxVersion > dbVersion)
}

func matchesRole(role string, inRecovery bool) bool {
	switch role {
	case "primary":
		return !inRecovery
	case "replica":
		return inRecovery
	default:
		return true
	}
}

func init() {
	inputs.Add("postgresql_extensible", func() telegraf.Input {
		return &Postgresql{
			Config: postgresql.Config{
				MaxIdle: 1,
				MaxOpen: 1,
			},
			PreparedStatements:  true,
			KeepIdleConnections: true,
		}
	})
}
