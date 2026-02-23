package template

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers"
)

func TestSerializer(t *testing.T) {
	var tests = []struct {
		name      string
		input     telegraf.Metric
		template  string
		output    []byte
		errReason string
	}{
		{
			name: "name",
			input: metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{},
				time.Unix(100, 0),
			),
			template: "{{ .Name }}",
			output:   []byte("cpu"),
		},
		{
			name: "time",
			input: metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{},
				time.Unix(100, 0),
			),
			template: "{{ .Time.Unix }}",
			output:   []byte("100"),
		},
		{
			name: "specific field",
			input: metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"x": 42.0,
					"y": 43.0,
				},
				time.Unix(100, 0),
			),
			template: `{{ .Field "x" }}`,
			output:   []byte("42"),
		},
		{
			name: "specific tag",
			input: metric.New(
				"cpu",
				map[string]string{
					"host": "localhost",
					"cpu":  "CPU0",
				},
				map[string]interface{}{},
				time.Unix(100, 0),
			),
			template: `{{ .Tag "cpu" }}`,
			output:   []byte("CPU0"),
		},
		{
			name: "all fields",
			input: metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"x": 42.0,
					"y": 43.0,
				},
				time.Unix(100, 0),
			),
			template: `{{ range $k, $v := .Fields }}{{$k}}={{$v}},{{end}}`,
			output:   []byte("x=42,y=43,"),
		},
		{
			name: "all tags",
			input: metric.New(
				"cpu",
				map[string]string{
					"host": "localhost",
					"cpu":  "CPU0",
				},
				map[string]interface{}{},
				time.Unix(100, 0),
			),
			template: `{{ range $k, $v := .Tags }}{{$k}}={{$v}},{{end}}`,
			output:   []byte("cpu=CPU0,host=localhost,"),
		},
		{
			name: "string",
			input: metric.New(
				"cpu",
				map[string]string{
					"host": "localhost",
					"cpu":  "CPU0",
				},
				map[string]interface{}{
					"x": 42.0,
					"y": 43.0,
				},
				time.Unix(100, 0),
			),
			template: "{{ .String }}",
			output:   []byte("cpu map[cpu:CPU0 host:localhost] map[x:42 y:43] 100000000000"),
		},
		{
			name: "complex",
			input: metric.New(
				"cpu",
				map[string]string{
					"tag1": "tag",
				},
				map[string]interface{}{
					"value": 42.0,
				},
				time.Unix(0, 0),
			),
			template: `{{ .Name }} {{ range $k, $v := .Fields}}{{$k}}={{$v}}{{end}} {{ .Tag "tag1" }} {{.Time.UnixNano}} literal`,
			output:   []byte("cpu value=42 tag 0 literal"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serializer := &Serializer{
				Template: tt.template,
			}
			require.NoError(t, serializer.Init())
			output, err := serializer.Serialize(tt.input)
			if tt.errReason != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errReason)
			}
			require.Equal(t, string(tt.output), string(output))
			// Ensure we get the same output in batch mode
			batchOutput, err := serializer.SerializeBatch([]telegraf.Metric{tt.input})
			if tt.errReason != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errReason)
			}
			require.Equal(t, string(tt.output), string(batchOutput))
		})
	}
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
	s := &Serializer{BatchTemplate: `{{ range $index, $metric := . }}{{$index}}: {{$metric.Name}} {{$metric.Field "value"}}
{{end}}`}
	require.NoError(t, s.Init())
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)
	require.Equal(
		t,
		`0: cpu 42
1: cpu 42
`, string(buf),
	)
	// A batch template should still work when serializing a single metric
	singleBuf, err := s.Serialize(m)
	require.NoError(t, err)
	require.Equal(t, "0: cpu 42\n", string(singleBuf))
}

func TestSerializeTrackingMetric(t *testing.T) {
	m := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(0, 0),
	)
	tm, _ := metric.WithTracking(m, func(_ telegraf.DeliveryInfo) {})

	s := &Serializer{Template: "{{ .Name }}"}
	require.NoError(t, s.Init())

	// Serialize should handle tracking metrics
	buf, err := s.Serialize(tm)
	require.NoError(t, err)
	require.Equal(t, "cpu", string(buf))

	// SerializeBatch should also handle tracking metrics
	batchBuf, err := s.SerializeBatch([]telegraf.Metric{tm})
	require.NoError(t, err)
	require.Equal(t, "cpu", string(batchBuf))
}

func TestSerializeBatchTrackingMetrics(t *testing.T) {
	m := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(0, 0),
	)
	tm, _ := metric.WithTracking(m, func(_ telegraf.DeliveryInfo) {})

	s := &Serializer{BatchTemplate: `{{ range $index, $metric := . }}{{$index}}: {{$metric.Name}} {{$metric.Field "value"}}
{{end}}`}
	require.NoError(t, s.Init())

	buf, err := s.SerializeBatch([]telegraf.Metric{tm, tm})
	require.NoError(t, err)
	require.Equal(t, "0: cpu 42\n1: cpu 42\n", string(buf))
}

func BenchmarkSerialize(b *testing.B) {
	s := &Serializer{}
	require.NoError(b, s.Init())
	metrics := serializers.BenchmarkMetrics(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := s.Serialize(metrics[i%len(metrics)])
		require.NoError(b, err)
	}
}

func BenchmarkSerializeBatch(b *testing.B) {
	s := &Serializer{}
	require.NoError(b, s.Init())
	m := serializers.BenchmarkMetrics(b)
	metrics := m[:]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := s.SerializeBatch(metrics)
		require.NoError(b, err)
	}
}
