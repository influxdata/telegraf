package ltsv

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	validLTSV1 = "time:2016-03-06T09:24:12Z\tstr1:value1\tint1:23\tint2:34\tfloat1:1.23\tbool1:true\tbool2:false\tignore_field1:foo\ttag1:tval1\tignore_tag1:bar\ttag2:tval2"
)

var validLTSV2 = [][]byte{
	[]byte("time:2016-03-06T09:24:12.012+09:00\tstr1:value1\tint1:23\tint2:34\tfloat1:1.23\tbool1:true\tbool2:fal"),
	[]byte("se\tignore_field1:foo\ttag1:tval1\tignore_tag1:bar\ttag2:tval2\ntime:2016-03-06T09:24:12.125+09:00\ts"),
	// NOTE: validLTSV2[2] contains an empty line, and it is safely ignored.
	[]byte("tr1:value2\ntime:2016-03-06T09:24:13.000+09:00\tstr1:value3\n\ntime:2016-03-06T09:24:15.999+09:00\tst"),
	// NOTE: validLTSV2[3] does not end with a newline, so you need to call Parse(nil) to parse the rest of data.
	[]byte("r1:value4"),
	nil,
}

var validLTSV3 = []string{
	"time:2016-03-06T09:24:12.000000000+09:00\tint1:1\ttag1:tval1",
	"time:2016-03-06T09:24:12.000000000+09:00\tint1:2\ttag1:tval1",
	"time:2016-03-06T09:24:12.000000000+09:00\tint1:3\ttag1:tval1",
	"time:2016-03-06T09:24:12.000000002+09:00\tint1:4\ttag1:tval1",
}

func TestParseLineValidLTSV(t *testing.T) {
	parser := LTSVParser{
		MetricName:                    "ltsv_test",
		TimeLabel:                     "time",
		TimeFormat:                    "2006-01-02T15:04:05Z07:00",
		StrFieldLabels:                []string{"str1"},
		IntFieldLabels:                []string{"int1", "int2"},
		FloatFieldLabels:              []string{"float1"},
		BoolFieldLabels:               []string{"bool1", "bool2", "bool3", "bool4"},
		TagLabels:                     []string{"tag1", "tag2"},
		DuplicatePointsModifierMethod: "no_op",
		DefaultTags: map[string]string{
			"log_host": "log.example.com",
		},
	}
	metric, err := parser.ParseLine(validLTSV1)
	assert.NoError(t, err)
	assert.NotNil(t, metric)
	assert.Equal(t, "ltsv_test", metric.Name())

	fields := metric.Fields()
	assert.Equal(t, map[string]interface{}{
		"str1":   "value1",
		"int1":   int64(23),
		"int2":   int64(34),
		"float1": float64(1.23),
		"bool1":  true,
		"bool2":  false,
	}, fields)
	assert.NotContains(t, fields, "ignore_field1", "ignore_tag1")

	tags := metric.Tags()
	assert.Equal(t, map[string]string{
		"log_host": "log.example.com",
		"tag1":     "tval1",
		"tag2":     "tval2",
	}, tags)
	assert.NotContains(t, tags, "ignore_field1", "ignore_tag1")
}

