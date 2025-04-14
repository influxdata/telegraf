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
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

// Test that a serial port URL can be parsed (Linux)
func TestParseSerialUrlLinux(t *testing.T) {
	plugin := Mavlink{
		URL: "serial:///dev/ttyACM0:115200",
	}

	err := plugin.Init()
	require.NoError(t, err)

	endpoint, ok := plugin.endpointConfig[0].(gomavlib.EndpointSerial)
	require.True(t, ok)
	require.Equal(t, "/dev/ttyACM0", endpoint.Device)
	require.Equal(t, 115200, endpoint.Baud)
}

// Test that a serial port URL can be parsed (Windows)
func TestParseSerialUrlWindows(t *testing.T) {
	plugin := Mavlink{
		URL: "serial://COM1:115200",
	}

	err := plugin.Init()
	require.NoError(t, err)

	endpoint, ok := plugin.endpointConfig[0].(gomavlib.EndpointSerial)
	require.True(t, ok)
	require.Equal(t, "COM1", endpoint.Device)
	require.Equal(t, 115200, endpoint.Baud)
}

// Test that a UDP client URL can be parsed.
func TestParseUDPClientUrl(t *testing.T) {
	plugin := Mavlink{
		URL: "udp://192.168.1.12:14550",
	}

	err := plugin.Init()
	require.NoError(t, err)

	endpoint, ok := plugin.endpointConfig[0].(gomavlib.EndpointUDPClient)
	require.True(t, ok)
	require.Equal(t, "192.168.1.12:14550", endpoint.Address)
}

// Test that a UDP server URL can be parsed.
func TestParseUDPServerUrl(t *testing.T) {
	plugin := Mavlink{
		URL: "udp://:14540",
	}

	err := plugin.Init()
	require.NoError(t, err)

	endpoint, ok := plugin.endpointConfig[0].(gomavlib.EndpointUDPServer)
	require.True(t, ok)
	require.Equal(t, ":14540", endpoint.Address)
}

// Test that a TCP client URL can be parsed.
func TestParseTCPClientUrl(t *testing.T) {
	plugin := Mavlink{
		URL: "tcp://192.168.1.12:14550",
	}

	err := plugin.Init()
	require.NoError(t, err)

	endpoint, ok := plugin.endpointConfig[0].(gomavlib.EndpointTCPClient)
	require.True(t, ok)
	require.Equal(t, "192.168.1.12:14550", endpoint.Address)
}

// Test that an invalid URL is caught.
func TestParseInvalidUrl(t *testing.T) {
	plugin := Mavlink{
		URL: "ftp://not-a-valid-fcu-url",
	}

	err := plugin.Init()
	require.ErrorContains(t, err, "unknown scheme \"ftp\"")
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
				},
				map[string]interface{}{
					"custom_mode":     uint64(3),
					"mavlink_version": uint64(5),
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
					TimeBootMs: 123,
					Roll:       1.234,
					Pitch:      0.463,
					Yaw:        -0.112,
					Rollspeed:  0.001,
					Pitchspeed: 0.002,
					Yawspeed:   0.003,
				},
				Checksum: 0,
			},
			expected: []telegraf.Metric{metric.New(
				"attitude",
				map[string]string{
					"sys_id": "1",
				},
				map[string]interface{}{
					"pitch":        float64(0.463),
					"roll":         float64(1.234),
					"yaw":          float64(-0.112),
					"pitchspeed":   float64(0.002),
					"rollspeed":    float64(0.001),
					"yawspeed":     float64(0.003),
					"time_boot_ms": uint64(123),
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
					Index:    0,
					TimeUsec: 12345,
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
				},
				map[string]interface{}{
					"index":     uint64(0),
					"time_usec": uint64(12345),
					"current_1": float64(14.0),
					"current_2": float64(15.0),
					"current_3": float64(16.0),
					"current_4": float64(17.0),
					"rpm_1":     int32(0),
					"rpm_2":     int32(1),
					"rpm_3":     int32(2),
					"rpm_4":     int32(3),
					"voltage_1": float64(10.0),
					"voltage_2": float64(11.0),
					"voltage_3": float64(12.0),
					"voltage_4": float64(13.0),
				},
				time.Unix(0, 0),
			)},
		},
	}

	tmpFilter, err := filter.Compile(make([]string, 0))
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := convertFrameToMetric(tt.input, tmpFilter)
			testutil.RequireMetricsStructureEqual(t, tt.expected, []telegraf.Metric{actual}, testutil.IgnoreTime())
		})
	}
}

// Test that the plugin can interface with a real ArduPilot container.
func TestArduPilotIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Start the docker container for ArduPilot
	t.Log("Starting ArduPilot container")
	container := testutil.Container{
		Image:        "radarku/ardupilot-sitl",
		ExposedPorts: []string{"5760"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(nat.Port("5760")),
		),
	}
	require.NoError(t, container.Start(), "failed to start ardupilot container")
	defer container.Terminate()

	t.Logf("ArduPilot is listening on port: %s", container.Ports["5760"])

	// Setup the plugin
	plugin := Mavlink{
		URL:      "tcp://127.0.0.1:" + container.Ports["5760"],
		SystemID: 254,
	}
	plugin.Log = testutil.Logger{}
	require.NoError(t, plugin.Init())

	// Collect the metrics and compare
	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))

	// Wait 5 seconds, then inspect metrics
	time.Sleep(time.Second * 5)
	require.NoError(t, plugin.Gather(&acc))
	actual := acc.GetTelegrafMetrics()
	plugin.Stop()

	// Expect to have received more than 10 metrics
	// The exact metrics received is non-deterministic because ArduPilot's
	// startup may differ between runs, but should be on the order of 100
	// metrics in 5 seconds. Not actually testing the content of metrics
	// here; that is tested in TestMavlinkDecoding.
	t.Logf("Received %d metrics", len(actual))
	require.Greater(t, len(actual), 10)
}
