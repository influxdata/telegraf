package nvidia_smi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseLineStandard(t *testing.T) {
	line := "85, 8114, 553, 7561, P2, 61, GeForce GTX 1070 Ti, GPU-d1911b8a-f5c8-5e66-057c-486561269de8, Default, 100, 93, 1, 0.0\n"
	tags, fields, err := parseLine(line)
	if err != nil {
		t.Fail()
	}
	if tags["name"] != "GeForce GTX 1070 Ti" {
		t.Fail()
	}
	if temp, ok := fields["temperature_gpu"].(int); ok && temp == 61 {
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
	line := "[Not Supported], 7606, 0, 7606, P0, 38, Tesla P4, GPU-xxx, Default, 0, 0, 0, 0.0\n"
	_, fields, err := parseLine(line)
	require.NoError(t, err)
	require.Equal(t, nil, fields["fan_speed"])
}
