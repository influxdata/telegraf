package grok

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var benchM telegraf.Metric

func Benchmark_ParseLine_CommonLogFormat(b *testing.B) {
	p := &Parser{
		Patterns: []string{"%{COMMON_LOG_FORMAT}"},
	}
	_ = p.Compile()

	var m telegraf.Metric
	for n := 0; n < b.N; n++ {
		m, _ = p.ParseLine(`127.0.0.1 user-identifier frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`)
	}
	benchM = m
}

func Benchmark_ParseLine_CombinedLogFormat(b *testing.B) {
	p := &Parser{
		Patterns: []string{"%{COMBINED_LOG_FORMAT}"},
	}
	_ = p.Compile()

	var m telegraf.Metric
	for n := 0; n < b.N; n++ {
		m, _ = p.ParseLine(`127.0.0.1 user-identifier frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326 "-" "Mozilla"`)
	}
	benchM = m
}

func Benchmark_ParseLine_CustomPattern(b *testing.B) {
	p := &Parser{
		Patterns: []string{"%{TEST_LOG_A}", "%{TEST_LOG_B}"},
		CustomPatterns: `
			DURATION %{NUMBER}[nuµm]?s
			RESPONSE_CODE %{NUMBER:response_code:tag}
			RESPONSE_TIME %{DURATION:response_time:duration}
			TEST_LOG_A %{NUMBER:myfloat:float} %{RESPONSE_CODE} %{IPORHOST:clientip} %{RESPONSE_TIME}
		`,
	}
	_ = p.Compile()

	var m telegraf.Metric
	for n := 0; n < b.N; n++ {
		m, _ = p.ParseLine(`[04/Jun/2016:12:41:45 +0100] 1.25 200 192.168.1.1 5.432µs 101`)
	}
	benchM = m
}

// Test a very simple parse pattern.
func TestSimpleParse(t *testing.T) {
	p := &Parser{
		Patterns: []string{"%{TESTLOG}"},
		CustomPatterns: `
			TESTLOG %{NUMBER:num:int} %{WORD:client}
		`,
	}
	assert.NoError(t, p.Compile())

	m, err := p.ParseLine(`142 bot`)
	assert.NoError(t, err)
	require.NotNil(t, m)

	assert.Equal(t,
		map[string]interface{}{
			"num":    int64(142),
			"client": "bot",
		},
		m.Fields())
}

// Verify that patterns with a regex lookahead fail at compile time.
func TestParsePatternsWithLookahead(t *testing.T) {
	p := &Parser{
		Patterns: []string{"%{MYLOG}"},
		CustomPatterns: `
			NOBOT ((?!bot|crawl).)*
			MYLOG %{NUMBER:num:int} %{NOBOT:client}
		`,
	}
	assert.NoError(t, p.Compile())

	_, err := p.ParseLine(`1466004605359052000 bot`)
	assert.Error(t, err)
}

func TestMeasurementName(t *testing.T) {
	p := &Parser{
		Measurement: "my_web_log",
		Patterns:    []string{"%{COMMON_LOG_FORMAT}"},
	}
	assert.NoError(t, p.Compile())

	// Parse an influxdb POST request
	m, err := p.ParseLine(`127.0.0.1 user-identifier frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`)
	require.NotNil(t, m)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"resp_bytes":   int64(2326),
			"auth":         "frank",
			"client_ip":    "127.0.0.1",
			"http_version": float64(1.0),
			"ident":        "user-identifier",
			"request":      "/apache_pb.gif",
		},
		m.Fields())
	assert.Equal(t, map[string]string{"verb": "GET", "resp_code": "200"}, m.Tags())
	assert.Equal(t, "my_web_log", m.Name())
}

func TestCLF_IPv6(t *testing.T) {
	p := &Parser{
		Measurement: "my_web_log",
		Patterns:    []string{"%{COMMON_LOG_FORMAT}"},
	}
	assert.NoError(t, p.Compile())

	m, err := p.ParseLine(`2001:0db8:85a3:0000:0000:8a2e:0370:7334 user-identifier frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`)
	require.NotNil(t, m)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"resp_bytes":   int64(2326),
			"auth":         "frank",
			"client_ip":    "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			"http_version": float64(1.0),
			"ident":        "user-identifier",
			"request":      "/apache_pb.gif",
		},
		m.Fields())
	assert.Equal(t, map[string]string{"verb": "GET", "resp_code": "200"}, m.Tags())
	assert.Equal(t, "my_web_log", m.Name())

	m, err = p.ParseLine(`::1 user-identifier frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`)
	require.NotNil(t, m)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"resp_bytes":   int64(2326),
			"auth":         "frank",
			"client_ip":    "::1",
			"http_version": float64(1.0),
			"ident":        "user-identifier",
			"request":      "/apache_pb.gif",
		},
		m.Fields())
	assert.Equal(t, map[string]string{"verb": "GET", "resp_code": "200"}, m.Tags())
	assert.Equal(t, "my_web_log", m.Name())
}

