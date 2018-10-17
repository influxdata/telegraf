package ssl

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/pkg/errors"
	"net"
	"strings"
	"time"
)

type Ssl struct {
	Servers []Server
}

type Server struct {
	Host    string
	Timeout int
}

var sampleConfig = `
  ## Server to check
  [[inputs.ssl.servers]]
    host = "google.com:443"
    timeout = 5
  ## Server to check
  [[inputs.ssl.servers]]
    host = "github.com"
    timeout = 5
`

func (s *Ssl) SampleConfig() string {
	return sampleConfig
}

func (s *Ssl) Description() string {
	return "Check expiration date and domains of ssl certificate"
}

func (s *Ssl) Gather(acc telegraf.Accumulator) error {
	for _, server := range s.Servers {
		slice := strings.Split(server.Host, ":")
		domain, port := slice[0], "443"
		if len(slice) > 1 {
			port = slice[1]
		}
		h := getServerAddress(domain, port)
		timeNow := time.Now()
		timeToExp := int64(0)
		fields := make(map[string]interface{})
		tags := make(map[string]string)
		certs, err := getServerCertsChain(domain, port, server.Timeout)

		if err != nil {
			acc.AddError(err)
		} else {
			cert := certs[0]
			if cert.NotAfter.UnixNano() < timeNow.UnixNano() {
				acc.AddError(errors.New("[" + h + "] cert has expired"))
			} else {
				timeToExp = int64(cert.NotAfter.Sub(timeNow) / time.Second)
			}
			if !isDomainInCertDnsNames(domain, cert.DNSNames) {
				acc.AddError(errors.New("[" + h + "] cert and domain mismatch"))
				timeToExp = int64(0)
			}
		}
		fields["time_to_expiration"] = timeToExp
		tags["domain"] = domain
		tags["port"] = port

		acc.AddFields("ssl", fields, tags)
	}
	return nil
}

func getServerCertsChain(d string, p string, t int) ([]*x509.Certificate, error) {
	h := getServerAddress(d, p)
	ipConn, err := net.DialTimeout("tcp", h, time.Duration(t)*time.Second)
	if err != nil {
		return nil, errors.New("[" + h + "] " + err.Error())
	}
	defer ipConn.Close()

	tlsConn := tls.Client(ipConn, &tls.Config{ServerName: d, InsecureSkipVerify: true})
	defer tlsConn.Close()

	err = tlsConn.Handshake()
	if err != nil {
		return nil, errors.New("[" + h + "] " + err.Error())
	}
	certs := tlsConn.ConnectionState().PeerCertificates
	if certs == nil || len(certs) < 1 {
		return nil, errors.New("[" + h + "] cert receive error")
	}
	return certs, nil
}

func getServerAddress(d string, p string) string {
	return d + ":" + p
}

func isDomainInCertDnsNames(domain string, certDnsNames []string) bool {
	for _, d := range certDnsNames {
		if domain == d {
			return true
		}
		if d[:1] == "*" && len(domain) >= len(d[2:]) {
			d = d[2:]
			if domain == d {
				return true
			}
			start := len(domain) - len(d) - 1
			if start >= 0 && domain[start:] == "."+d {
				return true
			}
		}
	}
	return false
}

func init() {
	inputs.Add("ssl", func() telegraf.Input {
		return &Ssl{}
	})
}
