package mqtt

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/common/mqtt"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
)

const servicePort = "1883"

func launchTestContainer(t *testing.T) *testutil.Container {
	conf, err := filepath.Abs(filepath.Join("testdata", "mosquitto.conf"))
	require.NoError(t, err, "missing file mosquitto.conf")

	container := testutil.Container{
		Image:        "eclipse-mosquitto:2",
		ExposedPorts: []string{servicePort},
		WaitingFor:   wait.ForListeningPort(servicePort),
		BindMounts: map[string]string{
			"/mosquitto/config/mosquitto.conf": conf,
		},
	}
	err = container.Start()
	require.NoError(t, err, "failed to start container")

	return &container
}

func TestConnectAndWriteIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t)
	defer container.Terminate()
	var url = fmt.Sprintf("%s:%s", container.Address, container.Ports[servicePort])
	s := serializers.NewInfluxSerializer()
	m := &MQTT{
		MqttConfig: mqtt.MqttConfig{
			Servers:   []string{url},
			KeepAlive: 30,
			Timeout:   config.Duration(5 * time.Second),
		},
		serializer: s,
		Log:        testutil.Logger{Name: "mqtt-default-integration-test"},
	}

	// Verify that we can connect to the MQTT broker
	require.NoError(t, m.Init())

	// Verify that we can connect to the MQTT broker
	require.NoError(t, m.Connect())

	// Verify that we can successfully write data to the mqtt broker
	require.NoError(t, m.Write(testutil.MockMetrics()))
}

func TestConnectAndWriteIntegrationMQTTv3(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t)
	defer container.Terminate()

	var url = fmt.Sprintf("%s:%s", container.Address, container.Ports[servicePort])
	s := serializers.NewInfluxSerializer()
	m := &MQTT{
		MqttConfig: mqtt.MqttConfig{
			Servers:   []string{url},
			Protocol:  "3.1.1",
			KeepAlive: 30,
			Timeout:   config.Duration(5 * time.Second),
		},
		serializer: s,
		Log:        testutil.Logger{Name: "mqttv311-integration-test"},
	}

	// Verify that we can connect to the MQTT broker
	require.NoError(t, m.Init())

	// Verify that we can connect to the MQTT broker
	require.NoError(t, m.Connect())

	// Verify that we can successfully write data to the mqtt broker
	require.NoError(t, m.Write(testutil.MockMetrics()))
}

func TestConnectAndWriteIntegrationMQTTv5(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t)
	defer container.Terminate()

	var url = fmt.Sprintf("%s:%s", container.Address, container.Ports[servicePort])
	m := &MQTT{
		MqttConfig: mqtt.MqttConfig{
			Servers:   []string{url},
			Protocol:  "5",
			KeepAlive: 30,
			Timeout:   config.Duration(5 * time.Second),
		},
		serializer: serializers.NewInfluxSerializer(),
		Log:        testutil.Logger{Name: "mqttv5-integration-test"},
	}

	// Verify that we can connect to the MQTT broker
	require.NoError(t, m.Init())
	require.NoError(t, m.Connect())

	// Verify that we can successfully write data to the mqtt broker
	require.NoError(t, m.Write(testutil.MockMetrics()))
}

