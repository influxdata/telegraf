package appdynamics_ma

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

/*************************************************************************
* Define AppD constants for test values
*************************************************************************/
const (
	testHost       = "http://127.0.0.1"
	testPort       = "8293"
	testMetricPath = "Custom Metrics|Telegraf|"
)

/*************************************************************************
* Function to create and return a test AppDynamicsMA using test vals
*************************************************************************/
func CreateTestAppDMa() AppDynamicsMA {
	a := AppDynamicsMA{
		Host:       testHost,
		Port:       testPort,
		MetricPath: testMetricPath}

	return a
}

/*************************************************************************
* Function to create and return a test Telegraf Metric Slice
*************************************************************************/
func CreateTestMetric() []telegraf.Metric {
	/*************************************************************************
	* Build a Local Test Metric
	*************************************************************************/
	// Declare Local var to hold the current time in Time format
	now := time.Now()
	// Declare a local map of strings to hold a few metric tags that will be
	// assigned to my test metric
	tags := map[string]string{
		"host":       "localhost",
		"datacenter": "us-east-1",
	}
	// Declare a local map of strings using types defined in the interface to
	// hold a few metric fields that will be assigned to my test metric
	fields := map[string]interface{}{
		"usage_idle": float64(90),
		"usage_busy": float64(10),
	}
	// Declare a new metric and pass the aforementioned properties since the
	// New function requires them to create a metric
	myMetric, err := metric.New("cpu", tags, fields, now)
	if err != nil {
		panic(err)
	}

	// Since my function requires a slice of metrics, here we go...
	myMetrics := make([]telegraf.Metric, 1)
	myMetrics[0] = myMetric

	return myMetrics
}

func TestBuildMetric(t *testing.T) {
	myTestMetric := CreateTestMetric()

	myTestSlice := BuildMetrics(myTestMetric, defaultMetricPath)

	myManualTestMetric1 := AppDynamicsJson{
		MetricName:     "Custom Metrics|Telegraf|cpu|us-east-1|localhost|usage_idle",
		AggregatorType: "AVERAGE",
		Value:          90}
	myManualTestMetric2 := AppDynamicsJson{
		MetricName:     "Custom Metrics|Telegraf|cpu|us-east-1|localhost|usage_busy",
		AggregatorType: "AVERAGE",
		Value:          10}

	var myManualTestSlice AppDynamicsSlice
	myManualTestSlice.Appdynamics_MA = append(myManualTestSlice.Appdynamics_MA, myManualTestMetric1)
	myManualTestSlice.Appdynamics_MA = append(myManualTestSlice.Appdynamics_MA, myManualTestMetric2)

	if !reflect.DeepEqual(myTestSlice, myManualTestSlice) {
		t.Errorf("Results not as expected.")
	}

}

func TestBadStatusCode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(`{ 'errors': [
    	'Something bad happened to the server.',
    	'Your query made the server very sad.'
  		]
		}`)
	}))
	defer ts.Close()

	a := CreateTestAppDMa()
	err := a.Connect()
	require.NoError(t, err)
	err = a.Write(testutil.MockMetrics())
	if err == nil {
		t.Errorf("error expected but none returned")
	} else {
		require.EqualError(t, fmt.Errorf("error POSTing metrics, Post http://127.0.0.1:8293/api/v1/metrics: dial tcp 127.0.0.1:8293: connect: connection refused\n"), err.Error())
	}
}
