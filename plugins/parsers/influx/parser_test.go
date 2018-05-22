package influx

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/require"
)

func Metric(v telegraf.Metric, err error) telegraf.Metric {
	if err != nil {
		panic(err)
	}
	return v
}

var DefaultTime = func() time.Time {
	return time.Unix(42, 0)
}

var ptests = []struct {
	name      string
	input     []byte
	timeFunc  func() time.Time
	precision time.Duration
	metrics   []telegraf.Metric
	err       error
}{
	{
		name:  "minimal",
		input: []byte("cpu value=42 0"),
		metrics: []telegraf.Metric{
			Metric(
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
		err: nil,
	},
	{
		name:  "minimal with newline",
		input: []byte("cpu value=42 0\n"),
		metrics: []telegraf.Metric{
			Metric(
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
		err: nil,
	},
	{
		name:  "measurement escape space",
		input: []byte(`c\ pu value=42`),
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
					"c pu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(42, 0),
				),
			),
		},
		err: nil,
	},
	{
		name:  "measurement escape comma",
		input: []byte(`c\,pu value=42`),
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
					"c,pu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(42, 0),
				),
			),
		},
		err: nil,
	},
	{
		name:  "tags",
		input: []byte(`cpu,cpu=cpu0,host=localhost value=42`),
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
					"cpu",
					map[string]string{
						"cpu":  "cpu0",
						"host": "localhost",
					},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(42, 0),
				),
			),
		},
		err: nil,
	},
	{
		name:  "tags escape unescapable",
		input: []byte(`cpu,ho\st=localhost value=42`),
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
					"cpu",
					map[string]string{
						`ho\st`: "localhost",
					},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(42, 0),
				),
			),
		},
		err: nil,
	},
	{
		name:  "tags escape equals",
		input: []byte(`cpu,ho\=st=localhost value=42`),
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
					"cpu",
					map[string]string{
						"ho=st": "localhost",
					},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(42, 0),
				),
			),
		},
		err: nil,
	},
	{
		name:  "tags escape comma",
		input: []byte(`cpu,ho\,st=localhost value=42`),
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
					"cpu",
					map[string]string{
						"ho,st": "localhost",
					},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(42, 0),
				),
			),
		},
		err: nil,
	},
	{
		name:  "field key escape not escapable",
		input: []byte(`cpu va\lue=42`),
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						`va\lue`: 42.0,
					},
					time.Unix(42, 0),
				),
			),
		},
		err: nil,
	},
	{
		name:  "field key escape equals",
		input: []byte(`cpu va\=lue=42`),
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						`va=lue`: 42.0,
					},
					time.Unix(42, 0),
				),
			),
		},
		err: nil,
	},
	{
		name:  "field key escape comma",
		input: []byte(`cpu va\,lue=42`),
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						`va,lue`: 42.0,
					},
					time.Unix(42, 0),
				),
			),
		},
		err: nil,
	},
	{
		name:  "field key escape space",
		input: []byte(`cpu va\ lue=42`),
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						`va lue`: 42.0,
					},
					time.Unix(42, 0),
				),
			),
		},
		err: nil,
	},
	{
		name:  "field int",
		input: []byte("cpu value=42i"),
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(42, 0),
				),
			),
		},
		err: nil,
	},
	{
		name:  "field int overflow dropped",
		input: []byte("cpu value=9223372036854775808i"),
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{},
					time.Unix(42, 0),
				),
			),
		},
		err: nil,
	},
	{
		name:  "field int max value",
		input: []byte("cpu value=9223372036854775807i"),
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": int64(9223372036854775807),
					},
					time.Unix(42, 0),
				),
			),
		},
		err: nil,
	},
	{
		name:  "field uint",
		input: []byte("cpu value=42u"),
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": uint64(42),
					},
					time.Unix(42, 0),
				),
			),
		},
		err: nil,
	},
	{
		name:  "field uint overflow dropped",
		input: []byte("cpu value=18446744073709551616u"),
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{},
					time.Unix(42, 0),
				),
			),
		},
		err: nil,
	},
	{
		name:  "field uint max value",
		input: []byte("cpu value=18446744073709551615u"),
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": uint64(18446744073709551615),
					},
					time.Unix(42, 0),
				),
			),
		},
		err: nil,
	},
	{
		name:  "field boolean",
		input: []byte("cpu value=true"),
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": true,
					},
					time.Unix(42, 0),
				),
			),
		},
		err: nil,
	},
	{
		name:  "field string",
		input: []byte(`cpu value="42"`),
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": "42",
					},
					time.Unix(42, 0),
				),
			),
		},
		err: nil,
	},
	{
		name:  "field string escape quote",
		input: []byte(`cpu value="how\"dy"`),
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						`value`: `how"dy`,
					},
					time.Unix(42, 0),
				),
			),
		},
		err: nil,
	},
	{
		name:  "field string escape backslash",
		input: []byte(`cpu value="how\\dy"`),
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						`value`: `how\dy`,
					},
					time.Unix(42, 0),
				),
			),
		},
		err: nil,
	},
	{
		name:  "no timestamp",
		input: []byte("cpu value=42"),
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(42, 0),
				),
			),
		},
		err: nil,
	},
	{
		name:  "no timestamp full precision",
		input: []byte("cpu value=42"),
		timeFunc: func() time.Time {
			return time.Unix(42, 123456789)
		},
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(42, 123456789),
				),
			),
		},
		err: nil,
	},
	{
		name:  "no timestamp partial precision",
		input: []byte("cpu value=42"),
		timeFunc: func() time.Time {
			return time.Unix(42, 123456789)
		},
		precision: 1 * time.Millisecond,
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(42, 123000000),
				),
			),
		},
		err: nil,
	},
	{
		name:  "multiple lines",
		input: []byte("cpu value=42\ncpu value=42"),
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(42, 0),
				),
			),
			Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(42, 0),
				),
			),
		},
		err: nil,
	},
	{
		name:    "invalid measurement only",
		input:   []byte("cpu"),
		metrics: nil,
		err: &ParseError{
			Offset: 3,
			msg:    ErrFieldParse.Error(),
			buf:    "cpu",
		},
	},
	{
		name:  "procstat",
		input: []byte("procstat,exe=bash,process_name=bash voluntary_context_switches=42i,memory_rss=5103616i,rlimit_memory_data_hard=2147483647i,cpu_time_user=0.02,rlimit_file_locks_soft=2147483647i,pid=29417i,cpu_time_nice=0,rlimit_memory_locked_soft=65536i,read_count=259i,rlimit_memory_vms_hard=2147483647i,memory_swap=0i,rlimit_num_fds_soft=1024i,rlimit_nice_priority_hard=0i,cpu_time_soft_irq=0,cpu_time=0i,rlimit_memory_locked_hard=65536i,realtime_priority=0i,signals_pending=0i,nice_priority=20i,cpu_time_idle=0,memory_stack=139264i,memory_locked=0i,rlimit_memory_stack_soft=8388608i,cpu_time_iowait=0,cpu_time_guest=0,cpu_time_guest_nice=0,rlimit_memory_data_soft=2147483647i,read_bytes=0i,rlimit_cpu_time_soft=2147483647i,involuntary_context_switches=2i,write_bytes=106496i,cpu_time_system=0,cpu_time_irq=0,cpu_usage=0,memory_vms=21659648i,memory_data=1576960i,rlimit_memory_stack_hard=2147483647i,num_threads=1i,cpu_time_stolen=0,rlimit_memory_rss_soft=2147483647i,rlimit_realtime_priority_soft=0i,num_fds=4i,write_count=35i,rlimit_signals_pending_soft=78994i,cpu_time_steal=0,rlimit_num_fds_hard=4096i,rlimit_file_locks_hard=2147483647i,rlimit_cpu_time_hard=2147483647i,rlimit_signals_pending_hard=78994i,rlimit_nice_priority_soft=0i,rlimit_memory_rss_hard=2147483647i,rlimit_memory_vms_soft=2147483647i,rlimit_realtime_priority_hard=0i 1517620624000000000"),
		metrics: []telegraf.Metric{
			Metric(
				metric.New(
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
						"cpu_time_stolen":               float64(0),
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
				),
			),
		},
		err: nil,
	},
}

