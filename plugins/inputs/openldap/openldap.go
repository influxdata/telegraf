package openldap

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"gopkg.in/ldap.v3"
)

type Openldap struct {
	Host               string
	Port               int
	SSL                string `toml:"ssl"` // Deprecated in 1.7; use TLS
	TLS                string `toml:"tls"`
	InsecureSkipVerify bool
	SSLCA              string `toml:"ssl_ca"` // Deprecated in 1.7; use TLSCA
	TLSCA              string `toml:"tls_ca"`
	BindDn             string
	BindPassword       string
	ReverseMetricNames bool
}

const sampleConfig string = `
  host = "localhost"
  port = 389

  # ldaps, starttls, or no encryption. default is an empty string, disabling all encryption.
  # note that port will likely need to be changed to 636 for ldaps
  # valid options: "" | "starttls" | "ldaps"
  tls = ""

  # skip peer certificate verification. Default is false.
  insecure_skip_verify = false

  # Path to PEM-encoded Root certificate to use to verify server certificate
  tls_ca = "/etc/ssl/certs.pem"

  # dn/password to bind with. If bind_dn is empty, an anonymous bind is performed.
  bind_dn = ""
  bind_password = ""

  # Reverse metric names so they sort more naturally. Recommended.
  # This defaults to false if unset, but is set to true when generating a new config
  reverse_metric_names = true
`

var searchBase = "cn=Monitor"
var searchFilter = "(|(objectClass=monitorCounterObject)(objectClass=monitorOperation)(objectClass=monitoredObject))"
var searchAttrs = []string{"monitorCounter", "monitorOpInitiated", "monitorOpCompleted", "monitoredInfo"}
var attrTranslate = map[string]string{
	"monitorCounter":     "",
	"monitoredInfo":      "",
	"monitorOpInitiated": "_initiated",
	"monitorOpCompleted": "_completed",
	"olmMDBPagesMax":     "_mdb_pages_max",
	"olmMDBPagesUsed":    "_mdb_pages_used",
	"olmMDBPagesFree":    "_mdb_pages_free",
	"olmMDBReadersMax":   "_mdb_readers_max",
	"olmMDBReadersUsed":  "_mdb_readers_used",
	"olmMDBEntries":      "_mdb_entries",
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
		SSL:                "",
		TLS:                "",
		InsecureSkipVerify: false,
		SSLCA:              "",
		TLSCA:              "",
		BindDn:             "",
		BindPassword:       "",
		ReverseMetricNames: false,
	}
}

// gather metrics
func (o *Openldap) Gather(acc telegraf.Accumulator) error {
	if o.TLS == "" {
		o.TLS = o.SSL
	}
	if o.TLSCA == "" {
		o.TLSCA = o.SSLCA
	}

	var err error
	var l *ldap.Conn
	if o.TLS != "" {
		// build tls config
		clientTLSConfig := tls.ClientConfig{
			TLSCA:              o.TLSCA,
			InsecureSkipVerify: o.InsecureSkipVerify,
		}
		tlsConfig, err := clientTLSConfig.TLSConfig()
		if err != nil {
			acc.AddError(err)
			return nil
		}
		if o.TLS == "ldaps" {
			l, err = ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", o.Host, o.Port), tlsConfig)
			if err != nil {
				acc.AddError(err)
				return nil
			}
		} else if o.TLS == "starttls" {
			l, err = ldap.Dial("tcp", fmt.Sprintf("%s:%d", o.Host, o.Port))
			if err != nil {
				acc.AddError(err)
				return nil
			}
			err = l.StartTLS(tlsConfig)
			if err != nil {
				acc.AddError(err)
				return nil
			}
		} else {
			acc.AddError(fmt.Errorf("Invalid setting for ssl: %s", o.TLS))
			return nil
		}
	} else {
		l, err = ldap.Dial("tcp", fmt.Sprintf("%s:%d", o.Host, o.Port))
	}

	if err != nil {
		acc.AddError(err)
		return nil
	}
	defer l.Close()

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
		metricName := dnToMetric(entry.DN, o)
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

// Convert a DN to metric name, eg cn=Read,cn=Waiters,cn=Monitor becomes waiters_read
// Assumes the last part of the DN is cn=Monitor and we want to drop it
func dnToMetric(dn string, o *Openldap) string {
	if o.ReverseMetricNames {
		var metricParts []string

		dn = strings.Trim(dn, " ")
		dn = strings.Replace(dn, " ", "_", -1)
		dn = strings.Replace(dn, "cn=", "", -1)
		dn = strings.ToLower(dn)
		metricParts = strings.Split(dn, ",")
		for i, j := 0, len(metricParts)-1; i < j; i, j = i+1, j-1 {
			metricParts[i], metricParts[j] = metricParts[j], metricParts[i]
		}
		return strings.Join(metricParts[1:], "_")
	} else {
		metricName := strings.Trim(dn, " ")
		metricName = strings.Replace(metricName, " ", "_", -1)
		metricName = strings.ToLower(metricName)
		metricName = strings.TrimPrefix(metricName, "cn=")
		metricName = strings.Replace(metricName, strings.ToLower("cn=Monitor"), "", -1)
		metricName = strings.Replace(metricName, "cn=", "_", -1)
		return strings.Replace(metricName, ",", "", -1)
	}
}

func init() {
	inputs.Add("openldap", func() telegraf.Input { return NewOpenldap() })
}
