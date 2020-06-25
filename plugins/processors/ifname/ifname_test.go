//+build localsnmp

package ifname

import (
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/snmp"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func _TestTable(t *testing.T) {
	tab, err := makeTable("IF-MIB::ifTable")
	require.NoError(t, err)

	config := snmp.ClientConfig{
		Version: 2,
		Timeout: internal.Duration{Duration: 5 * time.Second}, // doesn't work with 0 timeout
	}
	gs, err := snmp.NewWrapper(config, "127.0.0.1")
	require.NoError(t, err)

	err = gs.Connect()
	require.NoError(t, err)

	//could use ifIndex but oid index is always the same
	m, err := buildMap(&gs, tab, "ifDescr")
	require.NoError(t, err)

	for index, name := range m {
		fmt.Println(index, name)
	}
}

func _TestXTable(t *testing.T) {
	tab, err := makeTable("IF-MIB::ifXTable")
	require.NoError(t, err)

	config := snmp.ClientConfig{
		Version: 2,
		Timeout: internal.Duration{Duration: 5 * time.Second}, // doesn't work with 0 timeout
	}
	gs, err := snmp.NewWrapper(config, "127.0.0.1")
	require.NoError(t, err)

	err = gs.Connect()
	require.NoError(t, err)

	m, err := buildMap(&gs, tab, "ifName")
	require.NoError(t, err)

	for index, name := range m {
		fmt.Println(index, name)
	}
}

func TestIfName(t *testing.T) {
	d := IfName{
		SourceTag: "ifIndex",
		DestTag:   "ifName",
		AgentTag:  "agent",
		ClientConfig: snmp.ClientConfig{
			Version: 2,
			Timeout: internal.Duration{Duration: 5 * time.Second}, // doesn't work with 0 timeout
		},
	}
	err := d.Init()
	require.NoError(t, err)

	in := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"ifIndex": "1",
				"agent":   "127.0.0.1",
			},
			map[string]interface{}{},
			time.Unix(0, 0),
		),
	}

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"ifIndex": "1",
				"agent":   "127.0.0.1",
				"ifName":  "lo",
			},
			map[string]interface{}{},
			time.Unix(0, 0),
		),
	}

	out := d.Apply(in...)
	require.Len(t, out, 1)

	testutil.RequireMetricsEqual(t, expected, out)
}
