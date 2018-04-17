package signalfx

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/signalfx/golib/datapoint"
	"github.com/signalfx/golib/event"
	"github.com/signalfx/golib/sfxclient"
)

/*SignalFx plugin context*/
type SignalFx struct {
	APIToken           string
	BatchSize          int
	ChannelSize        int
	DatapointIngestURL string
	EventIngestURL     string
	Exclude            []string
	Include            []string
	ctx                context.Context
	client             *sfxclient.HTTPSink
	dps                chan *datapoint.Datapoint
	evts               chan *event.Event
	done               chan struct{}
	wg                 sync.WaitGroup
}

var sampleConfig = `
    ## SignalFx API Token
    APIToken = "my-secret-key" # required.

    ## BatchSize
    BatchSize = 1000

    ## Ingest URL
    DatapointIngestURL = "https://ingest.signalfx.com/v2/datapoint"
    EventIngestURL = "https://ingest.signalfx.com/v2/event"

    ## Exclude metrics by metric name
    Exclude = ["plugin.metric_name", ""]

    ## Events or String typed metrics are omitted by default,
    ## with the exception of host property events which are emitted by 
    ## the SignalFx Metadata Plugin.  If you require a string typed metric
    ## you must specify the metric name in the following list
    Include = ["plugin.metric_name", ""]
`

// NewSignalFx - returns a new context for the SignalFx output plugin
func NewSignalFx() *SignalFx {
	return &SignalFx{
		APIToken:           "",
		BatchSize:          1000,
		ChannelSize:        100000,
		DatapointIngestURL: "https://ingest.signalfx.com/v2/datapoint",
		EventIngestURL:     "https://ingest.signalfx.com/v2/event",
		Exclude:            []string{""},
		Include:            []string{""},
		done:               make(chan struct{}),
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
	s.dps = make(chan *datapoint.Datapoint, s.ChannelSize)
	s.evts = make(chan *event.Event, s.ChannelSize)
	s.wg.Add(2)
	go func() {
		s.emitDatapoints()
		s.wg.Done()
	}()
	go func() {
		s.emitEvents()
		s.wg.Done()
	}()
	log.Printf("I! Output [signalfx] batch size is %d\n", s.BatchSize)
	return nil
}

/*Close closes the connection to SignalFx*/
func (s *SignalFx) Close() error {
	close(s.done)  // drain the input channels
	s.wg.Wait()    // wait for the input channels to be drained
	s.client = nil // destroy the client
	return nil
}

/*Determine and assign a datapoint metric type based on telegraf metric type*/
func getMetricType(metric telegraf.Metric) (metricType datapoint.MetricType, err error) {
	switch metric.Type() {
	case telegraf.Counter:
		metricType = datapoint.Counter
		if metric.Name() == "mem" {
			metricType = datapoint.Gauge
		}
	case telegraf.Gauge:
		metricType = datapoint.Gauge
	case telegraf.Summary, telegraf.Histogram, telegraf.Untyped:
		metricType = datapoint.Gauge
		err = fmt.Errorf("histogram, summary, and untyped metrics will be sent as gauges")
	default:
		metricType = datapoint.Gauge
		err = fmt.Errorf("unrecognized metric type defaulting to gauge")
	}
	return metricType, err
}

/*Determine and assign a datapoint metric type based on telegraf metric type*/
func getMetricTypeAsString(metric telegraf.Metric) (metricType string, err error) {
	switch metric.Type() {
	case telegraf.Counter:
		metricType = "counter"
	case telegraf.Gauge:
		metricType = "gauge"
	case telegraf.Summary:
		metricType = "summary"
		err = fmt.Errorf("summary metrics will be sent as gauges")
	case telegraf.Histogram:
		metricType = "histogram"
		err = fmt.Errorf("histogram metrics will be sent as gauges")
	case telegraf.Untyped:
		metricType = "untyped"
		err = fmt.Errorf("untyped metrics will be sent as gauges")
	default:
		metricType = "unrecognized"
		err = fmt.Errorf("unrecognized metric type defaulting to gauge")
	}
	return metricType, err
}

func getIntegerValue(value interface{}) datapoint.Value {
	var metricValue datapoint.Value
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
	}
	return metricValue
}

func getUnsignedIntegerValue(value interface{}) datapoint.Value {
	var metricValue datapoint.Value
	switch value.(type) {
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
	}
	return metricValue
}

func getFloatingValue(value interface{}) datapoint.Value {
	var metricValue datapoint.Value
	switch value.(type) {
	case float64:
		metricValue = datapoint.NewFloatValue(value.(float64))
	case float32:
		metricValue = datapoint.NewFloatValue(float64(value.(float32)))
	}
	return metricValue
}

/*Determine and assign the datapoint value based on the telegraf value type*/
func getMetricValue(metric telegraf.Metric,
	field string) (datapoint.Value, error) {
	var err error
	var metricValue datapoint.Value
	var value = metric.Fields()[field]
	switch value.(type) {
	case int64, int32, int16, int8, int:
		metricValue = getIntegerValue(value)
	case uint64, uint32, uint16, uint8, uint:
		metricValue = getUnsignedIntegerValue(value)
	case float64, float32:
		metricValue = getFloatingValue(value)
	default:
		err = fmt.Errorf("unknown metric value type %s", reflect.TypeOf(value))
	}
	return metricValue, err
}

