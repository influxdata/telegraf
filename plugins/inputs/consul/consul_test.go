package consul

import (
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/influxdata/telegraf/testutil"
)

var sampleChecks = []*api.HealthCheck{
	{
		Node:        "localhost",
		CheckID:     "foo.health123",
		Name:        "foo.health",
		Status:      "passing",
		Notes:       "lorem ipsum",
		Output:      "OK",
		ServiceID:   "foo.123",
		ServiceName: "foo",
		ServiceTags: []string{"bar", "env:sandbox", "tagkey:value:stillvalue"},
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
		"node":                    "localhost",
		"service_name":            "foo",
		"check_id":                "foo.health123",
		"bar":                     "bar",
		"env:sandbox":             "env:sandbox",
		"tagkey:value:stillvalue": "tagkey:value:stillvalue",
	}

	var acc testutil.Accumulator

	consul := &Consul{}
	consul.GatherHealthCheck(&acc, sampleChecks)

	acc.AssertContainsTaggedFields(t, "consul_health_checks", expectedFields, expectedTags)
}

func TestGatherHealthCheckWithDelimitedTags(t *testing.T) {
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
		"bar":          "bar",
		"env":          "sandbox",
		"tagkey":       "value:stillvalue",
	}

	var acc testutil.Accumulator

	consul := &Consul{
		TagDelimiter: ":",
	}
	consul.GatherHealthCheck(&acc, sampleChecks)

	acc.AssertContainsTaggedFields(t, "consul_health_checks", expectedFields, expectedTags)
}

func TestGatherHealthCheckV2(t *testing.T) {
	expectedFields := map[string]interface{}{
		"passing":  1,
		"critical": 0,
		"warning":  0,
	}

	expectedTags := map[string]string{
		"node":                    "localhost",
		"check_id":                "foo.health123",
		"check_name":              "foo.health",
		"status":                  "passing",
		"service_id":              "foo.123",
		"service_name":            "foo",
		"bar":                     "bar",
		"env:sandbox":             "env:sandbox",
		"tagkey:value:stillvalue": "tagkey:value:stillvalue",
	}

	var acc testutil.Accumulator

	consul := &Consul{
		MetricVersion: 2,
	}
	consul.GatherHealthCheck(&acc, sampleChecks)

	acc.AssertContainsTaggedFields(t, "consul_health_checks", expectedFields, expectedTags)
}

func TestGatherHealthCheckWithDelimitedTagsV2(t *testing.T) {
	expectedFields := map[string]interface{}{
		"passing":  1,
		"critical": 0,
		"warning":  0,
	}

	expectedTags := map[string]string{
		"node":         "localhost",
		"check_id":     "foo.health123",
		"check_name":   "foo.health",
		"status":       "passing",
		"service_id":   "foo.123",
		"service_name": "foo",
		"bar":          "bar",
		"env":          "sandbox",
		"tagkey":       "value:stillvalue",
	}

	var acc testutil.Accumulator

	consul := &Consul{
		MetricVersion: 2,
		TagDelimiter:  ":",
	}
	consul.GatherHealthCheck(&acc, sampleChecks)

	acc.AssertContainsTaggedFields(t, "consul_health_checks", expectedFields, expectedTags)
}
