// Package x509_cert reports metrics from an SSL certificate.
package x509_cert

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
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

  ## Pass a different name into the TLS request (Server Name Indication)
  ##   example: server_name = "myhost.example.org"
  # server_name = ""

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
`
const description = "Reads metrics from a SSL certificate"

// X509Cert holds the configuration of the plugin.
type X509Cert struct {
	Sources    []string          `toml:"sources"`
	Timeout    internal.Duration `toml:"timeout"`
	ServerName string            `toml:"server_name"`
	tlsCfg     *tls.Config
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

func (c *X509Cert) locationToURL(location string) (*url.URL, error) {
	if strings.HasPrefix(location, "/") {
		location = "file://" + location
	}

	u, err := url.Parse(location)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cert location - %s", err.Error())
	}

	return u, nil
}

func (c *X509Cert) getCert(u *url.URL, timeout time.Duration) ([]*x509.Certificate, error) {
	switch u.Scheme {
	case "https":
		u.Scheme = "tcp"
		fallthrough
	case "udp", "udp4", "udp6":
		fallthrough
	case "tcp", "tcp4", "tcp6":
		ipConn, err := net.DialTimeout(u.Scheme, u.Host, timeout)
		if err != nil {
			return nil, err
		}
		defer ipConn.Close()

		if c.ServerName == "" {
			c.tlsCfg.ServerName = u.Hostname()
		} else {
			c.tlsCfg.ServerName = c.ServerName
		}

		c.tlsCfg.InsecureSkipVerify = true
		conn := tls.Client(ipConn, c.tlsCfg)
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
		var certs []*x509.Certificate
		for {
			block, rest := pem.Decode(bytes.TrimSpace(content))
			if block == nil {
				return nil, fmt.Errorf("failed to parse certificate PEM")
			}

			if block.Type == "CERTIFICATE" {
				cert, err := x509.ParseCertificate(block.Bytes)
				if err != nil {
					return nil, err
				}
				certs = append(certs, cert)
			}
			if rest == nil || len(rest) == 0 {
				break
			}
			content = rest
		}
		return certs, nil
	default:
		return nil, fmt.Errorf("unsuported scheme '%s' in location %s", u.Scheme, u.String())
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

func getTags(cert *x509.Certificate, location string) map[string]string {
	tags := map[string]string{
		"source":               location,
		"common_name":          cert.Subject.CommonName,
		"serial_number":        cert.SerialNumber.Text(16),
		"signature_algorithm":  cert.SignatureAlgorithm.String(),
		"public_key_algorithm": cert.PublicKeyAlgorithm.String(),
	}

	if len(cert.Subject.Organization) > 0 {
		tags["organization"] = cert.Subject.Organization[0]
	}
	if len(cert.Subject.OrganizationalUnit) > 0 {
		tags["organizational_unit"] = cert.Subject.OrganizationalUnit[0]
	}
	if len(cert.Subject.Country) > 0 {
		tags["country"] = cert.Subject.Country[0]
	}
	if len(cert.Subject.Province) > 0 {
		tags["province"] = cert.Subject.Province[0]
	}
	if len(cert.Subject.Locality) > 0 {
		tags["locality"] = cert.Subject.Locality[0]
	}

	tags["issuer_common_name"] = cert.Issuer.CommonName
	tags["issuer_serial_number"] = cert.Issuer.SerialNumber

	san := append(cert.DNSNames, cert.EmailAddresses...)
	for _, ip := range cert.IPAddresses {
		san = append(san, ip.String())
	}
	for _, uri := range cert.URIs {
		san = append(san, uri.String())
	}
	tags["san"] = strings.Join(san, ",")

	return tags
}

// Gather adds metrics into the accumulator.
func (c *X509Cert) Gather(acc telegraf.Accumulator) error {
	now := time.Now()

	for _, location := range c.Sources {
		u, err := c.locationToURL(location)
		if err != nil {
			acc.AddError(err)
			return nil
		}

		certs, err := c.getCert(u, c.Timeout.Duration*time.Second)
		if err != nil {
			acc.AddError(fmt.Errorf("cannot get SSL cert '%s': %s", location, err.Error()))
		}

		for i, cert := range certs {
			fields := getFields(cert, now)
			tags := getTags(cert, location)

			// The first certificate is the leaf/end-entity certificate which needs DNS
			// name validation against the URL hostname.
			opts := x509.VerifyOptions{
				Intermediates: x509.NewCertPool(),
			}
			if i == 0 {
				if c.ServerName == "" {
					opts.DNSName = u.Hostname()
				} else {
					opts.DNSName = c.ServerName
				}
				for j, cert := range certs {
					if j != 0 {
						opts.Intermediates.AddCert(cert)
					}
				}
			}
			if c.tlsCfg.RootCAs != nil {
				opts.Roots = c.tlsCfg.RootCAs
			}

			_, err = cert.Verify(opts)
			if err == nil {
				tags["verification"] = "valid"
				fields["verification_code"] = 0
			} else {
				tags["verification"] = "invalid"
				fields["verification_code"] = 1
				fields["verification_error"] = err.Error()
			}

			acc.AddFields("x509_cert", fields, tags)
		}
	}

	return nil
}

func (c *X509Cert) Init() error {
	tlsCfg, err := c.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}
	if tlsCfg == nil {
		tlsCfg = &tls.Config{}
	}

	c.tlsCfg = tlsCfg

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
