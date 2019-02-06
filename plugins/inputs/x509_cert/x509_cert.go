// Package x509_cert reports metrics from an SSL certificate.
package x509_cert

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	_tls "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
  ## List certificate sources
  sources = ["/etc/ssl/certs/ssl-cert-snakeoil.pem", "tcp://example.org:443"]

  ## Timeout for SSL connection
  # timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`
const description = "Reads metrics from a SSL certificate"

// X509Cert holds the configuration of the plugin.
type X509Cert struct {
	Sources []string          `toml:"sources"`
	Timeout internal.Duration `toml:"timeout"`
	_tls.ClientConfig
}

// Description returns description of the plugin.
func (c *X509Cert) Description() string {
	return description
}

// SampleConfig returns configuration sample for the plugin.
func (c *X509Cert) SampleConfig() string {
	return sampleConfig
}

func (c *X509Cert) getCert(location string, timeout time.Duration) ([]*x509.Certificate, error) {
	if strings.HasPrefix(location, "/") {
		location = "file://" + location
	}

	u, err := url.Parse(location)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cert location - %s\n", err.Error())
	}

	switch u.Scheme {
	case "https":
		u.Scheme = "tcp"
		fallthrough
	case "udp", "udp4", "udp6":
		fallthrough
	case "tcp", "tcp4", "tcp6":
		tlsCfg, err := c.ClientConfig.TLSConfig()
		if err != nil {
			return nil, err
		}

		ipConn, err := net.DialTimeout(u.Scheme, u.Host, timeout)
		if err != nil {
			return nil, err
		}
		defer ipConn.Close()

		if tlsCfg == nil {
			tlsCfg = &tls.Config{}
		}
		tlsCfg.ServerName = u.Hostname()
		conn := tls.Client(ipConn, tlsCfg)
		defer conn.Close()

		hsErr := conn.Handshake()
		if hsErr != nil {
			return nil, hsErr
		}

		certs := conn.ConnectionState().PeerCertificates

		return certs, nil
	case "file":
		content, err := ioutil.ReadFile(u.Path)
		if err != nil {
			return nil, err
		}

		block, _ := pem.Decode(content)
		if block == nil {
			return nil, fmt.Errorf("failed to parse certificate PEM")
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}

		return []*x509.Certificate{cert}, nil
	default:
		return nil, fmt.Errorf("unsuported scheme '%s' in location %s\n", u.Scheme, location)
	}
}

func getFields(cert *x509.Certificate, now time.Time) map[string]interface{} {
	age := int(now.Sub(cert.NotBefore).Seconds())
	expiry := int(cert.NotAfter.Sub(now).Seconds())
	startdate := cert.NotBefore.Unix()
	enddate := cert.NotAfter.Unix()

	fields := map[string]interface{}{
		"age":       age,
		"expiry":    expiry,
		"startdate": startdate,
		"enddate":   enddate,
	}

	return fields
}

func getTags(subject pkix.Name, location string) map[string]string {
	tags := map[string]string{
		"source":      location,
		"common_name": subject.CommonName,
	}

	if len(subject.Organization) > 0 {
		tags["organization"] = subject.Organization[0]
	}
	if len(subject.OrganizationalUnit) > 0 {
		tags["organizational_unit"] = subject.OrganizationalUnit[0]
	}
	if len(subject.Country) > 0 {
		tags["country"] = subject.Country[0]
	}
	if len(subject.Province) > 0 {
		tags["province"] = subject.Province[0]
	}
	if len(subject.Locality) > 0 {
		tags["locality"] = subject.Locality[0]
	}

	return tags
}

// Gather adds metrics into the accumulator.
func (c *X509Cert) Gather(acc telegraf.Accumulator) error {
	now := time.Now()

	for _, location := range c.Sources {
		certs, err := c.getCert(location, c.Timeout.Duration*time.Second)
		if err != nil {
			acc.AddError(fmt.Errorf("cannot get SSL cert '%s': %s", location, err.Error()))
		}

		for _, cert := range certs {
			fields := getFields(cert, now)
			tags := getTags(cert.Subject, location)

			acc.AddFields("x509_cert", fields, tags)
		}
	}

	return nil
}

func init() {
	inputs.Add("x509_cert", func() telegraf.Input {
		return &X509Cert{
			Sources: []string{},
			Timeout: internal.Duration{Duration: 5},
		}
	})
}
