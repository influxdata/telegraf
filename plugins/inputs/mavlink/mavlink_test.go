package mavlink

import (
	"testing"
)

// Test that a serial port URL can be parsed.
func TestParseSerialFcuUrl(t *testing.T) {
	testConfig = Mavlink{
		FcuUrl: "serial://dev/ttyACM0:115200",
	}

	conf, error := ParseMavlinkEndpointConfig(&testConfig)
	require.Equal(t, error, nil)
}

// Test that a UDP client URL can be parsed.
func TestParseUDPClientFcuUrl(t *testing.T) {
	testConfig = Mavlink{
		FcuUrl: "udp://192.168.1.12:14550",
	}

	conf, error := ParseMavlinkEndpointConfig(&testConfig)
	require.Equal(t, error, nil)
}

// Test that a UDP server URL can be parsed.
func TestParseUDPServerFcuUrl(t *testing.T) {
	testConfig = Mavlink{
		FcuUrl: "udp://:14540",
	}

	conf, error := ParseMavlinkEndpointConfig(&testConfig)
	require.Equal(t, error, nil)
}

// Test that a TCP client URL can be parsed.
func TestParseTCPClientFcuUrl(t *testing.T) {
	testConfig = Mavlink{
		FcuUrl: "tcp://192.168.1.12:14550",
	}

	conf, error := ParseMavlinkEndpointConfig(&testConfig)
	require.Equal(t, error, nil)
}

// Test that an invalid URL is caught.
func TestParseInvalidFcuUrl(t *testing.T) {
	testConfig = Mavlink{
		FcuUrl: "ftp://not-a-valid-fcu-url",
	}

	conf, error := ParseMavlinkEndpointConfig(&testConfig)
	require.Equal(t, error, "")
}

func TestStringContains(t *testing.T) {
	testArr := []string{"test1", "test2", "test3"}
	require.Equal(t, true, Contains(testArr, "test1"))
	require.Equal(t, true, Contains(testArr, "test2"))
	require.Equal(t, true, Contains(testArr, "test3"))
	require.Equal(t, false, Contains(testArr, "test4"))
}

func TestConvertToSnakeCase(t *testing.T) {
	require.Equal(t, "", ConvertToSnakeCase(""))
	require.Equal(t, "CamelCase", ConvertToSnakeCase("camel_case"))
	require.Equal(t, "CamelCamelCase", ConvertToSnakeCase("camel_camel_case"))
	require.Equal(t, "snake_case", ConvertToSnakeCase("snake_case"))
}
