package access_log

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
)

type AccessLogParser struct {
	MetricName  string
	DataType    string
	DefaultTags map[string]string
}

func (v *AccessLogParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	log := strings.TrimSpace(string(buf))

	// unless it's a string, separate out any fields in the buffer,
	// ignore anything but the last.
	if len(log) == 0 {
		return []telegraf.Metric{}, nil
	}

	metric, err := telegraf.NewMetric(v.MetricName, v.DefaultTags,
		parseAccessLog(log), time.Now().UTC())
	if err != nil {
		return nil, err
	}

	return []telegraf.Metric{metric}, nil
}

func (v *AccessLogParser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := v.Parse([]byte(line))

	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, fmt.Errorf("Can not parse the line: %s, for data format: value", line)
	}

	return metrics[0], nil
}

func (v *AccessLogParser) SetDefaultTags(tags map[string]string) {
	v.DefaultTags = tags
}

func parseAccessLog(log string) map[string]interface{} {
	// Parse Hostname
	hostname_regexp := regexp.MustCompile("\\s+([a-z0-9]+(-[a-z0-9]+)*\\.)+[a-z]{2,}\\s+")
	hostname := strings.TrimSpace(hostname_regexp.FindString(log))

	// Parse Request Method
	method_regexp := regexp.MustCompile("\\s+(GET|HEAD|POST|PUT|DELETE|TRACE|OPTIONS|CONNECT)\\s+")
	method := strings.TrimSpace(method_regexp.FindString(log))

	// Parse Request URI
	uri_regexp := regexp.MustCompile("\\s+(?:\\/([^?#]*))")
	uri := strings.Split(strings.TrimSpace(uri_regexp.FindString(log)), " ")[0]

	return map[string]interface{}{
		"hostname": hostname,
		"method":   method,
		"path":     uri,
	}
}
