package warp10

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type Warp10 struct {
	Prefix string

	WarpUrl string

	Token string

	Debug bool
}

var sampleConfig = `
  # prefix for metrics class Name
  prefix = "Prefix"
  ## POST HTTP(or HTTPS) ##
  # Url name of the Warp 10 server
  warp_url = "WarpUrl"
  # Token to access your app on warp 10
  token = "Token"
  # Debug true - Prints Warp communication
  debug = false
`

type MetricLine struct {
	Metric    string
	Timestamp int64
	Value     string
	Tags      string
}

func (o *Warp10) Connect() error {
	return nil
}

func (o *Warp10) Write(metrics []telegraf.Metric) error {

	var out io.Writer = ioutil.Discard
	if o.Debug {
		out = os.Stdout
	}

	if len(metrics) == 0 {
		return nil
	}
	var now = time.Now()
	collectString := make([]string, 0)
	index := 0
	for _, mm := range metrics {

		for k, v := range mm.Fields() {

			metric := &MetricLine{
				Metric:    fmt.Sprintf("%s%s", o.Prefix, mm.Name()+"."+k),
				Timestamp: now.Unix() * 1000000,
			}

			metricValue, err := buildValue(v)
			if err != nil {
				log.Printf("Warp: %s\n", err.Error())
				continue
			}
			metric.Value = metricValue

			tagsSlice := buildTags(mm.Tags())
			metric.Tags = strings.Join(tagsSlice, ",")

			messageLine := fmt.Sprintf("%d// %s{%s} %s\n", metric.Timestamp, metric.Metric, metric.Tags, metric.Value)

			collectString = append(collectString, messageLine)
			index += 1
		}
	}
	payload := fmt.Sprint(strings.Join(collectString, "\n"))
	//defer connection.Close()
	req, err := http.NewRequest("POST", o.WarpUrl, bytes.NewBufferString(payload))
	req.Header.Set("X-Warp10-Token", o.Token)
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Fprintf(out, "response Status: %#v", resp.Status)
	fmt.Fprintf(out, "response Headers: %#v", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Fprintf(out, "response Body: %#v", string(body))

	return nil
}

func buildTags(ptTags map[string]string) []string {
	sizeTags := len(ptTags)
	sizeTags += 1
	tags := make([]string, sizeTags)
	index := 0
	for k, v := range ptTags {
		tags[index] = fmt.Sprintf("%s=%s", k, v)
		index += 1
	}
	tags[index] = fmt.Sprintf("source=telegraf")
	sort.Strings(tags)
	return tags
}

func buildValue(v interface{}) (string, error) {
	var retv string
	switch p := v.(type) {
	case int64:
		retv = IntToString(int64(p))
	case string:
		retv = fmt.Sprintf("'%s'", p)
	case bool:
		retv = BoolToString(bool(p))
	case uint64:
		retv = UIntToString(uint64(p))
	case float64:
		retv = FloatToString(float64(p))
	default:
		retv = fmt.Sprintf("'%s'", p)
		//    return retv, fmt.Errorf("unexpected type %T with value %v for Warp", v, v)
	}
	return retv, nil
}

func IntToString(input_num int64) string {
	return strconv.FormatInt(input_num, 10)
}

func BoolToString(input_bool bool) string {
	return strconv.FormatBool(input_bool)
}

func UIntToString(input_num uint64) string {
	return strconv.FormatUint(input_num, 10)
}

func FloatToString(input_num float64) string {
	return strconv.FormatFloat(input_num, 'f', 6, 64)
}

func (o *Warp10) SampleConfig() string {
	return sampleConfig
}

func (o *Warp10) Description() string {
	return "Configuration for Warp server to send metrics to"
}

func (o *Warp10) Close() error {
	// Basically nothing to do for Warp10 here
	return nil
}

func init() {
	outputs.Add("warp10", func() telegraf.Output {
		return &Warp10{}
	})
}
