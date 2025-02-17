package whois

import (
	"fmt"
	"strings"
	"testing"
	"time"

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
	plugin := &Whois{
		Domains: []string{"example.com"},
		Timeout: config.Duration(5 * time.Second),
		Log:     testutil.Logger{},
	}

	require.NoError(t, plugin.Init(), "Unexpected error during Init()")
	acc := &testutil.Accumulator{}

	// Static mock WHOIS responses
	mockResponses := map[string]whoisparser.WhoisInfo{
		"example.com": {
			Domain: &whoisparser.Domain{
				ExpirationDateInTime: ptr(time.Unix(1755057600, 0)),
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

	plugin.whoisLookup = func(_ *whois.Client, domain string, _ string) (string, error) {
		return "WHOIS mock response for " + domain, nil
	}

	plugin.parseWhoisData = func(raw string) (whoisparser.WhoisInfo, error) {
		for domain, info := range mockResponses {
			if strings.Contains(raw, domain) { // Match requested domain
				return info, nil
			}
		}

		return whoisparser.WhoisInfo{}, whoisparser.ErrNotFoundDomain
	}

	require.NoError(t, plugin.Gather(acc))

	require.Equal(t, "example.com", acc.TagValue("whois", "domain"))
	domainStatus, found := acc.IntField("whois", "status_code")
	require.True(t, found, "Expected field status_code not found")
	require.Equal(t, int(3), domainStatus, "Expected status_code field mismatch")

	// Validate `expiration_timestamp` field (2025-08-13T04:00:00Z → Unix)
	expectedExpiration := int64(1755057600)
	expirationValue, found := acc.Int64Field("whois", "expiration_timestamp")
	require.True(t, found, "expiration_timestamp field missing")
	require.InDelta(t, expectedExpiration, expirationValue, 10)

	// Validate `expiry` field
	now := time.Now()
	expectedExpiry := int(expectedExpiration - now.Unix())
	expiryValue, found := acc.IntField("whois", "expiry")
	require.True(t, found, "expiry field missing")
	require.InDelta(t, expectedExpiry, expiryValue, 10) // Allow small delta due to execution time
}

// Test WHOIS Handling for an Invalid Domain
func TestWhoisGatherInvalidDomain(t *testing.T) {
	plugin := &Whois{
		Domains: []string{"invalid-domain.xyz"},
		Timeout: config.Duration(5 * time.Second),
		Log:     testutil.Logger{},
	}

	require.NoError(t, plugin.Init(), "Unexpected error during Init()")
	acc := &testutil.Accumulator{}

	// Mock `whoisLookup` to return a valid WHOIS response (not an error)
	plugin.whoisLookup = func(_ *whois.Client, _ string, _ string) (string, error) {
		return "WHOIS mock response for invalid-domain.xyz", nil
	}

	// Mock `parseWhoisData` to return ErrNotFoundDomain (simulating WHOIS parser failure)
	plugin.parseWhoisData = func(_ string) (whoisparser.WhoisInfo, error) {
		return whoisparser.WhoisInfo{}, whoisparser.ErrNotFoundDomain
	}

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