func TestParseValidLTSV(t *testing.T) {
	parser := LTSVParser{
		MetricName:                    "ltsv_test",
		TimeLabel:                     "time",
		TimeFormat:                    "2006-01-02T15:04:05Z07:00",
		StrFieldLabels:                []string{"str1"},
		IntFieldLabels:                []string{"int1", "int2"},
		FloatFieldLabels:              []string{"float1"},
		BoolFieldLabels:               []string{"bool1", "bool2", "bool3", "bool4"},
		TagLabels:                     []string{"tag1", "tag2"},
		DuplicatePointsModifierMethod: "no_op",
		DefaultTags: map[string]string{
			"log_host": "log.example.com",
		},
	}
	metrics, err := parser.Parse(validLTSV2[0])
	assert.NoError(t, err)
	assert.Len(t, metrics, 0)

	metrics, err = parser.Parse(validLTSV2[1])
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "ltsv_test", metrics[0].Name())

	fields := metrics[0].Fields()
	assert.Equal(t, map[string]interface{}{
		"str1":   "value1",
		"int1":   int64(23),
		"int2":   int64(34),
		"float1": float64(1.23),
		"bool1":  true,
		"bool2":  false,
	}, fields)
	assert.NotContains(t, fields, "ignore_field1", "ignore_tag1")

	tags := metrics[0].Tags()
	assert.Equal(t, map[string]string{
		"log_host": "log.example.com",
		"tag1":     "tval1",
		"tag2":     "tval2",
	}, tags)
	assert.NotContains(t, tags, "ignore_field1", "ignore_tag1")

	metrics, err = parser.Parse(validLTSV2[2])
	assert.NoError(t, err)
	assert.Len(t, metrics, 2)
	assert.Equal(t, "ltsv_test", metrics[0].Name())

	assert.Equal(t, map[string]interface{}{
		"str1": "value2",
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{
		"log_host": "log.example.com",
	}, metrics[0].Tags())

	assert.Equal(t, map[string]interface{}{
		"str1": "value3",
	}, metrics[1].Fields())
	assert.Equal(t, map[string]string{
		"log_host": "log.example.com",
	}, metrics[1].Tags())

	metrics, err = parser.Parse(validLTSV2[3])
	assert.NoError(t, err)
	assert.Len(t, metrics, 0)

	metrics, err = parser.Parse(validLTSV2[4])
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "ltsv_test", metrics[0].Name())

	assert.Equal(t, map[string]interface{}{
		"str1": "value4",
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{
		"log_host": "log.example.com",
	}, metrics[0].Tags())
}

func TestAlwaysAddTagDuplicatePointModifier(t *testing.T) {
	parser := LTSVParser{
		MetricName:                     "ltsv_test",
		TimeLabel:                      "time",
		TimeFormat:                     "2006-01-02T15:04:05.000000000Z07:00",
		IntFieldLabels:                 []string{"int1"},
		TagLabels:                      []string{"tag1"},
		DuplicatePointsModifierMethod:  "add_uniq_tag",
		DuplicatePointsModifierUniqTag: "uniq",
		DefaultTags: map[string]string{
			"log_host": "log.example.com",
			"uniq":     "0",
		},
	}

	metric, err := parser.ParseLine(validLTSV3[0])
	assert.NoError(t, err)
	assert.NotNil(t, metric)
	assert.Equal(t, "ltsv_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"int1": int64(1),
	}, metric.Fields())
	assert.Equal(t, map[string]string{
		"log_host": "log.example.com",
		"tag1":     "tval1",
		"uniq":     "0",
	}, metric.Tags())
	assert.Equal(t, "2016-03-06T09:24:12.000000000+09:00", metric.Time().Format(parser.TimeFormat))

	metric, err = parser.ParseLine(validLTSV3[1])
	assert.NoError(t, err)
	assert.NotNil(t, metric)
	assert.Equal(t, "ltsv_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"int1": int64(2),
	}, metric.Fields())
	assert.Equal(t, map[string]string{
		"log_host": "log.example.com",
		"tag1":     "tval1",
		"uniq":     "1",
	}, metric.Tags())
	assert.Equal(t, "2016-03-06T09:24:12.000000000+09:00", metric.Time().Format(parser.TimeFormat))

	metric, err = parser.ParseLine(validLTSV3[2])
	assert.NoError(t, err)
	assert.NotNil(t, metric)
	assert.Equal(t, "ltsv_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"int1": int64(3),
	}, metric.Fields())
	assert.Equal(t, map[string]string{
		"log_host": "log.example.com",
		"tag1":     "tval1",
		"uniq":     "2",
	}, metric.Tags())
	assert.Equal(t, "2016-03-06T09:24:12.000000000+09:00", metric.Time().Format(parser.TimeFormat))

	metric, err = parser.ParseLine(validLTSV3[3])
	assert.NoError(t, err)
	assert.NotNil(t, metric)
	assert.Equal(t, "ltsv_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"int1": int64(4),
	}, metric.Fields())
	assert.Equal(t, map[string]string{
		"log_host": "log.example.com",
		"tag1":     "tval1",
		"uniq":     "0",
	}, metric.Tags())
	assert.Equal(t, "2016-03-06T09:24:12.000000002+09:00", metric.Time().Format(parser.TimeFormat))
}

