//go:generate ../../../tools/readme_config_includer/generator
package snowpipe_streaming

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	gosql "database/sql"
	_ "embed"
	"encoding/pem"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/snowflakedb/gosnowflake"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

type SnowpipeStreaming struct {
	Account             string          `toml:"account"`
	User                string          `toml:"user"`
	PrivateKeyPath      string          `toml:"private_key_path"`
	PrivateKeyPass      string          `toml:"private_key_passphrase"`
	Role                string          `toml:"role"`
	Database            string          `toml:"database"`
	Schema              string          `toml:"schema"`
	Table               string          `toml:"table"`
	BatchSize           int             `toml:"batch_size"`
	RetryMax            int             `toml:"retry_max"`
	RetryDelay          config.Duration `toml:"retry_delay"`
	TimestampColumn     string          `toml:"timestamp_column"`
	TagColumns          []string        `toml:"tag_columns"`
	FieldColumns        []string        `toml:"field_columns"`
	CreateTable         bool            `toml:"create_table"`
	TableSchemaCacheTTL config.Duration `toml:"table_schema_cache_ttl"`

	Log telegraf.Logger `toml:"-"`

	db           *gosql.DB
	tableTmpl    *template.Template
	tableHasTmpl bool
	tagSet       map[string]bool
	fieldSet     map[string]bool

	schemaMu    sync.RWMutex
	schemaCache map[string]*tableSchema

	// For testing: allow overriding the connection opener
	openDB func() (*gosql.DB, error)
}

type tableSchema struct {
	columns   map[string]bool
	fetchedAt time.Time
}

func (*SnowpipeStreaming) SampleConfig() string {
	return sampleConfig
}

func (s *SnowpipeStreaming) Init() error {
	if s.Account == "" {
		return errors.New(`"account" is required`)
	}
	if s.User == "" {
		return errors.New(`"user" is required`)
	}
	if s.Database == "" {
		return errors.New(`"database" is required`)
	}
	if s.Schema == "" {
		return errors.New(`"schema" is required`)
	}
	if s.Table == "" {
		return errors.New(`"table" is required`)
	}

	if strings.Contains(s.Table, "{{") {
		tmpl, err := template.New("table").Parse(s.Table)
		if err != nil {
			return fmt.Errorf("parsing table template: %w", err)
		}
		s.tableTmpl = tmpl
		s.tableHasTmpl = true
	}

	if len(s.TagColumns) > 0 {
		s.tagSet = make(map[string]bool, len(s.TagColumns))
		for _, t := range s.TagColumns {
			s.tagSet[t] = true
		}
	}
	if len(s.FieldColumns) > 0 {
		s.fieldSet = make(map[string]bool, len(s.FieldColumns))
		for _, f := range s.FieldColumns {
			s.fieldSet[f] = true
		}
	}

	s.schemaCache = make(map[string]*tableSchema)

	return nil
}

func (s *SnowpipeStreaming) Connect() error {
	var db *gosql.DB
	var err error

	if s.openDB != nil {
		db, err = s.openDB()
	} else {
		var dsn string
		dsn, err = s.buildDSN()
		if err != nil {
			return fmt.Errorf("building DSN: %w", err)
		}
		db, err = gosql.Open("snowflake", dsn)
	}
	if err != nil {
		return fmt.Errorf("opening snowflake connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("pinging snowflake: %w", err)
	}

	s.db = db
	return nil
}

func (s *SnowpipeStreaming) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *SnowpipeStreaming) Write(metrics []telegraf.Metric) error {
	grouped := s.groupByTable(metrics)
	for tableName, rows := range grouped {
		if err := s.writeTable(tableName, rows); err != nil {
			return fmt.Errorf("writing to table %q: %w", tableName, err)
		}
	}
	return nil
}

func (s *SnowpipeStreaming) writeTable(tableName string, metrics []telegraf.Metric) error {
	if s.CreateTable {
		if err := s.ensureTable(tableName, metrics[0]); err != nil {
			return err
		}
	}

	for start := 0; start < len(metrics); start += s.BatchSize {
		end := start + s.BatchSize
		if end > len(metrics) {
			end = len(metrics)
		}
		if err := s.insertBatch(tableName, metrics[start:end]); err != nil {
			return err
		}
	}
	return nil
}

