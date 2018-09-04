// Package stdlib is the compatibility layer from pgx to database/sql.
//
// A database/sql connection can be established through sql.Open.
//
//	db, err := sql.Open("pgx", "postgres://pgx_md5:secret@localhost:5432/pgx_test?sslmode=disable")
//	if err != nil {
//		return err
//	}
//
// Or from a DSN string.
//
//	db, err := sql.Open("pgx", "user=postgres password=secret host=localhost port=5432 database=pgx_test sslmode=disable")
//	if err != nil {
//		return err
//	}
//
// A DriverConfig can be used to further configure the connection process. This
// allows configuring TLS configuration, setting a custom dialer, logging, and
// setting an AfterConnect hook.
//
//	driverConfig := stdlib.DriverConfig{
// 		ConnConfig: pgx.ConnConfig{
//			Logger:   logger,
//		},
//		AfterConnect: func(c *pgx.Conn) error {
//			// Ensure all connections have this temp table available
//			_, err := c.Exec("create temporary table foo(...)")
//			return err
//		},
//	}
//
//	stdlib.RegisterDriverConfig(&driverConfig)
//
//	db, err := sql.Open("pgx", driverConfig.ConnectionString("postgres://pgx_md5:secret@127.0.0.1:5432/pgx_test"))
//	if err != nil {
//		return err
//	}
//
// pgx uses standard PostgreSQL positional parameters in queries. e.g. $1, $2.
// It does not support named parameters.
//
//	db.QueryRow("select * from users where id=$1", userID)
//
// AcquireConn and ReleaseConn acquire and release a *pgx.Conn from the standard
// database/sql.DB connection pool. This allows operations that must be
// performed on a single connection, but should not be run in a transaction or
// to use pgx specific functionality.
//
//	conn, err := stdlib.AcquireConn(db)
//	if err != nil {
//		return err
//	}
//	defer stdlib.ReleaseConn(db, conn)
//
//	// do stuff with pgx.Conn
//
// It also can be used to enable a fast path for pgx while preserving
// compatibility with other drivers and database.
//
//	conn, err := stdlib.AcquireConn(db)
//	if err == nil {
//		// fast path with pgx
//		// ...
//		// release conn when done
//		stdlib.ReleaseConn(db, conn)
//	} else {
//		// normal path for other drivers and databases
//	}
package stdlib

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
)

// oids that map to intrinsic database/sql types. These will be allowed to be
// binary, anything else will be forced to text format
var databaseSqlOIDs map[pgtype.OID]bool

var pgxDriver *Driver

type ctxKey int

var ctxKeyFakeTx ctxKey = 0

var ErrNotPgx = errors.New("not pgx *sql.DB")

func init() {
	pgxDriver = &Driver{
		configs:     make(map[int64]*DriverConfig),
		fakeTxConns: make(map[*pgx.Conn]*sql.Tx),
	}
	sql.Register("pgx", pgxDriver)

	databaseSqlOIDs = make(map[pgtype.OID]bool)
	databaseSqlOIDs[pgtype.BoolOID] = true
	databaseSqlOIDs[pgtype.ByteaOID] = true
	databaseSqlOIDs[pgtype.CIDOID] = true
	databaseSqlOIDs[pgtype.DateOID] = true
	databaseSqlOIDs[pgtype.Float4OID] = true
	databaseSqlOIDs[pgtype.Float8OID] = true
	databaseSqlOIDs[pgtype.Int2OID] = true
	databaseSqlOIDs[pgtype.Int4OID] = true
	databaseSqlOIDs[pgtype.Int8OID] = true
	databaseSqlOIDs[pgtype.OIDOID] = true
	databaseSqlOIDs[pgtype.TimestampOID] = true
	databaseSqlOIDs[pgtype.TimestamptzOID] = true
	databaseSqlOIDs[pgtype.XIDOID] = true
}

type Driver struct {
	configMutex sync.Mutex
	configCount int64
	configs     map[int64]*DriverConfig

	fakeTxMutex sync.Mutex
	fakeTxConns map[*pgx.Conn]*sql.Tx
}

func (d *Driver) Open(name string) (driver.Conn, error) {
	var connConfig pgx.ConnConfig
	var afterConnect func(*pgx.Conn) error
	if len(name) >= 9 && name[0] == 0 {
		idBuf := []byte(name)[1:9]
		id := int64(binary.BigEndian.Uint64(idBuf))
		connConfig = d.configs[id].ConnConfig
		afterConnect = d.configs[id].AfterConnect
		name = name[9:]
	}

	parsedConfig, err := pgx.ParseConnectionString(name)
	if err != nil {
		return nil, err
	}
	connConfig = connConfig.Merge(parsedConfig)

	conn, err := pgx.Connect(connConfig)
	if err != nil {
		return nil, err
	}

	if afterConnect != nil {
		err = afterConnect(conn)
		if err != nil {
			return nil, err
		}
	}

	c := &Conn{conn: conn, driver: d, connConfig: connConfig}
	return c, nil
}

