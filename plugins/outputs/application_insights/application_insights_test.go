package application_insights

import (
	"math"
	"testing"
	"time"

	"github.com/microsoft/ApplicationInsights-Go/appinsights"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/outputs/application_insights/mocks"
	"github.com/influxdata/telegraf/testutil"
)

func TestConnectFailsIfNoIkey(t *testing.T) {
	transmitter := new(mocks.Transmitter)
	transmitter.On("Close").Return(closed)

	ai := ApplicationInsights{
		transmitter: transmitter,
		// Very long timeout to ensure we do not rely on timeouts for closing the transmitter
		Timeout: config.Duration(time.Hour),
		Log:     testutil.Logger{},
	}

	err := ai.Connect()
	require.Error(t, err)
}

func TestOutputCloseTimesOut(t *testing.T) {
	transmitter := new(mocks.Transmitter)
	transmitter.On("Close").Return(unfinished)

	ai := ApplicationInsights{
		transmitter: transmitter,
		Timeout:     config.Duration(time.Millisecond * 50),
		Log:         testutil.Logger{},
	}

	err := ai.Close()
	require.NoError(t, err)
	transmitter.AssertCalled(t, "Close")
}

func TestCloseRemovesDiagMsgListener(t *testing.T) {
	transmitter := new(mocks.Transmitter)
	transmitter.On("Close").Return(closed)

	diagMsgListener := new(mocks.DiagnosticsMessageListener)
	diagMsgListener.On("Remove")

	diagMsgSubscriber := new(mocks.DiagnosticsMessageSubscriber)
	diagMsgSubscriber.
		On("Subscribe", mock.AnythingOfType("appinsights.DiagnosticsMessageHandler")).
		Return(diagMsgListener)

	ai := ApplicationInsights{
		transmitter:             transmitter,
		Timeout:                 config.Duration(time.Hour),
		EnableDiagnosticLogging: true,
		diagMsgSubscriber:       diagMsgSubscriber,
		InstrumentationKey:      "1234", // Fake, but necessary to enable tracking
		Log:                     testutil.Logger{},
	}

	err := ai.Connect()
	require.NoError(t, err)
	diagMsgSubscriber.AssertCalled(t, "Subscribe", mock.AnythingOfType("appinsights.DiagnosticsMessageHandler"))

	err = ai.Close()
	require.NoError(t, err)
	transmitter.AssertCalled(t, "Close")
	diagMsgListener.AssertCalled(t, "Remove")
}

func TestAggregateMetricCreated(t *testing.T) {
	tests := []struct {
		name                        string
		fields                      map[string]interface{}
		valueField                  string
		countField                  string
		additionalMetricValueFields []string
	}{
		{"value and count", map[string]interface{}{"value": 16.5, "count": 23}, "value", "count", nil},
		{"value and samples", map[string]interface{}{"value": 16.5, "samples": 23}, "value", "samples", nil},
		{"sum and count", map[string]interface{}{"sum": 16.5, "count": 23}, "sum", "count", nil},
		{"sum and samples", map[string]interface{}{"samples": 23, "sum": 16.5}, "sum", "samples", nil},
		{"value and count, sum is wrong type", map[string]interface{}{"sum": "J23", "value": 16.5, "count": 23}, "value", "count", nil},
		{
			"with aggregates",
			map[string]interface{}{
				"value":  16.5,
				"count":  23,
				"min":    -2.1,
				"max":    34,
				"stddev": 3.4,
			},
			"value",
			"count",
			nil,
		},
		{
			"some aggregates with invalid values",
			map[string]interface{}{
				"value": 16.5,
				"count": 23,
				"min":   "min",
				"max":   []float64{3.4, 5.6},
				"stddev": struct {
					name  string
					value float64
				}{"delta", 7.0},
			},
			"value",
			"count",
			nil,
		},
		{
			"aggregate with additional fields",
			map[string]interface{}{"value": 16.5, "samples": 23, "alpha": -34e12, "bravo": -3, "charlie": "charlie"},
			"value",
			"samples",
			[]string{"alpha", "bravo"},
		},
	}

	for _, tt := range tests {
		tf := func(t *testing.T) {
			now := time.Now().UTC()

			transmitter := new(mocks.Transmitter)
			transmitter.On("Track", mock.Anything)
			metricName := "ShouldBeAggregateMetric"

			m := metric.New(
				metricName,
				nil, // tags
				tt.fields,
				now,
			)

			ai := ApplicationInsights{
				transmitter:        transmitter,
				InstrumentationKey: "1234", // Fake, but necessary to enable tracking
				Log:                testutil.Logger{},
			}

			err := ai.Connect()
			require.NoError(t, err)

			mSet := []telegraf.Metric{m}
			err = ai.Write(mSet)
			require.NoError(t, err)
			transmitter.AssertNumberOfCalls(t, "Track", 1+len(tt.additionalMetricValueFields))
			var pAggregateTelemetry *appinsights.AggregateMetricTelemetry
			require.IsType(t, pAggregateTelemetry, transmitter.Calls[len(transmitter.Calls)-1].Arguments.Get(0), "Expected last telemetry to be AggregateMetricTelemetry")
			aggregateTelemetry := transmitter.Calls[len(transmitter.Calls)-1].Arguments.Get(0).(*appinsights.AggregateMetricTelemetry)
			verifyAggregateTelemetry(t, m, tt.valueField, tt.countField, aggregateTelemetry)

			verifyAdditionalTelemetry(t, m, transmitter, tt.additionalMetricValueFields, metricName)
		}

		t.Run(tt.name, tf)
	}
}