func (s *SnowpipeStreaming) insertBatch(tableName string, metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	columns, allValues := s.metricsToRows(metrics)
	if len(columns) == 0 {
		return nil
	}

	query := s.buildInsertQuery(tableName, columns, len(metrics))

	flat := make([]interface{}, 0, len(columns)*len(metrics))
	for _, row := range allValues {
		flat = append(flat, row...)
	}

	var lastErr error
	for attempt := 0; attempt <= s.RetryMax; attempt++ {
		if attempt > 0 {
			delay := time.Duration(s.RetryDelay) * (1 << (attempt - 1))
			time.Sleep(delay)
		}

		_, err := s.db.Exec(query, flat...)
		if err == nil {
			return nil
		}
		lastErr = err

		if !isTransientError(err) {
			return fmt.Errorf("insert failed: %w", err)
		}
		s.Log.Warnf("Transient error on attempt %d/%d for table %q: %v", attempt+1, s.RetryMax+1, tableName, err)
	}

	return fmt.Errorf("insert failed after %d retries: %w", s.RetryMax, lastErr)
}

func (s *SnowpipeStreaming) metricsToRows(metrics []telegraf.Metric) ([]string, [][]interface{}) {
	columnOrder := s.buildColumnOrder(metrics[0])
	columnSet := make(map[string]bool, len(columnOrder))
	for _, c := range columnOrder {
		columnSet[c] = true
	}

	rows := make([][]interface{}, 0, len(metrics))
	for _, m := range metrics {
		row := s.metricToRow(m, columnOrder, columnSet)
		rows = append(rows, row)
	}

	return columnOrder, rows
}

func (s *SnowpipeStreaming) buildColumnOrder(m telegraf.Metric) []string {
	columns := make([]string, 0, 1+len(m.TagList())+len(m.FieldList()))

	if s.TimestampColumn != "" {
		columns = append(columns, s.TimestampColumn)
	}

	columns = append(columns, "name")

	for _, tag := range m.TagList() {
		if s.tagSet != nil && !s.tagSet[tag.Key] {
			continue
		}
		columns = append(columns, tag.Key)
	}

	for _, field := range m.FieldList() {
		if s.fieldSet != nil && !s.fieldSet[field.Key] {
			continue
		}
		columns = append(columns, field.Key)
	}

	return columns
}

func (s *SnowpipeStreaming) metricToRow(m telegraf.Metric, columns []string, columnSet map[string]bool) []interface{} {
	vals := make(map[string]interface{}, len(columns))

	if s.TimestampColumn != "" {
		vals[s.TimestampColumn] = m.Time()
	}
	vals["name"] = m.Name()

	for _, tag := range m.TagList() {
		if s.tagSet != nil && !s.tagSet[tag.Key] {
			continue
		}
		if columnSet[tag.Key] {
			vals[tag.Key] = tag.Value
		}
	}

	for _, field := range m.FieldList() {
		if s.fieldSet != nil && !s.fieldSet[field.Key] {
			continue
		}
		if columnSet[field.Key] {
			vals[field.Key] = sanitizeFieldValue(field.Value)
		}
	}

	row := make([]interface{}, len(columns))
	for i, col := range columns {
		row[i] = vals[col]
	}
	return row
}

func sanitizeFieldValue(v interface{}) interface{} {
	if f, ok := v.(float64); ok {
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return nil
		}
	}
	return v
}

func (s *SnowpipeStreaming) buildInsertQuery(tableName string, columns []string, numRows int) string {
	quoted := make([]string, len(columns))
	for i, c := range columns {
		quoted[i] = quoteIdent(c)
	}

	placeholders := make([]string, len(columns))
	for i := range columns {
		placeholders[i] = "?"
	}
	rowPlaceholder := "(" + strings.Join(placeholders, ", ") + ")"

	rowPlaceholders := make([]string, numRows)
	for i := range numRows {
		rowPlaceholders[i] = rowPlaceholder
	}

	return fmt.Sprintf("INSERT INTO %s.%s.%s (%s) VALUES %s",
		quoteIdent(s.Database),
		quoteIdent(s.Schema),
		quoteIdent(tableName),
		strings.Join(quoted, ", "),
		strings.Join(rowPlaceholders, ", "),
	)
}

