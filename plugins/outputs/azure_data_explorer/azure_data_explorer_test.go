package azure_data_explorer

import (
	"testing"

	"github.com/stretchr/testify/require"

	common_adx "github.com/influxdata/telegraf/plugins/common/adx"
	"github.com/influxdata/telegraf/testutil"
)

func TestInit(t *testing.T) {
	plugin := AzureDataExplorer{
		Log:    testutil.Logger{},
		client: &common_adx.Client{},
		Config: common_adx.Config{
			Endpoint: "someendpoint",
		},
	}

	err := plugin.Init()
	require.NoError(t, err)
}

func TestConnectBlankEndpointData(t *testing.T) {
	plugin := AzureDataExplorer{
		Log: testutil.Logger{},
	}
	require.ErrorContains(t, plugin.Connect(), "endpoint configuration cannot be empty")
}
