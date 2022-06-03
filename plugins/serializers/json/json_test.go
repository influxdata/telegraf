package json

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/toml"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestSerializeMetricFloat(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m := metric.New("cpu", tags, fields, now)

	s, _ := NewSerializer(0, "", "")
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)
	expS := []byte(fmt.Sprintf(`{"fields":{"usage_idle":91.5},"name":"cpu","tags":{"cpu":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
	require.Equal(t, string(expS), string(buf))
}

func TestSerialize_TimestampUnits(t *testing.T) {
	tests := []struct {
		name            string
		timestampUnits  time.Duration
		timestampFormat string
		expected        string
	}{
		{
			name:           "default of 1s",
			timestampUnits: 0,
			expected:       `{"fields":{"value":42},"name":"cpu","tags":{},"timestamp":1525478795}`,
		},
		{
			name:           "1ns",
			timestampUnits: 1 * time.Nanosecond,
			expected:       `{"fields":{"value":42},"name":"cpu","tags":{},"timestamp":1525478795123456789}`,
		},
		{
			name:           "1ms",
			timestampUnits: 1 * time.Millisecond,
			expected:       `{"fields":{"value":42},"name":"cpu","tags":{},"timestamp":1525478795123}`,
		},
		{
			name:           "10ms",
			timestampUnits: 10 * time.Millisecond,
			expected:       `{"fields":{"value":42},"name":"cpu","tags":{},"timestamp":152547879512}`,
		},
		{
			name:           "15ms is reduced to 10ms",
			timestampUnits: 15 * time.Millisecond,
			expected:       `{"fields":{"value":42},"name":"cpu","tags":{},"timestamp":152547879512}`,
		},
		{
			name:           "65ms is reduced to 10ms",
			timestampUnits: 65 * time.Millisecond,
			expected:       `{"fields":{"value":42},"name":"cpu","tags":{},"timestamp":152547879512}`,
		},
		{
			name:            "timestamp format",
			timestampFormat: "2006-01-02T15:04:05Z07:00",
			expected:        `{"fields":{"value":42},"name":"cpu","tags":{},"timestamp":"2018-05-05T00:06:35Z"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"value": 42.0,
				},
				time.Unix(1525478795, 123456789),
			)
			s, _ := NewSerializer(tt.timestampUnits, tt.timestampFormat, "")
			actual, err := s.Serialize(m)
			require.NoError(t, err)
			require.Equal(t, tt.expected+"\n", string(actual))
		})
	}
}

func TestSerializeMetricInt(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": int64(90),
	}
	m := metric.New("cpu", tags, fields, now)

	s, _ := NewSerializer(0, "", "")
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"fields":{"usage_idle":90},"name":"cpu","tags":{"cpu":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
	require.Equal(t, string(expS), string(buf))
}

func TestSerializeMetricString(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": "foobar",
	}
	m := metric.New("cpu", tags, fields, now)

	s, _ := NewSerializer(0, "", "")
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"fields":{"usage_idle":"foobar"},"name":"cpu","tags":{"cpu":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
	require.Equal(t, string(expS), string(buf))
}

func TestSerializeMultiFields(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle":  int64(90),
		"usage_total": 8559615,
	}
	m := metric.New("cpu", tags, fields, now)

	s, _ := NewSerializer(0, "", "")
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"fields":{"usage_idle":90,"usage_total":8559615},"name":"cpu","tags":{"cpu":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
	require.Equal(t, string(expS), string(buf))
}

func TestSerializeMetricWithEscapes(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu tag": "cpu0",
	}
	fields := map[string]interface{}{
		"U,age=Idle": int64(90),
	}
	m := metric.New("My CPU", tags, fields, now)

	s, _ := NewSerializer(0, "", "")
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"fields":{"U,age=Idle":90},"name":"My CPU","tags":{"cpu tag":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
	require.Equal(t, string(expS), string(buf))
}

func TestSerializeBatch(t *testing.T) {
	m := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(0, 0),
	)

	metrics := []telegraf.Metric{m, m}
	s, _ := NewSerializer(0, "", "")
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)
	require.Equal(t, []byte(`{"metrics":[{"fields":{"value":42},"name":"cpu","tags":{},"timestamp":0},{"fields":{"value":42},"name":"cpu","tags":{},"timestamp":0}]}`), buf)
}

