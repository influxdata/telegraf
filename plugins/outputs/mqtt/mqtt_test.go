package mqtt

import (
	"fmt"
	"github.com/influxdata/telegraf/metric"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

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
		Servers:    []string{url},
		serializer: s,
		KeepAlive:  30,
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
		Servers:    []string{url},
		Protocol:   "3.1.1",
		serializer: s,
		KeepAlive:  30,
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
	s := serializers.NewInfluxSerializer()
	m := &MQTT{
		Servers:    []string{url},
		Protocol:   "5",
		serializer: s,
		KeepAlive:  30,
		Log:        testutil.Logger{Name: "mqttv5-integration-test"},
	}

	// Verify that we can connect to the MQTT broker
	require.NoError(t, m.Init())
	require.NoError(t, m.Connect())

	// Verify that we can successfully write data to the mqtt broker
	require.NoError(t, m.Write(testutil.MockMetrics()))
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
		Servers:    []string{"tcp://localhost:502"},
		serializer: s,
		KeepAlive:  30,
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
