package aurora

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

var masterServer *httptest.Server

func getRawMetrics() string {
	return `assigner_launch_failures 0
cron_job_triggers 240
sla_cluster_mtta_ms 18
sla_disk_small_mttr_ms 1029
sla_cpu_small_mtta_ms 17
jvm_prop_java.endorsed.dirs /usr/lib/jvm/java-8-openjdk-amd64/jre/lib/endorsed
sla_role2/prod2/jobname2_job_uptime_50.00_sec 25`
}

func TestMain(m *testing.M) {
	metrics := getRawMetrics()

	masterRouter := http.NewServeMux()
	masterRouter.HandleFunc("/vars", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, metrics)
	})
	masterServer = httptest.NewServer(masterRouter)

	rc := m.Run()

	masterServer.Close()
	os.Exit(rc)
}

func TestConvertToNumeric(t *testing.T) {
	if _, isNumeric := convertToNumeric("0.000"); !isNumeric {
		t.Fatalf("0.000 should have been numeric")
	}
	if _, isNumeric := convertToNumeric("7"); !isNumeric {
		t.Fatalf("7 should have been numeric")
	}
	if boolVal, isNumeric := convertToNumeric("true"); !isNumeric {
		if val := boolVal.(int); val != 1 {
			t.Fatalf("true should have been converted to a 1")
		}
		t.Fatalf("true should have been numeric")
	}
	if boolVal, isNumeric := convertToNumeric("false"); !isNumeric {
		if val := boolVal.(int); val != 0 {
			t.Fatalf("false should have been converted to a 0")
		}
		t.Fatalf("false should have been numeric")
	}
	if _, isNumeric := convertToNumeric("&"); isNumeric {
		t.Fatalf("& should not be numeric")
	}
}

func TestIsJobMetric(t *testing.T) {
	var notJobMetrics = []string{
		"assigner_launch_failures",
		"cron_job_triggers",
		"sla_cluster_mtta_ms",
		"sla_disk_small_mttr_ms",
		"sla_cpu_small_mtta_ms",
	}
	for _, metric := range notJobMetrics {
		if isJobMetric(metric) {
			t.Fatalf("%s should not be a job metric", metric)
		}
	}
	var isJobMetrics = []string{
		"sla_role2/prod2/jobname2_job_uptime_50.00_sec",
	}
	for _, metric := range isJobMetrics {
		if !isJobMetric(metric) {
			t.Fatalf("%s should be a job metric", metric)
		}
	}
}

func TestParseJobSpecificMetric(t *testing.T) {
	var expectedFields = map[string]interface{}{
		"job_uptime_50.00_sec": 0,
	}
	var expectedTags = map[string]string{
		"role": "role2",
		"env":  "prod2",
		"job":  "jobname2",
	}
	key := "sla_role2/prod2/jobname2_job_uptime_50.00_sec"
	value := 0
	fields, tags := parseJobSpecificMetric(key, value)
	assert.Equal(t, fields, expectedFields)
	assert.Equal(t, tags, expectedTags)
}

func TestAuroraMaster(t *testing.T) {
	var acc testutil.Accumulator

	m := Aurora{
		Master:     masterServer.Listener.Addr().String(),
		Timeout:    10,
		HttpPrefix: "http",
		Numeric:    true,
	}

	err := m.Gather(&acc)
	if err != nil {
		t.Error(err)
	}

	var referenceMetrics = map[string]interface{}{
		"job_uptime_50.00_sec": 25.0,
	}
	acc.AssertContainsFields(t, "aurora", referenceMetrics)
}
