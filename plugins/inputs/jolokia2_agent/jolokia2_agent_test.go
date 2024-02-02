package jolokia2_agent_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	common "github.com/influxdata/telegraf/plugins/common/jolokia2"
	"github.com/influxdata/telegraf/plugins/inputs/jolokia2_agent"
	"github.com/influxdata/telegraf/testutil"
)

func TestScalarValues(t *testing.T) {
	config := `
	[jolokia2_agent]
		urls = ["%s"]

	[[jolokia2_agent.metric]]
		name  = "scalar_without_attribute"
		mbean = "scalar_without_attribute"

	[[jolokia2_agent.metric]]
		name  = "scalar_with_attribute"
		mbean = "scalar_with_attribute"
		paths = ["biz"]

	[[jolokia2_agent.metric]]
		name  = "scalar_with_attribute_and_path"
		mbean = "scalar_with_attribute_and_path"
		paths = ["biz/baz"]

	# This should return multiple series with different test tags.
	[[jolokia2_agent.metric]]
		name     = "scalar_with_key_pattern"
		mbean    = "scalar_with_key_pattern:test=*"
		tag_keys = ["test"]`

	response := `[{
		"request": {
			"mbean": "scalar_without_attribute",
			"type": "read"
		},
		"value": 123,
		"status": 200
	  }, {
		"request": {
			"mbean": "scalar_with_attribute",
			"attribute": "biz",
			"type": "read"
		},
		"value": 456,
		"status": 200
	  }, {
		"request": {
			"mbean": "scalar_with_attribute_and_path",
			"attribute": "biz",
			"path": "baz",
			"type": "read"
		},
		"value": 789,
		"status": 200
	  }, {
		"request": {
			"mbean": "scalar_with_key_pattern:test=*",
			"type": "read"
		},
		"value": {
			"scalar_with_key_pattern:test=foo": 123,
			"scalar_with_key_pattern:test=bar": 456
		},
		"status": 200
	  }]`

	server := setupServer(response)
	defer server.Close()
	plugin := SetupPlugin(t, fmt.Sprintf(config, server.URL))

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	acc.AssertContainsTaggedFields(t, "scalar_without_attribute", map[string]interface{}{
		"value": 123.0,
	}, map[string]string{
		"jolokia_agent_url": server.URL,
	})

	acc.AssertContainsTaggedFields(t, "scalar_with_attribute", map[string]interface{}{
		"biz": 456.0,
	}, map[string]string{
		"jolokia_agent_url": server.URL,
	})

	acc.AssertContainsTaggedFields(t, "scalar_with_attribute_and_path", map[string]interface{}{
		"biz.baz": 789.0,
	}, map[string]string{
		"jolokia_agent_url": server.URL,
	})

	acc.AssertContainsTaggedFields(t, "scalar_with_key_pattern", map[string]interface{}{
		"value": 123.0,
	}, map[string]string{
		"jolokia_agent_url": server.URL,
		"test":              "foo",
	})
	acc.AssertContainsTaggedFields(t, "scalar_with_key_pattern", map[string]interface{}{
		"value": 456.0,
	}, map[string]string{
		"jolokia_agent_url": server.URL,
		"test":              "bar",
	})
}

