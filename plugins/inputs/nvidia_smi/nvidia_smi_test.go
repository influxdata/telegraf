package nvidia_smi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseLineStandard(t *testing.T) {
	line := "41, 11264, 1074, 10190, P8, 32, GeForce RTX 2080 Ti, GPU-c97b7f88-c06d-650f-5339-f8dd0c1315c0, Default, 1, 4, 0, 24.33, 1, 16, 0, 0, 0, 300, 300, 405, 540\n"
	tags, fields, err := parseLine(line)
	if err != nil {
		t.Fail()
	}
	if tags["name"] != "GeForce RTX 2080 Ti" {
		t.Fail()
	}
	if temp, ok := fields["temperature_gpu"].(int); ok && temp != 32 {
		t.Fail()
	}
}

func TestParseLineEmptyLine(t *testing.T) {
	line := "\n"
	_, _, err := parseLine(line)
	if err == nil {
		t.Fail()
	}
}

func TestParseLineBad(t *testing.T) {
	line := "the quick brown fox jumped over the lazy dog"
	_, _, err := parseLine(line)
	if err == nil {
		t.Fail()
	}
}

func TestParseLineNotSupported(t *testing.T) {
	line := "[Not Supported], 11264, 1074, 10190, P8, 32, GeForce RTX 2080 Ti, GPU-c97b7f88-c06d-650f-5339-f8dd0c1315c0, Default, 1, 4, 0, 24.33, 1, 16, 0, 0, 0, 300, 300, 405, 540\n"
	_, fields, err := parseLine(line)
	require.NoError(t, err)
	require.Equal(t, nil, fields["fan_speed"])
}

func TestParseLineUnknownError(t *testing.T) {
	line := "[Unknown Error], 11264, 1074, 10190, P8, 32, GeForce RTX 2080 Ti, GPU-c97b7f88-c06d-650f-5339-f8dd0c1315c0, Default, 1, 4, 0, 24.33, 1, 16, 0, 0, 0, 300, 300, 405, 540\n"
	_, fields, err := parseLine(line)
	require.NoError(t, err)
	require.Equal(t, nil, fields["fan_speed"])
}
