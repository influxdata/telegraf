package ifname

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/snmp"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tab, err := makeTable("IF-MIB::ifTable")
	require.NoError(t, err)

	config := snmp.ClientConfig{
		Version: 2,
		Timeout: internal.Duration{Duration: 5 * time.Second}, // doesn't work with 0 timeout
	}
	gs, err := snmp.NewWrapper(config)
	require.NoError(t, err)
	err = gs.SetAgent("127.0.0.1")
	require.NoError(t, err)

	err = gs.Connect()
	require.NoError(t, err)

	//could use ifIndex but oid index is always the same
	m, err := buildMap(&gs, tab, "ifDescr")
	require.NoError(t, err)
	require.NotEmpty(t, m)
}

func TestXTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	tab, err := makeTable("IF-MIB::ifXTable")
	require.NoError(t, err)

	config := snmp.ClientConfig{
		Version: 2,
		Timeout: internal.Duration{Duration: 5 * time.Second}, // doesn't work with 0 timeout
	}
	gs, err := snmp.NewWrapper(config)
	require.NoError(t, err)
	err = gs.SetAgent("127.0.0.1")
	require.NoError(t, err)

	err = gs.Connect()
	require.NoError(t, err)

	m, err := buildMap(&gs, tab, "ifName")
	require.NoError(t, err)
	require.NotEmpty(t, m)
}

func TestIfName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	d := IfName{
		SourceTag: "ifIndex",
		DestTag:   "ifName",
		AgentTag:  "agent",
		CacheSize: 1000,
		ClientConfig: snmp.ClientConfig{
			Version: 2,
			Timeout: internal.Duration{Duration: 5 * time.Second}, // doesn't work with 0 timeout
		},
	}
	err := d.Init()
	require.NoError(t, err)
	d.getMap = func(agent string) (nameMap, error) {
		return map[uint64]string{
			1: "lo",
		}, nil
	}

	acc := testutil.Accumulator{}
	err = d.Start(&acc)

	require.NoError(t, err)

	m := testutil.MustMetric(
		"cpu",
		map[string]string{
			"ifIndex": "1",
			"agent":   "127.0.0.1",
		},
		map[string]interface{}{},
		time.Unix(0, 0),
	)

	expected := testutil.MustMetric(
		"cpu",
		map[string]string{
			"ifIndex": "1",
			"agent":   "127.0.0.1",
			"ifName":  "lo",
		},
		map[string]interface{}{},
		time.Unix(0, 0),
	)

	err = d.addTag(m)
	require.NoError(t, err)

	testutil.RequireMetricEqual(t, expected, m)
}
