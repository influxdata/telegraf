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
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte("cpu value=42 0\n"),
		},
		{
			name:         "multiple lines",
			maxLineBytes: 4096,
			bufferSize:   20,
			input: []telegraf.Metric{
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(0, 0),
				),
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte("cpu value=42 0\ncpu value=42 0\n"),
		},
		{
			name:         "exact fit",
			maxLineBytes: 4096,
			bufferSize:   15,
			input: []telegraf.Metric{
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte("cpu value=42 0\n"),
		},
		{
			name:         "continue on failed metrics",
			maxLineBytes: 4096,
			bufferSize:   15,
			input: []telegraf.Metric{
				metric.New(
					"",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(0, 0),
				),
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte("cpu value=42 0\n"),
		},
		{
			name:         "last metric failed regression",
			maxLineBytes: 4096,
			bufferSize:   15,
			input: []telegraf.Metric{
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(0, 0),
				),
				metric.New(
					"",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(0, 0),
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
	m := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(0, 0),
	)
	serializer := NewSerializer()
	serializer.SetFieldSortOrder(SortFields)
	reader := NewReader([]telegraf.Metric{m}, serializer)

	readbuf := make([]byte, 0)

	n, err := reader.Read(readbuf)
	require.NoError(t, err)
	require.Equal(t, 0, n)
}

func BenchmarkReader(b *testing.B) {
	m := metric.New(
		"procstat",
		map[string]string{
			"exe":          "bash",
			"process_name": "bash",
		},
		map[string]interface{}{
			"cpu_time":                      0,
			"cpu_time_guest":                float64(0),
			"cpu_time_guest_nice":           float64(0),
			"cpu_time_idle":                 float64(0),
			"cpu_time_iowait":               float64(0),
			"cpu_time_irq":                  float64(0),
			"cpu_time_nice":                 float64(0),
			"cpu_time_soft_irq":             float64(0),
			"cpu_time_steal":                float64(0),
			"cpu_time_system":               float64(0),
			"cpu_time_user":                 float64(0.02),
			"cpu_usage":                     float64(0),
			"involuntary_context_switches":  2,
			"memory_data":                   1576960,
			"memory_locked":                 0,
			"memory_rss":                    5103616,
			"memory_stack":                  139264,
			"memory_swap":                   0,
			"memory_vms":                    21659648,
			"nice_priority":                 20,
			"num_fds":                       4,
			"num_threads":                   1,
			"pid":                           29417,
			"read_bytes":                    0,
			"read_count":                    259,
			"realtime_priority":             0,
			"rlimit_cpu_time_hard":          2147483647,
			"rlimit_cpu_time_soft":          2147483647,
			"rlimit_file_locks_hard":        2147483647,
			"rlimit_file_locks_soft":        2147483647,
			"rlimit_memory_data_hard":       2147483647,
			"rlimit_memory_data_soft":       2147483647,
			"rlimit_memory_locked_hard":     65536,
			"rlimit_memory_locked_soft":     65536,
			"rlimit_memory_rss_hard":        2147483647,
			"rlimit_memory_rss_soft":        2147483647,
			"rlimit_memory_stack_hard":      2147483647,
			"rlimit_memory_stack_soft":      8388608,
			"rlimit_memory_vms_hard":        2147483647,
			"rlimit_memory_vms_soft":        2147483647,
			"rlimit_nice_priority_hard":     0,
			"rlimit_nice_priority_soft":     0,
			"rlimit_num_fds_hard":           4096,
			"rlimit_num_fds_soft":           1024,
			"rlimit_realtime_priority_hard": 0,
			"rlimit_realtime_priority_soft": 0,
			"rlimit_signals_pending_hard":   78994,
			"rlimit_signals_pending_soft":   78994,
			"signals_pending":               0,
			"voluntary_context_switches":    42,
			"write_bytes":                   106496,
			"write_count":                   35,
		},
		time.Unix(0, 1517620624000000000),
	)
	metrics := make([]telegraf.Metric, 1000)
	for i := range metrics {
		metrics[i] = m
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		readbuf := make([]byte, 4096)
		serializer := NewSerializer()
		reader := NewReader(metrics, serializer)
		for {
			_, err := reader.Read(readbuf)
			if err == io.EOF {
				break
			}

			if err != nil {
				panic(err.Error())
			}
		}
	}
}
