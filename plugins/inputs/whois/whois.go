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
				"status": 0, // Mark failure
			}, map[string]string{
				"domain":        domain,
				"domain_status": "UNKNOWN",
			})
			continue
		}

		// Extract expiration date
		expiration := parsedWhois.Domain.ExpirationDateInTime

		// Extract creation date
		created := parsedWhois.Domain.CreatedDateInTime

		// Extract updated date
		updated := parsedWhois.Domain.UpdatedDateInTime

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
		domainStatus := "UNKNOWN"
		if parsedWhois.Domain.Status != nil {
			domainStatus = simplifyStatus(parsedWhois.Domain.Status)
		}

		// Calculate expiry in seconds
		expiry := int(expiration.Sub(time.Now()).Seconds())

		// Add metrics
		fields := map[string]interface{}{
			"creation_timestamp":   created.Unix(),
			"dnssec_status_code":   dnssec,
			"expiration_timestamp": expiration.Unix(),
			"expiry":               expiry,
			"updated_timestamp":    updated.Unix(),
			"registrar":            registrar,
			"registrant":           registrant,
			"domain_status":        domainStatus,
			"status":               1,
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
func simplifyStatus(statusList []string) string {
	for _, status := range statusList {
		switch s := strings.ToLower(status); {
		case strings.Contains(s, "pendingdelete"):
			return "PENDING DELETE"
		case strings.Contains(s, "redemptionperiod"):
			return "EXPIRED"
		case strings.Contains(s, "clienttransferprohibited"), strings.Contains(s, "clientdeleteprohibited"):
			return "LOCKED"
		case s == "registered":
			return "REGISTERED"
		case s == "active":
			return "ACTIVE"
		}
	}
	return "UNKNOWN"
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
