package ifname

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/snmp"
	si "github.com/influxdata/telegraf/plugins/inputs/snmp"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
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
		Timeout: internal.Duration{Duration: 5 * time.Second}, // doesn't work with 0 timeout
	}
	gs, err := snmp.NewWrapper(config)
	require.NoError(t, err)
	err = gs.SetAgent("127.0.0.1")
	require.NoError(t, err)

	err = gs.Connect()
	require.NoError(t, err)

	//could use ifIndex but oid index is always the same
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
			Timeout: internal.Duration{Duration: 5 * time.Second}, // doesn't work with 0 timeout
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
		SourceTag: "ifIndex",
		DestTag:   "ifName",
		AgentTag:  "agent",
		CacheSize: 1000,
		ClientConfig: snmp.ClientConfig{
			Version: 2,
			Timeout: internal.Duration{Duration: 5 * time.Second}, // doesn't work with 0 timeout
		},
		CacheTTL: config.Duration(10 * time.Second),
	}

	// This test mocks the snmp transaction so don't run net-snmp
	// commands to look up table names.
	d.makeTable = func(agent string) (*si.Table, error) {
		return &si.Table{}, nil
	}
	err := d.Init()
	require.NoError(t, err)

	// Request the same agent multiple times in goroutines. The first
	// request should make the mocked remote call and the others
	// should block until the response is cached, then return the
	// cached response.

	expected := nameMap{
		1: "ifname1",
		2: "ifname2",
	}

	var wgRemote sync.WaitGroup
	var remoteCalls int32

	wgRemote.Add(1)
	d.getMapRemote = func(agent string) (nameMap, error) {
		atomic.AddInt32(&remoteCalls, 1)
		wgRemote.Wait() //don't return until all requests are made
		return expected, nil
	}

	const thMax = 3
	var wgReq sync.WaitGroup

	for th := 0; th < thMax; th++ {
		wgReq.Add(1)
		go func() {
			defer wgReq.Done()
			m, _, err := d.getMap("agent")
			require.NoError(t, err)
			require.Equal(t, expected, m)
		}()
	}

	//signal mocked remote call to finish
	wgRemote.Done()

	//wait for requests to finish
	wgReq.Wait()

	//remote call should only happen once
	require.Equal(t, int32(1), remoteCalls)

}
