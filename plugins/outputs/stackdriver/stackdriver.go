package stackdriver

import (
	"context"
	"fmt"
	"path"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"

	// Imports the Stackdriver Monitoring client package.
	monitoring "cloud.google.com/go/monitoring/apiv3"
	googlepb "github.com/golang/protobuf/ptypes/timestamp"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

// GCPStackdriver is the Google Stackdriver config info.
type GCPStackdriver struct {
	Project   string
	Namespace string

	client *monitoring.MetricClient
}

var sampleConfig = `
  # GCP Project
  project = "erudite-bloom-151019"

  # The namespace for the metric descriptor
  namespace = "telegraf"
`

// Connect initiates the primary connection to the GCP project.
func (s *GCPStackdriver) Connect() error {
	if s.Project == "" {
		return fmt.Errorf("Project is a required field for stackdriver output")
	}

	if s.Namespace == "" {
		return fmt.Errorf("Namespace is a required field for stackdriver output")
	}

	if s.client == nil {
		ctx := context.Background()

		// Creates a client
		client, err := monitoring.NewMetricClient(ctx)
		if err != nil {
			return err
		}

		s.client = client
	}

	return nil
}

// Write writes the metrics to Google Cloud Stackdriver.
func (s *GCPStackdriver) Write(metrics []telegraf.Metric) error {
	ctx := context.Background()

	for _, m := range metrics {
		// Writes time series data
		for k, v := range m.Fields() {
			var value *monitoringpb.TypedValue

			switch vt := v.(type) {
			case float64:
				value = &monitoringpb.TypedValue{
					Value: &monitoringpb.TypedValue_DoubleValue{
						DoubleValue: v.(float64),
					},
				}
			case int64:
				value = &monitoringpb.TypedValue{
					Value: &monitoringpb.TypedValue_Int64Value{
						Int64Value: v.(int64),
					},
				}
			case bool:
				value = &monitoringpb.TypedValue{
					Value: &monitoringpb.TypedValue_BoolValue{
						BoolValue: v.(bool),
					},
				}
			case string:
				value = &monitoringpb.TypedValue{
					Value: &monitoringpb.TypedValue_StringValue{
						StringValue: v.(string),
					},
				}
			default:
				return fmt.Errorf("Unsupported type %T", vt)
			}

			// Prepares an individual data point
			dataPoint := &monitoringpb.Point{
				Interval: &monitoringpb.TimeInterval{
					EndTime: &googlepb.Timestamp{
						Seconds: m.Time().Unix(),
					},
				},
				Value: value,
			}

			if err := s.client.CreateTimeSeries(ctx, &monitoringpb.CreateTimeSeriesRequest{
				Name: monitoring.MetricProjectPath(s.Project),
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type:   path.Join("custom.googleapis.com", s.Namespace, m.Name(), k),
							Labels: m.Tags(),
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "global",
							Labels: map[string]string{
								"project_id": s.Project,
							},
						},
						Points: []*monitoringpb.Point{
							dataPoint,
						},
					},
				},
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

// Close will terminate the session to the backend, returning error if an issue arises.
func (s *GCPStackdriver) Close() error {
	return nil
}

// SampleConfig returns the formatted sample configuration for the plugin.
func (s *GCPStackdriver) SampleConfig() string {
	return sampleConfig
}

// Description returns the human-readable function definition of the plugin.
func (s *GCPStackdriver) Description() string {
	return "Configuration for Google Cloud Stackdriver to send metrics to"
}

func newGCPStackdriver() *GCPStackdriver {
	return &GCPStackdriver{}
}

func init() {
	outputs.Add("stackdriver", func() telegraf.Output {
		return newGCPStackdriver()
	})
}
