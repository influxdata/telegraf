package whois

import (
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/likexian/whois"
	"github.com/likexian/whois-parser"
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

func (w *Whois) Gather(acc telegraf.Accumulator) error {
	for _, domain := range w.Domains {
		w.Log.Debugf("Fetching WHOIS data for %q using WHOIS server %q with timeout: %v", domain, w.Server, w.Timeout)

		// Fetch WHOIS raw data
		rawWhois, err := w.whoisLookup(w.client, domain, w.Server)
		if err != nil {
			acc.AddError(fmt.Errorf("whois query failed for %q: %w", domain, err))
			continue
		}

		// Parse WHOIS data using whois-parser
		parsedWhois, err := w.parseWhoisData(rawWhois)
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

		w.Log.Debugf("Parsed WHOIS data for %s: %+v", domain, parsedWhois)

		// Extract expiration date
		var expirationTimestamp int64
		var expiry int
		if parsedWhois.Domain.ExpirationDateInTime != nil {
			expirationTimestamp = parsedWhois.Domain.ExpirationDateInTime.Unix()

			// Calculate expiry in seconds
			expiry = int(time.Until(*parsedWhois.Domain.ExpirationDateInTime).Seconds())
		}

		// Extract creation date
		var creationTimestamp int64
		if parsedWhois.Domain.CreatedDateInTime != nil {
			creationTimestamp = parsedWhois.Domain.CreatedDateInTime.Unix()
		}

		// Extract updated date
		var updatedTimestamp int64
		if parsedWhois.Domain.UpdatedDateInTime != nil {
			updatedTimestamp = parsedWhois.Domain.UpdatedDateInTime.Unix()
		}

		// Extract registrar name (handle nil)
		registrar := ""
		if parsedWhois.Registrar != nil {
			registrar = parsedWhois.Registrar.Name
		}

		// Extract registrant name (handle nil)
		registrant := ""
		if parsedWhois.Registrant != nil {
			registrant = parsedWhois.Registrant.Name
		}

		// Add metrics
		fields := map[string]interface{}{
			"creation_timestamp":   creationTimestamp,
			"dnssec_enabled":       parsedWhois.Domain.DNSSec,
			"expiration_timestamp": expirationTimestamp,
			"expiry":               expiry,
			"updated_timestamp":    updatedTimestamp,
			"registrar":            registrar,
			"registrant":           registrant,
			"status_code":          simplifyStatus(parsedWhois.Domain.Status, nil),
			"name_servers":         strings.Join(parsedWhois.Domain.NameServers, ","),
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

// Plugin registration
func init() {
	inputs.Add("whois", func() telegraf.Input {
		return &Whois{
			Timeout: config.Duration(5 * time.Second),
		}
	})
}
