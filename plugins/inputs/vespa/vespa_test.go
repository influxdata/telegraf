package vespa

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/require"
)

func TestVespaMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, response)
	}))
	defer ts.Close()

	v := Vespa{
		Url: ts.URL,
	}

	var acc testutil.Accumulator
	err := acc.GatherError(v.Gather)
	require.NoError(t, err)

	tests := []struct {
		measurement string
		fields      map[string]interface{}
		tags        map[string]string
	}{
		{
			"vespa",
			map[string]interface{}{
				"memory_virt": float64(2386239488),
				"memory_rss":  float64(280096768),
				"cpu":         float64(2.1065675340768),
			},
			map[string]string{
				"metrictype":   "system",
				"instance":     "container-clustercontroller",
				"clustername":  "cluster-controllers",
				"vespaVersion": "7.136.13",
			},
		},
		{
			"vespa",
			map[string]interface{}{
				"jdisc.gc.ms.average": float64(0),
			},
			map[string]string{
				"metrictype":   "standard",
				"instance":     "container-clustercontroller",
				"gcName":       "G1OldGeneration",
				"clustername":  "cluster-controllers",
				"vespaVersion": "7.136.13",
			},
		},
		{
			"vespa",
			map[string]interface{}{
				"cpu.util":  float64(38.7),
				"disk.util": float64(35),
				"mem.util":  float64(53),
			},
			map[string]string{
				"applicationId": "tenant.app.instance",
				"host":          "some-host",
				"zone":          "some-zone",
				"clusterId":     "container/default",
			},
		},
		{
			"vespa",
			map[string]interface{}{
				"net.in.bytes":  float64(123),
				"net.out.bytes": float64(456),
			},
			map[string]string{
				"applicationId": "tenant.app.instance",
				"host":          "some-host",
				"zone":          "some-zone",
				"clusterId":     "container/default",
			},
		},
	}

	for _, test := range tests {
		acc.AssertContainsTaggedFields(t, test.measurement, test.fields, test.tags)
	}
}

var response = `
{
  "nodes": [
    {
      "hostname": "some-host",
      "role": "container/default/0/1",
      "node": {
        "timestamp": 1581340861,
        "metrics": [
          {
            "values": {
              "cpu.util": 38.7,
              "disk.util": 35,
              "mem.util": 53
            },
            "dimensions": {
              "applicationId": "tenant.app.instance",
              "host": "some-host",
              "zone": "some-zone",
              "clusterId": "container/default"
            }
          },
          {
            "values": {
              "net.in.bytes": 123,
              "net.out.bytes": 456
            },
            "dimensions": {
              "applicationId": "tenant.app.instance",
              "host": "some-host",
              "zone": "some-zone",
              "clusterId": "container/default"
            }
          }
        ]
      },
      "services": [
        {
          "name": "vespa.container-clustercontroller",
          "timestamp": 1580311023,
          "status": {
            "code": "up",
            "description": "Data collected successfully"
          },
          "metrics": [
            {
              "values": {
                "memory_virt": 2386239488,
                "memory_rss": 280096768,
                "cpu": 2.1065675340768
              },
              "dimensions": {
                "metrictype": "system",
                "instance": "container-clustercontroller",
                "clustername": "cluster-controllers",
                "vespaVersion": "7.136.13"
              }
            },
            {
              "values": {
                "jdisc.gc.ms.average": 0
              },
              "dimensions": {
                "metrictype": "standard",
                "instance": "container-clustercontroller",
                "gcName": "G1OldGeneration",
                "clustername": "cluster-controllers",
                "vespaVersion": "7.136.13"
              }
            }
          ]
        }
      ]
    }
  ]
}
`
