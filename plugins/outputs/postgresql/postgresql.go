package postgresql

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/coocood/freecache"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/outputs/postgresql/template"
	"github.com/influxdata/telegraf/plugins/outputs/postgresql/utils"
	"github.com/influxdata/toml"
)

type dbh interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	Exec(ctx context.Context, sql string, arguments ...interface{}) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

var sampleConfig = `
  ## specify address via a url matching:
  ##   postgres://[pqgotest[:password]]@localhost[/dbname]\
  ##       ?sslmode=[disable|verify-ca|verify-full]
  ## or a simple string:
  ##   host=localhost user=pqotest password=... sslmode=... dbname=app_production
  ##
  ## All connection parameters are optional. Also supported are PG environment vars
  ## e.g. PGPASSWORD, PGHOST, PGUSER, PGDATABASE 
  ## all supported vars here: https://www.postgresql.org/docs/current/libpq-envars.html
  ##
  ## Non-standard parameters:
  ##   pool_max_conns (default: 1) - Maximum size of connection pool for parallel (per-batch per-table) inserts.
  ##   pool_min_conns (default: 0) - Minimum size of connection pool.
  ##   pool_max_conn_lifetime (default: 0s) - Maximum age of a connection before closing.
  ##   pool_max_conn_idle_time (default: 0s) - Maximum idle time of a connection before closing.
  ##   pool_health_check_period (default: 0s) - Duration between health checks on idle connections.
  ##
  ## Without the dbname parameter, the driver will default to a database
  ## with the same name as the user. This dbname is just for instantiating a
  ## connection with the server and doesn't restrict the databases we are trying
  ## to grab metrics for.
  ##
  #connection = "host=localhost user=postgres sslmode=verify-full"

  ## Postgres schema to use.
  schema = "public"

  ## Store tags as foreign keys in the metrics table. Default is false.
  tags_as_foreign_keys = false

  ## Suffix to append to table name (measurement name) for the foreign tag table.
  tag_table_suffix = "_tag"

  ## Deny inserting metrics if the foreign tag can't be inserted.
  foreign_tag_constraint = false

  ## Store all tags as a JSONB object in a single 'tags' column.
  tags_as_jsonb = false

  ## Store all fields as a JSONB object in a single 'fields' column.
  fields_as_jsonb = false

  ## Templated statements to execute when creating a new table.
  create_templates = [
    '''CREATE TABLE {{.table}} ({{.columns}})''',
  ]

  ## Templated statements to execute when adding columns to a table.
  ## Set to an empty list to disable. Points containing tags for which there is no column will be skipped. Points
  ## containing fields for which there is no column will have the field omitted.
  add_column_templates = [
    '''ALTER TABLE {{.table}} ADD COLUMN IF NOT EXISTS {{.columns|join ", ADD COLUMN IF NOT EXISTS "}}''',
  ]

  ## Templated statements to execute when creating a new tag table.
  tag_table_create_templates = [
    '''CREATE TABLE {{.table}} ({{.columns}}, PRIMARY KEY (tag_id))''',
  ]

  ## Templated statements to execute when adding columns to a tag table.
  ## Set to an empty list to disable. Points containing tags for which there is no column will be skipped.
  tag_table_add_column_templates = [
    '''ALTER TABLE {{.table}} ADD COLUMN IF NOT EXISTS {{.columns|join ", ADD COLUMN IF NOT EXISTS "}}''',
  ]

  ## When using pool_max_conns>1, an a temporary error occurs, the query is retried with an incremental backoff. This
  ## controls the maximum backoff duration.
  retry_max_backoff = "15s"

  ## Enable & set the log level for the Postgres driver.
  # log_level = "info" # trace, debug, info, warn, error, none
`

type Postgresql struct {
	Connection                 string
	Schema                     string
	TagsAsForeignKeys          bool
	TagTableSuffix             string
	ForeignTagConstraint       bool
	TagsAsJsonb                bool
	FieldsAsJsonb              bool
	CreateTemplates            []*template.Template
	AddColumnTemplates         []*template.Template
	TagTableCreateTemplates    []*template.Template
	TagTableAddColumnTemplates []*template.Template
	RetryMaxBackoff            config.Duration
	LogLevel                   string

	dbContext       context.Context
	dbContextCancel func()
	db              *pgxpool.Pool
	tableManager    *TableManager
	tagsCache       *freecache.Cache

	writeChan      chan *TableSource
	writeWaitGroup *utils.WaitGroup

	Logger telegraf.Logger
}

func init() {
	outputs.Add("postgresql", func() telegraf.Output { return newPostgresql() })
}

