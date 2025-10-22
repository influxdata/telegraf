// Test Suite
package jenkins

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestJobRequestHierarchyName(t *testing.T) {
	tests := []struct {
		name     string
		input    jobRequest
		expected string
	}{
		{
			name:     "empty",
			input:    jobRequest{},
			expected: "",
		},
		{
			name: "parents",
			input: jobRequest{
				name:    "1",
				parents: []string{"3", "2"},
			},
			expected: "3/2/1",
		},
		{
			name: "parents special character",
			input: jobRequest{
				name:    "job 3",
				parents: []string{"job 1", "job 2"},
			},
			expected: "job 1/job 2/job 3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.input.hierarchyName())
		})
	}
}

func TestJobRequestURL(t *testing.T) {
	tests := []struct {
		name     string
		input    jobRequest
		expected string
	}{
		{
			name: "parents",
			input: jobRequest{
				name:    "1",
				parents: []string{"3", "2"},
			},
			expected: "/job/3/job/2/job/1/api/json",
		},
		{
			name: "parents special character",
			input: jobRequest{
				name:    "job 3",
				parents: []string{"job 1", "job 2"},
			},
			expected: "/job/job%201/job/job%202/job/job%203/api/json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.input.url())
		})
	}
}

func TestResultCode(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"SUCCESS", 0},
		{"Failure", 1},
		{"NOT_BUILT", 2},
		{"UNSTABLE", 3},
		{"ABORTED", 4},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			require.Equal(t, tt.expected, mapResultCode(tt.input))
		})
	}
}

type mockHandler struct {
	// responseMap is the path to response interface
	// we will output the serialized response in json when serving http
	// example '/computer/api/json': *gojenkins.
	responseMap map[string]interface{}
}

func (h mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	o, ok := h.responseMap[r.URL.RequestURI()]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	b, err := json.Marshal(o)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(b) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Write(b) //nolint:errcheck // ignore the returned error as the tests will fail anyway
}

func TestInitFail(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		expected string
	}{
		{
			name:     "bad jenkins config",
			address:  "http://a bad url",
			expected: `invalid character " " in host name`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup plugin
			plugin := &Jenkins{
				URL:             tt.address,
				Log:             testutil.Logger{},
				ResponseTimeout: config.Duration(time.Second),
			}

			err := plugin.initialize(&http.Client{Transport: &http.Transport{}})
			require.ErrorContains(t, err, tt.expected)
		})
	}
}

func TestInit(t *testing.T) {
	mh := mockHandler{
		responseMap: map[string]interface{}{
			"/api/json": struct{}{},
		},
	}
	ts := httptest.NewServer(mh)
	defer ts.Close()
	mockClient := &http.Client{Transport: &http.Transport{}}
	tests := []struct {
		// name of the test
		name        string
		jobInclude  []string
		jobExclude  []string
		nodeExclude []string
	}{
		{
			name: "default",
		},
		{
			name:        "with filters",
			jobInclude:  []string{"jobA", "jobB"},
			jobExclude:  []string{"job1", "job2"},
			nodeExclude: []string{"node1", "node2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the plugin
			plugin := &Jenkins{
				URL:             ts.URL,
				ResponseTimeout: config.Duration(time.Second),
				JobInclude:      tt.jobInclude,
				JobExclude:      tt.jobExclude,
				NodeExclude:     tt.nodeExclude,
				Log:             testutil.Logger{},
			}
			require.NoError(t, plugin.initialize(mockClient))

			// Check the default values
			require.Equal(t, 5, plugin.MaxConnections)
			require.Equal(t, 10, plugin.MaxSubJobPerLayer)
		})
	}
}

