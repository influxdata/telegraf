package appdynamics

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/stretchr/testify/assert"
)

// TestAppdynamicsError - attempt to initialize Appdynamics with invalid controller user name value
func TestAppdynamicsError(t *testing.T) {
	a := Appdynamics{
		ControllerTierURL:     "https://foo.saas.appdynamics.com/controller/rest/applications/bar/tiers/baz?output=JSON",
		ControllerUserName:    "apiuser@foo.bar.com",
		ControllerPassword:    "pass123",
		AgentURL:              "http://localhost:8293/machineagent/metrics?name=Server|Component:%d|Custom+Metrics|",
	}

	assert.Error(t, a.Connect())
}

// TestAppdynamicsOK - successfully initialize Appdynamics and process metrics calls
func TestAppdynamicsOK(t *testing.T) {
	// channel to collect received calls
	ch := make(chan string, 1)

	h := func(w http.ResponseWriter, r *http.Request) {
		s := r.URL.String()
		fmt.Fprintf(w, "Hi there, I love %s!", s)
		ch <- r.URL.RawQuery
	}
	http.HandleFunc("/", h)
	go http.ListenAndServe(":8293", nil)
	time.Sleep(time.Millisecond * 100)

	a := Appdynamics{
		ControllerTierURL:     "https://foo.saas.appdynamics.com/controller/rest/applications/bar/tiers/baz?output=JSON",
		ControllerUserName:    "apiuser@foo.bar",
		ControllerPassword:    "pass123",
		AgentURL:              "http://localhost:8293/machineagent/metrics?name=Server|Component:%d|Custom+Metrics|",
	}

	// this error is expected since we are not connecting to actual controller
	assert.Error(t, a.Connect())
	// reset agent url value with '123' tier id
	a.AgentURL = fmt.Sprintf(a.AgentURL, 123)
	assert.Equal(t, a.AgentURL, "http://localhost:8293/machineagent/metrics?name=Server|Component:123|Custom+Metrics|")

	tm := time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)
	// counter type - appd-type: sum
	m, _ := telegraf.NewMetric(
		"foo",
		map[string]string{"metric_type": "counter"},
		map[string]interface{}{"value": float64(1.23)},
		tm,
	)
	metrics := []telegraf.Metric{m}
	assert.NoError(t, a.Write(metrics))
	call := <-ch
	assert.Equal(t, "name=Server|Component:123|Custom+Metrics|foo&value=1.23&type=sum", call)

	// gauge type - appd-type: average
	m, _ = telegraf.NewMetric(
		"foo",
		map[string]string{"metric_type": "gauge"},
		map[string]interface{}{"value": float64(4.56)},
		tm,
	)
	metrics = []telegraf.Metric{m}
	assert.NoError(t, a.Write(metrics))
	call = <-ch
	assert.Equal(t, "name=Server|Component:123|Custom+Metrics|foo&value=4.56&type=average", call)

	// other type - defaults to appd-type: sum
	m, _ = telegraf.NewMetric(
		"foo",
		map[string]string{"metric_type": "other"},
		map[string]interface{}{"value": float64(7.89)},
		tm,
	)
	metrics = []telegraf.Metric{m}
	assert.NoError(t, a.Write(metrics))
	call = <-ch
	assert.Equal(t, "name=Server|Component:123|Custom+Metrics|foo&value=7.89&type=sum", call)

	// invalid: missing value
	m, _ = telegraf.NewMetric(
		"foo",
		map[string]string{"metric_type": "bar"},
		map[string]interface{}{"values": float64(7.89)},
		tm,
	)
	metrics = []telegraf.Metric{m}
	assert.NoError(t, a.Write(metrics))
	select {
	case call = <-ch:
		t.Error("No messages expected, but got: ", call)
	default:
	}
}
