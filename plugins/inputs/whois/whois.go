//go:generate ../../../tools/readme_config_includer/generator

package whois

import (
	_ "embed"
	"errors"
	"fmt"
	"regexp"
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
	Domains            []string        `toml:"domains"`
	Server             string          `toml:"server"`
	Timeout            config.Duration `toml:"timeout"`
	ReferralChainQuery bool            `toml:"referral_chain_query"`
	Log                telegraf.Logger `toml:"-"`

	client *whois.Client
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
	w.client.SetDisableReferralChain(!w.ReferralChainQuery)

	if w.Server == "" {
		w.Server = "whois.iana.org"
	}

	return nil
}

var domainRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{0,253}[a-zA-Z0-9]\.[a-zA-Z]{2,}$`)

func isValidDomain(domain string) bool {
	return domainRegex.MatchString(domain)
}

func (w *Whois) Gather(acc telegraf.Accumulator) error {
	for _, domain := range w.Domains {
		if !isValidDomain(domain) {
			acc.AddError(fmt.Errorf("invalid domain format: %q", domain))
			continue
		}

		w.Log.Tracef("Fetching WHOIS data for %q using WHOIS server %q with timeout: %v", domain, w.Server, w.Timeout)

		// Fetch WHOIS raw data
		raw, err := w.client.Whois(domain, w.Server)
		if err != nil {
			acc.AddError(fmt.Errorf("whois query failed for %q: %w", domain, err))
			continue
		}

		// Parse WHOIS data using whois-parser
		data, err := whoisparser.Parse(raw)
		if err != nil {
			// Skip metric recording for these errors
			if errors.Is(err, whoisparser.ErrDomainDataInvalid) {
				acc.AddError(fmt.Errorf("whois parsing failed for %q: %w", domain, err))
				continue
			}

			var status string
			switch {
			case errors.Is(err, whoisparser.ErrNotFoundDomain):
				status = "not found"
			case errors.Is(err, whoisparser.ErrReservedDomain):
				status = "reserved"
			case errors.Is(err, whoisparser.ErrPremiumDomain):
				status = "premium"
			case errors.Is(err, whoisparser.ErrBlockedDomain):
				status = "blocked"
			case errors.Is(err, whoisparser.ErrDomainLimitExceed):
				status = "limit exceeded"
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
		var expiry int64
		if data.Domain.ExpirationDateInTime != nil {
			expirationTimestamp = data.Domain.ExpirationDateInTime.Unix()

			// Calculate expiry in seconds
			expiry = int64(time.Until(*data.Domain.ExpirationDateInTime).Seconds())
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
		registrar := "not set"
		if data.Registrar != nil {
			registrar = data.Registrar.Name
		}

		// Extract registrant name (handle nil)
		registrant := "not set"
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
		}
	})
}
