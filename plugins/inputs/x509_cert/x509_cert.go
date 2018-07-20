package x509_cert

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	_tls "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
  ## List of local SSL files
  # files = ["/etc/ssl/certs/ssl-cert-snakeoil.pem"]
  ## List of servers
  # servers = ["tcp://example.org:443"]
  ## Timeout for SSL connection
  # timeout = 5
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
	Servers []string          `toml:"servers"`
	Files   []string          `toml:"files"`
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

func (c *X509Cert) getRemoteCert(server string, timeout time.Duration) (*x509.Certificate, error) {
	tlsCfg, err := c.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	network := "tcp"
	host_port := server
	vals := strings.Split(server, "://")

	if len(vals) > 1 {
		network = vals[0]
		host_port = vals[1]
	}

	ipConn, err := net.DialTimeout(network, host_port, timeout)
	if err != nil {
		return nil, err
	}
	defer ipConn.Close()

	conn := tls.Client(ipConn, tlsCfg)
	defer conn.Close()

	hsErr := conn.Handshake()
	if hsErr != nil {
		return nil, hsErr
	}

	certs := conn.ConnectionState().PeerCertificates

	if certs == nil || len(certs) < 1 {
		return nil, fmt.Errorf("couldn't get remote certificate")
	}

	return certs[0], nil
}

func getLocalCert(filename string) (*x509.Certificate, error) {
	content, err := ioutil.ReadFile(filename)
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

	return cert, nil
}

func getFields(cert *x509.Certificate, now time.Time) map[string]interface{} {
	age := int(now.Sub(cert.NotBefore).Seconds())
	expiry := int(cert.NotAfter.Sub(now).Seconds())
	startdate := cert.NotBefore.Unix()
	enddate := cert.NotAfter.Unix()
	valid := expiry > 0

	fields := map[string]interface{}{
		"age":       age,
		"expiry":    expiry,
		"startdate": startdate,
		"enddate":   enddate,
		"valid":     valid,
	}

	return fields
}

// Gather adds metrics into the accumulator.
func (c *X509Cert) Gather(acc telegraf.Accumulator) error {
	now := time.Now()

	for _, server := range c.Servers {
		cert, err := c.getRemoteCert(server, c.Timeout.Duration*time.Second)
		if err != nil {
			return fmt.Errorf("cannot get remote SSL cert '%s': %s", server, err)
		}

		tags := map[string]string{
			"server": server,
		}

		fields := getFields(cert, now)

		acc.AddFields("x509_cert", fields, tags)
	}

	for _, file := range c.Files {
		cert, err := getLocalCert(file)
		if err != nil {
			return fmt.Errorf("cannot get local SSL cert '%s': %s", file, err)
		}

		tags := map[string]string{
			"file": file,
		}

		fields := getFields(cert, now)

		acc.AddFields("x509_cert", fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("x509_cert", func() telegraf.Input {
		return &X509Cert{
			Files:   []string{},
			Servers: []string{},
			Timeout: internal.Duration{Duration: 5},
		}
	})
}