func TestGatherFail(t *testing.T) {
	tests := []struct {
		name     string
		response map[string]interface{}
		expected string
	}{
		{
			name: "bad node data",
			response: map[string]interface{}{
				"/api/json": struct{}{},
				"/computer/api/json": nodeResponse{
					Computers: []node{
						{},
						{},
						{},
					},
				},
			},
			expected: "error empty node name",
		},
		{
			name: "bad inner jobs",
			response: map[string]interface{}{
				"/computer/api/json": nil,
				"/api/json": &jobResponse{
					Jobs: []innerJob{
						{Name: "job1"},
					},
				},
			},
			expected: "[/job/job1/api/json] 404 Not Found",
		},
		{
			name: "bad build info",
			response: map[string]interface{}{
				"/computer/api/json": nil,
				"/api/json": &jobResponse{
					Jobs: []innerJob{
						{Name: "job1"},
					},
				},
				"/job/job1/api/json": &jobResponse{
					LastBuild: jobBuild{
						Number: 1,
					},
				},
			},
			expected: "[/job/job1/1/api/json] 404 Not Found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Lookup the response using the URI
				response, ok := tt.response[r.URL.RequestURI()]
				if !ok {
					w.WriteHeader(http.StatusNotFound)
					return
				}

				// Encode the response to JSON
				buf, err := json.Marshal(response)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				if len(buf) == 0 {
					w.WriteHeader(http.StatusNoContent)
					return
				}

				// Send the response
				if _, err := w.Write(buf); err != nil {
					t.Logf("writing failed: %v", err)
					t.Fail()
				}
			}))
			defer ts.Close()

			// Setup the plugin
			plugin := &Jenkins{
				Log:             testutil.Logger{},
				URL:             ts.URL,
				ResponseTimeout: config.Duration(time.Second),
			}

			// Collect the data and check for the expected error
			var acc testutil.Accumulator
			require.ErrorContains(t, acc.GatherError(plugin.Gather), tt.expected)

			expected := []telegraf.Metric{
				metric.New(
					"jenkins",
					map[string]string{
						"source": "127.0.0.1",
						"port":   "",
					},
					map[string]interface{}{
						"busy_executors":  0,
						"total_executors": 0,
					},
					time.Unix(0, 0),
				),
			}

			// Check the resulting metrics
			options := []cmp.Option{
				testutil.IgnoreTime(),
				testutil.IgnoreTags("port"),
			}

			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, options...)
		})
	}
}

