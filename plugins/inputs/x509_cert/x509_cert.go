// Package x509_cert reports metrics from an SSL certificate.
//
//go:generate ../../../tools/readme_config_includer/generator
package x509_cert

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pion/dtls/v2"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/globpath"
	"github.com/influxdata/telegraf/plugins/common/proxy"
	commontls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

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
	collectedUrls, err := c.collectCertURLs()
	if err != nil {
		acc.AddError(fmt.Errorf("getting some certificates failed: %w", err))
	}

	for _, location := range append(c.locations, collectedUrls...) {
		certs, err := c.getCert(location, time.Duration(c.Timeout))
		if err != nil {
			acc.AddError(fmt.Errorf("cannot get SSL cert '%s': %s", location, err.Error()))
		}
		dnsName, err := c.serverName(location)
		if err != nil {
			acc.AddError(fmt.Errorf("resolving name of %q failed: %w", location, err))
			continue
		}
		for i, cert := range certs {
			fields := getFields(cert, now)
			tags := getTags(cert, location.String())

			// The first certificate is the leaf/end-entity certificate which
			// needs DNS name validation against the URL hostname.
			opts := x509.VerifyOptions{
				Intermediates: x509.NewCertPool(),
				KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
				Roots:         c.tlsCfg.RootCAs,
				DNSName:       dnsName,
			}
			// Reset DNS name to only use it for the leaf node
			dnsName = ""

			// Add all returned certs to the pool if intermediates except for
			// the leaf node and ourself
			for j, c := range certs[1:] {
				if i != j {
					opts.Intermediates.AddCert(c)
				}
			}

			if _, err := cert.Verify(opts); err == nil {
				tags["verification"] = "valid"
				fields["verification_code"] = 0
			} else {
				c.Log.Debugf("Invalid certificate at index %2d!", i)
				c.Log.Debugf("  cert DNS names:    %v", cert.DNSNames)
				c.Log.Debugf("  cert IP addresses: %v", cert.IPAddresses)
				c.Log.Debugf("  opts.DNSName:      %v", opts.DNSName)
				c.Log.Debugf("  verify options:    %v", opts)
				c.Log.Debugf("  verify error:      %v", err)
				c.Log.Debugf("  location:          %v", location)
				c.Log.Debugf("  tlsCfg.ServerName: %v", c.tlsCfg.ServerName)
				c.Log.Debugf("  ServerName:        %v", c.ServerName)
				tags["verification"] = "invalid"
				fields["verification_code"] = 1
				fields["verification_error"] = err.Error()
			}

			acc.AddFields("x509_cert", fields, tags)
			if c.ExcludeRootCerts {
				break
			}
		}
	}

	return nil
}

func (c *X509Cert) sourcesToURLs() error {
	for _, source := range c.Sources {
		if strings.HasPrefix(source, "file://") ||
			strings.HasPrefix(source, "/") {
			source = filepath.ToSlash(strings.TrimPrefix(source, "file://"))
			g, err := globpath.Compile(source)
			if err != nil {
				return fmt.Errorf("could not compile glob %v: %v", source, err)
			}
			c.globpaths = append(c.globpaths, g)
		} else {
			if strings.Index(source, ":\\") == 1 {
				source = "file://" + filepath.ToSlash(source)
			}
			u, err := url.Parse(source)
			if err != nil {
				return fmt.Errorf("failed to parse cert location - %s", err.Error())
			}
			c.locations = append(c.locations, u)
		}
	}

	return nil
}

func (c *X509Cert) serverName(u *url.URL) (string, error) {
	if c.tlsCfg.ServerName != "" {
		return c.tlsCfg.ServerName, nil
	}
	return u.Hostname(), nil
}

