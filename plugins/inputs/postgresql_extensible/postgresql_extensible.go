//go:generate ../../../tools/readme_config_includer/generator
package postgresql_extensible

import (
	"context"
	_ "embed"
	"fmt"
	"math"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	// Required for SQL framework driver
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/common/postgresql"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// Columns never reported as a field or tag
var ignoredColumns = map[string]bool{"stats_reset": true}

// Values accepted by role
var roles = []string{"any", "primary", "replica"}

// Values accepted by numeric_type
var numericTypes = []string{"string", "float"}

// Column types that string_columns_as_tags turns into tags
var textTypes = map[string]bool{"TEXT": true, "VARCHAR": true, "BPCHAR": true, "CHAR": true, "NAME": true}

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
	NumericType               string          `toml:"numeric_type"`
	Log                       telegraf.Logger `toml:"-"`
	postgresql.Config

	service *postgresql.Service

	// Pools of the databases other than the one of the address, kept between
	// collections while keep_idle_connections is set. They are created and
	// read by the goroutines processing the databases, hence the lock.
	pools   map[string]*postgresql.Service
	poolsMu sync.Mutex

	includeRegex []*regexp.Regexp
	excludeRegex []*regexp.Regexp

	dbVersion       int
	inRecovery      bool
	currentDatabase string
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
	NumericType         string          `toml:"numeric_type"`

	additionalTags map[string]bool
	stringTags     bool
	numericFloat   bool
}

// job is a single unit of work: one query executed in one database
type job struct {
	datname string
	query   *query
}

// resultSet is what the rows of one query result have in common. The values
// are overwritten by every Scan, the rest is built once per result.
type resultSet struct {
	columns []string
	values  []any
	// numeric names the columns of the NUMERIC type
	numeric map[string]bool
	// text names the columns of the text types
	text map[string]bool
	// debug tells whether every column value is logged
	debug bool
}

// collection is what the databases of one Gather have in common.
type collection struct {
	queries                []*query
	timestamp              time.Time
	connectionsPerDatabase int
	// connDatname is the database the pool of the address is connected to
	connDatname string
	// debug tells whether the connections of this collection are counted
	// for the debug line at its end
	debug bool
	// closed sums the connections of the pools closed while the collection
	// runs. A pool records its own count just before it goes away, so the
	// debug line at the end can still include it.
	closed atomic.Int64
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
	if p.Role == "" {
		p.Role = "any"
	}
	if err := choice.Check(p.Role, roles); err != nil {
		return fmt.Errorf("config option 'role': %w", err)
	}
	if p.NumericType == "" {
		p.NumericType = "string"
	}
	if err := choice.Check(p.NumericType, numericTypes); err != nil {
		return fmt.Errorf("config option 'numeric_type': %w", err)
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
		if q.Role == "" {
			q.Role = p.Role
		}
		if err := choice.Check(q.Role, roles); err != nil {
			return fmt.Errorf("query %d: config option 'role': %w", i, err)
		}

		q.stringTags = p.StringColumnsAsTags
		if q.StringColumnsAsTags != nil {
			q.stringTags = *q.StringColumnsAsTags
		}
		if q.NumericType == "" {
			q.NumericType = p.NumericType
		}
		if err := choice.Check(q.NumericType, numericTypes); err != nil {
			return fmt.Errorf("query %d: config option 'numeric_type': %w", i, err)
		}
		q.numericFloat = q.NumericType == "float"

		q.additionalTags = make(map[string]bool)
		if q.Tagvalue != "" {
			for tag := range strings.SplitSeq(q.Tagvalue, ",") {
				q.additionalTags[tag] = true
			}
		}
		p.Queries[i] = q
	}
	p.Config.SimpleProtocol = !p.PreparedStatements

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

	if !p.KeepIdleConnections {
		defer p.closeIdleConnections()
	}

	if err := p.refreshMetadata(ctx); err != nil {
		acc.AddError(err)
	}
	if p.dbVersion == 0 {
		return nil
	}

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

	timestamp := time.Now()

	// Determine the databases to run the queries in
	datnames := []string{p.service.ConnectionDatabase}
	connDatname := p.service.ConnectionDatabase
	if len(p.includeRegex) > 0 || len(p.excludeRegex) > 0 {
		datnames = p.datnames
		// Without a dbname in the address the server connects to the database
		// named after the user, not to the ConnectionDatabase. The single
		// database mode is excluded here only to keep the behavior of old
		// configurations.
		connDatname = p.currentDatabase
	}
	if len(datnames) == 0 {
		return nil
	}

	p.releasePools(datnames)

	coll := &collection{
		queries:                queries,
		timestamp:              timestamp,
		connDatname:            connDatname,
		connectionsPerDatabase: p.MaxConnectionsPerDatabase,
		debug:                  p.Log.Level() >= telegraf.Debug,
	}

	// The semaphore is acquired before spawning the goroutine, so that a
	// long database list does not pile up goroutines waiting for their turn
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, p.MaxParallelDatabases)
	for _, datname := range datnames {
		semaphore <- struct{}{}
		wg.Go(func() {
			defer func() { <-semaphore }()

			if err := p.gatherInDatabase(ctx, acc, datname, coll); err != nil {
				acc.AddError(fmt.Errorf("database %q: %w", datname, err))
			}
		})
	}
	wg.Wait()

	// Report the connections of this collection before the deferred close of
	// the idle ones runs
	if coll.debug {
		used := coll.closed.Load() + p.openConnections()
		var kept int64
		if p.KeepIdleConnections {
			kept = used
		}
		p.Log.Debugf("Collection used %d connections, %d kept open", used, kept)
	}

	return nil
}

