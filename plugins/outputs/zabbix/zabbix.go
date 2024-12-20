//go:generate ../../../tools/readme_config_includer/generator
package zabbix

import (
	_ "embed"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/datadope-io/go-zabbix/v2"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
)

// zabbixSender is an interface to send autoregister data to Zabbix.
// It is implemented by Zabbix.Sender.
// Created to be able to mock Zabbix.Sender in tests.
type zabbixSender interface {
	Send(packet *zabbix.Packet) (res zabbix.Response, err error)
	SendMetrics(metrics []*zabbix.Metric) (resActive zabbix.Response, resTrapper zabbix.Response, err error)
	RegisterHost(hostname string, hostMetadata string) error
}

// Zabbix allows pushing metrics to Zabbix software
type Zabbix struct {
	Address                    string          `toml:"address"`
	AgentActive                bool            `toml:"agent_active"`
	KeyPrefix                  string          `toml:"key_prefix"`
	HostTag                    string          `toml:"host_tag"`
	SkipMeasurementPrefix      bool            `toml:"skip_measurement_prefix"`
	LLDSendInterval            config.Duration `toml:"lld_send_interval"`
	LLDClearInterval           config.Duration `toml:"lld_clear_interval"`
	Autoregister               string          `toml:"autoregister"`
	AutoregisterResendInterval config.Duration `toml:"autoregister_resend_interval"`
	Log                        telegraf.Logger `toml:"-"`

	// lldHandler handles low level discovery data
	lldHandler zabbixLLD
	// lldLastSend store the last LLD send to known where to send it again
	lldLastSend time.Time
	// autoregisterLastSend stores the last time autoregister data was sent to Zabbix for each host.
	autoregisterLastSend map[string]time.Time
	// sender is the interface to send data to Zabbix.
	sender zabbixSender
}

//go:embed sample.conf
var sampleConfig string

func (*Zabbix) SampleConfig() string {
	return sampleConfig
}

// Connect does nothing, Write() would initiate connection in each call.
// Checking if Zabbix server is alive in this step does not allow Telegraf
// to start if there is a temporal connection problem with the server.
func (*Zabbix) Connect() error {
	return nil
}

// Init initializes LLD and autoregister maps. Copy config values to them. Configure Logger.
func (z *Zabbix) Init() error {
	// Add port to address if not present
	if _, _, err := net.SplitHostPort(z.Address); err != nil {
		z.Address = net.JoinHostPort(z.Address, "10051")
	}

	z.sender = zabbix.NewSender(z.Address)
	// Initialize autoregisterLastSend map with size one, as the most common scenario is to have one host.
	z.autoregisterLastSend = make(map[string]time.Time, 1)
	z.lldLastSend = time.Now()

	z.lldHandler = zabbixLLD{
		log:           z.Log,
		hostTag:       z.HostTag,
		clearInterval: z.LLDClearInterval,
		lastClear:     time.Now(),
		current:       make(map[uint64]lldInfo, 100),
	}

	return nil
}

func (*Zabbix) Close() error {
	return nil
}

// Write sends metrics to Zabbix server
func (z *Zabbix) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	zbxMetrics := make([]*zabbix.Metric, 0, len(metrics))

	for _, metric := range metrics {
		hostname, err := getHostname(z.HostTag, metric)
		if err != nil {
			z.Log.Errorf("Error getting hostname for metric %v: %v", metric, err)
			continue
		}

		zbxMetrics = append(zbxMetrics, z.processMetric(metric)...)

		// Handle hostname for autoregister
		z.autoregisterAdd(hostname)

		// Process LLD data
		err = z.lldHandler.Add(metric)
		if err != nil {
			z.Log.Errorf("Error processing LLD for metric %v: %v", metric, err)
		}
	}

	// Send LLD data if enough time has passed
	if time.Since(z.lldLastSend) > time.Duration(z.LLDSendInterval) {
		z.lldLastSend = time.Now()
		for _, lldMetric := range z.lldHandler.Push() {
			zbxMetrics = append(zbxMetrics, z.processMetric(lldMetric)...)
		}
	}

	// Send metrics to Zabbix server
	err := z.sendZabbixMetrics(zbxMetrics)

	// Send autoregister data after sending metrics.
	z.autoregisterPush()

	return err
}

// sendZabbixMetrics sends metrics to Zabbix server
func (z *Zabbix) sendZabbixMetrics(zbxMetrics []*zabbix.Metric) error {
	if len(zbxMetrics) == 0 {
		return nil
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
	sort.Slice(zbxMetrics, func(i, j int) bool {
		return zbxMetrics[i].Clock < zbxMetrics[j].Clock
	})

	packet := zabbix.NewPacket(zbxMetrics, z.AgentActive)
	_, err := z.sender.Send(packet)

	return err
}

// processMetric converts a Telegraf metric to a list of Zabbix metrics.
// Ignore metrics with no hostname.
func (z Zabbix) processMetric(metric telegraf.Metric) []*zabbix.Metric {
	zbxMetrics := make([]*zabbix.Metric, 0, len(metric.FieldList()))

	for _, field := range metric.FieldList() {
		zbxMetric, err := z.buildZabbixMetric(metric, field.Key, field.Value)
		if err != nil {
			z.Log.Errorf("Error converting telegraf metric to Zabbix format: %v", err)
			continue
		}

		zbxMetrics = append(zbxMetrics, zbxMetric)
	}

	return zbxMetrics
}

// buildZabbixMetric builds a Zabbix metric from a Telegraf metric, for one particular value.
func (z Zabbix) buildZabbixMetric(metric telegraf.Metric, fieldName string, value interface{}) (*zabbix.Metric, error) {
	hostname, err := getHostname(z.HostTag, metric)
	if err != nil {
		return nil, fmt.Errorf("error getting hostname: %w", err)
	}

	metricValue, err := internal.ToString(value)
	if err != nil {
		return nil, fmt.Errorf("error converting value: %w", err)
	}

	key := z.KeyPrefix + metric.Name() + "." + fieldName
	if z.SkipMeasurementPrefix {
		key = z.KeyPrefix + fieldName
	}

	// Ignore host tag.
	// We want to add tags to the key in alphabetical order. Eg.:
	// lld.dns_query.query_time_ms[DOMAIN,RECORD_TYPE,SERVER]
	// TagList already return the tags in alphabetical order.
	tagValues := make([]string, 0, len(metric.TagList()))

	for _, tag := range metric.TagList() {
		if tag.Key == z.HostTag {
			continue
		}

		// Get tag values in the same order as the tag keys in the tags slice.
		tagValues = append(tagValues, tag.Value)
	}

	if len(tagValues) != 0 {
		key = fmt.Sprintf("%v[%v]", key, strings.Join(tagValues, ","))
	}

	return zabbix.NewMetric(hostname, key, metricValue, z.AgentActive, metric.Time().Unix()), nil
}

func init() {
	outputs.Add("zabbix", func() telegraf.Output {
		return &Zabbix{
			KeyPrefix:                  "telegraf.",
			HostTag:                    "host",
			AutoregisterResendInterval: config.Duration(time.Minute * 30),
			LLDSendInterval:            config.Duration(time.Minute * 10),
			LLDClearInterval:           config.Duration(time.Hour),
		}
	})
}

// getHostname returns the hostname from the tags, or the system hostname if not found.
func getHostname(hostTag string, metric telegraf.Metric) (string, error) {
	if hostname, ok := metric.GetTag(hostTag); ok {
		return hostname, nil
	}

	return os.Hostname()
}
