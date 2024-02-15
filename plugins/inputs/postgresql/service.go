package postgresql

import (
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

// Based on parseURLSettings() at https://github.com/jackc/pgx/blob/master/pgconn/config.go
func toKeyValue(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", fmt.Errorf("parsing URI failed: %w", err)
	}

	// Check the protocol
	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return "", fmt.Errorf("invalid connection protocol: %s", u.Scheme)
	}

	quoteIfNecessary := func(v string) string {
		if !strings.ContainsAny(v, ` ='\`) {
			return v
		}
		r := strings.ReplaceAll(v, `\`, `\\`)
		r = strings.ReplaceAll(r, `'`, `\'`)
		return "'" + r + "'"
	}

	// Extract the parameters
	parts := make([]string, 0, len(u.Query())+5)
	if u.User != nil {
		parts = append(parts, "user="+quoteIfNecessary(u.User.Username()))
		if password, found := u.User.Password(); found {
			parts = append(parts, "password="+quoteIfNecessary(password))
		}
	}

	// Handle multiple host:port's in url.Host by splitting them into host,host,host and port,port,port.
	hostParts := strings.Split(u.Host, ",")
	hosts := make([]string, 0, len(hostParts))
	ports := make([]string, 0, len(hostParts))
	var anyPortSet bool
	for _, host := range hostParts {
		if host == "" {
			continue
		}

		h, p, err := net.SplitHostPort(host)
		if err != nil {
			if !strings.Contains(err.Error(), "missing port") {
				return "", fmt.Errorf("failed to process host %q: %w", host, err)
			}
			h = host
		}
		anyPortSet = anyPortSet || err == nil
		hosts = append(hosts, h)
		ports = append(ports, p)
	}
	if len(hosts) > 0 {
		parts = append(parts, "host="+strings.Join(hosts, ","))
	}
	if anyPortSet {
		parts = append(parts, "port="+strings.Join(ports, ","))
	}

	database := strings.TrimLeft(u.Path, "/")
	if database != "" {
		parts = append(parts, "dbname="+quoteIfNecessary(database))
	}

	for k, v := range u.Query() {
		parts = append(parts, k+"="+quoteIfNecessary(strings.Join(v, ",")))
	}

	// Required to produce a repeatable output e.g. for tags or testing
	sort.Strings(parts)
	return strings.Join(parts, " "), nil
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
	addr := addrSecret.String()
	defer addrSecret.Destroy()

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

var sanitizer = regexp.MustCompile(`(\s|^)((?:password|sslcert|sslkey|sslmode|sslrootcert)\s?=\s?(?:(?:'(?:[^'\\]|\\.)*')|(?:\S+)))`)

// SanitizedAddress utility function to strip sensitive information from the connection string.
func (p *Service) SanitizedAddress() (string, error) {
	if p.OutputAddress != "" {
		return p.OutputAddress, nil
	}

	// Get the address
	addrSecret, err := p.Address.Get()
	if err != nil {
		return "", fmt.Errorf("getting address for sanitization failed: %w", err)
	}
	addr := addrSecret.String()
	addrSecret.Destroy()

	// Make sure we convert URI-formatted strings into key-values
	if strings.HasPrefix(addr, "postgres://") || strings.HasPrefix(addr, "postgresql://") {
		if addr, err = toKeyValue(addr); err != nil {
			return "", err
		}
	}

	// Sanitize the string using a regular expression
	sanitized := sanitizer.ReplaceAllString(addr, "")
	return strings.TrimSpace(sanitized), nil
}

// GetConnectDatabase utility function for getting the database to which the connection was made
// If the user set the output address use that before parsing anything else.
func (p *Service) GetConnectDatabase(connectionString string) (string, error) {
	connConfig, err := pgx.ParseConfig(connectionString)
	if err == nil && len(connConfig.Database) != 0 {
		return connConfig.Database, nil
	}

	return "postgres", nil
}
