package wavefront

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type Wavefront struct {
	Prefix string

	Host string
	Port int

	Debug bool
	Metric_separator string
}

var sanitizedChars = strings.NewReplacer("*", "-", `%`, "-", "#", "-")

var sampleConfig = `
  ## prefix for metrics keys
  prefix = "my.specific.prefix."

  ## Telnet Mode ##
  ## DNS name of the wavefront proxy server in telnet mode
  host = "wavefront.example.com"

  ## Port of the Wavefront proxy server in telnet mode
  port = 4242

  ## Debug true - Prints Wavefront communication
  debug = false
`

type MetricLine struct {
	Metric    string
	Timestamp int64
	Value     string
	Tags      string
}

func (w *Wavefront) Connect() error {
	// Test Connection to OpenTSDB Server
	if w.Metric_separator == "" {
		w.Metric_separator = "_"
	}
	uri := fmt.Sprintf("%s:%d", w.Host, w.Port)
	tcpAddr, err := net.ResolveTCPAddr("tcp", uri)
	if err != nil {
		return fmt.Errorf("Wavefront: TCP address cannot be resolved")
	}
	connection, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return fmt.Errorf("Wavefront: TCP connect fail")
	}
	defer connection.Close()
	return nil
}

func (w *Wavefront) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}
	now := time.Now()

	// Send Data with telnet / socket communication
	uri := fmt.Sprintf("%s:%d", w.Host, w.Port)
	tcpAddr, _ := net.ResolveTCPAddr("tcp", uri)
	connection, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return fmt.Errorf("Wavefront: TCP connect fail")
	}
	defer connection.Close()

	for _, m := range metrics {
		for _, metric := range buildMetrics(m, now, w) {
			messageLine := fmt.Sprintf("put %s %v %s %s\n",
				metric.Metric, metric.Timestamp, metric.Value, metric.Tags)
			if w.Debug {
				fmt.Print(messageLine)
			}
			_, err := connection.Write([]byte(messageLine))
			if err != nil {
				return fmt.Errorf("Wavefront: TCP writing error %s", err.Error())
			}
		}
	}

	return nil
}

func buildTags(mTags map[string]string) []string {
	tags := make([]string, len(mTags))
	index := 0
	for k, v := range mTags {
		tags[index] = sanitizedChars.Replace(fmt.Sprintf("%s=\"%s\"", k, v))
		index++
	}
	sort.Strings(tags)
	return tags
}

func buildMetrics(m telegraf.Metric, now time.Time, w *Wavefront) []*MetricLine {
	ret := []*MetricLine{}
	for fieldName, value := range m.Fields() {
		metric := &MetricLine{
			Metric: sanitizedChars.Replace(fmt.Sprintf("%s%s%s%s",
				w.Prefix, m.Name(), w.Metric_separator, fieldName)),
			Timestamp: now.Unix(),
		}
		metricValue, buildError := buildValue(value, metric.Metric)
		if buildError != nil {
			fmt.Printf("Wavefront: %s\n", buildError.Error())
			continue
		}
		metric.Value = metricValue
		tagsSlice := buildTags(m.Tags())
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
	return "Configuration for OpenTSDB server to send metrics to"
}

func (w *Wavefront) Close() error {
	return nil
}

func init() {
	outputs.Add("wavefront", func() telegraf.Output {
		return &Wavefront{}
	})
}