func newPostgresql() *Postgresql {
	p := &Postgresql{
		Logger: models.NewLogger("outputs", "postgresql", ""),
	}
	if err := toml.Unmarshal([]byte(p.SampleConfig()), p); err != nil {
		panic(err.Error())
	}
	return p
}

func (p *Postgresql) SampleConfig() string { return sampleConfig }
func (p *Postgresql) Description() string  { return "Send metrics to PostgreSQL" }

// Connect establishes a connection to the target database and prepares the cache
func (p *Postgresql) Connect() error {
	poolConfig, err := pgxpool.ParseConfig(p.Connection)
	if err != nil {
		return err
	}
	parsedConfig, _ := pgx.ParseConfig(p.Connection)
	if _, ok := parsedConfig.Config.RuntimeParams["pool_max_conns"]; !ok {
		// The pgx default for pool_max_conns is 4. However we want to default to 1.
		poolConfig.MaxConns = 1
	}

	if p.LogLevel != "" {
		poolConfig.ConnConfig.Logger = utils.PGXLogger{p.Logger}
		poolConfig.ConnConfig.LogLevel, err = pgx.LogLevelFromString(p.LogLevel)
		if err != nil {
			return fmt.Errorf("invalid log level")
		}
	}

	// Yes, we're not supposed to store the context. However since we don't receive a context, we have to.
	p.dbContext, p.dbContextCancel = context.WithCancel(context.Background())
	p.db, err = pgxpool.ConnectConfig(p.dbContext, poolConfig)
	if err != nil {
		p.Logger.Errorf("Couldn't connect to server\n%v", err)
		return err
	}
	p.tableManager = NewTableManager(p)

	if p.TagsAsForeignKeys {
		p.tagsCache = freecache.NewCache(5 * 1024 * 1024) // 5MB
	}

	maxConns := int(p.db.Stat().MaxConns())
	if maxConns > 1 {
		p.writeChan = make(chan *TableSource)
		p.writeWaitGroup = utils.NewWaitGroup()
		for i := 0; i < maxConns; i++ {
			p.writeWaitGroup.Add(1)
			go p.writeWorker(p.dbContext)
		}
	}

	return nil
}

// Close closes the connection(s) to the database.
func (p *Postgresql) Close() error {
	if p.writeChan != nil {
		// We're using async mode. Gracefully close with timeout.
		close(p.writeChan)
		select {
		case <-p.writeWaitGroup.C():
		case <-time.NewTimer(time.Second * 5).C:
		}
	}

	// Die!
	p.dbContextCancel()
	p.db.Close()
	p.tableManager = nil
	return nil
}

func (p *Postgresql) Write(metrics []telegraf.Metric) error {
	tableSources := NewTableSources(p, metrics)

	if p.db.Stat().MaxConns() > 1 {
		return p.writeConcurrent(tableSources)
	} else {
		return p.writeSequential(tableSources)
	}
}

func (p *Postgresql) writeSequential(tableSources map[string]*TableSource) error {
	tx, err := p.db.Begin(p.dbContext)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback(p.dbContext)

	for _, tableSource := range tableSources {
		err := p.writeMetricsFromMeasure(p.dbContext, tx, tableSource)
		if err != nil {
			if isTempError(err) {
				return err
			}
			p.Logger.Errorf("write error (permanent, dropping sub-batch): %v", err)
		}
	}

	if err := tx.Commit(p.dbContext); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

func (p *Postgresql) writeConcurrent(tableSources map[string]*TableSource) error {
	for _, tableSource := range tableSources {
		select {
		case p.writeChan <- tableSource:
		case <-p.dbContext.Done():
			return nil
		}
	}
	return nil
}

func (p *Postgresql) writeWorker(ctx context.Context) {
	defer p.writeWaitGroup.Done()
	for {
		select {
		case tableSource, ok := <-p.writeChan:
			if !ok {
				return
			}
			if err := p.writeRetry(ctx, tableSource); err != nil {
				p.Logger.Errorf("write error (permanent, dropping sub-batch): %v", err)
			}
		case <-p.dbContext.Done():
			return
		}
	}
}

// This is a subset of net.Error
type maybeTempError interface {
	error
	Temporary() bool
}

// isTempError reports whether the error received during a metric write operation is temporary or permanent.
// A temporary error is one that if the write were retried at a later time, that it might succeed.
// Note however that this applies to the transaction as a whole, not the individual operation. Meaning for example a
// write might come in that needs a new table created, but another worker already created the table in between when we
// checked for it, and tried to create it. In this case, the operation error is permanent, as we can try `CREATE TABLE`
// again and it will still fail. But if we retry the transaction from scratch, when we perform the table check we'll see
// it exists, so we consider the error temporary.
func isTempError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr); pgErr != nil {
		// https://www.postgresql.org/docs/12/errcodes-appendix.html
		errClass := pgErr.Code[:2]
		switch errClass {
		case "42": // Syntax Error or Access Rule Violation
			switch pgErr.Code {
			case "42701": // duplicate_column
				return true
			case "42P07": // duplicate_table
				return true
			}
		case "53": // Insufficient Resources
			return true
		case "57": // Operator Intervention
			return true
		case "23": // Integrity Constraint Violation
			switch pgErr.Code {
			case "23505": // unique_violation
				if strings.Contains(err.Error(), "pg_type_typname_nsp_index") {
					// Happens when you try to create 2 tables simultaneously.
					return true
				}
			}
		}
		// Assume that any other error that comes from postgres is a permanent error
		return false
	}

	if mtErr := maybeTempError(nil); errors.As(err, &mtErr) {
		return mtErr.Temporary()
	}

	// Assume that any other error is permanent.
	// This may mean that we incorrectly discard data that could have been retried, but the alternative is that we get
	// stuck retrying data that will never succeed, causing good data to be dropped because the buffer fills up.
	return false
}

