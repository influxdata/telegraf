package graphite

import (
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/internal/templating"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func BenchmarkParse(b *testing.B) {
	p, err := NewGraphiteParser("_", []string{
		"*.* .wrong.measurement*",
		"servers.* .host.measurement*",
		"servers.localhost .host.measurement*",
		"*.localhost .host.measurement*",
		"*.*.cpu .host.measurement*",
		"a.b.c .host.measurement*",
		"influxd.*.foo .host.measurement*",
		"prod.*.mem .host.measurement*",
	}, nil)

	if err != nil {
		b.Fatalf("unexpected error creating parser, got %v", err)
	}

	for i := 0; i < b.N; i++ {
		_, err := p.Parse([]byte("servers.localhost.cpu.load 11 1435077219"))
		require.NoError(b, err)
	}
}

func TestTemplateApply(t *testing.T) {
	var tests = []struct {
		test        string
		input       string
		template    string
		measurement string
		tags        map[string]string
		err         string
	}{
		{
			test:        "metric only",
			input:       "cpu",
			template:    "measurement",
			measurement: "cpu",
		},
		{
			test:        "metric with single series",
			input:       "cpu.server01",
			template:    "measurement.hostname",
			measurement: "cpu",
			tags:        map[string]string{"hostname": "server01"},
		},
		{
			test:        "metric with multiple series",
			input:       "cpu.us-west.server01",
			template:    "measurement.region.hostname",
			measurement: "cpu",
			tags:        map[string]string{"hostname": "server01", "region": "us-west"},
		},
		{
			test:        "metric with multiple tags",
			input:       "server01.example.org.cpu.us-west",
			template:    "hostname.hostname.hostname.measurement.region",
			measurement: "cpu",
			tags:        map[string]string{"hostname": "server01.example.org", "region": "us-west"},
		},
		{
			test: "no metric",
			tags: make(map[string]string),
			err:  `no measurement specified for template. ""`,
		},
		{
			test:        "ignore unnamed",
			input:       "foo.cpu",
			template:    "measurement",
			measurement: "foo",
			tags:        make(map[string]string),
		},
		{
			test:        "name shorter than template",
			input:       "foo",
			template:    "measurement.A.B.C",
			measurement: "foo",
			tags:        make(map[string]string),
		},
		{
			test:        "wildcard measurement at end",
			input:       "prod.us-west.server01.cpu.load",
			template:    "env.zone.host.measurement*",
			measurement: "cpu.load",
			tags:        map[string]string{"env": "prod", "zone": "us-west", "host": "server01"},
		},
		{
			test:        "skip fields",
			input:       "ignore.us-west.ignore-this-too.cpu.load",
			template:    ".zone..measurement*",
			measurement: "cpu.load",
			tags:        map[string]string{"zone": "us-west"},
		},
		{
			test:        "conjoined fields",
			input:       "prod.us-west.server01.cpu.util.idle.percent",
			template:    "env.zone.host.measurement.measurement.field*",
			measurement: "cpu.util",
			tags:        map[string]string{"env": "prod", "zone": "us-west", "host": "server01"},
		},
		{
			test:        "multiple fields",
			input:       "prod.us-west.server01.cpu.util.idle.percent.free",
			template:    "env.zone.host.measurement.measurement.field.field.reading",
			measurement: "cpu.util",
			tags:        map[string]string{"env": "prod", "zone": "us-west", "host": "server01", "reading": "free"},
		},
	}

	for _, test := range tests {
		tmpl, err := templating.NewDefaultTemplateWithPattern(test.template)
		if errstr(err) != test.err {
			t.Fatalf("err does not match.  expected %v, got %v", test.err, err)
		}
		if err != nil {
			// If we erred out,it was intended and the following tests won't work
			continue
		}

		measurement, tags, _, _ := tmpl.Apply(test.input, DefaultSeparator)
		if measurement != test.measurement {
			t.Fatalf("name parse failer.  expected %v, got %v", test.measurement, measurement)
		}
		if len(tags) != len(test.tags) {
			t.Fatalf("unexpected number of tags.  expected %v, got %v", test.tags, tags)
		}
		for k, v := range test.tags {
			if tags[k] != v {
				t.Fatalf("unexpected tag value for tags[%s].  expected %q, got %q", k, v, tags[k])
			}
		}
	}
}

