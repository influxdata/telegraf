package openldap

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/ldap.v2"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Openldap struct {
	Host               string
	Port               int
	Ssl                string
	InsecureSkipVerify bool
	SslCa              string
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
  ssl = ""

  # skip peer certificate verification. Default is false.
  insecure_skip_verify = false

  # Path to PEM-encoded Root certificate to use to verify server certificate
  ssl_ca = "/etc/ssl/certs.pem"

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
		Ssl:                "",
		InsecureSkipVerify: false,
		SslCa:              "",
		BindDn:             "",
		BindPassword:       "",
		ReverseMetricNames: false,
	}
}

// gather metrics
func (o *Openldap) Gather(acc telegraf.Accumulator) error {
	var err error
	var l *ldap.Conn
	if o.Ssl != "" {
		// build tls config
		tlsConfig, err := internal.GetTLSConfig("", "", o.SslCa, o.InsecureSkipVerify)
		if err != nil {
			acc.AddError(err)
			return nil
		}
		if o.Ssl == "ldaps" {
			l, err = ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", o.Host, o.Port), tlsConfig)
			if err != nil {
				acc.AddError(err)
				return nil
			}
		} else if o.Ssl == "starttls" {
			l, err = ldap.Dial("tcp", fmt.Sprintf("%s:%d", o.Host, o.Port))
			if err != nil {
				acc.AddError(err)
				return nil
			}
			err = l.StartTLS(tlsConfig)
		} else {
			acc.AddError(fmt.Errorf("Invalid setting for ssl: %s", o.Ssl))
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
