// Package x509_cert reports metrics from an SSL certificate.
//
//go:generate ../../../tools/readme_config_includer/generator
package x509_cert

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pion/dtls/v2"
	"golang.org/x/crypto/ocsp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/globpath"
	"github.com/influxdata/telegraf/plugins/common/proxy"
	commontls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// Regexp for handling file URIs containing a drive letter and leading slash
var reDriveLetter = regexp.MustCompile(`^/([a-zA-Z]:/)`)

// X509Cert holds the configuration of the plugin.
type X509Cert struct {
	Sources          []string        `toml:"sources"`
	Timeout          config.Duration `toml:"timeout"`
	ServerName       string          `toml:"server_name"`
	ExcludeRootCerts bool            `toml:"exclude_root_certs"`
	Log              telegraf.Logger `toml:"-"`
	commontls.ClientConfig
	proxy.TCPProxy

	tlsCfg    *tls.Config
	locations []*url.URL
	globpaths []*globpath.GlobPath

	classification map[string]string
}

func (*X509Cert) SampleConfig() string {
	return sampleConfig
}

func (c *X509Cert) Init() error {
	// Check if we do have at least one source
	if len(c.Sources) == 0 {
		return errors.New("no source configured")
	}

	// Check the server name and transfer it if necessary
	if c.ClientConfig.ServerName != "" && c.ServerName != "" {
		return fmt.Errorf("both server_name (%q) and tls_server_name (%q) are set, but they are mutually exclusive", c.ServerName, c.ClientConfig.ServerName)
	} else if c.ServerName != "" {
		// Store the user-provided server-name in the TLS configuration
		c.ClientConfig.ServerName = c.ServerName
	}

	// Normalize the sources, handle files and file-globbing
	if err := c.sourcesToURLs(); err != nil {
		return err
	}

	// Create the TLS configuration
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

// Gather adds metrics into the accumulator.
func (c *X509Cert) Gather(acc telegraf.Accumulator) error {
	now := time.Now()

	collectedUrls := append(c.locations, c.collectCertURLs()...)
	for _, location := range collectedUrls {
		certs, ocspresp, err := c.getCert(location, time.Duration(c.Timeout))
		if err != nil {
			acc.AddError(fmt.Errorf("cannot get SSL cert %q: %w", location, err))
		}

		// Add all returned certs to the pool of intermediates except for
		// the leaf node which has to come first
		intermediates := x509.NewCertPool()
		if len(certs) > 1 {
			for _, c := range certs[1:] {
				intermediates.AddCert(c)
			}
		}

		dnsName := c.serverName(location)
		results := make([]error, 0, len(certs))
		c.classification = make(map[string]string)
		for _, cert := range certs {
			// The first certificate is the leaf/end-entity certificate which
			// needs DNS name validation against the URL hostname.
			opts := x509.VerifyOptions{
				Intermediates: intermediates,
				KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
				Roots:         c.tlsCfg.RootCAs,
				DNSName:       dnsName,
			}
			// Reset DNS name to only use it for the leaf node
			dnsName = ""

			// Do the processing
			results = append(results, c.processCertificate(cert, opts))
		}

		for i, cert := range certs {
			fields := getFields(cert, now)
			tags := getTags(cert, location.String())

			// Extract the verification result
			err := results[i]
			if err == nil {
				tags["verification"] = "valid"
				fields["verification_code"] = 0
			} else {
				tags["verification"] = "invalid"
				fields["verification_code"] = 1
				fields["verification_error"] = err.Error()
			}
			// OCSPResponse only for leaf cert
			if i == 0 && ocspresp != nil && len(*ocspresp) > 0 {
				var ocspissuer *x509.Certificate
				for _, chaincert := range certs[1:] {
					if cert.Issuer.CommonName == chaincert.Subject.CommonName &&
						cert.Issuer.SerialNumber == chaincert.Subject.SerialNumber {
						ocspissuer = chaincert
						break
					}
				}
				resp, err := ocsp.ParseResponse(*ocspresp, ocspissuer)
				if err != nil {
					if ocspissuer == nil {
						tags["ocsp_stapled"] = "no"
						fields["ocsp_error"] = err.Error()
					} else {
						ocspissuer = nil // retry parsing w/out issuer cert
						resp, err = ocsp.ParseResponse(*ocspresp, ocspissuer)
					}
				}
				if err != nil {
					tags["ocsp_stapled"] = "no"
					fields["ocsp_error"] = err.Error()
				} else {
					tags["ocsp_stapled"] = "yes"
					if ocspissuer != nil {
						tags["ocsp_verified"] = "yes"
					} else {
						tags["ocsp_verified"] = "no"
					}
					// resp.Status: 0=Good 1=Revoked 2=Unknown
					fields["ocsp_status_code"] = resp.Status
					switch resp.Status {
					case 0:
						tags["ocsp_status"] = "good"
					case 1:
						tags["ocsp_status"] = "revoked"
						// Status=Good: revoked_at always = -62135596800
						fields["ocsp_revoked_at"] = resp.RevokedAt.Unix()
					default:
						tags["ocsp_status"] = "unknown"
					}
					fields["ocsp_produced_at"] = resp.ProducedAt.Unix()
					fields["ocsp_this_update"] = resp.ThisUpdate.Unix()
					fields["ocsp_next_update"] = resp.NextUpdate.Unix()
				}
			} else {
				tags["ocsp_stapled"] = "no"
			}

			// Determine the classification
			sig := hex.EncodeToString(cert.Signature)
			if class, found := c.classification[sig]; found {
				tags["type"] = class
			} else {
				tags["type"] = "leaf"
			}

			acc.AddFields("x509_cert", fields, tags)
			if c.ExcludeRootCerts {
				break
			}
		}
	}

	return nil
}

func (c *X509Cert) processCertificate(certificate *x509.Certificate, opts x509.VerifyOptions) error {
	chains, err := certificate.Verify(opts)
	if err != nil {
		c.Log.Debugf("Invalid certificate %v", certificate.SerialNumber.Text(16))
		c.Log.Debugf("  cert DNS names:    %v", certificate.DNSNames)
		c.Log.Debugf("  cert IP addresses: %v", certificate.IPAddresses)
		c.Log.Debugf("  cert subject:      %v", certificate.Subject)
		c.Log.Debugf("  cert issuer:       %v", certificate.Issuer)
		c.Log.Debugf("  opts.DNSName:      %v", opts.DNSName)
		c.Log.Debugf("  verify options:    %v", opts)
		c.Log.Debugf("  verify error:      %v", err)
		c.Log.Debugf("  tlsCfg.ServerName: %v", c.tlsCfg.ServerName)
		c.Log.Debugf("  ServerName:        %v", c.ServerName)
	}

	// Check if the certificate is a root-certificate.
	// The only reliable way to distinguish root certificates from
	// intermediates is the fact that root certificates are self-signed,
	// i.e. you can verify the certificate with its own public key.
	rootErr := certificate.CheckSignature(certificate.SignatureAlgorithm, certificate.RawTBSCertificate, certificate.Signature)
	if rootErr == nil {
		sig := hex.EncodeToString(certificate.Signature)
		c.classification[sig] = "root"
	}

	// Identify intermediate certificates
	for _, chain := range chains {
		// All nodes except the first one are of intermediate or CA type.
		// Mark them as such. We never add leaf nodes to the classification
		// so in the end if a cert is NOT in the classification it is a true
		// leaf node.
		for _, cert := range chain[1:] {
			// Never change a classification if we already have one
			sig := hex.EncodeToString(cert.Signature)
			if _, found := c.classification[sig]; found {
				continue
			}

			// We found an intermediate certificate which is not a CA. This
			// should never happen actually.
			if !cert.IsCA {
				c.classification[sig] = "unknown"
				continue
			}

			// The only reliable way to distinguish root certificates from
			// intermediates is the fact that root certificates are self-signed,
			// i.e. you can verify the certificate with its own public key.
			rootErr := cert.CheckSignature(cert.SignatureAlgorithm, cert.RawTBSCertificate, cert.Signature)
			if rootErr != nil {
				c.classification[sig] = "intermediate"
			} else {
				c.classification[sig] = "root"
			}
		}
	}

	return err
}

func (c *X509Cert) sourcesToURLs() error {
	for _, source := range c.Sources {
		if strings.HasPrefix(source, "file://") || strings.HasPrefix(source, "/") {
			source = filepath.ToSlash(strings.TrimPrefix(source, "file://"))
			// Removing leading slash in Windows path containing a drive-letter
			// like "file:///C:/Windows/..."
			source = reDriveLetter.ReplaceAllString(source, "$1")
			g, err := globpath.Compile(source)
			if err != nil {
				return fmt.Errorf("could not compile glob %q: %w", source, err)
			}
			c.globpaths = append(c.globpaths, g)
		} else {
			if strings.Index(source, ":\\") == 1 {
				source = "file://" + filepath.ToSlash(source)
			}
			u, err := url.Parse(source)
			if err != nil {
				return fmt.Errorf("failed to parse cert location: %w", err)
			}
			c.locations = append(c.locations, u)
		}
	}

	return nil
}

func (c *X509Cert) serverName(u *url.URL) string {
	if c.tlsCfg.ServerName != "" {
		return c.tlsCfg.ServerName
	}
	return u.Hostname()
}

func (c *X509Cert) getCert(u *url.URL, timeout time.Duration) ([]*x509.Certificate, *[]byte, error) {
	protocol := u.Scheme
	switch u.Scheme {
	case "udp", "udp4", "udp6":
		ipConn, err := net.DialTimeout(u.Scheme, u.Host, timeout)
		if err != nil {
			return nil, nil, err
		}
		defer ipConn.Close()

		dtlsCfg := &dtls.Config{
			InsecureSkipVerify: true,
			Certificates:       c.tlsCfg.Certificates,
			RootCAs:            c.tlsCfg.RootCAs,
			ServerName:         c.serverName(u),
		}
		conn, err := dtls.Client(ipConn, dtlsCfg)
		if err != nil {
			return nil, nil, err
		}
		defer conn.Close()

		rawCerts := conn.ConnectionState().PeerCertificates
		var certs []*x509.Certificate
		for _, rawCert := range rawCerts {
			parsed, err := x509.ParseCertificate(rawCert)
			if err != nil {
				return nil, nil, err
			}

			if parsed != nil {
				certs = append(certs, parsed)
			}
		}

		return certs, nil, nil
	case "https":
		protocol = "tcp"
		if u.Port() == "" {
			u.Host += ":443"
		}
		fallthrough
	case "tcp", "tcp4", "tcp6":
		dialer, err := c.Proxy()
		if err != nil {
			return nil, nil, err
		}
		ipConn, err := dialer.DialTimeout(protocol, u.Host, timeout)
		if err != nil {
			return nil, nil, err
		}
		defer ipConn.Close()

		downloadTLSCfg := c.tlsCfg.Clone()
		downloadTLSCfg.ServerName = c.serverName(u)
		downloadTLSCfg.InsecureSkipVerify = true

		conn := tls.Client(ipConn, downloadTLSCfg)
		defer conn.Close()

		hsErr := conn.Handshake()
		if hsErr != nil {
			return nil, nil, hsErr
		}

		certs := conn.ConnectionState().PeerCertificates
		ocspresp := conn.ConnectionState().OCSPResponse

		return certs, &ocspresp, nil
	case "file":
		content, err := os.ReadFile(u.Path)
		if err != nil {
			return nil, nil, err
		}
		var certs []*x509.Certificate
		for {
			block, rest := pem.Decode(bytes.TrimSpace(content))
			if block == nil {
				return nil, nil, fmt.Errorf("failed to parse certificate PEM")
			}

			if block.Type == "CERTIFICATE" {
				cert, err := x509.ParseCertificate(block.Bytes)
				if err != nil {
					return nil, nil, err
				}
				certs = append(certs, cert)
			}
			if len(rest) == 0 {
				break
			}
			content = rest
		}
		return certs, nil, nil
	case "smtp":
		ipConn, err := net.DialTimeout("tcp", u.Host, timeout)
		if err != nil {
			return nil, nil, err
		}
		defer ipConn.Close()

		downloadTLSCfg := c.tlsCfg.Clone()
		downloadTLSCfg.ServerName = c.serverName(u)
		downloadTLSCfg.InsecureSkipVerify = true

		smtpConn, err := smtp.NewClient(ipConn, u.Host)
		if err != nil {
			return nil, nil, err
		}

		err = smtpConn.Hello(downloadTLSCfg.ServerName)
		if err != nil {
			return nil, nil, err
		}

		id, err := smtpConn.Text.Cmd("STARTTLS")
		if err != nil {
			return nil, nil, err
		}

		smtpConn.Text.StartResponse(id)
		defer smtpConn.Text.EndResponse(id)
		_, _, err = smtpConn.Text.ReadResponse(220)
		if err != nil {
			return nil, nil, fmt.Errorf("did not get 220 after STARTTLS: %w", err)
		}

		tlsConn := tls.Client(ipConn, downloadTLSCfg)
		defer tlsConn.Close()

		hsErr := tlsConn.Handshake()
		if hsErr != nil {
			return nil, nil, hsErr
		}

		certs := tlsConn.ConnectionState().PeerCertificates
		ocspresp := tlsConn.ConnectionState().OCSPResponse

		return certs, &ocspresp, nil
	default:
		return nil, nil, fmt.Errorf("unsupported scheme %q in location %s", u.Scheme, u.String())
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

func (c *X509Cert) collectCertURLs() []*url.URL {
	var urls []*url.URL

	for _, path := range c.globpaths {
		files := path.Match()
		if len(files) <= 0 {
			c.Log.Errorf("could not find file: %v", path.GetRoots())
			continue
		}
		for _, file := range files {
			fn := filepath.ToSlash(file)
			urls = append(urls, &url.URL{Scheme: "file", Path: fn})
		}
	}

	return urls
}

func init() {
	inputs.Add("x509_cert", func() telegraf.Input {
		return &X509Cert{
			Timeout: config.Duration(5 * time.Second),
		}
	})
}
