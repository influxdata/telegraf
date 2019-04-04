package wavefront

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs/wavefront"
)

// WavefrontSerializer : WavefrontSerializer struct
type WavefrontSerializer struct {
	Prefix         string
	UseStrict      bool
	SourceOverride []string
}

// catch many of the invalid chars that could appear in a metric or tag name
var sanitizedChars = strings.NewReplacer(
	"!", "-", "@", "-", "#", "-", "$", "-", "%", "-", "^", "-", "&", "-",
	"*", "-", "(", "-", ")", "-", "+", "-", "`", "-", "'", "-", "\"", "-",
	"[", "-", "]", "-", "{", "-", "}", "-", ":", "-", ";", "-", "<", "-",
	">", "-", ",", "-", "?", "-", "/", "-", "\\", "-", "|", "-", " ", "-",
	"=", "-",
)

// catch many of the invalid chars that could appear in a metric or tag name
var strictSanitizedChars = strings.NewReplacer(
	"!", "-", "@", "-", "#", "-", "$", "-", "%", "-", "^", "-", "&", "-",
	"*", "-", "(", "-", ")", "-", "+", "-", "`", "-", "'", "-", "\"", "-",
	"[", "-", "]", "-", "{", "-", "}", "-", ":", "-", ";", "-", "<", "-",
	">", "-", "?", "-", "\\", "-", "|", "-", " ", "-", "=", "-",
)

var tagValueReplacer = strings.NewReplacer("\"", "\\\"", "*", "-")

var pathReplacer = strings.NewReplacer("_", ".")

func NewSerializer(prefix string, useStrict bool, sourceOverride []string) (*WavefrontSerializer, error) {
	s := &WavefrontSerializer{
		Prefix:         prefix,
		UseStrict:      useStrict,
		SourceOverride: sourceOverride,
	}
	return s, nil
}

// Serialize : Serialize based on Wavefront format
func (s *WavefrontSerializer) Serialize(m telegraf.Metric) ([]byte, error) {
	out := []byte{}
	metricSeparator := "."

	for fieldName, value := range m.Fields() {
		var name string

		if fieldName == "value" {
			name = fmt.Sprintf("%s%s", s.Prefix, m.Name())
		} else {
			name = fmt.Sprintf("%s%s%s%s", s.Prefix, m.Name(), metricSeparator, fieldName)
		}

		if s.UseStrict {
			name = strictSanitizedChars.Replace(name)
		} else {
			name = sanitizedChars.Replace(name)
		}

		name = pathReplacer.Replace(name)

		metric := &wavefront.MetricPoint{
			Metric:    name,
			Timestamp: m.Time().Unix(),
		}

		metricValue, buildError := buildValue(value, metric.Metric)
		if buildError != nil {
			// bad value continue to next metric
			continue
		}
		metric.Value = metricValue

		source, tags := buildTags(m.Tags(), s)
		metric.Source = source
		metric.Tags = tags

		out = append(out, formatMetricPoint(metric, s)...)
	}
	return out, nil
}

func (s *WavefrontSerializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	var batch bytes.Buffer
	for _, m := range metrics {
		buf, err := s.Serialize(m)
		if err != nil {
			return nil, err
		}
		_, err = batch.Write(buf)
		if err != nil {
			return nil, err
		}
	}
	return batch.Bytes(), nil
}

func buildTags(mTags map[string]string, s *WavefrontSerializer) (string, map[string]string) {

	// Remove all empty tags.
	for k, v := range mTags {
		if v == "" {
			delete(mTags, k)
		}
	}

	var source string

	if src, ok := mTags["source"]; ok {
		source = src
		delete(mTags, "source")
	} else {
		sourceTagFound := false
		for _, src := range s.SourceOverride {
			for k, v := range mTags {
				if k == src {
					source = v
					mTags["telegraf_host"] = mTags["host"]
					sourceTagFound = true
					delete(mTags, k)
					break
				}
			}
			if sourceTagFound {
				break
			}
		}

		if !sourceTagFound {
			source = mTags["host"]
		}
	}

	delete(mTags, "host")

	return tagValueReplacer.Replace(source), mTags
}

func buildValue(v interface{}, name string) (float64, error) {
	switch p := v.(type) {
	case bool:
		if p {
			return 1, nil
		} else {
			return 0, nil
		}
	case int64:
		return float64(v.(int64)), nil
	case uint64:
		return float64(v.(uint64)), nil
	case float64:
		return v.(float64), nil
	case string:
		// return an error but don't log
		return 0, fmt.Errorf("string type not supported")
	default:
		// return an error and log a debug message
		err := fmt.Errorf("unexpected type: %T, with value: %v, for :%s", v, v, name)
		log.Printf("D! Serializer [wavefront] %s\n", err.Error())
		return 0, err
	}
}

func formatMetricPoint(metricPoint *wavefront.MetricPoint, s *WavefrontSerializer) []byte {
	var buffer bytes.Buffer
	buffer.WriteString("\"")
	buffer.WriteString(metricPoint.Metric)
	buffer.WriteString("\" ")
	buffer.WriteString(strconv.FormatFloat(metricPoint.Value, 'f', 6, 64))
	buffer.WriteString(" ")
	buffer.WriteString(strconv.FormatInt(metricPoint.Timestamp, 10))
	buffer.WriteString(" source=\"")
	buffer.WriteString(metricPoint.Source)
	buffer.WriteString("\"")

	for k, v := range metricPoint.Tags {
		buffer.WriteString(" \"")
		if s.UseStrict {
			buffer.WriteString(strictSanitizedChars.Replace(k))
		} else {
			buffer.WriteString(sanitizedChars.Replace(k))
		}
		buffer.WriteString("\"=\"")
		buffer.WriteString(tagValueReplacer.Replace(v))
		buffer.WriteString("\"")
	}

	buffer.WriteString("\n")

	return buffer.Bytes()
}