func TestIntegrationMQTTv3(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	conf, err := filepath.Abs(filepath.Join("testdata", "mosquitto.conf"))
	require.NoError(t, err, "missing file mosquitto.conf")

	container := testutil.Container{
		Image:        "eclipse-mosquitto:2",
		ExposedPorts: []string{servicePort},
		WaitingFor:   wait.ForListeningPort(servicePort),
		BindMounts: map[string]string{
			"/mosquitto/config/mosquitto.conf": conf,
		},
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	// Setup the parser / serializer pair
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	serializer := serializers.NewInfluxSerializer()

	// Setup the plugin
	url := fmt.Sprintf("tcp://%s:%s", container.Address, container.Ports[servicePort])
	topic := "testv3"
	plugin := &MQTT{
		MqttConfig: mqtt.MqttConfig{
			Servers:       []string{url},
			KeepAlive:     30,
			Timeout:       config.Duration(5 * time.Second),
			AutoReconnect: true,
		},
		Topic: topic + "/{{.PluginName}}",
		Log:   testutil.Logger{Name: "mqttv3-integration-test"},
	}
	plugin.SetSerializer(serializer)
	require.NoError(t, plugin.Init())

	// Prepare the receiver message
	var acc testutil.Accumulator
	onMessage := func(_ paho.Client, msg paho.Message) {
		metrics, err := parser.Parse(msg.Payload())
		if err != nil {
			acc.AddError(err)
			return
		}

		for _, m := range metrics {
			m.AddTag("topic", msg.Topic())
			acc.AddMetric(m)
		}
	}

	// Startup the plugin and subscribe to the topic
	require.NoError(t, plugin.Connect())
	defer plugin.Close()

	// Add routing for the messages
	subscriptionPattern := topic + "/#"
	plugin.client.AddRoute(subscriptionPattern, onMessage)

	// Subscribe to the topic
	topics := map[string]byte{subscriptionPattern: byte(plugin.QoS)}
	require.NoError(t, plugin.client.SubscribeMultiple(topics, onMessage))

	// Setup and execute the test case
	input := make([]telegraf.Metric, 0, 3)
	expected := make([]telegraf.Metric, 0, len(input))
	for i := 0; i < cap(input); i++ {
		name := fmt.Sprintf("test%d", i)
		m := testutil.TestMetric(i, name)
		input = append(input, m)

		e := m.Copy()
		e.AddTag("topic", topic+"/"+name)
		expected = append(expected, e)
	}
	require.NoError(t, plugin.Write(input))

	// Verify the result
	require.Eventually(t, func() bool {
		return acc.NMetrics() >= uint64(len(expected))
	}, time.Second, 100*time.Millisecond)
	require.NoError(t, plugin.Close())

	require.Empty(t, acc.Errors)
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func TestMQTTv5Properties(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t)
	defer container.Terminate()

	tests := []struct {
		name       string
		properties *mqtt.PublishProperties
	}{
		{
			name:       "no publish properties",
			properties: nil,
		},
		{
			name:       "content type set",
			properties: &mqtt.PublishProperties{ContentType: "text/plain"},
		},
		{
			name:       "response topic set",
			properties: &mqtt.PublishProperties{ResponseTopic: "test/topic"},
		},
		{
			name:       "message expiry set",
			properties: &mqtt.PublishProperties{MessageExpiry: config.Duration(10 * time.Minute)},
		},
		{
			name:       "topic alias set",
			properties: &mqtt.PublishProperties{TopicAlias: new(uint16)},
		},
		{
			name:       "user properties set",
			properties: &mqtt.PublishProperties{UserProperties: map[string]string{"key": "value"}},
		},
	}

	topic := "testv3"
	url := fmt.Sprintf("%s:%s", container.Address, container.Ports[servicePort])
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &MQTT{
				MqttConfig: mqtt.MqttConfig{
					Servers:       []string{url},
					Protocol:      "5",
					KeepAlive:     30,
					Timeout:       config.Duration(5 * time.Second),
					AutoReconnect: true,
				},
				Topic: topic,
				Log:   testutil.Logger{Name: "mqttv5-integration-test"},
			}

			// Setup the metric serializer
			serializer := serializers.NewInfluxSerializer()
			plugin.SetSerializer(serializer)

			// Verify that we can connect to the MQTT broker
			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.Connect())

			// Verify that we can successfully write data to the mqtt broker
			require.NoError(t, plugin.Write(testutil.MockMetrics()))
		})
	}
}

func TestMissingServers(t *testing.T) {
	plugin := &MQTT{}
	require.ErrorContains(t, plugin.Init(), "no servers specified")
}

func TestMQTTTopicGenerationTemplateIsValid(t *testing.T) {
	tests := []struct {
		name          string
		topic         string
		expectedError string
	}{
		{
			name:          "a valid pattern is accepted",
			topic:         "this/is/valid",
			expectedError: "",
		},
		{
			name:          "an invalid pattern is rejected",
			topic:         "this/is/#/invalid",
			expectedError: "found forbidden character # in the topic name this/is/#/invalid",
		},
		{
			name:          "an invalid pattern is rejected",
			topic:         "this/is/+/invalid",
			expectedError: "found forbidden character + in the topic name this/is/+/invalid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MQTT{
				Topic: tt.topic,
				MqttConfig: mqtt.MqttConfig{
					Servers: []string{"tcp://localhost:1883"},
				},
			}
			err := m.Init()
			if tt.expectedError != "" {
				require.ErrorContains(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGenerateTopicName(t *testing.T) {
	s := serializers.NewInfluxSerializer()
	m := &MQTT{
		MqttConfig: mqtt.MqttConfig{
			Servers:   []string{"tcp://localhost:1883"},
			KeepAlive: 30,
			Timeout:   config.Duration(5 * time.Second),
		},
		serializer: s,
		Log:        testutil.Logger{},
	}
	tests := []struct {
		name    string
		pattern string
		want    string
	}{
		{
			name:    "matches default legacy format",
			pattern: "telegraf/{{ .Hostname }}/{{ .PluginName }}",
			want:    "telegraf/hostname/metric-name",
		},
		{
			name:    "respect hardcoded strings",
			pattern: "this/is/a/topic",
			want:    "this/is/a/topic",
		},
		{
			name:    "allows the use of tags",
			pattern: "{{ .TopicPrefix }}/{{ .Tag \"tag1\" }}",
			want:    "prefix/value1",
		},
		{
			name:    "uses the plugin name when no pattern is provided",
			pattern: "",
			want:    "metric-name",
		},
		{
			name:    "ignores tag when tag does not exists",
			pattern: "{{ .TopicPrefix }}/{{ .Tag \"not-a-tag\" }}",
			want:    "prefix",
		},
		{
			name:    "ignores empty forward slashes",
			pattern: "double//slashes//are//ignored",
			want:    "double/slashes/are/ignored",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.Topic = tt.pattern
			m.TopicPrefix = "prefix"
			met := metric.New(
				"metric-name",
				map[string]string{"tag1": "value1"},
				map[string]interface{}{"value": 123},
				time.Date(2022, time.November, 10, 23, 0, 0, 0, time.UTC),
			)
			require.NoError(t, m.Init())
			actual, err := m.generator.Generate("hostname", met)
			require.NoError(t, err)
			require.Equal(t, tt.want, actual)
		})
	}
}