func TestCustomInfluxdbHttpd(t *testing.T) {
	p := &Parser{
		Patterns: []string{`\[httpd\] %{COMBINED_LOG_FORMAT} %{UUID:uuid:drop} %{NUMBER:response_time_us:int}`},
	}
	assert.NoError(t, p.Compile())

	// Parse an influxdb POST request
	m, err := p.ParseLine(`[httpd] ::1 - - [14/Jun/2016:11:33:29 +0100] "POST /write?consistency=any&db=telegraf&precision=ns&rp= HTTP/1.1" 204 0 "-" "InfluxDBClient" 6f61bc44-321b-11e6-8050-000000000000 2513`)
	require.NotNil(t, m)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"resp_bytes":       int64(0),
			"auth":             "-",
			"client_ip":        "::1",
			"http_version":     float64(1.1),
			"ident":            "-",
			"referrer":         "-",
			"request":          "/write?consistency=any&db=telegraf&precision=ns&rp=",
			"response_time_us": int64(2513),
			"agent":            "InfluxDBClient",
		},
		m.Fields())
	assert.Equal(t, map[string]string{"verb": "POST", "resp_code": "204"}, m.Tags())

	// Parse an influxdb GET request
	m, err = p.ParseLine(`[httpd] ::1 - - [14/Jun/2016:12:10:02 +0100] "GET /query?db=telegraf&q=SELECT+bytes%2Cresponse_time_us+FROM+logparser_grok+WHERE+http_method+%3D+%27GET%27+AND+response_time_us+%3E+0+AND+time+%3E+now%28%29+-+1h HTTP/1.1" 200 578 "http://localhost:8083/" "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.84 Safari/537.36" 8a3806f1-3220-11e6-8006-000000000000 988`)
	require.NotNil(t, m)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"resp_bytes":       int64(578),
			"auth":             "-",
			"client_ip":        "::1",
			"http_version":     float64(1.1),
			"ident":            "-",
			"referrer":         "http://localhost:8083/",
			"request":          "/query?db=telegraf&q=SELECT+bytes%2Cresponse_time_us+FROM+logparser_grok+WHERE+http_method+%3D+%27GET%27+AND+response_time_us+%3E+0+AND+time+%3E+now%28%29+-+1h",
			"response_time_us": int64(988),
			"agent":            "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.84 Safari/537.36",
		},
		m.Fields())
	assert.Equal(t, map[string]string{"verb": "GET", "resp_code": "200"}, m.Tags())
}

// common log format
// 127.0.0.1 user-identifier frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326
func TestBuiltinCommonLogFormat(t *testing.T) {
	p := &Parser{
		Patterns: []string{"%{COMMON_LOG_FORMAT}"},
	}
	assert.NoError(t, p.Compile())

	// Parse an influxdb POST request
	m, err := p.ParseLine(`127.0.0.1 user-identifier frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`)
	require.NotNil(t, m)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"resp_bytes":   int64(2326),
			"auth":         "frank",
			"client_ip":    "127.0.0.1",
			"http_version": float64(1.0),
			"ident":        "user-identifier",
			"request":      "/apache_pb.gif",
		},
		m.Fields())
	assert.Equal(t, map[string]string{"verb": "GET", "resp_code": "200"}, m.Tags())
}

// common log format
// 127.0.0.1 user1234 frank1234 [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326
func TestBuiltinCommonLogFormatWithNumbers(t *testing.T) {
	p := &Parser{
		Patterns: []string{"%{COMMON_LOG_FORMAT}"},
	}
	assert.NoError(t, p.Compile())

	// Parse an influxdb POST request
	m, err := p.ParseLine(`127.0.0.1 user1234 frank1234 [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`)
	require.NotNil(t, m)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"resp_bytes":   int64(2326),
			"auth":         "frank1234",
			"client_ip":    "127.0.0.1",
			"http_version": float64(1.0),
			"ident":        "user1234",
			"request":      "/apache_pb.gif",
		},
		m.Fields())
	assert.Equal(t, map[string]string{"verb": "GET", "resp_code": "200"}, m.Tags())
}

