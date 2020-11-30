package config_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/config"
	"github.com/stretchr/testify/require"

	_ "github.com/influxdata/telegraf/plugins/inputs/all"
)

// TestListPluginTypes tests that the config api can scrape all existing plugins
// for type information to build a schema.
func TestListPluginTypes(t *testing.T) {
	config.NewConfig() // initalizes API
	pluginConfigs := config.API.ListPluginTypes()
	require.Greater(t, len(pluginConfigs), 10)
	b, _ := json.Marshal(pluginConfigs)
	fmt.Println(string(b))

	// find the gnmi plugin
	var gnmi config.PluginConfig
	for _, conf := range pluginConfigs {
		if conf.Name == "gnmi" {
			gnmi = conf
			break
		}
	}

	// find the cloudwatch plugin
	var cloudwatch config.PluginConfig
	for _, conf := range pluginConfigs {
		if conf.Name == "cloudwatch" {
			cloudwatch = conf
			break
		}
	}

	// validate a slice of objects
	require.EqualValues(t, "array", gnmi.Config["Subscriptions"].Type)
	require.EqualValues(t, "object", gnmi.Config["Subscriptions"].SubType)
	require.NotNil(t, gnmi.Config["Subscriptions"].SubFields)
	require.EqualValues(t, "string", gnmi.Config["Subscriptions"].SubFields["Name"].Type)

	// validate a slice of pointer objects
	require.EqualValues(t, "array", cloudwatch.Config["Metrics"].Type)
	require.EqualValues(t, "object", cloudwatch.Config["Metrics"].SubType)
	require.NotNil(t, cloudwatch.Config["Metrics"].SubFields)
	require.EqualValues(t, "array", cloudwatch.Config["Metrics"].SubFields["StatisticExclude"].Type)
	require.EqualValues(t, "array", cloudwatch.Config["Metrics"].SubFields["MetricNames"].Type)

	// validate a map of strings
	require.EqualValues(t, "map", gnmi.Config["Aliases"].Type)
	require.EqualValues(t, "string", gnmi.Config["Aliases"].SubType)

	// check a default value
	require.EqualValues(t, "proto", gnmi.Config["Encoding"].Default)
	require.EqualValues(t, 10*1e9, gnmi.Config["Redial"].Default)

	// check anonymous composed fields
	require.EqualValues(t, "bool", gnmi.Config["InsecureSkipVerify"].Type)

	// check named composed fields

}