func TestSimpleMetricCreated(t *testing.T) {
	tests := []struct {
		name                        string
		fields                      map[string]interface{}
		primaryMetricValueField     string
		additionalMetricValueFields []string
	}{
		{"just a single value field", map[string]interface{}{"value": 16.5}, "value", nil},
		{"single field not named value", map[string]interface{}{"first": 32.9}, "first", nil},
		{"value but no count", map[string]interface{}{"value": 16.5, "other": "bulba"}, "", []string{"value"}},
		{"count but no value", map[string]interface{}{"v1": "v1Val", "count": 23}, "", []string{"count"}},
		{"neither value nor count", map[string]interface{}{"v1": "alpha", "v2": 45.8}, "", []string{"v2"}},
		{"value is of wrong type", map[string]interface{}{"value": "alpha", "count": 15}, "", []string{"count"}},
		{"count is of wrong type", map[string]interface{}{"value": 23.77, "count": 7.5}, "", []string{"count", "value"}},
		{"count is out of range", map[string]interface{}{"value": -98.45e4, "count": math.MaxUint64 - uint64(20)}, "", []string{"value", "count"}},
		{"several additional fields", map[string]interface{}{"alpha": 10, "bravo": "bravo", "charlie": 30, "delta": 40.7}, "", []string{"alpha", "charlie", "delta"}},
	}

	for _, tt := range tests {
		tf := func(t *testing.T) {
			now := time.Now().UTC()

			transmitter := new(mocks.Transmitter)
			transmitter.On("Track", mock.Anything)
			metricName := "ShouldBeSimpleMetric"

			m := metric.New(
				metricName,
				nil, // tags
				tt.fields,
				now,
			)

			ai := ApplicationInsights{
				transmitter:        transmitter,
				InstrumentationKey: "1234", // Fake, but necessary to enable tracking
				Log:                testutil.Logger{},
			}

			err := ai.Connect()
			require.NoError(t, err)

			mSet := []telegraf.Metric{m}
			err = ai.Write(mSet)
			require.NoError(t, err)

			expectedNumberOfCalls := len(tt.additionalMetricValueFields)
			if tt.primaryMetricValueField != "" {
				expectedNumberOfCalls++
			}

			transmitter.AssertNumberOfCalls(t, "Track", expectedNumberOfCalls)
			if tt.primaryMetricValueField != "" {
				var pMetricTelemetry *appinsights.MetricTelemetry
				require.IsType(t, pMetricTelemetry, transmitter.Calls[0].Arguments.Get(0), "First created telemetry should be simple MetricTelemetry")
				metricTelemetry := transmitter.Calls[0].Arguments.Get(0).(*appinsights.MetricTelemetry)

				var expectedTelemetryName string
				if tt.primaryMetricValueField == "value" {
					expectedTelemetryName = m.Name()
				} else {
					expectedTelemetryName = m.Name() + "_" + tt.primaryMetricValueField
				}
				verifySimpleTelemetry(t, m, tt.primaryMetricValueField, expectedTelemetryName, metricTelemetry)
			}

			verifyAdditionalTelemetry(t, m, transmitter, tt.additionalMetricValueFields, metricName)
		}

		t.Run(tt.name, tf)
	}
}

