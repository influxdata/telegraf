//go:generate ../../../tools/readme_config_includer/generator

package whois

import (
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/likexian/whois"
	"github.com/likexian/whois-parser"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Whois struct {
	Domains []string        `toml:"domains"`
	Server  string          `toml:"server"`
	Timeout config.Duration `toml:"timeout"`
	Log     telegraf.Logger `toml:"-"`

	client         *whois.Client
	whoisLookup    func(client *whois.Client, domain, server string) (string, error)
	parseWhoisData func(raw string) (whoisparser.WhoisInfo, error)
}

func (*Whois) SampleConfig() string {
	return sampleConfig
}

func (w *Whois) Init() error {
	if len(w.Domains) == 0 {
		return errors.New("no domains configured")
	}

	if w.Timeout <= 0 {
		return errors.New("timeout has to be greater than zero")
	}

	w.client = whois.NewClient()
	w.client.SetTimeout(time.Duration(w.Timeout))
	w.client.SetDisableReferralChain(true)

	if w.whoisLookup == nil {
		w.whoisLookup = func(client *whois.Client, domain, _ string) (string, error) {
			return client.Whois(domain, w.Server)
		}
	}

	if w.parseWhoisData == nil {
		w.parseWhoisData = whoisparser.Parse
	}

	return nil
}

func (w *Whois) Gather(acc telegraf.Accumulator) error {
	for _, domain := range w.Domains {
		w.Log.Tracef("Fetching WHOIS data for %q using WHOIS server %q with timeout: %v", domain, w.Server, w.Timeout)

		// Fetch WHOIS raw data
		raw, err := w.whoisLookup(w.client, domain, w.Server)
		if err != nil {
			acc.AddError(fmt.Errorf("whois query failed for %q: %w", domain, err))
			continue
		}

		// Parse WHOIS data using whois-parser
		data, err := w.parseWhoisData(raw)
		if err != nil {
			// Skip metric recording for these errors
			if errors.Is(err, whoisparser.ErrDomainDataInvalid) {
				acc.AddError(fmt.Errorf("whois parsing failed for %q: %w", domain, err))
				continue
			}

			var status string
			switch {
			case errors.Is(err, whoisparser.ErrNotFoundDomain):
				status = "domainNotFound"
			case errors.Is(err, whoisparser.ErrReservedDomain):
				status = "reservedDomain"
			case errors.Is(err, whoisparser.ErrPremiumDomain):
				status = "reservedDomain"
			case errors.Is(err, whoisparser.ErrBlockedDomain):
				status = "blockedDomain"
			case errors.Is(err, whoisparser.ErrDomainLimitExceed):
				status = "domainLimitExceed"
			default:
				status = "unknown"
			}

			acc.AddFields(
				"whois",
				map[string]interface{}{
					"error": err.Error(),
				},
				map[string]string{
					"domain": domain,
					"status": status,
				},
			)

			continue
		}

		// Extract expiration date
		var expirationTimestamp int64
		var expiry int
		if data.Domain.ExpirationDateInTime != nil {
			expirationTimestamp = data.Domain.ExpirationDateInTime.Unix()

			// Calculate expiry in seconds
			expiry = int(time.Until(*data.Domain.ExpirationDateInTime).Seconds())
		}

		// Extract creation date
		var creationTimestamp int64
		if data.Domain.CreatedDateInTime != nil {
			creationTimestamp = data.Domain.CreatedDateInTime.Unix()
		}

		// Extract updated date
		var updatedTimestamp int64
		if data.Domain.UpdatedDateInTime != nil {
			updatedTimestamp = data.Domain.UpdatedDateInTime.Unix()
		}

		// Extract registrar name (handle nil)
		var registrar string
		if data.Registrar != nil {
			registrar = data.Registrar.Name
		}

		// Extract registrant name (handle nil)
		var registrant string
		if data.Registrant != nil {
			registrant = data.Registrant.Name
		}

		// Extract status (handle empty)
		status := "unknown"
		if len(data.Domain.Status) > 0 {
			status = strings.Join(data.Domain.Status, ",")
		}

		// Add metrics
		fields := map[string]interface{}{
			"creation_timestamp":   creationTimestamp,
			"dnssec_enabled":       data.Domain.DNSSec,
			"expiration_timestamp": expirationTimestamp,
			"expiry":               expiry,
			"updated_timestamp":    updatedTimestamp,
			"registrar":            registrar,
			"registrant":           registrant,
			"name_servers":         strings.Join(data.Domain.NameServers, ","),
		}
		tags := map[string]string{
			"domain": domain,
			"status": status,
		}

		acc.AddFields("whois", fields, tags)
	}

	return nil
}

// Plugin registration
func init() {
	inputs.Add("whois", func() telegraf.Input {
		return &Whois{
			Timeout: config.Duration(30 * time.Second),
			Server:  "whois.iana.org",
		}
	})
}
