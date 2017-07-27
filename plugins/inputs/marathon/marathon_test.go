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
var outputMetrics map[string]map[string]interface{}
var testServer *httptest.Server

func generateMetrics() {
	inputMetrics = map[string]interface{}{
		"counters": map[string]interface{}{
			"org.eclipse.jetty.servlet.ServletContextHandler.active-dispatches": map[string]interface{}{
				"count": 1,
			},
		},
		"gauges": map[string]interface{}{
			"jvm.buffers.direct.capacity": map[string]interface{}{
				"value": 1,
			},
		},
		"histograms": map[string]interface{}{
			"service.mesosphere.marathon.state.MarathonStore.FrameworkId.read-data-size": map[string]interface{}{
				"count": 1,
				"max":   1,
				"mean":  1,
				"min":   1,
			},
		},
		"meters": map[string]interface{}{
			"service.mesosphere.marathon.state.GroupRepository.read-requests": map[string]interface{}{
				"count":     1,
				"mean_rate": 0.1,
			},
		},
		"timers": map[string]interface{}{
			"service.mesosphere.marathon.MarathonSchedulerService.run": map[string]interface{}{
				"m1_rate": 0.1,
				"m5_rate": 0.01,
			},
		},
	}

	outputMetrics = map[string]map[string]interface{}{
		"org_eclipse_jetty_servlet_ServletContextHandler_active-dispatches": {
			"count": 1.0,
		},
		"jvm_buffers_direct_capacity": {
			"value": 1.0,
		},
		"service_mesosphere_marathon_state_MarathonStore_FrameworkId_read-data-size": {
			"count": 1.0,
			"max":   1.0,
			"mean":  1.0,
			"min":   1.0,
		},
		"service_mesosphere_marathon_state_GroupRepository_read-requests": {
			"count":     1.0,
			"mean_rate": 0.1,
		},
		"service_mesosphere_marathon_MarathonSchedulerService_run": {
			"m1_rate": 0.1,
			"m5_rate": 0.01,
		},
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

	for measurement, fields := range outputMetrics {
		acc.AssertContainsFields(t, measurement, fields)
	}
}