func TestParseMissingMeasurement(t *testing.T) {
	_, err := NewGraphiteParser("", []string{"a.b.c"}, nil)
	if err == nil {
		t.Fatalf("expected error creating parser, got nil")
	}
}

func TestParseLine(t *testing.T) {
	testTime := time.Now().Round(time.Second)
	epochTime := testTime.Unix()
	strTime := strconv.FormatInt(epochTime, 10)

	var tests = []struct {
		test        string
		input       string
		measurement string
		tags        map[string]string
		value       float64
		time        time.Time
		template    string
		err         string
	}{
		{
			test:        "normal case",
			input:       `cpu.foo.bar 50 ` + strTime,
			template:    "measurement.foo.bar",
			measurement: "cpu",
			tags: map[string]string{
				"foo": "foo",
				"bar": "bar",
			},
			value: 50,
			time:  testTime,
		},
		{
			test:        "normal case with tag",
			input:       `cpu.foo.bar;tag1=value1 50 ` + strTime,
			template:    "measurement.foo.bar",
			measurement: "cpu",
			tags: map[string]string{
				"foo":  "foo",
				"bar":  "bar",
				"tag1": "value1",
			},
			value: 50,
			time:  testTime,
		},
		{
			test:        "wrong tag names",
			input:       `cpu.foo.bar;tag!1=value1;tag^2=value2 50 ` + strTime,
			template:    "measurement.foo.bar",
			measurement: "cpu",
			tags: map[string]string{
				"foo": "foo",
				"bar": "bar",
			},
			value: 50,
			time:  testTime,
		},
		{
			test:        "empty tag name",
			input:       `cpu.foo.bar;=value1 50 ` + strTime,
			template:    "measurement.foo.bar",
			measurement: "cpu",
			tags: map[string]string{
				"foo": "foo",
				"bar": "bar",
			},
			value: 50,
			time:  testTime,
		},
		{
			test:        "wrong tag value",
			input:       `cpu.foo.bar;tag1=~value1 50 ` + strTime,
			template:    "measurement.foo.bar",
			measurement: "cpu",
			tags: map[string]string{
				"foo": "foo",
				"bar": "bar",
			},
			value: 50,
			time:  testTime,
		},
		{
			test:        "empty tag value",
			input:       `cpu.foo.bar;tag1= 50 ` + strTime,
			template:    "measurement.foo.bar",
			measurement: "cpu",
			tags: map[string]string{
				"foo": "foo",
				"bar": "bar",
			},
			value: 50,
			time:  testTime,
		},
		{
			test:        "metric only with float value",
			input:       `cpu 50.554 ` + strTime,
			measurement: "cpu",
			template:    "measurement",
			value:       50.554,
			time:        testTime,
		},
		{
			test:     "missing metric",
			input:    `1419972457825`,
			template: "measurement",
			err:      `received "1419972457825" which doesn't have required fields`,
		},
		{
			test:     "should error parsing invalid float",
			input:    `cpu 50.554z 1419972457825`,
			template: "measurement",
			err:      `field "cpu" value: strconv.ParseFloat: parsing "50.554z": invalid syntax`,
		},
		{
			test:     "should error parsing invalid int",
			input:    `cpu 50z 1419972457825`,
			template: "measurement",
			err:      `field "cpu" value: strconv.ParseFloat: parsing "50z": invalid syntax`,
		},
		{
			test:     "should error parsing invalid time",
			input:    `cpu 50.554 14199724z57825`,
			template: "measurement",
			err:      `field "cpu" time: strconv.ParseFloat: parsing "14199724z57825": invalid syntax`,
		},
		{
			test:     "measurement* and field* (invalid)",
			input:    `prod.us-west.server01.cpu.util.idle.percent 99.99 1419972457825`,
			template: "env.zone.host.measurement*.field*",
			err:      `either 'field*' or 'measurement*' can be used in each template (but not both together): "env.zone.host.measurement*.field*"`,
		},
	}

	for _, test := range tests {
		p, err := NewGraphiteParser("", []string{test.template}, nil)
		if err != nil {
			t.Fatalf("unexpected error creating graphite parser: %v", err)
		}

		m, err := p.ParseLine(test.input)
		if errstr(err) != test.err {
			t.Fatalf("err does not match.  expected %v, got %v", test.err, err)
		}
		if err != nil {
			// If we erred out,it was intended and the following tests won't work
			continue
		}
		if m.Name() != test.measurement {
			t.Fatalf("name parse failer.  expected %v, got %v",
				test.measurement, m.Name())
		}
		if len(m.Tags()) != len(test.tags) {
			t.Fatalf("tags len mismatch.  expected %d, got %d",
				len(test.tags), len(m.Tags()))
		}
		f := m.Fields()["value"].(float64)
		if f != test.value {
			t.Fatalf("floatValue value mismatch.  expected %v, got %v",
				test.value, f)
		}
		if m.Time().UnixNano()/1000000 != test.time.UnixNano()/1000000 {
			t.Fatalf("time value mismatch.  expected %v, got %v",
				test.time.UnixNano(), m.Time().UnixNano())
		}
	}
}

