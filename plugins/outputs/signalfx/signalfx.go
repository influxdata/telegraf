package signalfx

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/signalfx/golib/datapoint"
	"github.com/signalfx/golib/event"
	"github.com/signalfx/golib/sfxclient"
)

/*SignalFx plugin context*/
type SignalFx struct {
	APIToken           string
	DatapointIngestURL string
	EventIngestURL     string
	Exclude            []string
	ctx                context.Context
	client             *sfxclient.HTTPSink
}

var sampleConfig = `
    ## SignalFx API Token
    APIToken = "my-secret-key" # required.

    ## Ingest URL
    DatapointIngestURL = "https://ingest.signalfx.com/v2/datapoint"
    EventIngestURL = "https://ingest.signalfx.com/v2/event"
    
    ## Exclude metrics by metric name
    Exclude = ["system.uptime_format", ""]
`

// NewSignalFx - returns a new context for the SignalFx output plugin
func NewSignalFx() *SignalFx {
	return &SignalFx{
		APIToken:           "",
		DatapointIngestURL: "https://ingest.signalfx.com/v2/datapoint",
		EventIngestURL:     "https://ingest.signalfx.com/v2/event",
		Exclude:            []string{""},
	}
}

/*Description returns a description for the plugin*/
func (s *SignalFx) Description() string {
	return "Send metrics to SignalFx"
}

/*SampleConfig returns the sample configuration for the plugin*/
func (s *SignalFx) SampleConfig() string {
	return sampleConfig
}

/*Connect establishes a connection to SignalFx*/
func (s *SignalFx) Connect() error {
	// Make a connection to the URL here
	s.client = sfxclient.NewHTTPSink()
	s.client.AuthToken = s.APIToken
	s.client.DatapointEndpoint = s.DatapointIngestURL
	s.client.EventEndpoint = s.EventIngestURL
	s.ctx = context.Background()
	return nil
}

/*Close closes the connection to SignalFx*/
func (s *SignalFx) Close() error {
	s.ctx.Done()
	s.client = nil
	return nil
}

/*Determine and assign a datapoint metric type based on telegraf metric type*/
func getMetricType(metric telegraf.Metric) (datapoint.MetricType, error) {
	var err error
	var metricType datapoint.MetricType
	switch metric.Type() {
	case telegraf.Counter:
		metricType = datapoint.Counter
		if metric.Name() == "mem" {
			metricType = datapoint.Gauge
		}
	case telegraf.Gauge:
		metricType = datapoint.Gauge
	case telegraf.Untyped:
		// For untyped metrics, default to Gauge
		metricType = datapoint.Gauge
	default:
		err = fmt.Errorf("unable to determine metric type")
	}
	return metricType, err
}

/*Determine and assign a datapoint metric type based on telegraf metric type*/
func getMetricTypeAsString(metric telegraf.Metric) (string, error) {
	var err error
	var metricType string
	switch metric.Type() {
	case telegraf.Counter:
		metricType = "counter"
	case telegraf.Gauge:
		metricType = "gauge"
	case telegraf.Untyped:
		// For untyped metrics, default to Gauge
		metricType = "untyped"
	default:
		err = fmt.Errorf("unable to determine metric type")
	}
	return metricType, err
}

/*Determine and assign the datapoint value based on the telegraf value type*/
func getMetricValue(metric telegraf.Metric,
	field string) (datapoint.Value, error) {
	var err error
	var metricValue datapoint.Value
	var value = metric.Fields()[field]
	switch value.(type) {
	case int64:
		metricValue = datapoint.NewIntValue(value.(int64))
	case int32:
		metricValue = datapoint.NewIntValue(int64(value.(int32)))
	case int16:
		metricValue = datapoint.NewIntValue(int64(value.(int16)))
	case int8:
		metricValue = datapoint.NewIntValue(int64(value.(int8)))
	case int:
		metricValue = datapoint.NewIntValue(int64(value.(int)))
	case uint64:
		metricValue = datapoint.NewIntValue(int64(value.(uint64)))
	case uint32:
		metricValue = datapoint.NewIntValue(int64(value.(uint32)))
	case uint16:
		metricValue = datapoint.NewIntValue(int64(value.(uint16)))
	case uint8:
		metricValue = datapoint.NewIntValue(int64(value.(uint8)))
	case uint:
		metricValue = datapoint.NewIntValue(int64(value.(uint)))
	case float64:
		metricValue = datapoint.NewFloatValue(value.(float64))
	case float32:
		metricValue = datapoint.NewFloatValue(float64(value.(float32)))
	default:
		err = fmt.Errorf("unknown metric value type %s", reflect.TypeOf(value))
	}
	return metricValue, err
}

