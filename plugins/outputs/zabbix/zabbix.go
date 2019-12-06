package zabbix

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/datadope-io/go-zabbix"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type Zabbix struct {
	Host                   string
	Port                   int
	AgentActive            bool `toml:"agent_active"`
	Prefix                 string
	SkipMeasurementPrefix  bool `toml:"skip_measurement_prefix"`
	Autoregister           string
	AutoregisterSendPeriod internal.Duration
	autoregisterLastSend   map[string]time.Time
}

var sampleConfig = `
	## Address of zabbix host
	host = "zabbix.example.com"

	## Port of the Zabbix server
	port = 10051

	## Send metrics as type "Zabbix agent (active)"
	agent_active = false

	## Add prefix to all keys sent to Zabbix (set by default)
	prefix = "telegraf."

	## Skip measurement prefix to all keys sent to Zabbix (false by default)
	skip_measurement_prefix = false

	## This field will be sent as HostMetadata to Zabbix Server to autoregister the host.
	autoregister = "Example aaa222111cccdddaaaa"

	## This is the period with which self-registrations are sent.
	## Only applies if autoregister is defined in config file
	## Set by default
	autoregister_send_period = "30m"
`

func (z *Zabbix) SampleConfig() string {
	return sampleConfig
}

func (z *Zabbix) Description() string {
	return "Configuration for sender to Zabbix server"
}

// Connect does nothing, Write() would initiate connection in each call.
// Checking if Zabbix server is alive in this step does not allow Telegraf
// to start if there is a temporal connection problem with the server.
func (z *Zabbix) Connect() error {
	return nil
}

func (z *Zabbix) Close() error {
	return nil
}

func (z *Zabbix) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	var zbxMetrics []*zabbix.Metric

	for _, metric := range metrics {
		hostname, ok := metric.Tags()["host"]
		if !ok {
			log.Printf("W! Missing hostname in metric %v", metric)
			continue
		}

		// Save the different values seen in "host" tag to autoregister them later
		if z.Autoregister != "" {
			_, exists := z.autoregisterLastSend[hostname]
			if !exists {
				z.autoregisterLastSend[hostname] = time.Time{}
			}
		}

		for fieldName, value := range metric.Fields() {
			metricValue, err := buildValue(value)
			if err != nil {
				log.Printf("E! Error converting value: %v", err)
				continue
			}

			key := fmt.Sprintf("%s%s.%s", z.Prefix, metric.Name(), fieldName)

			if z.SkipMeasurementPrefix {
				key = fmt.Sprintf("%s%s", z.Prefix, fieldName)
			}

			// We want to add tags to the key in alphabetical order. Eg.:
			// lld.dns_query.query_time_ms[DOMAIN,RECORD_TYPE,SERVER]
			tags := []string{}
			for key := range metric.Tags() {
				if key == "host" {
					continue
				}
				tags = append(tags, key)
			}
			sort.Strings(tags)

			if len(tags) != 0 {
				tagValues := []string{}

				for _, tag := range tags {
					tagValues = append(tagValues, metric.Tags()[tag])
				}

				key = fmt.Sprintf("%v[%v]", key, strings.Join(tagValues, ","))
			}

			zbxMetric := zabbix.NewMetric(hostname, key, metricValue, z.AgentActive, metric.Time().Unix())
			zbxMetrics = append(zbxMetrics, zbxMetric)
		}
	}

	// Sort metrics by time.
	// Avoid extra work in Zabbix when generating the trends.
	// If values are not sent in clock order, trend generation is forced to
	// make more database operations.
	// When a value is received with a new hour, trend is flushed to the
	// database.
	// If later a value is received with the previous hour, new trend is
	// flushed, old one is retrieved from database and updated.
	// When a new value with the new hour is received, old trend is flushed,
	// new trend retrieved from database and updated.
	if len(zbxMetrics) > 0 {
		sort.Slice(zbxMetrics, func(i, j int) bool {
			return zbxMetrics[i].Clock < zbxMetrics[j].Clock
		})
	}

	// All metrics are of the same kind (trappers or agent active)
	packet := zabbix.NewPacket(zbxMetrics, z.AgentActive)
	sender := zabbix.NewSender(z.Host, z.Port)

	_, err := sender.Send(packet)
	if err != nil {
		return fmt.Errorf("Zabbix: Sender writing error: %v", err)
	}

	// For each "host" tag seen, send an autoregister request to Zabbix server.
	// z.AutoregisterSendPeriod is the interval at which requests are resend.
	if z.Autoregister != "" {
		s := zabbix.NewSender(z.Host, z.Port)
		for hostname, timeLastSend := range z.autoregisterLastSend {
			if time.Since(timeLastSend) > z.AutoregisterSendPeriod.Duration {
				log.Printf("D! [output.zabbix] Autoregistering host %v", hostname)
				err = s.RegisterHost(hostname, z.Autoregister)
				if err != nil {
					log.Printf("E! [output.zabbix] Autoregistering host %s: %v", hostname, err)
				}
				z.autoregisterLastSend[hostname] = time.Now()
			}
		}
	}

	return nil
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
	case string:
		retv = p
	case bool:
		retv = BoolToString(p)
	default:
		return retv, fmt.Errorf("unexpected type %T with value %v for Zabbix", v, v)
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

func BoolToString(input_num bool) string {
	x := "0"
	if input_num {
		x = "1"
	}
	return x
}

func init() {
	outputs.Add("zabbix", func() telegraf.Output {
		return &Zabbix{
			Port:                   10051,
			Prefix:                 "telegraf.",
			SkipMeasurementPrefix:  false,
			AutoregisterSendPeriod: internal.Duration{Duration: time.Minute * 30},
			autoregisterLastSend:   map[string]time.Time{},
		}
	})
}
