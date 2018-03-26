package application_insights

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"

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
	InstrumentationKey string
	Timeout            internal.Duration
	DiagMsgSubscriber  DiagnosticsMessageSubscriber
	Transmitter        TelemetryTransmitter
	diagMsgListener    appinsights.DiagnosticsMessageListener
}

var (
	sampleConfig = `
## Instrumentation key of the Application Insights resource.
instrumentationKey = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxx"

## Timeout on close. If not provided, will default to 5s. 0s means no timeout (not recommended).
# timeout = "5s"
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
	if a.Timeout.Duration == 0 {
		a.Timeout = internal.Duration{Duration: time.Second * 5}
	}

	if a.Transmitter == nil {
		a.useDefaultTransmitter()
	}

	if a.Transmitter != nil && a.DiagMsgSubscriber != nil {
		a.diagMsgListener = a.DiagMsgSubscriber.Subscribe(func(msg string) error {
			logOutputMsg("%s", msg)
			return nil
		})
	}

	return nil
}

func (a *ApplicationInsights) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 || a.Transmitter == nil || a.InstrumentationKey == "" {
		return nil
	}

	for _, metric := range metrics {
		if canBeTreatedAsAggregate(metric) {
			if telemetry := createAggregateMetricTelemetry(metric); telemetry != nil {
				a.Transmitter.Track(telemetry)
				continue
			}
		}

		if telemetry := createSimpleMetricTelemetry(metric); telemetry != nil {
			a.Transmitter.Track(telemetry)
		}
	}

	return nil
}

func (a *ApplicationInsights) Close() error {
	if a.Transmitter == nil {
		return nil
	}

	select {
	case <-a.Transmitter.Close():
		logOutputMsg("Closed")
	case <-time.After(a.Timeout.Duration):
		logOutputMsg("Close operation timed out after %v", a.Timeout.Duration)
	}

	if a.diagMsgListener != nil {
		// We want to listen to diagnostic messages during closing
		// That is why we stop listening only after Close() ends (or a timeout occurs)
		a.diagMsgListener.Remove()
	}

	return nil
}

func init() {
	outputs.Add("application_insights", func() telegraf.Output {
		return &ApplicationInsights{
			DiagMsgSubscriber: appinsightsDiagnosticsMessageSubscriber{},
		}
	})
}

func (a *ApplicationInsights) useDefaultTransmitter() {
	// Env variable, if present, overrides what is in the configuration
	iKey := os.Getenv("APPINSIGHTS_INSTRUMENTATIONKEY")
	if iKey != "" {
		a.InstrumentationKey = iKey
	}

	if a.InstrumentationKey != "" {
		a.Transmitter = NewAppinsightsTransmitter(a.InstrumentationKey)
	}
}

func canBeTreatedAsAggregate(metric telegraf.Metric) bool {
	// If we can identify the Sum and Count properties, then we are good
	hasSum := metric.HasField("value") || metric.HasField("sum")
	hasCount := metric.HasField("count") || metric.HasField("samples")
	return hasSum && hasCount
}

func createSimpleMetricTelemetry(metric telegraf.Metric) *appinsights.MetricTelemetry {
	telemetry := appinsights.NewMetricTelemetry(metric.Name(), 0.0)
	telemetry.Properties = metric.Tags()
	telemetry.Timestamp = metric.Time()

	err := setFloat64TelemetryProperty(&(telemetry.Value), []string{"value"}, metric, nil)
	if err == nil {
		addRemainingMeasurements(telemetry, metric, []string{"value"})
	} else {
		// Either the metric does not have a "value" field, or it was not convertable to float64
		// Heuristics: if the metric has just one field that is convertible to float64, take it as value
		// Otherwise give up
		convertedFieldCount := 0
		var metricValue float64

		for _, value := range metric.Fields() {
			convertedValue, err := toFloat64(value)
			if err == nil {
				convertedFieldCount++
				metricValue = convertedValue
			}
		}

		if convertedFieldCount != 1 {
			return nil
		} else {
			telemetry.Value = metricValue
		}
	}

	return telemetry
}

func createAggregateMetricTelemetry(metric telegraf.Metric) *appinsights.AggregateMetricTelemetry {
	telemetry := appinsights.NewAggregateMetricTelemetry(metric.Name())
	telemetry.Properties = metric.Tags()
	telemetry.Timestamp = metric.Time()

	var usedFields []string
	var err error

	// Get the sum of all individual measurements(mandatory property)
	err = setFloat64TelemetryProperty(&(telemetry.Value), []string{"sum", "value"}, metric, &usedFields)
	if err != nil {
		return nil
	}

	// Get the count of measurements (mandatory property)
	err = setIntTelemetryProperty(&(telemetry.Count), []string{"count", "samples"}, metric, &usedFields)
	if err != nil {
		return nil
	}

	// We attempt to set Min, Max, StdDev and Variance fields but do not really care if they are not present
	setFloat64TelemetryProperty(&(telemetry.Min), []string{"min"}, metric, &usedFields)
	setFloat64TelemetryProperty(&(telemetry.Max), []string{"max"}, metric, &usedFields)
	setFloat64TelemetryProperty(&(telemetry.StdDev), []string{"stddev"}, metric, &usedFields)
	setFloat64TelemetryProperty(&(telemetry.Variance), []string{"variance"}, metric, &usedFields)

	addRemainingMeasurements(telemetry, metric, usedFields)

	return telemetry
}

func setFloat64TelemetryProperty(
	telemetryProperty *float64,
	candidateFields []string,
	metric telegraf.Metric,
	usedFields *[]string) error {

	for _, fieldName := range candidateFields {
		if metric.HasField(fieldName) {
			metricValue, err := toFloat64(metric.Fields()[fieldName])
			if err != nil {
				return err
			}

			*telemetryProperty = metricValue

			if usedFields != nil {
				*usedFields = append(*usedFields, fieldName)
			}

			return nil
		}
	}

	return fmt.Errorf("No field from the candidate list was found in the metric")
}

func setIntTelemetryProperty(
	telemetryProperty *int,
	candidateFields []string,
	metric telegraf.Metric,
	usedFields *[]string) error {

	for _, fieldName := range candidateFields {
		if metric.HasField(fieldName) {
			metricValue, err := toInt(metric.Fields()[fieldName])
			if err != nil {
				return err
			}

			*telemetryProperty = metricValue

			if usedFields != nil {
				*usedFields = append(*usedFields, fieldName)
			}

			return nil
		}
	}

	return fmt.Errorf("No field from the candidate list was found in the metric")
}

func addRemainingMeasurements(telemetry appinsights.Telemetry, metric telegraf.Metric, usedFields []string) {
	properties := telemetry.GetProperties()

	for fieldName, fieldValue := range metric.Fields() {
		if contains(usedFields, fieldName) {
			continue
		}

		float64Value, err := toFloat64(fieldValue)
		if err != nil {
			continue
		}

		// We would like to add the field to telemetry.Measurements here
		// But because of issue https://github.com/Microsoft/ApplicationInsights-Go/issues/12
		// this is currently impossible. So we are just going to add the measurement as a property instead.
		// This is what we want to be able to do, eventually:
		//     telemetry.GetMeasurements()[fieldName] = float64Value
		// (plus handle name conflics)
		addUniquelyNamedProperty(properties, fieldName, fmt.Sprintf("%g", float64Value))
	}
}

func contains(set []string, val string) bool {
	for _, elem := range set {
		if elem == val {
			return true
		}
	}

	return false
}

func addUniquelyNamedProperty(properties map[string]string, name string, value string) {
	if _, found := properties[name]; !found {
		properties[name] = value
	} else {
		suffix := 1
		for {
			_, found = properties[name+"_"+strconv.Itoa(suffix)]
			if !found {
				break
			}
			suffix++
		}
		properties[name+"_"+strconv.Itoa(suffix)] = value
	}
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
		i := int(-1)
		j := uint(i)
		is32Bit = (j == math.MaxUint32)
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
			if v > math.MaxInt64 {
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
