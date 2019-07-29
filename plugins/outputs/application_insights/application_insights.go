package application_insights

import (
	"fmt"
	"log"
	"math"
	"time"
	"unsafe"

	"github.com/Microsoft/ApplicationInsights-Go/appinsights"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type TelemetryTransmitter interface {
	Track(appinsights.Telemetry)
	Close() <-chan struct{}
}

type DiagnosticsMessageSubscriber interface {
	Subscribe(appinsights.DiagnosticsMessageHandler) appinsights.DiagnosticsMessageListener
}

type ApplicationInsights struct {
	InstrumentationKey      string
	Timeout                 internal.Duration
	EnableDiagnosticLogging bool
	ContextTagSources       map[string]string
	diagMsgSubscriber       DiagnosticsMessageSubscriber
	transmitter             TelemetryTransmitter
	diagMsgListener         appinsights.DiagnosticsMessageListener
}

const (
	Error   = "E! "
	Warning = "W! "
	Info    = "I! "
	Debug   = "D! "
)

var (
	sampleConfig = `
  ## Instrumentation key of the Application Insights resource.
  instrumentation_key = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxx"

  ## Timeout for closing (default: 5s).
  # timeout = "5s"

  ## Enable additional diagnostic logging.
  # enable_diagnostic_logging = false

  ## Context Tag Sources add Application Insights context tags to a tag value.
  ##
  ## For list of allowed context tag keys see:
  ## https://github.com/Microsoft/ApplicationInsights-Go/blob/master/appinsights/contracts/contexttagkeys.go
  # [outputs.application_insights.context_tag_sources]
  #   "ai.cloud.role" = "kubernetes_container_name"
  #   "ai.cloud.roleInstance" = "kubernetes_pod_name"
`
	is32Bit        bool
	is32BitChecked bool
)

func (a *ApplicationInsights) SampleConfig() string {
	return sampleConfig
}

func (a *ApplicationInsights) Description() string {
	return "Send metrics to Azure Application Insights"
}

func (a *ApplicationInsights) Connect() error {
	if a.InstrumentationKey == "" {
		return fmt.Errorf("Instrumentation key is required")
	}

	if a.transmitter == nil {
		a.transmitter = NewTransmitter(a.InstrumentationKey)
	}

	if a.EnableDiagnosticLogging && a.diagMsgSubscriber != nil {
		a.diagMsgListener = a.diagMsgSubscriber.Subscribe(func(msg string) error {
			logOutputMsg(Info, "%s", msg)
			return nil
		})
	}

	return nil
}

func (a *ApplicationInsights) Write(metrics []telegraf.Metric) error {
	for _, metric := range metrics {
		allMetricTelemetry := a.createTelemetry(metric)
		for _, telemetry := range allMetricTelemetry {
			a.transmitter.Track(telemetry)
		}
	}

	return nil
}

func (a *ApplicationInsights) Close() error {
	if a.diagMsgListener != nil {
		// We want to listen to diagnostic messages during closing
		// That is why we stop listening only after Close() ends (or a timeout occurs)
		defer a.diagMsgListener.Remove()
	}

	if a.transmitter == nil {
		return nil
	}

	select {
	case <-a.transmitter.Close():
		logOutputMsg(Info, "Closed")
	case <-time.After(a.Timeout.Duration):
		logOutputMsg(Warning, "Close operation timed out after %v", a.Timeout.Duration)
	}

	return nil
}

func (a *ApplicationInsights) createTelemetry(metric telegraf.Metric) []appinsights.Telemetry {
	aggregateTelemetry, usedFields := a.createAggregateMetricTelemetry(metric)
	if aggregateTelemetry != nil {
		telemetry := a.createTelemetryForUnusedFields(metric, usedFields)
		telemetry = append(telemetry, aggregateTelemetry)
		return telemetry
	}

	fields := metric.Fields()
	if len(fields) == 1 && metric.FieldList()[0].Key == "value" {
		// Just use metric name as the telemetry name
		telemetry := a.createSimpleMetricTelemetry(metric, "value", false)
		if telemetry != nil {
			return []appinsights.Telemetry{telemetry}
		} else {
			return nil
		}
	} else {
		// AppInsights does not support multi-dimensional metrics at the moment, so we need to disambiguate resulting telemetry
		// by adding field name as the telemetry name suffix
		retval := a.createTelemetryForUnusedFields(metric, nil)
		return retval
	}
}

func (a *ApplicationInsights) createSimpleMetricTelemetry(metric telegraf.Metric, fieldName string, useFieldNameInTelemetryName bool) *appinsights.MetricTelemetry {
	telemetryValue, err := getFloat64TelemetryPropertyValue([]string{fieldName}, metric, nil)
	if err != nil {
		return nil
	}

	var telemetryName string
	if useFieldNameInTelemetryName {
		telemetryName = metric.Name() + "_" + fieldName
	} else {
		telemetryName = metric.Name()
	}
	telemetry := appinsights.NewMetricTelemetry(telemetryName, telemetryValue)
	telemetry.Properties = metric.Tags()
	a.addContextTags(metric, telemetry)
	telemetry.Timestamp = metric.Time()
	return telemetry
}

func (a *ApplicationInsights) createAggregateMetricTelemetry(metric telegraf.Metric) (*appinsights.AggregateMetricTelemetry, []string) {
	usedFields := make([]string, 0, 6) // We will use up to 6 fields

	// Get the sum of all individual measurements(mandatory property)
	telemetryValue, err := getFloat64TelemetryPropertyValue([]string{"sum", "value"}, metric, &usedFields)
	if err != nil {
		return nil, nil
	}

	// Get the count of measurements (mandatory property)
	telemetryCount, err := getIntTelemetryPropertyValue([]string{"count", "samples"}, metric, &usedFields)
	if err != nil {
		return nil, nil
	}

	telemetry := appinsights.NewAggregateMetricTelemetry(metric.Name())
	telemetry.Value = telemetryValue
	telemetry.Count = telemetryCount
	telemetry.Properties = metric.Tags()
	a.addContextTags(metric, telemetry)
	telemetry.Timestamp = metric.Time()

	// We attempt to set min, max, variance and stddev fields but do not really care if they are not present--
	// they are not essential for aggregate metric.
	// By convention AppInsights prefers stddev over variance, so to be consistent, we test for stddev after testing for variance.
	telemetry.Min, _ = getFloat64TelemetryPropertyValue([]string{"min"}, metric, &usedFields)
	telemetry.Max, _ = getFloat64TelemetryPropertyValue([]string{"max"}, metric, &usedFields)
	telemetry.Variance, _ = getFloat64TelemetryPropertyValue([]string{"variance"}, metric, &usedFields)
	telemetry.StdDev, _ = getFloat64TelemetryPropertyValue([]string{"stddev"}, metric, &usedFields)

	return telemetry, usedFields
}

func (a *ApplicationInsights) createTelemetryForUnusedFields(metric telegraf.Metric, usedFields []string) []appinsights.Telemetry {
	fields := metric.Fields()
	retval := make([]appinsights.Telemetry, 0, len(fields))

	for fieldName := range fields {
		if contains(usedFields, fieldName) {
			continue
		}

		telemetry := a.createSimpleMetricTelemetry(metric, fieldName, true)
		if telemetry != nil {
			retval = append(retval, telemetry)
		}
	}

	return retval
}

func (a *ApplicationInsights) addContextTags(metric telegraf.Metric, telemetry appinsights.Telemetry) {
	for contextTagName, tagSourceName := range a.ContextTagSources {
		if contextTagValue, found := metric.GetTag(tagSourceName); found {
			telemetry.ContextTags()[contextTagName] = contextTagValue
		}
	}
}

func getFloat64TelemetryPropertyValue(
	candidateFields []string,
	metric telegraf.Metric,
	usedFields *[]string) (float64, error) {

	for _, fieldName := range candidateFields {
		fieldValue, found := metric.GetField(fieldName)
		if !found {
			continue
		}

		metricValue, err := toFloat64(fieldValue)
		if err != nil {
			continue
		}

		if usedFields != nil {
			*usedFields = append(*usedFields, fieldName)
		}

		return metricValue, nil
	}

	return 0.0, fmt.Errorf("No field from the candidate list was found in the metric")
}

func getIntTelemetryPropertyValue(
	candidateFields []string,
	metric telegraf.Metric,
	usedFields *[]string) (int, error) {

	for _, fieldName := range candidateFields {
		fieldValue, found := metric.GetField(fieldName)
		if !found {
			continue
		}

		metricValue, err := toInt(fieldValue)
		if err != nil {
			continue
		}

		if usedFields != nil {
			*usedFields = append(*usedFields, fieldName)
		}

		return metricValue, nil
	}

	return 0, fmt.Errorf("No field from the candidate list was found in the metric")
}

func contains(set []string, val string) bool {
	for _, elem := range set {
		if elem == val {
			return true
		}
	}

	return false
}

func toFloat64(value interface{}) (float64, error) {
	// Out of all Golang numerical types Telegraf only uses int64, unit64 and float64 for fields
	switch v := value.(type) {
	case int64:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float64:
		return v, nil
	}

	return 0.0, fmt.Errorf("[%s] cannot be converted to a float64 value", value)
}

func toInt(value interface{}) (int, error) {
	if !is32BitChecked {
		is32BitChecked = true
		var i int
		if unsafe.Sizeof(i) == 4 {
			is32Bit = true
		} else {
			is32Bit = false
		}
	}

	// Out of all Golang numerical types Telegraf only uses int64, unit64 and float64 for fields
	switch v := value.(type) {
	case uint64:
		if is32Bit {
			if v > math.MaxInt32 {
				return 0, fmt.Errorf("Value [%d] out of range of 32-bit integers", v)
			}
		} else {
			if v > math.MaxInt64 {
				return 0, fmt.Errorf("Value [%d] out of range of 64-bit integers", v)
			}
		}

		return int(v), nil

	case int64:
		if is32Bit {
			if v > math.MaxInt32 || v < math.MinInt32 {
				return 0, fmt.Errorf("Value [%d] out of range of 32-bit integers", v)
			}
		}

		return int(v), nil
	}

	return 0.0, fmt.Errorf("[%s] cannot be converted to an int value", value)
}

func logOutputMsg(level string, format string, v ...interface{}) {
	log.Printf(level+"[outputs.application_insights] "+format, v...)
}

func init() {
	outputs.Add("application_insights", func() telegraf.Output {
		return &ApplicationInsights{
			Timeout:           internal.Duration{Duration: time.Second * 5},
			diagMsgSubscriber: diagnosticsMessageSubscriber{},
			// It is very common to set Cloud.RoleName and Cloud.RoleInstance context properties, hence initial capacity of two
			ContextTagSources: make(map[string]string, 2),
		}
	})
}
