package mqtt

import (
	"fmt"
	"path/filepath"
	"sync"
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
	influxSerializer "github.com/influxdata/telegraf/plugins/serializers/influx"
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
	s := &influxSerializer.Serializer{}
	require.NoError(t, s.Init())
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
	s := &influxSerializer.Serializer{}
	require.NoError(t, s.Init())

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

	url := fmt.Sprintf("%s:%s", container.Address, container.Ports[servicePort])
	s := &influxSerializer.Serializer{}
	require.NoError(t, s.Init())

	m := &MQTT{
		MqttConfig: mqtt.MqttConfig{
			Servers:   []string{url},
			Protocol:  "5",
			KeepAlive: 30,
			Timeout:   config.Duration(5 * time.Second),
		},
		serializer: s,
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
	serializer := &influxSerializer.Serializer{}
	require.NoError(t, serializer.Init())

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
		Topic:  topic + "/{{.PluginName}}",
		Layout: "non-batch",
		Log:    testutil.Logger{Name: "mqttv3-integration-test"},
	}
	plugin.SetSerializer(serializer)
	require.NoError(t, plugin.Init())

	// Prepare the receiver message
	var acc testutil.Accumulator
	onMessage := createMetricMessageHandler(&acc, parser)

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
			serializer := &influxSerializer.Serializer{}
			require.NoError(t, serializer.Init())

			plugin.SetSerializer(serializer)

			// Verify that we can connect to the MQTT broker
			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.Connect())

			// Verify that we can successfully write data to the mqtt broker
			require.NoError(t, plugin.Write(testutil.MockMetrics()))
		})
	}
}

