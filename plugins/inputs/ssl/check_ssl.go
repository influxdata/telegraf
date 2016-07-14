package ssl

import (
	"crypto/sha1"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var signatureAlgorithm = [...]string{
	"UnknownSignatureAlgorithm",
	"MD2WithRSA",
	"MD5WithRSA",
	"SHA1WithRSA",
	"SHA256WithRSA",
	"SHA384WithRSA",
	"SHA512WithRSA",
	"DSAWithSHA1",
	"DSAWithSHA256",
	"ECDSAWithSHA1",
	"ECDSAWithSHA256",
	"ECDSAWithSHA384",
	"ECDSAWithSHA512",
}

var publicKeyAlgorithm = [...]string{
	"UnknownPublicKeyAlgorithm",
	"RSA",
	"DAS",
	"ECDSA",
}

func SHA1Hash(data []byte) string {
	h := sha1.New()
	h.Write(data)
	return fmt.Sprintf("%X", h.Sum(nil))
}

// SSLCerts struct
type SSLCerts struct {
	SHA1                string
	SubjectKeyId        string
	Version             int
	SignatureAlgorithm  string
	PublicKeyAlgorithm  string
	Subject             string
	DNSNames            []string
	NotBefore, NotAfter string
	ExpiresIn           int64
	Issuer              string
	AuthorityKeyId      string
}

type CheckExpire struct {
	// Server to check
	Servers []string

	// SSL server port number (Default 443)
	Port string

	// Timeout in seconds. 0 means no timeout
	Timeout int

	// Skip SSL Verify?
	Skipverify bool `toml:"skip_verify"`
}

// Description returns the plugin Description
func (c *CheckExpire) Description() string {
	return "Days left until SSL cert is expired"
}

var sampleConfig = `
  ## server name list default ["github.com"] )
  servers = ["github.com"]
  ## Set timeout (default 5 seconds)
  timeout = 5
  ## SSL Port (Default 443)
  port = "443"
  ## SSL Skip Verification validity of certificates
  skip_verify = false
`

// SampleConfig returns the plugin SampleConfig
func (c *CheckExpire) SampleConfig() string {
	return sampleConfig
}

// Connect to server and retrieve chain certificates
func (c *CheckExpire) checkHost(server string) ([]SSLCerts, error) {

	canonicalName := server + ":" + c.Port
	tout := time.Duration(c.Timeout) * time.Second
	//Connect network
	ipConn, err := net.DialTimeout("tcp", canonicalName, tout)
	if err != nil {
		return nil, err
	}
	defer ipConn.Close()

	// Configure tls to look at domainName
	config := tls.Config{ServerName: canonicalName,
		InsecureSkipVerify: c.Skipverify}

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

	sslcerts := make([]SSLCerts, len(certs))
	timeNow := time.Now()

	for i, cert := range certs {
		s := SSLCerts{
			SHA1:         SHA1Hash(cert.Raw),
			SubjectKeyId: fmt.Sprintf("%X", cert.SubjectKeyId),
			Version:      cert.Version, SignatureAlgorithm: signatureAlgorithm[cert.SignatureAlgorithm],
			PublicKeyAlgorithm: publicKeyAlgorithm[cert.PublicKeyAlgorithm],
			Subject:            cert.Subject.CommonName,
			DNSNames:           cert.DNSNames,
			NotBefore:          cert.NotBefore.Local().String(),
			NotAfter:           cert.NotAfter.Local().String(),
			ExpiresIn:          int64(cert.NotAfter.Sub(timeNow).Hours() / 24),
			Issuer:             cert.Issuer.CommonName,
			AuthorityKeyId:     fmt.Sprintf("%X", cert.AuthorityKeyId),
		}
		sslcerts[i] = s

	}

	return sslcerts, nil
}

// Gather gets all metric fields and tags and returns any errors it encounters
func (c *CheckExpire) Gather(acc telegraf.Accumulator) error {
	c.setDefaultValues()
	for _, server := range c.Servers {
		// Prepare data
		tags := map[string]string{"server": server}
		// Gather data
		certs, err := c.checkHost(server)
		if err != nil {
			return err
		}
		daysToExpire := certs[0].ExpiresIn
		fields := map[string]interface{}{"days_to_expire": daysToExpire}
		// Add metrics
		acc.AddFields("expire_time", fields, tags)
	}
	return nil
}

func (c *CheckExpire) setDefaultValues() {
	if len(c.Servers) == 0 {
		c.Servers = []string{"github.com"}
	}

	if len(c.Port) == 0 {
		c.Port = "443"
	}

	if c.Timeout == 0 {
		c.Timeout = 5
	}
}

func init() {
	inputs.Add("check_ssl", func() telegraf.Input {
		return &CheckExpire{}
	})
}
