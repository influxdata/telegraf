package postgresql

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"

	"github.com/influxdata/telegraf/config"
)

var socketRegexp = regexp.MustCompile(`/\.s\.PGSQL\.\d+$`)
var sanitizer = regexp.MustCompile(`(\s|^)((?:password|sslcert|sslkey|sslmode|sslrootcert)\s?=\s?(?:(?:'(?:[^'\\]|\\.)*')|(?:\S+)))`)

type Config struct {
	Address       config.Secret   `toml:"address"`
	OutputAddress string          `toml:"outputaddress"`
	MaxIdle       int             `toml:"max_idle"`
	MaxOpen       int             `toml:"max_open"`
	MaxLifetime   config.Duration `toml:"max_lifetime"`
	IsPgBouncer   bool            `toml:"-"`
}

func (c *Config) CreateService() (*Service, error) {
	addrSecret, err := c.Address.Get()
	if err != nil {
		return nil, fmt.Errorf("getting address failed: %w", err)
	}
	addr := addrSecret.String()
	defer addrSecret.Destroy()

	if c.Address.Empty() || addr == "localhost" {
		addr = "host=localhost sslmode=disable"
		if err := c.Address.Set([]byte(addr)); err != nil {
			return nil, err
		}
	}

	connConfig, err := pgx.ParseConfig(addr)
	if err != nil {
		return nil, err
	}
	// Remove the socket name from the path
	connConfig.Host = socketRegexp.ReplaceAllLiteralString(connConfig.Host, "")

	// Specific support to make it work with PgBouncer too
	// See https://github.com/influxdata/telegraf/issues/3253#issuecomment-357505343
	if c.IsPgBouncer {
		// Remove DriveConfig and revert it by the ParseConfig method
		// See https://github.com/influxdata/telegraf/issues/9134
		connConfig.PreferSimpleProtocol = true
	}

	// Provide the connection string without sensitive information for use as
	// tag or other output properties
	sanitizedAddr, err := c.sanitizedAddress()
	if err != nil {
		return nil, err
	}

	return &Service{
		SanitizedAddress:   sanitizedAddr,
		ConnectionDatabase: connectionDatabase(sanitizedAddr),
		maxIdle:            c.MaxIdle,
		maxOpen:            c.MaxOpen,
		maxLifetime:        time.Duration(c.MaxLifetime),
		dsn:                stdlib.RegisterConnConfig(connConfig),
	}, nil
}

// connectionDatabase determines the database to which the connection was made
func connectionDatabase(sanitizedAddr string) string {
	connConfig, err := pgx.ParseConfig(sanitizedAddr)
	if err != nil || connConfig.Database == "" {
		return "postgres"
	}

	return connConfig.Database
}

// sanitizedAddress strips sensitive information from the connection string.
// If the user set the output address use that before parsing anything else.
func (c *Config) sanitizedAddress() (string, error) {
	if c.OutputAddress != "" {
		return c.OutputAddress, nil
	}

	// Get the address
	addrSecret, err := c.Address.Get()
	if err != nil {
		return "", fmt.Errorf("getting address for sanitization failed: %w", err)
	}
	defer addrSecret.Destroy()

	// Make sure we convert URI-formatted strings into key-values
	addr := addrSecret.TemporaryString()
	if strings.HasPrefix(addr, "postgres://") || strings.HasPrefix(addr, "postgresql://") {
		if addr, err = toKeyValue(addr); err != nil {
			return "", err
		}
	}

	// Sanitize the string using a regular expression
	sanitized := sanitizer.ReplaceAllString(addr, "")
	return strings.TrimSpace(sanitized), nil
}

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
