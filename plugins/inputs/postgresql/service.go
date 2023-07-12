package postgresql

import (
	"bytes"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
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
	Address       config.Secret   `toml:"address"`
	OutputAddress string          `toml:"outputaddress"`
	MaxIdle       int             `toml:"max_idle"`
	MaxOpen       int             `toml:"max_open"`
	MaxLifetime   config.Duration `toml:"max_lifetime"`
	IsPgBouncer   bool            `toml:"-"`
	DB            *sql.DB
}

var socketRegexp = regexp.MustCompile(`/\.s\.PGSQL\.\d+$`)

// Start starts the ServiceInput's service, whatever that may be
func (p *Service) Start(telegraf.Accumulator) (err error) {
	addrSecret, err := p.Address.Get()
	if err != nil {
		return fmt.Errorf("getting address failed: %w", err)
	}
	addr := string(addrSecret)
	defer config.ReleaseSecret(addrSecret)

	if p.Address.Empty() || addr == "localhost" {
		addr = "host=localhost sslmode=disable"
		if err := p.Address.Set([]byte(addr)); err != nil {
			return err
		}
	}

	connConfig, err := pgx.ParseConfig(addr)
	if err != nil {
		return err
	}

	// Remove the socket name from the path
	connConfig.Host = socketRegexp.ReplaceAllLiteralString(connConfig.Host, "")

	// Specific support to make it work with PgBouncer too
	// See https://github.com/influxdata/telegraf/issues/3253#issuecomment-357505343
	if p.IsPgBouncer {
		// Remove DriveConfig and revert it by the ParseConfig method
		// See https://github.com/influxdata/telegraf/issues/9134
		connConfig.PreferSimpleProtocol = true
	}

	connectionString := stdlib.RegisterConnConfig(connConfig)
	if p.DB, err = sql.Open("pgx", connectionString); err != nil {
		return err
	}

	p.DB.SetMaxOpenConns(p.MaxOpen)
	p.DB.SetMaxIdleConns(p.MaxIdle)
	p.DB.SetConnMaxLifetime(time.Duration(p.MaxLifetime))

	return nil
}

// Stop stops the services and closes any necessary channels and connections
func (p *Service) Stop() {
	p.DB.Close()
}

var kvMatcher, _ = regexp.Compile(`(password|sslcert|sslkey|sslmode|sslrootcert)=\S+ ?`)

// SanitizedAddress utility function to strip sensitive information from the connection string.
func (p *Service) SanitizedAddress() (sanitizedAddress string, err error) {
	if p.OutputAddress != "" {
		return p.OutputAddress, nil
	}

	addr, err := p.Address.Get()
	if err != nil {
		return sanitizedAddress, fmt.Errorf("getting address for sanitization failed: %w", err)
	}
	defer config.ReleaseSecret(addr)

	var canonicalizedAddress string
	if bytes.HasPrefix(addr, []byte("postgres://")) || bytes.HasPrefix(addr, []byte("postgresql://")) {
		if canonicalizedAddress, err = parseURL(string(addr)); err != nil {
			return sanitizedAddress, err
		}
	} else {
		canonicalizedAddress = string(addr)
	}

	return kvMatcher.ReplaceAllString(canonicalizedAddress, ""), nil
}

// GetConnectDatabase utility function for getting the database to which the connection was made
func (p *Service) GetConnectDatabase(connectionString string) (string, error) {
	connConfig, err := pgx.ParseConfig(connectionString)
	if err != nil {
		return "", fmt.Errorf("connection string parsing failed: %w", err)
	}

	if len(connConfig.Database) != 0 {
		return connConfig.Database, nil
	}

	return "postgres", nil
}
