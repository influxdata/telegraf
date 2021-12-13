package wavefront

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/outputs/wavefront"
)

func TestBuildTags(t *testing.T) {
	var tagTests = []struct {
		ptIn      map[string]string
		outTags   map[string]string
		outSource string
	}{
		{
			map[string]string{"one": "two", "three": "four", "host": "testHost"},
			map[string]string{"one": "two", "three": "four"},
			"testHost",
		},
		{
			map[string]string{"aaa": "bbb", "host": "testHost"},
			map[string]string{"aaa": "bbb"},
			"testHost",
		},
		{
			map[string]string{"bbb": "789", "aaa": "123", "host": "testHost"},
			map[string]string{"aaa": "123", "bbb": "789"},
			"testHost",
		},
		{
			map[string]string{"host": "aaa", "dc": "bbb"},
			map[string]string{"dc": "bbb"},
			"aaa",
		},
		{
			map[string]string{"instanceid": "i-0123456789", "host": "aaa", "dc": "bbb"},
			map[string]string{"dc": "bbb", "telegraf_host": "aaa"},
			"i-0123456789",
		},
		{
			map[string]string{"instance-id": "i-0123456789", "host": "aaa", "dc": "bbb"},
			map[string]string{"dc": "bbb", "telegraf_host": "aaa"},
			"i-0123456789",
		},
		{
			map[string]string{"instanceid": "i-0123456789", "host": "aaa", "hostname": "ccc", "dc": "bbb"},
			map[string]string{"dc": "bbb", "hostname": "ccc", "telegraf_host": "aaa"},
			"i-0123456789",
		},
		{
			map[string]string{"instanceid": "i-0123456789", "host": "aaa", "snmp_host": "ccc", "dc": "bbb"},
			map[string]string{"dc": "bbb", "snmp_host": "ccc", "telegraf_host": "aaa"},
			"i-0123456789",
		},
		{
			map[string]string{"host": "aaa", "snmp_host": "ccc", "dc": "bbb"},
			map[string]string{"dc": "bbb", "telegraf_host": "aaa"},
			"ccc",
		},
	}
	s := WavefrontSerializer{SourceOverride: []string{"instanceid", "instance-id", "hostname", "snmp_host", "node_host"}}

	for _, tt := range tagTests {
		source, tags := buildTags(tt.ptIn, &s)
		if !reflect.DeepEqual(tags, tt.outTags) {
			t.Errorf("\nexpected\t%+v\nreceived\t%+v\n", tt.outTags, tags)
		}
		if source != tt.outSource {
			t.Errorf("\nexpected\t%s\nreceived\t%s\n", tt.outSource, source)
		}
	}
}

func TestBuildTagsHostTag(t *testing.T) {
	var tagTests = []struct {
		ptIn      map[string]string
		outTags   map[string]string
		outSource string
	}{
		{
			map[string]string{"one": "two", "host": "testHost", "snmp_host": "snmpHost"},
			map[string]string{"telegraf_host": "testHost", "one": "two"},
			"snmpHost",
		},
	}
	s := WavefrontSerializer{SourceOverride: []string{"snmp_host"}}

	for _, tt := range tagTests {
		source, tags := buildTags(tt.ptIn, &s)
		if !reflect.DeepEqual(tags, tt.outTags) {
			t.Errorf("\nexpected\t%+v\nreceived\t%+v\n", tt.outTags, tags)
		}
		if source != tt.outSource {
			t.Errorf("\nexpected\t%s\nreceived\t%s\n", tt.outSource, source)
		}
	}
}

func TestFormatMetricPoint(t *testing.T) {
	var pointTests = []struct {
		ptIn *wavefront.MetricPoint
		out  string
	}{
		{
			&wavefront.MetricPoint{
				Metric:    "cpu.idle",
				Value:     1,
				Timestamp: 1554172967,
				Source:    "testHost",
				Tags:      map[string]string{"aaa": "bbb"},
			},
			"\"cpu.idle\" 1.000000 1554172967 source=\"testHost\" \"aaa\"=\"bbb\"\n",
		},
		{
			&wavefront.MetricPoint{
				Metric:    "cpu.idle",
				Value:     1,
				Timestamp: 1554172967,
				Source:    "testHost",
				Tags:      map[string]string{"sp&c!al/chars,": "get*replaced"},
			},
			"\"cpu.idle\" 1.000000 1554172967 source=\"testHost\" \"sp-c-al-chars-\"=\"get-replaced\"\n",
		},
	}

	s := WavefrontSerializer{}

	for _, pt := range pointTests {
		bout := formatMetricPoint(new(buffer), pt.ptIn, &s)
		sout := string(bout[:])
		if sout != pt.out {
			t.Errorf("\nexpected\t%s\nreceived\t%s\n", pt.out, sout)
		}
	}
}

func TestUseStrict(t *testing.T) {
	var pointTests = []struct {
		ptIn *wavefront.MetricPoint
		out  string
	}{
		{
			&wavefront.MetricPoint{
				Metric:    "cpu.idle",
				Value:     1,
				Timestamp: 1554172967,
				Source:    "testHost",
				Tags:      map[string]string{"sp&c!al/chars,": "get*replaced"},
			},
			"\"cpu.idle\" 1.000000 1554172967 source=\"testHost\" \"sp-c-al/chars,\"=\"get-replaced\"\n",
		},
	}

	s := WavefrontSerializer{UseStrict: true}

	for _, pt := range pointTests {
		bout := formatMetricPoint(new(buffer), pt.ptIn, &s)
		sout := string(bout[:])
		if sout != pt.out {
			t.Errorf("\nexpected\t%s\nreceived\t%s\n", pt.out, sout)
		}
	}
}

