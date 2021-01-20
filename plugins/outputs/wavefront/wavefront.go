package wavefront

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	wavefront "github.com/wavefronthq/wavefront-sdk-go/senders"
)

const maxTagLength = 254

type Wavefront struct {
	Url             string
	Token           string
	Host            string
	Port            int
	Prefix          string
	SimpleFields    bool
	MetricSeparator string
	ConvertPaths    bool
	ConvertBool     bool
	UseRegex        bool
	UseStrict       bool
	TruncateTags    bool
	ImmediateFlush  bool
	SourceOverride  []string
	StringToNumber  map[string][]map[string]float64

	sender wavefront.Sender
	Log    telegraf.Logger
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
var sanitizedRegex = regexp.MustCompile("[^a-zA-Z\\d_.-]")

var tagValueReplacer = strings.NewReplacer("*", "-")

var pathReplacer = strings.NewReplacer("_", "_")

var sampleConfig = `
  ## Url for Wavefront Direct Ingestion or using HTTP with Wavefront Proxy
  ## If using Wavefront Proxy, also specify port. example: http://proxyserver:2878
  url = "https://metrics.wavefront.com"

  ## Authentication Token for Wavefront. Only required if using Direct Ingestion
  #token = "DUMMY_TOKEN"  
  
  ## DNS name of the wavefront proxy server. Do not use if url is specified
  #host = "wavefront.example.com"

  ## Port that the Wavefront proxy server listens on. Do not use if url is specified
  #port = 2878

  ## prefix for metrics keys
  #prefix = "my.specific.prefix."

  ## whether to use "value" for name of simple fields. default is false
  #simple_fields = false

  ## character to use between metric and field name.  default is . (dot)
  #metric_separator = "."

  ## Convert metric name paths to use metricSeparator character
  ## When true will convert all _ (underscore) characters in final metric name. default is true
  #convert_paths = true

  ## Use Strict rules to sanitize metric and tag names from invalid characters
  ## When enabled forward slash (/) and comma (,) will be accepted
  #use_strict = false

  ## Use Regex to sanitize metric and tag names from invalid characters
  ## Regex is more thorough, but significantly slower. default is false
  #use_regex = false

  ## point tags to use as the source name for Wavefront (if none found, host will be used)
  #source_override = ["hostname", "address", "agent_host", "node_host"]

  ## whether to convert boolean values to numeric values, with false -> 0.0 and true -> 1.0. default is true
  #convert_bool = true

  ## Truncate metric tags to a total of 254 characters for the tag name value. Wavefront will reject any 
  ## data point exceeding this limit if not truncated. Defaults to 'false' to provide backwards compatibility.
  #truncate_tags = false

  ## Flush the internal buffers after each batch. This effectively bypasses the background sending of metrics
  ## normally done by the Wavefront SDK. This can be used if you are experiencing buffer overruns. The sending 
  ## of metrics will block for a longer time, but this will be handled gracefully by the internal buffering in
  ## Telegraf.
  #immediate_flush = true

  ## Define a mapping, namespaced by metric prefix, from string values to numeric values
  ##   deprecated in 1.9; use the enum processor plugin
  #[[outputs.wavefront.string_to_number.elasticsearch]]
  #  green = 1.0
  #  yellow = 0.5
  #  red = 0.0
`

type MetricPoint struct {
	Metric    string
	Value     float64
	Timestamp int64
	Source    string
	Tags      map[string]string
}

func (w *Wavefront) Connect() error {

	if len(w.StringToNumber) > 0 {
		w.Log.Warn("The string_to_number option is deprecated; please use the enum processor instead")
	}

	flushSeconds := 5
	if w.ImmediateFlush {
		flushSeconds = 86400 // Set a very long flush interval if we're flushing directly
	}
	if w.Url != "" {
		w.Log.Debug("connecting over http/https using Url: %s", w.Url)
		sender, err := wavefront.NewDirectSender(&wavefront.DirectConfiguration{
			Server:               w.Url,
			Token:                w.Token,
			FlushIntervalSeconds: flushSeconds,
		})
		if err != nil {
			return fmt.Errorf("Wavefront: Could not create Wavefront Sender for Url: %s", w.Url)
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
			return fmt.Errorf("Wavefront: Could not create Wavefront Sender for Host: %q and Port: %d", w.Host, w.Port)
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
					return fmt.Errorf("Wavefront sending error: %v", err)
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
			} else {
				return 0, nil
			}
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
					val, hasVal := mapping[string(p)]
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

func (w *Wavefront) SampleConfig() string {
	return sampleConfig
}

func (w *Wavefront) Description() string {
	return "Configuration for Wavefront server to send metrics to"
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