func TestTagsAppliedToTelemetry(t *testing.T) {
	tests := []struct {
		name              string
		fields            map[string]interface{}
		tags              map[string]string
		metricValueFields []string
	}{
		{
			"value but no count",
			map[string]interface{}{"value": 16.5, "alpha": 3.5, "bravo": 17},
			map[string]string{"alpha": "a tag is not a field", "charlie": "charlie"},
			[]string{"value", "alpha", "bravo"},
		},
	}

	for _, tt := range tests {
		tf := func(t *testing.T) {
			now := time.Now().UTC()

			transmitter := new(mocks.Transmitter)
			transmitter.On("Track", mock.Anything)
			metricName := "ShouldBeSimpleMetric"

			m := metric.New(
				metricName,
				tt.tags,
				tt.fields,
				now,
			)

			ai := ApplicationInsights{
				transmitter:        transmitter,
				InstrumentationKey: "1234", // Fake, but necessary to enable tracking
				Log:                testutil.Logger{},
			}

			err := ai.Connect()
			require.NoError(t, err)

			mSet := []telegraf.Metric{m}
			err = ai.Write(mSet)
			require.NoError(t, err)
			transmitter.AssertNumberOfCalls(t, "Track", len(tt.metricValueFields))
			transmitter.AssertCalled(t, "Track", mock.AnythingOfType("*appinsights.MetricTelemetry"))

			// Will verify that all original tags are present in telemetry.Properties map
			verifyAdditionalTelemetry(t, m, transmitter, tt.metricValueFields, metricName)
		}

		t.Run(tt.name, tf)
	}
}

func TestContextTagsSetOnSimpleTelemetry(t *testing.T) {
	now := time.Now().UTC()

	transmitter := new(mocks.Transmitter)
	transmitter.On("Track", mock.Anything)

	m := metric.New(
		"SimpleMetric",
		map[string]string{"kubernetes_container_name": "atcsvc", "kubernetes_pod_name": "bunkie17554"},
		map[string]interface{}{"value": 23.0},
		now,
	)

	ai := ApplicationInsights{
		transmitter:        transmitter,
		InstrumentationKey: "1234", // Fake, but necessary to enable tracking
		ContextTagSources: map[string]string{
			"ai.cloud.role":         "kubernetes_container_name",
			"ai.cloud.roleInstance": "kubernetes_pod_name",
			"ai.user.id":            "nonexistent",
		},
		Log: testutil.Logger{},
	}

	err := ai.Connect()
	require.NoError(t, err)

	mSet := []telegraf.Metric{m}
	err = ai.Write(mSet)
	require.NoError(t, err)
	transmitter.AssertNumberOfCalls(t, "Track", 1)
	metricTelemetry := transmitter.Calls[0].Arguments.Get(0).(*appinsights.MetricTelemetry)
	cloudTags := metricTelemetry.Tags.Cloud()
	require.Equal(t, "atcsvc", cloudTags.GetRole())
	require.Equal(t, "bunkie17554", cloudTags.GetRoleInstance())
}

func TestContextTagsSetOnAggregateTelemetry(t *testing.T) {
	now := time.Now().UTC()

	transmitter := new(mocks.Transmitter)
	transmitter.On("Track", mock.Anything)

	m := metric.New(
		"AggregateMetric",
		map[string]string{"kubernetes_container_name": "atcsvc", "kubernetes_pod_name": "bunkie17554"},
		map[string]interface{}{"value": 23.0, "count": 5},
		now,
	)

	ai := ApplicationInsights{
		transmitter:        transmitter,
		InstrumentationKey: "1234", // Fake, but necessary to enable tracking
		ContextTagSources: map[string]string{
			"ai.cloud.role":         "kubernetes_container_name",
			"ai.cloud.roleInstance": "kubernetes_pod_name",
			"ai.user.id":            "nonexistent",
		},
		Log: testutil.Logger{},
	}

	err := ai.Connect()
	require.NoError(t, err)

	mSet := []telegraf.Metric{m}
	err = ai.Write(mSet)
	require.NoError(t, err)
	transmitter.AssertNumberOfCalls(t, "Track", 1)
	metricTelemetry := transmitter.Calls[0].Arguments.Get(0).(*appinsights.AggregateMetricTelemetry)
	cloudTags := metricTelemetry.Tags.Cloud()
	require.Equal(t, "atcsvc", cloudTags.GetRole())
	require.Equal(t, "bunkie17554", cloudTags.GetRoleInstance())
}