// combined log format
// 127.0.0.1 user-identifier frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326 "-" "Mozilla"
func TestBuiltinCombinedLogFormat(t *testing.T) {
	p := &Parser{
		Patterns: []string{"%{COMBINED_LOG_FORMAT}"},
	}
	assert.NoError(t, p.Compile())

	// Parse an influxdb POST request
	m, err := p.ParseLine(`127.0.0.1 user-identifier frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326 "-" "Mozilla"`)
	require.NotNil(t, m)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"resp_bytes":   int64(2326),
			"auth":         "frank",
			"client_ip":    "127.0.0.1",
			"http_version": float64(1.0),
			"ident":        "user-identifier",
			"request":      "/apache_pb.gif",
			"referrer":     "-",
			"agent":        "Mozilla",
		},
		m.Fields())
	assert.Equal(t, map[string]string{"verb": "GET", "resp_code": "200"}, m.Tags())
}

func TestCompileStringAndParse(t *testing.T) {
	p := &Parser{
		Patterns: []string{"%{TEST_LOG_A}"},
		CustomPatterns: `
			DURATION %{NUMBER}[nuµm]?s
			RESPONSE_CODE %{NUMBER:response_code:tag}
			RESPONSE_TIME %{DURATION:response_time:duration}
			TEST_LOG_A %{NUMBER:myfloat:float} %{RESPONSE_CODE} %{IPORHOST:clientip} %{RESPONSE_TIME}
		`,
	}
	assert.NoError(t, p.Compile())

	metricA, err := p.ParseLine(`1.25 200 192.168.1.1 5.432µs`)
	require.NotNil(t, metricA)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"clientip":      "192.168.1.1",
			"myfloat":       float64(1.25),
			"response_time": int64(5432),
		},
		metricA.Fields())
	assert.Equal(t, map[string]string{"response_code": "200"}, metricA.Tags())
}

func TestCompileErrorsOnInvalidPattern(t *testing.T) {
	p := &Parser{
		Patterns: []string{"%{TEST_LOG_A}", "%{TEST_LOG_B}"},
		CustomPatterns: `
			DURATION %{NUMBER}[nuµm]?s
			RESPONSE_CODE %{NUMBER:response_code:tag}
			RESPONSE_TIME %{DURATION:response_time:duration}
			TEST_LOG_A %{NUMBER:myfloat:float} %{RESPONSE_CODE} %{IPORHOST:clientip} %{RESPONSE_TIME}
		`,
	}
	assert.Error(t, p.Compile())

	metricA, _ := p.ParseLine(`1.25 200 192.168.1.1 5.432µs`)
	require.Nil(t, metricA)
}

func TestParsePatternsWithoutCustom(t *testing.T) {
	p := &Parser{
		Patterns: []string{"%{POSINT:ts:ts-epochnano} response_time=%{POSINT:response_time:int} mymetric=%{NUMBER:metric:float}"},
	}
	assert.NoError(t, p.Compile())

	metricA, err := p.ParseLine(`1466004605359052000 response_time=20821 mymetric=10890.645`)
	require.NotNil(t, metricA)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"response_time": int64(20821),
			"metric":        float64(10890.645),
		},
		metricA.Fields())
	assert.Equal(t, map[string]string{}, metricA.Tags())
	assert.Equal(t, time.Unix(0, 1466004605359052000), metricA.Time())
}

func TestParseEpochNano(t *testing.T) {
	p := &Parser{
		Patterns: []string{"%{MYAPP}"},
		CustomPatterns: `
			MYAPP %{POSINT:ts:ts-epochnano} response_time=%{POSINT:response_time:int} mymetric=%{NUMBER:metric:float}
		`,
	}
	assert.NoError(t, p.Compile())

	metricA, err := p.ParseLine(`1466004605359052000 response_time=20821 mymetric=10890.645`)
	require.NotNil(t, metricA)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"response_time": int64(20821),
			"metric":        float64(10890.645),
		},
		metricA.Fields())
	assert.Equal(t, map[string]string{}, metricA.Tags())
	assert.Equal(t, time.Unix(0, 1466004605359052000), metricA.Time())
}

func TestParseEpoch(t *testing.T) {
	p := &Parser{
		Patterns: []string{"%{MYAPP}"},
		CustomPatterns: `
			MYAPP %{POSINT:ts:ts-epoch} response_time=%{POSINT:response_time:int} mymetric=%{NUMBER:metric:float}
		`,
	}
	assert.NoError(t, p.Compile())

	metricA, err := p.ParseLine(`1466004605 response_time=20821 mymetric=10890.645`)
	require.NotNil(t, metricA)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"response_time": int64(20821),
			"metric":        float64(10890.645),
		},
		metricA.Fields())
	assert.Equal(t, map[string]string{}, metricA.Tags())
	assert.Equal(t, time.Unix(1466004605, 0), metricA.Time())
}

