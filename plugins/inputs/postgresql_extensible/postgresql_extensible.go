//go:generate ../../../tools/readme_config_includer/generator
package postgresql_extensible

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	// Required for SQL framework driver
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
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
	DatnameInclude            []string        `toml:"datname_include"`
	DatnameExclude            []string        `toml:"datname_exclude"`
	MetadataRefreshInterval   config.Duration `toml:"metadata_refresh_interval"`
	MaxConnectionsPerDatabase int             `toml:"max_connections_per_database"`
	MaxParallelDatabases      int             `toml:"max_parallel_databases"`
	KeepIdleConnections       *bool           `toml:"keep_idle_connections"`
	Log                       telegraf.Logger `toml:"-"`
	postgresql.Config

	service *postgresql.Service

	// Pools of the databases other than the one of the address, kept between
	// collections while keep_idle_connections is set. They are created and
	// read by the goroutines processing the databases, hence the lock.
	pools   map[string]*postgresql.Service
	poolsMu sync.Mutex

	keepIdle bool

	datnameFilter filter.Filter

	dbVersion       int
	currentDatabase string
	datnames        []string
	metadataUpdated time.Time
}

type query struct {
	Sqlquery    string `toml:"sqlquery"`
	Script      string `toml:"script"`
	Version     int    `deprecated:"1.28.0;use minVersion to specify minimal DB version this query supports"`
	MinVersion  int    `toml:"min_version"`
	MaxVersion  int    `toml:"max_version"`
	Tagvalue    string `toml:"tagvalue"`
	Measurement string `toml:"measurement"`
	Timestamp   string `toml:"timestamp"`
	Withdbname  bool   `deprecated:"1.22.4;use the sqlquery option to specify database to use"`

	additionalTags map[string]bool
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
	// Without keep_idle_connections follow max_idle, as setting it to zero was
	// the way to not keep the connection open in previous versions
	p.keepIdle = p.MaxIdle > 0
	if p.KeepIdleConnections != nil {
		p.keepIdle = *p.KeepIdleConnections
	}

	if len(p.DatnameInclude) > 0 || len(p.DatnameExclude) > 0 {
		f, err := filter.NewIncludeExcludeFilter(p.DatnameInclude, p.DatnameExclude)
		if err != nil {
			return fmt.Errorf("creating database filter failed: %w", err)
		}
		p.datnameFilter = f
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
	ctx := context.Background()

	if !p.keepIdle {
		defer p.closeIdleConnections()
	}

	if err := p.refreshMetadata(ctx); err != nil {
		if p.metadataRequired() {
			acc.AddError(err)
			if p.dbVersion == 0 {
				return nil
			}
		} else {
			// Servers without pg_settings such as the PgBouncer admin console
			// still run the queries not limited to a version
			p.dbVersion = 0
		}
	}

	queries := make([]*query, 0, len(p.Queries))
	for i := range p.Queries {
		q := &p.Queries[i]
		if !matchesVersion(q, p.dbVersion) {
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
	if p.datnameFilter != nil {
		datnames = p.datnames
		// Without a dbname in the address the server connects to the database
		// named after the user, not to the ConnectionDatabase. The single
		// database mode is excluded here only to keep the behavior of old
		// configurations.
		connDatname = p.currentDatabase
	}

	p.releasePools(datnames)

	if len(datnames) == 0 {
		return nil
	}

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
		if p.keepIdle {
			kept = used
		}
		p.Log.Debugf("Collection used %d connections, %d kept open", used, kept)
	}

	return nil
}

// openConnections counts the connections currently open in all pools
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

// gatherInDatabase picks the pool of the given database and runs the queries on it
func (p *Postgresql) gatherInDatabase(ctx context.Context, acc telegraf.Accumulator, datname string, coll *collection) error {
	if datname == coll.connDatname {
		p.gatherDatabase(ctx, acc, p.service, datname, coll)
		return nil
	}

	if p.keepIdle {
		service, err := p.poolFor(datname, coll.connectionsPerDatabase)
		if err != nil {
			return err
		}
		p.gatherDatabase(ctx, acc, service, datname, coll)
		return nil
	}

	// Create the pools one at a time as they all access the same address secret
	p.poolsMu.Lock()
	service, err := p.newPool(datname, coll.connectionsPerDatabase)
	p.poolsMu.Unlock()
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
				acc.AddError(err)
			}
		})
	}
	wg.Wait()
}

func (p *Postgresql) Stop() {
	p.releasePools(nil)
	p.service.Stop()
}

// refreshMetadata updates the cached server information if a refresh is
// due. A collection between two refreshes does not read it again.
func (p *Postgresql) refreshMetadata(ctx context.Context) error {
	if !refreshDue(p.metadataUpdated, time.Now(), time.Duration(p.MetadataRefreshInterval)) {
		return nil
	}

	row := p.service.DB.QueryRowContext(ctx,
		`SELECT setting::integer / 100, current_database() FROM pg_settings WHERE name = 'server_version_num'`)
	if err := row.Scan(&p.dbVersion, &p.currentDatabase); err != nil {
		return fmt.Errorf("querying metadata failed: %w", err)
	}

	if p.datnameFilter != nil {
		if err := p.refreshDatnames(ctx); err != nil {
			return err
		}
	}

	p.metadataUpdated = time.Now()
	return nil
}

// metadataRequired tells whether the queries depend on the server information
func (p *Postgresql) metadataRequired() bool {
	return p.datnameFilter != nil
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
		if p.datnameFilter.Match(datname) {
			datnames = append(datnames, datname)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("reading database list failed: %w", err)
	}

	if p.metadataUpdated.IsZero() || !slices.Equal(datnames, p.datnames) {
		if len(datnames) == 0 {
			p.Log.Warn("No database matches the datname_include/datname_exclude filters")
		} else {
			p.Log.Debugf("Databases matching the filters: %s", strings.Join(datnames, ", "))
		}
	}
	p.datnames = datnames

	return nil
}

func (p *Postgresql) gatherMetricsFromQuery(
	ctx context.Context,
	acc telegraf.Accumulator,
	service *postgresql.Service,
	j job,
	timestamp time.Time,
) error {
	if err := p.runQuery(ctx, acc, service, j, timestamp); err != nil {
		return fmt.Errorf("database %q: query %q: %w", j.datname, j.query.Measurement, err)
	}
	return nil
}

func (p *Postgresql) runQuery(
	ctx context.Context,
	acc telegraf.Accumulator,
	service *postgresql.Service,
	j job,
	timestamp time.Time,
) error {
	rows, err := service.DB.QueryContext(ctx, j.query.Sqlquery)
	if err != nil {
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return err
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

		if v, ok := value.([]byte); ok {
			fields[col] = string(v)
		} else {
			fields[col] = value
		}
	}
	acc.AddFields(q.Measurement, fields, tags, timestamp)
}

func refreshDue(updated, now time.Time, interval time.Duration) bool {
	if updated.IsZero() || interval <= 0 || now.Before(updated) {
		return true
	}
	return !now.Before(updated.Truncate(interval).Add(interval))
}

func matchesVersion(q *query, dbVersion int) bool {
	return q.MinVersion <= dbVersion && (q.MaxVersion == 0 || q.MaxVersion > dbVersion)
}

func init() {
	inputs.Add("postgresql_extensible", func() telegraf.Input {
		return &Postgresql{
			MaxIdle:            1,
			MaxOpen:            1,
			PreparedStatements: true,
		}
	})
}