func TestParse(t *testing.T) {
	testTime := time.Now().Round(time.Second)
	epochTime := testTime.Unix()
	strTime := strconv.FormatInt(epochTime, 10)

	var tests = []struct {
		test        string
		input       []byte
		measurement string
		tags        map[string]string
		value       float64
		time        time.Time
		template    string
		err         string
	}{
		{
			test:        "normal case",
			input:       []byte(`cpu.foo.bar 50 ` + strTime),
			template:    "measurement.foo.bar",
			measurement: "cpu",
			tags: map[string]string{
				"foo": "foo",
				"bar": "bar",
			},
			value: 50,
			time:  testTime,
		},
		{
			test:        "normal case with tag",
			input:       []byte(`cpu.foo.bar;tag1=value1 50 ` + strTime),
			template:    "measurement.foo.bar",
			measurement: "cpu",
			tags: map[string]string{
				"foo":  "foo",
				"bar":  "bar",
				"tag1": "value1",
			},
			value: 50,
			time:  testTime,
		},

		{
			test:        "metric only with float value",
			input:       []byte(`cpu 50.554 ` + strTime),
			measurement: "cpu",
			template:    "measurement",
			value:       50.554,
			time:        testTime,
		},
		{
			test:     "missing metric",
			input:    []byte(`1419972457825`),
			template: "measurement",
			err:      `received "1419972457825" which doesn't have required fields`,
		},
		{
			test:     "should error parsing invalid float",
			input:    []byte(`cpu 50.554z 1419972457825`),
			template: "measurement",
			err:      `field "cpu" value: strconv.ParseFloat: parsing "50.554z": invalid syntax`,
		},
		{
			test:     "should error parsing invalid int",
			input:    []byte(`cpu 50z 1419972457825`),
			template: "measurement",
			err:      `field "cpu" value: strconv.ParseFloat: parsing "50z": invalid syntax`,
		},
		{
			test:     "should error parsing invalid time",
			input:    []byte(`cpu 50.554 14199724z57825`),
			template: "measurement",
			err:      `field "cpu" time: strconv.ParseFloat: parsing "14199724z57825": invalid syntax`,
		},
		{
			test:     "measurement* and field* (invalid)",
			input:    []byte(`prod.us-west.server01.cpu.util.idle.percent 99.99 1419972457825`),
			template: "env.zone.host.measurement*.field*",
			err:      `either 'field*' or 'measurement*' can be used in each template (but not both together): "env.zone.host.measurement*.field*"`,
		},
	}

	for _, test := range tests {
		p, err := NewGraphiteParser("", []string{test.template}, nil)
		if err != nil {
			t.Fatalf("unexpected error creating graphite parser: %v", err)
		}

		metrics, err := p.Parse(test.input)
		if errstr(err) != test.err {
			t.Fatalf("err does not match.  expected [%v], got [%v]", test.err, err)
		}
		if err != nil {
			// If we erred out,it was intended and the following tests won't work
			continue
		}
		if metrics[0].Name() != test.measurement {
			t.Fatalf("name parse failer.  expected %v, got %v",
				test.measurement, metrics[0].Name())
		}
		if len(metrics[0].Tags()) != len(test.tags) {
			t.Fatalf("tags len mismatch.  expected %d, got %d",
				len(test.tags), len(metrics[0].Tags()))
		}
		f := metrics[0].Fields()["value"].(float64)
		if metrics[0].Fields()["value"] != f {
			t.Fatalf("floatValue value mismatch.  expected %v, got %v",
				test.value, f)
		}
		if metrics[0].Time().UnixNano()/1000000 != test.time.UnixNano()/1000000 {
			t.Fatalf("time value mismatch.  expected %v, got %v",
				test.time.UnixNano(), metrics[0].Time().UnixNano())
		}
	}
}