func TestAddTagDuplicatePointModifier(t *testing.T) {
	parser := LTSVParser{
		MetricName:                     "ltsv_test",
		TimeLabel:                      "time",
		TimeFormat:                     "2006-01-02T15:04:05.000000000Z07:00",
		IntFieldLabels:                 []string{"int1"},
		TagLabels:                      []string{"tag1"},
		DuplicatePointsModifierMethod:  "add_uniq_tag",
		DuplicatePointsModifierUniqTag: "uniq",
		DefaultTags: map[string]string{
			"log_host": "log.example.com",
		},
	}

	metric, err := parser.ParseLine(validLTSV3[0])
	assert.NoError(t, err)
	assert.NotNil(t, metric)
	assert.Equal(t, "ltsv_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"int1": int64(1),
	}, metric.Fields())
	assert.Equal(t, map[string]string{
		"log_host": "log.example.com",
		"tag1":     "tval1",
	}, metric.Tags())
	assert.Equal(t, "2016-03-06T09:24:12.000000000+09:00", metric.Time().Format(parser.TimeFormat))

	metric, err = parser.ParseLine(validLTSV3[1])
	assert.NoError(t, err)
	assert.NotNil(t, metric)
	assert.Equal(t, "ltsv_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"int1": int64(2),
	}, metric.Fields())
	assert.Equal(t, map[string]string{
		"log_host": "log.example.com",
		"tag1":     "tval1",
		"uniq":     "1",
	}, metric.Tags())
	assert.Equal(t, "2016-03-06T09:24:12.000000000+09:00", metric.Time().Format(parser.TimeFormat))

	metric, err = parser.ParseLine(validLTSV3[2])
	assert.NoError(t, err)
	assert.NotNil(t, metric)
	assert.Equal(t, "ltsv_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"int1": int64(3),
	}, metric.Fields())
	assert.Equal(t, map[string]string{
		"log_host": "log.example.com",
		"tag1":     "tval1",
		"uniq":     "2",
	}, metric.Tags())
	assert.Equal(t, "2016-03-06T09:24:12.000000000+09:00", metric.Time().Format(parser.TimeFormat))

	metric, err = parser.ParseLine(validLTSV3[3])
	assert.NoError(t, err)
	assert.NotNil(t, metric)
	assert.Equal(t, "ltsv_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"int1": int64(4),
	}, metric.Fields())
	assert.Equal(t, map[string]string{
		"log_host": "log.example.com",
		"tag1":     "tval1",
	}, metric.Tags())
	assert.Equal(t, "2016-03-06T09:24:12.000000002+09:00", metric.Time().Format(parser.TimeFormat))
}

func TestIncTimeDuplicatePointModifier(t *testing.T) {
	parser := LTSVParser{
		MetricName:                       "ltsv_test",
		TimeLabel:                        "time",
		TimeFormat:                       "2006-01-02T15:04:05.000000000Z07:00",
		IntFieldLabels:                   []string{"int1"},
		TagLabels:                        []string{"tag1"},
		DuplicatePointsModifierMethod:    "increment_time",
		DuplicatePointsIncrementDuration: time.Nanosecond,
		DefaultTags: map[string]string{
			"log_host": "log.example.com",
		},
	}

	metric, err := parser.ParseLine(validLTSV3[0])
	assert.NoError(t, err)
	assert.NotNil(t, metric)
	assert.Equal(t, "ltsv_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"int1": int64(1),
	}, metric.Fields())
	assert.Equal(t, map[string]string{
		"log_host": "log.example.com",
		"tag1":     "tval1",
	}, metric.Tags())
	assert.Equal(t, "2016-03-06T09:24:12.000000000+09:00", metric.Time().Format(parser.TimeFormat))

	metric, err = parser.ParseLine(validLTSV3[1])
	assert.NoError(t, err)
	assert.NotNil(t, metric)
	assert.Equal(t, "ltsv_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"int1": int64(2),
	}, metric.Fields())
	assert.Equal(t, map[string]string{
		"log_host": "log.example.com",
		"tag1":     "tval1",
	}, metric.Tags())
	assert.Equal(t, "2016-03-06T09:24:12.000000001+09:00", metric.Time().Format(parser.TimeFormat))

	metric, err = parser.ParseLine(validLTSV3[2])
	assert.NoError(t, err)
	assert.NotNil(t, metric)
	assert.Equal(t, "ltsv_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"int1": int64(3),
	}, metric.Fields())
	assert.Equal(t, map[string]string{
		"log_host": "log.example.com",
		"tag1":     "tval1",
	}, metric.Tags())
	assert.Equal(t, "2016-03-06T09:24:12.000000002+09:00", metric.Time().Format(parser.TimeFormat))

	metric, err = parser.ParseLine(validLTSV3[3])
	assert.NoError(t, err)
	assert.NotNil(t, metric)
	assert.Equal(t, "ltsv_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"int1": int64(4),
	}, metric.Fields())
	assert.Equal(t, map[string]string{
		"log_host": "log.example.com",
		"tag1":     "tval1",
	}, metric.Tags())
	assert.Equal(t, "2016-03-06T09:24:12.000000003+09:00", metric.Time().Format(parser.TimeFormat))
}

