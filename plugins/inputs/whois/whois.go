package whois

import (
	_ "embed"
	"errors"
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
	Domains            []string        `toml:"domains"`
	Server             string          `toml:"server"`
	Timeout            config.Duration `toml:"timeout"`
	IncludeNameServers bool            `toml:"include_name_servers"`
	Log                telegraf.Logger `toml:"-"`

	Client         *whois.Client
	WhoisLookup    func(client *whois.Client, domain, server string) (string, error)
	ParseWhoisData func(raw string) (whoisparser.WhoisInfo, error)
}

func (*Whois) SampleConfig() string {
	return sampleConfig
}

func (w *Whois) Gather(acc telegraf.Accumulator) error {
	now := time.Now()

	if w.Client == nil {
		if err := w.Init(); err != nil {
			return err
		}
	}

	for _, domain := range w.Domains {
		w.Log.Debugf("Fetching WHOIS data for: %s", domain)
		w.Log.Debugf("Using WHOIS server: %s with timeout: %v", w.Server, w.Timeout)

		// Fetch WHOIS raw data
		rawWhois, err := w.WhoisLookup(w.Client, domain, w.Server)
		if err != nil {
			w.Log.Errorf("WHOIS query failed for %s: %v", domain, err)
			acc.AddError(err)

			// Always register a metric, even on failure
			acc.AddFields("whois", map[string]interface{}{
				"status": 0, // Mark failure
			}, map[string]string{
				"domain":        domain,
				"domain_status": "UNKNOWN",
			})
			continue
		}

		// Parse WHOIS data using whois-parser
		parsedWhois, err := w.ParseWhoisData(rawWhois)
		if err != nil {
			w.Log.Errorf("WHOIS parsing failed for %s: %v", domain, err)
			acc.AddError(err)

			// Always register a metric, even on failure
			acc.AddFields("whois", map[string]interface{}{
				"status": 0, // Mark failure
			}, map[string]string{
				"domain":        domain,
				"domain_status": "UNKNOWN",
			})
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
		if expiration == nil {
			w.Log.Warnf("No expiration date found for %s", domain)
			continue
		}

		// Extract creation date
		created := parsedWhois.Domain.CreatedDateInTime
		if created == nil {
			w.Log.Warnf("No created date found for %s", domain)
			continue
		}

		// Extract updated date
		updated := parsedWhois.Domain.UpdatedDateInTime
		if updated == nil {
			w.Log.Warnf("No updated date found for %s", domain)
			continue
		}

		// Extract DNSSEC status
		dnssec := parsedWhois.Domain.DNSSec

		// Extract NameServers status
		nameServers := parsedWhois.Domain.NameServers
		if len(nameServers) == 0 {
			w.Log.Warnf("No name servers found for %s", domain)
			continue
		}

		// Extract registrar name (handle nil)
		registrar := ""
		if parsedWhois.Registrar != nil {
			registrar = parsedWhois.Registrar.Name
		}

		// Extract status (handle nil)
		domainStatus := "UNKNOWN"
		if parsedWhois.Domain.Status != nil {
			domainStatus = simplifyStatus(parsedWhois.Domain.Status)
		}

		// Calculate expiry in seconds
		expiry := int(expiration.Sub(now).Seconds())

		// Add metrics
		fields := map[string]interface{}{
			"creation_timestamp":   created.Unix(),
			"dnssec_enabled":       boolToInt(dnssec),
			"expiration_timestamp": expiration.Unix(),
			"expiry":               expiry,
			"updated_timestamp":    updated.Unix(),
			"registrar":            registrar,
			"domain_status":        domainStatus,
			"status":               1,
		}
		tags := map[string]string{
			"domain": domain,
		}

		if w.IncludeNameServers {
			fields["name_servers"] = strings.Join(nameServers, ",")
		}

		acc.AddFields("whois", fields, tags)
	}

	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
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

	w.Client = whois.NewClient()
	w.Client.SetTimeout(time.Duration(w.Timeout))

	if w.WhoisLookup == nil {
		w.WhoisLookup = func(client *whois.Client, domain, server string) (string, error) {
			return client.Whois(domain, server)
		}
	}
	if w.ParseWhoisData == nil {
		w.ParseWhoisData = func(raw string) (whoisparser.WhoisInfo, error) {
			return whoisparser.Parse(raw)
		}
	}

	return nil
}

// Plugin registration
func init() {
	inputs.Add("whois", func() telegraf.Input {
		return &Whois{
			IncludeNameServers: true,
			Timeout:            config.Duration(5 * time.Second),
		}
	})
}