func TestParseEpochDecimal(t *testing.T) {
	var tests = []struct {
		name    string
		line    string
		noMatch bool
		err     error
		tags    map[string]string
		fields  map[string]interface{}
		time    time.Time
	}{
		{
			name: "ns precision",
			line: "1466004605.359052000 value=42",
			tags: map[string]string{},
			fields: map[string]interface{}{
				"value": int64(42),
			},
			time: time.Unix(0, 1466004605359052000),
		},
		{
			name: "ms precision",
			line: "1466004605.359 value=42",
			tags: map[string]string{},
			fields: map[string]interface{}{
				"value": int64(42),
			},
			time: time.Unix(0, 1466004605359000000),
		},
		{
			name: "second precision",
			line: "1466004605 value=42",
			tags: map[string]string{},
			fields: map[string]interface{}{
				"value": int64(42),
			},
			time: time.Unix(0, 1466004605000000000),
		},
		{
			name: "sub ns precision",
			line: "1466004605.123456789123 value=42",
			tags: map[string]string{},
			fields: map[string]interface{}{
				"value": int64(42),
			},
			time: time.Unix(0, 1466004605123456789),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{
				Patterns: []string{"%{NUMBER:ts:ts-epoch} value=%{NUMBER:value:int}"},
			}
			assert.NoError(t, parser.Compile())
			m, err := parser.ParseLine(tt.line)

			if tt.noMatch {
				require.Nil(t, m)
				require.Nil(t, err)
				return
			}

			require.Equal(t, tt.err, err)

			require.NotNil(t, m)
			require.Equal(t, tt.tags, m.Tags())
			require.Equal(t, tt.fields, m.Fields())
			require.Equal(t, tt.time, m.Time())
		})
	}
}

func TestParseEpochErrors(t *testing.T) {
	p := &Parser{
		Patterns: []string{"%{MYAPP}"},
		CustomPatterns: `
			MYAPP %{WORD:ts:ts-epoch} response_time=%{POSINT:response_time:int} mymetric=%{NUMBER:metric:float}
		`,
	}
	assert.NoError(t, p.Compile())

	_, err := p.ParseLine(`foobar response_time=20821 mymetric=10890.645`)
	assert.NoError(t, err)

	p = &Parser{
		Patterns: []string{"%{MYAPP}"},
		CustomPatterns: `
			MYAPP %{WORD:ts:ts-epochnano} response_time=%{POSINT:response_time:int} mymetric=%{NUMBER:metric:float}
		`,
	}
	assert.NoError(t, p.Compile())

	_, err = p.ParseLine(`foobar response_time=20821 mymetric=10890.645`)
	assert.NoError(t, err)
}

func TestParseGenericTimestamp(t *testing.T) {
	p := &Parser{
		Patterns: []string{`\[%{HTTPDATE:ts:ts}\] response_time=%{POSINT:response_time:int} mymetric=%{NUMBER:metric:float}`},
	}
	assert.NoError(t, p.Compile())

	metricA, err := p.ParseLine(`[09/Jun/2016:03:37:03 +0000] response_time=20821 mymetric=10890.645`)
	require.NotNil(t, metricA)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"response_time": int64(20821),
			"metric":        float64(10890.645),
		},
		metricA.Fields())
	assert.Equal(t, map[string]string{}, metricA.Tags())
	assert.Equal(t, time.Unix(1465443423, 0).UTC(), metricA.Time().UTC())

	metricB, err := p.ParseLine(`[09/Jun/2016:03:37:04 +0000] response_time=20821 mymetric=10890.645`)
	require.NotNil(t, metricB)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"response_time": int64(20821),
			"metric":        float64(10890.645),
		},
		metricB.Fields())
	assert.Equal(t, map[string]string{}, metricB.Tags())
	assert.Equal(t, time.Unix(1465443424, 0).UTC(), metricB.Time().UTC())
}

