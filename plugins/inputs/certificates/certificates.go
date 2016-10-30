package certificates

import (
	"crypto/tls"
	"net/http"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	"fmt"
)

// this is the cert data structure that will hold the values that I need to deal with
type Certificates struct {
	SHA1                string
	SubjectKeyId        string
	Version             int
	SignatureAlgorithm  string
	PublicKeyAlgorithm  string
	Subject             string
	DNSNames            []string
	NotBefore, NotAfter string
	ExpiresIn           string
	Issuer              string
	AuthorityKeyId      string
}

// sample config for the user
var sampleConfig = ``

// return the sample config to te user
func (j *Certificates) SampleConfig() string {
	return sampleConfig
}

// return the description of the telegraf input
func (j *Certificates) Description() string {
	return "Reads metrics from an SSL Certificate"
}

// return the sha hash of the cert field
func SHA1Hash(data []byte) string {
	h := sha1.New()
	h.Write(data)
	return fmt.Sprintf("%X", h.Sum(nil))
}


func checkHost(domainName string, skipVerify bool) ([]SSLCerts, error) {

	//Connect network
	ipConn, err := net.DialTimeout("tcp", domainName, 10000*time.Millisecond)
	if err != nil {
		return nil, err
	}
	defer ipConn.Close()

	// Configure tls to look at domainName
	config := tls.Config{ServerName: domainName,
		InsecureSkipVerify: skipVerify}

	// Connect to tls
	conn := tls.Client(ipConn, &config)
	defer conn.Close()

	// Handshake with TLS to get certs
	hsErr := conn.Handshake()
	if hsErr != nil {
		return nil, hsErr
	}

    // get the certs
	certs := conn.ConnectionState().PeerCertificates

    // check t0 make sure we have certs
	if certs == nil || len(certs) < 1 {
		return nil, errors.New("Could not get server's certificate from the TLS connection.")
	}

    // compile the list of certs
	sslcerts := make([]SSLCerts, len(certs))

    // this will go through each cert in the list and get the details and then compile them into
	// a data structure that can then be iterated over.
	// The qurestion is should I just dump each of these instead of building them or continue w/
	// the data structure and then iterate over that in the end like the commandline tool does.
	for i, cert := range certs {
		s := SSLCerts{SHA1: SHA1Hash(cert.Raw), SubjectKeyId: fmt.Sprintf("%X", cert.SubjectKeyId),
			Version: cert.Version, SignatureAlgorithm: signatureAlgorithm[cert.SignatureAlgorithm],
			PublicKeyAlgorithm: publicKeyAlgorithm[cert.PublicKeyAlgorithm],
			Subject:            cert.Subject.CommonName,
			DNSNames:           cert.DNSNames,
			NotBefore:          cert.NotBefore.Local().String(),
			NotAfter:           cert.NotAfter.Local().String(),
			ExpiresIn:          ExpiresIn(cert.NotAfter.Local()),
			Issuer:             cert.Issuer.CommonName,
			AuthorityKeyId:     fmt.Sprintf("%X", cert.AuthorityKeyId),
		}
		sslcerts[i] = s

	}

	return sslcerts, nil
}

func (j *Certificates) Gather(acc telegraf.Accumulator) error {
}

func init() {
}

