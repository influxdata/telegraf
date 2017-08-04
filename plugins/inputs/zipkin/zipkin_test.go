package zipkin

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/telegraf/testutil"
)

func TestZipkinPlugin(t *testing.T) {
	mockAcc := testutil.Accumulator{}

	tests := []struct {
		name           string
		thriftDataFile string //path name to a binary thrift data file which contains test data
		wantErr        bool
		want           []testutil.Metric
	}{
		{
			name:           "threespan",
			thriftDataFile: "testdata/threespans.dat",
			want: []testutil.Metric{
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":           "8090652509916334619",
						"parent_id":    "22964302721410078",
						"trace_id":     "22c4fc8ab3669045",
						"service_name": "trivial",
						"name":         "Child",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(53106) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851331000).UTC(),
				},
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":             "8090652509916334619",
						"parent_id":      "22964302721410078",
						"trace_id":       "22c4fc8ab3669045",
						"name":           "Child",
						"service_name":   "trivial",
						"annotation":     "trivial", //base64: dHJpdmlhbA==
						"endpoint_host":  "127.0.0.1",
						"annotation_key": "lc",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(53106) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851331000).UTC(),
				},
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":           "103618986556047333",
						"parent_id":    "22964302721410078",
						"trace_id":     "22c4fc8ab3669045",
						"service_name": "trivial",
						"name":         "Child",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(50410) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360904552000).UTC(),
				},
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":             "103618986556047333",
						"parent_id":      "22964302721410078",
						"trace_id":       "22c4fc8ab3669045",
						"name":           "Child",
						"service_name":   "trivial",
						"annotation":     "trivial", //base64: dHJpdmlhbA==
						"endpoint_host":  "127.0.0.1",
						"annotation_key": "lc",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(50410) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360904552000).UTC(),
				},
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":           "22964302721410078",
						"parent_id":    "22964302721410078",
						"trace_id":     "22c4fc8ab3669045",
						"service_name": "trivial",
						"name":         "Parent",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(103680) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851318000).UTC(),
				},
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"service_name":  "trivial",
						"annotation":    "Starting child #0",
						"endpoint_host": "127.0.0.1",
						"id":            "22964302721410078",
						"parent_id":     "22964302721410078",
						"trace_id":      "22c4fc8ab3669045",
						"name":          "Parent",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(103680) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851318000).UTC(),
				},
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"service_name":  "trivial",
						"annotation":    "Starting child #1",
						"endpoint_host": "127.0.0.1",
						"id":            "22964302721410078",
						"parent_id":     "22964302721410078",
						"trace_id":      "22c4fc8ab3669045",
						"name":          "Parent",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(103680) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851318000).UTC(),
				},
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"parent_id":     "22964302721410078",
						"trace_id":      "22c4fc8ab3669045",
						"name":          "Parent",
						"service_name":  "trivial",
						"annotation":    "A Log",
						"endpoint_host": "127.0.0.1",
						"id":            "22964302721410078",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(103680) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851318000).UTC(),
				},
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"trace_id":       "22c4fc8ab3669045",
						"service_name":   "trivial",
						"annotation":     "trivial", //base64: dHJpdmlhbA==
						"annotation_key": "lc",
						"id":             "22964302721410078",
						"parent_id":      "22964302721410078",
						"name":           "Parent",
						"endpoint_host":  "127.0.0.1",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(103680) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851318000).UTC(),
				},
			},
			wantErr: false,
		},
		{
			name:           "distributed_trace_sample",
			thriftDataFile: "testdata/distributed_trace_sample.dat",
			want: []testutil.Metric{
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":           "6802735349851856000",
						"parent_id":    "6802735349851856000",
						"trace_id":     "5e682bc21ce99c80",
						"service_name": "go-zipkin-testclient",
						"name":         "main.dud",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(1) * time.Microsecond).Nanoseconds(),
					},
					//Time: time.Unix(1, 0).UTC(),
					Time: time.Unix(0, 1433330263415871*int64(time.Microsecond)).UTC(),
				},
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"annotation":    "cs",
						"endpoint_host": "0.0.0.0:9410",
						"id":            "6802735349851856000",
						"parent_id":     "6802735349851856000",
						"trace_id":      "5e682bc21ce99c80",
						"name":          "main.dud",
						"service_name":  "go-zipkin-testclient",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(1) * time.Microsecond).Nanoseconds(),
					},
					//Time: time.Unix(1, 0).UTC(),
					Time: time.Unix(0, 1433330263415871*int64(time.Microsecond)).UTC(),
				},
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"annotation":    "cr",
						"endpoint_host": "0.0.0.0:9410",
						"id":            "6802735349851856000",
						"parent_id":     "6802735349851856000",
						"trace_id":      "5e682bc21ce99c80",
						"name":          "main.dud",
						"service_name":  "go-zipkin-testclient",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(1) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1433330263415871*int64(time.Microsecond)).UTC(),
				},
			},
		},
	}

	z := &Zipkin{
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
			if err := postThriftData(tt.thriftDataFile, z.address); err != nil {
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

func postThriftData(datafile, address string) error {
	dat, err := ioutil.ReadFile(datafile)
	if err != nil {
		return fmt.Errorf("could not read from data file %s", datafile)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/api/v1/spans", address), bytes.NewReader(dat))

	if err != nil {
		return fmt.Errorf("HTTP request creation failed")
	}

	req.Header.Set("Content-Type", "application/x-thrift")
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP POST request to zipkin endpoint %s failed %v", address, err)
	}

	return nil
}
