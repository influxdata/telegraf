package mavlink

import (
	"testing"
	"time"

	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/common"
	"github.com/bluenviron/gomavlib/v3/pkg/frame"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

// Test that all URL types can be parsed successfully
func TestParseURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected gomavlib.EndpointConf
	}{
		{
			name: "serial Linux",
			url:  "serial:///dev/ttyACM0:115200",
			expected: gomavlib.EndpointSerial{
				Device: "/dev/ttyACM0",
				Baud:   115200,
			},
		},
		{
			name: "serial Linux with default baudrate",
			url:  "serial:///dev/ttyACM0",
			expected: gomavlib.EndpointSerial{
				Device: "/dev/ttyACM0",
				Baud:   57600,
			},
		},
		{
			name: "serial Windows",
			url:  "serial://COM1:115200",
			expected: gomavlib.EndpointSerial{
				Device: "COM1",
				Baud:   115200,
			},
		},
		{
			name: "serial Windows with default baudrate",
			url:  "serial://COM1",
			expected: gomavlib.EndpointSerial{
				Device: "COM1",
				Baud:   57600,
			},
		},
		{
			name: "UDP client",
			url:  "udp://192.168.1.12:14550",
			expected: gomavlib.EndpointUDPClient{
				Address: "192.168.1.12:14550",
			},
		},
		{
			name: "UDP client with default port",
			url:  "udp://192.168.1.12",
			expected: gomavlib.EndpointUDPClient{
				Address: "192.168.1.12:14550",
			},
		},
		{
			name: "UDP server",
			url:  "udpserver://:14550",
			expected: gomavlib.EndpointUDPServer{
				Address: "0.0.0.0:14550",
			},
		},
		{
			name: "UDP server with default port",
			url:  "udpserver://",
			expected: gomavlib.EndpointUDPServer{
				Address: "0.0.0.0:14550",
			},
		},
		{
			name: "UDP server on localhost",
			url:  "udpserver://127.0.0.1",
			expected: gomavlib.EndpointUDPServer{
				Address: "127.0.0.1:14550",
			},
		},
		{
			name: "TCP client",
			url:  "tcp://192.168.1.12:5760",
			expected: gomavlib.EndpointTCPClient{
				Address: "192.168.1.12:5760",
			},
		},
		{
			name: "TCP client with default port",
			url:  "tcp://192.168.1.12",
			expected: gomavlib.EndpointTCPClient{
				Address: "192.168.1.12:5760",
			},
		},
		{
			name: "TCP server",
			url:  "tcpserver://:5761",
			expected: gomavlib.EndpointTCPServer{
				Address: "0.0.0.0:5761",
			},
		},
		{
			name: "TCP server with default port",
			url:  "tcpserver://",
			expected: gomavlib.EndpointTCPServer{
				Address: "0.0.0.0:5760",
			},
		},
		{
			name: "TCP server on localhost",
			url:  "tcpserver://127.0.0.1",
			expected: gomavlib.EndpointTCPServer{
				Address: "127.0.0.1:5760",
			},
		},
		{
			name: "Default connection",
			url:  "",
			expected: gomavlib.EndpointTCPClient{
				Address: "127.0.0.1:5760",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the plugin
			plugin := &Mavlink{URL: tt.url}
			require.NoError(t, plugin.Init())

			// Check the resulting endpoint configuration
			require.EqualValues(t, tt.expected, plugin.endpointConfig)
		})
	}
}