func TestParseGenericTimestampNotFound(t *testing.T) {
	p := &Parser{
		Patterns: []string{`\[%{NOTSPACE:ts:ts}\] response_time=%{POSINT:response_time:int} mymetric=%{NUMBER:metric:float}`},
	}
	assert.NoError(t, p.Compile())

	metricA, err := p.ParseLine(`[foobar] response_time=20821 mymetric=10890.645`)
	require.NotNil(t, metricA)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"response_time": int64(20821),
			"metric":        float64(10890.645),
		},
		metricA.Fields())
	assert.Equal(t, map[string]string{}, metricA.Tags())
}

func TestCompileFileAndParse(t *testing.T) {
	p := &Parser{
		Patterns:           []string{"%{TEST_LOG_A}", "%{TEST_LOG_B}"},
		CustomPatternFiles: []string{"./testdata/test-patterns"},
	}
	assert.NoError(t, p.Compile())

	metricA, err := p.ParseLine(`[04/Jun/2016:12:41:45 +0100] 1.25 200 192.168.1.1 5.432µs 101`)
	require.NotNil(t, metricA)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"clientip":      "192.168.1.1",
			"myfloat":       float64(1.25),
			"response_time": int64(5432),
			"myint":         int64(101),
		},
		metricA.Fields())
	assert.Equal(t, map[string]string{"response_code": "200"}, metricA.Tags())
	assert.Equal(t,
		time.Date(2016, time.June, 4, 12, 41, 45, 0, time.FixedZone("foo", 60*60)).Nanosecond(),
		metricA.Time().Nanosecond())

	metricB, err := p.ParseLine(`[04/06/2016--12:41:45] 1.25 mystring dropme nomodifier`)
	require.NotNil(t, metricB)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"myfloat":    1.25,
			"mystring":   "mystring",
			"nomodifier": "nomodifier",
		},
		metricB.Fields())
	assert.Equal(t, map[string]string{}, metricB.Tags())
	assert.Equal(t,
		time.Date(2016, time.June, 4, 12, 41, 45, 0, time.FixedZone("foo", 60*60)).Nanosecond(),
		metricB.Time().Nanosecond())
}

func TestCompileNoModifiersAndParse(t *testing.T) {
	p := &Parser{
		Patterns: []string{"%{TEST_LOG_C}"},
		CustomPatterns: `
			DURATION %{NUMBER}[nuµm]?s
			TEST_LOG_C %{NUMBER:myfloat} %{NUMBER} %{IPORHOST:clientip} %{DURATION:rt}
		`,
	}
	assert.NoError(t, p.Compile())

	metricA, err := p.ParseLine(`1.25 200 192.168.1.1 5.432µs`)
	require.NotNil(t, metricA)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"clientip": "192.168.1.1",
			"myfloat":  "1.25",
			"rt":       "5.432µs",
		},
		metricA.Fields())
	assert.Equal(t, map[string]string{}, metricA.Tags())
}

func TestCompileNoNamesAndParse(t *testing.T) {
	p := &Parser{
		Patterns: []string{"%{TEST_LOG_C}"},
		CustomPatterns: `
			DURATION %{NUMBER}[nuµm]?s
			TEST_LOG_C %{NUMBER} %{NUMBER} %{IPORHOST} %{DURATION}
		`,
	}
	assert.NoError(t, p.Compile())

	metricA, err := p.ParseLine(`1.25 200 192.168.1.1 5.432µs`)
	require.Nil(t, metricA)
	assert.NoError(t, err)
}

func TestParseNoMatch(t *testing.T) {
	p := &Parser{
		Patterns:           []string{"%{TEST_LOG_A}", "%{TEST_LOG_B}"},
		CustomPatternFiles: []string{"./testdata/test-patterns"},
	}
	assert.NoError(t, p.Compile())

	metricA, err := p.ParseLine(`[04/Jun/2016:12:41:45 +0100] notnumber 200 192.168.1.1 5.432µs 101`)
	assert.NoError(t, err)
	assert.Nil(t, metricA)
}

func TestCompileErrors(t *testing.T) {
	// Compile fails because there are multiple timestamps:
	p := &Parser{
		Patterns: []string{"%{TEST_LOG_A}", "%{TEST_LOG_B}"},
		CustomPatterns: `
			TEST_LOG_A %{HTTPDATE:ts1:ts-httpd} %{HTTPDATE:ts2:ts-httpd} %{NUMBER:mynum:int}
		`,
	}
	assert.Error(t, p.Compile())

	// Compile fails because file doesn't exist:
	p = &Parser{
		Patterns:           []string{"%{TEST_LOG_A}", "%{TEST_LOG_B}"},
		CustomPatternFiles: []string{"/tmp/foo/bar/baz"},
	}
	assert.Error(t, p.Compile())
}