// openConnections counts the connections currently open in the pool of the
// connection database and in the pools kept for the other databases
func (p *Postgresql) openConnections() int64 {
	count := int64(p.service.DB.Stats().OpenConnections)

	p.poolsMu.Lock()
	defer p.poolsMu.Unlock()
	for _, service := range p.pools {
		count += int64(service.DB.Stats().OpenConnections)
	}

	return count
}

// releasePools closes the pools of the databases missing from datnames
// between collections
func (p *Postgresql) releasePools(datnames []string) {
	p.poolsMu.Lock()
	defer p.poolsMu.Unlock()

	queried := make(map[string]bool, len(datnames))
	for _, datname := range datnames {
		queried[datname] = true
	}

	for datname, service := range p.pools {
		if queried[datname] {
			continue
		}
		service.Stop()
		delete(p.pools, datname)
	}
}

// newPool creates a pool for the given database, holding as many
// connections as the database may use at a time
func (p *Postgresql) newPool(datname string, connections int) (*postgresql.Service, error) {
	service, err := p.Config.CreateServiceForDatabase(datname)
	if err != nil {
		return nil, err
	}
	if err := service.Start(); err != nil {
		return nil, err
	}
	service.DB.SetMaxOpenConns(connections)
	service.DB.SetMaxIdleConns(connections)

	return service, nil
}

// poolFor returns the pool of the given database, creating it if this is the
// first collection querying that database
func (p *Postgresql) poolFor(datname string, connections int) (*postgresql.Service, error) {
	p.poolsMu.Lock()
	defer p.poolsMu.Unlock()

	if service, found := p.pools[datname]; found {
		return service, nil
	}

	service, err := p.newPool(datname, connections)
	if err != nil {
		return nil, err
	}

	if p.pools == nil {
		p.pools = make(map[string]*postgresql.Service)
	}
	p.pools[datname] = service

	return service, nil
}

// gatherInDatabase runs the queries in the given database. The connection
// database uses the pool created at startup, any other database gets one of
// its own, kept between collections while keep_idle_connections is set.
func (p *Postgresql) gatherInDatabase(ctx context.Context, acc telegraf.Accumulator, datname string, coll *collection) error {
	if datname == coll.connDatname {
		p.gatherDatabase(ctx, acc, p.service, datname, coll)
		return nil
	}

	if p.KeepIdleConnections {
		service, err := p.poolFor(datname, coll.connectionsPerDatabase)
		if err != nil {
			return err
		}
		p.gatherDatabase(ctx, acc, service, datname, coll)
		return nil
	}

	service, err := p.newPool(datname, coll.connectionsPerDatabase)
	if err != nil {
		return err
	}

	p.gatherDatabase(ctx, acc, service, datname, coll)

	if coll.debug {
		coll.closed.Add(int64(service.DB.Stats().OpenConnections))
	}
	service.Stop()

	return nil
}