func TestObjectValues(t *testing.T) {
	config := `
	[jolokia2_agent]
		urls = ["%s"]

	[[jolokia2_agent.metric]]
		name     = "object_without_attribute"
		mbean    = "object_without_attribute"
		tag_keys = ["foo"]

	[[jolokia2_agent.metric]]
		name  = "object_with_attribute"
		mbean = "object_with_attribute"
		paths = ["biz"]

	[[jolokia2_agent.metric]]
		name  = "object_with_attribute_and_path"
		mbean = "object_with_attribute_and_path"
		paths = ["biz/baz"]

	# This will generate two separate request objects.
	[[jolokia2_agent.metric]]
		name  = "object_with_branching_paths"
		mbean = "object_with_branching_paths"
		paths = ["foo/fiz", "foo/faz"]

	# This should return multiple series with different test tags.
	[[jolokia2_agent.metric]]
		name     = "object_with_key_pattern"
		mbean    = "object_with_key_pattern:test=*"
		tag_keys = ["test"]

	[[jolokia2_agent.metric]]
		name  = "ColumnFamily"
		mbean = "org.apache.cassandra.metrics:keyspace=*,name=EstimatedRowSizeHistogram,scope=schema_columns,type=ColumnFamily"
		tag_keys = ["keyspace", "name", "scope"]`

	response, err := os.ReadFile("./testdata/response.json")
	require.NoError(t, err)

	server := setupServer(string(response))
	defer server.Close()
	plugin := SetupPlugin(t, fmt.Sprintf(config, server.URL))

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	acc.AssertContainsTaggedFields(t, "object_without_attribute", map[string]interface{}{
		"biz": 123.0,
		"baz": 456.0,
	}, map[string]string{
		"jolokia_agent_url": server.URL,
	})

	acc.AssertContainsTaggedFields(t, "object_with_attribute", map[string]interface{}{
		"biz.fiz": 123.0,
		"biz.faz": 456.0,
	}, map[string]string{
		"jolokia_agent_url": server.URL,
	})

	acc.AssertContainsTaggedFields(t, "object_with_attribute_and_path", map[string]interface{}{
		"biz.baz.bing": 123.0,
		"biz.baz.bang": 456.0,
	}, map[string]string{
		"jolokia_agent_url": server.URL,
	})

	acc.AssertContainsTaggedFields(t, "object_with_branching_paths", map[string]interface{}{
		"foo.fiz.bing": 123.0,
		"foo.faz.bang": 456.0,
	}, map[string]string{
		"jolokia_agent_url": server.URL,
	})

	acc.AssertContainsTaggedFields(t, "object_with_key_pattern", map[string]interface{}{
		"fiz": 123.0,
	}, map[string]string{
		"test":              "foo",
		"jolokia_agent_url": server.URL,
	})

	acc.AssertContainsTaggedFields(t, "object_with_key_pattern", map[string]interface{}{
		"biz": 456.0,
	}, map[string]string{
		"test":              "bar",
		"jolokia_agent_url": server.URL,
	})
}

func TestStatusCodes(t *testing.T) {
	config := `
	[jolokia2_agent]
		urls = ["%s"]

	[[jolokia2_agent.metric]]
		name     = "ok"
		mbean    = "ok"

	[[jolokia2_agent.metric]]
		name       = "not_found"
		mbean      = "not_found"

	[[jolokia2_agent.metric]]
		name       = "unknown"
		mbean      = "unknown"`

	response := `[{
		"request": {
			"mbean": "ok",
			"type": "read"
		},
		"value": 1,
		"status": 200
	}, {
		"request": {
			"mbean": "not_found",
			"type": "read"
		},
		"status": 404
	}, {
		"request": {
			"mbean": "unknown",
			"type": "read"
		},
		"status": 500
	}]`

	server := setupServer(response)
	defer server.Close()
	plugin := SetupPlugin(t, fmt.Sprintf(config, server.URL))

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	acc.AssertContainsTaggedFields(t, "ok", map[string]interface{}{
		"value": 1.0,
	}, map[string]string{
		"jolokia_agent_url": server.URL,
	})

	acc.AssertDoesNotContainMeasurement(t, "not_found")
	acc.AssertDoesNotContainMeasurement(t, "unknown")
}

func TestTagRenaming(t *testing.T) {
	config := `
	[jolokia2_agent]
		default_tag_prefix = "DEFAULT_PREFIX_"
		urls = ["%s"]

	[[jolokia2_agent.metric]]
		name     = "default_tag_prefix"
		mbean    = "default_tag_prefix:biz=baz,fiz=faz"
		tag_keys = ["biz", "fiz"]

	[[jolokia2_agent.metric]]
		name       = "custom_tag_prefix"
		mbean      = "custom_tag_prefix:biz=baz,fiz=faz"
		tag_keys   = ["biz", "fiz"]
		tag_prefix = "CUSTOM_PREFIX_"`

	response := `[{
		"request": {
			"mbean": "default_tag_prefix:biz=baz,fiz=faz",
			"type": "read"
		},
		"value": 123,
		"status": 200
	}, {
		"request": {
			"mbean": "custom_tag_prefix:biz=baz,fiz=faz",
			"type": "read"
		},
		"value": 123,
		"status": 200
	}]`

	server := setupServer(response)
	defer server.Close()
	plugin := SetupPlugin(t, fmt.Sprintf(config, server.URL))

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	acc.AssertContainsTaggedFields(t, "default_tag_prefix", map[string]interface{}{
		"value": 123.0,
	}, map[string]string{
		"DEFAULT_PREFIX_biz": "baz",
		"DEFAULT_PREFIX_fiz": "faz",
		"jolokia_agent_url":  server.URL,
	})

	acc.AssertContainsTaggedFields(t, "custom_tag_prefix", map[string]interface{}{
		"value": 123.0,
	}, map[string]string{
		"CUSTOM_PREFIX_biz": "baz",
		"CUSTOM_PREFIX_fiz": "faz",
		"jolokia_agent_url": server.URL,
	})
}