type DriverConfig struct {
	pgx.ConnConfig
	AfterConnect func(*pgx.Conn) error // function to call on every new connection
	driver       *Driver
	id           int64
}

// ConnectionString encodes the DriverConfig into the original connection
// string. DriverConfig must be registered before calling ConnectionString.
func (c *DriverConfig) ConnectionString(original string) string {
	if c.driver == nil {
		panic("DriverConfig must be registered before calling ConnectionString")
	}

	buf := make([]byte, 9)
	binary.BigEndian.PutUint64(buf[1:], uint64(c.id))
	buf = append(buf, original...)
	return string(buf)
}

func (d *Driver) registerDriverConfig(c *DriverConfig) {
	d.configMutex.Lock()

	c.driver = d
	c.id = d.configCount
	d.configs[d.configCount] = c
	d.configCount++

	d.configMutex.Unlock()
}

func (d *Driver) unregisterDriverConfig(c *DriverConfig) {
	d.configMutex.Lock()
	delete(d.configs, c.id)
	d.configMutex.Unlock()
}

// RegisterDriverConfig registers a DriverConfig for use with Open.
func RegisterDriverConfig(c *DriverConfig) {
	pgxDriver.registerDriverConfig(c)
}

// UnregisterDriverConfig removes a DriverConfig registration.
func UnregisterDriverConfig(c *DriverConfig) {
	pgxDriver.unregisterDriverConfig(c)
}

type Conn struct {
	conn       *pgx.Conn
	psCount    int64 // Counter used for creating unique prepared statement names
	driver     *Driver
	connConfig pgx.ConnConfig
}

func (c *Conn) Prepare(query string) (driver.Stmt, error) {
	return c.PrepareContext(context.Background(), query)
}

func (c *Conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if !c.conn.IsAlive() {
		return nil, driver.ErrBadConn
	}

	name := fmt.Sprintf("pgx_%d", c.psCount)
	c.psCount++

	ps, err := c.conn.PrepareEx(ctx, name, query, nil)
	if err != nil {
		return nil, err
	}

	restrictBinaryToDatabaseSqlTypes(ps)

	return &Stmt{ps: ps, conn: c}, nil
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *Conn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if !c.conn.IsAlive() {
		return nil, driver.ErrBadConn
	}

	if pconn, ok := ctx.Value(ctxKeyFakeTx).(**pgx.Conn); ok {
		*pconn = c.conn
		return fakeTx{}, nil
	}

	var pgxOpts pgx.TxOptions
	switch sql.IsolationLevel(opts.Isolation) {
	case sql.LevelDefault:
	case sql.LevelReadUncommitted:
		pgxOpts.IsoLevel = pgx.ReadUncommitted
	case sql.LevelReadCommitted:
		pgxOpts.IsoLevel = pgx.ReadCommitted
	case sql.LevelSnapshot:
		pgxOpts.IsoLevel = pgx.RepeatableRead
	case sql.LevelSerializable:
		pgxOpts.IsoLevel = pgx.Serializable
	default:
		return nil, errors.Errorf("unsupported isolation: %v", opts.Isolation)
	}

	if opts.ReadOnly {
		pgxOpts.AccessMode = pgx.ReadOnly
	}

	return c.conn.BeginEx(ctx, &pgxOpts)
}

func (c *Conn) Exec(query string, argsV []driver.Value) (driver.Result, error) {
	if !c.conn.IsAlive() {
		return nil, driver.ErrBadConn
	}

	args := valueToInterface(argsV)
	commandTag, err := c.conn.Exec(query, args...)
	return driver.RowsAffected(commandTag.RowsAffected()), err
}

func (c *Conn) ExecContext(ctx context.Context, query string, argsV []driver.NamedValue) (driver.Result, error) {
	if !c.conn.IsAlive() {
		return nil, driver.ErrBadConn
	}

	args := namedValueToInterface(argsV)

	commandTag, err := c.conn.ExecEx(ctx, query, nil, args...)
	return driver.RowsAffected(commandTag.RowsAffected()), err
}

