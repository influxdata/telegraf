package opcua_listener

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/opcua"
	"github.com/influxdata/telegraf/plugins/common/opcua/input"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"
)

const servicePort = "4840"

type OPCTags struct {
	Name           string
	Namespace      string
	IdentifierType string
	Identifier     string
	Want           interface{}
}

func MapOPCTag(tags OPCTags) (out input.NodeSettings) {
	out.FieldName = tags.Name
	out.Namespace = tags.Namespace
	out.IdentifierType = tags.IdentifierType
	out.Identifier = tags.Identifier
	return out
}

func TestSubscribeClientIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := testutil.Container{
		Image:        "open62541/open62541",
		ExposedPorts: []string{servicePort},
		WaitingFor:   wait.ForListeningPort(nat.Port(servicePort)),
	}
	err := container.Start()
	require.NoError(t, err, "failed to start container")
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	var testopctags = []OPCTags{
		{"ProductName", "0", "i", "2261", "open62541 OPC UA Server"},
		{"ProductUri", "0", "i", "2262", "http://open62541.org"},
		{"ManufacturerName", "0", "i", "2263", "open62541"},
		{"badnode", "1", "i", "1337", nil},
		{"goodnode", "1", "s", "the.answer", int32(42)},
		{"DateTime", "1", "i", "51037", "0001-01-01T00:00:00Z"},
	}
	var tagsRemaining = make([]string, 0, len(testopctags))
	for i, tag := range testopctags {
		if tag.Want != nil {
			tagsRemaining = append(tagsRemaining, testopctags[i].Name)
		}
	}

	subscribeConfig := SubscribeClientConfig{
		InputClientConfig: input.InputClientConfig{
			OpcUAClientConfig: opcua.OpcUAClientConfig{
				Endpoint:       fmt.Sprintf("opc.tcp://%s:%s", container.Address, container.Ports[servicePort]),
				SecurityPolicy: "None",
				SecurityMode:   "None",
				AuthMethod:     "Anonymous",
				ConnectTimeout: config.Duration(10 * time.Second),
				RequestTimeout: config.Duration(1 * time.Second),
				Workarounds:    opcua.OpcUAWorkarounds{},
			},
			MetricName: "testing",
			RootNodes:  make([]input.NodeSettings, 0),
			Groups:     make([]input.NodeGroupSettings, 0),
		},
		SubscriptionInterval: 0,
	}
	for _, tags := range testopctags {
		subscribeConfig.RootNodes = append(subscribeConfig.RootNodes, MapOPCTag(tags))
	}
	o, err := subscribeConfig.CreateSubscribeClient(testutil.Logger{})
	require.NoError(t, err)

	err = o.Init()
	require.NoError(t, err, "Initialization")
	err = o.Connect()
	require.NoError(t, err, "Connect")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	res, err := o.StartStreamValues(ctx)
	require.NoError(t, err)

	for {
		select {
		case m := <-res:
			for fieldName, fieldValue := range m.Fields() {
				for _, tag := range testopctags {
					if fieldName != tag.Name {
						continue
					}

					if tag.Want == nil {
						t.Errorf("Tag: %s has value: %v", tag.Name, fieldValue)
						return
					}

					require.Equal(t, tag.Want, fieldValue)

					newRemaining := make([]string, 0, len(tagsRemaining))
					for _, remainingTag := range tagsRemaining {
						if fieldName != remainingTag {
							newRemaining = append(newRemaining, remainingTag)
							break
						}
					}

					if len(newRemaining) <= 0 {
						return
					}

					tagsRemaining = newRemaining
				}
			}

		case <-ctx.Done():
			msg := ""
			for _, tag := range tagsRemaining {
				msg += tag + ", "
			}

			t.Errorf("Tags %s are remaining without a received value", msg)
			return
		}
	}
}

