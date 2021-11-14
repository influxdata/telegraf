package nomad

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestNomadStats(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/v1/metrics" {
			w.WriteHeader(http.StatusOK)
			_, err := fmt.Fprintln(w, responseKeyMetrics)
			require.NoError(t, err)
		}
	}))
	defer ts.Close()

	labelFilter, _ := filter.NewIncludeExcludeFilter([]string{"host"}, nil)

	n := &Nomad{
		URL:         ts.URL,
		labelFilter: labelFilter,
	}

	var acc testutil.Accumulator
	err := acc.GatherError(n.Gather)
	require.NoError(t, err)

	fields := map[string]interface{}{
		"value": float32(500),
	}
	tags := map[string]string{
		"node_scheduling_eligibility": "eligible",
		"host":                        "node1",
		"node_id":                     "2bbff078-8473-a9de-6c5e-42b4e053e12f",
		"datacenter":                  "dc1",
		"node_class":                  "none",
		"node_status":                 "ready",
	}
	acc.AssertContainsTaggedFields(t, "nomad.client.allocated.cpu", fields, tags)

	fields = map[string]interface{}{
		"count": int(7),
		"max":   float64(1),
		"min":   float64(1),
		"mean":  float64(1),
		"rate":  float64(0.7),
		"sum":   float64(7),
		"sumsq": float64(0),
	}
	tags = map[string]string{
		"host": "node1",
	}
	acc.AssertContainsTaggedFields(t, "nomad.nomad.rpc.query", fields, tags)

	fields = map[string]interface{}{
		"count": int(20),
		"max":   float64(0.03747599944472313),
		"min":   float64(0.003459000028669834),
		"rate":  float64(0.026318199979141355),
		"sum":   float64(0.26318199979141355),
		"sumsq": float64(0),
		"mean":  float64(0.013159099989570678),
	}
	tags = map[string]string{
		"host": "node1",
	}
	acc.AssertContainsTaggedFields(t, "nomad.memberlist.gossip", fields, tags)

}

var responseKeyMetrics = `
{
	"Counters": [
		{
		  "Count": 7,
		  "Labels": {
			"host": "node1"
		  },
		  "Max": 1,
		  "Mean": 1,
		  "Min": 1,
		  "Name": "nomad.nomad.rpc.query",
		  "Rate": 0.7,
		  "Stddev": 0,
		  "Sum": 7
		}
	  ],
	"Gauges": [
		{
		  "Labels": {
			"node_scheduling_eligibility": "eligible",
			"host": "node1",
			"node_id": "2bbff078-8473-a9de-6c5e-42b4e053e12f",
			"datacenter": "dc1",
			"node_class": "none",
			"node_status": "ready"
		  },
		  "Name": "nomad.client.allocated.cpu",
		  "Value": 500
		}
	  ],
	"Points" : [],
	"Samples" : [
		{
		  "Count": 20,
		  "Labels": {
			"host": "node1"
		  },
		  "Max": 0.03747599944472313,
		  "Mean": 0.013159099989570678,
		  "Min": 0.003459000028669834,
		  "Name": "nomad.memberlist.gossip",
		  "Rate": 0.026318199979141355,
		  "Stddev": 0.009523742715522742,
		  "Sum": 0.26318199979141355
		}
	  ],
	"Timestamp": "2021-11-13 22:39:00 +0000 UTC"
  }
`
