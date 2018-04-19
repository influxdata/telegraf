package influx

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/require"
)

func TestReader(t *testing.T) {
	tests := []struct {
		name         string
		maxLineBytes int
		bufferSize   int
		input        []telegraf.Metric
		expected     []byte
	}{
		{
			name:         "minimal",
			maxLineBytes: 4096,
			bufferSize:   20,
			input: []telegraf.Metric{
				MustMetric(
					metric.New(
						"cpu",
						map[string]string{},
						map[string]interface{}{
							"value": 42.0,
						},
						time.Unix(0, 0),
					),
				),
			},
			expected: []byte("cpu value=42 0\n"),
		},
		{
			name:         "multiple lines",
			maxLineBytes: 4096,
			bufferSize:   20,
			input: []telegraf.Metric{
				MustMetric(
					metric.New(
						"cpu",
						map[string]string{},
						map[string]interface{}{
							"value": 42.0,
						},
						time.Unix(0, 0),
					),
				),
				MustMetric(
					metric.New(
						"cpu",
						map[string]string{},
						map[string]interface{}{
							"value": 42.0,
						},
						time.Unix(0, 0),
					),
				),
			},
			expected: []byte("cpu value=42 0\ncpu value=42 0\n"),
		},
		{
			name:         "exact fit",
			maxLineBytes: 4096,
			bufferSize:   15,
			input: []telegraf.Metric{
				MustMetric(
					metric.New(
						"cpu",
						map[string]string{},
						map[string]interface{}{
							"value": 42.0,
						},
						time.Unix(0, 0),
					),
				),
			},
			expected: []byte("cpu value=42 0\n"),
		},
		{
			name:         "continue on failed metrics",
			maxLineBytes: 4096,
			bufferSize:   15,
			input: []telegraf.Metric{
				MustMetric(
					metric.New(
						"",
						map[string]string{},
						map[string]interface{}{
							"value": 42.0,
						},
						time.Unix(0, 0),
					),
				),
				MustMetric(
					metric.New(
						"cpu",
						map[string]string{},
						map[string]interface{}{
							"value": 42.0,
						},
						time.Unix(0, 0),
					),
				),
			},
			expected: []byte("cpu value=42 0\n"),
		},
		{
			name:         "last metric failed regression",
			maxLineBytes: 4096,
			bufferSize:   15,
			input: []telegraf.Metric{
				MustMetric(
					metric.New(
						"cpu",
						map[string]string{},
						map[string]interface{}{
							"value": 42.0,
						},
						time.Unix(0, 0),
					),
				),
				MustMetric(
					metric.New(
						"",
						map[string]string{},
						map[string]interface{}{
							"value": 42.0,
						},
						time.Unix(0, 0),
					),
				),
			},
			expected: []byte("cpu value=42 0\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serializer := NewSerializer()
			serializer.SetMaxLineBytes(tt.maxLineBytes)
			serializer.SetFieldSortOrder(SortFields)
			reader := NewReader(tt.input, serializer)

			data := new(bytes.Buffer)
			readbuf := make([]byte, tt.bufferSize)

			total := 0
			for {
				n, err := reader.Read(readbuf)
				total += n
				if err == io.EOF {
					break
				}

				data.Write(readbuf[:n])
				require.NoError(t, err)
			}
			require.Equal(t, tt.expected, data.Bytes())
			require.Equal(t, len(tt.expected), total)
		})
	}
}

func TestZeroLengthBufferNoError(t *testing.T) {
	m := MustMetric(
		metric.New(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"value": 42.0,
			},
			time.Unix(0, 0),
		),
	)
	serializer := NewSerializer()
	serializer.SetFieldSortOrder(SortFields)
	reader := NewReader([]telegraf.Metric{m}, serializer)

	readbuf := make([]byte, 0)

	n, err := reader.Read(readbuf)
	require.NoError(t, err)
	require.Equal(t, 0, n)
}
