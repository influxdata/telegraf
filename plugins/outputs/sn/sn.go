package sn

import (
	"fmt"
	"time"
	"log"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type SN struct {
	Prefix string
	Url string
	Username         string
	Password         string
	HttpBatchSize int
	Debug bool
}

var sanitizedChars = strings.NewReplacer("@", "-", "*", "-", " ", "_",
	`%`, "-", "#", "-", "$", "-", ":", "_")

var sampleConfig = `
  ## prefix for metrics keys
  prefix = "telegraf."
  
  ##username and password to access the MID api
  username = "admin"
  password = "admin"
  
  ## url of the metric api on the MID side
  url = "http://127.0.0.1:9080/api/mid/sa/metrics"

  ## Debug true - Prints SN communication
  debug = false
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

func (o *SN) Connect() error {
	// Test Connection to SN Server
	

	//TODO - create http client and test conneectivity against a ping resource
	return nil
}

func (o *SN) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	u, err := url.Parse(o.Url)
	if err != nil {
		return fmt.Errorf("Error in parsing host url: %s", err.Error())
	}

	return o.WriteHttp(metrics, u)
}

func (o *SN) WriteHttp(metrics []telegraf.Metric, u *url.URL) error {
	http := SNHttp{
		Host:      u.Host,
		Url:       o.Url,
		Username:      o.Username,
		Password:  o.Password,
		BatchSize: o.HttpBatchSize,
		Debug:     o.Debug,
	}

	for _, m := range metrics {
		now := m.UnixNano() / int64(time.Millisecond)
		tags := cleanTags(m.Tags())

		for fieldName, value := range m.Fields() {
			switch value.(type) {
			case int64:
			case uint64:
			case float64:
			default:
				log.Printf("D! SN does not support metric value: [%s] of type [%T].\n", value, value)
				continue
			}

			metric := &HttpMetric{
				Metric: sanitizedChars.Replace(fmt.Sprintf("%s%s_%s",
					o.Prefix, m.Name(), fieldName)),
				Tags:      tags,
				Timestamp: now,
				Value:     value,
				Source:    "telegraf",
				Node:	   tags["host"] ,	
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



func cleanTags(tags map[string]string) map[string]string {
	tagSet := make(map[string]string, len(tags))
	for k, v := range tags {
		tagSet[sanitizedChars.Replace(k)] = sanitizedChars.Replace(v)
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
		return retv, fmt.Errorf("unexpected type %T with value %v for SN", v, v)
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

func (o *SN) SampleConfig() string {
	return sampleConfig
}

func (o *SN) Description() string {
	return "Configuration for SN server to send metrics to"
}

func (o *SN) Close() error {
	return nil
}

func init() {
	outputs.Add("sn", func() telegraf.Output {
		return &SN{}
	})
}
