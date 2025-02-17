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

	// Ensure timeout is valid
	if w.Timeout <= 0 {
		w.Log.Debugf("Invalid timeout, setting default to 5s")
		w.Timeout = config.Duration(5 * time.Second)
	}

	w.client = whois.NewClient()
	w.client.SetTimeout(time.Duration(w.Timeout))

	if w.whoisLookup == nil {
		w.whoisLookup = func(client *whois.Client, domain, server string) (string, error) {
			return client.Whois(domain, server)
		}
	}
	if w.parseWhoisData == nil {
		w.parseWhoisData = func(raw string) (whoisparser.WhoisInfo, error) {
			return whoisparser.Parse(raw)
		}
	}

	return nil
}

func (w *Whois) Gather(acc telegraf.Accumulator) error {
	for _, domain := range w.Domains {
		w.Log.Debugf("Fetching WHOIS data for %q using WHOIS server %q with timeout: %v", domain, w.Server, w.Timeout)

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
			if errors.Is(err, whoisparser.ErrDomainLimitExceed) || errors.Is(err, whoisparser.ErrDomainDataInvalid) {
				acc.AddError(fmt.Errorf("whois parsing failed for %q: %w", domain, err))
				continue
			}

			acc.AddFields("whois", map[string]interface{}{
				"status_code": simplifyStatus(nil, err),
			}, map[string]string{
				"domain": domain,
			})
			continue
		}

		w.Log.Debugf("Parsed WHOIS data for %s: %+v", domain, data)

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

		// Add metrics
		fields := map[string]interface{}{
			"creation_timestamp":   creationTimestamp,
			"dnssec_enabled":       data.Domain.DNSSec,
			"expiration_timestamp": expirationTimestamp,
			"expiry":               expiry,
			"updated_timestamp":    updatedTimestamp,
			"registrar":            registrar,
			"registrant":           registrant,
			"status_code":          simplifyStatus(data.Domain.Status, nil),
			"name_servers":         strings.Join(data.Domain.NameServers, ","),
		}
		tags := map[string]string{
			"domain": domain,
		}

		acc.AddFields("whois", fields, tags)
	}

	return nil
}

func simplifyStatus(statusList []string, err error) int {
	// Handle WHOIS parser errors
	if err != nil {
		if errors.Is(err, whoisparser.ErrNotFoundDomain) {
			return 6
		}
		if errors.Is(err, whoisparser.ErrReservedDomain) {
			return 7
		}
		if errors.Is(err, whoisparser.ErrPremiumDomain) {
			return 8
		}
		if errors.Is(err, whoisparser.ErrBlockedDomain) {
			return 9
		}
	}

	// Handle nil case explicitly
	if statusList == nil {
		return 0 // UNKNOWN
	}

	// Process WHOIS status strings
	for _, status := range statusList {
		switch s := strings.ToLower(status); {
		case strings.Contains(s, "pendingdelete"):
			return 1 // PENDING DELETE
		case strings.Contains(s, "redemptionperiod"):
			return 2 // EXPIRED
		case strings.Contains(s, "clienttransferprohibited"), strings.Contains(s, "clientdeleteprohibited"):
			return 3 // LOCKED
		case s == "registered":
			return 4 // REGISTERED
		case s == "active":
			return 5 // ACTIVE
		}
	}
	return 0 // UNKNOWN
}

// Plugin registration
func init() {
	inputs.Add("whois", func() telegraf.Input {
		return &Whois{
			Timeout: config.Duration(5 * time.Second),
		}
	})
}
