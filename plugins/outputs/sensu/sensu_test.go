package sensu

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/testutil"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/require"
)

func TestResolveEventEndpointUrl(t *testing.T) {
	agentAPIURL := "http://127.0.0.1:3031"
	backendAPIURL := "http://127.0.0.1:8080"
	entityNamespace := "test-namespace"
	emptyString := ""
	tests := []struct {
		name                string
		plugin              *Sensu
		expectedEndpointURL string
	}{
		{
			name: "agent event endpoint",
			plugin: &Sensu{
				AgentAPIURL: &agentAPIURL,
				Log:         testutil.Logger{},
			},
			expectedEndpointURL: "http://127.0.0.1:3031/events",
		},
		{
			name: "backend event endpoint with default namespace",
			plugin: &Sensu{
				AgentAPIURL:   &agentAPIURL,
				BackendAPIURL: &backendAPIURL,
				Log:           testutil.Logger{},
			},
			expectedEndpointURL: "http://127.0.0.1:8080/api/core/v2/namespaces/default/events",
		},
		{
			name: "backend event endpoint with namespace declared",
			plugin: &Sensu{
				AgentAPIURL:   &agentAPIURL,
				BackendAPIURL: &backendAPIURL,
				Entity: &SensuEntity{
					Namespace: &entityNamespace,
				},
				Log: testutil.Logger{},
			},
			expectedEndpointURL: "http://127.0.0.1:8080/api/core/v2/namespaces/test-namespace/events",
		},
		{
			name: "agent event endpoint due to empty AgentAPIURL",
			plugin: &Sensu{
				AgentAPIURL: &emptyString,
				Log:         testutil.Logger{},
			},
			expectedEndpointURL: "http://127.0.0.1:3031/events",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plugin.setEndpointURL()
			require.Equal(t, err, error(nil))
			require.Equal(t, tt.expectedEndpointURL, tt.plugin.EndpointURL)
		})
	}
}

func TestConnectAndWrite(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	testURL := fmt.Sprintf("http://%s", ts.Listener.Addr().String())
	testAPIKey := "a0b1c2d3-e4f5-g6h7-i8j9-k0l1m2n3o4p5"
	testCheck := "telegraf"
	testEntity := "entity1"
	testNamespace := "default"
	testHandler := "influxdb"
	testTagName := "myTagName"
	testTagValue := "myTagValue"
	expectedAuthHeader := fmt.Sprintf("Key %s", testAPIKey)
	expectedURL := fmt.Sprintf("/api/core/v2/namespaces/%s/events", testNamespace)
	expectedPointName := "cpu"
	expectedPointValue := float64(42)

	plugin := &Sensu{
		AgentAPIURL:   nil,
		BackendAPIURL: &testURL,
		APIKey:        &testAPIKey,
		Check: &SensuCheck{
			Name: &testCheck,
		},
		Entity: &SensuEntity{
			Name:      &testEntity,
			Namespace: &testNamespace,
		},
		Metrics: &SensuMetrics{
			Handlers: []string{testHandler},
		},
		Tags: map[string]string{testTagName: testTagValue},
		Log:  testutil.Logger{},
	}

	t.Run("connect", func(t *testing.T) {
		err := plugin.Connect()
		require.NoError(t, err)
	})

	t.Run("write", func(t *testing.T) {
		ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, expectedURL, r.URL.String())
			require.Equal(t, expectedAuthHeader, r.Header.Get("Authorization"))
			// let's make sure what we received is a valid Sensu event that contains all of the expected data
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			receivedEvent := &corev2.Event{}
			err = json.Unmarshal(body, receivedEvent)
			require.NoError(t, err)
			require.Equal(t, testCheck, receivedEvent.Check.Name)
			require.Equal(t, testEntity, receivedEvent.Entity.Name)
			require.NotEmpty(t, receivedEvent.Metrics)
			require.Equal(t, true, choice.Contains(testHandler, receivedEvent.Metrics.Handlers))
			require.NotEmpty(t, receivedEvent.Metrics.Points)
			pointFound := false
			tagFound := false
			for _, p := range receivedEvent.Metrics.Points {
				if p.Name == expectedPointName+".value" && p.Value == expectedPointValue {
					pointFound = true
					require.NotEmpty(t, p.Tags)
					for _, t := range p.Tags {
						if t.Name == testTagName && t.Value == testTagValue {
							tagFound = true
						}
					}
				}
			}
			require.Equal(t, true, pointFound)
			require.Equal(t, true, tagFound)
			w.WriteHeader(http.StatusCreated)
		})
		err := plugin.Write([]telegraf.Metric{testutil.TestMetric(expectedPointValue, expectedPointName)})
		require.NoError(t, err)
	})
}

func TestGetFloat(t *testing.T) {
	tests := []struct {
		name           string
		value          interface{}
		expectedReturn float64
	}{
		{
			name:           "getfloat with float64",
			value:          float64(42),
			expectedReturn: 42,
		},
		{
			name:           "getfloat with float32",
			value:          float32(42),
			expectedReturn: 42,
		},
		{
			name:           "getfloat with int64",
			value:          int64(42),
			expectedReturn: 42,
		},
		{
			name:           "getfloat with int32",
			value:          int32(42),
			expectedReturn: 42,
		},
		{
			name:           "getfloat with int",
			value:          int(42),
			expectedReturn: 42,
		},
		{
			name:           "getfloat with uint64",
			value:          uint64(42),
			expectedReturn: 42,
		},
		{
			name:           "getfloat with uint32",
			value:          uint32(42),
			expectedReturn: 42,
		},
		{
			name:           "getfloat with uint",
			value:          uint(42),
			expectedReturn: 42,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expectedReturn, getFloat(tt.value))
		})
	}
	// Since math.NaN() == math.NaN() returns false
	t.Run("getfloat NaN special case", func(t *testing.T) {
		f := getFloat("42")
		require.True(t, math.IsNaN(f))
	})
}
