package whois

import (
	"fmt"
	// "strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/likexian/whois"
	"github.com/likexian/whois-parser"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

// Make sure Whois implements telegraf.Input
var _ telegraf.Input = &Whois{}

func ptr(t time.Time) *time.Time {
	return &t
}

func TestSimplifyStatus(t *testing.T) {
	tests := []struct {
		input    []string
		err      error
		expected int
	}{
		// WHOIS status strings
		{[]string{"clientTransferProhibited"}, nil, 3},
		{[]string{"pendingDelete"}, nil, 1},
		{[]string{"redemptionPeriod"}, nil, 2},
		{[]string{"active"}, nil, 5},
		{[]string{"registered"}, nil, 4},
		{[]string{"unknownStatus"}, nil, 0},

		// WHOIS error cases
		{nil, whoisparser.ErrNotFoundDomain, 6},
		{nil, whoisparser.ErrReservedDomain, 7},
		{nil, whoisparser.ErrPremiumDomain, 8},
		{nil, whoisparser.ErrBlockedDomain, 9},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("status_%d", tt.expected), func(t *testing.T) {
			result := simplifyStatus(tt.input, tt.err)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestWhoisConfigInitialization(t *testing.T) {
	tests := []struct {
		name               string
		domains            []string
		server             string
		IncludeNameServers bool
		timeout            config.Duration
		expectErr          bool
	}{
		{
			name:      "Valid Configuration",
			domains:   []string{"example.com", "google.com"},
			timeout:   config.Duration(10 * time.Second),
			expectErr: false,
		},
		{
			name:      "No Domains Configured",
			domains:   nil,
			timeout:   config.Duration(5 * time.Second),
			expectErr: true,
		},
		{
			name:      "Invalid Timeout (Zero Value)",
			domains:   []string{"example.com"},
			timeout:   config.Duration(0),
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plugin := &Whois{
				Domains: test.domains,
				Timeout: test.timeout,
				Server:  test.server,
				Log:     testutil.Logger{},
			}

			err := plugin.Init()

			if test.expectErr {
				require.Error(t, err, "Expected error but got none")
				return
			}

			require.NoError(t, err, "Unexpected error during Init()")
		})
	}
}

func TestWhoisGatherStaticMockResponses(t *testing.T) {
	expirationTimestamp := time.Now().Add(60 * 24 * time.Hour).Unix()

	plugin := &Whois{
		Domains: []string{"example.com"},
		Timeout: config.Duration(5 * time.Second),
		Log:     testutil.Logger{},
		client:  whois.NewClient(),
		whoisLookup: func(_ *whois.Client, domain string, _ string) (string, error) {
			return "WHOIS mock response for " + domain, nil
		},
		parseWhoisData: func(raw string) (whoisparser.WhoisInfo, error) {
			mockResponses := map[string]whoisparser.WhoisInfo{
				"WHOIS mock response for example.com": {
					Domain: &whoisparser.Domain{
						ExpirationDateInTime: ptr(time.Unix(expirationTimestamp, 0)),
						CreatedDateInTime:    ptr(time.Unix(1609459200, 0)),
						UpdatedDateInTime:    ptr(time.Unix(1680307200, 0)),
						Status:               []string{"clientTransferProhibited"},
						NameServers:          []string{"ns1.example.com", "ns2.example.com"},
					},
					Registrar: &whoisparser.Contact{
						Name: "RESERVED-Internet Assigned Numbers Authority",
					},
				},
			}

			if info, found := mockResponses[raw]; found {
				return info, nil
			}

			return whoisparser.WhoisInfo{}, whoisparser.ErrNotFoundDomain
		},
	}

	require.NoError(t, plugin.Init(), "Unexpected error during Init()")
	acc := &testutil.Accumulator{}

	err := plugin.Gather(acc)
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"whois",
			map[string]string{
				"domain": "example.com",
			},
			map[string]interface{}{
				"status_code":          3, // LOCKED
				"creation_timestamp":   1609459200,
				"updated_timestamp":    1680307200,
				"expiration_timestamp": expirationTimestamp,
				"expiry":               int64(86399),
				"registrar":            "RESERVED-Internet Assigned Numbers Authority",
				"name_servers":         "ns1.example.com,ns2.example.com",
				"dnssec_enabled":       false,
				"registrant":           "",
			},
			time.Unix(0, 0),
		),
	}

	// Validate expected vs actual metrics
	opts := []cmp.Option{
		testutil.SortMetrics(),
		testutil.IgnoreTime(),
		testutil.IgnoreFields("expiry"),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), opts...)
}

// Test WHOIS Handling for an Invalid Domain
func TestWhoisGatherInvalidDomain(t *testing.T) {
	plugin := &Whois{
		Domains: []string{"invalid-domain.xyz"},
		Timeout: config.Duration(5 * time.Second),
		Log:     testutil.Logger{},
		client:  whois.NewClient(),
		whoisLookup: func(_ *whois.Client, _ string, _ string) (string, error) {
			return "WHOIS mock response for invalid-domain.xyz", nil
		},
		parseWhoisData: func(_ string) (whoisparser.WhoisInfo, error) {
			return whoisparser.WhoisInfo{}, whoisparser.ErrNotFoundDomain
		},
	}

	require.NoError(t, plugin.Init(), "Unexpected error during Init()")
	acc := &testutil.Accumulator{}

	err := plugin.Gather(acc)
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"whois",
			map[string]string{
				"domain": "invalid-domain.xyz",
			},
			map[string]interface{}{
				"status_code": 6, // Expecting "ErrNotFoundDomain"
			},
			time.Time{},
		),
	}

	// Validate expected vs actual metrics
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}
