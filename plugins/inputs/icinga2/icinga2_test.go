package icinga2

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestGatherServicesStatus(t *testing.T) {
	s := `{
    "results": [
        {
            "attrs": {
                "check_command": "check-bgp-juniper-netconf",
                "display_name": "eq-par.dc2.fr",
                "host_name": "someserverfqdn.net",
                "name": "ef017af8-c684-4f3f-bb20-0dfe9fcd3dbe",
                "state": 0
            },
            "joins": {},
            "meta": {},
            "name": "eq-par.dc2.fr!ef017af8-c684-4f3f-bb20-0dfe9fcd3dbe",
            "type": "Service"
        }
    ]
}
`

	checks := Result{}
	require.NoError(t, json.Unmarshal([]byte(s), &checks))

	icinga2 := new(Icinga2)
	icinga2.Log = testutil.Logger{}
	icinga2.ObjectType = "services"
	icinga2.Server = "https://localhost:5665"

	var acc testutil.Accumulator
	icinga2.GatherStatus(&acc, checks.Results)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"icinga2_services",
			map[string]string{
				"display_name":  "eq-par.dc2.fr",
				"check_command": "check-bgp-juniper-netconf",
				"state":         "ok",
				"source":        "someserverfqdn.net",
				"server":        "localhost",
				"port":          "5665",
				"scheme":        "https",
			},
			map[string]interface{}{
				"name":       "ef017af8-c684-4f3f-bb20-0dfe9fcd3dbe",
				"state_code": 0,
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestGatherHostsStatus(t *testing.T) {
	s := `{
    "results": [
        {
            "attrs": {
                "address": "192.168.1.1",
                "check_command": "ping",
                "display_name": "apache",
                "name": "webserver",
                "state": 2.0
            },
            "joins": {},
            "meta": {},
            "name": "webserver",
            "type": "Host"
        }
    ]
}
`

	checks := Result{}
	require.NoError(t, json.Unmarshal([]byte(s), &checks))

	var acc testutil.Accumulator

	icinga2 := new(Icinga2)
	icinga2.Log = testutil.Logger{}
	icinga2.ObjectType = "hosts"
	icinga2.Server = "https://localhost:5665"

	icinga2.GatherStatus(&acc, checks.Results)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"icinga2_hosts",
			map[string]string{
				"display_name":  "apache",
				"check_command": "ping",
				"state":         "critical",
				"source":        "webserver",
				"server":        "localhost",
				"port":          "5665",
				"scheme":        "https",
			},
			map[string]interface{}{
				"name":       "webserver",
				"state_code": 2,
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}
