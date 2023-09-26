//go:generate ../../../tools/readme_config_includer/generator
package ldap

import (
	"crypto/tls"
	_ "embed"
	"fmt"
	"net/url"
	"time"

	"github.com/go-ldap/ldap/v3"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	commontls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type LDAP struct {
	Server            string        `toml:"server"`
	Dialect           string        `toml:"dialect"`
	BindDn            string        `toml:"bind_dn"`
	BindPassword      config.Secret `toml:"bind_password"`
	ReverseFieldNames bool          `toml:"reverse_field_names"`
	commontls.ClientConfig

	tlsCfg   *tls.Config
	requests []request
	mode     string
	host     string
	port     string
}

type request struct {
	query   *ldap.SearchRequest
	convert func(*ldap.SearchResult, time.Time) []telegraf.Metric
}

func (*LDAP) SampleConfig() string {
	return sampleConfig
}

func (l *LDAP) Init() error {
	if l.Server == "" {
		l.Server = "ldap://localhost:389"
	}

	u, err := url.Parse(l.Server)
	if err != nil {
		return fmt.Errorf("parsing server failed: %w", err)
	}

	// Verify the server setting and set the defaults
	var tlsEnable bool
	switch u.Scheme {
	case "ldap":
		if u.Port() == "" {
			u.Host = u.Host + ":389"
		}
		tlsEnable = false
	case "starttls":
		if u.Port() == "" {
			u.Host = u.Host + ":389"
		}
		tlsEnable = true
	case "ldaps":
		if u.Port() == "" {
			u.Host = u.Host + ":636"
		}
		tlsEnable = true
	default:
		return fmt.Errorf("invalid scheme: %q", u.Scheme)
	}
	l.mode = u.Scheme
	l.Server = u.Host
	l.host, l.port = u.Hostname(), u.Port()

	// Force TLS depending on the selected mode
	l.ClientConfig.Enable = &tlsEnable

	// Setup TLS configuration
	tlsCfg, err := l.ClientConfig.TLSConfig()
	if err != nil {
		return fmt.Errorf("creating TLS config failed: %w", err)
	}
	l.tlsCfg = tlsCfg

	// Initialize the search request(s)
	switch l.Dialect {
	case "", "openldap":
		l.requests = l.newOpenLDAPConfig()
	case "389ds":
		l.requests = l.new389dsConfig()
	default:
		return fmt.Errorf("invalid dialect %q", l.Dialect)
	}

	return nil
}

func (l *LDAP) Gather(acc telegraf.Accumulator) error {
	// Connect
	conn, err := l.connect()
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()

	// Query the server
	for _, req := range l.requests {
		now := time.Now()
		result, err := conn.Search(req.query)
		if err != nil {
			acc.AddError(err)
			continue
		}

		// Collect metrics
		for _, m := range req.convert(result, now) {
			acc.AddMetric(m)
		}
	}

	return nil
}

func (l *LDAP) connect() (*ldap.Conn, error) {
	var conn *ldap.Conn
	switch l.mode {
	case "ldap":
		var err error
		conn, err = ldap.Dial("tcp", l.Server)
		if err != nil {
			return nil, err
		}
	case "ldaps":
		var err error
		conn, err = ldap.DialTLS("tcp", l.Server, l.tlsCfg)
		if err != nil {
			return nil, err
		}
	case "starttls":
		var err error
		conn, err = ldap.Dial("tcp", l.Server)
		if err != nil {
			return nil, err
		}
		if err := conn.StartTLS(l.tlsCfg); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid tls_mode: %s", l.mode)
	}

	if l.BindDn == "" && l.BindPassword.Empty() {
		return conn, nil
	}

	// Bind username and password
	passwd, err := l.BindPassword.Get()
	if err != nil {
		return nil, fmt.Errorf("getting password failed: %w", err)
	}
	defer passwd.Destroy()

	if err := conn.Bind(l.BindDn, passwd.String()); err != nil {
		return nil, fmt.Errorf("binding credentials failed: %w", err)
	}

	return conn, nil
}

func init() {
	inputs.Add("ldap", func() telegraf.Input { return &LDAP{} })
}
