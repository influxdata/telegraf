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

func TestGatherHealtCheck(t *testing.T) {
	expectedFields := map[string]interface{}{
		"check_id":   "foo.health123",
		"check_name": "foo.health",
		"status":     "passing",
		"service_id": "foo.123",
	}

	expectedTags := map[string]string{
		"node":         "localhost",
		"service_name": "foo",
	}

	var acc testutil.Accumulator

	consul := &Consul{}
	consul.GatherHealthCheck(&acc, sampleChecks)

	acc.AssertContainsTaggedFields(t, "consul_health_checks", expectedFields, expectedTags)
}
