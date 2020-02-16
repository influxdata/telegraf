// Package x509_cert reports metrics from an SSL certificate.
package x509_crl

import (
	"bytes"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const x509CrlMeasurement = "x509_crl"

const sampleConfig = `
  ## List CRLs sources
  sources = ["/tmp/crl.pem"]
`
const description = "Reads metrics from PEM encoded X509 CRL files"

// X509CRL holds the configuration of the plugin.
type X509CRL struct {
	Sources []string `toml:"sources"`
}

// Description returns description of the plugin.
func (configuration *X509CRL) Description() string {
	return description
}

// SampleConfig returns configuration sample for the plugin.
func (configuration *X509CRL) SampleConfig() string {
	return sampleConfig
}

func (configuration *X509CRL) sourceToURL(source string) (*url.URL, error) {
	if strings.HasPrefix(source, "/") {
		source = "file://" + source
	}

	u, err := url.Parse(source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CRL source - %s", err.Error())
	}

	return u, nil
}

func (configuration *X509CRL) getCRL(crlURL *url.URL) ([]*pkix.CertificateList, error) {
	switch crlURL.Scheme {
	case "file":
		content, err := ioutil.ReadFile(crlURL.Path)
		if err != nil {
			return nil, err
		}
		var crls []*pkix.CertificateList
		for {
			block, rest := pem.Decode(bytes.TrimSpace(content))
			if block == nil {
				return nil, fmt.Errorf("failed to parse CRL PEM")
			}

			if block.Type == "X509 CRL" {
				crl, err := x509.ParseCRL(block.Bytes)
				if err != nil {
					return nil, err
				}
				crls = append(crls, crl)
			}
			if rest == nil || len(rest) == 0 {
				break
			}
			content = rest
		}
		return crls, nil
	default:
		return nil, fmt.Errorf("unsuported scheme '%s' in location %s", crlURL.Scheme, crlURL.String())
	}
}

func getFields(crl *pkix.CertificateList) map[string]interface{} {
	return map[string]interface{}{
		"start_date":           crl.TBSCertList.ThisUpdate.Unix() * 1000, // EPOCH in ms
		"end_date":             crl.TBSCertList.NextUpdate.Unix() * 1000, // EPOCH in ms
		"has_expired":          crl.HasExpired(time.Now()),
		"revoked_certificates": len(crl.TBSCertList.RevokedCertificates),
	}
}

func getTags(crl *pkix.CertificateList, location string) map[string]string {
	return map[string]string{
		"source":  location,
		"issuer":  crl.TBSCertList.Issuer.String(),
		"version": strconv.Itoa(crl.TBSCertList.Version),
	}
}

// Gather adds metrics into the accumulator.
func (configuration *X509CRL) Gather(acc telegraf.Accumulator) error {

	for _, source := range configuration.Sources {
		crlURL, err := configuration.sourceToURL(source)
		if err != nil {
			acc.AddError(err)
			return nil
		}

		crls, err := configuration.getCRL(crlURL)
		if err != nil {
			acc.AddError(fmt.Errorf("cannot get SSL crl '%s': %s", source, err.Error()))
		}

		for _, crl := range crls {
			fields := getFields(crl)
			tags := getTags(crl, source)

			acc.AddFields(x509CrlMeasurement, fields, tags)
		}
	}

	return nil
}

func (configuration *X509CRL) Init() error {
	return nil
}

func init() {
	inputs.Add(x509CrlMeasurement, func() telegraf.Input {
		return &X509CRL{
			Sources: []string{},
		}
	})
}
