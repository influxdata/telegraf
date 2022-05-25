package signalfx

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/datapoint/dpsink"
	"github.com/signalfx/golib/v3/event"
	"github.com/signalfx/golib/v3/sfxclient"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//init initializes the plugin context
func init() {
	outputs.Add("signalfx", func() telegraf.Output {
		return NewSignalFx()
	})
}

// SignalFx plugin context
type SignalFx struct {
	AccessToken        string   `toml:"access_token"`
	SignalFxRealm      string   `toml:"signalfx_realm"`
	IngestURL          string   `toml:"ingest_url"`
	IncludedEventNames []string `toml:"included_event_names"`

	Log telegraf.Logger `toml:"-"`

	includedEventSet map[string]bool
	client           dpsink.Sink

	ctx    context.Context
	cancel context.CancelFunc
}

// GetMetricType returns the equivalent telegraf ValueType for a signalfx metric type
func GetMetricType(mtype telegraf.ValueType) (metricType datapoint.MetricType) {
	switch mtype {
	case telegraf.Counter:
		metricType = datapoint.Counter
	case telegraf.Gauge:
		metricType = datapoint.Gauge
	case telegraf.Summary:
		metricType = datapoint.Gauge
	case telegraf.Histogram:
		metricType = datapoint.Gauge
	case telegraf.Untyped:
		metricType = datapoint.Gauge
	default:
		metricType = datapoint.Gauge
	}
	return metricType
}

// NewSignalFx - returns a new context for the SignalFx output plugin
func NewSignalFx() *SignalFx {
	ctx, cancel := context.WithCancel(context.Background())
	return &SignalFx{
		AccessToken:        "",
		SignalFxRealm:      "",
		IngestURL:          "",
		IncludedEventNames: []string{""},
		ctx:                ctx,
		cancel:             cancel,
		client:             sfxclient.NewHTTPSink(),
	}
}

// Connect establishes a connection to SignalFx
func (s *SignalFx) Connect() error {
	client := s.client.(*sfxclient.HTTPSink)
	client.AuthToken = s.AccessToken

	if s.IngestURL != "" {
		client.DatapointEndpoint = datapointEndpointForIngestURL(s.IngestURL)
		client.EventEndpoint = eventEndpointForIngestURL(s.IngestURL)
	} else if s.SignalFxRealm != "" { //nolint: revive // "Simplifying" if c {...} else {... return } would not simplify anything at all in this case
		client.DatapointEndpoint = datapointEndpointForRealm(s.SignalFxRealm)
		client.EventEndpoint = eventEndpointForRealm(s.SignalFxRealm)
	} else {
		return errors.New("signalfx_realm or ingest_url must be configured")
	}

	return nil
}

// Close closes any connections to SignalFx
func (s *SignalFx) Close() error {
	s.cancel()
	s.client.(*sfxclient.HTTPSink).Client.CloseIdleConnections()
	return nil
}

func (s *SignalFx) ConvertToSignalFx(metrics []telegraf.Metric) ([]*datapoint.Datapoint, []*event.Event) {
	var dps []*datapoint.Datapoint
	var events []*event.Event

	for _, metric := range metrics {
		s.Log.Debugf("Processing the following measurement: %v", metric)
		var timestamp = metric.Time()

		metricType := GetMetricType(metric.Type())
		for field, val := range metric.Fields() {
			// Copy the metric tags because they are meant to be treated as
			// immutable
			var metricDims = metric.Tags()

			// Generate the metric name
			metricName := getMetricName(metric.Name(), field)

			// Get the metric value as a datapoint value
			if metricValue, err := datapoint.CastMetricValueWithBool(val); err == nil {
				var dp = datapoint.New(metricName,
					metricDims,
					metricValue,
					metricType,
					timestamp)

				s.Log.Debugf("Datapoint: %v", dp.String())

				dps = append(dps, dp)
			} else {
				// Skip if it's not an explicitly included event
				if !s.isEventIncluded(metricName) {
					continue
				}

				// We've already type checked field, so set property with value
				metricProps := map[string]interface{}{"message": val}
				var ev = event.NewWithProperties(metricName,
					event.AGENT,
					metricDims,
					metricProps,
					timestamp)

				s.Log.Debugf("Event: %v", ev.String())

				events = append(events, ev)
			}
		}
	}

	return dps, events
}

// Write call back for writing metrics
func (s *SignalFx) Write(metrics []telegraf.Metric) error {
	dps, events := s.ConvertToSignalFx(metrics)

	if len(dps) > 0 {
		err := s.client.AddDatapoints(s.ctx, dps)
		if err != nil {
			return err
		}
	}

	if len(events) > 0 {
		if err := s.client.AddEvents(s.ctx, events); err != nil {
			// If events error out but we successfully sent some datapoints,
			// don't return an error so that it won't ever retry -- that way we
			// don't send the same datapoints twice.
			if len(dps) == 0 {
				return err
			}
			s.Log.Errorf("Failed to send SignalFx event: %v", err)
		}
	}

	return nil
}

// isEventIncluded - checks whether a metric name for an event was put on the whitelist
func (s *SignalFx) isEventIncluded(name string) bool {
	if s.includedEventSet == nil {
		s.includedEventSet = make(map[string]bool, len(s.includedEventSet))
		for _, include := range s.IncludedEventNames {
			s.includedEventSet[include] = true
		}
	}
	return s.includedEventSet[name]
}

// getMetricName combines telegraf fields and tags into a full metric name
func getMetricName(metric string, field string) string {
	name := metric

	// Include field in metric name when it adds to the metric name
	if field != "value" {
		name = fmt.Sprintf("%s.%s", name, field)
	}

	return name
}

// ingestURLForRealm returns the base ingest URL for a particular SignalFx
// realm
func ingestURLForRealm(realm string) string {
	return fmt.Sprintf("https://ingest.%s.signalfx.com", realm)
}

// datapointEndpointForRealm returns the endpoint to which datapoints should be
// POSTed for a particular realm.
func datapointEndpointForRealm(realm string) string {
	return datapointEndpointForIngestURL(ingestURLForRealm(realm))
}

// datapointEndpointForRealm returns the endpoint to which datapoints should be
// POSTed for a particular ingest base URL.
func datapointEndpointForIngestURL(ingestURL string) string {
	return strings.TrimRight(ingestURL, "/") + "/v2/datapoint"
}

// eventEndpointForRealm returns the endpoint to which events should be
// POSTed for a particular realm.
func eventEndpointForRealm(realm string) string {
	return eventEndpointForIngestURL(ingestURLForRealm(realm))
}

// eventEndpointForRealm returns the endpoint to which events should be
// POSTed for a particular ingest base URL.
func eventEndpointForIngestURL(ingestURL string) string {
	return strings.TrimRight(ingestURL, "/") + "/v2/event"
}
