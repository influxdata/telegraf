package openldap

import (
	"fmt"
	"gopkg.in/ldap.v2"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Openldap struct {
	Host               string
	Port               int
	Ssl                bool
	InsecureSkipverify bool
	SslCA              string
	BindDn             string
	BindPassword       string
}

const sampleConfig string = `
  host = "localhost"
  port = 389
  # starttls. Default is false.
  tls = false
  # skip peer certificate verification. Default is false.
  tls_skipverify = false
  # Path to PEM-encoded Root certificate to use to verify server certificate
  tls_cacertificate = "/etc/ssl/certs.pem"

  # dn/password to bind with. If bind_dn is empty, an anonymous bind is performed.
  bind_dn = ""
  bind_password = ""
`

var searchBase = "cn=Monitor"
var searchFilter = "(|(objectClass=monitorCounterObject)(objectClass=monitorOperation))"
var searchAttrs = []string{"monitorCounter", "monitorOpInitiated", "monitorOpCompleted"}
var attrTranslate = map[string]string{
	"monitorCounter":     "",
	"monitorOpInitiated": "_initiated",
	"monitorOpCompleted": "_completed",
}

func (o *Openldap) SampleConfig() string {
	return sampleConfig
}

func (o *Openldap) Description() string {
	return "OpenLDAP cn=Monitor plugin"
}

// return an initialized Openldap
func NewOpenldap() *Openldap {
	return &Openldap{
		Host:               "localhost",
		Port:               389,
		Ssl:                false,
		InsecureSkipverify: false,
		SslCA:              "",
		BindDn:             "",
		BindPassword:       "",
	}
}

// gather metrics
func (o *Openldap) Gather(acc telegraf.Accumulator) error {
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", o.Host, o.Port))
	if err != nil {
		acc.AddError(err)
		return nil
	}
	defer l.Close()

	// TLS
	if o.Ssl {
		// build tls config
		tlsConfig, err := internal.GetTLSConfig("", "", o.SslCA, o.InsecureSkipverify)
		if err != nil {
			acc.AddError(err)
			return nil
		}

		// configure StartTLS
		err = l.StartTLS(tlsConfig)
		if err != nil {
			acc.AddError(err)
			return nil
		}
	}

	// username/password bind
	if o.BindDn != "" && o.BindPassword != "" {
		err = l.Bind(o.BindDn, o.BindPassword)
		if err != nil {
			acc.AddError(err)
			return nil
		}
	}

	searchRequest := ldap.NewSearchRequest(
		searchBase,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		searchFilter,
		searchAttrs,
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		acc.AddError(err)
		return nil
	}

	gatherSearchResult(sr, o, acc)

	return nil
}

func gatherSearchResult(sr *ldap.SearchResult, o *Openldap, acc telegraf.Accumulator) {
	fields := map[string]interface{}{}
	tags := map[string]string{
		"server": o.Host,
		"port":   strconv.Itoa(o.Port),
	}
	for _, entry := range sr.Entries {
		metricName := dnToMetric(entry.DN, searchBase)
		for _, attr := range entry.Attributes {
			if len(attr.Values[0]) >= 1 {
				if v, err := strconv.ParseInt(attr.Values[0], 10, 64); err == nil {
					fields[metricName+attrTranslate[attr.Name]] = v
				}
			}
		}
	}
	acc.AddFields("openldap", fields, tags)
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
	inputs.Add("openldap", func() telegraf.Input { return NewOpenldap() })
}
