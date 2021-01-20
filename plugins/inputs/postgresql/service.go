package postgresql

import (
	"database/sql"
	"fmt"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/stdlib"
	"net"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
)

// pulled from lib/pq
// ParseURL no longer needs to be used by clients of this library since supplying a URL as a
// connection string to sql.Open() is now supported:
//
//	sql.Open("postgres", "postgres://bob:secret@1.2.3.4:5432/mydb?sslmode=verify-full")
//
// It remains exported here for backwards-compatibility.
//
// ParseURL converts a url to a connection string for driver.Open.
// Example:
//
//	"postgres://bob:secret@1.2.3.4:5432/mydb?sslmode=verify-full"
//
// converts to:
//
//	"user=bob password=secret host=1.2.3.4 port=5432 dbname=mydb sslmode=verify-full"
//
// A minimal example:
//
//	"postgres://"
//
// This will be blank, causing driver.Open to use all of the defaults
func parseURL(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return "", fmt.Errorf("invalid connection protocol: %s", u.Scheme)
	}

	var kvs []string
	escaper := strings.NewReplacer(` `, `\ `, `'`, `\'`, `\`, `\\`)
	accrue := func(k, v string) {
		if v != "" {
			kvs = append(kvs, k+"="+escaper.Replace(v))
		}
	}

	if u.User != nil {
		v := u.User.Username()
		accrue("user", v)

		v, _ = u.User.Password()
		accrue("password", v)
	}

	if host, port, err := net.SplitHostPort(u.Host); err != nil {
		accrue("host", u.Host)
	} else {
		accrue("host", host)
		accrue("port", port)
	}

	if u.Path != "" {
		accrue("dbname", u.Path[1:])
	}

	q := u.Query()
	for k := range q {
		accrue(k, q.Get(k))
	}

	sort.Strings(kvs) // Makes testing easier (not a performance concern)
	return strings.Join(kvs, " "), nil
}

// Service common functionality shared between the postgresql and postgresql_extensible
// packages.
type Service struct {
	Address       string
	Outputaddress string
	MaxIdle       int
	MaxOpen       int
	MaxLifetime   internal.Duration
	DB            *sql.DB
	IsPgBouncer   bool
}

// Start starts the ServiceInput's service, whatever that may be
func (p *Service) Start(telegraf.Accumulator) (err error) {
	const localhost = "host=localhost sslmode=disable"

	if p.Address == "" || p.Address == "localhost" {
		p.Address = localhost
	}

	connectionString := p.Address

	// Specific support to make it work with PgBouncer too
	// See https://github.com/influxdata/telegraf/issues/3253#issuecomment-357505343
	if p.IsPgBouncer {
		d := &stdlib.DriverConfig{
			ConnConfig: pgx.ConnConfig{
				PreferSimpleProtocol: true,
				RuntimeParams: map[string]string{
					"client_encoding": "UTF8",
				},
				CustomConnInfo: func(c *pgx.Conn) (*pgtype.ConnInfo, error) {
					info := c.ConnInfo.DeepCopy()
					info.RegisterDataType(pgtype.DataType{
						Value: &pgtype.OIDValue{},
						Name:  "int8OID",
						OID:   pgtype.Int8OID,
					})
					// Newer versions of pgbouncer need this defined. See the discussion here:
					// https://github.com/jackc/pgx/issues/649
					info.RegisterDataType(pgtype.DataType{
						Value: &pgtype.OIDValue{},
						Name:  "numericOID",
						OID:   pgtype.NumericOID,
					})

					return info, nil
				},
			},
		}
		stdlib.RegisterDriverConfig(d)
		connectionString = d.ConnectionString(p.Address)
	}

	if p.DB, err = sql.Open("pgx", connectionString); err != nil {
		return err
	}

	p.DB.SetMaxOpenConns(p.MaxOpen)
	p.DB.SetMaxIdleConns(p.MaxIdle)
	p.DB.SetConnMaxLifetime(p.MaxLifetime.Duration)

	return nil
}

// Stop stops the services and closes any necessary channels and connections
func (p *Service) Stop() {
	p.DB.Close()
}

var kvMatcher, _ = regexp.Compile("(password|sslcert|sslkey|sslmode|sslrootcert)=\\S+ ?")

// SanitizedAddress utility function to strip sensitive information from the connection string.
func (p *Service) SanitizedAddress() (sanitizedAddress string, err error) {
	var (
		canonicalizedAddress string
	)

	if p.Outputaddress != "" {
		return p.Outputaddress, nil
	}

	if strings.HasPrefix(p.Address, "postgres://") || strings.HasPrefix(p.Address, "postgresql://") {
		if canonicalizedAddress, err = parseURL(p.Address); err != nil {
			return sanitizedAddress, err
		}
	} else {
		canonicalizedAddress = p.Address
	}

	sanitizedAddress = kvMatcher.ReplaceAllString(canonicalizedAddress, "")

	return sanitizedAddress, err
}