func TestSerializeBatchSkipInf(t *testing.T) {
	metrics := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"inf":       math.Inf(1),
				"time_idle": 42,
			},
			time.Unix(0, 0),
		),
	}

	s, err := NewSerializer(0, "", "")
	require.NoError(t, err)
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)
	require.Equal(t, []byte(`{"metrics":[{"fields":{"time_idle":42},"name":"cpu","tags":{},"timestamp":0}]}`), buf)
}

func TestSerializeBatchSkipInfAllFields(t *testing.T) {
	metrics := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"inf": math.Inf(1),
			},
			time.Unix(0, 0),
		),
	}

	s, err := NewSerializer(0, "", "")
	require.NoError(t, err)
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)
	require.Equal(t, []byte(`{"metrics":[{"fields":{},"name":"cpu","tags":{},"timestamp":0}]}`), buf)
}

func TestSerializeTransformationNonBatch(t *testing.T) {
	var tests = []struct {
		name     string
		filename string
	}{
		{
			name:     "non-batch transformation test",
			filename: "testcases/transformation_single.conf",
		},
	}
	parser := influx.NewParser(influx.NewMetricHandler())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := filepath.FromSlash(tt.filename)
			cfg, header, err := loadTestConfiguration(filename)
			require.NoError(t, err)

			// Get the input metrics
			metrics, err := testutil.ParseMetricsFrom(header, "Input:", parser)
			require.NoError(t, err)

			// Get the expectations
			expectedArray, err := loadJSON(strings.TrimSuffix(filename, ".conf") + "_out.json")
			require.NoError(t, err)
			expected := expectedArray.([]interface{})

			// Serialize
			serializer, err := NewSerializer(cfg.TimestampUnits, cfg.TimestampFormat, cfg.Transformation)
			require.NoError(t, err)
			for i, m := range metrics {
				buf, err := serializer.Serialize(m)
				require.NoError(t, err)

				// Compare
				var actual interface{}
				require.NoError(t, json.Unmarshal(buf, &actual))
				fmt.Printf("actual: %v\n", actual)
				fmt.Printf("expected: %v\n", expected[i])
				require.EqualValuesf(t, expected[i], actual, "mismatch in %d", i)
			}
		})
	}
}

func TestSerializeTransformationBatch(t *testing.T) {
	var tests = []struct {
		name     string
		filename string
	}{
		{
			name:     "batch transformation test",
			filename: "testcases/transformation_batch.conf",
		},
	}
	parser := influx.NewParser(influx.NewMetricHandler())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := filepath.FromSlash(tt.filename)
			cfg, header, err := loadTestConfiguration(filename)
			require.NoError(t, err)

			// Get the input metrics
			metrics, err := testutil.ParseMetricsFrom(header, "Input:", parser)
			require.NoError(t, err)

			// Get the expectations
			expected, err := loadJSON(strings.TrimSuffix(filename, ".conf") + "_out.json")
			require.NoError(t, err)

			// Serialize
			serializer, err := NewSerializer(cfg.TimestampUnits, cfg.TimestampFormat, cfg.Transformation)
			require.NoError(t, err)
			buf, err := serializer.SerializeBatch(metrics)
			require.NoError(t, err)

			// Compare
			var actual interface{}
			require.NoError(t, json.Unmarshal(buf, &actual))
			require.EqualValues(t, expected, actual)
		})
	}
}

type Config struct {
	TimestampUnits  time.Duration `toml:"json_timestamp_units"`
	TimestampFormat string        `toml:"json_timestamp_format"`
	Transformation  string        `toml:"json_transformation"`
}

func loadTestConfiguration(filename string) (*Config, []string, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}

	header := make([]string, 0)
	for _, line := range strings.Split(string(buf), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			header = append(header, line)
		}
	}
	var cfg Config
	err = toml.Unmarshal(buf, &cfg)
	return &cfg, header, err
}

func loadJSON(filename string) (interface{}, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var data interface{}
	err = json.Unmarshal(buf, &data)
	return data, err
}
