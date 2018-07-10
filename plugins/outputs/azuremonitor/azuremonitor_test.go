package azuremonitor

import (
	"net/http"
	"testing"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
)

func TestConnectionMSI(t *testing.T) {
	azm := AzureMonitor{}
}

// MockMetrics returns a mock []telegraf.Metric object for using in unit tests
// of telegraf output sinks.
func getMockMetrics() []telegraf.Metric {
	metrics := make([]telegraf.Metric, 0)
	// Create a new point batch
	metrics = append(metrics, getTestMetric(1.0))
	return metrics
}

// TestMetric Returns a simple test point:
//     measurement -> "test1" or name
//     tags -> "tag1":"value1"
//     value -> value
//     time -> time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
func getTestMetric(value interface{}, name ...string) telegraf.Metric {
	if value == nil {
		panic("Cannot use a nil value")
	}
	measurement := "test1"
	if len(name) > 0 {
		measurement = name[0]
	}
	tags := map[string]string{"tag1": "value1"}
	pt, _ := metric.New(
		measurement,
		tags,
		map[string]interface{}{"value": value},
		time.Now().UTC(),
	)
	return pt
}

func TestAzureMonitor_Write(t *testing.T) {
	type fields struct {
		useMsi              bool
		ResourceID          string
		Region              string
		Timeout             internal.Duration
		AzureSubscriptionID string
		AzureTenantID       string
		AzureClientID       string
		AzureClientSecret   string
		StringsAsDimensions bool
		url                 string
		auth                autorest.Authorizer
		client              *http.Client
		cache               map[time.Time]map[uint64]*aggregate
	}
	type args struct {
		metrics []telegraf.Metric
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AzureMonitor{
				useMsi:              tt.fields.useMsi,
				ResourceID:          tt.fields.ResourceID,
				Region:              tt.fields.Region,
				Timeout:             tt.fields.Timeout,
				AzureSubscriptionID: tt.fields.AzureSubscriptionID,
				AzureTenantID:       tt.fields.AzureTenantID,
				AzureClientID:       tt.fields.AzureClientID,
				AzureClientSecret:   tt.fields.AzureClientSecret,
				StringsAsDimensions: tt.fields.StringsAsDimensions,
				url:                 tt.fields.url,
				auth:                tt.fields.auth,
				client:              tt.fields.client,
				cache:               tt.fields.cache,
			}
			if err := a.Write(tt.args.metrics); (err != nil) != tt.wantErr {
				t.Errorf("AzureMonitor.Write() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