func TestParseErrors(t *testing.T) {
	// Parse fails because the pattern doesn't exist
	p := &Parser{
		Patterns: []string{"%{TEST_LOG_B}"},
		CustomPatterns: `
			TEST_LOG_A %{HTTPDATE:ts:ts-httpd} %{WORD:myword:int} %{}
		`,
	}
	assert.Error(t, p.Compile())
	_, err := p.ParseLine(`[04/Jun/2016:12:41:45 +0100] notnumber 200 192.168.1.1 5.432µs 101`)
	assert.Error(t, err)

	// Parse fails because myword is not an int
	p = &Parser{
		Patterns: []string{"%{TEST_LOG_A}"},
		CustomPatterns: `
			TEST_LOG_A %{HTTPDATE:ts:ts-httpd} %{WORD:myword:int}
		`,
	}
	assert.NoError(t, p.Compile())
	_, err = p.ParseLine(`04/Jun/2016:12:41:45 +0100 notnumber`)
	assert.Error(t, err)

	// Parse fails because myword is not a float
	p = &Parser{
		Patterns: []string{"%{TEST_LOG_A}"},
		CustomPatterns: `
			TEST_LOG_A %{HTTPDATE:ts:ts-httpd} %{WORD:myword:float}
		`,
	}
	assert.NoError(t, p.Compile())
	_, err = p.ParseLine(`04/Jun/2016:12:41:45 +0100 notnumber`)
	assert.Error(t, err)

	// Parse fails because myword is not a duration
	p = &Parser{
		Patterns: []string{"%{TEST_LOG_A}"},
		CustomPatterns: `
			TEST_LOG_A %{HTTPDATE:ts:ts-httpd} %{WORD:myword:duration}
		`,
	}
	assert.NoError(t, p.Compile())
	_, err = p.ParseLine(`04/Jun/2016:12:41:45 +0100 notnumber`)
	assert.Error(t, err)

	// Parse fails because the time layout is wrong.
	p = &Parser{
		Patterns: []string{"%{TEST_LOG_A}"},
		CustomPatterns: `
			TEST_LOG_A %{HTTPDATE:ts:ts-unix} %{WORD:myword:duration}
		`,
	}
	assert.NoError(t, p.Compile())
	_, err = p.ParseLine(`04/Jun/2016:12:41:45 +0100 notnumber`)
	assert.Error(t, err)
}

func TestTsModder(t *testing.T) {
	tsm := &tsModder{}

	reftime := time.Date(2006, time.December, 1, 1, 1, 1, int(time.Millisecond), time.UTC)
	modt := tsm.tsMod(reftime)
	assert.Equal(t, reftime, modt)
	modt = tsm.tsMod(reftime)
	assert.Equal(t, reftime.Add(time.Microsecond*1), modt)
	modt = tsm.tsMod(reftime)
	assert.Equal(t, reftime.Add(time.Microsecond*2), modt)
	modt = tsm.tsMod(reftime)
	assert.Equal(t, reftime.Add(time.Microsecond*3), modt)

	reftime = time.Date(2006, time.December, 1, 1, 1, 1, int(time.Microsecond), time.UTC)
	modt = tsm.tsMod(reftime)
	assert.Equal(t, reftime, modt)
	modt = tsm.tsMod(reftime)
	assert.Equal(t, reftime.Add(time.Nanosecond*1), modt)
	modt = tsm.tsMod(reftime)
	assert.Equal(t, reftime.Add(time.Nanosecond*2), modt)
	modt = tsm.tsMod(reftime)
	assert.Equal(t, reftime.Add(time.Nanosecond*3), modt)

	reftime = time.Date(2006, time.December, 1, 1, 1, 1, int(time.Microsecond)*999, time.UTC)
	modt = tsm.tsMod(reftime)
	assert.Equal(t, reftime, modt)
	modt = tsm.tsMod(reftime)
	assert.Equal(t, reftime.Add(time.Nanosecond*1), modt)
	modt = tsm.tsMod(reftime)
	assert.Equal(t, reftime.Add(time.Nanosecond*2), modt)
	modt = tsm.tsMod(reftime)
	assert.Equal(t, reftime.Add(time.Nanosecond*3), modt)

	reftime = time.Date(2006, time.December, 1, 1, 1, 1, 0, time.UTC)
	modt = tsm.tsMod(reftime)
	assert.Equal(t, reftime, modt)
	modt = tsm.tsMod(reftime)
	assert.Equal(t, reftime.Add(time.Millisecond*1), modt)
	modt = tsm.tsMod(reftime)
	assert.Equal(t, reftime.Add(time.Millisecond*2), modt)
	modt = tsm.tsMod(reftime)
	assert.Equal(t, reftime.Add(time.Millisecond*3), modt)

	reftime = time.Time{}
	modt = tsm.tsMod(reftime)
	assert.Equal(t, reftime, modt)
}

