package nagios

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validOutput1 = `PING OK - Packet loss = 0%, RTA = 0.30 ms|rta=0.298000ms;4000.000000;6000.000000;0.000000 pl=0%;80;90;0;100
This is a long output
with three lines
`
const validOutput2 = "TCP OK - 0.008 second response time on port 80|time=0.008457s;;;0.000000;10.000000"
const validOutput3 = "TCP OK - 0.008 second response time on port 80|time=0.008457"
const invalidOutput3 = "PING OK - Packet loss = 0%, RTA = 0.30 ms"
const invalidOutput4 = "PING OK - Packet loss = 0%, RTA = 0.30 ms| =3;;;; dgasdg =;;;; sff=;;;;"

func TestParseValidOutput(t *testing.T) {
	parser := NagiosParser{
		MetricName: "nagios_test",
	}

	// Output1
	metrics, err := parser.Parse([]byte(validOutput1))
	require.NoError(t, err)
	assert.Len(t, metrics, 2)
	// rta
	assert.Equal(t, "rta", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value":    float64(0.298),
		"warning":  float64(4000),
		"critical": float64(6000),
		"min":      float64(0),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{"unit": "ms"}, metrics[0].Tags())
	// pl
	assert.Equal(t, "pl", metrics[1].Name())
	assert.Equal(t, map[string]interface{}{
		"value":    float64(0),
		"warning":  float64(80),
		"critical": float64(90),
		"min":      float64(0),
		"max":      float64(100),
	}, metrics[1].Fields())
	assert.Equal(t, map[string]string{"unit": "%"}, metrics[1].Tags())

	// Output2
	metrics, err = parser.Parse([]byte(validOutput2))
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	// time
	assert.Equal(t, "time", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": float64(0.008457),
		"min":   float64(0),
		"max":   float64(10),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{"unit": "s"}, metrics[0].Tags())

	// Output3
	metrics, err = parser.Parse([]byte(validOutput3))
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	// time
	assert.Equal(t, "time", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": float64(0.008457),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())
}

func TestParseInvalidOutput(t *testing.T) {
	parser := NagiosParser{
		MetricName: "nagios_test",
	}

	// invalidOutput3
	metrics, err := parser.Parse([]byte(invalidOutput3))
	require.NoError(t, err)
	assert.Len(t, metrics, 0)

	// invalidOutput4
	metrics, err = parser.Parse([]byte(invalidOutput4))
	require.NoError(t, err)
	assert.Len(t, metrics, 0)

}