func TestIntegrationMQTTLayoutNonBatch(t *testing.T) {
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
	serializer := &influxSerializer.Serializer{}
	require.NoError(t, serializer.Init())

	// Setup the plugin
	url := fmt.Sprintf("tcp://%s:%s", container.Address, container.Ports[servicePort])
	topic := "test_nonbatch"
	plugin := &MQTT{
		MqttConfig: mqtt.MqttConfig{
			Servers:       []string{url},
			KeepAlive:     30,
			Timeout:       config.Duration(5 * time.Second),
			AutoReconnect: true,
		},
		Topic:  topic + "/{{.PluginName}}",
		Layout: "non-batch",
		Log:    testutil.Logger{Name: "mqttv3-integration-test"},
	}
	plugin.SetSerializer(serializer)
	require.NoError(t, plugin.Init())

	// Prepare the receiver message
	var acc testutil.Accumulator
	onMessage := createMetricMessageHandler(&acc, parser)

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
		m := metric.New(
			name,
			map[string]string{"case": "mqtt"},
			map[string]interface{}{"value": i},
			time.Unix(1676470949, 0),
		)
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

func TestIntegrationMQTTLayoutBatch(t *testing.T) {
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
	serializer := &influxSerializer.Serializer{}
	require.NoError(t, serializer.Init())

	// Setup the plugin
	url := fmt.Sprintf("tcp://%s:%s", container.Address, container.Ports[servicePort])
	topic := "test_batch"
	plugin := &MQTT{
		MqttConfig: mqtt.MqttConfig{
			Servers:       []string{url},
			KeepAlive:     30,
			Timeout:       config.Duration(5 * time.Second),
			AutoReconnect: true,
		},
		Topic:  topic + "/{{.PluginName}}",
		Layout: "batch",
		Log:    testutil.Logger{Name: "mqttv3-integration-test-"},
	}
	plugin.SetSerializer(serializer)
	require.NoError(t, plugin.Init())

	// Prepare the receiver message
	var acc testutil.Accumulator
	onMessage := createMetricMessageHandler(&acc, parser)

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
	input := make([]telegraf.Metric, 0, 6)
	expected := make([]telegraf.Metric, 0, len(input))
	for i := 0; i < cap(input); i++ {
		name := fmt.Sprintf("test%d", i%3)
		m := metric.New(
			name,
			map[string]string{
				"case": "mqtt",
				"id":   fmt.Sprintf("test%d", i),
			},
			map[string]interface{}{"value": i},
			time.Unix(1676470949, 0),
		)
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
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.SortMetrics())
}

func TestIntegrationMQTTLayoutField(t *testing.T) {
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

	// Setup the plugin
	url := fmt.Sprintf("tcp://%s:%s", container.Address, container.Ports[servicePort])
	topic := "test_field"
	plugin := &MQTT{
		MqttConfig: mqtt.MqttConfig{
			Servers:       []string{url},
			KeepAlive:     30,
			Timeout:       config.Duration(5 * time.Second),
			AutoReconnect: true,
		},
		Topic:  topic + `/{{.PluginName}}/{{.Tag "source"}}`,
		Layout: "field",
		Log:    testutil.Logger{Name: "mqttv3-integration-test-"},
	}
	require.NoError(t, plugin.Init())

	// Startup the plugin and subscribe to the topic
	require.NoError(t, plugin.Connect())
	defer plugin.Close()

	// Prepare the message receiver
	var received []message
	var mtx sync.Mutex
	onMessage := func(_ paho.Client, msg paho.Message) {
		mtx.Lock()
		defer mtx.Unlock()
		received = append(received, message{msg.Topic(), msg.Payload()})
	}

	// Add routing for the messages
	subscriptionPattern := topic + "/#"
	plugin.client.AddRoute(subscriptionPattern, onMessage)

	// Subscribe to the topic
	topics := map[string]byte{subscriptionPattern: byte(plugin.QoS)}
	require.NoError(t, plugin.client.SubscribeMultiple(topics, onMessage))

	// Setup and execute the test case
	input := []telegraf.Metric{
		metric.New(
			"modbus",
			map[string]string{
				"source":   "device 1",
				"type":     "Machine A",
				"location": "main building",
				"status":   "ok",
			},
			map[string]interface{}{
				"temperature":   21.4,
				"serial number": "324nlk234r5u9834t",
				"working hours": 123,
				"supplied":      true,
			},
			time.Unix(1676522982, 0),
		),
		metric.New(
			"modbus",
			map[string]string{
				"source":   "device 2",
				"type":     "Machine B",
				"location": "main building",
				"status":   "offline",
			},
			map[string]interface{}{
				"temperature": 25.0,
				"supplied":    false,
			},
			time.Unix(1676522982, 0),
		),
	}

	expected := []string{
		topic + "/modbus/device 1/temperature" + " " + "21.4",
		topic + "/modbus/device 1/serial number" + " " + "324nlk234r5u9834t",
		topic + "/modbus/device 1/supplied" + " " + "true",
		topic + "/modbus/device 1/working hours" + " " + "123",
		topic + "/modbus/device 2/temperature" + " " + "25",
		topic + "/modbus/device 2/supplied" + " " + "false",
	}
	require.NoError(t, plugin.Write(input))

	// Verify the result
	require.Eventually(t, func() bool {
		mtx.Lock()
		defer mtx.Unlock()
		return len(received) >= len(expected)
	}, time.Second, 100*time.Millisecond)
	require.NoError(t, plugin.Close())

	actual := make([]string, 0, len(received))
	for _, msg := range received {
		actual = append(actual, msg.topic+" "+string(msg.payload))
	}
	require.ElementsMatch(t, expected, actual)
}

func TestIntegrationMQTTLayoutHomieV4(t *testing.T) {
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

	// Setup the plugin
	url := fmt.Sprintf("tcp://%s:%s", container.Address, container.Ports[servicePort])
	topic := "homie"
	plugin := &MQTT{
		MqttConfig: mqtt.MqttConfig{
			Servers:       []string{url},
			KeepAlive:     30,
			Timeout:       config.Duration(5 * time.Second),
			AutoReconnect: true,
		},
		Topic:           topic + "/{{.PluginName}}",
		HomieDeviceName: `{{.PluginName}}`,
		HomieNodeID:     `{{.Tag "source"}}`,
		Layout:          "homie-v4",
		Log:             testutil.Logger{Name: "mqttv3-integration-test-"},
	}
	require.NoError(t, plugin.Init())

	// Startup the plugin and subscribe to the topic
	require.NoError(t, plugin.Connect())
	defer plugin.Close()

	// Prepare the message receiver
	var received []message
	var mtx sync.Mutex
	onMessage := func(_ paho.Client, msg paho.Message) {
		mtx.Lock()
		defer mtx.Unlock()
		received = append(received, message{msg.Topic(), msg.Payload()})
	}

	// Add routing for the messages
	subscriptionPattern := topic + "/#"
	plugin.client.AddRoute(subscriptionPattern, onMessage)

	// Subscribe to the topic
	topics := map[string]byte{subscriptionPattern: byte(plugin.QoS)}
	require.NoError(t, plugin.client.SubscribeMultiple(topics, onMessage))

	// Setup and execute the test case
	input := []telegraf.Metric{
		metric.New(
			"modbus",
			map[string]string{
				"source":   "device 1",
				"type":     "Machine A",
				"location": "main building",
				"status":   "ok",
			},
			map[string]interface{}{
				"temperature":   21.4,
				"serial number": "324nlk234r5u9834t",
				"working hours": 123,
				"supplied":      true,
			},
			time.Unix(1676522982, 0),
		),
		metric.New(
			"modbus",
			map[string]string{
				"source":   "device 2",
				"type":     "Machine B",
				"location": "main building",
				"status":   "offline",
			},
			map[string]interface{}{
				"supplied": false,
			},
			time.Unix(1676522982, 0),
		),
		metric.New(
			"modbus",
			map[string]string{
				"source":       "device 2",
				"type":         "Machine B",
				"location":     "main building",
				"status":       "online",
				"in operation": "yes",
			},
			map[string]interface{}{
				"Temperature": 25.38,
				"Voltage":     24.1,
				"Current":     100.0,
				"Throughput":  12345,
				"Load [%]":    81.2,
				"account no":  "T3L3GrAf",
				"supplied":    true,
			},
			time.Unix(1676542982, 0),
		),
	}

	dev1Props := "location,serial-number,source,status,supplied,temperature,type,working-hours"
	dev2Props := "account-no,current,in-operation,load,location,source,status,supplied,temperature,"
	dev2Props += "throughput,type,voltage"
	expected := []string{
		topic + "/modbus/$homie" + " " + "4.0",
		topic + "/modbus/$name" + " " + "modbus",
		topic + "/modbus/$state" + " " + "ready",
		topic + "/modbus/$nodes" + " " + "device-1",
		topic + "/modbus/device-1/$name" + " " + "device 1",
		topic + "/modbus/device-1/$properties" + " " + dev1Props,
		topic + "/modbus/device-1/location" + " " + "main building",
		topic + "/modbus/device-1/location/$name" + " " + "location",
		topic + "/modbus/device-1/location/$datatype" + " " + "string",
		topic + "/modbus/device-1/status" + " " + "ok",
		topic + "/modbus/device-1/status/$name" + " " + "status",
		topic + "/modbus/device-1/status/$datatype" + " " + "string",
		topic + "/modbus/device-1/type" + " " + "Machine A",
		topic + "/modbus/device-1/type/$name" + " " + "type",
		topic + "/modbus/device-1/type/$datatype" + " " + "string",
		topic + "/modbus/device-1/source" + " " + "device 1",
		topic + "/modbus/device-1/source/$name" + " " + "source",
		topic + "/modbus/device-1/source/$datatype" + " " + "string",
		topic + "/modbus/device-1/temperature" + " " + "21.4",
		topic + "/modbus/device-1/temperature/$name" + " " + "temperature",
		topic + "/modbus/device-1/temperature/$datatype" + " " + "float",
		topic + "/modbus/device-1/serial-number" + " " + "324nlk234r5u9834t",
		topic + "/modbus/device-1/serial-number/$name" + " " + "serial number",
		topic + "/modbus/device-1/serial-number/$datatype" + " " + "string",
		topic + "/modbus/device-1/working-hours" + " " + "123",
		topic + "/modbus/device-1/working-hours/$name" + " " + "working hours",
		topic + "/modbus/device-1/working-hours/$datatype" + " " + "integer",
		topic + "/modbus/device-1/supplied" + " " + "true",
		topic + "/modbus/device-1/supplied/$name" + " " + "supplied",
		topic + "/modbus/device-1/supplied/$datatype" + " " + "boolean",
		topic + "/modbus/$nodes" + " " + "device-1,device-2",

		topic + "/modbus/device-2/$name" + " " + "device 2",
		topic + "/modbus/device-2/$properties" + " " + "location,source,status,supplied,type",
		topic + "/modbus/device-2/location" + " " + "main building",
		topic + "/modbus/device-2/location/$name" + " " + "location",
		topic + "/modbus/device-2/location/$datatype" + " " + "string",
		topic + "/modbus/device-2/status" + " " + "offline",
		topic + "/modbus/device-2/status/$name" + " " + "status",
		topic + "/modbus/device-2/status/$datatype" + " " + "string",
		topic + "/modbus/device-2/type" + " " + "Machine B",
		topic + "/modbus/device-2/type/$name" + " " + "type",
		topic + "/modbus/device-2/type/$datatype" + " " + "string",
		topic + "/modbus/device-2/source" + " " + "device 2",
		topic + "/modbus/device-2/source/$name" + " " + "source",
		topic + "/modbus/device-2/source/$datatype" + " " + "string",
		topic + "/modbus/device-2/supplied" + " " + "false",
		topic + "/modbus/device-2/supplied/$name" + " " + "supplied",
		topic + "/modbus/device-2/supplied/$datatype" + " " + "boolean",

		topic + "/modbus/device-2/$properties" + " " + dev2Props,
		topic + "/modbus/device-2/location" + " " + "main building",
		topic + "/modbus/device-2/location/$name" + " " + "location",
		topic + "/modbus/device-2/location/$datatype" + " " + "string",
		topic + "/modbus/device-2/in-operation" + " " + "yes",
		topic + "/modbus/device-2/in-operation/$name" + " " + "in operation",
		topic + "/modbus/device-2/in-operation/$datatype" + " " + "string",
		topic + "/modbus/device-2/status" + " " + "online",
		topic + "/modbus/device-2/status/$name" + " " + "status",
		topic + "/modbus/device-2/status/$datatype" + " " + "string",
		topic + "/modbus/device-2/type" + " " + "Machine B",
		topic + "/modbus/device-2/type/$name" + " " + "type",
		topic + "/modbus/device-2/type/$datatype" + " " + "string",
		topic + "/modbus/device-2/source" + " " + "device 2",
		topic + "/modbus/device-2/source/$name" + " " + "source",
		topic + "/modbus/device-2/source/$datatype" + " " + "string",
		topic + "/modbus/device-2/temperature" + " " + "25.38",
		topic + "/modbus/device-2/temperature/$name" + " " + "Temperature",
		topic + "/modbus/device-2/temperature/$datatype" + " " + "float",
		topic + "/modbus/device-2/voltage" + " " + "24.1",
		topic + "/modbus/device-2/voltage/$name" + " " + "Voltage",
		topic + "/modbus/device-2/voltage/$datatype" + " " + "float",
		topic + "/modbus/device-2/current" + " " + "100",
		topic + "/modbus/device-2/current/$name" + " " + "Current",
		topic + "/modbus/device-2/current/$datatype" + " " + "float",
		topic + "/modbus/device-2/throughput" + " " + "12345",
		topic + "/modbus/device-2/throughput/$name" + " " + "Throughput",
		topic + "/modbus/device-2/throughput/$datatype" + " " + "integer",
		topic + "/modbus/device-2/load" + " " + "81.2",
		topic + "/modbus/device-2/load/$name" + " " + "Load [%]",
		topic + "/modbus/device-2/load/$datatype" + " " + "float",
		topic + "/modbus/device-2/account-no" + " " + "T3L3GrAf",
		topic + "/modbus/device-2/account-no/$name" + " " + "account no",
		topic + "/modbus/device-2/account-no/$datatype" + " " + "string",
		topic + "/modbus/device-2/supplied" + " " + "true",
		topic + "/modbus/device-2/supplied/$name" + " " + "supplied",
		topic + "/modbus/device-2/supplied/$datatype" + " " + "boolean",

		topic + "/modbus/$state" + " " + "lost",
	}
	require.NoError(t, plugin.Write(input))
	require.NoError(t, plugin.Close())

	// Verify the result
	require.Eventually(t, func() bool {
		mtx.Lock()
		defer mtx.Unlock()
		return len(received) >= len(expected)
	}, time.Second, 100*time.Millisecond)

	actual := make([]string, 0, len(received))
	for _, msg := range received {
		actual = append(actual, msg.topic+" "+string(msg.payload))
	}
	require.ElementsMatch(t, expected, actual)
}

func createMetricMessageHandler(acc telegraf.Accumulator, parser telegraf.Parser) paho.MessageHandler {
	return func(_ paho.Client, msg paho.Message) {
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
	s := &influxSerializer.Serializer{}
	require.NoError(t, s.Init())

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