func TestParseNaN(t *testing.T) {
	p, err := NewGraphiteParser("", []string{"measurement*"}, nil)
	require.NoError(t, err)

	m, err := p.ParseLine("servers.localhost.cpu_load NaN 1435077219")
	require.NoError(t, err)

	expected := testutil.MustMetric(
		"servers.localhost.cpu_load",
		map[string]string{},
		map[string]interface{}{
			"value": math.NaN(),
		},
		time.Unix(1435077219, 0),
	)

	testutil.RequireMetricEqual(t, expected, m)
}

func TestParseInf(t *testing.T) {
	p, err := NewGraphiteParser("", []string{"measurement*"}, nil)
	require.NoError(t, err)

	m, err := p.ParseLine("servers.localhost.cpu_load +Inf 1435077219")
	require.NoError(t, err)

	expected := testutil.MustMetric(
		"servers.localhost.cpu_load",
		map[string]string{},
		map[string]interface{}{
			"value": math.Inf(1),
		},
		time.Unix(1435077219, 0),
	)

	testutil.RequireMetricEqual(t, expected, m)
}

func TestFilterMatchDefault(t *testing.T) {
	p, err := NewGraphiteParser("", []string{"servers.localhost .host.measurement*"}, nil)
	if err != nil {
		t.Fatalf("unexpected error creating parser, got %v", err)
	}

	exp := metric.New("miss.servers.localhost.cpu_load",
		map[string]string{},
		map[string]interface{}{"value": float64(11)},
		time.Unix(1435077219, 0))

	m, err := p.ParseLine("miss.servers.localhost.cpu_load 11 1435077219")
	require.NoError(t, err)

	require.Equal(t, exp, m)
}

func TestFilterMatchMultipleMeasurement(t *testing.T) {
	p, err := NewGraphiteParser("", []string{"servers.localhost .host.measurement.measurement*"}, nil)
	if err != nil {
		t.Fatalf("unexpected error creating parser, got %v", err)
	}

	exp := metric.New("cpu.cpu_load.10",
		map[string]string{"host": "localhost"},
		map[string]interface{}{"value": float64(11)},
		time.Unix(1435077219, 0))

	m, err := p.ParseLine("servers.localhost.cpu.cpu_load.10 11 1435077219")
	require.NoError(t, err)

	require.Equal(t, exp, m)
}

func TestFilterMatchMultipleMeasurementSeparator(t *testing.T) {
	p, err := NewGraphiteParser("_",
		[]string{"servers.localhost .host.measurement.measurement*"},
		nil,
	)
	require.NoError(t, err)

	exp := metric.New("cpu_cpu_load_10",
		map[string]string{"host": "localhost"},
		map[string]interface{}{"value": float64(11)},
		time.Unix(1435077219, 0))

	m, err := p.ParseLine("servers.localhost.cpu.cpu_load.10 11 1435077219")
	require.NoError(t, err)

	require.Equal(t, exp, m)
}

func TestFilterMatchSingle(t *testing.T) {
	p, err := NewGraphiteParser("", []string{"servers.localhost .host.measurement*"}, nil)
	if err != nil {
		t.Fatalf("unexpected error creating parser, got %v", err)
	}

	exp := metric.New("cpu_load",
		map[string]string{"host": "localhost"},
		map[string]interface{}{"value": float64(11)},
		time.Unix(1435077219, 0))

	m, err := p.ParseLine("servers.localhost.cpu_load 11 1435077219")
	require.NoError(t, err)

	require.Equal(t, exp, m)
}

