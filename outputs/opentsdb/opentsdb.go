package opentsdb

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/influxdb/influxdb/client"
	"github.com/influxdb/telegraf/outputs"
)

type OpenTSDB struct {
	Prefix string

	Host string
	Port int

	Debug bool
}

var sampleConfig = `
	# prefix for metrics keys
	prefix = "my.specific.prefix."

	## Telnet Mode ##
	# DNS name of the OpenTSDB server in telnet mode
	host = "opentsdb.example.com"

	# Port of the OpenTSDB server in telnet mode
	port = 4242

	# Debug true - Prints OpenTSDB communication
	debug = false
`

type MetricLine struct {
	Metric    string
	Timestamp int64
	Value     string
	Tags      string
}

func (o *OpenTSDB) Connect() error {
	// Test Connection to OpenTSDB Server
	uri := fmt.Sprintf("%s:%d", o.Host, o.Port)
	tcpAddr, err := net.ResolveTCPAddr("tcp", uri)
	if err != nil {
		return fmt.Errorf("OpenTSDB: TCP address cannot be resolved")
	}
	connection, err := net.DialTCP("tcp", nil, tcpAddr)
	defer connection.Close()
	if err != nil {
		return fmt.Errorf("OpenTSDB: Telnet connect fail")
	}
	return nil
}

func (o *OpenTSDB) Write(bp client.BatchPoints) error {
	if len(bp.Points) == 0 {
		return nil
	}
	var timeNow = time.Now()
	// Send Data with telnet / socket communication
	uri := fmt.Sprintf("%s:%d", o.Host, o.Port)
	tcpAddr, _ := net.ResolveTCPAddr("tcp", uri)
	connection, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return fmt.Errorf("OpenTSDB: Telnet connect fail")
	}
	for _, pt := range bp.Points {
		metric := &MetricLine{
			Metric:    fmt.Sprintf("%s%s", o.Prefix, pt.Measurement),
			Timestamp: timeNow.Unix(),
		}
		metricValue, buildError := buildValue(bp, pt)
		if buildError != nil {
			fmt.Printf("OpenTSDB: %s\n", buildError.Error())
			continue
		}
		metric.Value = metricValue

		tagsSlice := buildTags(bp.Tags, pt.Tags)
		metric.Tags = fmt.Sprint(strings.Join(tagsSlice, " "))

		messageLine := fmt.Sprintf("put %s %v %s %s\n", metric.Metric, metric.Timestamp, metric.Value, metric.Tags)
		if o.Debug {
			fmt.Print(messageLine)
		}
		_, err := connection.Write([]byte(messageLine))
		if err != nil {
			return fmt.Errorf("OpenTSDB: Telnet writing error %s", err.Error())
		}
	}
	defer connection.Close()

	return nil
}

func buildTags(bpTags map[string]string, ptTags map[string]string) []string {
	tags := make([]string, (len(bpTags) + len(ptTags)))
	index := 0
	for k, v := range bpTags {
		tags[index] = fmt.Sprintf("%s=%s", k, v)
		index += 1
	}
	for k, v := range ptTags {
		tags[index] = fmt.Sprintf("%s=%s", k, v)
		index += 1
	}
	sort.Strings(tags)
	return tags
}

func buildValue(bp client.BatchPoints, pt client.Point) (string, error) {
	var retv string
	var v = pt.Fields["value"]
	switch p := v.(type) {
	case int64:
		retv = IntToString(int64(p))
	case uint64:
		retv = UIntToString(uint64(p))
	case float64:
		retv = FloatToString(float64(p))
	default:
		return retv, fmt.Errorf("unexpected type %T with value %v for OpenTSDB", v, v)
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

func (o *OpenTSDB) SampleConfig() string {
	return sampleConfig
}

func (o *OpenTSDB) Description() string {
	return "Configuration for OpenTSDB server to send metrics to"
}

func (o *OpenTSDB) Close() error {
	return nil
}

func init() {
	outputs.Add("opentsdb", func() outputs.Output {
		return &OpenTSDB{}
	})
}
