// Test Suite
package jenkins

import (
	"encoding/json"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestResultCode(t *testing.T) {
	tests := []struct {
		input  string
		output int
	}{
		{"SUCCESS", 0},
		{"Failure", 1},
		{"NOT_BUILT", 2},
		{"UNSTABLE", 3},
		{"ABORTED", 4},
	}
	for _, test := range tests {
		output := mapResultCode(test.input)
		if output != test.output {
			t.Errorf("Expected %d, got %d\n", test.output, output)
		}
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

func TestInitialize(t *testing.T) {
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
		name    string
		input   *JenkinsBuilds
		output  *JenkinsBuilds
		wantErr bool
	}{
		{
			name: "bad jenkins config",
			input: &JenkinsBuilds{
				Log:             testutil.Logger{},
				URL:             "http://a bad url",
				ResponseTimeout: config.Duration(time.Microsecond),
			},
			wantErr: true,
		},
		{
			name: "has filter",
			input: &JenkinsBuilds{
				Log:             testutil.Logger{},
				URL:             ts.URL,
				ResponseTimeout: config.Duration(time.Microsecond),
				JobInclude:      []string{"jobA", "jobB"},
				JobExclude:      []string{"job1", "job2"},
			},
		},
		{
			name: "default config",
			input: &JenkinsBuilds{
				Log:             testutil.Logger{},
				URL:             ts.URL,
				ResponseTimeout: config.Duration(time.Microsecond),
			},
			output: &JenkinsBuilds{
				Log:                testutil.Logger{},
				MaxIdleConnections: 5,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			te := test.input.initialize(mockClient)
			if !test.wantErr && te != nil {
				t.Fatalf("%s: failed %s, expected to be nil", test.name, te.Error())
			} else if test.wantErr && te == nil {
				t.Fatalf("%s: expected err, got nil", test.name)
			}
			if test.output != nil {
				if test.input.client == nil {
					t.Fatalf("%s: failed %v, jenkins instance shouldn't be nil", test.name, te)
				}
				if test.input.MaxIdleConnections != test.output.MaxIdleConnections {
					t.Fatalf("%s: different MaxConnections Expected %d, got %d\n", test.name, test.output.MaxIdleConnections, test.input.MaxIdleConnections)
				}
			}
		})
	}
}