func closed() <-chan struct{} {
	closed := make(chan struct{})
	close(closed)
	return closed
}

func unfinished() <-chan struct{} {
	unfinished := make(chan struct{})
	return unfinished
}

func verifyAggregateTelemetry(
	t *testing.T,
	m telegraf.Metric,
	valueField string,
	countField string,
	telemetry *appinsights.AggregateMetricTelemetry,
) {
	verifyAggregateField := func(fieldName string, telemetryValue float64) {
		metricRawFieldValue, found := m.Fields()[fieldName]
		if !found {
			return
		}

		if _, err := toFloat64(metricRawFieldValue); err == nil {
			require.EqualValues(t, metricRawFieldValue, telemetryValue, "Telemetry property %s does not match the metric field", fieldName)
		}
	}
	require.Equal(t, m.Name(), telemetry.Name, "Telemetry name should be the same as metric name")
	require.EqualValues(t, m.Fields()[valueField], telemetry.Value, "Telemetry value does not match metric value field")
	require.EqualValues(t, m.Fields()[countField], telemetry.Count, "Telemetry sample count does not mach metric sample count field")
	verifyAggregateField("min", telemetry.Min)
	verifyAggregateField("max", telemetry.Max)
	verifyAggregateField("stdev", telemetry.StdDev)
	verifyAggregateField("variance", telemetry.Variance)
	require.Equal(t, m.Time(), telemetry.Timestamp, "Telemetry and metric timestamps do not match")
	assertMapContains(t, m.Tags(), telemetry.Properties)
}

func verifySimpleTelemetry(
	t *testing.T,
	m telegraf.Metric,
	valueField string,
	expectedTelemetryName string,
	telemetry *appinsights.MetricTelemetry,
) {
	require.Equal(t, expectedTelemetryName, telemetry.Name, "Telemetry name is not what was expected")
	require.EqualValues(t, m.Fields()[valueField], telemetry.Value, "Telemetry value does not match metric value field")
	require.Equal(t, m.Time(), telemetry.Timestamp, "Telemetry and metric timestamps do not match")
	assertMapContains(t, m.Tags(), telemetry.Properties)
}

func verifyAdditionalTelemetry(
	t *testing.T,
	m telegraf.Metric,
	transmitter *mocks.Transmitter,
	additionalMetricValueFields []string,
	telemetryNamePrefix string,
) {
	for _, fieldName := range additionalMetricValueFields {
		expectedTelemetryName := telemetryNamePrefix + "_" + fieldName
		telemetry := findTransmittedTelemetry(transmitter, expectedTelemetryName)
		require.NotNil(t, telemetry, "Expected telemetry named %s to be created, but could not find it", expectedTelemetryName)
		if telemetry != nil {
			verifySimpleTelemetry(t, m, fieldName, expectedTelemetryName, telemetry)
		}
	}
}

func findTransmittedTelemetry(transmitter *mocks.Transmitter, telemetryName string) *appinsights.MetricTelemetry {
	for _, call := range transmitter.Calls {
		telemetry, isMetricTelemetry := call.Arguments.Get(0).(*appinsights.MetricTelemetry)
		if isMetricTelemetry && telemetry.Name == telemetryName {
			return telemetry
		}
	}

	return nil
}

func assertMapContains(t *testing.T, expected, actual map[string]string) {
	if expected == nil && actual == nil {
		return
	}

	require.NotNil(t, expected, "Maps not equal: expected is nil but actual is not")
	require.NotNil(t, actual, "Maps not equal: actual is nil but expected is not")

	for k, v := range expected {
		av, ok := actual[k]
		require.True(t, ok, "Actual map does not contain a value for key '%s'", k)
		require.Equal(t, v, av, "The expected value for key '%s' is '%s' but the actual value is '%s", k, v, av)
	}
}
