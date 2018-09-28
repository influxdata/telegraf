package wavefront

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"

	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type Wavefront struct {
	Prefix          string
	Host            string
	Port            int
	SimpleFields    bool
	MetricSeparator string
	ConvertPaths    bool
	ConvertBool     bool
	UseRegex        bool
	SourceOverride  []string
	StringToNumber  map[string][]map[string]float64
}

// catch many of the invalid chars that could appear in a metric or tag name
var sanitizedChars = strings.NewReplacer(
	"!", "-", "@", "-", "#", "-", "$", "-", "%", "-", "^", "-", "&", "-",
	"*", "-", "(", "-", ")", "-", "+", "-", "`", "-", "'", "-", "\"", "-",
	"[", "-", "]", "-", "{", "-", "}", "-", ":", "-", ";", "-", "<", "-",
	">", "-", ",", "-", "?", "-", "/", "-", "\\", "-", "|", "-", " ", "-",
	"=", "-",
)

// instead of Replacer which may miss some special characters we can use a regex pattern, but this is significantly slower than Replacer
var sanitizedRegex = regexp.MustCompile("[^a-zA-Z\\d_.-]")

var tagValueReplacer = strings.NewReplacer("\"", "\\\"", "*", "-")

var pathReplacer = strings.NewReplacer("_", "_")

var sampleConfig = `
  ## DNS name of the wavefront proxy server
  host = "wavefront.example.com"

  ## Port that the Wavefront proxy server listens on
  port = 2878

  ## prefix for metrics keys
  #prefix = "my.specific.prefix."

  ## whether to use "value" for name of simple fields
  #simple_fields = false

  ## character to use between metric and field name.  defaults to . (dot)
  #metric_separator = "."

  ## Convert metric name paths to use metricSeperator character
  ## When true (default) will convert all _ (underscore) chartacters in final metric name
  #convert_paths = true

  ## Use Regex to sanitize metric and tag names from invalid characters
  ## Regex is more thorough, but significantly slower
  #use_regex = false

  ## point tags to use as the source name for Wavefront (if none found, host will be used)
  #source_override = ["hostname", "agent_host", "node_host"]

  ## whether to convert boolean values to numeric values, with false -> 0.0 and true -> 1.0.  default true
  #convert_bool = true

  ## Define a mapping, namespaced by metric prefix, from string values to numeric values
  ## The example below maps "green" -> 1.0, "yellow" -> 0.5, "red" -> 0.0 for
  ## any metrics beginning with "elasticsearch"
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
	if w.ConvertPaths && w.MetricSeparator == "_" {
		w.ConvertPaths = false
	}
	if w.ConvertPaths {
		pathReplacer = strings.NewReplacer("_", w.MetricSeparator)
	}

	// Test Connection to Wavefront proxy Server
	uri := fmt.Sprintf("%s:%d", w.Host, w.Port)
	_, err := net.ResolveTCPAddr("tcp", uri)
	if err != nil {
		return fmt.Errorf("Wavefront: TCP address cannot be resolved %s", err.Error())
	}
	connection, err := net.Dial("tcp", uri)
	if err != nil {
		return fmt.Errorf("Wavefront: TCP connect fail %s", err.Error())
	}
	defer connection.Close()
	return nil
}

func (w *Wavefront) Write(metrics []telegraf.Metric) error {

	// Send Data to Wavefront proxy Server
	uri := fmt.Sprintf("%s:%d", w.Host, w.Port)
	connection, err := net.Dial("tcp", uri)
	if err != nil {
		return fmt.Errorf("Wavefront: TCP connect fail %s", err.Error())
	}
	defer connection.Close()
	connection.SetWriteDeadline(time.Now().Add(5 * time.Second))

	for _, m := range metrics {
		for _, metricPoint := range buildMetrics(m, w) {
			metricLine := formatMetricPoint(metricPoint, w)
			_, err := connection.Write([]byte(metricLine))
			if err != nil {
				return fmt.Errorf("Wavefront: TCP writing error %s", err.Error())
			}
		}
	}

	return nil
}

func buildMetrics(m telegraf.Metric, w *Wavefront) []*MetricPoint {
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
			log.Printf("D! Output [wavefront] %s\n", buildError.Error())
			continue
		}
		metric.Value = metricValue

		source, tags := buildTags(m.Tags(), w)
		metric.Source = source
		metric.Tags = tags

		ret = append(ret, metric)
	}
	return ret
}

func buildTags(mTags map[string]string, w *Wavefront) (string, map[string]string) {

	// Remove all empty tags.
	for k, v := range mTags {
		if v == "" {
			delete(mTags, k)
		}
	}

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

	delete(mTags, "host")

	return tagValueReplacer.Replace(source), mTags
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

func formatMetricPoint(metricPoint *MetricPoint, w *Wavefront) string {
	buffer := bytes.NewBufferString("")
	buffer.WriteString(metricPoint.Metric)
	buffer.WriteString(" ")
	buffer.WriteString(strconv.FormatFloat(metricPoint.Value, 'f', 6, 64))
	buffer.WriteString(" ")
	buffer.WriteString(strconv.FormatInt(metricPoint.Timestamp, 10))
	buffer.WriteString(" source=\"")
	buffer.WriteString(metricPoint.Source)
	buffer.WriteString("\"")

	for k, v := range metricPoint.Tags {
		buffer.WriteString(" ")
		if w.UseRegex {
			buffer.WriteString(sanitizedRegex.ReplaceAllLiteralString(k, "-"))
		} else {
			buffer.WriteString(sanitizedChars.Replace(k))
		}
		buffer.WriteString("=\"")
		buffer.WriteString(tagValueReplacer.Replace(v))
		buffer.WriteString("\"")
	}

	buffer.WriteString("\n")

	return buffer.String()
}

func (w *Wavefront) SampleConfig() string {
	return sampleConfig
}

func (w *Wavefront) Description() string {
	return "Configuration for Wavefront server to send metrics to"
}

func (w *Wavefront) Close() error {
	return nil
}

func init() {
	outputs.Add("wavefront", func() telegraf.Output {
		return &Wavefront{
			MetricSeparator: ".",
			ConvertPaths:    true,
			ConvertBool:     true,
		}
	})
}
