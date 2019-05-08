package kairosdb

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

var (
	allowedChars = regexp.MustCompile(`[^a-zA-Z0-9-_./\p{L}]`)
	hypenChars   = strings.NewReplacer(
		"@", "-",
		"*", "-",
		`%`, "-",
		"#", "-",
		"$", "-")
	defaultHttpPath  = "/api/v1/datapoints"
	defaultSeperator = "_"
)

type KairosDB struct {
	Prefix string

	Host string
	Port int

	Username string
	Password string

	HttpBatchSize int // deprecated httpBatchSize form in 1.8
	HttpPath      string

	Debug bool

	Separator string
}

var sampleConfig = `
  ## prefix for metrics keys
  prefix = "my.specific.prefix."

  ## DNS name of the KairosDB server
  ## Using "kairosdb.example.com" or "tcp://kairosdb.example.com" will use the
  ## telnet API. "http://kairosdb.example.com" will use the Http API.
  host = "http://kairosdb.example.com"

  ## Port of the KairosDB server
  port = 4242

  ## HTTP basic authentication
  ## Leave username or password empty to disable
  username = ""
  password = ""

  ## Number of data points to send to KairosDB in Http requests.
  ## Not used with telnet API.
  http_batch_size = 50

  ## URI Path for Http requests to KairosDB.
  ## Used in cases where KairosDB is located behind a reverse proxy.
  http_path = "/api/v1/datapoints"

  ## Debug true - Prints KairosDB communication
  debug = false

  ## Separator separates measurement name from field
  separator = "_"
`

func ToLineFormat(tags map[string]string) string {
	tagsArray := make([]string, len(tags))
	index := 0
	for k, v := range tags {
		tagsArray[index] = fmt.Sprintf("%s=%s", k, v)
		index++
	}
	sort.Strings(tagsArray)
	return strings.Join(tagsArray, " ")
}

func (o *KairosDB) Connect() error {
	if !strings.HasPrefix(o.Host, "http") && !strings.HasPrefix(o.Host, "tcp") {
		o.Host = "tcp://" + o.Host
	}
	// Test Connection to KairosDB Server
	u, err := url.Parse(o.Host)
	if err != nil {
		return fmt.Errorf("Error in parsing host url: %s", err.Error())
	}

	uri := fmt.Sprintf("%s:%d", u.Host, o.Port)
	tcpAddr, err := net.ResolveTCPAddr("tcp", uri)
	if err != nil {
		return fmt.Errorf("KairosDB TCP address cannot be resolved: %s", err)
	}
	connection, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return fmt.Errorf("KairosDB Telnet connect fail: %s", err)
	}
	defer connection.Close()
	return nil
}

func (o *KairosDB) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	u, err := url.Parse(o.Host)
	if err != nil {
		return fmt.Errorf("Error in parsing host url: %s", err.Error())
	}

	if u.Scheme == "" || u.Scheme == "tcp" {
		return o.WriteTelnet(metrics, u)
	} else if u.Scheme == "http" || u.Scheme == "https" {
		return o.WriteHttp(metrics, u)
	} else {
		return fmt.Errorf("Unknown scheme in host parameter.")
	}
}

func (o *KairosDB) WriteHttp(metrics []telegraf.Metric, u *url.URL) error {
	http := kairosDBHttp{
		Host:      u.Host,
		Port:      o.Port,
		Username:  o.Username,
		Password:  o.Password,
		Scheme:    u.Scheme,
		User:      u.User,
		BatchSize: o.HttpBatchSize,
		Path:      o.HttpPath,
		Debug:     o.Debug,
	}

	for _, m := range metrics {
		now := m.Time().UnixNano() / 1000000
		tags := cleanTags(m.Tags())

		for fieldName, value := range m.Fields() {
			switch value.(type) {
			case int64:
			case uint64:
			case float64:
			default:
				log.Printf("D! KairosDB does not support metric value: [%s] of type [%T].\n", value, value)
				continue
			}

			metric := &HttpMetric{
				Metric: sanitize(fmt.Sprintf("%s%s%s%s",
					o.Prefix, m.Name(), o.Separator, fieldName)),
				Tags:      tags,
				Timestamp: now,
				Value:     value,
			}

			if err := http.sendDataPoint(metric); err != nil {
				return err
			}
		}
	}

	if err := http.flush(); err != nil {
		return err
	}

	return nil
}

func (o *KairosDB) WriteTelnet(metrics []telegraf.Metric, u *url.URL) error {
	// Send Data with telnet / socket communication
	uri := fmt.Sprintf("%s:%d", u.Host, o.Port)
	tcpAddr, _ := net.ResolveTCPAddr("tcp", uri)
	connection, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return fmt.Errorf("KairosDB: Telnet connect fail")
	}
	defer connection.Close()

	for _, m := range metrics {
		now := m.Time().UnixNano() / 1000000
		tags := ToLineFormat(cleanTags(m.Tags()))

		for fieldName, value := range m.Fields() {
			switch value.(type) {
			case int64:
			case uint64:
			case float64:
			default:
				log.Printf("D! KairosDB does not support metric value: [%s] of type [%T].\n", value, value)
				continue
			}

			metricValue, buildError := buildValue(value)
			if buildError != nil {
				log.Printf("E! KairosDB: %s\n", buildError.Error())
				continue
			}

			messageLine := fmt.Sprintf("put %s %v %s %s\n",
				sanitize(fmt.Sprintf("%s%s%s%s", o.Prefix, m.Name(), o.Separator, fieldName)),
				now, metricValue, tags)

			_, err := connection.Write([]byte(messageLine))
			if err != nil {
				return fmt.Errorf("KairosDB: Telnet writing error %s", err.Error())
			}
		}
	}

	return nil
}

func cleanTags(tags map[string]string) map[string]string {
	tagSet := make(map[string]string, len(tags))
	for k, v := range tags {
		val := sanitize(v)
		if val != "" {
			tagSet[sanitize(k)] = val
		}
	}
	return tagSet
}

func buildValue(v interface{}) (string, error) {
	var retv string
	switch p := v.(type) {
	case int64:
		retv = IntToString(int64(p))
	case uint64:
		retv = UIntToString(uint64(p))
	case float64:
		retv = FloatToString(float64(p))
	default:
		return retv, fmt.Errorf("unexpected type %T with value %v for KairosDB", v, v)
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

func (o *KairosDB) SampleConfig() string {
	return sampleConfig
}

func (o *KairosDB) Description() string {
	return "Configuration for KairosDB server to send metrics to"
}

func (o *KairosDB) Close() error {
	return nil
}

func sanitize(value string) string {
	// Apply special hypenation rules to preserve backwards compatibility
	value = hypenChars.Replace(value)
	// Replace any remaining illegal chars
	return allowedChars.ReplaceAllLiteralString(value, "_")
}

func init() {
	outputs.Add("kairosdb", func() telegraf.Output {
		return &KairosDB{
			HttpPath:  defaultHttpPath,
			Separator: defaultSeperator,
		}
	})
}