func (c *Conn) Query(query string, argsV []driver.Value) (driver.Rows, error) {
	if !c.conn.IsAlive() {
		return nil, driver.ErrBadConn
	}

	if !c.connConfig.PreferSimpleProtocol {
		ps, err := c.conn.Prepare("", query)
		if err != nil {
			return nil, err
		}

		restrictBinaryToDatabaseSqlTypes(ps)
		return c.queryPrepared("", argsV)
	}

	rows, err := c.conn.Query(query, valueToInterface(argsV)...)
	if err != nil {
		return nil, err
	}

	// Preload first row because otherwise we won't know what columns are available when database/sql asks.
	more := rows.Next()
	return &Rows{rows: rows, skipNext: true, skipNextMore: more}, nil
}

func (c *Conn) QueryContext(ctx context.Context, query string, argsV []driver.NamedValue) (driver.Rows, error) {
	if !c.conn.IsAlive() {
		return nil, driver.ErrBadConn
	}

	if !c.connConfig.PreferSimpleProtocol {
		ps, err := c.conn.PrepareEx(ctx, "", query, nil)
		if err != nil {
			return nil, err
		}

		restrictBinaryToDatabaseSqlTypes(ps)
		return c.queryPreparedContext(ctx, "", argsV)
	}

	rows, err := c.conn.QueryEx(ctx, query, nil, namedValueToInterface(argsV)...)
	if err != nil {
		return nil, err
	}

	// Preload first row because otherwise we won't know what columns are available when database/sql asks.
	more := rows.Next()
	return &Rows{rows: rows, skipNext: true, skipNextMore: more}, nil
}

func (c *Conn) queryPrepared(name string, argsV []driver.Value) (driver.Rows, error) {
	if !c.conn.IsAlive() {
		return nil, driver.ErrBadConn
	}

	args := valueToInterface(argsV)

	rows, err := c.conn.Query(name, args...)
	if err != nil {
		return nil, err
	}

	return &Rows{rows: rows}, nil
}

func (c *Conn) queryPreparedContext(ctx context.Context, name string, argsV []driver.NamedValue) (driver.Rows, error) {
	if !c.conn.IsAlive() {
		return nil, driver.ErrBadConn
	}

	args := namedValueToInterface(argsV)

	rows, err := c.conn.QueryEx(ctx, name, nil, args...)
	if err != nil {
		return nil, err
	}

	return &Rows{rows: rows}, nil
}

func (c *Conn) Ping(ctx context.Context) error {
	if !c.conn.IsAlive() {
		return driver.ErrBadConn
	}

	return c.conn.Ping(ctx)
}

// Anything that isn't a database/sql compatible type needs to be forced to
// text format so that pgx.Rows.Values doesn't decode it into a native type
// (e.g. []int32)
func restrictBinaryToDatabaseSqlTypes(ps *pgx.PreparedStatement) {
	for i := range ps.FieldDescriptions {
		intrinsic, _ := databaseSqlOIDs[ps.FieldDescriptions[i].DataType]
		if !intrinsic {
			ps.FieldDescriptions[i].FormatCode = pgx.TextFormatCode
		}
	}
}

type Stmt struct {
	ps   *pgx.PreparedStatement
	conn *Conn
}

func (s *Stmt) Close() error {
	return s.conn.conn.Deallocate(s.ps.Name)
}

func (s *Stmt) NumInput() int {
	return len(s.ps.ParameterOIDs)
}

func (s *Stmt) Exec(argsV []driver.Value) (driver.Result, error) {
	return s.conn.Exec(s.ps.Name, argsV)
}

func (s *Stmt) ExecContext(ctx context.Context, argsV []driver.NamedValue) (driver.Result, error) {
	return s.conn.ExecContext(ctx, s.ps.Name, argsV)
}

func (s *Stmt) Query(argsV []driver.Value) (driver.Rows, error) {
	return s.conn.queryPrepared(s.ps.Name, argsV)
}

func (s *Stmt) QueryContext(ctx context.Context, argsV []driver.NamedValue) (driver.Rows, error) {
	return s.conn.queryPreparedContext(ctx, s.ps.Name, argsV)
}

type Rows struct {
	rows         *pgx.Rows
	values       []interface{}
	skipNext     bool
	skipNextMore bool
}

func (r *Rows) Columns() []string {
	fieldDescriptions := r.rows.FieldDescriptions()
	names := make([]string, 0, len(fieldDescriptions))
	for _, fd := range fieldDescriptions {
		names = append(names, fd.Name)
	}
	return names
}

// ColumnTypeDatabaseTypeName return the database system type name.
func (r *Rows) ColumnTypeDatabaseTypeName(index int) string {
	return strings.ToUpper(r.rows.FieldDescriptions()[index].DataTypeName)
}

