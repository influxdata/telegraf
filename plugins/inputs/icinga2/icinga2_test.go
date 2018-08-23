package icinga2

import (
	"encoding/json"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func TestGatherServicesStatus(t *testing.T) {

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
		"name":       "ef017af8-c684-4f3f-bb20-0dfe9fcd3dbe",
		"state_code": 0,
	}
	tags := map[string]string{
		"display_name":  "eq-par.dc2.fr",
		"check_command": "check-bgp-juniper-netconf",
		"state":         "ok",
		"source":        "localhost",
		"port":          "5665",
		"scheme":        "https",
	}

	var acc testutil.Accumulator

	icinga2 := new(Icinga2)
	icinga2.ObjectType = "services"
	icinga2.Server = "https://localhost:5665"
	icinga2.GatherStatus(&acc, checks.Results)
	acc.AssertContainsTaggedFields(t, "icinga2_services", fields, tags)
}

func TestGatherHostsStatus(t *testing.T) {

	s := `{"results":[
    {
      "attrs": {
				"name": "webserver",
        "address": "192.168.1.1",
        "check_command": "ping",
        "display_name": "apache",
        "state": 2
      },
      "joins": {},
      "meta": {},
      "name": "webserver",
      "type": "Host"
    }
  ]}`

	checks := Result{}
	json.Unmarshal([]byte(s), &checks)
	fields := map[string]interface{}{
		"name":       "webserver",
		"state_code": 2,
	}
	tags := map[string]string{
		"display_name":  "apache",
		"check_command": "ping",
		"state":         "critical",
		"source":        "localhost",
		"port":          "5665",
		"scheme":        "https",
	}

	var acc testutil.Accumulator

	icinga2 := new(Icinga2)
	icinga2.ObjectType = "hosts"
	icinga2.Server = "https://localhost:5665"
	icinga2.GatherStatus(&acc, checks.Results)
	acc.AssertContainsTaggedFields(t, "icinga2_hosts", fields, tags)
}
