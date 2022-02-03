package snmp

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestTrapLookup(t *testing.T) {
	tests := []struct {
		name     string
		oid      string
		expected MibEntry
	}{
		{
			name: "Known trap OID",
			oid:  ".1.3.6.1.6.3.1.1.5.1",
			expected: MibEntry{
				MibName: "TGTEST-MIB",
				OidText: "coldStart",
			},
		},
		{
			name: "Known trap value OID",
			oid:  ".1.3.6.1.2.1.1.3.0",
			expected: MibEntry{
				MibName: "TGTEST-MIB",
				OidText: "sysUpTimeInstance",
			},
		},
		{
			name: "Unknown enterprise sub-OID",
			oid:  ".1.3.6.1.4.1.0.1.2.3",
			expected: MibEntry{
				MibName: "TGTEST-MIB",
				OidText: "enterprises.0.1.2.3",
			},
		},
		{
			name:     "Unknown MIB",
			oid:      ".1.2.3",
			expected: MibEntry{OidText: "iso.2.3"},
		},
	}

	// Load the MIBs
	require.NoError(t, LoadMibsFromPath([]string{"testdata/mibs"}, testutil.Logger{}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the actual test
			actual, err := TrapLookup(tt.oid)
			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestTrapLookupFail(t *testing.T) {
	tests := []struct {
		name     string
		oid      string
		expected string
	}{
		{
			name:     "New top level OID",
			oid:      ".3.6.1.3.0",
			expected: "Could not find node for OID 3.6.1.3.0",
		},
		{
			name:     "Malformed OID",
			oid:      ".1.3.dod.1.3.0",
			expected: "could not convert OID .1.3.dod.1.3.0: strconv.ParseUint: parsing \"dod\": invalid syntax",
		},
	}

	// Load the MIBs
	require.NoError(t, LoadMibsFromPath([]string{"testdata/mibs"}, testutil.Logger{}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the actual test
			_, err := TrapLookup(tt.oid)
			require.EqualError(t, err, tt.expected)
		})
	}
}