func TestFieldRenaming(t *testing.T) {
	config := `
	[jolokia2_agent]
		default_field_prefix    = "DEFAULT_PREFIX_"
		default_field_separator = "_DEFAULT_SEPARATOR_"

		urls = ["%s"]

	[[jolokia2_agent.metric]]
		name  = "default_field_modifiers"
		mbean = "default_field_modifiers"

	[[jolokia2_agent.metric]]
		name            = "custom_field_modifiers"
		mbean           = "custom_field_modifiers"
		field_prefix    = "CUSTOM_PREFIX_"
		field_separator = "_CUSTOM_SEPARATOR_"

	[[jolokia2_agent.metric]]
		name            = "field_prefix_substitution"
		mbean           = "field_prefix_substitution:foo=*"
		field_prefix    = "$1_"

	[[jolokia2_agent.metric]]
		name         = "field_name_substitution"
		mbean        = "field_name_substitution:foo=*"
		field_prefix = ""
		field_name   = "$1"`

	response := `[{
		"request": {
			"mbean": "default_field_modifiers",
			"type": "read"
		},
		"value": {
			"hello": { "world": 123 }
		},
		"status": 200
	}, {
		"request": {
			"mbean": "custom_field_modifiers",
			"type": "read"
		},
		"value": {
			"hello": { "world": 123 }
		},
		"status": 200
	}, {
		"request": {
			"mbean": "field_prefix_substitution:foo=*",
			"type": "read"
		},
		"value": {
			"field_prefix_substitution:foo=biz": 123,
			"field_prefix_substitution:foo=baz": 456
		},
		"status": 200
	}, {
		"request": {
			"mbean": "field_name_substitution:foo=*",
			"type": "read"
		},
		"value": {
			"field_name_substitution:foo=biz": 123,
			"field_name_substitution:foo=baz": 456
		},
		"status": 200
	}]`

	server := setupServer(response)
	defer server.Close()
	plugin := SetupPlugin(t, fmt.Sprintf(config, server.URL))

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	acc.AssertContainsTaggedFields(t, "default_field_modifiers", map[string]interface{}{
		"DEFAULT_PREFIX_hello_DEFAULT_SEPARATOR_world": 123.0,
	}, map[string]string{
		"jolokia_agent_url": server.URL,
	})

	acc.AssertContainsTaggedFields(t, "custom_field_modifiers", map[string]interface{}{
		"CUSTOM_PREFIX_hello_CUSTOM_SEPARATOR_world": 123.0,
	}, map[string]string{
		"jolokia_agent_url": server.URL,
	})

	acc.AssertContainsTaggedFields(t, "field_prefix_substitution", map[string]interface{}{
		"biz_value": 123.0,
		"baz_value": 456.0,
	}, map[string]string{
		"jolokia_agent_url": server.URL,
	})

	acc.AssertContainsTaggedFields(t, "field_name_substitution", map[string]interface{}{
		"biz": 123.0,
		"baz": 456.0,
	}, map[string]string{
		"jolokia_agent_url": server.URL,
	})
}

