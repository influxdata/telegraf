package application_insights

import (
	"fmt"
	"math"
	"time"
	"unsafe"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
)

type TelemetryTransmitter interface {
	Track(appinsights.Telemetry)
	Close() <-chan struct{}
}

type DiagnosticsMessageSubscriber interface {
	Subscribe(appinsights.DiagnosticsMessageHandler) appinsights.DiagnosticsMessageListener
}

type ApplicationInsights struct {
	InstrumentationKey      string            `toml:"instrumentation_key"`
	EndpointURL             string            `toml:"endpoint_url"`
	Timeout                 config.Duration   `toml:"timeout"`
	EnableDiagnosticLogging bool              `toml:"enable_diagnostic_logging"`
	ContextTagSources       map[string]string `toml:"context_tag_sources"`
	Log                     telegraf.Logger   `toml:"-"`

	diagMsgSubscriber DiagnosticsMessageSubscriber
	transmitter       TelemetryTransmitter
	diagMsgListener   appinsights.DiagnosticsMessageListener
}

var (
	is32Bit        bool
	is32BitChecked bool
)

func (a *ApplicationInsights) Connect() error {
	if a.InstrumentationKey == "" {
		return fmt.Errorf("instrumentation key is required")
	}

	if a.transmitter == nil {
		a.transmitter = NewTransmitter(a.InstrumentationKey, a.EndpointURL)
	}

	if a.EnableDiagnosticLogging && a.diagMsgSubscriber != nil {
		a.diagMsgListener = a.diagMsgSubscriber.Subscribe(func(msg string) error {
			a.Log.Info(msg)
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
		a.Log.Info("Closed")
	case <-time.After(time.Duration(a.Timeout)):
		a.Log.Warnf("Close operation timed out after %v", time.Duration(a.Timeout))
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
		}
		return nil
	}
	// AppInsights does not support multi-dimensional metrics at the moment, so we need to disambiguate resulting telemetry
	// by adding field name as the telemetry name suffix
	return a.createTelemetryForUnusedFields(metric, nil)
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
	usedFields *[]string,
) (float64, error) {
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

	return 0.0, fmt.Errorf("no field from the candidate list was found in the metric")
}

func getIntTelemetryPropertyValue(
	candidateFields []string,
	metric telegraf.Metric,
	usedFields *[]string,
) (int, error) {
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

	return 0, fmt.Errorf("no field from the candidate list was found in the metric")
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
				return 0, fmt.Errorf("value [%d] out of range of 32-bit integers", v)
			}
		} else {
			if v > math.MaxInt64 {
				return 0, fmt.Errorf("value [%d] out of range of 64-bit integers", v)
			}
		}

		return int(v), nil

	case int64:
		if is32Bit {
			if v > math.MaxInt32 || v < math.MinInt32 {
				return 0, fmt.Errorf("value [%d] out of range of 32-bit integers", v)
			}
		}

		return int(v), nil
	}

	return 0.0, fmt.Errorf("[%s] cannot be converted to an int value", value)
}

func init() {
	outputs.Add("application_insights", func() telegraf.Output {
		return &ApplicationInsights{
			Timeout:           config.Duration(time.Second * 5),
			diagMsgSubscriber: diagnosticsMessageSubscriber{},
			// It is very common to set Cloud.RoleName and Cloud.RoleInstance context properties, hence initial capacity of two
			ContextTagSources: make(map[string]string, 2),
		}
	})
}
