package ifname

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/snmp"
	si "github.com/influxdata/telegraf/plugins/inputs/snmp"
	"github.com/influxdata/telegraf/testutil"
)

func TestTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	d := IfName{}
	d.Init()
	tab, err := d.makeTable("IF-MIB::ifTable")
	require.NoError(t, err)

	config := snmp.ClientConfig{
		Version: 2,
		Timeout: internal.Duration{Duration: 5 * time.Second}, // Doesn't work with 0 timeout
	}
	gs, err := snmp.NewWrapper(config)
	require.NoError(t, err)
	err = gs.SetAgent("127.0.0.1")
	require.NoError(t, err)

	err = gs.Connect()
	require.NoError(t, err)

	// Could use ifIndex but oid index is always the same
	m, err := buildMap(gs, tab, "ifDescr")
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
			Timeout: internal.Duration{Duration: 5 * time.Second}, // Doesn't work with 0 timeout
		},
	}
	err := d.Init()
	require.NoError(t, err)

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

func TestGetMap(t *testing.T) {
	d := IfName{
		CacheSize: 1000,
		CacheTTL:  config.Duration(10 * time.Second),
	}

	// Don't run net-snmp commands to look up table names.
	d.makeTable = func(agent string) (*si.Table, error) {
		return &si.Table{}, nil
	}
	err := d.Init()
	require.NoError(t, err)

	expected := nameMap{
		1: "ifname1",
		2: "ifname2",
	}

	var remoteCalls int32

	// Mock the snmp transaction
	d.getMapRemote = func(agent string) (nameMap, error) {
		atomic.AddInt32(&remoteCalls, 1)
		return expected, nil
	}
	m, age, err := d.getMap("agent")
	require.NoError(t, err)
	require.Zero(t, age) // Age is zero when map comes from getMapRemote
	require.Equal(t, expected, m)

	// Remote call should happen the first time getMap runs
	require.Equal(t, int32(1), remoteCalls)

	var wg sync.WaitGroup
	const thMax = 3
	for th := 0; th < thMax; th++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m, age, err := d.getMap("agent")
			require.NoError(t, err)
			require.NotZero(t, age) // Age is nonzero when map comes from cache
			require.Equal(t, expected, m)
		}()
	}

	wg.Wait()

	// Remote call should not happen subsequent times getMap runs
	require.Equal(t, int32(1), remoteCalls)
}
