package processors

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs/sematext/sender"
	"math/rand"
	"sync"
	"time"
)

const (
	metainfoRetryIntervalSeconds = 60
	maxMetainfoRetries           = 10
)

// MetricMetainfo contains metainfo about a single metric
type MetricMetainfo struct {
	token       string
	name        string
	namespace   string
	semType     SematextMetricType
	numericType NumericType
	label       string
	description string
	host        string
}

// SematextMetricType is an enumeration of metric types expected by Sematext backend
type SematextMetricType int

// Possible values for the ValueType enum.
const (
	_ SematextMetricType = iota
	Counter
	Gauge
)

func (s SematextMetricType) String() string {
	return [...]string{"", "Counter", "Gauge"}[s]
}

// NumericType represents metric's data type
type NumericType int

const (
	UnsupportedNumericType NumericType = iota
	Long
	Double
	Bool
)

func (s NumericType) String() string {
	return [...]string{"", "Long", "Double", "Bool"}[s]
}

// Metainfo is a processor that extracts metainfo from telegraf metrics and sends it to Sematext backend
type Metainfo struct {
	log         telegraf.Logger
	token       string
	sentMetrics map[string]*MetricMetainfo
	lock        sync.Mutex
	sendChannel chan bool
	serializer  *MetainfoSerializer

	metainfoURL  string
	senderConfig *sender.Config
	sender       *sender.Sender
}

// NewMetainfo creates a new Metainfo processor
func NewMetainfo(log telegraf.Logger, token string, receiverURL string, senderConfig *sender.Config) BatchProcessor {
	sentMetricsMap := make(map[string]*MetricMetainfo)
	return &Metainfo{
		log:          log,
		token:        token,
		sentMetrics:  sentMetricsMap,
		sendChannel:  make(chan bool, 1),
		metainfoURL:  receiverURL + "/write?db=metainfo",
		senderConfig: senderConfig,
		sender:       sender.NewSender(senderConfig),
		serializer:   NewMetainfoSerializer(log),
	}
}

// Process contains core logic of Metainfo processor
func (m *Metainfo) Process(metrics []telegraf.Metric) ([]telegraf.Metric, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	newMetrics := make(map[string]*MetricMetainfo)

	for _, metric := range metrics {
		for _, field := range metric.FieldList() {
			mInfo, mKey := processMetric(m.token, metric, field, m.sentMetrics)
			if mInfo != nil {
				newMetrics[mKey] = mInfo
			}
		}
	}

	if len(newMetrics) > 0 {
		go m.sendMetainfo(newMetrics)

		// add newMetrics to sentMetrics so they are not sent again from other goroutines
		for k, v := range newMetrics {
			m.sentMetrics[k] = v
		}
	}

	return metrics, nil
}

func (m *Metainfo) sendMetainfo(newMetrics map[string]*MetricMetainfo) {
	reqCounter := 0

	mInfoSlice := make([]*MetricMetainfo, 0, len(newMetrics))
	for _, mInfo := range newMetrics {
		mInfoSlice = append(mInfoSlice, mInfo)
	}
	body := m.serializer.Write(mInfoSlice)

	for {
		select {
		case <-m.sendChannel:
			return
		default:
			m.log.Infof("sending metainfo to Sematext endpoint %s", m.metainfoURL)
			res, err := m.sender.Request("POST", m.metainfoURL, "text/plain; charset=utf-8", body)
			reqCounter++
			if err != nil {
				// possibly non-recoverable, return without retrying
				m.log.Errorf("can't send metainfo to Sematext endpoint %s, error: %v",
					m.metainfoURL, err)
				return
			}
			responseContent := response(res)
			res.Body.Close()

			success := res.StatusCode >= 200 && res.StatusCode < 300

			if success {
				m.log.Infof("successfully sent metainfo to Sematext endpoint %s, batch size : %d",
					m.metainfoURL, len(newMetrics))
				return
			}

			m.log.Errorf("can't send the metainfo to Sematext endpoint %s, response code: %d, response: %s",
				m.metainfoURL, res.StatusCode, responseContent)

			badRequest := res.StatusCode >= 400 && res.StatusCode < 500
			if badRequest {
				m.log.Infof("no retry for bad requests, response code was %d", res.StatusCode)
				return
			}

			if reqCounter >= maxMetainfoRetries {
				m.log.Warnf("max retries (%d) exceeded, cancelling the request permanently",
					maxMetainfoRetries)
				return
			}

			nextSleepIntervalSec := int32(reqCounter * metainfoRetryIntervalSeconds)
			m.log.Infof("metainfo sending retry in %d seconds", nextSleepIntervalSec)
			time.Sleep(time.Second * time.Duration(rand.Int31n(nextSleepIntervalSec)))
		}
	}
}

func processMetric(token string, metric telegraf.Metric, field *telegraf.Field,
	sentMetrics map[string]*MetricMetainfo) (*MetricMetainfo, string) {
	host, set := metric.GetTag(telegrafHostTag)
	// skip if no host tag
	if set {
		key := buildMetricKey(host, metric.Name(), field.Key)

		_, set := sentMetrics[key]
		if !set {
			return buildMetainfo(token, host, metric, field), key
		}
	}

	return nil, ""
}

func buildMetainfo(token string, host string, metric telegraf.Metric, field *telegraf.Field) *MetricMetainfo {
	semType := getSematextMetricType(metric.Type())
	numericType := getSematextNumericType(field)

	if numericType == UnsupportedNumericType || numericType == Bool {
		return nil
	}

	label := fmt.Sprintf("%s.%s", metric.Name(), field.Key)

	return &MetricMetainfo{
		token:       token,
		name:        field.Key,
		namespace:   metric.Name(),
		semType:     semType,
		numericType: numericType,
		label:       label,
		description: "",
		host:        host,
	}
}

// Close clears the resources used by Metainfo processor
func (m *Metainfo) Close() {
	// close the channel to metainfo sender goroutines + close the sender
	close(m.sendChannel)
	m.sender.Close()
}

func getSematextMetricType(metricType telegraf.ValueType) SematextMetricType {
	var semType SematextMetricType
	switch metricType {
	case telegraf.Counter:
		semType = Counter
	default:
		semType = Gauge
	}

	return semType
}

func getSematextNumericType(field *telegraf.Field) NumericType {
	var numType NumericType
	switch field.Value.(type) {
	case float64:
		numType = Double
	case uint64:
		numType = Long
	case int64:
		numType = Long
	case bool:
		numType = Bool
	default:
		numType = UnsupportedNumericType
	}

	return numType
}

func buildMetricKey(host string, namespace string, name string) string {
	return fmt.Sprintf("%s-%s.%s", host, namespace, name)
}
