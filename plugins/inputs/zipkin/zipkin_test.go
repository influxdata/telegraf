package zipkin

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

func TestZipkinPlugin(t *testing.T) {
	mockAcc := testutil.Accumulator{}

	tests := []struct {
		name        string
		datafile    string // data file which contains test data
		contentType string
		wantErr     bool
		want        []testutil.Metric
	}{
		{
			name:        "threespan",
			datafile:    "testdata/threespans.dat",
			contentType: "application/x-thrift",
			want: []testutil.Metric{
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":           "7047c59776af8a1b",
						"parent_id":    "5195e96239641e",
						"trace_id":     "22c4fc8ab3669045",
						"service_name": "trivial",
						"name":         "child",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(53106) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851331000).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":             "7047c59776af8a1b",
						"parent_id":      "5195e96239641e",
						"trace_id":       "22c4fc8ab3669045",
						"name":           "child",
						"service_name":   "trivial",
						"annotation":     "trivial", //base64: dHJpdmlhbA==
						"endpoint_host":  "127.0.0.1",
						"annotation_key": "lc",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(53106) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851331000).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":           "17020eb55a8bfe5",
						"parent_id":    "5195e96239641e",
						"trace_id":     "22c4fc8ab3669045",
						"service_name": "trivial",
						"name":         "child",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(50410) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360904552000).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":             "17020eb55a8bfe5",
						"parent_id":      "5195e96239641e",
						"trace_id":       "22c4fc8ab3669045",
						"name":           "child",
						"service_name":   "trivial",
						"annotation":     "trivial", //base64: dHJpdmlhbA==
						"endpoint_host":  "127.0.0.1",
						"annotation_key": "lc",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(50410) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360904552000).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":           "5195e96239641e",
						"parent_id":    "5195e96239641e",
						"trace_id":     "22c4fc8ab3669045",
						"service_name": "trivial",
						"name":         "parent",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(103680) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851318000).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"service_name":  "trivial",
						"annotation":    "Starting child #0",
						"endpoint_host": "127.0.0.1",
						"id":            "5195e96239641e",
						"parent_id":     "5195e96239641e",
						"trace_id":      "22c4fc8ab3669045",
						"name":          "parent",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(103680) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851318000).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"service_name":  "trivial",
						"annotation":    "Starting child #1",
						"endpoint_host": "127.0.0.1",
						"id":            "5195e96239641e",
						"parent_id":     "5195e96239641e",
						"trace_id":      "22c4fc8ab3669045",
						"name":          "parent",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(103680) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851318000).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"parent_id":     "5195e96239641e",
						"trace_id":      "22c4fc8ab3669045",
						"name":          "parent",
						"service_name":  "trivial",
						"annotation":    "A Log",
						"endpoint_host": "127.0.0.1",
						"id":            "5195e96239641e",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(103680) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851318000).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"trace_id":       "22c4fc8ab3669045",
						"service_name":   "trivial",
						"annotation":     "trivial", //base64: dHJpdmlhbA==
						"annotation_key": "lc",
						"id":             "5195e96239641e",
						"parent_id":      "5195e96239641e",
						"name":           "parent",
						"endpoint_host":  "127.0.0.1",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(103680) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851318000).UTC(),
					Type: telegraf.Untyped,
				},
			},
			wantErr: false,
		},
		{
			name:        "distributed_trace_sample",
			datafile:    "testdata/distributed_trace_sample.dat",
			contentType: "application/x-thrift",
			want: []testutil.Metric{
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":           "5e682bc21ce99c80",
						"parent_id":    "5e682bc21ce99c80",
						"trace_id":     "5e682bc21ce99c80",
						"service_name": "go-zipkin-testclient",
						"name":         "main.dud",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(1) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1433330263415871*int64(time.Microsecond)).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"annotation":    "cs",
						"endpoint_host": "0.0.0.0:9410",
						"id":            "5e682bc21ce99c80",
						"parent_id":     "5e682bc21ce99c80",
						"trace_id":      "5e682bc21ce99c80",
						"name":          "main.dud",
						"service_name":  "go-zipkin-testclient",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(1) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1433330263415871*int64(time.Microsecond)).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"annotation":    "cr",
						"endpoint_host": "0.0.0.0:9410",
						"id":            "5e682bc21ce99c80",
						"parent_id":     "5e682bc21ce99c80",
						"trace_id":      "5e682bc21ce99c80",
						"name":          "main.dud",
						"service_name":  "go-zipkin-testclient",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(1) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1433330263415871*int64(time.Microsecond)).UTC(),
					Type: telegraf.Untyped,
				},
			},
		},
		{
			name:        "JSON rather than thrift",
			datafile:    "testdata/json/brave-tracer-example.json",
			contentType: "application/json",
			want: []testutil.Metric{
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":           "b26412d1ac16767d",
						"name":         "http:/hi2",
						"parent_id":    "7312f822d43d0fd8",
						"service_name": "test",
						"trace_id":     "7312f822d43d0fd8",
					},
					Fields: map[string]interface{}{
						"duration_ns": int64(3000000),
					},
					Time: time.Unix(0, 1503031538791000*int64(time.Microsecond)).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"annotation":    "sr",
						"endpoint_host": "192.168.0.8:8010",
						"id":            "b26412d1ac16767d",
						"name":          "http:/hi2",
						"parent_id":     "7312f822d43d0fd8",
						"service_name":  "test",
						"trace_id":      "7312f822d43d0fd8",
					},
					Fields: map[string]interface{}{
						"duration_ns": int64(3000000),
					},
					Time: time.Unix(0, 1503031538791000*int64(time.Microsecond)).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"annotation":    "ss",
						"endpoint_host": "192.168.0.8:8010",
						"id":            "b26412d1ac16767d",
						"name":          "http:/hi2",
						"parent_id":     "7312f822d43d0fd8",
						"service_name":  "test",
						"trace_id":      "7312f822d43d0fd8",
					},
					Fields: map[string]interface{}{
						"duration_ns": int64(3000000),
					},
					Time: time.Unix(0, 1503031538791000*int64(time.Microsecond)).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"annotation":     "Demo2Application",
						"annotation_key": "mvc.controller.class",
						"endpoint_host":  "192.168.0.8:8010",
						"id":             "b26412d1ac16767d",
						"name":           "http:/hi2",
						"parent_id":      "7312f822d43d0fd8",
						"service_name":   "test",
						"trace_id":       "7312f822d43d0fd8",
					},
					Fields: map[string]interface{}{
						"duration_ns": int64(3000000),
					},
					Time: time.Unix(0, 1503031538791000*int64(time.Microsecond)).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"annotation":     "hi2",
						"annotation_key": "mvc.controller.method",
						"endpoint_host":  "192.168.0.8:8010",
						"id":             "b26412d1ac16767d",
						"name":           "http:/hi2",
						"parent_id":      "7312f822d43d0fd8",
						"service_name":   "test",
						"trace_id":       "7312f822d43d0fd8",
					},
					Fields: map[string]interface{}{
						"duration_ns": int64(3000000),
					},
					Time: time.Unix(0, 1503031538791000*int64(time.Microsecond)).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"annotation":     "192.168.0.8:test:8010",
						"annotation_key": "spring.instance_id",
						"endpoint_host":  "192.168.0.8:8010",
						"id":             "b26412d1ac16767d",
						"name":           "http:/hi2",
						"parent_id":      "7312f822d43d0fd8",
						"service_name":   "test",
						"trace_id":       "7312f822d43d0fd8",
					},
					Fields: map[string]interface{}{
						"duration_ns": int64(3000000),
					},
					Time: time.Unix(0, 1503031538791000*int64(time.Microsecond)).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":           "b26412d1ac16767d",
						"name":         "http:/hi2",
						"parent_id":    "7312f822d43d0fd8",
						"service_name": "test",
						"trace_id":     "7312f822d43d0fd8",
					},
					Fields: map[string]interface{}{
						"duration_ns": int64(10000000),
					},
					Time: time.Unix(0, 1503031538786000*int64(time.Microsecond)).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"annotation":    "cs",
						"endpoint_host": "192.168.0.8:8010",
						"id":            "b26412d1ac16767d",
						"name":          "http:/hi2",
						"parent_id":     "7312f822d43d0fd8",
						"service_name":  "test",
						"trace_id":      "7312f822d43d0fd8",
					},
					Fields: map[string]interface{}{
						"duration_ns": int64(10000000),
					},
					Time: time.Unix(0, 1503031538786000*int64(time.Microsecond)).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"annotation":    "cr",
						"endpoint_host": "192.168.0.8:8010",
						"id":            "b26412d1ac16767d",
						"name":          "http:/hi2",
						"parent_id":     "7312f822d43d0fd8",
						"service_name":  "test",
						"trace_id":      "7312f822d43d0fd8",
					},
					Fields: map[string]interface{}{
						"duration_ns": int64(10000000),
					},
					Time: time.Unix(0, 1503031538786000*int64(time.Microsecond)).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"annotation":     "localhost",
						"annotation_key": "http.host",
						"endpoint_host":  "192.168.0.8:8010",
						"id":             "b26412d1ac16767d",
						"name":           "http:/hi2",
						"parent_id":      "7312f822d43d0fd8",
						"service_name":   "test",
						"trace_id":       "7312f822d43d0fd8",
					},
					Fields: map[string]interface{}{
						"duration_ns": int64(10000000),
					},
					Time: time.Unix(0, 1503031538786000*int64(time.Microsecond)).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"annotation":     "GET",
						"annotation_key": "http.method",
						"endpoint_host":  "192.168.0.8:8010",
						"id":             "b26412d1ac16767d",
						"name":           "http:/hi2",
						"parent_id":      "7312f822d43d0fd8",
						"service_name":   "test",
						"trace_id":       "7312f822d43d0fd8",
					},
					Fields: map[string]interface{}{
						"duration_ns": int64(10000000),
					},
					Time: time.Unix(0, 1503031538786000*int64(time.Microsecond)).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"annotation":     "/hi2",
						"annotation_key": "http.path",
						"endpoint_host":  "192.168.0.8:8010",
						"id":             "b26412d1ac16767d",
						"name":           "http:/hi2",
						"parent_id":      "7312f822d43d0fd8",
						"service_name":   "test",
						"trace_id":       "7312f822d43d0fd8",
					},
					Fields: map[string]interface{}{
						"duration_ns": int64(10000000),
					},
					Time: time.Unix(0, 1503031538786000*int64(time.Microsecond)).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"annotation":     "http://localhost:8010/hi2",
						"annotation_key": "http.url",
						"endpoint_host":  "192.168.0.8:8010",
						"id":             "b26412d1ac16767d",
						"name":           "http:/hi2",
						"parent_id":      "7312f822d43d0fd8",
						"service_name":   "test",
						"trace_id":       "7312f822d43d0fd8",
					},
					Fields: map[string]interface{}{
						"duration_ns": int64(10000000),
					},
					Time: time.Unix(0, 1503031538786000*int64(time.Microsecond)).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"annotation":     "192.168.0.8:test:8010",
						"annotation_key": "spring.instance_id",
						"endpoint_host":  "192.168.0.8:8010",
						"id":             "b26412d1ac16767d",
						"name":           "http:/hi2",
						"parent_id":      "7312f822d43d0fd8",
						"service_name":   "test",
						"trace_id":       "7312f822d43d0fd8",
					},
					Fields: map[string]interface{}{
						"duration_ns": int64(10000000),
					},
					Time: time.Unix(0, 1503031538786000*int64(time.Microsecond)).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":           "7312f822d43d0fd8",
						"name":         "http:/hi",
						"parent_id":    "7312f822d43d0fd8",
						"service_name": "test",
						"trace_id":     "7312f822d43d0fd8",
					},
					Fields: map[string]interface{}{
						"duration_ns": int64(23393000),
					},
					Time: time.Unix(0, 1503031538778000*int64(time.Microsecond)).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"annotation":    "sr",
						"endpoint_host": "192.168.0.8:8010",
						"id":            "7312f822d43d0fd8",
						"name":          "http:/hi",
						"parent_id":     "7312f822d43d0fd8",
						"service_name":  "test",
						"trace_id":      "7312f822d43d0fd8",
					},
					Fields: map[string]interface{}{
						"duration_ns": int64(23393000),
					},
					Time: time.Unix(0, 1503031538778000*int64(time.Microsecond)).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"annotation":    "ss",
						"endpoint_host": "192.168.0.8:8010",
						"id":            "7312f822d43d0fd8",
						"name":          "http:/hi",
						"parent_id":     "7312f822d43d0fd8",
						"service_name":  "test",
						"trace_id":      "7312f822d43d0fd8",
					},
					Fields: map[string]interface{}{
						"duration_ns": int64(23393000),
					},
					Time: time.Unix(0, 1503031538778000*int64(time.Microsecond)).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"annotation":     "Demo2Application",
						"annotation_key": "mvc.controller.class",
						"endpoint_host":  "192.168.0.8:8010",
						"id":             "7312f822d43d0fd8",
						"name":           "http:/hi",
						"parent_id":      "7312f822d43d0fd8",
						"service_name":   "test",
						"trace_id":       "7312f822d43d0fd8",
					},
					Fields: map[string]interface{}{
						"duration_ns": int64(23393000),
					},
					Time: time.Unix(0, 1503031538778000*int64(time.Microsecond)).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"annotation":     "hi",
						"annotation_key": "mvc.controller.method",
						"endpoint_host":  "192.168.0.8:8010",
						"id":             "7312f822d43d0fd8",
						"name":           "http:/hi",
						"parent_id":      "7312f822d43d0fd8",
						"service_name":   "test",
						"trace_id":       "7312f822d43d0fd8",
					},
					Fields: map[string]interface{}{
						"duration_ns": int64(23393000),
					},
					Time: time.Unix(0, 1503031538778000*int64(time.Microsecond)).UTC(),
					Type: telegraf.Untyped,
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"annotation":     "192.168.0.8:test:8010",
						"annotation_key": "spring.instance_id",
						"endpoint_host":  "192.168.0.8:8010",
						"id":             "7312f822d43d0fd8",
						"name":           "http:/hi",
						"parent_id":      "7312f822d43d0fd8",
						"service_name":   "test",
						"trace_id":       "7312f822d43d0fd8",
					},
					Fields: map[string]interface{}{
						"duration_ns": int64(23393000),
					},
					Time: time.Unix(0, 1503031538778000*int64(time.Microsecond)).UTC(),
					Type: telegraf.Untyped,
				},
			},
		},
	}

	// Workaround for Go 1.8
	// https://github.com/golang/go/issues/18806
	DefaultNetwork = "tcp4"

	z := &Zipkin{
		Log:  testutil.Logger{},
		Path: "/api/v1/spans",
		Port: 0,
	}

	err := z.Start(&mockAcc)
	if err != nil {
		t.Fatal("Failed to start zipkin server")
	}

	defer z.Stop()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAcc.ClearMetrics()
			if err := postThriftData(tt.datafile, z.address, tt.contentType); err != nil {
				t.Fatalf("Posting data to http endpoint /api/v1/spans failed. Error: %s\n", err)
			}
			mockAcc.Wait(len(tt.want)) //Since the server is running concurrently, we need to wait for the number of data points we want to test to be added to the Accumulator.
			if len(mockAcc.Errors) > 0 != tt.wantErr {
				t.Fatalf("Got unexpected errors. want error = %v, errors = %v\n", tt.wantErr, mockAcc.Errors)
			}

			var got []testutil.Metric
			for _, m := range mockAcc.Metrics {
				got = append(got, *m)
			}
			if !cmp.Equal(tt.want, got) {
				t.Fatalf("Got != Want\n %s", cmp.Diff(tt.want, got))
			}
		})
	}
	mockAcc.ClearMetrics()
	z.Stop()
	// Make sure there is no erroneous error on shutdown
	if len(mockAcc.Errors) != 0 {
		t.Fatal("Expected no errors on shutdown")
	}
}

func postThriftData(datafile, address, contentType string) error {
	dat, err := os.ReadFile(datafile)
	if err != nil {
		return fmt.Errorf("could not read from data file %s", datafile)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/api/v1/spans", address), bytes.NewReader(dat))
	if err != nil {
		return fmt.Errorf("HTTP request creation failed")
	}

	req.Header.Set("Content-Type", contentType)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP POST request to zipkin endpoint %s failed %v", address, err)
	}

	defer resp.Body.Close()

	return nil
}