func TestTsModder_Rollover(t *testing.T) {
	tsm := &tsModder{}

	reftime := time.Date(2006, time.December, 1, 1, 1, 1, int(time.Millisecond), time.UTC)
	modt := tsm.tsMod(reftime)
	for i := 1; i < 1000; i++ {
		modt = tsm.tsMod(reftime)
	}
	assert.Equal(t, reftime.Add(time.Microsecond*999+time.Nanosecond), modt)

	reftime = time.Date(2006, time.December, 1, 1, 1, 1, int(time.Microsecond), time.UTC)
	modt = tsm.tsMod(reftime)
	for i := 1; i < 1001; i++ {
		modt = tsm.tsMod(reftime)
	}
	assert.Equal(t, reftime.Add(time.Nanosecond*1000), modt)
}

func TestShortPatternRegression(t *testing.T) {
	p := &Parser{
		Patterns: []string{"%{TS_UNIX:timestamp:ts-unix} %{NUMBER:value:int}"},
		CustomPatterns: `
		  TS_UNIX %{DAY} %{MONTH} %{MONTHDAY} %{HOUR}:%{MINUTE}:%{SECOND} %{TZ} %{YEAR}
		`,
	}
	require.NoError(t, p.Compile())

	metric, err := p.ParseLine(`Wed Apr 12 13:10:34 PST 2017 42`)
	require.NoError(t, err)
	require.NotNil(t, metric)

	require.Equal(t,
		map[string]interface{}{
			"value": int64(42),
		},
		metric.Fields())
}

func TestTimezoneEmptyCompileFileAndParse(t *testing.T) {
	p := &Parser{
		Patterns:           []string{"%{TEST_LOG_A}", "%{TEST_LOG_B}"},
		CustomPatternFiles: []string{"./testdata/test-patterns"},
		Timezone:           "",
	}
	assert.NoError(t, p.Compile())

	metricA, err := p.ParseLine(`[04/Jun/2016:12:41:45 +0100] 1.25 200 192.168.1.1 5.432µs 101`)
	require.NotNil(t, metricA)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"clientip":      "192.168.1.1",
			"myfloat":       float64(1.25),
			"response_time": int64(5432),
			"myint":         int64(101),
		},
		metricA.Fields())
	assert.Equal(t, map[string]string{"response_code": "200"}, metricA.Tags())
	assert.Equal(t, int64(1465040505000000000), metricA.Time().UnixNano())

	metricB, err := p.ParseLine(`[04/06/2016--12:41:45] 1.25 mystring dropme nomodifier`)
	require.NotNil(t, metricB)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"myfloat":    1.25,
			"mystring":   "mystring",
			"nomodifier": "nomodifier",
		},
		metricB.Fields())
	assert.Equal(t, map[string]string{}, metricB.Tags())
	assert.Equal(t, int64(1465044105000000000), metricB.Time().UnixNano())
}

func TestTimezoneMalformedCompileFileAndParse(t *testing.T) {
	p := &Parser{
		Patterns:           []string{"%{TEST_LOG_A}", "%{TEST_LOG_B}"},
		CustomPatternFiles: []string{"./testdata/test-patterns"},
		Timezone:           "Something/Weird",
	}
	assert.NoError(t, p.Compile())

	metricA, err := p.ParseLine(`[04/Jun/2016:12:41:45 +0100] 1.25 200 192.168.1.1 5.432µs 101`)
	require.NotNil(t, metricA)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"clientip":      "192.168.1.1",
			"myfloat":       float64(1.25),
			"response_time": int64(5432),
			"myint":         int64(101),
		},
		metricA.Fields())
	assert.Equal(t, map[string]string{"response_code": "200"}, metricA.Tags())
	assert.Equal(t, int64(1465040505000000000), metricA.Time().UnixNano())

	metricB, err := p.ParseLine(`[04/06/2016--12:41:45] 1.25 mystring dropme nomodifier`)
	require.NotNil(t, metricB)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"myfloat":    1.25,
			"mystring":   "mystring",
			"nomodifier": "nomodifier",
		},
		metricB.Fields())
	assert.Equal(t, map[string]string{}, metricB.Tags())
	assert.Equal(t, int64(1465044105000000000), metricB.Time().UnixNano())
}