func TestNoOpDuplicatePointModifier(t *testing.T) {
	parser := LTSVParser{
		MetricName:                    "ltsv_test",
		TimeLabel:                     "time",
		TimeFormat:                    "2006-01-02T15:04:05.000000000Z07:00",
		IntFieldLabels:                []string{"int1"},
		TagLabels:                     []string{"tag1"},
		DuplicatePointsModifierMethod: "no_op",
		DefaultTags: map[string]string{
			"log_host": "log.example.com",
		},
	}

	var buf bytes.Buffer
	for _, line := range validLTSV3 {
		buf.WriteString(line)
		buf.WriteByte(byte('\n'))
	}

	metrics, err := parser.Parse(buf.Bytes())
	assert.NoError(t, err)
	// NOTE: Even though 4 metrics are created here, 3 of these will be merged on
	// a InfluxDB database.
	assert.Len(t, metrics, 4)

	assert.Equal(t, "ltsv_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"int1": int64(1),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{
		"log_host": "log.example.com",
		"tag1":     "tval1",
	}, metrics[0].Tags())
	assert.Equal(t, "2016-03-06T09:24:12.000000000+09:00", metrics[0].Time().Format(parser.TimeFormat))

	assert.Equal(t, "ltsv_test", metrics[1].Name())
	assert.Equal(t, map[string]interface{}{
		"int1": int64(2),
	}, metrics[1].Fields())
	assert.Equal(t, map[string]string{
		"log_host": "log.example.com",
		"tag1":     "tval1",
	}, metrics[1].Tags())
	assert.Equal(t, "2016-03-06T09:24:12.000000000+09:00", metrics[1].Time().Format(parser.TimeFormat))

	assert.Equal(t, "ltsv_test", metrics[2].Name())
	assert.Equal(t, map[string]interface{}{
		"int1": int64(3),
	}, metrics[2].Fields())
	assert.Equal(t, map[string]string{
		"log_host": "log.example.com",
		"tag1":     "tval1",
	}, metrics[2].Tags())
	assert.Equal(t, "2016-03-06T09:24:12.000000000+09:00", metrics[2].Time().Format(parser.TimeFormat))

	assert.Equal(t, "ltsv_test", metrics[3].Name())
	assert.Equal(t, map[string]interface{}{
		"int1": int64(4),
	}, metrics[3].Fields())
	assert.Equal(t, map[string]string{
		"log_host": "log.example.com",
		"tag1":     "tval1",
	}, metrics[3].Tags())
	assert.Equal(t, "2016-03-06T09:24:12.000000002+09:00", metrics[3].Time().Format(parser.TimeFormat))
}

func TestInvalidDuplicatePointsModifierMethod(t *testing.T) {
	parser := LTSVParser{
		DuplicatePointsModifierMethod: "",
	}
	metric, err := parser.ParseLine("")
	assert.Error(t, err)
	assert.Nil(t, metric)
}