func TestParseNoMatch(t *testing.T) {
	p, err := NewGraphiteParser("", []string{"servers.*.cpu .host.measurement.cpu.measurement"}, nil)
	if err != nil {
		t.Fatalf("unexpected error creating parser, got %v", err)
	}

	exp := metric.New("servers.localhost.memory.VmallocChunk",
		map[string]string{},
		map[string]interface{}{"value": float64(11)},
		time.Unix(1435077219, 0))

	m, err := p.ParseLine("servers.localhost.memory.VmallocChunk 11 1435077219")
	require.NoError(t, err)

	require.Equal(t, exp, m)
}

func TestFilterMatchWildcard(t *testing.T) {
	p, err := NewGraphiteParser("", []string{"servers.* .host.measurement*"}, nil)
	if err != nil {
		t.Fatalf("unexpected error creating parser, got %v", err)
	}

	exp := metric.New("cpu_load",
		map[string]string{"host": "localhost"},
		map[string]interface{}{"value": float64(11)},
		time.Unix(1435077219, 0))

	m, err := p.ParseLine("servers.localhost.cpu_load 11 1435077219")
	require.NoError(t, err)

	require.Equal(t, exp, m)
}

func TestFilterMatchExactBeforeWildcard(t *testing.T) {
	p, err := NewGraphiteParser("", []string{
		"servers.* .wrong.measurement*",
		"servers.localhost .host.measurement*"}, nil)
	if err != nil {
		t.Fatalf("unexpected error creating parser, got %v", err)
	}

	exp := metric.New("cpu_load",
		map[string]string{"host": "localhost"},
		map[string]interface{}{"value": float64(11)},
		time.Unix(1435077219, 0))

	m, err := p.ParseLine("servers.localhost.cpu_load 11 1435077219")
	require.NoError(t, err)

	require.Equal(t, exp, m)
}

func TestFilterMatchMostLongestFilter(t *testing.T) {
	p, err := NewGraphiteParser("", []string{
		"*.* .wrong.measurement*",
		"servers.* .wrong.measurement*",
		"servers.localhost .wrong.measurement*",
		"servers.localhost.cpu .host.resource.measurement*", // should match this
		"*.localhost .wrong.measurement*",
	}, nil)

	if err != nil {
		t.Fatalf("unexpected error creating parser, got %v", err)
	}

	m, err := p.ParseLine("servers.localhost.cpu.cpu_load 11 1435077219")
	require.NoError(t, err)

	value, ok := m.GetTag("host")
	require.True(t, ok)
	require.Equal(t, "localhost", value)

	value, ok = m.GetTag("resource")
	require.True(t, ok)
	require.Equal(t, "cpu", value)
}

func TestFilterMatchMultipleWildcards(t *testing.T) {
	p, err := NewGraphiteParser("", []string{
		"*.* .wrong.measurement*",
		"servers.* .host.measurement*", // should match this
		"servers.localhost .wrong.measurement*",
		"*.localhost .wrong.measurement*",
	}, nil)

	if err != nil {
		t.Fatalf("unexpected error creating parser, got %v", err)
	}

	exp := metric.New("cpu_load",
		map[string]string{"host": "server01"},
		map[string]interface{}{"value": float64(11)},
		time.Unix(1435077219, 0))

	m, err := p.ParseLine("servers.server01.cpu_load 11 1435077219")
	require.NoError(t, err)

	require.Equal(t, exp, m)
}

func TestParseDefaultTags(t *testing.T) {
	p, err := NewGraphiteParser("", []string{"servers.localhost .host.measurement*"}, map[string]string{
		"region": "us-east",
		"zone":   "1c",
		"host":   "should not set",
	})
	if err != nil {
		t.Fatalf("unexpected error creating parser, got %v", err)
	}

	m, err := p.ParseLine("servers.localhost.cpu_load 11 1435077219")
	require.NoError(t, err)

	value, ok := m.GetTag("host")
	require.True(t, ok)
	require.Equal(t, "localhost", value)

	value, ok = m.GetTag("region")
	require.True(t, ok)
	require.Equal(t, "us-east", value)

	value, ok = m.GetTag("zone")
	require.True(t, ok)
	require.Equal(t, "1c", value)
}