func TestMetricMbeanMatching(t *testing.T) {
	config := `
	[jolokia2_agent]
		urls = ["%s"]

	[[jolokia2_agent.metric]]
		name = "mbean_name_and_object_keys"
		mbean = "test1:foo=bar,fizz=buzz"

	[[jolokia2_agent.metric]]
		name = "mbean_name_and_unordered_object_keys"
		mbean = "test2:fizz=buzz,foo=bar"

	[[jolokia2_agent.metric]]
		name = "mbean_name_and_attributes"
		mbean = "test3"
		paths = ["foo", "bar"]

	[[jolokia2_agent.metric]]
		name = "mbean_name_and_attribute_with_paths"
		mbean = "test4"
		paths = ["flavor/chocolate", "flavor/strawberry"]
	`

	response := `[{
		"request": {
			"mbean": "test1:foo=bar,fizz=buzz",
			"type": "read"
		},
		"value": 123,
		"status": 200
	}, {
		"request": {
			"mbean": "test2:foo=bar,fizz=buzz",
			"type": "read"
		},
		"value": 123,
		"status": 200
	}, {
		"request": {
			"mbean": "test3",
			"attribute": "foo",
			"type": "read"
		},
		"value": 123,
		"status": 200
	}, {
		"request": {
			"mbean": "test3",
			"attribute": "bar",
			"type": "read"
		},
		"value": 456,
		"status": 200
	}, {
		"request": {
			"mbean": "test4",
			"attribute": "flavor",
			"path": "chocolate",
			"type": "read"
		},
		"value": 123,
		"status": 200
	}, {
		"request": {
			"mbean": "test4",
			"attribute": "flavor",
			"path": "strawberry",
			"type": "read"
		},
		"value": 456,
		"status": 200
	}]`

	server := setupServer(response)
	defer server.Close()
	plugin := SetupPlugin(t, fmt.Sprintf(config, server.URL))

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	acc.AssertContainsTaggedFields(t, "mbean_name_and_object_keys", map[string]interface{}{
		"value": 123.0,
	}, map[string]string{
		"jolokia_agent_url": server.URL,
	})

	acc.AssertContainsTaggedFields(t, "mbean_name_and_unordered_object_keys", map[string]interface{}{
		"value": 123.0,
	}, map[string]string{
		"jolokia_agent_url": server.URL,
	})

	acc.AssertContainsTaggedFields(t, "mbean_name_and_attributes", map[string]interface{}{
		"foo": 123.0,
		"bar": 456.0,
	}, map[string]string{
		"jolokia_agent_url": server.URL,
	})

	acc.AssertContainsTaggedFields(t, "mbean_name_and_attribute_with_paths", map[string]interface{}{
		"flavor.chocolate":  123.0,
		"flavor.strawberry": 456.0,
	}, map[string]string{
		"jolokia_agent_url": server.URL,
	})
}

func TestMetricCompaction(t *testing.T) {
	config := `
	[jolokia2_agent]
		urls = ["%s"]

	[[jolokia2_agent.metric]]
		name     = "compact_metric"
		mbean    = "scalar_value:flavor=chocolate"
		tag_keys = ["flavor"]

	[[jolokia2_agent.metric]]
		name     = "compact_metric"
		mbean    = "scalar_value:flavor=vanilla"
		tag_keys = ["flavor"]

	[[jolokia2_agent.metric]]
		name     = "compact_metric"
		mbean    = "object_value1:flavor=chocolate"
		tag_keys = ["flavor"]

	[[jolokia2_agent.metric]]
		name     = "compact_metric"
		mbean    = "object_value2:flavor=chocolate"
		tag_keys = ["flavor"]`

	response := `[{
		"request": {
			"mbean": "scalar_value:flavor=chocolate",
			"type": "read"
		},
		"value": 123,
		"status": 200
	}, {
		"request": {
			"mbean": "scalar_value:flavor=vanilla",
			"type": "read"
		},
		"value": 999,
		"status": 200
	}, {
		"request": {
			"mbean": "object_value1:flavor=chocolate",
			"type": "read"
		},
		"value": {
			"foo": 456
		},
		"status": 200
	}, {
		"request": {
			"mbean": "object_value2:flavor=chocolate",
			"type": "read"
		},
		"value": {
			"bar": 789
		},
		"status": 200
	}]`

	server := setupServer(response)
	defer server.Close()
	plugin := SetupPlugin(t, fmt.Sprintf(config, server.URL))

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	acc.AssertContainsTaggedFields(t, "compact_metric", map[string]interface{}{
		"value": 123.0,
		"foo":   456.0,
		"bar":   789.0,
	}, map[string]string{
		"flavor":            "chocolate",
		"jolokia_agent_url": server.URL,
	})

	acc.AssertContainsTaggedFields(t, "compact_metric", map[string]interface{}{
		"value": 999.0,
	}, map[string]string{
		"flavor":            "vanilla",
		"jolokia_agent_url": server.URL,
	})
}

