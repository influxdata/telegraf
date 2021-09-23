package burrow

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

// remap uri to json file, eg: /v3/kafka -> ./testdata/v3_kafka.json
func getResponseJSON(requestURI string) ([]byte, int) {
	uri := strings.TrimLeft(requestURI, "/")
	mappedFile := strings.Replace(uri, "/", "_", -1)
	jsonFile := fmt.Sprintf("./testdata/%s.json", mappedFile)

	code := 200
	_, err := os.Stat(jsonFile)
	if err != nil {
		code = 404
		jsonFile = "./testdata/error.json"
	}

	// respond with file
	b, _ := os.ReadFile(jsonFile)
	return b, code
}

// return mocked HTTP server
func getHTTPServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, code := getResponseJSON(r.RequestURI)
		w.WriteHeader(code)
		w.Header().Set("Content-Type", "application/json")
		// Ignore the returned error as the test will fail anyway
		//nolint:errcheck,revive
		w.Write(body)
	}))
}

// return mocked HTTP server with basic auth
func getHTTPServerBasicAuth() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		username, password, authOK := r.BasicAuth()
		if !authOK {
			http.Error(w, "Not authorized", 401)
			return
		}

		if username != "test" && password != "test" {
			http.Error(w, "Not authorized", 401)
			return
		}

		// ok, continue
		body, code := getResponseJSON(r.RequestURI)
		w.WriteHeader(code)
		w.Header().Set("Content-Type", "application/json")
		// Ignore the returned error as the test will fail anyway
		//nolint:errcheck,revive
		w.Write(body)
	}))
}

// test burrow_topic measurement
func TestBurrowTopic(t *testing.T) {
	s := getHTTPServer()
	defer s.Close()

	plugin := &burrow{Servers: []string{s.URL}}
	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Gather(acc))

	fields := []map[string]interface{}{
		// topicA
		{"offset": int64(459178195)},
		{"offset": int64(459178022)},
		{"offset": int64(456491598)},
	}
	tags := []map[string]string{
		// topicA
		{"cluster": "clustername1", "topic": "topicA", "partition": "0"},
		{"cluster": "clustername1", "topic": "topicA", "partition": "1"},
		{"cluster": "clustername1", "topic": "topicA", "partition": "2"},
	}

	require.Empty(t, acc.Errors)
	require.Equal(t, true, acc.HasMeasurement("burrow_topic"))
	for i := 0; i < len(fields); i++ {
		acc.AssertContainsTaggedFields(t, "burrow_topic", fields[i], tags[i])
	}
}

// test burrow_partition measurement
func TestBurrowPartition(t *testing.T) {
	s := getHTTPServer()
	defer s.Close()

	plugin := &burrow{
		Servers: []string{s.URL},
	}
	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Gather(acc))

	fields := []map[string]interface{}{
		{
			"status":      "OK",
			"status_code": 1,
			"lag":         int64(0),
			"offset":      int64(431323195),
			"timestamp":   int64(1515609490008),
		},
		{
			"status":      "OK",
			"status_code": 1,
			"lag":         int64(0),
			"offset":      int64(431322962),
			"timestamp":   int64(1515609490008),
		},
		{
			"status":      "OK",
			"status_code": 1,
			"lag":         int64(0),
			"offset":      int64(428636563),
			"timestamp":   int64(1515609490008),
		},
	}
	tags := []map[string]string{
		{"cluster": "clustername1", "group": "group1", "topic": "topicA", "partition": "0", "owner": "kafka1"},
		{"cluster": "clustername1", "group": "group1", "topic": "topicA", "partition": "1", "owner": "kafka2"},
		{"cluster": "clustername1", "group": "group1", "topic": "topicA", "partition": "2", "owner": "kafka3"},
	}

	require.Empty(t, acc.Errors)
	require.Equal(t, true, acc.HasMeasurement("burrow_partition"))

	for i := 0; i < len(fields); i++ {
		acc.AssertContainsTaggedFields(t, "burrow_partition", fields[i], tags[i])
	}
}

// burrow_group
func TestBurrowGroup(t *testing.T) {
	s := getHTTPServer()
	defer s.Close()

	plugin := &burrow{
		Servers: []string{s.URL},
	}
	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Gather(acc))

	fields := []map[string]interface{}{
		{
			"status":          "OK",
			"status_code":     1,
			"partition_count": 3,
			"total_lag":       int64(0),
			"lag":             int64(0),
			"offset":          int64(431323195 + 431322962 + 428636563),
			"timestamp":       int64(1515609490008),
		},
	}

	tags := []map[string]string{
		{"cluster": "clustername1", "group": "group1"},
	}

	require.Empty(t, acc.Errors)
	require.Equal(t, true, acc.HasMeasurement("burrow_group"))

	for i := 0; i < len(fields); i++ {
		acc.AssertContainsTaggedFields(t, "burrow_group", fields[i], tags[i])
	}
}

// collect from multiple servers
func TestMultipleServers(t *testing.T) {
	s1 := getHTTPServer()
	defer s1.Close()

	s2 := getHTTPServer()
	defer s2.Close()

	plugin := &burrow{
		Servers: []string{s1.URL, s2.URL},
	}
	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Gather(acc))

	require.Exactly(t, 14, len(acc.Metrics))
	require.Empty(t, acc.Errors)
}

// collect multiple times
func TestMultipleRuns(t *testing.T) {
	s := getHTTPServer()
	defer s.Close()

	plugin := &burrow{
		Servers: []string{s.URL},
	}
	for i := 0; i < 4; i++ {
		acc := &testutil.Accumulator{}
		require.NoError(t, plugin.Gather(acc))

		require.Exactly(t, 7, len(acc.Metrics))
		require.Empty(t, acc.Errors)
	}
}

// collect from http basic auth server
func TestBasicAuthConfig(t *testing.T) {
	s := getHTTPServerBasicAuth()
	defer s.Close()

	plugin := &burrow{
		Servers:  []string{s.URL},
		Username: "test",
		Password: "test",
	}

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Gather(acc))

	require.Exactly(t, 7, len(acc.Metrics))
	require.Empty(t, acc.Errors)
}

// collect from whitelisted clusters
func TestFilterClusters(t *testing.T) {
	s := getHTTPServer()
	defer s.Close()

	plugin := &burrow{
		Servers:         []string{s.URL},
		ClustersInclude: []string{"wrongname*"}, // clustername1 -> no match
	}

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Gather(acc))

	// no match by cluster
	require.Exactly(t, 0, len(acc.Metrics))
	require.Empty(t, acc.Errors)
}

// collect from whitelisted groups
func TestFilterGroups(t *testing.T) {
	s := getHTTPServer()
	defer s.Close()

	plugin := &burrow{
		Servers:       []string{s.URL},
		GroupsInclude: []string{"group?"}, // group1 -> match
		TopicsExclude: []string{"*"},      // exclude all
	}

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Gather(acc))

	require.Exactly(t, 1, len(acc.Metrics))
	require.Empty(t, acc.Errors)
}

// collect from whitelisted topics
func TestFilterTopics(t *testing.T) {
	s := getHTTPServer()
	defer s.Close()

	plugin := &burrow{
		Servers:       []string{s.URL},
		TopicsInclude: []string{"topic?"}, // topicA -> match
		GroupsExclude: []string{"*"},      // exclude all
	}

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Gather(acc))

	require.Exactly(t, 3, len(acc.Metrics))
	require.Empty(t, acc.Errors)
}