func (c *X509Cert) getCert(u *url.URL, timeout time.Duration) ([]*x509.Certificate, error) {
	protocol := u.Scheme
	switch u.Scheme {
	case "udp", "udp4", "udp6":
		ipConn, err := net.DialTimeout(u.Scheme, u.Host, timeout)
		if err != nil {
			return nil, err
		}
		defer ipConn.Close()

		serverName, err := c.serverName(u)
		if err != nil {
			return nil, err
		}

		dtlsCfg := &dtls.Config{
			InsecureSkipVerify: true,
			Certificates:       c.tlsCfg.Certificates,
			RootCAs:            c.tlsCfg.RootCAs,
			ServerName:         serverName,
		}
		conn, err := dtls.Client(ipConn, dtlsCfg)
		if err != nil {
			return nil, err
		}
		defer conn.Close()

		rawCerts := conn.ConnectionState().PeerCertificates
		var certs []*x509.Certificate
		for _, rawCert := range rawCerts {
			parsed, err := x509.ParseCertificate(rawCert)
			if err != nil {
				return nil, err
			}

			if parsed != nil {
				certs = append(certs, parsed)
			}
		}

		return certs, nil
	case "https":
		protocol = "tcp"
		if u.Port() == "" {
			u.Host += ":443"
		}
		fallthrough
	case "tcp", "tcp4", "tcp6":
		dialer, err := c.Proxy()
		if err != nil {
			return nil, err
		}
		ipConn, err := dialer.DialTimeout(protocol, u.Host, timeout)
		if err != nil {
			return nil, err
		}
		defer ipConn.Close()

		serverName, err := c.serverName(u)
		if err != nil {
			return nil, err
		}

		downloadTLSCfg := c.tlsCfg.Clone()
		downloadTLSCfg.ServerName = serverName
		downloadTLSCfg.InsecureSkipVerify = true

		conn := tls.Client(ipConn, downloadTLSCfg)
		defer conn.Close()

		hsErr := conn.Handshake()
		if hsErr != nil {
			return nil, hsErr
		}

		certs := conn.ConnectionState().PeerCertificates

		return certs, nil
	case "file":
		content, err := os.ReadFile(u.Path)
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
			if len(rest) == 0 {
				break
			}
			content = rest
		}
		return certs, nil
	case "smtp":
		ipConn, err := net.DialTimeout("tcp", u.Host, timeout)
		if err != nil {
			return nil, err
		}
		defer ipConn.Close()

		serverName, err := c.serverName(u)
		if err != nil {
			return nil, err
		}

		downloadTLSCfg := c.tlsCfg.Clone()
		downloadTLSCfg.ServerName = serverName
		downloadTLSCfg.InsecureSkipVerify = true

		smtpConn, err := smtp.NewClient(ipConn, u.Host)
		if err != nil {
			return nil, err
		}

		err = smtpConn.Hello(downloadTLSCfg.ServerName)
		if err != nil {
			return nil, err
		}

		id, err := smtpConn.Text.Cmd("STARTTLS")
		if err != nil {
			return nil, err
		}

		smtpConn.Text.StartResponse(id)
		defer smtpConn.Text.EndResponse(id)
		_, _, err = smtpConn.Text.ReadResponse(220)
		if err != nil {
			return nil, fmt.Errorf("did not get 220 after STARTTLS: %s", err.Error())
		}

		tlsConn := tls.Client(ipConn, downloadTLSCfg)
		defer tlsConn.Close()

		hsErr := tlsConn.Handshake()
		if hsErr != nil {
			return nil, hsErr
		}

		certs := tlsConn.ConnectionState().PeerCertificates

		return certs, nil
	default:
		return nil, fmt.Errorf("unsupported scheme '%s' in location %s", u.Scheme, u.String())
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

func (c *X509Cert) collectCertURLs() ([]*url.URL, error) {
	var urls []*url.URL

	for _, path := range c.globpaths {
		files := path.Match()
		if len(files) <= 0 {
			c.Log.Errorf("could not find file: %v", path)
			continue
		}
		for _, file := range files {
			file = "file://" + file
			u, err := url.Parse(file)
			if err != nil {
				return urls, fmt.Errorf("failed to parse cert location - %s", err.Error())
			}
			urls = append(urls, u)
		}
	}

	return urls, nil
}

func init() {
	inputs.Add("x509_cert", func() telegraf.Input {
		return &X509Cert{
			Timeout: config.Duration(5 * time.Second),
		}
	})
}
