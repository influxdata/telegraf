package marathon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

var inputMetrics map[string]interface{}
var outputMetrics map[string]interface{}
var testServer *httptest.Server

func generateMetrics() {
	inputMetrics = map[string]interface{}{
		"gauges": map[string]interface{}{
			"jvm.buffers.direct.capacity": map[string]interface{}{
				"count": 2,
			},
		},
		"timers": map[string]interface{}{
			"service.mesosphere.marathon.MarathonSchedulerService.run": map[string]interface{}{
				"m1_rate": 0.1,
			},
		},
	}

	outputMetrics = map[string]interface{}{
		"gauges_jvm.buffers.direct.capacity_count":                                2.0,
		"timers_service.mesosphere.marathon.MarathonSchedulerService.run_m1_rate": 0.1,
	}
}

func TestMain(m *testing.M) {
	generateMetrics()

	testRouter := http.NewServeMux()
	testRouter.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(inputMetrics)
	})
	testServer = httptest.NewServer(testRouter)

	rc := m.Run()

	testServer.Close()
	os.Exit(rc)
}

func TestMetricsGather(t *testing.T) {
	var acc testutil.Accumulator

	m := Marathon{
		Servers: []string{testServer.Listener.Addr().String()},
	}

	err := m.Gather(&acc)

	if err != nil {
		t.Errorf(err.Error())
	}

	acc.AssertContainsFields(t, "marathon", outputMetrics)
}

func TestMetricsFilter(t *testing.T) {
	m := Marathon{
		MetricTypes: []string{
			"gauges",
		},
	}
	b := []string{
		"counters", "timers", "meters", "histrograms",
	}

	m.filterMetrics(inputMetrics)

	for _, k := range b {
		if _, ok := inputMetrics[k]; ok {
			t.Errorf("Found key %s, it should be gone.", k)
		}
	}

	for _, k := range m.MetricTypes {
		if _, ok := inputMetrics[k]; !ok {
			t.Errorf("Didn't find key %s, it should be present.", k)
		}
	}
}