func TestTimezoneEuropeCompileFileAndParse(t *testing.T) {
	p := &Parser{
		Patterns:           []string{"%{TEST_LOG_A}", "%{TEST_LOG_B}"},
		CustomPatternFiles: []string{"./testdata/test-patterns"},
		Timezone:           "Europe/Berlin",
	}
	assert.NoError(t, p.Compile())

	metricA, err := p.ParseLine(`[04/Jun/2016:12:41:45 +0100] 1.25 200 192.168.1.1 5.432µs 101`)
	require.NotNil(t, metricA)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"clientip":      "192.168.1.1",
			"myfloat":       float64(1.25),
			"response_time": int64(5432),
			"myint":         int64(101),
		},
		metricA.Fields())
	assert.Equal(t, map[string]string{"response_code": "200"}, metricA.Tags())
	assert.Equal(t, int64(1465040505000000000), metricA.Time().UnixNano())

	metricB, err := p.ParseLine(`[04/06/2016--12:41:45] 1.25 mystring dropme nomodifier`)
	require.NotNil(t, metricB)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"myfloat":    1.25,
			"mystring":   "mystring",
			"nomodifier": "nomodifier",
		},
		metricB.Fields())
	assert.Equal(t, map[string]string{}, metricB.Tags())
	assert.Equal(t, int64(1465036905000000000), metricB.Time().UnixNano())
}

func TestTimezoneAmericasCompileFileAndParse(t *testing.T) {
	p := &Parser{
		Patterns:           []string{"%{TEST_LOG_A}", "%{TEST_LOG_B}"},
		CustomPatternFiles: []string{"./testdata/test-patterns"},
		Timezone:           "Canada/Eastern",
	}
	assert.NoError(t, p.Compile())

	metricA, err := p.ParseLine(`[04/Jun/2016:12:41:45 +0100] 1.25 200 192.168.1.1 5.432µs 101`)
	require.NotNil(t, metricA)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"clientip":      "192.168.1.1",
			"myfloat":       float64(1.25),
			"response_time": int64(5432),
			"myint":         int64(101),
		},
		metricA.Fields())
	assert.Equal(t, map[string]string{"response_code": "200"}, metricA.Tags())
	assert.Equal(t, int64(1465040505000000000), metricA.Time().UnixNano())

	metricB, err := p.ParseLine(`[04/06/2016--12:41:45] 1.25 mystring dropme nomodifier`)
	require.NotNil(t, metricB)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"myfloat":    1.25,
			"mystring":   "mystring",
			"nomodifier": "nomodifier",
		},
		metricB.Fields())
	assert.Equal(t, map[string]string{}, metricB.Tags())
	assert.Equal(t, int64(1465058505000000000), metricB.Time().UnixNano())
}

func TestTimezoneLocalCompileFileAndParse(t *testing.T) {
	p := &Parser{
		Patterns:           []string{"%{TEST_LOG_A}", "%{TEST_LOG_B}"},
		CustomPatternFiles: []string{"./testdata/test-patterns"},
		Timezone:           "Local",
	}
	assert.NoError(t, p.Compile())

	metricA, err := p.ParseLine(`[04/Jun/2016:12:41:45 +0100] 1.25 200 192.168.1.1 5.432µs 101`)
	require.NotNil(t, metricA)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"clientip":      "192.168.1.1",
			"myfloat":       float64(1.25),
			"response_time": int64(5432),
			"myint":         int64(101),
		},
		metricA.Fields())
	assert.Equal(t, map[string]string{"response_code": "200"}, metricA.Tags())
	assert.Equal(t, int64(1465040505000000000), metricA.Time().UnixNano())

	metricB, err := p.ParseLine(`[04/06/2016--12:41:45] 1.25 mystring dropme nomodifier`)
	require.NotNil(t, metricB)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]interface{}{
			"myfloat":    1.25,
			"mystring":   "mystring",
			"nomodifier": "nomodifier",
		},
		metricB.Fields())
	assert.Equal(t, map[string]string{}, metricB.Tags())
	assert.Equal(t, time.Date(2016, time.June, 4, 12, 41, 45, 0, time.Local).UnixNano(), metricB.Time().UnixNano())
}

func TestNewlineInPatterns(t *testing.T) {
	p := &Parser{
		Patterns: []string{`
			%{SYSLOGTIMESTAMP:timestamp}
		`},
	}
	require.NoError(t, p.Compile())
	m, err := p.ParseLine("Apr 10 05:11:57")
	require.NoError(t, err)
	require.NotNil(t, m)
}
