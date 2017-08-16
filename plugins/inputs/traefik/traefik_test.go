package traefik

import (
	"encoding/json"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func TestGatherHealthCheck(t *testing.T) {

	s := `{
    "pid": 1,
    "uptime": "1m53.450952875s",
    "uptime_sec": 113.450952875,
    "time": "2017-04-14 09:32:00.350042707 +0000 UTC",
    "unixtime": 1492162320,
    "status_code_count": {},
    "total_status_code_count": {
      "200": 7,
      "404": 6
    },
    "count": 0,
    "total_count": 13,
    "total_response_time": "15.202713ms",
    "total_response_time_sec": 0.015202713,
    "average_response_time": "1.169439ms",
    "average_response_time_sec": 0.001169439
  }`

	check := HealthCheck{}

	json.Unmarshal([]byte(s), &check)

	records := map[string]interface{}{
		"total_response_time_sec":   0.015202713,
		"average_response_time_sec": 0.001169439,
		"total_count":               13,
		"200":                       7,
		"404":                       6,
	}
	tags := map[string]string{
		"instance": "default",
	}

	var acc testutil.Accumulator

	traefik := new(Traefik)
	traefik.Instance = "default"
	traefik.GatherHealthCheck(&acc, check)

	acc.AssertContainsTaggedFields(t, "traefik_healthchecks", records, tags)
}