func parseMetricType(metric telegraf.Metric) (metricType datapoint.MetricType, metricTypeString string) {
	var err error
	// Parse the metric type
	metricType, err = getMetricType(metric)
	if err != nil {
		log.Printf("D! Outputs [signalfx] getMetricType() %s {%s}\n", err, metric)
	}
	metricTypeString, err = getMetricTypeAsString(metric)
	if err != nil {
		log.Printf("D! Outputs [signalfx] getMetricTypeAsString()  %s {%s}\n", err, metric)
	}
	return metricType, metricTypeString
}

func getMetricName(metric telegraf.Metric, field string, dims map[string]string, props map[string]interface{}) string {
	var name = metric.Name()

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
func modifyDimensions(name string, metricTypeString string, dims map[string]string, props map[string]interface{}) error {
	var err error
	// Add common dimensions
	dims["agent"] = "telegraf"
	dims["telegraf_type"] = metricTypeString

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

func (s *SignalFx) shouldSkipMetric(metricName string, metricTypeString string, metricDims map[string]string, metricProps map[string]interface{}) bool {
	// Check if the metric is explicitly excluded
	if excluded := s.isExcluded(metricName); excluded {
		log.Println("D! Outputs [signalfx] excluding the following metric: ", metricName)
		return true
	}

	// Modify the dimensions of the metric and skip the metric if the dimensions are malformed
	if err := modifyDimensions(metricName, metricTypeString, metricDims, metricProps); err != nil {
		return true
	}

	return false
}

func (s *SignalFx) emitDatapoints() {
	var buf []*datapoint.Datapoint
	for {
		select {
		case <-s.done:
			return
		case dp := <-s.dps:
			buf = append(buf, dp)
			s.fillAndSendDatapoints(buf)
			buf = buf[:0]
		}
	}
}

func (s *SignalFx) fillAndSendDatapoints(buf []*datapoint.Datapoint) {
outer:
	for {
		select {
		case dp := <-s.dps:
			buf = append(buf, dp)
			if len(buf) >= s.BatchSize {
				if err := s.client.AddDatapoints(s.ctx, buf); err != nil {
					log.Println("E! Output [signalfx] ", err)
				}
				buf = buf[:0]
			}
		default:
			break outer
		}
	}
	if len(buf) > 0 {
		if err := s.client.AddDatapoints(s.ctx, buf); err != nil {
			log.Println("E! Output [signalfx] ", err)
		}
	}
}

func (s *SignalFx) emitEvents() {
	var buf []*event.Event
	for {
		select {
		case <-s.done:
			return
		case e := <-s.evts:
			buf = append(buf, e)
			s.fillAndSendEvents(buf)
			buf = buf[:0]
		}
	}
}

func (s *SignalFx) fillAndSendEvents(buf []*event.Event) {
outer:
	for {
		select {
		case e := <-s.evts:
			buf = append(buf, e)
			if len(buf) >= s.BatchSize {
				if err := s.client.AddEvents(s.ctx, buf); err != nil {
					log.Println("E! Output [signalfx] ", err)
				}
				buf = buf[:0]
			}
		default:
			break outer
		}
	}
	if len(buf) > 0 {
		if err := s.client.AddEvents(s.ctx, buf); err != nil {
			log.Println("E! Output [signalfx] ", err)
		}
	}
}

// GetObjects - converts telegraf metrics to signalfx datapoints and events, and pushes them on to the supplied channels
func (s *SignalFx) GetObjects(metrics []telegraf.Metric, dps chan *datapoint.Datapoint, evts chan *event.Event) {
	for _, metric := range metrics {
		var timestamp = metric.Time()
		var metricType datapoint.MetricType
		var metricTypeString string

		metricType, metricTypeString = parseMetricType(metric)

		for field := range metric.Fields() {
			var metricName string
			var metricProps = make(map[string]interface{})
			var metricDims = metric.Tags()

			// Get metric name
			metricName = getMetricName(metric, field, metricDims, metricProps)

			if s.shouldSkipMetric(metric.Name(), metricTypeString, metricDims, metricProps) {
				continue
			}

			// Get the metric value as a datapoint value
			if metricValue, err := getMetricValue(metric, field); err == nil {
				var dp = datapoint.New(metricName,
					metricDims,
					metricValue.(datapoint.Value),
					metricType,
					timestamp)

				// log metric
				log.Println("D! Output [signalfx] ", dp.String())

				// Add metric as a datapoint
				dps <- dp
			} else {
				// Skip if it's not an sfx metric and it's not included
				if _, isSFX := metric.Tags()["sf_metric"]; !isSFX && !s.isIncluded(metricName) {
					continue
				}

				// We've already type checked field, so set property with value
				metricProps["message"] = metric.Fields()[field]
				var ev = event.NewWithProperties(metricName,
					event.AGENT,
					metricDims,
					metricProps,
					timestamp)

				// log event
				log.Println("D! Output [signalfx] ", ev.String())

				// Add event
				evts <- ev
			}
		}
	}
}

/*Write call back for writing metrics*/
func (s *SignalFx) Write(metrics []telegraf.Metric) error {
	s.GetObjects(metrics, s.dps, s.evts)
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

// isIncluded - checks whether a metric name was put on the include list
func (s *SignalFx) isIncluded(name string) bool {
	for _, include := range s.Include {
		if name == include {
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