func (p *Postgresql) writeRetry(ctx context.Context, tableSource *TableSource) error {
	backoff := time.Duration(0)
	for {
		tx, err := p.db.Begin(ctx)
		if err != nil {
			return err
		}

		err = p.writeMetricsFromMeasure(ctx, tx, tableSource)
		if err == nil {
			tx.Commit(ctx)
			return nil
		}

		tx.Rollback(ctx)
		if !isTempError(err) {
			return err
		}
		p.Logger.Errorf("write error (retry in %s): %v", backoff, err)
		tableSource.Reset()
		time.Sleep(backoff)

		if backoff == 0 {
			backoff = time.Millisecond * 250
		} else {
			backoff *= 2
			if backoff > time.Duration(p.RetryMaxBackoff) {
				backoff = time.Duration(p.RetryMaxBackoff)
			}
		}
	}
}

// Writes the metrics from a specified measure. All the provided metrics must belong to the same measurement.
func (p *Postgresql) writeMetricsFromMeasure(ctx context.Context, db dbh, tableSource *TableSource) error {
	err := p.tableManager.MatchSource(ctx, db, tableSource)
	if err != nil {
		return err
	}

	if p.TagsAsForeignKeys {
		if err := p.WriteTagTable(ctx, db, tableSource); err != nil {
			if p.ForeignTagConstraint {
				return fmt.Errorf("writing to tag table '%s': %s", tableSource.Name()+p.TagTableSuffix, err)
			} else {
				// log and continue. As the admin can correct the issue, and tags don't change over time, they can be
				// added from future metrics after issue is corrected.
				p.Logger.Errorf("writing to tag table '%s': %s", tableSource.Name()+p.TagTableSuffix, err)
			}
		}
	}

	fullTableName := utils.FullTableName(p.Schema, tableSource.Name())
	if _, err := db.CopyFrom(ctx, fullTableName, tableSource.ColumnNames(), tableSource); err != nil {
		return err
	}

	return nil
}

func (p *Postgresql) WriteTagTable(ctx context.Context, db dbh, tableSource *TableSource) error {
	ttsrc := NewTagTableSource(tableSource)

	// Check whether we have any tags to insert
	if !ttsrc.Next() {
		return nil
	}
	ttsrc.Reset()

	// need a transaction so that if it errors, we don't roll back the parent transaction, just the tags
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	ident := pgx.Identifier{ttsrc.postgresql.Schema, ttsrc.Name()}
	identTemp := pgx.Identifier{ttsrc.Name() + "_temp"}
	sql := fmt.Sprintf("CREATE TEMP TABLE %s (LIKE %s) ON COMMIT DROP", identTemp.Sanitize(), ident.Sanitize())
	if _, err := tx.Exec(ctx, sql); err != nil {
		return fmt.Errorf("creating tags temp table: %w", err)
	}

	if _, err := tx.CopyFrom(ctx, identTemp, ttsrc.ColumnNames(), ttsrc); err != nil {
		return fmt.Errorf("copying into tags temp table: %w", err)
	}

	if _, err := tx.Exec(ctx, fmt.Sprintf("INSERT INTO %s SELECT * FROM %s ORDER BY tag_id ON CONFLICT (tag_id) DO NOTHING", ident.Sanitize(), identTemp.Sanitize())); err != nil {
		return fmt.Errorf("inserting into tags table: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	ttsrc.UpdateCache()
	return nil
}
