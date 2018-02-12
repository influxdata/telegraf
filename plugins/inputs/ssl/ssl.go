package ssl

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"crypto/x509"
	"time"
	"net"
	"crypto/tls"
	"github.com/pkg/errors"
	"strconv"
)

type Ssl struct {
	Servers []Server
}

type Server struct {
	Domain string
	Port int
	Timeout int
}

var sampleConfig  = `
  ## Server to check
  [[inputs.ssl.servers]]
    domain = "google.com"
    port = 443
    timeout = 5
  ## Server to check
  [[inputs.ssl.servers]]
    domain = "github.com"
    port = 443
    timeout = 5
`

func (s *Ssl) SampleConfig() string {
	return sampleConfig
}

func (s *Ssl) Description() string {
	return "Check expiration date and domains of ssl certificate"
}

func (s *Ssl) Gather(acc telegraf.Accumulator) error  {
	for _, server := range s.Servers {
		h := getServerAddress(server.Domain, server.Port)
		timeNow := time.Now()
		timeToExp := int64(0)
		fields := make(map[string]interface{})
		tags := make(map[string]string)
		certs, err := getServerCertsChain(server.Domain, server.Port, server.Timeout)

		if err != nil {
			acc.AddError(err)
		} else {
			cert := certs[0]
			if cert.NotAfter.UnixNano() < timeNow.UnixNano() {
				acc.AddError(errors.New("[" + h + "] cert has expired"))
			} else {
				timeToExp = int64(cert.NotAfter.Sub(timeNow))
			}
			if !isStringInSlice(server.Domain, cert.DNSNames) {
				acc.AddError(errors.New("[" + h + "] cert and domain mismatch"))
				timeToExp = int64(0)
			}
		}
		fields["time_to_expiration"] = timeToExp
		tags["domain"] = server.Domain
		tags["port"] = strconv.FormatInt(int64(server.Port), 10)

		acc.AddFields("ssl", fields, tags)
	}
	return nil
}

func getServerCertsChain(d string, p int, t int) ([]*x509.Certificate, error) {
	h := getServerAddress(d, p)
	ipConn, err := net.DialTimeout("tcp", h, time.Duration(t) * time.Second)
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

func getServerAddress(d string, p int) string {
	return d + ":" + strconv.FormatInt(int64(p), 10)
}

func isStringInSlice(n string, s []string) bool {
	for _, d := range s {
		if n == d {
			return true
		}
	}
	return false
}

func init() {
	inputs.Add("ssl", func() telegraf.Input {
		return &Ssl{}
	})
}
