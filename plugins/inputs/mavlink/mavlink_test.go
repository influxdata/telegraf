package mavlink

import (
	"testing"

	"github.com/chrisdalke/gomavlib/v3"
	"github.com/stretchr/testify/require"
)

// Test that a serial port URL can be parsed (Linux)
func TestParseSerialFcuUrlLinux(t *testing.T) {
	testConfig := Mavlink{
		FcuURL: "serial:///dev/ttyACM0:115200",
	}

	config, err := ParseMavlinkEndpointConfig(testConfig.FcuURL)
	require.NoError(t, err)
	endpoint, ok := config[0].(gomavlib.EndpointSerial)
	require.True(t, ok)
	require.Equal(t, "/dev/ttyACM0", endpoint.Device)
	require.Equal(t, 115200, endpoint.Baud)
}

// Test that a serial port URL can be parsed (Windows)
func TestParseSerialFcuUrlWindows(t *testing.T) {
	testConfig := Mavlink{
		FcuURL: "serial://COM1:115200",
	}

	config, err := ParseMavlinkEndpointConfig(testConfig.FcuURL)
	require.NoError(t, err)
	endpoint, ok := config[0].(gomavlib.EndpointSerial)
	require.True(t, ok)
	require.Equal(t, "COM1", endpoint.Device)
	require.Equal(t, 115200, endpoint.Baud)
}

// Test that a UDP client URL can be parsed.
func TestParseUDPClientFcuUrl(t *testing.T) {
	testConfig := Mavlink{
		FcuURL: "udp://192.168.1.12:14550",
	}

	config, err := ParseMavlinkEndpointConfig(testConfig.FcuURL)
	require.NoError(t, err)
	endpoint, ok := config[0].(gomavlib.EndpointUDPClient)
	require.True(t, ok)
	require.Equal(t, "192.168.1.12:14550", endpoint.Address)
}

// Test that a UDP server URL can be parsed.
func TestParseUDPServerFcuUrl(t *testing.T) {
	testConfig := Mavlink{
		FcuURL: "udp://:14540",
	}

	config, err := ParseMavlinkEndpointConfig(testConfig.FcuURL)
	require.NoError(t, err)
	endpoint, ok := config[0].(gomavlib.EndpointUDPServer)
	require.True(t, ok)
	require.Equal(t, ":14540", endpoint.Address)
}

// Test that a TCP client URL can be parsed.
func TestParseTCPClientFcuUrl(t *testing.T) {
	testConfig := Mavlink{
		FcuURL: "tcp://192.168.1.12:14550",
	}

	config, err := ParseMavlinkEndpointConfig(testConfig.FcuURL)
	require.NoError(t, err)
	endpoint, ok := config[0].(gomavlib.EndpointTCPClient)
	require.True(t, ok)
	require.Equal(t, "192.168.1.12:14550", endpoint.Address)
}

// Test that an invalid URL is caught.
func TestParseInvalidFcuUrl(t *testing.T) {
	testConfig := Mavlink{
		FcuURL: "ftp://not-a-valid-fcu-url",
	}

	_, err := ParseMavlinkEndpointConfig(testConfig.FcuURL)
	require.Equal(t, "could not parse fcu_url ftp://not-a-valid-fcu-url", err.Error())
}

func TestConvertToSnakeCase(t *testing.T) {
	require.Equal(t, "", ConvertToSnakeCase(""))
	require.Equal(t, "camel_case", ConvertToSnakeCase("CamelCase"))
	require.Equal(t, "camel_camel_case", ConvertToSnakeCase("CamelCamelCase"))
	require.Equal(t, "snake_case", ConvertToSnakeCase("snake_case"))
	require.Equal(t, "snake_case", ConvertToSnakeCase("SNAKE_CASE"))
}