func TestGatherNodeData(t *testing.T) {
	tests := []struct {
		name     string
		response map[string]interface{}
		expected []telegraf.Metric
	}{
		{
			name: "empty monitor data",
			response: map[string]interface{}{
				"/api/json": struct{}{},
				"/computer/api/json": nodeResponse{
					Computers: []node{
						{DisplayName: "master"},
						{DisplayName: "node1"},
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"jenkins",
					map[string]string{
						"source": "127.0.0.1",
						"port":   "",
					},
					map[string]interface{}{
						"busy_executors":  0,
						"total_executors": 0,
					},
					time.Unix(0, 0),
				),
				metric.New(
					"jenkins_node",
					map[string]string{
						"source":    "127.0.0.1",
						"port":      "",
						"node_name": "master",
						"status":    "online",
					},
					map[string]interface{}{
						"num_executors": int64(0),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "normal response",
			response: map[string]interface{}{
				"/api/json": struct{}{},
				"/computer/api/json": nodeResponse{
					BusyExecutors:  4,
					TotalExecutors: 8,
					Computers: []node{
						{
							DisplayName: "master",
							MonitorData: monitorData{
								HudsonNodeMonitorsArchitectureMonitor: "linux",
								HudsonNodeMonitorsResponseTimeMonitor: &responseTimeMonitor{
									Average: 10032,
								},
								HudsonNodeMonitorsDiskSpaceMonitor: &nodeSpaceMonitor{
									Path: "/path/1",
									Size: 123,
								},
								HudsonNodeMonitorsTemporarySpaceMonitor: &nodeSpaceMonitor{
									Path: "/path/2",
									Size: 245,
								},
								HudsonNodeMonitorsSwapSpaceMonitor: &swapSpaceMonitor{
									SwapAvailable:   212,
									SwapTotal:       500,
									MemoryAvailable: 101,
									MemoryTotal:     500,
								},
							},
							Offline: false,
						},
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"jenkins",
					map[string]string{
						"source": "127.0.0.1",
						"port":   "",
					},
					map[string]interface{}{
						"busy_executors":  4,
						"total_executors": 8,
					},
					time.Unix(0, 0),
				),
				metric.New(
					"jenkins_node",
					map[string]string{
						"source":    "127.0.0.1",
						"port":      "",
						"node_name": "master",
						"status":    "online",
						"arch":      "linux",
						"disk_path": "/path/1",
						"temp_path": "/path/2",
					},
					map[string]interface{}{
						"num_executors":    int64(0),
						"response_time":    int64(10032),
						"disk_available":   float64(123),
						"temp_available":   float64(245),
						"swap_available":   float64(212),
						"swap_total":       float64(500),
						"memory_available": float64(101),
						"memory_total":     float64(500),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "filtered nodes included",
			response: map[string]interface{}{
				"/api/json": struct{}{},
				"/computer/api/json": nodeResponse{
					BusyExecutors:  4,
					TotalExecutors: 8,
					Computers: []node{
						{DisplayName: "filtered-1"},
						{DisplayName: "filtered-1"},
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"jenkins",
					map[string]string{
						"source": "127.0.0.1",
						"port":   "",
					},
					map[string]interface{}{
						"busy_executors":  4,
						"total_executors": 8,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "filtered nodes excluded",
			response: map[string]interface{}{
				"/api/json": struct{}{},
				"/computer/api/json": nodeResponse{
					BusyExecutors:  4,
					TotalExecutors: 8,
					Computers: []node{
						{DisplayName: "ignore-1"},
						{DisplayName: "ignore-2"},
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"jenkins",
					map[string]string{
						"source": "127.0.0.1",
						"port":   "",
					},
					map[string]interface{}{
						"busy_executors":  4,
						"total_executors": 8,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "slave offline",
			response: map[string]interface{}{
				"/api/json": struct{}{},
				"/computer/api/json": nodeResponse{
					BusyExecutors:  4,
					TotalExecutors: 8,
					Computers: []node{
						{
							DisplayName:  "slave",
							MonitorData:  monitorData{},
							NumExecutors: 1,
							Offline:      true,
						},
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"jenkins",
					map[string]string{
						"source": "127.0.0.1",
						"port":   "",
					},
					map[string]interface{}{
						"busy_executors":  4,
						"total_executors": 8,
					},
					time.Unix(0, 0),
				),
				metric.New(
					"jenkins_node",
					map[string]string{
						"source":    "127.0.0.1",
						"port":      "",
						"node_name": "slave",
						"status":    "offline",
					},
					map[string]interface{}{
						"num_executors": 1,
					},
					time.Unix(0, 0),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Lookup the response using the URI
				response, ok := tt.response[r.URL.RequestURI()]
				if !ok {
					w.WriteHeader(http.StatusNotFound)
					return
				}

				// Encode the response to JSON
				buf, err := json.Marshal(response)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				if len(buf) == 0 {
					w.WriteHeader(http.StatusNoContent)
					return
				}

				// Send the response
				if _, err := w.Write(buf); err != nil {
					t.Logf("writing failed: %v", err)
					t.Fail()
				}
			}))
			defer ts.Close()

			// Setup the plugin
			plugin := &Jenkins{
				Log:             testutil.Logger{},
				URL:             ts.URL,
				ResponseTimeout: config.Duration(time.Second),
				NodeExclude:     []string{"ignore-1", "ignore-2"},
				NodeInclude:     []string{"master", "slave"},
			}

			// Collect the data
			var acc testutil.Accumulator
			require.NoError(t, acc.GatherError(plugin.Gather))

			// Check the resulting metrics
			options := []cmp.Option{
				testutil.IgnoreTime(),
				testutil.IgnoreTags("port"),
			}

			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, tt.expected, actual, options...)
		})
	}
}

func TestGatherLabels(t *testing.T) {
	response := map[string]interface{}{
		"/api/json": struct{}{},
		"/computer/api/json": nodeResponse{
			BusyExecutors:  4,
			TotalExecutors: 8,
			Computers: []node{
				{
					DisplayName: "master",
					AssignedLabels: []label{
						{"project_a"},
						{"testing"},
					},
					MonitorData: monitorData{
						HudsonNodeMonitorsResponseTimeMonitor: &responseTimeMonitor{
							Average: 54321,
						},
					},
				},
				{
					DisplayName: "secondary",
					MonitorData: monitorData{
						HudsonNodeMonitorsResponseTimeMonitor: &responseTimeMonitor{
							Average: 12345,
						},
					},
				},
			},
		},
	}

	expected := []telegraf.Metric{
		metric.New(
			"jenkins",
			map[string]string{
				"source": "127.0.0.1",
				"port":   "",
			},
			map[string]interface{}{
				"busy_executors":  4,
				"total_executors": 8,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"jenkins_node",
			map[string]string{
				"source":    "127.0.0.1",
				"port":      "",
				"node_name": "master",
				"status":    "online",
				"labels":    "project_a,testing",
			},
			map[string]interface{}{
				"num_executors": int64(0),
				"response_time": int64(54321),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"jenkins_node",
			map[string]string{
				"source":    "127.0.0.1",
				"port":      "",
				"node_name": "secondary",
				"status":    "online",
				"labels":    "none",
			},
			map[string]interface{}{
				"num_executors": int64(0),
				"response_time": int64(12345),
			},
			time.Unix(0, 0),
		),
	}

	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Lookup the response using the URI
		response, ok := response[r.URL.RequestURI()]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Encode the response to JSON
		buf, err := json.Marshal(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(buf) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Send the response
		if _, err := w.Write(buf); err != nil {
			t.Logf("writing failed: %v", err)
			t.Fail()
		}
	}))
	defer ts.Close()

	// Setup the plugin
	plugin := &Jenkins{
		Log:             testutil.Logger{},
		URL:             ts.URL,
		ResponseTimeout: config.Duration(time.Second),
		NodeLabelsAsTag: true,
	}

	// Collect the data
	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(plugin.Gather))

	// Check the resulting metrics
	options := []cmp.Option{
		testutil.IgnoreTime(),
		testutil.IgnoreTags("port"),
	}

	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, options...)
}

func TestGatherJobs(t *testing.T) {
	tests := []struct {
		name     string
		response map[string]interface{}
		expected []telegraf.Metric
	}{
		{
			name: "empty job",
			response: map[string]interface{}{
				"/api/json": &jobResponse{},
			},
			expected: []telegraf.Metric{
				metric.New(
					"jenkins",
					map[string]string{
						"source": "127.0.0.1",
						"port":   "",
					},
					map[string]interface{}{
						"busy_executors":  0,
						"total_executors": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "without build",
			response: map[string]interface{}{
				"/api/json": &jobResponse{
					Jobs: []innerJob{
						{Name: "job1"},
					},
				},
				"/job/job1/api/json": &jobResponse{},
			},
			expected: []telegraf.Metric{
				metric.New(
					"jenkins",
					map[string]string{
						"source": "127.0.0.1",
						"port":   "",
					},
					map[string]interface{}{
						"busy_executors":  0,
						"total_executors": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "ignore building job",
			response: map[string]interface{}{
				"/api/json": &jobResponse{
					Jobs: []innerJob{
						{Name: "job1"},
					},
				},
				"/job/job1/api/json": &jobResponse{
					LastBuild: jobBuild{
						Number: 1,
					},
				},
				"/job/job1/1/api/json": &buildResponse{
					Building: true,
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"jenkins",
					map[string]string{
						"source": "127.0.0.1",
						"port":   "",
					},
					map[string]interface{}{
						"busy_executors":  0,
						"total_executors": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "ignore old build",
			response: map[string]interface{}{
				"/api/json": &jobResponse{
					Jobs: []innerJob{
						{Name: "job1"},
					},
				},
				"/job/job1/api/json": &jobResponse{
					LastBuild: jobBuild{
						Number: 2,
					},
				},
				"/job/job1/2/api/json": &buildResponse{
					Building:  false,
					Timestamp: 100,
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"jenkins",
					map[string]string{
						"source": "127.0.0.1",
						"port":   "",
					},
					map[string]interface{}{
						"busy_executors":  0,
						"total_executors": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "normal",
			response: map[string]interface{}{
				"/api/json": &jobResponse{
					Jobs: []innerJob{
						{Name: "job1"},
						{Name: "job2"},
					},
				},
				"/job/job1/api/json": &jobResponse{
					LastBuild: jobBuild{
						Number: 3,
					},
				},
				"/job/job2/api/json": &jobResponse{
					LastBuild: jobBuild{
						Number: 1,
					},
				},
				"/job/job1/3/api/json": &buildResponse{
					Building:  false,
					Result:    "SUCCESS",
					Duration:  25558,
					Number:    3,
					Timestamp: (time.Now().Unix() - int64(time.Minute.Seconds())) * 1000,
				},
				"/job/job2/1/api/json": &buildResponse{
					Building:  false,
					Result:    "FAILURE",
					Duration:  1558,
					Number:    1,
					Timestamp: (time.Now().Unix() - int64(time.Minute.Seconds())) * 1000,
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"jenkins",
					map[string]string{
						"source": "127.0.0.1",
						"port":   "",
					},
					map[string]interface{}{
						"busy_executors":  0,
						"total_executors": 0,
					},
					time.Unix(0, 0),
				),
				metric.New(
					"jenkins_job",
					map[string]string{
						"source":  "127.0.0.1",
						"port":    "",
						"name":    "job1",
						"result":  "SUCCESS",
						"parents": "",
					},
					map[string]interface{}{
						"duration":    int64(25558),
						"number":      int64(3),
						"result_code": 0,
					},
					time.Unix(0, 0),
				),
				metric.New(
					"jenkins_job",
					map[string]string{
						"source":  "127.0.0.1",
						"port":    "",
						"name":    "job2",
						"result":  "FAILURE",
						"parents": "",
					},
					map[string]interface{}{
						"duration":    int64(1558),
						"number":      int64(1),
						"result_code": 1,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "with space",
			response: map[string]interface{}{
				"/api/json": &jobResponse{
					Jobs: []innerJob{
						{Name: "job 1"},
					},
				},
				"/job/job%201/api/json": &jobResponse{
					LastBuild: jobBuild{
						Number: 3,
					},
				},
				"/job/job%201/3/api/json": &buildResponse{
					Building:  false,
					Result:    "SUCCESS",
					Duration:  25558,
					Number:    3,
					Timestamp: (time.Now().Unix() - int64(time.Minute.Seconds())) * 1000,
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"jenkins",
					map[string]string{
						"source": "127.0.0.1",
						"port":   "",
					},
					map[string]interface{}{
						"busy_executors":  0,
						"total_executors": 0,
					},
					time.Unix(0, 0),
				),
				metric.New(
					"jenkins_job",
					map[string]string{
						"source":  "127.0.0.1",
						"port":    "",
						"name":    "job 1",
						"result":  "SUCCESS",
						"parents": "",
					},
					map[string]interface{}{
						"duration":    int64(25558),
						"number":      int64(3),
						"result_code": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "gather metrics for nested jobs with space exercising append slice behaviour",
			response: map[string]interface{}{
				"/api/json": &jobResponse{
					Jobs: []innerJob{
						{Name: "l1"},
					},
				},
				"/job/l1/api/json": &jobResponse{
					Jobs: []innerJob{
						{Name: "l2"},
					},
				},
				"/job/l1/job/l2/api/json": &jobResponse{
					Jobs: []innerJob{
						{Name: "job 1"},
					},
				},
				"/job/l1/job/l2/job/job%201/api/json": &jobResponse{
					Jobs: []innerJob{
						{Name: "job 2"},
					},
				},
				"/job/l1/job/l2/job/job%201/job/job%202/api/json": &jobResponse{
					LastBuild: jobBuild{
						Number: 3,
					},
				},
				"/job/l1/job/l2/job/job%201/job/job%202/3/api/json": &buildResponse{
					Building:  false,
					Result:    "SUCCESS",
					Duration:  25558,
					Timestamp: (time.Now().Unix() - int64(time.Minute.Seconds())) * 1000,
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"jenkins",
					map[string]string{
						"source": "127.0.0.1",
						"port":   "",
					},
					map[string]interface{}{
						"busy_executors":  0,
						"total_executors": 0,
					},
					time.Unix(0, 0),
				),
				metric.New(
					"jenkins_job",
					map[string]string{
						"source":  "127.0.0.1",
						"port":    "",
						"name":    "job 2",
						"result":  "SUCCESS",
						"parents": "l1/l2/job 1",
					},
					map[string]interface{}{
						"duration":    int64(25558),
						"result_code": 0,
						"number":      0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "gather sub jobs, jobs filter",
			response: map[string]interface{}{
				"/api/json": &jobResponse{
					Jobs: []innerJob{
						{Name: "apps"},
						{Name: "ignore-1"},
					},
				},
				"/job/ignore-1/api/json": &jobResponse{},
				"/job/apps/api/json": &jobResponse{
					Jobs: []innerJob{
						{Name: "k8s-cloud"},
						{Name: "chronograf"},
						{Name: "ignore-all"},
					},
				},
				"/job/apps/job/ignore-all/api/json": &jobResponse{
					Jobs: []innerJob{
						{Name: "1"},
						{Name: "2"},
					},
				},
				"/job/apps/job/ignore-all/job/1/api/json": &jobResponse{
					LastBuild: jobBuild{
						Number: 1,
					},
				},
				"/job/apps/job/ignore-all/job/2/api/json": &jobResponse{
					LastBuild: jobBuild{
						Number: 1,
					},
				},
				"/job/apps/job/chronograf/api/json": &jobResponse{
					LastBuild: jobBuild{
						Number: 1,
					},
				},
				"/job/apps/job/k8s-cloud/api/json": &jobResponse{
					Jobs: []innerJob{
						{Name: "PR-100"},
						{Name: "PR-101"},
						{Name: "PR-ignore2"},
						{Name: "PR 1"},
						{Name: "PR ignore"},
					},
				},
				"/job/apps/job/k8s-cloud/job/PR%20ignore/api/json": &jobResponse{
					LastBuild: jobBuild{
						Number: 1,
					},
				},
				"/job/apps/job/k8s-cloud/job/PR-ignore2/api/json": &jobResponse{
					LastBuild: jobBuild{
						Number: 1,
					},
				},
				"/job/apps/job/k8s-cloud/job/PR-100/api/json": &jobResponse{
					LastBuild: jobBuild{
						Number: 1,
					},
				},
				"/job/apps/job/k8s-cloud/job/PR-101/api/json": &jobResponse{
					LastBuild: jobBuild{
						Number: 4,
					},
				},
				"/job/apps/job/k8s-cloud/job/PR%201/api/json": &jobResponse{
					LastBuild: jobBuild{
						Number: 1,
					},
				},
				"/job/apps/job/chronograf/1/api/json": &buildResponse{
					Building:  false,
					Result:    "FAILURE",
					Duration:  1558,
					Number:    1,
					Timestamp: (time.Now().Unix() - int64(time.Minute.Seconds())) * 1000,
				},
				"/job/apps/job/k8s-cloud/job/PR-101/4/api/json": &buildResponse{
					Building:  false,
					Result:    "SUCCESS",
					Duration:  76558,
					Number:    4,
					Timestamp: (time.Now().Unix() - int64(time.Minute.Seconds())) * 1000,
				},
				"/job/apps/job/k8s-cloud/job/PR-100/1/api/json": &buildResponse{
					Building:  false,
					Result:    "SUCCESS",
					Duration:  91558,
					Number:    1,
					Timestamp: (time.Now().Unix() - int64(time.Minute.Seconds())) * 1000,
				},
				"/job/apps/job/k8s-cloud/job/PR%201/1/api/json": &buildResponse{
					Building:  false,
					Result:    "SUCCESS",
					Duration:  87832,
					Number:    1,
					Timestamp: (time.Now().Unix() - int64(time.Minute.Seconds())) * 1000,
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"jenkins",
					map[string]string{
						"source": "127.0.0.1",
						"port":   "",
					},
					map[string]interface{}{
						"busy_executors":  0,
						"total_executors": 0,
					},
					time.Unix(0, 0),
				),
				metric.New(
					"jenkins_job",
					map[string]string{
						"source":  "127.0.0.1",
						"port":    "",
						"name":    "PR 1",
						"result":  "SUCCESS",
						"parents": "apps/k8s-cloud",
					},
					map[string]interface{}{
						"duration":    int64(87832),
						"number":      int64(1),
						"result_code": 0,
					},
					time.Unix(0, 0),
				),
				metric.New(
					"jenkins_job",
					map[string]string{
						"source":  "127.0.0.1",
						"port":    "",
						"name":    "PR-100",
						"result":  "SUCCESS",
						"parents": "apps/k8s-cloud",
					},
					map[string]interface{}{
						"duration":    int64(91558),
						"number":      int64(1),
						"result_code": 0,
					},
					time.Unix(0, 0),
				),
				metric.New(
					"jenkins_job",
					map[string]string{
						"source":  "127.0.0.1",
						"port":    "",
						"name":    "PR-101",
						"result":  "SUCCESS",
						"parents": "apps/k8s-cloud",
					},
					map[string]interface{}{
						"duration":    int64(76558),
						"number":      int64(4),
						"result_code": 0,
					},
					time.Unix(0, 0),
				),
				metric.New(
					"jenkins_job",
					map[string]string{
						"source":  "127.0.0.1",
						"port":    "",
						"name":    "chronograf",
						"result":  "FAILURE",
						"parents": "apps",
					},
					map[string]interface{}{
						"duration":    int64(1558),
						"number":      int64(1),
						"result_code": 1,
					},
					time.Unix(0, 0),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Lookup the response using the URI
				response, ok := tt.response[r.URL.RequestURI()]
				if !ok {
					// Shortcut unrelated endpoints
					if r.URL.RequestURI() != "/computer/api/json" {
						w.WriteHeader(http.StatusNotFound)
						return
					}
				}

				// Encode the response to JSON
				buf, err := json.Marshal(response)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				if len(buf) == 0 {
					w.WriteHeader(http.StatusNoContent)
					return
				}

				// Send the response
				if _, err := w.Write(buf); err != nil {
					t.Logf("writing failed: %v", err)
					t.Fail()
				}
			}))
			defer ts.Close()

			// Setup the plugin
			plugin := &Jenkins{
				URL:             ts.URL,
				MaxBuildAge:     config.Duration(time.Hour),
				ResponseTimeout: config.Duration(time.Second),
				JobInclude:      []string{"*"},
				JobExclude: []string{
					"ignore-1",
					"apps/ignore-all/*",
					"apps/k8s-cloud/PR-ignore2",
					"apps/k8s-cloud/PR ignore",
				},
				Log: testutil.Logger{},
			}

			// Collect the data
			var acc testutil.Accumulator
			require.NoError(t, acc.GatherError(plugin.Gather))

			// Check the resulting metrics
			options := []cmp.Option{
				testutil.IgnoreTime(),
				testutil.SortMetrics(),
				testutil.IgnoreTags("port"),
			}

			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, tt.expected, actual, options...)
		})
	}
}
