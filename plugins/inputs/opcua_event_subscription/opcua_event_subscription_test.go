package opcua_event_subscription

import (
	"fmt"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type TempConfig struct {
	Endpoint       string   `toml:"endpoint"`
	Interval       string   `toml:"interval"`
	EventType      string   `toml:"event_type"`
	NodeIDs        []string `toml:"node_ids"`
	SourceNames    []string `toml:"source_names"`
	Fields         []string `toml:"fields"`
	SecurityMode   string   `toml:"security_mode"`
	SecurityPolicy string   `toml:"security_policy"`
	Certificate    string   `toml:"certificate"`
	PrivateKey     string   `toml:"private_key"`
}

func LoadSampleConfigToPlugin() (*OpcuaEventSubscription, error) {
	plugin := &OpcuaEventSubscription{}
	sampleConfig := plugin.SampleConfig()
	tempConfig := &TempConfig{}

	err := toml.Unmarshal([]byte(sampleConfig), tempConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal sample config: %v", err)
	}

	plugin.Endpoint = tempConfig.Endpoint
	plugin.Interval = config.Duration(time.Second * 10) // Default to 10s for simplicity
	plugin.EventType = NodeIDWrapper{}
	if err := plugin.EventType.UnmarshalText([]byte(tempConfig.EventType)); err != nil {
		return nil, fmt.Errorf("failed to unmarshal EventType: %v", err)
	}

	for _, nodeIDStr := range tempConfig.NodeIDs {
		nodeIDWrapper := NodeIDWrapper{}
		if err := nodeIDWrapper.UnmarshalText([]byte(nodeIDStr)); err != nil {
			return nil, fmt.Errorf("failed to unmarshal NodeID: %v", err)
		}
		plugin.NodeIDs = append(plugin.NodeIDs, nodeIDWrapper)
	}

	plugin.SourceNames = tempConfig.SourceNames
	plugin.Fields = tempConfig.Fields
	plugin.SecurityMode = tempConfig.SecurityMode
	plugin.SecurityPolicy = tempConfig.SecurityPolicy
	plugin.Certificate = tempConfig.Certificate
	plugin.PrivateKey = tempConfig.PrivateKey

	return plugin, nil
}

func TestStart(t *testing.T) {
	plugin, err := LoadSampleConfigToPlugin()
	require.NoError(t, err)

	plugin.Log = testutil.Logger{}

	acc := &testutil.Accumulator{}

	err = plugin.Start(acc)
	require.NoError(t, err)

	require.NotNil(t, plugin.ClientManager)
	require.NotNil(t, plugin.SubscriptionManager)
	require.NotNil(t, plugin.NotificationHandler)

	// Clean up after the test
	plugin.Stop()
}

func TestGather(t *testing.T) {
	plugin, err := LoadSampleConfigToPlugin()
	require.NoError(t, err)
	plugin.Log = testutil.Logger{}

	acc := &testutil.Accumulator{}

	err = plugin.Start(acc)
	require.NoError(t, err)

	err = plugin.Gather(acc)
	require.NoError(t, err)

	plugin.Stop()
}

func TestStop(t *testing.T) {
	plugin, err := LoadSampleConfigToPlugin()
	require.NoError(t, err)
	plugin.Log = testutil.Logger{}

	acc := &testutil.Accumulator{}

	err = plugin.Start(acc)
	require.NoError(t, err)

	plugin.SubscriptionManager = &SubscriptionManager{}
	plugin.ClientManager = &ClientManager{}

	plugin.Stop()
	require.Nil(t, plugin.Cancel)
	require.Nil(t, plugin.ClientManager.Client)
}
