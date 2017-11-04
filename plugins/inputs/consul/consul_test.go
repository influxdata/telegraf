package consul

import (
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/influxdata/telegraf/testutil"
)

var sampleChecks = []*api.HealthCheck{
	&api.HealthCheck{
		Node:        "localhost",
		CheckID:     "foo.health123",
		Name:        "foo.health",
		Status:      "passing",
		Notes:       "lorem ipsum",
		Output:      "OK",
		ServiceID:   "foo.123",
		ServiceName: "foo",
	},
}

func TestGatherHealthCheck(t *testing.T) {
	expectedFields := map[string]interface{}{
		"check_name": "foo.health",
		"status":     "passing",
		"passing":    1,
		"critical":   0,
		"warning":    0,
		"service_id": "foo.123",
	}

	expectedTags := map[string]string{
		"node":         "localhost",
		"service_name": "foo",
		"check_id":     "foo.health123",
	}

	var acc testutil.Accumulator

	consul := &Consul{}
	consul.GatherHealthCheck(&acc, sampleChecks)

	acc.AssertContainsTaggedFields(t, "consul_health_checks", expectedFields, expectedTags)
}
