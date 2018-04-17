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

var (
	sampleConfig = `
## Instrumentation key of the Application Insights resource.
instrumentationKey = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxx"

## Timeout on close. If not provided, will default to 5s. 0s means no timeout (not recommended).
# timeout = "5s"

## Determines whether diagnostic logging (for Application Insights endpoint traffic) is enabled. Default is false.
# enableDiagnosticLogging = "true"

## ContextTagSources dictionary instructs the Application Insights plugin to set Application Insights context tags using metric properties.
## In this dictionary keys are Application Insights context tags to set, and values are names of metric properties to use as source of data.
## For example:
# [outputs.application_insights.contextTagSources]
# "ai.cloud.role" = "kubernetes_container_name"
# "ai.cloud.roleInstance" = "kubernetes_pod_name"
## will set the ai.cloud.role context tag to the value of kubernetes_container_name property (if present), 
## and the ai.cloud.roleInstance context tag to the value of kubernetes_pod_name property.
## For list of all context tag keys see https://github.com/Microsoft/ApplicationInsights-Go/blob/master/appinsights/contracts/contexttagkeys.go
`
	is32Bit        bool
	is32BitChecked bool
)

func (a *ApplicationInsights) SampleConfig() string {
	return sampleConfig
}

func (a *ApplicationInsights) Description() string {
	return "Send telegraf metrics to Azure Application Insights"
}

func (a *ApplicationInsights) Connect() error {
	if a.transmitter == nil && a.InstrumentationKey != "" {
		a.transmitter = NewAppinsightsTransmitter(a.InstrumentationKey)
	}

	if a.EnableDiagnosticLogging && a.diagMsgSubscriber != nil {
		a.diagMsgListener = a.diagMsgSubscriber.Subscribe(func(msg string) error {
			logOutputMsg("%s", msg)
			return nil
		})
	}

	return nil
}

func (a *ApplicationInsights) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 || a.transmitter == nil || a.InstrumentationKey == "" {
		return nil
	}

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
		logOutputMsg("Closed")
	case <-time.After(a.Timeout.Duration):
		logOutputMsg("Close operation timed out after %v", a.Timeout.Duration)
	}

	return nil
}

func init() {
	outputs.Add("application_insights", func() telegraf.Output {
		return &ApplicationInsights{
			Timeout:           internal.Duration{Duration: time.Second * 5},
			diagMsgSubscriber: appinsightsDiagnosticsMessageSubscriber{},
			// It is very common to set Cloud.RoleName and Cloud.RoleInstance context properties, hence initial capacity of two
			ContextTagSources: make(map[string]string, 2),
		}
	})
}

func canBeTreatedAsAggregate(metric telegraf.Metric) bool {
	// If we can identify the Sum and Count properties, then we are good
	hasSum := metric.HasField("value") || metric.HasField("sum")
	hasCount := metric.HasField("count") || metric.HasField("samples")
	return hasSum && hasCount
}

func (a *ApplicationInsights) createTelemetry(metric telegraf.Metric) []appinsights.Telemetry {
	if canBeTreatedAsAggregate(metric) {
		aggregateTelemetry, usedFields := a.createAggregateMetricTelemetry(metric)
		if aggregateTelemetry != nil {
			remainingTelemetry := a.createTelemetryForUnusedFields(metric, usedFields)
			retval := make([]appinsights.Telemetry, 1, 1+len(remainingTelemetry))
			retval[0] = aggregateTelemetry
			retval = append(retval, remainingTelemetry...)
			return retval
		}
	}

	fields := metric.Fields()
	if len(fields) == 1 {
		// Just use metric name as the telemetry name
		telemetry := a.createSimpleMetricTelemetry(metric, metric.FieldList()[0].Key, false)
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
	var telemetryValue float64
	err := setFloat64TelemetryProperty(&telemetryValue, []string{fieldName}, metric, nil)
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
	var telemetryValue float64
	var telemetryCount int
	usedFields := make([]string, 0, 6) // We will use up to 6 fields
	var err error

	// Get the sum of all individual measurements(mandatory property)
	err = setFloat64TelemetryProperty(&telemetryValue, []string{"sum", "value"}, metric, &usedFields)
	if err != nil {
		return nil, nil
	}

	// Get the count of measurements (mandatory property)
	err = setIntTelemetryProperty(&telemetryCount, []string{"count", "samples"}, metric, &usedFields)
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
	setFloat64TelemetryProperty(&(telemetry.Min), []string{"min"}, metric, &usedFields)
	setFloat64TelemetryProperty(&(telemetry.Max), []string{"max"}, metric, &usedFields)
	setFloat64TelemetryProperty(&(telemetry.Variance), []string{"variance"}, metric, &usedFields)
	setFloat64TelemetryProperty(&(telemetry.StdDev), []string{"stddev"}, metric, &usedFields)

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

func setFloat64TelemetryProperty(
	telemetryProperty *float64,
	candidateFields []string,
	metric telegraf.Metric,
	usedFields *[]string) error {

	for _, fieldName := range candidateFields {
		fieldValue, found := metric.GetField(fieldName)
		if !found {
			continue
		}

		metricValue, err := toFloat64(fieldValue)
		if err != nil {
			continue
		}

		*telemetryProperty = metricValue

		if usedFields != nil {
			*usedFields = append(*usedFields, fieldName)
		}

		return nil
	}

	return fmt.Errorf("No field from the candidate list was found in the metric")
}

func setIntTelemetryProperty(
	telemetryProperty *int,
	candidateFields []string,
	metric telegraf.Metric,
	usedFields *[]string) error {

	for _, fieldName := range candidateFields {
		fieldValue, found := metric.GetField(fieldName)
		if !found {
			continue
		}

		metricValue, err := toInt(fieldValue)
		if err != nil {
			continue
		}

		*telemetryProperty = metricValue

		if usedFields != nil {
			*usedFields = append(*usedFields, fieldName)
		}

		return nil
	}

	return fmt.Errorf("No field from the candidate list was found in the metric")
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
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float32:
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

	switch v := value.(type) {
	case int:
		return v, nil
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case uint8:
		return int(v), nil
	case uint16:
		return int(v), nil

	case uint:
		if is32Bit {
			if v > math.MaxInt32 {
				return 0, fmt.Errorf("Value [%d] out of range of 32-bit integers", v)
			}
		} else {
			if uint64(v) > math.MaxInt64 {
				return 0, fmt.Errorf("Value [%d] out of range of 64-bit integers", v)
			}
		}

		return int(v), nil

	case uint32:
		if is32Bit {
			if v > math.MaxInt32 {
				return 0, fmt.Errorf("Value [%d] out of range of 32-bit integers", v)
			}
		}

		return int(v), nil

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

func logOutputMsg(format string, v ...interface{}) {
	log.Printf("Output [application_insights] "+format, v...)
}
