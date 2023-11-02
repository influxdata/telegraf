package snmp_lookup

import (
	"testing"

	"github.com/influxdata/telegraf/internal/snmp"
	"github.com/influxdata/telegraf/plugins/common/parallel"
	si "github.com/influxdata/telegraf/plugins/inputs/snmp"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name     string
		plugin   Lookup
		expected string
	}{
		{
			name: "defaults",
			plugin: Lookup{
				AgentTag:        "source",
				IndexTag:        "index",
				CacheSize:       defaultCacheSize,
				ParallelLookups: defaultParallelLookups,
				ClientConfig:    *snmp.DefaultClientConfig(),
				CacheTTL:        defaultCacheTTL,
			},
		},
		{
			name: "empty",
		},
		{
			name: "wrong SNMP client config",
			plugin: Lookup{
				ClientConfig: snmp.ClientConfig{
					Version: 99,
				},
			},
			expected: "parsing SNMP client config: invalid version",
		},
		{
			name: "netsnmp translator",
			plugin: Lookup{
				ClientConfig: snmp.ClientConfig{
					Translator: "netsnmp",
				},
			},
		},
		{
			name: "unknown translator",
			plugin: Lookup{
				ClientConfig: snmp.ClientConfig{
					Translator: "unknown",
				},
			},
			expected: `invalid agent.snmp_translator value "unknown"`,
		},
		{
			name: "table init",
			plugin: Lookup{
				Tags: []si.Field{
					{
						Name: "ifName",
						Oid:  ".1.3.6.1.2.1.31.1.1.1.1",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.plugin.Log = testutil.Logger{}

			if tt.expected == "" {
				require.NoError(t, tt.plugin.Init())
			} else {
				require.ErrorContains(t, tt.plugin.Init(), tt.expected)
			}
		})
	}
}

func TestStart(t *testing.T) {
	acc := &testutil.NopAccumulator{}
	p := Lookup{}
	require.NoError(t, p.Init())

	p.Ordered = true
	require.NoError(t, p.Start(acc))
	require.IsType(t, &parallel.Ordered{}, p.parallel)
	p.Stop()

	p.Ordered = false
	require.NoError(t, p.Start(acc))
	require.IsType(t, &parallel.Unordered{}, p.parallel)
	p.Stop()
}

func TestRegistry(t *testing.T) {
	require.Contains(t, processors.Processors, "snmp_lookup")
	require.IsType(t, &Lookup{}, processors.Processors["snmp_lookup"]())
}