// Test that some mavlink messages are correctly decoded into telegraf metrics.
func TestMavlinkDecoding(t *testing.T) {
	tests := []struct {
		name     string
		input    *frame.V2Frame
		expected []telegraf.Metric
	}{
		{
			name: "Heartbeat",
			input: &frame.V2Frame{
				SequenceNumber: 1,
				SystemID:       1,
				ComponentID:    1,
				Message: &common.MessageHeartbeat{
					Type:           0,
					Autopilot:      1,
					BaseMode:       2,
					CustomMode:     3,
					SystemStatus:   4,
					MavlinkVersion: 5,
				},
				Checksum: 0,
			},
			expected: []telegraf.Metric{metric.New(
				"heartbeat",
				map[string]string{
					"sys_id": "1",
					"source": "udpserver://",
				},
				map[string]interface{}{
					"custom_mode":     uint32(3),
					"mavlink_version": uint8(5),
				},
				time.Unix(0, 0),
			)},
		},
		{
			name: "Attitude",
			input: &frame.V2Frame{
				SequenceNumber: 1,
				SystemID:       1,
				ComponentID:    1,
				Message: &common.MessageAttitude{
					TimeBootMs: uint32(123),
					Roll:       float32(1.234),
					Pitch:      float32(0.463),
					Yaw:        float32(-0.112),
					Rollspeed:  float32(0.001),
					Pitchspeed: float32(0.002),
					Yawspeed:   float32(0.003),
				},
				Checksum: 0,
			},
			expected: []telegraf.Metric{metric.New(
				"attitude",
				map[string]string{
					"sys_id": "1",
					"source": "udpserver://",
				},
				map[string]interface{}{
					"pitch":        float32(0.463),
					"roll":         float32(1.234),
					"yaw":          float32(-0.112),
					"pitchspeed":   float32(0.002),
					"rollspeed":    float32(0.001),
					"yawspeed":     float32(0.003),
					"time_boot_ms": uint32(123),
				},
				time.Unix(0, 0),
			)},
		},
		{
			name: "ESC Status",
			input: &frame.V2Frame{
				SequenceNumber: 1,
				SystemID:       1,
				ComponentID:    1,
				Message: &common.MessageEscStatus{
					Index:    uint8(0),
					TimeUsec: uint64(12345),
					Rpm:      [4]int32{0, 1, 2, 3},
					Voltage:  [4]float32{10.0, 11.0, 12.0, 13.0},
					Current:  [4]float32{14.0, 15.0, 16.0, 17.0},
				},
				Checksum: 0,
			},
			expected: []telegraf.Metric{metric.New(
				"esc_status",
				map[string]string{
					"sys_id": "1",
					"source": "udpserver://",
				},
				map[string]interface{}{
					"index":     uint8(0),
					"time_usec": uint64(12345),
					"current_1": float32(14.0),
					"current_2": float32(15.0),
					"current_3": float32(16.0),
					"current_4": float32(17.0),
					"rpm_1":     int32(0),
					"rpm_2":     int32(1),
					"rpm_3":     int32(2),
					"rpm_4":     int32(3),
					"voltage_1": float32(10.0),
					"voltage_2": float32(11.0),
					"voltage_3": float32(12.0),
					"voltage_4": float32(13.0),
				},
				time.Unix(0, 0),
			)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the plugin
			plugin := &Mavlink{URL: "udpserver://"}
			require.NoError(t, plugin.Init())
			acc := testutil.Accumulator{}

			plugin.handleFrame(&acc, tt.input)

			// Check that accumulator contains the metric
			testutil.RequireMetricsStructureEqual(t, tt.expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

// Test that the plugin can interface with a real ArduPilot container.
func TestArduPilotIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Start the docker container for ArduPilot
	container := testutil.Container{
		Image:        "radarku/ardupilot-sitl",
		ExposedPorts: []string{"5760"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(nat.Port("5760")),
		),
	}
	require.NoError(t, container.Start(), "failed to start ardupilot container")
	defer container.Terminate()

	// Setup the plugin
	plugin := Mavlink{
		URL:                    "tcp://127.0.0.1:" + container.Ports["5760"],
		SystemID:               254,
		StreamRequestFrequency: 4,
		Log:                    testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	// Collect the metrics and compare
	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Expect to have received more than 10 metrics. The exact metrics received
	// is non-deterministic because ArduPilot's startup may vary between runs,
	// but should be on the order of 100 metrics in 5 seconds.
	// Content of metrics is not tested here as we don't know what to expect.
	require.Eventually(t, func() bool {
		return acc.NMetrics() >= 10
	}, 10*time.Second, 100*time.Millisecond, "less than 10 metrics received")
}
