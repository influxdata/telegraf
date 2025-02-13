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
	if w.client == nil {
		if err := w.Init(); err != nil {
			return err
		}
	}

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
			acc.AddError(fmt.Errorf("whois parsing failed for %q: %w", domain, err))
			continue
		}

		w.Log.Debugf("Parsed WHOIS data for %s: %+v", domain, parsedWhois)

		// Prevent nil pointer panic
		if parsedWhois.Domain == nil {
			w.Log.Warnf("No domain info found for %s", domain)

			// Always register a metric, even on failure
			acc.AddFields("whois", map[string]interface{}{
				"status_code": 0,
			}, map[string]string{
				"domain": domain,
			})
			continue
		}

		// Extract expiration date
		var expirationTimestamp int64
		var expiry int
		expiration := parsedWhois.Domain.ExpirationDateInTime
		if expiration != nil {
			expirationTimestamp = expiration.Unix()

			// Calculate expiry in seconds
			expiry = int(time.Until(*expiration).Seconds())
		}

		// Extract creation date
		var creationTimestamp int64
		created := parsedWhois.Domain.CreatedDateInTime
		if created != nil {
			creationTimestamp = created.Unix()
		}

		// Extract updated date
		var updatedTimestamp int64
		updated := parsedWhois.Domain.UpdatedDateInTime
		if updated != nil {
			updatedTimestamp = updated.Unix()
		}

		// Extract DNSSEC status
		dnssec := parsedWhois.Domain.DNSSec

		// Extract NameServers status
		nameServers := parsedWhois.Domain.NameServers

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

		// Extract status (handle nil)
		domainStatus := 0
		if parsedWhois.Domain.Status != nil {
			domainStatus = simplifyStatus(parsedWhois.Domain.Status)
		}

		// Add metrics
		fields := map[string]interface{}{
			"creation_timestamp":   creationTimestamp,
			"dnssec_enabled":       dnssec,
			"expiration_timestamp": expirationTimestamp,
			"expiry":               expiry,
			"updated_timestamp":    updatedTimestamp,
			"registrar":            registrar,
			"registrant":           registrant,
			"status_code":          domainStatus,
			"name_servers":         strings.Join(nameServers, ","),
		}
		tags := map[string]string{
			"domain": domain,
		}

		acc.AddFields("whois", fields, tags)
	}

	return nil
}

// simplifyStatus maps raw WHOIS statuses to a simpler classification
func simplifyStatus(statusList []string) int {
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
