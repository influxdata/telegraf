package wavefront

import (
	"fmt"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"log"
)

type Wavefront struct {
	Prefix          string
	Host            string
	Port            int
	SimpleFields    bool
	MetricSeparator string
	ConvertPaths    bool
	UseRegex        bool
	SourceOverride  []string
	DebugAll        bool
}

// catch many of the invalid chars that could appear in a metric or tag name
var sanitizedChars = strings.NewReplacer(
	"!", "-", "@", "-", "#", "-", "$", "-", "%", "-", "^", "-", "&", "-",
	"*", "-", "(", "-", ")", "-", "+", "-", "`", "-", "'", "-", "\"", "-",
	"[", "-", "]", "-", "{", "-", "}", "-", ":", "-", ";", "-", "<", "-",
	">", "-", ",", "-", "?", "-", "/", "-", "\\", "-", "|", "-", " ", "-",
)
// instead of Replacer which may miss some special characters we can use a regex pattern, but this is significantly slower than Replacer
var sanitizedRegex, _ = regexp.Compile("[^a-zA-Z\\d_.-]")

var tagValueReplacer = strings.NewReplacer("\"", "\\\"", "*", "-")

var pathReplacer = strings.NewReplacer("_", "_")

var sampleConfig = `
  ## prefix for metrics keys
  #prefix = "my.specific.prefix."

  ## DNS name of the wavefront proxy server
  host = "wavefront.example.com"

  ## Port that the Wavefront proxy server listens on
  port = 2878

  ## wether to use "value" for name of simple fields
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
  #source_override = ["hostname", "snmp_host", "node_host"]

  ## Print additional debug information requires debug = true at the agent level
  #debug_all = false
`

type MetricLine struct {
	Metric    string
	Value     string
	Timestamp int64
	Tags      string
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
	tcpAddr, err := net.ResolveTCPAddr("tcp", uri)
	if err != nil {
		return fmt.Errorf("Wavefront: TCP address cannot be resolved %s", err.Error())
	}
	connection, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return fmt.Errorf("Wavefront: TCP connect fail %s", err.Error())
	}
	defer connection.Close()
	return nil
}

func (w *Wavefront) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	// Send Data to Wavefront proxy Server
	uri := fmt.Sprintf("%s:%d", w.Host, w.Port)
	tcpAddr, _ := net.ResolveTCPAddr("tcp", uri)
	connection, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return fmt.Errorf("Wavefront: TCP connect fail %s", err.Error())
	}
	defer connection.Close()

	for _, m := range metrics {
		for _, metric := range buildMetrics(m, w) {
			messageLine := fmt.Sprintf("%s %s %v %s\n",	metric.Metric, metric.Value, metric.Timestamp, metric.Tags)
			log.Printf("D! Output [wavefront] %s", messageLine)
			_, err := connection.Write([]byte(messageLine))
			if err != nil {
				return fmt.Errorf("Wavefront: TCP writing error %s", err.Error())
			}
		}
	}

	return nil
}

func buildTags(mTags map[string]string, w *Wavefront) []string {
	sourceTagFound := false

	for _, s := range w.SourceOverride {
		for k, v := range mTags {
			if k == s {
				mTags["source"] = v
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
		mTags["source"] = mTags["host"]
	}
	mTags["telegraf_host"] = mTags["host"]
	delete(mTags, "host")

	tags := make([]string, len(mTags))
	index := 0
	for k, v := range mTags {
		if w.UseRegex {
			tags[index] = fmt.Sprintf("%s=\"%s\"", sanitizedRegex.ReplaceAllString(k, "-"), tagValueReplacer.Replace(v))
		} else {
			tags[index] = fmt.Sprintf("%s=\"%s\"", sanitizedChars.Replace(k), tagValueReplacer.Replace(v))
		}

		index++
	}

	sort.Strings(tags)
	return tags
}

func buildMetrics(m telegraf.Metric, w *Wavefront) []*MetricLine {
	if w.DebugAll {
		log.Printf("D! Output [wavefront] original name: %s\n", m.Name())
	}

	ret := []*MetricLine{}
	for fieldName, value := range m.Fields() {
		if w.DebugAll {
			log.Printf("D! Output [wavefront] original field: %s\n", fieldName)
		}

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

		metric := &MetricLine{
			Metric: name,
			Timestamp: m.UnixNano() / 1000000000,
		}
		metricValue, buildError := buildValue(value, metric.Metric)
		if buildError != nil {
			log.Printf("E! Output [wavefront] %s\n", buildError.Error())
			continue
		}
		metric.Value = metricValue
		tagsSlice := buildTags(m.Tags(), w)
		metric.Tags = fmt.Sprint(strings.Join(tagsSlice, " "))
		ret = append(ret, metric)
	}
	return ret
}

func buildValue(v interface{}, name string) (string, error) {
	var retv string
	switch p := v.(type) {
	case int64:
		retv = IntToString(int64(p))
	case uint64:
		retv = UIntToString(uint64(p))
	case float64:
		retv = FloatToString(float64(p))
	default:
		return retv, fmt.Errorf("unexpected type: %T, with value: %v, for: %s", v, v, name)
	}
	return retv, nil
}

func IntToString(input_num int64) string {
	return strconv.FormatInt(input_num, 10)
}

func UIntToString(input_num uint64) string {
	return strconv.FormatUint(input_num, 10)
}

func FloatToString(input_num float64) string {
	return strconv.FormatFloat(input_num, 'f', 6, 64)
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
			ConvertPaths: true,
		}
	})
}