func TestParseDefaultTemplateTags(t *testing.T) {
	p, err := NewGraphiteParser("", []string{"servers.localhost .host.measurement* zone=1c"}, map[string]string{
		"region": "us-east",
		"host":   "should not set",
	})
	if err != nil {
		t.Fatalf("unexpected error creating parser, got %v", err)
	}

	m, err := p.ParseLine("servers.localhost.cpu_load 11 1435077219")
	require.NoError(t, err)

	value, ok := m.GetTag("host")
	require.True(t, ok)
	require.Equal(t, "localhost", value)

	value, ok = m.GetTag("region")
	require.True(t, ok)
	require.Equal(t, "us-east", value)

	value, ok = m.GetTag("zone")
	require.True(t, ok)
	require.Equal(t, "1c", value)
}

func TestParseDefaultTemplateTagsOverridGlobal(t *testing.T) {
	p, err := NewGraphiteParser("", []string{"servers.localhost .host.measurement* zone=1c,region=us-east"}, map[string]string{
		"region": "shot not be set",
		"host":   "should not set",
	})
	if err != nil {
		t.Fatalf("unexpected error creating parser, got %v", err)
	}

	m, err := p.ParseLine("servers.localhost.cpu_load 11 1435077219")
	_ = m
	require.NoError(t, err)

	value, ok := m.GetTag("host")
	require.True(t, ok)
	require.Equal(t, "localhost", value)

	value, ok = m.GetTag("region")
	require.True(t, ok)
	require.Equal(t, "us-east", value)

	value, ok = m.GetTag("zone")
	require.True(t, ok)
	require.Equal(t, "1c", value)
}

func TestParseTemplateWhitespace(t *testing.T) {
	p, err := NewGraphiteParser("",
		[]string{"servers.localhost        .host.measurement*           zone=1c"},
		map[string]string{
			"region": "us-east",
			"host":   "should not set",
		})
	if err != nil {
		t.Fatalf("unexpected error creating parser, got %v", err)
	}

	m, err := p.ParseLine("servers.localhost.cpu_load 11 1435077219")
	require.NoError(t, err)

	value, ok := m.GetTag("host")
	require.True(t, ok)
	require.Equal(t, "localhost", value)

	value, ok = m.GetTag("region")
	require.True(t, ok)
	require.Equal(t, "us-east", value)

	value, ok = m.GetTag("zone")
	require.True(t, ok)
	require.Equal(t, "1c", value)
}

// Test basic functionality of ApplyTemplate
func TestApplyTemplate(t *testing.T) {
	p, err := NewGraphiteParser("_",
		[]string{"current.* measurement.measurement"},
		nil)
	require.NoError(t, err)

	measurement, _, _, err := p.ApplyTemplate("current.users")
	require.NoError(t, err)
	require.Equal(t, "current_users", measurement)
}

// Test basic functionality of ApplyTemplate
func TestApplyTemplateNoMatch(t *testing.T) {
	p, err := NewGraphiteParser(".",
		[]string{"foo.bar measurement.measurement"},
		nil)
	require.NoError(t, err)

	measurement, _, _, err := p.ApplyTemplate("current.users")
	require.NoError(t, err)
	require.Equal(t, "current.users", measurement)
}

// Test that most specific template is chosen
func TestApplyTemplateSpecific(t *testing.T) {
	p, err := NewGraphiteParser("_",
		[]string{
			"current.* measurement.measurement",
			"current.*.* measurement.measurement.service",
		}, nil)
	require.NoError(t, err)

	measurement, tags, _, _ := p.ApplyTemplate("current.users.facebook")
	require.Equal(t, "current_users", measurement)

	service, ok := tags["service"]
	if !ok {
		t.Error("Expected for template to apply a 'service' tag, but not found")
	}
	if service != "facebook" {
		t.Errorf("Expected service='facebook' tag, got service='%s'", service)
	}
}

