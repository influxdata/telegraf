//go:generate ../../../tools/readme_config_includer/generator

package rdap

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/openrdap/rdap"
	"golang.org/x/net/idna"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

const maxDomainLength = 253

type RDAP struct {
	Domains []string        `toml:"domains"`
	Server  string          `toml:"server"`
	Timeout config.Duration `toml:"timeout"`
	Log     telegraf.Logger `toml:"-"`

	client *rdap.Client
	server *url.URL
}

func (*RDAP) SampleConfig() string {
	return sampleConfig
}

func (r *RDAP) Init() error {
	if len(r.Domains) == 0 {
		return errors.New("no domains configured")
	}

	if r.Timeout <= 0 {
		return errors.New("timeout has to be greater than zero")
	}

	if r.Server != "" {
		u, err := url.Parse(r.Server)
		if err != nil {
			return fmt.Errorf("parsing server URL failed: %w", err)
		}
		if u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("server %q is not a valid URL", r.Server)
		}
		r.server = u
	}

	r.client = &rdap.Client{UserAgent: "Telegraf/rdap"}

	return nil
}

var asciiDomainRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,63}$`)

func isValidDomain(domain string) bool {
	if len(domain) > maxDomainLength {
		return false
	}

	if asciiDomainRegex.MatchString(domain) {
		return true
	}

	// Convert internationalized domain names to Punycode before checking
	p := idna.New(idna.MapForLookup(), idna.StrictDomainName(true))
	ascii, err := p.ToASCII(domain)
	if err != nil {
		return false
	}

	return asciiDomainRegex.MatchString(ascii)
}

func (r *RDAP) Gather(acc telegraf.Accumulator) error {
	for _, domain := range r.Domains {
		if !isValidDomain(domain) {
			acc.AddError(fmt.Errorf("invalid domain format: %q", domain))
			continue
		}

		data, status, err := r.query(domain)
		if err != nil {
			// A missing object still produces a metric so dashboards can
			// alert on it, everything else is a hard error.
			if status == "not found" {
				acc.AddFields(
					"rdap",
					map[string]interface{}{"error": err.Error()},
					map[string]string{"domain": domain, "status": status},
				)
				continue
			}
			acc.AddError(fmt.Errorf("rdap query failed for %q: %w", domain, err))
			continue
		}

		var creation, expiration, updated int64
		for _, event := range data.Events {
			ts, err := time.Parse(time.RFC3339, event.Date)
			if err != nil {
				continue
			}
			switch strings.ToLower(event.Action) {
			case "registration":
				creation = ts.Unix()
			case "expiration":
				expiration = ts.Unix()
			case "last changed":
				updated = ts.Unix()
			}
		}

		var expiry int64
		if expiration > 0 {
			expiry = expiration - time.Now().Unix()
		}

		fields := map[string]interface{}{
			"creation_timestamp":   creation,
			"dnssec_enabled":       dnssecEnabled(data),
			"expiration_timestamp": expiration,
			"expiry":               expiry,
			"updated_timestamp":    updated,
			"registrar":            entityName(data, "registrar"),
			"registrant":           entityName(data, "registrant"),
			"name_servers":         nameservers(data),
		}
		tags := map[string]string{
			"domain": domain,
			"status": status,
		}

		acc.AddFields("rdap", fields, tags)
	}

	return nil
}

func (r *RDAP) query(domain string) (*rdap.Domain, string, error) {
	req := &rdap.Request{
		Type:  rdap.DomainRequest,
		Query: domain,
	}
	if r.server != nil {
		req.Server = r.server
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.Timeout))
	defer cancel()
	req = req.WithContext(ctx)

	r.Log.Tracef("Fetching RDAP data for %q", domain)

	resp, err := r.client.Do(req)
	if err != nil {
		var cerr *rdap.ClientError
		if errors.As(err, &cerr) && cerr.Type == rdap.ObjectDoesNotExist {
			return nil, "not found", err
		}
		return nil, "unknown", err
	}

	data, ok := resp.Object.(*rdap.Domain)
	if !ok {
		return nil, "unknown", fmt.Errorf("unexpected response type %T", resp.Object)
	}

	status := "unknown"
	if len(data.Status) > 0 {
		status = strings.Join(data.Status, ",")
	}

	return data, status, nil
}

func dnssecEnabled(d *rdap.Domain) bool {
	if d.SecureDNS == nil {
		return false
	}
	return d.SecureDNS.DelegationSigned != nil && *d.SecureDNS.DelegationSigned
}

// entityName returns the vCard "fn" (falling back to "org") of the first
// entity carrying the requested role.
func entityName(d *rdap.Domain, role string) string {
	for i := range d.Entities {
		e := d.Entities[i]
		for _, have := range e.Roles {
			if !strings.EqualFold(have, role) {
				continue
			}
			if e.VCard == nil {
				return "not set"
			}
			if name := e.VCard.Name(); name != "" {
				return name
			}
			if org := e.VCard.Org(); org != "" {
				return org
			}
			return "not set"
		}
	}
	return "not set"
}

func nameservers(d *rdap.Domain) string {
	names := make([]string, 0, len(d.Nameservers))
	for _, ns := range d.Nameservers {
		if ns.LDHName != "" {
			names = append(names, strings.ToLower(ns.LDHName))
		}
	}
	return strings.Join(names, ",")
}

func init() {
	inputs.Add("rdap", func() telegraf.Input {
		return &RDAP{
			Timeout: config.Duration(30 * time.Second),
		}
	})
}