func TestParser(t *testing.T) {
	for _, tt := range ptests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewMetricHandler()
			handler.SetTimeFunc(DefaultTime)
			if tt.timeFunc != nil {
				handler.SetTimeFunc(tt.timeFunc)
			}
			if tt.precision > 0 {
				handler.SetTimePrecision(tt.precision)
			}
			parser := NewParser(handler)

			metrics, err := parser.Parse(tt.input)
			require.Equal(t, tt.err, err)

			require.Equal(t, len(tt.metrics), len(metrics))
			for i, expected := range tt.metrics {
				require.Equal(t, expected.Name(), metrics[i].Name())
				require.Equal(t, expected.Tags(), metrics[i].Tags())
				require.Equal(t, expected.Fields(), metrics[i].Fields())
				require.Equal(t, expected.Time(), metrics[i].Time())
			}
		})
	}
}

func BenchmarkParser(b *testing.B) {
	for _, tt := range ptests {
		b.Run(tt.name, func(b *testing.B) {
			handler := NewMetricHandler()
			parser := NewParser(handler)
			for n := 0; n < b.N; n++ {
				metrics, err := parser.Parse(tt.input)
				_ = err
				_ = metrics
			}
		})
	}
}

func TestSeriesParser(t *testing.T) {
	var tests = []struct {
		name      string
		input     []byte
		timeFunc  func() time.Time
		precision time.Duration
		metrics   []telegraf.Metric
		err       error
	}{
		{
			name:    "empty",
			input:   []byte(""),
			metrics: []telegraf.Metric{},
		},
		{
			name:  "minimal",
			input: []byte("cpu"),
			metrics: []telegraf.Metric{
				Metric(
					metric.New(
						"cpu",
						map[string]string{},
						map[string]interface{}{},
						time.Unix(0, 0),
					),
				),
			},
		},
		{
			name:  "tags",
			input: []byte("cpu,a=x,b=y"),
			metrics: []telegraf.Metric{
				Metric(
					metric.New(
						"cpu",
						map[string]string{
							"a": "x",
							"b": "y",
						},
						map[string]interface{}{},
						time.Unix(0, 0),
					),
				),
			},
		},
		{
			name:    "missing tag value",
			input:   []byte("cpu,a="),
			metrics: []telegraf.Metric{},
			err: &ParseError{
				Offset: 6,
				msg:    ErrTagParse.Error(),
				buf:    "cpu,a=",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewMetricHandler()
			handler.SetTimeFunc(DefaultTime)
			if tt.timeFunc != nil {
				handler.SetTimeFunc(tt.timeFunc)
			}
			if tt.precision > 0 {
				handler.SetTimePrecision(tt.precision)
			}
			parser := NewSeriesParser(handler)

			metrics, err := parser.Parse(tt.input)
			require.Equal(t, tt.err, err)

			require.Equal(t, len(tt.metrics), len(metrics))
			for i, expected := range tt.metrics {
				require.Equal(t, expected.Name(), metrics[i].Name())
				require.Equal(t, expected.Tags(), metrics[i].Tags())
			}
		})
	}
}