func TestApplyTemplateTags(t *testing.T) {
	p, err := NewGraphiteParser("_",
		[]string{"current.* measurement.measurement region=us-west"}, nil)
	require.NoError(t, err)

	measurement, tags, _, _ := p.ApplyTemplate("current.users")
	require.Equal(t, "current_users", measurement)

	region, ok := tags["region"]
	if !ok {
		t.Error("Expected for template to apply a 'region' tag, but not found")
	}
	if region != "us-west" {
		t.Errorf("Expected region='us-west' tag, got region='%s'", region)
	}
}

func TestApplyTemplateField(t *testing.T) {
	p, err := NewGraphiteParser("_",
		[]string{"current.* measurement.measurement.field"}, nil)
	require.NoError(t, err)

	measurement, _, field, err := p.ApplyTemplate("current.users.logged_in")
	require.NoError(t, err)
	require.Equal(t, "current_users", measurement)

	if field != "logged_in" {
		t.Errorf("Parser.ApplyTemplate unexpected result. got %s, exp %s",
			field, "logged_in")
	}
}

func TestApplyTemplateMultipleFieldsTogether(t *testing.T) {
	p, err := NewGraphiteParser("_",
		[]string{"current.* measurement.measurement.field.field"}, nil)
	require.NoError(t, err)

	measurement, _, field, err := p.ApplyTemplate("current.users.logged_in.ssh")
	require.NoError(t, err)
	require.Equal(t, "current_users", measurement)

	if field != "logged_in_ssh" {
		t.Errorf("Parser.ApplyTemplate unexpected result. got %s, exp %s",
			field, "logged_in_ssh")
	}
}

func TestApplyTemplateMultipleFieldsApart(t *testing.T) {
	p, err := NewGraphiteParser("_",
		[]string{"current.* measurement.measurement.field.method.field"}, nil)
	require.NoError(t, err)

	measurement, _, field, err := p.ApplyTemplate("current.users.logged_in.ssh.total")
	require.NoError(t, err)
	require.Equal(t, "current_users", measurement)

	if field != "logged_in_total" {
		t.Errorf("Parser.ApplyTemplate unexpected result. got %s, exp %s",
			field, "logged_in_total")
	}
}

func TestApplyTemplateGreedyField(t *testing.T) {
	p, err := NewGraphiteParser("_",
		[]string{"current.* measurement.measurement.field*"}, nil)
	require.NoError(t, err)

	measurement, _, field, err := p.ApplyTemplate("current.users.logged_in")
	require.NoError(t, err)
	require.Equal(t, "current_users", measurement)

	if field != "logged_in" {
		t.Errorf("Parser.ApplyTemplate unexpected result. got %s, exp %s",
			field, "logged_in")
	}
}

func TestApplyTemplateOverSpecific(t *testing.T) {
	p, err := NewGraphiteParser(
		".",
		[]string{
			"measurement.host.metric.metric.metric",
		},
		nil,
	)
	require.NoError(t, err)

	measurement, tags, _, err := p.ApplyTemplate("net.server001.a.b 2")
	require.NoError(t, err)
	require.Equal(t, "net", measurement)
	require.Equal(t,
		map[string]string{"host": "server001", "metric": "a.b"},
		tags)
}

func TestApplyTemplateMostSpecificTemplate(t *testing.T) {
	p, err := NewGraphiteParser(
		".",
		[]string{
			"measurement.host.metric",
			"measurement.host.metric.metric.metric",
			"measurement.host.metric.metric",
		},
		nil,
	)
	require.NoError(t, err)

	measurement, tags, _, err := p.ApplyTemplate("net.server001.a.b.c 2")
	require.NoError(t, err)
	require.Equal(t, "net", measurement)
	require.Equal(t,
		map[string]string{"host": "server001", "metric": "a.b.c"},
		tags)

	measurement, tags, _, err = p.ApplyTemplate("net.server001.a.b 2")
	require.NoError(t, err)
	require.Equal(t, "net", measurement)
	require.Equal(t,
		map[string]string{"host": "server001", "metric": "a.b"},
		tags)
}

// Test Helpers
func errstr(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
