package ssl

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/errchan"
	"github.com/influxdata/telegraf/plugins/inputs"
	"net"
	"time"
)

type CheckExpire struct {
	// Server to check
	Servers []string

	// Timeout in seconds. 0 means no timeout
	Timeout string
}

// Description returns the plugin Description
func (c *CheckExpire) Description() string {
	return "time left until SSL cert is expired"
}

var sampleConfig = `
  ## server name list default [] )
  servers = ["github.com:443"]
  ## Set timeout (default 5 seconds)
  timeout = "5s"
`

// SampleConfig returns the plugin SampleConfig
func (c *CheckExpire) SampleConfig() string {
	return sampleConfig
}

// Connect to server and retrieve chain certificates
func (c *CheckExpire) checkHost(server string) ([]*x509.Certificate, error) {

	tout, err := time.ParseDuration(c.Timeout)
	if err != nil {
		return nil, err
	}
	//Connect network
	ipConn, err := net.DialTimeout("tcp", server, tout)
	if err != nil {
		return nil, err
	}
	defer ipConn.Close()

	// Configure tls to not verify if site is secure
	config := tls.Config{ServerName: server, InsecureSkipVerify: true}

	// Connect to tls
	conn := tls.Client(ipConn, &config)
	defer conn.Close()

	// Handshake with TLS to get certs
	hsErr := conn.Handshake()
	if hsErr != nil {
		return nil, hsErr
	}

	certs := conn.ConnectionState().PeerCertificates

	if certs == nil || len(certs) < 1 {
		return nil, errors.New("Could not get server's certificate from the TLS connection.")
	}
	return certs, nil

}

// Gather gets all metric fields and tags and returns any errors it encounters
func (c *CheckExpire) Gather(acc telegraf.Accumulator) error {
	errChan := errchan.New(len(c.Servers))
	for _, server := range c.Servers {
		// Prepare data
		tags := map[string]string{"server": server}
		// Gather data
		var timeToExpire time.Duration
		timeNow := time.Now()
		certs, err := c.checkHost(server)
		errChan.C <- err
		if err != nil {
			timeToExpire = 0
		} else {
			timeToExpire = certs[0].NotAfter.Sub(timeNow)
		}
		fields := map[string]interface{}{"time_to_expire": timeToExpire.Seconds()}
		// Add metrics
		acc.AddFields("ssl_cert", fields, tags)
	}
	return errChan.Error()
}

func init() {
	inputs.Add("check_ssl", func() telegraf.Input {
		return &CheckExpire{Timeout: "5s"}
	})
}