// closeIdleConnections closes the idle connections of the connection
// database by temporarily dropping the idle limit of the pool
func (p *Postgresql) closeIdleConnections() {
	p.service.DB.SetMaxIdleConns(0)
	p.service.DB.SetMaxIdleConns(max(p.MaxIdle, p.MaxConnectionsPerDatabase))
}

// gatherDatabase runs the queries in one database, as many at a time as
// that database may use connections
func (p *Postgresql) gatherDatabase(ctx context.Context, acc telegraf.Accumulator, service *postgresql.Service, datname string, coll *collection) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, coll.connectionsPerDatabase)
	for _, q := range coll.queries {
		semaphore <- struct{}{}
		j := job{datname: datname, query: q}
		wg.Go(func() {
			defer func() { <-semaphore }()

			if err := p.gatherMetricsFromQuery(ctx, acc, service, j, coll.timestamp); err != nil {
				acc.AddError(fmt.Errorf("database %q: %w", j.datname, err))
			}
		})
	}
	wg.Wait()
}

func (p *Postgresql) Stop() {
	p.releasePools(nil)
	p.service.Stop()
}

// refreshMetadata updates the cached server version, role and list of
// databases if a refresh is due. A collection between two refreshes does
// not read the server information again.
func (p *Postgresql) refreshMetadata(ctx context.Context) error {
	if !refreshDue(p.metadataUpdated, time.Now(), time.Duration(p.MetadataRefreshInterval)) {
		return nil
	}

	row := p.service.DB.QueryRowContext(ctx,
		`SELECT setting::integer / 100, pg_is_in_recovery(), current_database() FROM pg_settings WHERE name = 'server_version_num'`)
	if err := row.Scan(&p.dbVersion, &p.inRecovery, &p.currentDatabase); err != nil {
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

	if !slices.Equal(datnames, p.datnames) {
		if len(datnames) == 0 {
			p.Log.Warn("No database matches the datname_include/datname_exclude filters")
		} else {
			p.Log.Debugf("Databases matching the filters: %s", strings.Join(datnames, ", "))
		}
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

	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return err
	}
	numeric := make(map[string]bool, len(columnTypes))
	text := make(map[string]bool, len(columnTypes))
	for _, ct := range columnTypes {
		switch typeName := ct.DatabaseTypeName(); {
		case typeName == "NUMERIC":
			numeric[ct.Name()] = true
		case textTypes[typeName]:
			text[ct.Name()] = true
		}
	}

	// Prepare the row buffers once and reuse them for every row
	values := make([]any, len(columns))
	pointers := make([]any, len(columns))
	for i := range values {
		pointers[i] = &values[i]
	}

	rs := &resultSet{
		columns: columns,
		values:  values,
		numeric: numeric,
		text:    text,
		// Logging every value is only done on demand: the arguments of a
		// disabled Debugf call are still boxed and allocated for every
		// column of every row
		debug: p.Log.Level() >= telegraf.Debug,
	}

	for rows.Next() {
		if err := rows.Scan(pointers...); err != nil {
			return err
		}
		p.accRow(acc, rs, j, timestamp)
	}
	return rows.Err()
}

func (p *Postgresql) accRow(acc telegraf.Accumulator, rs *resultSet, j job, timestamp time.Time) {
	q := j.query

	// extract the database name from the row if available
	dbname := j.datname
	for i, col := range rs.columns {
		if col != "datname" {
			continue
		}
		if datname, ok := rs.values[i].(string); ok {
			dbname = datname
		}
		break
	}

	tags := make(map[string]string, len(q.additionalTags)+2)
	tags["server"] = p.service.SanitizedAddress
	tags["db"] = dbname

	fields := make(map[string]any, len(rs.columns))
	for i, col := range rs.columns {
		value := rs.values[i]
		if rs.debug {
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

		if q.numericFloat && rs.numeric[col] {
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

		if q.stringTags && rs.text[col] {
			if v, err := internal.ToString(value); err == nil {
				tags[col] = v
				continue
			}
		}

		if v, ok := value.([]byte); ok {
			fields[col] = string(v)
		} else {
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
			MaxIdle:             1,
			MaxOpen:             1,
			PreparedStatements:  true,
			KeepIdleConnections: true,
		}
	})
}