func TestJolokia2_ClientAuthRequest(t *testing.T) {
	var username string
	var password string
	var requests []map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, _ = r.BasicAuth()

		body, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(body, &requests))

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	plugin := SetupPlugin(t, fmt.Sprintf(`
		[jolokia2_agent]
			urls = ["%s/jolokia"]
			username = "sally"
			password = "seashore"
		[[jolokia2_agent.metric]]
			name  = "hello"
			mbean = "hello:foo=bar"
	`, server.URL))

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	require.EqualValuesf(t, "sally", username, "Expected to post with username %s, but was %s", "sally", username)
	require.EqualValuesf(t, "seashore", password, "Expected to post with password %s, but was %s", "seashore", password)
	require.NotZero(t, len(requests), "Expected to post a request body, but was empty.")

	request := requests[0]["mbean"]
	require.EqualValuesf(t, "hello:foo=bar", request, "Expected to query mbean %s, but was %s", "hello:foo=bar", request)
}

func TestFillFields(t *testing.T) {
	complexPoint := map[string]interface{}{"Value": []interface{}{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}}
	scalarPoint := []interface{}{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	results := map[string]interface{}{}
	common.NewPointBuilder(common.Metric{Name: "test", Mbean: "complex"}, []string{"this", "that"}, "/").FillFields("", complexPoint, results)
	require.Equal(t, map[string]interface{}{}, results)

	results = map[string]interface{}{}
	common.NewPointBuilder(common.Metric{Name: "test", Mbean: "scalar"}, []string{"this", "that"}, "/").FillFields("", scalarPoint, results)
	require.Equal(t, map[string]interface{}{}, results)
}

func TestIntegrationArtemis(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Start the docker container
	container := testutil.Container{
		Image:        "apache/activemq-artemis",
		ExposedPorts: []string{"61616", "8161"},
		WaitingFor: wait.ForAll(
			wait.ForLog("Artemis Console available at"),
			wait.ForListeningPort(nat.Port("8161")),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	// Setup the plugin
	port := container.Ports["8161"]
	endpoint := "http://" + container.Address + ":" + port + "/console/jolokia/"
	plugin := &jolokia2_agent.JolokiaAgent{
		URLs:     []string{endpoint},
		Username: "artemis",
		Password: "artemis",
		Metrics: []common.MetricConfig{
			{
				Name:    "artemis",
				Mbean:   `org.apache.activemq.artemis:broker="*",component=addresses,address="*",subcomponent=queues,routing-type="anycast",queue="*"`,
				TagKeys: []string{"queue", "subcomponent"},
			},
		},
	}

	// Setup the expectations
	expected := []telegraf.Metric{
		metric.New(
			"artemis",
			map[string]string{
				"jolokia_agent_url": endpoint,
				"queue":             "ExpiryQueue",
				"subcomponent":      "queues",
			},
			map[string]interface{}{
				"AcknowledgeAttempts":             float64(0),
				"Address":                         "ExpiryQueue",
				"AutoDelete":                      false,
				"ConfigurationManaged":            false,
				"ConsumerCount":                   float64(0),
				"ConsumersBeforeDispatch":         float64(0),
				"DeadLetterAddress":               "DLQ",
				"DelayBeforeDispatch":             float64(0),
				"DeliveringCount":                 float64(0),
				"DeliveringSize":                  float64(0),
				"Durable":                         false,
				"DurableDeliveringCount":          float64(0),
				"DurableDeliveringSize":           float64(0),
				"DurableMessageCount":             float64(0),
				"DurablePersistentSize":           float64(0),
				"DurableScheduledCount":           float64(0),
				"DurableScheduledSize":            float64(0),
				"Enabled":                         true,
				"Exclusive":                       true,
				"ExpiryAddress":                   "ExpiryQueue",
				"FirstMessageAsJSON":              "[{}]",
				"GroupBuckets":                    float64(0),
				"GroupCount":                      float64(0),
				"GroupRebalance":                  false,
				"GroupRebalancePauseDispatch":     false,
				"ID":                              float64(0),
				"LastValue":                       false,
				"MaxConsumers":                    float64(0),
				"MessageCount":                    float64(0),
				"MessagesAcknowledged":            float64(0),
				"MessagesAdded":                   float64(0),
				"MessagesExpired":                 float64(0),
				"MessagesKilled":                  float64(0),
				"Name":                            "ExpiryQueue",
				"Paused":                          false,
				"PersistentSize":                  float64(0),
				"PreparedTransactionMessageCount": float64(0),
				"PurgeOnNoConsumers":              false,
				"RetroactiveResource":             false,
				"RingSize":                        float64(0),
				"RoutingType":                     "ANYCAST",
				"ScheduledCount":                  float64(0),
				"ScheduledSize":                   float64(0),
				"Temporary":                       false,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"artemis",
			map[string]string{
				"jolokia_agent_url": endpoint,
				"queue":             "DLQ",
				"subcomponent":      "queues",
			},
			map[string]interface{}{
				"AcknowledgeAttempts":             float64(0),
				"Address":                         "DLQ",
				"AutoDelete":                      false,
				"ConfigurationManaged":            false,
				"ConsumerCount":                   float64(0),
				"ConsumersBeforeDispatch":         float64(0),
				"DeadLetterAddress":               "DLQ",
				"DelayBeforeDispatch":             float64(0),
				"DeliveringCount":                 float64(0),
				"DeliveringSize":                  float64(0),
				"Durable":                         false,
				"DurableDeliveringCount":          float64(0),
				"DurableDeliveringSize":           float64(0),
				"DurableMessageCount":             float64(0),
				"DurablePersistentSize":           float64(0),
				"DurableScheduledCount":           float64(0),
				"DurableScheduledSize":            float64(0),
				"Enabled":                         true,
				"Exclusive":                       true,
				"ExpiryAddress":                   "ExpiryQueue",
				"FirstMessageAsJSON":              "[{}]",
				"GroupBuckets":                    float64(0),
				"GroupCount":                      float64(0),
				"GroupRebalance":                  false,
				"GroupRebalancePauseDispatch":     false,
				"ID":                              float64(0),
				"LastValue":                       false,
				"MaxConsumers":                    float64(0),
				"MessageCount":                    float64(0),
				"MessagesAcknowledged":            float64(0),
				"MessagesAdded":                   float64(0),
				"MessagesExpired":                 float64(0),
				"MessagesKilled":                  float64(0),
				"Name":                            "DLQ",
				"Paused":                          false,
				"PersistentSize":                  float64(0),
				"PreparedTransactionMessageCount": float64(0),
				"PurgeOnNoConsumers":              false,
				"RetroactiveResource":             false,
				"RingSize":                        float64(0),
				"RoutingType":                     "ANYCAST",
				"ScheduledCount":                  float64(0),
				"ScheduledSize":                   float64(0),
				"Temporary":                       false,
			},
			time.Unix(0, 0),
		),
	}

	// Collect the metrics and compare
	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsStructureEqual(t, expected, actual, testutil.SortMetrics(), testutil.IgnoreTime())
}

func setupServer(resp string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, resp)
	}))
}

func SetupPlugin(t *testing.T, conf string) telegraf.Input {
	table, err := toml.Parse([]byte(conf))
	if err != nil {
		t.Fatalf("Unable to parse config! %v", err)
	}

	for name := range table.Fields {
		object := table.Fields[name]
		if name == "jolokia2_agent" {
			plugin := jolokia2_agent.JolokiaAgent{
				Metrics:               []common.MetricConfig{},
				DefaultFieldSeparator: ".",
			}

			if err := toml.UnmarshalTable(object.(*ast.Table), &plugin); err != nil {
				t.Fatalf("Unable to parse jolokia_agent plugin config! %v", err)
			}

			return &plugin
		}
	}

	return nil
}