// ColumnTypeLength returns the length of the column type if the column is a
// variable length type. If the column is not a variable length type ok
// should return false.
func (r *Rows) ColumnTypeLength(index int) (int64, bool) {
	return r.rows.FieldDescriptions()[index].Length()
}

// ColumnTypePrecisionScale should return the precision and scale for decimal
// types. If not applicable, ok should be false.
func (r *Rows) ColumnTypePrecisionScale(index int) (precision, scale int64, ok bool) {
	return r.rows.FieldDescriptions()[index].PrecisionScale()
}

// ColumnTypeScanType returns the value type that can be used to scan types into.
func (r *Rows) ColumnTypeScanType(index int) reflect.Type {
	return r.rows.FieldDescriptions()[index].Type()
}

func (r *Rows) Close() error {
	r.rows.Close()
	return nil
}

func (r *Rows) Next(dest []driver.Value) error {
	if r.values == nil {
		r.values = make([]interface{}, len(r.rows.FieldDescriptions()))
		for i, fd := range r.rows.FieldDescriptions() {
			switch fd.DataType {
			case pgtype.BoolOID:
				r.values[i] = &pgtype.Bool{}
			case pgtype.ByteaOID:
				r.values[i] = &pgtype.Bytea{}
			case pgtype.CIDOID:
				r.values[i] = &pgtype.CID{}
			case pgtype.DateOID:
				r.values[i] = &pgtype.Date{}
			case pgtype.Float4OID:
				r.values[i] = &pgtype.Float4{}
			case pgtype.Float8OID:
				r.values[i] = &pgtype.Float8{}
			case pgtype.Int2OID:
				r.values[i] = &pgtype.Int2{}
			case pgtype.Int4OID:
				r.values[i] = &pgtype.Int4{}
			case pgtype.Int8OID:
				r.values[i] = &pgtype.Int8{}
			case pgtype.OIDOID:
				r.values[i] = &pgtype.OIDValue{}
			case pgtype.TimestampOID:
				r.values[i] = &pgtype.Timestamp{}
			case pgtype.TimestamptzOID:
				r.values[i] = &pgtype.Timestamptz{}
			case pgtype.XIDOID:
				r.values[i] = &pgtype.XID{}
			default:
				r.values[i] = &pgtype.GenericText{}
			}
		}
	}

	var more bool
	if r.skipNext {
		more = r.skipNextMore
		r.skipNext = false
	} else {
		more = r.rows.Next()
	}

	if !more {
		if r.rows.Err() == nil {
			return io.EOF
		} else {
			return r.rows.Err()
		}
	}

	err := r.rows.Scan(r.values...)
	if err != nil {
		return err
	}

	for i, v := range r.values {
		dest[i], err = v.(driver.Valuer).Value()
		if err != nil {
			return err
		}
	}

	return nil
}

func valueToInterface(argsV []driver.Value) []interface{} {
	args := make([]interface{}, 0, len(argsV))
	for _, v := range argsV {
		if v != nil {
			args = append(args, v.(interface{}))
		} else {
			args = append(args, nil)
		}
	}
	return args
}

func namedValueToInterface(argsV []driver.NamedValue) []interface{} {
	args := make([]interface{}, 0, len(argsV))
	for _, v := range argsV {
		if v.Value != nil {
			args = append(args, v.Value.(interface{}))
		} else {
			args = append(args, nil)
		}
	}
	return args
}

type fakeTx struct{}

func (fakeTx) Commit() error { return nil }

func (fakeTx) Rollback() error { return nil }

func AcquireConn(db *sql.DB) (*pgx.Conn, error) {
	driver, ok := db.Driver().(*Driver)
	if !ok {
		return nil, ErrNotPgx
	}

	var conn *pgx.Conn
	ctx := context.WithValue(context.Background(), ctxKeyFakeTx, &conn)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	driver.fakeTxMutex.Lock()
	driver.fakeTxConns[conn] = tx
	driver.fakeTxMutex.Unlock()

	return conn, nil
}

func ReleaseConn(db *sql.DB, conn *pgx.Conn) error {
	var tx *sql.Tx
	var ok bool

	driver := db.Driver().(*Driver)
	driver.fakeTxMutex.Lock()
	tx, ok = driver.fakeTxConns[conn]
	if ok {
		delete(driver.fakeTxConns, conn)
		driver.fakeTxMutex.Unlock()
	} else {
		driver.fakeTxMutex.Unlock()
		return errors.Errorf("can't release conn that is not acquired")
	}

	return tx.Rollback()
}
