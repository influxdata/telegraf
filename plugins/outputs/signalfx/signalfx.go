package signalfx

import (
	"context"
	"log"

	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/outputs/signalfx/parse"
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
	exclude            map[string]bool
	include            map[string]bool
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

// GetMetricType casts a telegraf ValueType to a signalfx metric type
func GetMetricType(mtype telegraf.ValueType) (metricType datapoint.MetricType, metricTypeString string) {
	switch mtype {
	case telegraf.Counter:
		metricTypeString = "counter"
		metricType = datapoint.Counter
	case telegraf.Gauge:
		metricTypeString = "gauge"
		metricType = datapoint.Gauge
	case telegraf.Summary:
		metricTypeString = "summary"
		metricType = datapoint.Gauge
		log.Println("D! Output [signalfx] GetMetricType() summary metrics will be sent as gauges")
	case telegraf.Histogram:
		metricTypeString = "histogram"
		metricType = datapoint.Gauge
		log.Println("D! Output [signalfx] GetMetricType() histogram metrics will be sent as gauges")
	case telegraf.Untyped:
		metricTypeString = "untyped"
		metricType = datapoint.Gauge
		log.Println("D! Output [signalfx] GetMetricType() untyped metrics will be sent as gauges")
	default:
		metricTypeString = "unrecognized"
		metricType = datapoint.Gauge
		log.Println("D! Output [signalfx] GetMetricType() unrecognized metric type defaulting to gauge")
	}
	return
}

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

		metricType, metricTypeString = GetMetricType(metric.Type())

		for field, val := range metric.Fields() {
			// Copy the metric tags because they are meant to be treated as
			// immutable
			var metricDims = metric.Tags()

			// Generate the metric name
			var metricName, isSFX = parse.GetMetricName(metric.Name(), field, metricDims)

			// Check if the metric is explicitly excluded
			if s.isExcluded(metricName) {
				log.Println("D! Outputs [signalfx] excluding the following metric: ", metricName)
				continue
			}

			// If eligible, move the dimension "property" to properties
			metricProps, propErr := parse.ExtractProperty(metricName, metricDims)
			if propErr != nil {
				log.Printf("E! Output [signalfx] %v", propErr)
				continue
			}

			// Add common dimensions
			metricDims["agent"] = "telegraf"
			metricDims["telegraf_type"] = metricTypeString
			parse.SetPluginDimension(metric.Name(), metricDims)
			parse.RemoveSFXDimensions(metricDims)

			// Get the metric value as a datapoint value
			if metricValue, err := datapoint.CastMetricValue(val); err == nil {
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
				// Skip if it's not an sfx event and it's not included
				if !isSFX && !s.isIncluded(metricName) {
					continue
				}

				// We've already type checked field, so set property with value
				metricProps["message"] = val
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
	if s.exclude == nil {
		s.exclude = make(map[string]bool, len(s.Exclude))
		for _, exclude := range s.Exclude {
			s.exclude[exclude] = true
		}
	}
	return s.exclude[name]
}

// isIncluded - checks whether a metric name was put on the include list
func (s *SignalFx) isIncluded(name string) bool {
	if s.include == nil {
		s.include = make(map[string]bool, len(s.Include))
		for _, include := range s.Include {
			s.include[include] = true
		}
	}
	return s.include[name]
}

/*init initializes the plugin context*/
func init() {
	outputs.Add("signalfx", func() telegraf.Output {
		return NewSignalFx()
	})
}
