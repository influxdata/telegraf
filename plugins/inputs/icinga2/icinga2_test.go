package icinga2

import (
	"encoding/json"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func TestGatherStatus(t *testing.T) {

	s := `{"results":[
    {
      "attrs": {
        "check_command": "check-bgp-juniper-netconf",
        "display_name": "eq-par.dc2.fr",
        "name": "ef017af8-c684-4f3f-bb20-0dfe9fcd3dbe",
        "state": 0
      },
      "joins": {},
      "meta": {},
      "name": "eq-par.dc2.fr!ef017af8-c684-4f3f-bb20-0dfe9fcd3dbe",
      "type": "Service"
    }
  ]}`

	checks := Result{}
	json.Unmarshal([]byte(s), &checks)
	fields := map[string]interface{}{
		"name":  "ef017af8-c684-4f3f-bb20-0dfe9fcd3dbe",
		"state": 0,
	}
	tags := map[string]string{
		"display_name":  "eq-par.dc2.fr",
		"check_command": "check-bgp-juniper-netconf",
		"source":        "https://localhost:5665",
	}

	var acc testutil.Accumulator

	icinga2 := new(Icinga2)
	icinga2.ObjectType = "services"
	icinga2.Server = "https://localhost:5665"
	icinga2.GatherStatus(&acc, checks.Results)
	acc.AssertContainsTaggedFields(t, "icinga2_services_status", fields, tags)
}