func (s *SnowpipeStreaming) groupByTable(metrics []telegraf.Metric) map[string][]telegraf.Metric {
	groups := make(map[string][]telegraf.Metric)
	for _, m := range metrics {
		name := s.resolveTableName(m)
		groups[name] = append(groups[name], m)
	}
	return groups
}

func (s *SnowpipeStreaming) resolveTableName(m telegraf.Metric) string {
	if !s.tableHasTmpl {
		return s.Table
	}
	var b strings.Builder
	if err := s.tableTmpl.Execute(&b, m); err != nil {
		s.Log.Errorf("Executing table template: %v, falling back to literal table name", err)
		return s.Table
	}
	return b.String()
}

func (s *SnowpipeStreaming) ensureTable(tableName string, sample telegraf.Metric) error {
	s.schemaMu.RLock()
	cached, exists := s.schemaCache[tableName]
	s.schemaMu.RUnlock()

	ttl := time.Duration(s.TableSchemaCacheTTL)
	if exists && time.Since(cached.fetchedAt) < ttl {
		return s.evolveSchema(tableName, sample, cached)
	}

	s.schemaMu.Lock()
	defer s.schemaMu.Unlock()

	if err := s.createTableIfNotExists(tableName, sample); err != nil {
		return err
	}

	schema, err := s.fetchTableSchema(tableName)
	if err != nil {
		return err
	}
	s.schemaCache[tableName] = schema

	return s.evolveSchemaLocked(tableName, sample, schema)
}

func (s *SnowpipeStreaming) createTableIfNotExists(tableName string, sample telegraf.Metric) error {
	columns := s.buildColumnOrder(sample)
	colDefs := make([]string, len(columns))
	for i, col := range columns {
		colDefs[i] = fmt.Sprintf("%s %s", quoteIdent(col), s.sqlTypeFor(col, sample))
	}

	query := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s.%s.%s (%s)",
		quoteIdent(s.Database),
		quoteIdent(s.Schema),
		quoteIdent(tableName),
		strings.Join(colDefs, ", "),
	)

	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("creating table %q: %w", tableName, err)
	}
	return nil
}

func (s *SnowpipeStreaming) sqlTypeFor(col string, m telegraf.Metric) string {
	if col == s.TimestampColumn {
		return "TIMESTAMP_NTZ"
	}
	if col == "name" {
		return "VARCHAR"
	}
	for _, tag := range m.TagList() {
		if tag.Key == col {
			return "VARCHAR"
		}
	}
	for _, field := range m.FieldList() {
		if field.Key == col {
			return goTypeToSnowflake(field.Value)
		}
	}
	return "VARCHAR"
}

func goTypeToSnowflake(v interface{}) string {
	switch v.(type) {
	case int, int8, int16, int32, int64:
		return "NUMBER"
	case uint, uint8, uint16, uint32, uint64:
		return "NUMBER"
	case float32, float64:
		return "DOUBLE"
	case bool:
		return "BOOLEAN"
	default:
		return "VARCHAR"
	}
}

func (s *SnowpipeStreaming) fetchTableSchema(tableName string) (*tableSchema, error) {
	query := fmt.Sprintf(
		"SELECT COLUMN_NAME FROM %s.INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = '%s' AND TABLE_NAME = '%s'",
		s.Database,
		strings.ToUpper(s.Schema),
		strings.ToUpper(tableName),
	)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("fetching schema for %q: %w", tableName, err)
	}
	defer rows.Close()

	cols := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		cols[strings.ToUpper(name)] = true
	}

	return &tableSchema{columns: cols, fetchedAt: time.Now()}, rows.Err()
}

