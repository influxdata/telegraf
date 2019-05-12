package wavefront

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

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

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

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
	const metricSeparator = "."
	var out []byte
	buf := pbFree.Get().(*buffer)

	for fieldName, value := range m.Fields() {
		var name string

		if fieldName == "value" {
			name = s.Prefix + m.Name()
		} else {
			name = s.Prefix + m.Name() + metricSeparator + fieldName
		}

		if s.UseStrict {
			name = strictSanitizedChars.Replace(name)
		} else {
			name = sanitizedChars.Replace(name)
		}

		name = pathReplacer.Replace(name)

		metricValue, buildError := buildValue(value, name)
		if buildError != nil {
			// bad value continue to next metric
			continue
		}
		source, tags := buildTags(m.Tags(), s)
		metric := wavefront.MetricPoint{
			Metric:    name,
			Timestamp: m.Time().Unix(),
			Value:     metricValue,
			Source:    source,
			Tags:      tags,
		}
		out = append(out, formatMetricPoint(buf, &metric, s)...)
	}

	pbFree.Put(buf)
	return out, nil
}

func (s *WavefrontSerializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	var batch []byte
	for _, m := range metrics {
		buf, err := s.Serialize(m)
		if err != nil {
			return nil, err
		}
		batch = append(batch, buf...)
	}
	return batch, nil
}

func findSourceTag(mTags map[string]string, s *WavefrontSerializer) string {
	if src, ok := mTags["source"]; ok {
		delete(mTags, "source")
		return src
	}
	for _, src := range s.SourceOverride {
		if source, ok := mTags[src]; ok {
			delete(mTags, src)
			mTags["telegraf_host"] = mTags["host"]
			return source
		}
	}
	return mTags["host"]
}

func buildTags(mTags map[string]string, s *WavefrontSerializer) (string, map[string]string) {
	// Remove all empty tags.
	for k, v := range mTags {
		if v == "" {
			delete(mTags, k)
		}
	}
	source := findSourceTag(mTags, s)
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
		return float64(p), nil
	case uint64:
		return float64(p), nil
	case float64:
		return p, nil
	case string:
		// return an error but don't log
		return 0, errors.New("string type not supported")
	default:
		// return an error and log a debug message
		err := fmt.Errorf("unexpected type: %T, with value: %v, for :%s", v, v, name)
		log.Printf("D! Serializer [wavefront] %s\n", err.Error())
		return 0, err
	}
}

func formatMetricPoint(b *buffer, metricPoint *wavefront.MetricPoint, s *WavefrontSerializer) []byte {
	b.Reset()

	b.WriteChar('"')
	b.WriteString(metricPoint.Metric)
	b.WriteString(`" `)
	b.WriteFloat64(metricPoint.Value)
	b.WriteChar(' ')
	b.WriteUnit64(uint64(metricPoint.Timestamp))
	b.WriteString(` source="`)
	b.WriteString(metricPoint.Source)
	b.WriteChar('"')

	for k, v := range metricPoint.Tags {
		b.WriteString(` "`)
		if s.UseStrict {
			b.WriteString(strictSanitizedChars.Replace(k))
		} else {
			b.WriteString(sanitizedChars.Replace(k))
		}
		b.WriteString(`"="`)
		b.WriteString(tagValueReplacer.Replace(v))
		b.WriteChar('"')
	}

	b.WriteChar('\n')

	return *b
}

// pbFree is the print buffer pool
var pbFree = sync.Pool{
	New: func() interface{} {
		b := make(buffer, 0, 128)
		return &b
	},
}

// Use a fast and simple buffer for constructing statsd messages
type buffer []byte

func (b *buffer) Reset() { *b = (*b)[:0] }

func (b *buffer) Write(p []byte) {
	*b = append(*b, p...)
}

func (b *buffer) WriteString(s string) {
	*b = append(*b, s...)
}

// This is named WriteChar instead of WriteByte because the 'stdmethods' check
// of 'go vet' wants WriteByte to have the signature:
//
// 	func (b *buffer) WriteByte(c byte) error { ... }
//
func (b *buffer) WriteChar(c byte) {
	*b = append(*b, c)
}

func (b *buffer) WriteUnit64(val uint64) {
	*b = strconv.AppendUint(*b, val, 10)
}

func (b *buffer) WriteFloat64(val float64) {
	*b = strconv.AppendFloat(*b, val, 'f', 6, 64)
}
