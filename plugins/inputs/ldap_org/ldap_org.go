package ldap_org

import (
	"fmt"
	"strconv"

	"gopkg.in/ldap.v2"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
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
	SearchBase	   string `toml:"searchBase"`
	RetAttr		   string
	Filter		   string
}

const sampleConfig string = `
  # This is an high load plugin. Tipically once a day run is sufficient.
  interval = "24h"

  # LDAP Host and post to query
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

  # Where to count metrics
  # For instance ou=<metric_name>,o=myorg,c=en
  # In searchBase look for "retAttr=*", then for each DN look for Filter and count results.
  searchBase = "o=myorg,c=en"
  retAttr = "ou"
  filter = "(objectClass=*)"
`

func (o *Openldap) SampleConfig() string {
	return sampleConfig
}

func (o *Openldap) Description() string {
	return "LDAP Count by Org plugin"
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
        	SearchBase:         "",
        	RetAttr:            "",
		Filter:		    "",
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

	var searchFilter = fmt.Sprintf("(%s=*)", o.RetAttr)
	var searchAttrs = []string{ o.RetAttr }
	metrics := make(map[string]int)

	searchRequest := ldap.NewSearchRequest(
		o.SearchBase,
		ldap.ScopeSingleLevel,
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

	for _, entry := range sr.Entries {
		metric_name := entry.GetAttributeValue(o.RetAttr)
		secondSearchBase := entry.DN
		secondSearchRequest := ldap.NewSearchRequest(
			secondSearchBase,
			ldap.ScopeWholeSubtree,
			ldap.NeverDerefAliases,
			0,
			0,
			false,
			o.Filter,
			[]string{ "dn" },
			nil,
		)
		
		second_sr, err := l.Search(secondSearchRequest)
		if err != nil {
			acc.AddError(err)
			return nil
		}
		
		metrics[metric_name] = len(second_sr.Entries)
	}

	gatherSearchResult(metrics, o, acc)
	return nil
}


func gatherSearchResult(measures map[string]int, o *Openldap, acc telegraf.Accumulator) {
	fields := map[string]interface{}{}
	tags := map[string]string{
		"server": o.Host,
		"port":   strconv.Itoa(o.Port),
		"base":   o.SearchBase,
	}

	for k, v := range measures {
		fields[k] = v
	}
	acc.AddFields("ldap_org", fields, tags)
	return
}


func init() {
	inputs.Add("ldap_org", func() telegraf.Input { return NewOpenldap() })
}
