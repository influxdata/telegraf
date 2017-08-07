package kairosdb

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHttpWriteNormal(t *testing.T) {
	actualPayload := []map[string]interface{}{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewDecoder(r.Body).Decode(&actualPayload)
		if !assert.NoError(t, err) {
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	now := time.Now()
	subject := &httpOutput{url: ts.URL, client: &http.Client{Timeout: time.Second}}
	m1, _ := metric.New("name123", nil, map[string]interface{}{"field1": 0.5, "field2": 5}, now)
	m2, _ := metric.New("name234", nil, map[string]interface{}{"field2": 6}, now)
	err := subject.Write([]telegraf.Metric{m1, m2})
	require.NoError(t, err)
	require.Len(t, actualPayload, 3)

	expectedTs := float64(now.UnixNano() / int64(time.Millisecond/time.Nanosecond))
	expected := []map[string]interface{}{
		{
			"name":      "name123.field1",
			"timestamp": expectedTs,
			"value":     0.5,
		},
		{
			"name":      "name123.field2",
			"timestamp": expectedTs,
			"value":     float64(5),
		},
		{
			"name":      "name234.field2",
			"timestamp": expectedTs,
			"value":     float64(6),
		},
	}
	for _, e := range expected {
		require.Contains(t, actualPayload, e)
	}
}

func TestHttpWriteHttpError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	subject := &httpOutput{url: ts.URL, client: &http.Client{Timeout: time.Second}}

	dummy, _ := metric.New("name123", nil, map[string]interface{}{"field1": 0.5}, time.Now())
	err := subject.Write([]telegraf.Metric{dummy})
	require.Error(t, err)
}

func TestPopulateDatapoint(t *testing.T) {
	now := time.Now()
	tags := map[string]string{"tag1": "val1", "tag2": "val2"}
	m1, _ := metric.New("name123", tags, map[string]interface{}{"ignored": 0}, now)
	expectedFloat := datapoint{
		Name:      "name123.field1",
		Timestamp: int64(now.UnixNano() / int64(time.Millisecond/time.Nanosecond)),
		Value:     float64(0.5),
		Tags:      tags,
	}
	actual, err := populateDatapoint(m1, "field1", 0.5)
	assert.NoError(t, err)
	assert.Equal(t, expectedFloat, actual)

	m2, _ := metric.New("name123", tags, map[string]interface{}{"ignored": 0}, now)
	expectedLong := datapoint{
		Name:      "name123.field1",
		Timestamp: int64(now.UnixNano() / int64(time.Millisecond/time.Nanosecond)),
		Value:     1234,
		Tags:      tags,
	}
	actual, err = populateDatapoint(m2, "field1", 1234)
	assert.NoError(t, err)
	assert.Equal(t, expectedLong, actual)

	m3, _ := metric.New("name123", tags, map[string]interface{}{"ignored": 0}, now)
	actual, err = populateDatapoint(m3, "field1", "unsupported")
	assert.Error(t, err)
}
