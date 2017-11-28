package sumologic

import (
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSumoLogicWriteError(t *testing.T) {
	s := SumoLogic{
		Prefix:       "my.prefix",
		CollectorUrl: "http://localhost:8080",
	}
	// Init metrics
	m1, _ := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"mymeasurement": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	// Prepare point list
	var metrics []telegraf.Metric
	metrics = append(metrics, m1)
	// Error
	err1 := s.Connect()
	require.NoError(t, err1)
	err2 := s.Write(metrics)
	require.Error(t, err2)
	assert.Contains(t, err2.Error(), "error posting metrics to sumologic server")
}

func TestSumoLogicOK(t *testing.T) {
	s := SumoLogic{
		Prefix:       "my.prefix",
		CollectorUrl: "http://localhost:8080",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(`{"status":"ok"}`)
	}))
	defer ts.Close()

	s.CollectorUrl = ts.URL

	// Init metrics
	m1, _ := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"mymeasurement": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	// Prepare point list
	var metrics []telegraf.Metric
	metrics = append(metrics, m1)

	err := s.Connect()
	require.NoError(t, err)
	err = s.Write(metrics)
	require.NoError(t, err)
}

func TestBadStatusCode(t *testing.T) {
	s := SumoLogic{
		Prefix:       "my.prefix",
		CollectorUrl: "http://localhost:8080",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(`{ 'errors': 	'Something bad happened to the server.' }`)
	}))
	defer ts.Close()

	s.CollectorUrl = ts.URL
	// Init metrics
	m1, _ := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"mymeasurement": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	// Prepare point list
	var metrics []telegraf.Metric
	metrics = append(metrics, m1)

	err := s.Connect()
	require.NoError(t, err)
	err = s.Write(metrics)
	if err == nil {
		t.Errorf("error expected but none returned")
	} else {
		require.EqualError(t, fmt.Errorf("Received bad status code from server, 500\n"), err.Error())
	}
}