func TestSerializeMetricFloat(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu":  "cpu0",
		"host": "realHost",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m := metric.New("cpu", tags, fields, now)

	s := WavefrontSerializer{}
	buf, _ := s.Serialize(m)
	mS := strings.Split(strings.TrimSpace(string(buf)), "\n")

	expS := []string{fmt.Sprintf("\"cpu.usage.idle\" 91.500000 %d source=\"realHost\" \"cpu\"=\"cpu0\"", now.UnixNano()/1000000000)}
	require.Equal(t, expS, mS)
}

func TestSerializeMetricInt(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu":  "cpu0",
		"host": "realHost",
	}
	fields := map[string]interface{}{
		"usage_idle": int64(91),
	}
	m := metric.New("cpu", tags, fields, now)

	s := WavefrontSerializer{}
	buf, _ := s.Serialize(m)
	mS := strings.Split(strings.TrimSpace(string(buf)), "\n")

	expS := []string{fmt.Sprintf("\"cpu.usage.idle\" 91.000000 %d source=\"realHost\" \"cpu\"=\"cpu0\"", now.UnixNano()/1000000000)}
	require.Equal(t, expS, mS)
}

func TestSerializeMetricBoolTrue(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu":  "cpu0",
		"host": "realHost",
	}
	fields := map[string]interface{}{
		"usage_idle": true,
	}
	m := metric.New("cpu", tags, fields, now)

	s := WavefrontSerializer{}
	buf, _ := s.Serialize(m)
	mS := strings.Split(strings.TrimSpace(string(buf)), "\n")

	expS := []string{fmt.Sprintf("\"cpu.usage.idle\" 1.000000 %d source=\"realHost\" \"cpu\"=\"cpu0\"", now.UnixNano()/1000000000)}
	require.Equal(t, expS, mS)
}

func TestSerializeMetricBoolFalse(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu":  "cpu0",
		"host": "realHost",
	}
	fields := map[string]interface{}{
		"usage_idle": false,
	}
	m := metric.New("cpu", tags, fields, now)

	s := WavefrontSerializer{}
	buf, _ := s.Serialize(m)
	mS := strings.Split(strings.TrimSpace(string(buf)), "\n")

	expS := []string{fmt.Sprintf("\"cpu.usage.idle\" 0.000000 %d source=\"realHost\" \"cpu\"=\"cpu0\"", now.UnixNano()/1000000000)}
	require.Equal(t, expS, mS)
}

func TestSerializeMetricFieldValue(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu":  "cpu0",
		"host": "realHost",
	}
	fields := map[string]interface{}{
		"value": int64(91),
	}
	m := metric.New("cpu", tags, fields, now)

	s := WavefrontSerializer{}
	buf, err := s.Serialize(m)
	require.NoError(t, err)
	mS := strings.Split(strings.TrimSpace(string(buf)), "\n")

	expS := []string{fmt.Sprintf("\"cpu\" 91.000000 %d source=\"realHost\" \"cpu\"=\"cpu0\"", now.UnixNano()/1000000000)}
	require.Equal(t, expS, mS)
}

func TestSerializeMetricPrefix(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu":  "cpu0",
		"host": "realHost",
	}
	fields := map[string]interface{}{
		"usage_idle": int64(91),
	}
	m := metric.New("cpu", tags, fields, now)

	s := WavefrontSerializer{Prefix: "telegraf."}
	buf, err := s.Serialize(m)
	require.NoError(t, err)
	mS := strings.Split(strings.TrimSpace(string(buf)), "\n")

	expS := []string{fmt.Sprintf("\"telegraf.cpu.usage.idle\" 91.000000 %d source=\"realHost\" \"cpu\"=\"cpu0\"", now.UnixNano()/1000000000)}
	require.Equal(t, expS, mS)
}

func benchmarkMetrics(b *testing.B) [4]telegraf.Metric {
	b.Helper()
	now := time.Now()
	tags := map[string]string{
		"cpu":  "cpu0",
		"host": "realHost",
	}
	newMetric := func(v interface{}) telegraf.Metric {
		fields := map[string]interface{}{
			"usage_idle": v,
		}
		m := metric.New("cpu", tags, fields, now)
		return m
	}
	return [4]telegraf.Metric{
		newMetric(91.5),
		newMetric(91),
		newMetric(true),
		newMetric(false),
	}
}

func BenchmarkSerialize(b *testing.B) {
	var s WavefrontSerializer
	metrics := benchmarkMetrics(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := s.Serialize(metrics[i%len(metrics)])
		require.NoError(b, err)
	}
}

func BenchmarkSerializeBatch(b *testing.B) {
	var s WavefrontSerializer
	m := benchmarkMetrics(b)
	metrics := m[:]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := s.SerializeBatch(metrics)
		require.NoError(b, err)
	}
}
