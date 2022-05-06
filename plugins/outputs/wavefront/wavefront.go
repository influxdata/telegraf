package wavefront

import (
	"fmt"
	"regexp"
	"strings"

	wavefront "github.com/wavefronthq/wavefront-sdk-go/senders"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

const maxTagLength = 254

type Wavefront struct {
	URL             string                          `toml:"url"`
	Token           string                          `toml:"token"`
	Host            string                          `toml:"host"`
	Port            int                             `toml:"port"`
	Prefix          string                          `toml:"prefix"`
	SimpleFields    bool                            `toml:"simple_fields"`
	MetricSeparator string                          `toml:"metric_separator"`
	ConvertPaths    bool                            `toml:"convert_paths"`
	ConvertBool     bool                            `toml:"convert_bool"`
	UseRegex        bool                            `toml:"use_regex"`
	UseStrict       bool                            `toml:"use_strict"`
	TruncateTags    bool                            `toml:"truncate_tags"`
	ImmediateFlush  bool                            `toml:"immediate_flush"`
	SourceOverride  []string                        `toml:"source_override"`
	StringToNumber  map[string][]map[string]float64 `toml:"string_to_number" deprecated:"1.9.0;use the enum processor instead"`

	sender wavefront.Sender
	Log    telegraf.Logger `toml:"-"`
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

// instead of Replacer which may miss some special characters we can use a regex pattern, but this is significantly slower than Replacer
var sanitizedRegex = regexp.MustCompile(`[^a-zA-Z\d_.-]`)

var tagValueReplacer = strings.NewReplacer("*", "-")

var pathReplacer = strings.NewReplacer("_", "_")

type MetricPoint struct {
	Metric    string
	Value     float64
	Timestamp int64
	Source    string
	Tags      map[string]string
}

func (w *Wavefront) Connect() error {
	flushSeconds := 5
	if w.ImmediateFlush {
		flushSeconds = 86400 // Set a very long flush interval if we're flushing directly
	}
	if w.URL != "" {
		w.Log.Debug("connecting over http/https using Url: %s", w.URL)
		sender, err := wavefront.NewDirectSender(&wavefront.DirectConfiguration{
			Server:               w.URL,
			Token:                w.Token,
			FlushIntervalSeconds: flushSeconds,
		})
		if err != nil {
			return fmt.Errorf("could not create Wavefront Sender for Url: %s", w.URL)
		}
		w.sender = sender
	} else {
		w.Log.Debugf("connecting over tcp using Host: %q and Port: %d", w.Host, w.Port)
		sender, err := wavefront.NewProxySender(&wavefront.ProxyConfiguration{
			Host:                 w.Host,
			MetricsPort:          w.Port,
			FlushIntervalSeconds: flushSeconds,
		})
		if err != nil {
			return fmt.Errorf("could not create Wavefront Sender for Host: %q and Port: %d", w.Host, w.Port)
		}
		w.sender = sender
	}

	if w.ConvertPaths && w.MetricSeparator == "_" {
		w.ConvertPaths = false
	}
	if w.ConvertPaths {
		pathReplacer = strings.NewReplacer("_", w.MetricSeparator)
	}
	return nil
}

func (w *Wavefront) Write(metrics []telegraf.Metric) error {
	for _, m := range metrics {
		for _, point := range w.buildMetrics(m) {
			err := w.sender.SendMetric(point.Metric, point.Value, point.Timestamp, point.Source, point.Tags)
			if err != nil {
				if isRetryable(err) {
					if flushErr := w.sender.Flush(); flushErr != nil {
						w.Log.Errorf("wavefront flushing error: %v", flushErr)
					}
					return fmt.Errorf("wavefront sending error: %v", err)
				}
				w.Log.Errorf("non-retryable error during Wavefront.Write: %v", err)
				w.Log.Debugf("Non-retryable metric data: Name: %v, Value: %v, Timestamp: %v, Source: %v, PointTags: %v ", point.Metric, point.Value, point.Timestamp, point.Source, point.Tags)
			}
		}
	}
	if w.ImmediateFlush {
		w.Log.Debugf("Flushing batch of %d points", len(metrics))
		return w.sender.Flush()
	}
	return nil
}

func (w *Wavefront) buildMetrics(m telegraf.Metric) []*MetricPoint {
	ret := []*MetricPoint{}

	for fieldName, value := range m.Fields() {
		var name string
		if !w.SimpleFields && fieldName == "value" {
			name = fmt.Sprintf("%s%s", w.Prefix, m.Name())
		} else {
			name = fmt.Sprintf("%s%s%s%s", w.Prefix, m.Name(), w.MetricSeparator, fieldName)
		}

		if w.UseRegex {
			name = sanitizedRegex.ReplaceAllLiteralString(name, "-")
		} else if w.UseStrict {
			name = strictSanitizedChars.Replace(name)
		} else {
			name = sanitizedChars.Replace(name)
		}

		if w.ConvertPaths {
			name = pathReplacer.Replace(name)
		}

		metric := &MetricPoint{
			Metric:    name,
			Timestamp: m.Time().Unix(),
		}

		metricValue, buildError := buildValue(value, metric.Metric, w)
		if buildError != nil {
			w.Log.Debugf("Error building tags: %s\n", buildError.Error())
			continue
		}
		metric.Value = metricValue

		source, tags := w.buildTags(m.Tags())
		metric.Source = source
		metric.Tags = tags

		ret = append(ret, metric)
	}
	return ret
}

func (w *Wavefront) buildTags(mTags map[string]string) (string, map[string]string) {
	// Remove all empty tags.
	for k, v := range mTags {
		if v == "" {
			delete(mTags, k)
		}
	}

	// find source, use source_override property if needed
	var source string
	if s, ok := mTags["source"]; ok {
		source = s
		delete(mTags, "source")
	} else {
		sourceTagFound := false
		for _, s := range w.SourceOverride {
			for k, v := range mTags {
				if k == s {
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
	source = tagValueReplacer.Replace(source)

	// remove default host tag
	delete(mTags, "host")

	// sanitize tag keys and values
	tags := make(map[string]string)
	for k, v := range mTags {
		var key string
		if w.UseRegex {
			key = sanitizedRegex.ReplaceAllLiteralString(k, "-")
		} else if w.UseStrict {
			key = strictSanitizedChars.Replace(k)
		} else {
			key = sanitizedChars.Replace(k)
		}
		val := tagValueReplacer.Replace(v)
		if w.TruncateTags {
			if len(key) > maxTagLength {
				w.Log.Warnf("Tag key length > 254. Skipping tag: %s", key)
				continue
			}
			if len(key)+len(val) > maxTagLength {
				w.Log.Debugf("Key+value length > 254: %s", key)
				val = val[:maxTagLength-len(key)]
			}
		}
		tags[key] = val
	}

	return source, tags
}

func buildValue(v interface{}, name string, w *Wavefront) (float64, error) {
	switch p := v.(type) {
	case bool:
		if w.ConvertBool {
			if p {
				return 1, nil
			}
			return 0, nil
		}
	case int64:
		return float64(v.(int64)), nil
	case uint64:
		return float64(v.(uint64)), nil
	case float64:
		return v.(float64), nil
	case string:
		for prefix, mappings := range w.StringToNumber {
			if strings.HasPrefix(name, prefix) {
				for _, mapping := range mappings {
					val, hasVal := mapping[p]
					if hasVal {
						return val, nil
					}
				}
			}
		}
		return 0, fmt.Errorf("unexpected type: %T, with value: %v, for: %s", v, v, name)
	default:
		return 0, fmt.Errorf("unexpected type: %T, with value: %v, for: %s", v, v, name)
	}
	return 0, fmt.Errorf("unexpected type: %T, with value: %v, for: %s", v, v, name)
}

func (w *Wavefront) Close() error {
	w.sender.Close()
	return nil
}

func init() {
	outputs.Add("wavefront", func() telegraf.Output {
		return &Wavefront{
			Token:           "DUMMY_TOKEN",
			MetricSeparator: ".",
			ConvertPaths:    true,
			ConvertBool:     true,
			TruncateTags:    false,
			ImmediateFlush:  true,
		}
	})
}

// TODO: Currently there's no canonical way to exhaust all
// retryable/non-retryable errors from wavefront, so this implementation just
// handles known non-retryable errors in a case-by-case basis and assumes all
// other errors are retryable.
// A support ticket has been filed against wavefront to provide a canonical way
// to distinguish between retryable and non-retryable errors (link is not
// public).
func isRetryable(err error) bool {
	if err != nil {
		// "empty metric name" errors are non-retryable as retry will just keep
		// getting the same error again and again.
		if strings.Contains(err.Error(), "empty metric name") {
			return false
		}
	}
	return true
}