func parseMetricType(metric telegraf.Metric) (datapoint.MetricType, string, error) {
	var err error
	var metricType datapoint.MetricType
	var metricTypeString string
	// Parse the metric type
	if metricType, err = getMetricType(metric); err == nil {
		metricTypeString, err = getMetricTypeAsString(metric)
	}
	return metricType, metricTypeString, err
}

func getMetricName(metric telegraf.Metric, field string, dims map[string]string, props map[string]interface{}) string {
	var name string
	name = metric.Name()

	// If sf_prefix is provided
	if metric.HasTag("sf_prefix") {
		name = dims["sf_prefix"]
	}

	// Include field when it adds to the metric name
	if field != "value" {
		name = name + "." + field
	}

	// If sf_metric is provided
	if metric.HasTag("sf_metric") {
		// If sf_metric is provided
		name = dims["sf_metric"]
	}

	return name
}

// Modify the dimensions of the metric according to the following rules
func modifyDimensions(name string, metricType string, dims map[string]string, props map[string]interface{}) error {
	var err error
	// Add common dimensions
	dims["agent"] = "telegraf"
	dims["telegraf_type"] = metricType

	// If the plugin doesn't define a plugin name use metric.Name()
	if _, in := dims["plugin"]; !in {
		dims["plugin"] = name
	}

	// remove sf_prefix if it exists in the dimension map
	if _, in := dims["sf_prefix"]; in {
		delete(dims, "sf_prefix")
	}

	// if sfMetric exists
	if sfMetric, in := dims["sf_metric"]; in {
		// if the metric is a metadata object
		if sfMetric == "objects.host-meta-data" {
			// If property exists remap it
			if _, in := dims["property"]; in {
				props["property"] = dims["property"]
				delete(dims, "property")
			} else {
				// This is a malformed metadata event
				err = fmt.Errorf("E! Output [signalfx] objects.host-metadata object doesn't have a property")
			}
			// remove the sf_metric dimension
			delete(dims, "sf_metric")
		}
	}
	return err
}

/*Write call back for writing metrics*/
func (s *SignalFx) Write(metrics []telegraf.Metric) error {
	for _, metric := range metrics {
		var datapoints = []*datapoint.Datapoint{}
		var events = []*event.Event{}
		var timestamp = metric.Time()
		var err error
		var metricType datapoint.MetricType
		var metricTypeString string

		// Parse metric type information
		if metricType, metricTypeString, err = parseMetricType(metric); err != nil {
			log.Println("E! Output [signalfx] Unable to parse metric type for metric ", metric.String())
			continue
		}

		for field := range metric.Fields() {
			var metricValue datapoint.Value
			var metricName string
			var metricProps = make(map[string]interface{})
			var metricDims = metric.Tags()

			// Get metric name
			metricName = getMetricName(metric, field, metricDims, metricProps)

			// Check if the metric is explicitly excluded
			if excluded := s.isExcluded(metricName); excluded {
				log.Println("D! Outputs [signalfx] excluding the following metric: ", metricName)
				continue
			}

			// Modify the dimensions of the metric and skip the metric if the dimensions are malformed
			if err = modifyDimensions(metric.Name(), metricTypeString, metricDims, metricProps); err != nil {
				continue
			}

			// Get the metric value as a datapoint value
			if metricValue, err = getMetricValue(metric, field); err == nil {
				var datapoint = datapoint.New(metricName,
					metricDims,
					metricValue.(datapoint.Value),
					metricType,
					timestamp)

				// log metric
				log.Println("D! Output [signalfx] ", datapoint.String())

				// Add metric as a datapoint
				datapoints = append(datapoints, datapoint)
			} else {
				// We've already type checked field, so set property with value
				metricProps["message"] = metric.Fields()[field]
				var event = event.NewWithProperties(metricName,
					event.AGENT,
					metricDims,
					metricProps,
					timestamp)

				// log event
				log.Println("D! Output [signalfx] ", event.String())

				// Add event
				events = append(events, event)
			}
		}
		err = s.client.AddDatapoints(s.ctx, datapoints)
		if err != nil {
			log.Println("E! Output [signalfx] ", err)
		}
		err = s.client.AddEvents(s.ctx, events)
		if err != nil {
			log.Println("E! Output [signalfx] ", err)
		}
	}
	return nil
}

// isExcluded - checks whether a metric name was put on the exclude list
func (s *SignalFx) isExcluded(name string) bool {
	for _, exclude := range s.Exclude {
		if name == exclude {
			return true
		}
	}
	return false
}

/*init initializes the plugin context*/
func init() {
	outputs.Add("signalfx", func() telegraf.Output {
		return NewSignalFx()
	})
}
