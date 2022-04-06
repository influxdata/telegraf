package jolokia2_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs/jolokia2/common"
	"github.com/influxdata/telegraf/plugins/inputs/jolokia2/jolokia2_agent"
	"github.com/influxdata/telegraf/plugins/inputs/jolokia2/jolokia2_proxy"
	"github.com/influxdata/telegraf/testutil"
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
)

func TestJolokia2_ScalarValues(t *testing.T) {
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

func TestJolokia2_ObjectValues(t *testing.T) {
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

	response := `[{
		"request": {
			"mbean": "object_without_attribute",
			"type": "read"
		},
		"value": {
			"biz": 123,
			"baz": 456
		},
		"status": 200
	}, {
		"request": {
			"mbean": "object_with_attribute",
			"attribute": "biz",
			"type": "read"
		},
		"value": {
			"fiz": 123,
			"faz": 456
		},
		"status": 200
	}, {
		"request": {
			"mbean": "object_with_branching_paths",
			"attribute": "foo",
			"path": "fiz",
			"type": "read"
		},
		"value": {
			"bing": 123
		},
		"status": 200
	}, {
		"request": {
			"mbean": "object_with_branching_paths",
			"attribute": "foo",
			"path": "faz",
			"type": "read"
		},
		"value": {
			"bang": 456
		},
		"status": 200
	}, {
		"request": {
			"mbean": "object_with_attribute_and_path",
			"attribute": "biz",
			"path": "baz",
			"type": "read"
		},
		"value": {
			"bing": 123,
			"bang": 456
		},
		"status": 200
	}, {
		"request": {
			"mbean": "object_with_key_pattern:test=*",
			"type": "read"
		},
		"value": {
			"object_with_key_pattern:test=foo": {
				"fiz": 123
			},
			"object_with_key_pattern:test=bar": {
				"biz": 456
			}
		},
		"status": 200
	}, {
		"request": {
		  "mbean": "org.apache.cassandra.metrics:keyspace=*,name=EstimatedRowSizeHistogram,scope=schema_columns,type=ColumnFamily",
		  "type": "read"
		},
		"value": {
		  "org.apache.cassandra.metrics:keyspace=system,name=EstimatedRowSizeHistogram,scope=schema_columns,type=ColumnFamily": {
			"Value": [
				0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
			]
		  }
		},
		"status": 200
	  }]`

	server := setupServer(response)
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

func TestJolokia2_StatusCodes(t *testing.T) {
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

func TestJolokia2_TagRenaming(t *testing.T) {
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

func TestJolokia2_FieldRenaming(t *testing.T) {
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

func TestJolokia2_MetricMbeanMatching(t *testing.T) {
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

func TestJolokia2_MetricCompaction(t *testing.T) {
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

func TestJolokia2_ProxyTargets(t *testing.T) {
	config := `
	[jolokia2_proxy]
		url = "%s"

	[[jolokia2_proxy.target]]
		url = "service:jmx:rmi:///jndi/rmi://target1:9010/jmxrmi"

	[[jolokia2_proxy.target]]
		url = "service:jmx:rmi:///jndi/rmi://target2:9010/jmxrmi"

	[[jolokia2_proxy.metric]]
		name  = "hello"
		mbean = "hello:foo=bar"`

	response := `[{
		"request": {
			"type": "read",
			"mbean": "hello:foo=bar",
			"target": {
				"url": "service:jmx:rmi:///jndi/rmi://target1:9010/jmxrmi"
			}
		},
		"value": 123,
		"status": 200
	}, {
		"request": {
			"type": "read",
			"mbean": "hello:foo=bar",
			"target": {
				"url": "service:jmx:rmi:///jndi/rmi://target2:9010/jmxrmi"
			}
		},
		"value": 456,
		"status": 200
	}]`

	server := setupServer(response)
	defer server.Close()
	plugin := SetupPlugin(t, fmt.Sprintf(config, server.URL))

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	acc.AssertContainsTaggedFields(t, "hello", map[string]interface{}{
		"value": 123.0,
	}, map[string]string{
		"jolokia_proxy_url": server.URL,
		"jolokia_agent_url": "service:jmx:rmi:///jndi/rmi://target1:9010/jmxrmi",
	})
	acc.AssertContainsTaggedFields(t, "hello", map[string]interface{}{
		"value": 456.0,
	}, map[string]string{
		"jolokia_proxy_url": server.URL,
		"jolokia_agent_url": "service:jmx:rmi:///jndi/rmi://target2:9010/jmxrmi",
	})
}

func TestFillFields(t *testing.T) {
	complexPoint := map[string]interface{}{"Value": []interface{}{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}}
	scalarPoint := []interface{}{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	results := map[string]interface{}{}
	common.NewPointBuilder(common.Metric{Name: "test", Mbean: "complex"}, []string{"this", "that"}, "/").FillFields("", complexPoint, results)
	require.Equal(t, map[string]interface{}{}, results)

	results = map[string]interface{}{}
	common.NewPointBuilder(common.Metric{Name: "test", Mbean: "scalar"}, []string{"this", "that"}, "/").FillFields("", scalarPoint, results)
	require.Equal(t, map[string]interface{}{}, results)
}

func setupServer(resp string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Ignore the returned error as the tests will fail anyway
		//nolint:errcheck,revive
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
		switch name {
		case "jolokia2_agent":
			plugin := jolokia2_agent.JolokiaAgent{
				Metrics:               []common.MetricConfig{},
				DefaultFieldSeparator: ".",
			}

			if err := toml.UnmarshalTable(object.(*ast.Table), &plugin); err != nil {
				t.Fatalf("Unable to parse jolokia_agent plugin config! %v", err)
			}

			return &plugin

		case "jolokia2_proxy":
			plugin := jolokia2_proxy.JolokiaProxy{
				Metrics:               []common.MetricConfig{},
				DefaultFieldSeparator: ".",
			}

			if err := toml.UnmarshalTable(object.(*ast.Table), &plugin); err != nil {
				t.Fatalf("Unable to parse jolokia_proxy plugin config! %v", err)
			}

			return &plugin
		}
	}

	return nil
}