func (s *SnowpipeStreaming) evolveSchema(tableName string, sample telegraf.Metric, schema *tableSchema) error {
	s.schemaMu.Lock()
	defer s.schemaMu.Unlock()
	return s.evolveSchemaLocked(tableName, sample, schema)
}

func (s *SnowpipeStreaming) evolveSchemaLocked(tableName string, sample telegraf.Metric, schema *tableSchema) error {
	needed := s.buildColumnOrder(sample)
	for _, col := range needed {
		if schema.columns[strings.ToUpper(col)] {
			continue
		}
		sqlType := s.sqlTypeFor(col, sample)
		alter := fmt.Sprintf("ALTER TABLE %s.%s.%s ADD COLUMN %s %s",
			quoteIdent(s.Database),
			quoteIdent(s.Schema),
			quoteIdent(tableName),
			quoteIdent(col),
			sqlType,
		)
		if _, err := s.db.Exec(alter); err != nil {
			s.Log.Warnf("Failed to add column %q to %q: %v", col, tableName, err)
			continue
		}
		schema.columns[strings.ToUpper(col)] = true
		s.Log.Infof("Added column %q (%s) to table %q", col, sqlType, tableName)
	}
	return nil
}

func (s *SnowpipeStreaming) buildDSN() (string, error) {
	cfg := &gosnowflake.Config{
		Account:  s.Account,
		User:     s.User,
		Database: s.Database,
		Schema:   s.Schema,
		Role:     s.Role,
	}

	if s.PrivateKeyPath != "" {
		key, err := loadPrivateKey(s.PrivateKeyPath, s.PrivateKeyPass)
		if err != nil {
			return "", fmt.Errorf("loading private key: %w", err)
		}
		cfg.Authenticator = gosnowflake.AuthTypeJwt
		cfg.PrivateKey = key
	}

	dsn, err := gosnowflake.DSN(cfg)
	if err != nil {
		return "", fmt.Errorf("building snowflake DSN: %w", err)
	}
	return dsn, nil
}

func loadPrivateKey(path, passphrase string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	var keyBytes []byte
	if passphrase != "" {
		keyBytes, err = x509.DecryptPEMBlock(block, []byte(passphrase)) //nolint:staticcheck // SA1019: required for PKCS#5 keys
		if err != nil {
			return nil, fmt.Errorf("decrypting private key: %w", err)
		}
	} else {
		keyBytes = block.Bytes
	}

	parsed, err := x509.ParsePKCS8PrivateKey(keyBytes)
	if err != nil {
		// Fall back to PKCS1
		key, err2 := x509.ParsePKCS1PrivateKey(keyBytes)
		if err2 != nil {
			return nil, fmt.Errorf("parsing private key (PKCS8: %v, PKCS1: %v)", err, err2)
		}
		return key, nil
	}

	rsaKey, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not RSA")
	}
	return rsaKey, nil
}

func isTransientError(err error) bool {
	if err == nil {
		return false
	}
	var sfErr *gosnowflake.SnowflakeError
	if errors.As(err, &sfErr) {
		// HTTP 429, 503, and internal server errors are transient
		switch {
		case sfErr.Number == gosnowflake.ErrCodeServiceUnavailable:
			return true
		case sfErr.Number == gosnowflake.ErrCodeFailedToConnect:
			return true
		}
	}

	msg := err.Error()
	return strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "i/o timeout") ||
		strings.Contains(msg, "service unavailable")
}

func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func init() {
	outputs.Add("snowpipe_streaming", func() telegraf.Output {
		return &SnowpipeStreaming{
			BatchSize:           1000,
			RetryMax:            3,
			RetryDelay:          config.Duration(1 * time.Second),
			TimestampColumn:     "timestamp",
			TableSchemaCacheTTL: config.Duration(5 * time.Minute),
		}
	})
}

// Compile-time interface check
var _ telegraf.Output = (*SnowpipeStreaming)(nil)

// connectContext is a helper for future use with context-aware connections
func (s *SnowpipeStreaming) connectContext(_ context.Context) error {
	return s.Connect()
}
