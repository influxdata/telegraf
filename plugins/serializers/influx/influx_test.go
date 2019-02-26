package influx

import (
	"math"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/require"
)

func MustMetric(v telegraf.Metric, err error) telegraf.Metric {
	if err != nil {
		panic(err)
	}
	return v
}

var tests = []struct {
	name        string
	maxBytes    int
	typeSupport FieldTypeSupport
	input       telegraf.Metric
	output      []byte
	errReason   string
}{
	{
		name: "minimal",
		input: MustMetric(
			metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"value": 42.0,
				},
				time.Unix(0, 0),
			),
		),
		output: []byte("cpu value=42 0\n"),
	},
	{
		name: "multiple tags",
		input: MustMetric(
			metric.New(
				"cpu",
				map[string]string{
					"host": "localhost",
					"cpu":  "CPU0",
				},
				map[string]interface{}{
					"value": 42.0,
				},
				time.Unix(0, 0),
			),
		),
		output: []byte("cpu,cpu=CPU0,host=localhost value=42 0\n"),
	},
	{
		name: "multiple fields",
		input: MustMetric(
			metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"x": 42.0,
					"y": 42.0,
				},
				time.Unix(0, 0),
			),
		),
		output: []byte("cpu x=42,y=42 0\n"),
	},
	{
		name: "float NaN",
		input: MustMetric(
			metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"x": math.NaN(),
					"y": 42,
				},
				time.Unix(0, 0),
			),
		),
		output: []byte("cpu y=42i 0\n"),
	},
	{
		name: "float NaN only",
		input: MustMetric(
			metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"value": math.NaN(),
				},
				time.Unix(0, 0),
			),
		),
		errReason: NoFields,
	},
	{
		name: "float Inf",
		input: MustMetric(
			metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"value": math.Inf(1),
					"y":     42,
				},
				time.Unix(0, 0),
			),
		),
		output: []byte("cpu y=42i 0\n"),
	},
	{
		name: "integer field",
		input: MustMetric(
			metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"value": 42,
				},
				time.Unix(0, 0),
			),
		),
		output: []byte("cpu value=42i 0\n"),
	},
	{
		name: "integer field 64-bit",
		input: MustMetric(
			metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"value": int64(123456789012345),
				},
				time.Unix(0, 0),
			),
		),
		output: []byte("cpu value=123456789012345i 0\n"),
	},
	{
		name: "uint field",
		input: MustMetric(
			metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"value": uint64(42),
				},
				time.Unix(0, 0),
			),
		),
		output:      []byte("cpu value=42u 0\n"),
		typeSupport: UintSupport,
	},
	{
		name: "uint field max value",
		input: MustMetric(
			metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"value": uint64(18446744073709551615),
				},
				time.Unix(0, 0),
			),
		),
		output:      []byte("cpu value=18446744073709551615u 0\n"),
		typeSupport: UintSupport,
	},
	{
		name: "uint field no uint support",
		input: MustMetric(
			metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"value": uint64(42),
				},
				time.Unix(0, 0),
			),
		),
		output: []byte("cpu value=42i 0\n"),
	},
	{
		name: "uint field no uint support overflow",
		input: MustMetric(
			metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"value": uint64(18446744073709551615),
				},
				time.Unix(0, 0),
			),
		),
		output: []byte("cpu value=9223372036854775807i 0\n"),
	},
	{
		name: "bool field",
		input: MustMetric(
			metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"value": true,
				},
				time.Unix(0, 0),
			),
		),
		output: []byte("cpu value=true 0\n"),
	},
	{
		name: "string field",
		input: MustMetric(
			metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"value": "howdy",
				},
				time.Unix(0, 0),
			),
		),
		output: []byte("cpu value=\"howdy\" 0\n"),
	},
	{
		name: "timestamp",
		input: MustMetric(
			metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"value": 42.0,
				},
				time.Unix(1519194109, 42),
			),
		),
		output: []byte("cpu value=42 1519194109000000042\n"),
	},
	{
		name:     "split fields exact",
		maxBytes: 33,
		input: MustMetric(
			metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"abc": 123,
					"def": 456,
				},
				time.Unix(1519194109, 42),
			),
		),
		output: []byte("cpu abc=123i 1519194109000000042\ncpu def=456i 1519194109000000042\n"),
	},
	{
		name:     "split fields extra",
		maxBytes: 34,
		input: MustMetric(
			metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"abc": 123,
					"def": 456,
				},
				time.Unix(1519194109, 42),
			),
		),
		output: []byte("cpu abc=123i 1519194109000000042\ncpu def=456i 1519194109000000042\n"),
	},
	{
		name:     "split_fields_overflow",
		maxBytes: 43,
		input: MustMetric(
			metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"abc": 123,
					"def": 456,
					"ghi": 789,
					"jkl": 123,
				},
				time.Unix(1519194109, 42),
			),
		),
		output: []byte("cpu abc=123i,def=456i 1519194109000000042\ncpu ghi=789i,jkl=123i 1519194109000000042\n"),
	},
	{
		name: "name newline",
		input: MustMetric(
			metric.New(
				"c\npu",
				map[string]string{},
				map[string]interface{}{
					"value": 42,
				},
				time.Unix(0, 0),
			),
		),
		output: []byte("c\\npu value=42i 0\n"),
	},
	{
		name: "tag newline",
		input: MustMetric(
			metric.New(
				"cpu",
				map[string]string{
					"host": "x\ny",
				},
				map[string]interface{}{
					"value": 42,
				},
				time.Unix(0, 0),
			),
		),
		output: []byte("cpu,host=x\\ny value=42i 0\n"),
	},
	{
		name: "string newline",
		input: MustMetric(
			metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"value": "x\ny",
				},
				time.Unix(0, 0),
			),
		),
		output: []byte("cpu value=\"x\ny\" 0\n"),
	},
	{
		name:     "need more space",
		maxBytes: 32,
		input: MustMetric(
			metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"abc": 123,
					"def": 456,
				},
				time.Unix(1519194109, 42),
			),
		),
		output:    nil,
		errReason: NeedMoreSpace,
	},
	{
		name: "no fields",
		input: MustMetric(
			metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{},
				time.Unix(0, 0),
			),
		),
		errReason: NoFields,
	},
	{
		name: "procstat",
		input: MustMetric(
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
		output: []byte("procstat,exe=bash,process_name=bash cpu_time=0i,cpu_time_guest=0,cpu_time_guest_nice=0,cpu_time_idle=0,cpu_time_iowait=0,cpu_time_irq=0,cpu_time_nice=0,cpu_time_soft_irq=0,cpu_time_steal=0,cpu_time_stolen=0,cpu_time_system=0,cpu_time_user=0.02,cpu_usage=0,involuntary_context_switches=2i,memory_data=1576960i,memory_locked=0i,memory_rss=5103616i,memory_stack=139264i,memory_swap=0i,memory_vms=21659648i,nice_priority=20i,num_fds=4i,num_threads=1i,pid=29417i,read_bytes=0i,read_count=259i,realtime_priority=0i,rlimit_cpu_time_hard=2147483647i,rlimit_cpu_time_soft=2147483647i,rlimit_file_locks_hard=2147483647i,rlimit_file_locks_soft=2147483647i,rlimit_memory_data_hard=2147483647i,rlimit_memory_data_soft=2147483647i,rlimit_memory_locked_hard=65536i,rlimit_memory_locked_soft=65536i,rlimit_memory_rss_hard=2147483647i,rlimit_memory_rss_soft=2147483647i,rlimit_memory_stack_hard=2147483647i,rlimit_memory_stack_soft=8388608i,rlimit_memory_vms_hard=2147483647i,rlimit_memory_vms_soft=2147483647i,rlimit_nice_priority_hard=0i,rlimit_nice_priority_soft=0i,rlimit_num_fds_hard=4096i,rlimit_num_fds_soft=1024i,rlimit_realtime_priority_hard=0i,rlimit_realtime_priority_soft=0i,rlimit_signals_pending_hard=78994i,rlimit_signals_pending_soft=78994i,signals_pending=0i,voluntary_context_switches=42i,write_bytes=106496i,write_count=35i 1517620624000000000\n"),
	},
}

func TestSerializer(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serializer := NewSerializer()
			serializer.SetMaxLineBytes(tt.maxBytes)
			serializer.SetFieldSortOrder(SortFields)
			serializer.SetFieldTypeSupport(tt.typeSupport)
			output, err := serializer.Serialize(tt.input)
			if tt.errReason != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errReason)
			}
			require.Equal(t, string(tt.output), string(output))
		})
	}
}

func BenchmarkSerializer(b *testing.B) {
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			serializer := NewSerializer()
			serializer.SetMaxLineBytes(tt.maxBytes)
			serializer.SetFieldTypeSupport(tt.typeSupport)
			for n := 0; n < b.N; n++ {
				output, err := serializer.Serialize(tt.input)
				_ = err
				_ = output
			}
		})
	}
}

func TestSerialize_SerializeBatch(t *testing.T) {
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

	metrics := []telegraf.Metric{m, m}

	serializer := NewSerializer()
	serializer.SetFieldSortOrder(SortFields)
	output, err := serializer.SerializeBatch(metrics)
	require.NoError(t, err)
	require.Equal(t, []byte("cpu value=42 0\ncpu value=42 0\n"), output)
}