func TestSubscribeClientConfig(t *testing.T) {
	toml := `
[[inputs.opcua_listener]]
name = "localhost"
endpoint = "opc.tcp://localhost:4840"
connect_timeout = "10s"
request_timeout = "5s"
subscription_interval = "200ms"
security_policy = "auto"
security_mode = "auto"
certificate = "/etc/telegraf/cert.pem"
private_key = "/etc/telegraf/key.pem"
auth_method = "Anonymous"
username = ""
password = ""
nodes = [
  {name="name",  namespace="1", identifier_type="s", identifier="one"},
  {name="name2", namespace="2", identifier_type="s", identifier="two"},
]

[[inputs.opcua_listener.group]]
name = "foo"
namespace = "3"
identifier_type = "i"
tags = [["tag1", "val1"], ["tag2", "val2"]]
nodes = [{name="name3", identifier="3000", tags=[["tag3", "val3"]]}]

[[inputs.opcua_listener.group]]
name = "bar"
namespace = "0"
identifier_type = "i"
tags = [["tag1", "val1"], ["tag2", "val2"]]
nodes = [{name="name4", identifier="4000", tags=[["tag1", "override"]]}]

[inputs.opcua_listener.workarounds]
additional_valid_status_codes = ["0xC0"]
`

	c := config.NewConfig()
	err := c.LoadConfigData([]byte(toml))
	require.NoError(t, err)

	require.Len(t, c.Inputs, 1)

	o, ok := c.Inputs[0].Input.(*OpcUaListener)
	require.True(t, ok)

	require.Equal(t, "localhost", o.SubscribeClientConfig.MetricName)
	require.Equal(t, "opc.tcp://localhost:4840", o.SubscribeClientConfig.Endpoint)
	require.Equal(t, config.Duration(10*time.Second), o.SubscribeClientConfig.ConnectTimeout)
	require.Equal(t, config.Duration(5*time.Second), o.SubscribeClientConfig.RequestTimeout)
	require.Equal(t, config.Duration(200*time.Millisecond), o.SubscribeClientConfig.SubscriptionInterval)
	require.Equal(t, "auto", o.SubscribeClientConfig.SecurityPolicy)
	require.Equal(t, "auto", o.SubscribeClientConfig.SecurityMode)
	require.Equal(t, "/etc/telegraf/cert.pem", o.SubscribeClientConfig.Certificate)
	require.Equal(t, "/etc/telegraf/key.pem", o.SubscribeClientConfig.PrivateKey)
	require.Equal(t, "Anonymous", o.SubscribeClientConfig.AuthMethod)
	require.Equal(t, "", o.SubscribeClientConfig.Username)
	require.Equal(t, "", o.SubscribeClientConfig.Password)
	require.Equal(t, []input.NodeSettings{
		{
			FieldName:      "name",
			Namespace:      "1",
			IdentifierType: "s",
			Identifier:     "one",
		},
		{
			FieldName:      "name2",
			Namespace:      "2",
			IdentifierType: "s",
			Identifier:     "two",
		},
	}, o.SubscribeClientConfig.RootNodes)
	require.Equal(t, []input.NodeGroupSettings{
		{
			MetricName:     "foo",
			Namespace:      "3",
			IdentifierType: "i",
			TagsSlice:      [][]string{{"tag1", "val1"}, {"tag2", "val2"}},
			Nodes: []input.NodeSettings{{
				FieldName:  "name3",
				Identifier: "3000",
				TagsSlice:  [][]string{{"tag3", "val3"}},
			}},
		},
		{
			MetricName:     "bar",
			Namespace:      "0",
			IdentifierType: "i",
			TagsSlice:      [][]string{{"tag1", "val1"}, {"tag2", "val2"}},
			Nodes: []input.NodeSettings{{
				FieldName:  "name4",
				Identifier: "4000",
				TagsSlice:  [][]string{{"tag1", "override"}},
			}},
		},
	}, o.SubscribeClientConfig.Groups)
	require.Equal(t, opcua.OpcUAWorkarounds{AdditionalValidStatusCodes: []string{"0xC0"}}, o.SubscribeClientConfig.Workarounds)
}
