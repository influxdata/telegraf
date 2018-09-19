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
const validOutput4 = "OK: Load average: 0.00, 0.01, 0.05 | 'load1'=0.00;~:4;@0:6;0; 'load5'=0.01;3;0:5;0; 'load15'=0.05;0:2;0:4;0;"
const invalidOutput3 = "PING OK - Packet loss = 0%, RTA = 0.30 ms"
const invalidOutput4 = "PING OK - Packet loss = 0%, RTA = 0.30 ms| =3;;;; dgasdg =;;;; sff=;;;;"

func TestParseValidOutput(t *testing.T) {
	parser := NagiosParser{
		MetricName: "nagios_test",
	}

	// Output1
	metrics, err := parser.Parse([]byte(validOutput1))
	require.NoError(t, err)
	require.Len(t, metrics, 2)
	// rta
	assert.Equal(t, "rta", metrics[0].Tags()["perfdata"])
	assert.Equal(t, map[string]interface{}{
		"value":       float64(0.298),
		"warning_lt":  float64(0),
		"warning_gt":  float64(4000),
		"critical_lt": float64(0),
		"critical_gt": float64(6000),
		"min":         float64(0),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{"unit": "ms", "perfdata": "rta"}, metrics[0].Tags())
	// pl
	assert.Equal(t, "pl", metrics[1].Tags()["perfdata"])
	assert.Equal(t, map[string]interface{}{
		"value":       float64(0),
		"warning_lt":  float64(0),
		"warning_gt":  float64(80),
		"critical_lt": float64(0),
		"critical_gt": float64(90),
		"min":         float64(0),
		"max":         float64(100),
	}, metrics[1].Fields())
	assert.Equal(t, map[string]string{"unit": "%", "perfdata": "pl"}, metrics[1].Tags())

	// Output2
	metrics, err = parser.Parse([]byte(validOutput2))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	// time
	assert.Equal(t, "time", metrics[0].Tags()["perfdata"])
	assert.Equal(t, map[string]interface{}{
		"value": float64(0.008457),
		"min":   float64(0),
		"max":   float64(10),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{"unit": "s", "perfdata": "time"}, metrics[0].Tags())

	// Output3
	metrics, err = parser.Parse([]byte(validOutput3))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	// time
	assert.Equal(t, "time", metrics[0].Tags()["perfdata"])
	assert.Equal(t, map[string]interface{}{
		"value": float64(0.008457),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{"perfdata": "time"}, metrics[0].Tags())

	// Output4
	metrics, err = parser.Parse([]byte(validOutput4))
	require.NoError(t, err)
	require.Len(t, metrics, 3)
	// load
	// const validOutput4 = "OK: Load average: 0.00, 0.01, 0.05 | 'load1'=0.00;0:4;0:6;0; 'load5'=0.01;0:3;0:5;0; 'load15'=0.05;0:2;0:4;0;"
	assert.Equal(t, map[string]interface{}{
		"value":       float64(0.00),
		"warning_lt":  MinFloat64,
		"warning_gt":  float64(4),
		"critical_le": float64(0),
		"critical_ge": float64(6),
		"min":         float64(0),
	}, metrics[0].Fields())

	assert.Equal(t, map[string]string{"perfdata": "load1"}, metrics[0].Tags())
}

func TestParseInvalidOutput(t *testing.T) {
	parser := NagiosParser{
		MetricName: "nagios_test",
	}

	// invalidOutput3
	metrics, err := parser.Parse([]byte(invalidOutput3))
	require.NoError(t, err)
	require.Len(t, metrics, 0)

	// invalidOutput4
	metrics, err = parser.Parse([]byte(invalidOutput4))
	require.NoError(t, err)
	require.Len(t, metrics, 0)

}

func TestParseThreshold(t *testing.T) {
	tests := []struct {
		input string
		eMin  float64
		eMax  float64
		eErr  error
	}{
		{
			input: "10",
			eMin:  0,
			eMax:  10,
			eErr:  nil,
		},
		{
			input: "10:",
			eMin:  10,
			eMax:  MaxFloat64,
			eErr:  nil,
		},
		{
			input: "~:10",
			eMin:  MinFloat64,
			eMax:  10,
			eErr:  nil,
		},
		{
			input: "10:20",
			eMin:  10,
			eMax:  20,
			eErr:  nil,
		},
		{
			input: "10:20",
			eMin:  10,
			eMax:  20,
			eErr:  nil,
		},
		{
			input: "10:20:30",
			eMin:  0,
			eMax:  0,
			eErr:  ErrBadThresholdFormat,
		},
	}

	for i := range tests {
		min, max, err := parseThreshold(tests[i].input)
		require.Equal(t, tests[i].eMin, min)
		require.Equal(t, tests[i].eMax, max)
		require.Equal(t, tests[i].eErr, err)
	}
}
