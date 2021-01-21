package ldap_response

import (
	ctls "crypto/tls"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gopkg.in/ldap.v2"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Ldap struct {
	Host               string
	Port               int
	Ssl                string
	InsecureSkipVerify bool
	SslCa              string
	BindDn             string
	BindPassword       string
	SearchBase         string
	SearchFilter       string
	SearchAttributes   []string
}

const sampleConfig string = `
  host = "localhost"
  port = 389

  # ldaps, starttls, or no encryption. default is an empty string, disabling all encryption.
  # note that port will likely need to be changed to 636 for ldaps
  # valid options: "" | "starttls" | "ldaps"
  ssl = ""

  # skip peer certificate verification. Default is false.
  insecure_skip_verify = false

  # Path to PEM-encoded Root certificate to use to verify server certificate
  ssl_ca = "/etc/ssl/certs.pem"

  # dn/password to bind with. If bind_dn is empty, an anonymous bind is performed.
  bind_dn = ""
  bind_password = ""

  # base entry for searches
  search_base = ""

  # ldap search to perform. defaults to "(objectClass=*)" if unspecified.
  search_filter = ""

  # the attributes to return as fields. defaults to "objectclass" if unspecified.
  search_attributes = [
    "attribute1",
    "attribute2",
  ]
`

var DefaultSearchFilter = "(objectClass=*)"
var DefaultSearchAttributes = []string{"objectclass"}

func (l *Ldap) SampleConfig() string {
	return sampleConfig
}

func (l *Ldap) Description() string {
	return "LDAP Response Input Plugin"
}

// return an initialized Ldap
func NewLdap() *Ldap {
	return &Ldap{
		Host: "localhost",
		Port: 389,
	}
}

// gather metrics
func (l *Ldap) Gather(acc telegraf.Accumulator) error {
	var err error
	var server *ldap.Conn
	beforeConnect := time.Now()
	if l.Ssl != "" {
		// build tls config
		clientConfig := &tls.ClientConfig{
			TLSCA:              l.SslCa,
			InsecureSkipVerify: l.InsecureSkipVerify,
		}
		var tlsConfig *ctls.Config
		tlsConfig, err = clientConfig.TLSConfig()
		if err != nil {
			acc.AddError(err)
			return nil
		}
		if l.Ssl == "ldaps" {
			server, err = ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", l.Host, l.Port), tlsConfig)
			if err != nil {
				acc.AddError(err)
				return nil
			}
		} else if l.Ssl == "starttls" {
			server, err = ldap.Dial("tcp", fmt.Sprintf("%s:%d", l.Host, l.Port))
			if err != nil {
				acc.AddError(err)
				return nil
			}
			err = server.StartTLS(tlsConfig)
		} else {
			acc.AddError(fmt.Errorf("Invalid setting for ssl: %s", l.Ssl))
			return nil
		}
	} else {
		server, err = ldap.Dial("tcp", fmt.Sprintf("%s:%d", l.Host, l.Port))
	}
	afterConnect := time.Now()

	if err != nil {
		acc.AddError(err)
		return nil
	}
	defer server.Close()
	server.SetTimeout(300 * time.Second)

	// username/password bind
	beforeBind := time.Now()
	if l.BindDn != "" && l.BindPassword != "" {
		err = server.Bind(l.BindDn, l.BindPassword)
		if err != nil {
			acc.AddError(err)
			return nil
		}
	}
	afterBind := time.Now()

	if l.SearchFilter == "" {
		l.SearchFilter = DefaultSearchFilter
	}
	if len(l.SearchAttributes) == 0 {
		l.SearchAttributes = DefaultSearchAttributes
	}

	searchRequest := ldap.NewSearchRequest(
		l.SearchBase,
		ldap.ScopeSingleLevel,
		ldap.NeverDerefAliases,
		1000,
		60,
		false,
		l.SearchFilter,
		l.SearchAttributes,
		nil,
	)

	beforeSearch := time.Now()
	sr, err := server.Search(searchRequest)
	afterSearch := time.Now()
	if err != nil {
		acc.AddError(err)
		return nil
	}

	fields := map[string]interface{}{
		"connect_time_ms": float64(afterConnect.Sub(beforeConnect).Nanoseconds()) / 1000 / 1000,
		"bind_time_ms":    float64(afterBind.Sub(beforeBind).Nanoseconds()) / 1000 / 1000,
		"query_time_ms":   float64(afterSearch.Sub(beforeSearch).Nanoseconds()) / 1000 / 1000,
		"total_time_ms":   float64(afterSearch.Sub(beforeConnect).Nanoseconds()) / 1000 / 1000,
	}

	gatherSearchResult(fields, sr, l, acc)

	return nil
}

func gatherSearchResult(fields map[string]interface{}, sr *ldap.SearchResult, l *Ldap, acc telegraf.Accumulator) {
	tags := map[string]string{
		"server": l.Host,
		"port":   strconv.Itoa(l.Port),
	}
	for _, entry := range sr.Entries {
		metricName := dnToMetric(entry.DN, l.SearchBase)
		for _, attr := range entry.Attributes {
			if len(attr.Values[0]) >= 1 {
				if v, err := strconv.ParseInt(attr.Values[0], 10, 64); err == nil {
					fields[metricName+attr.Name] = v
				}
			}
		}
	}
	acc.AddFields("ldap_response", fields, tags)
	return
}

// Convert a DN to metric name, eg cn=Read,cn=Waiters,cn=Monitor to read_waiters
func dnToMetric(dn, searchBase string) string {
	metricName := strings.Trim(dn, " ")
	metricName = strings.Replace(metricName, " ", "_", -1)
	metricName = strings.ToLower(metricName)
	metricName = strings.TrimPrefix(metricName, "cn=")
	metricName = strings.Replace(metricName, strings.ToLower(searchBase), "", -1)
	metricName = strings.Replace(metricName, "cn=", "_", -1)
	return strings.Replace(metricName, ",", "", -1)
}

func init() {
	inputs.Add("ldap_response", func() telegraf.Input { return NewLdap() })
}
