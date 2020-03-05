package newrelic

import (
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/newrelic/newrelic-telemetry-sdk-go/telemetry"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	nr := &NewRelic{
		MetricPrefix: "Test",
		InsightsKey:  "12345",
	}

	err := nr.Connect()
	require.NoError(t, err)

	err = nr.Write(testutil.MockMetrics())
	require.NoError(t, err)
}

func TestNewRelic_Write(t *testing.T) {
	type fields struct {
		harvestor    *telemetry.Harvester
		InsightsKey  string
		MetricPrefix string
	}
	type args struct {
		metrics []telegraf.Metric
	}
	tests := []struct {
		name    string
		fields  fields
		metrics []telegraf.Metric
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Test: Basic mock metric write",
			fields: fields{
				InsightsKey:  "insightskey",
				MetricPrefix: "test1",
			},
			metrics: testutil.MockMetrics(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nr := &NewRelic{
				harvestor:    tt.fields.harvestor,
				InsightsKey:  tt.fields.InsightsKey,
				MetricPrefix: tt.fields.MetricPrefix,
			}
			if err := nr.Write(tt.metrics); (err != nil) != tt.wantErr {
				t.Errorf("NewRelic.Write() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewRelic_Connect(t *testing.T) {

	tests := []struct {
		name     string
		newrelic *NewRelic
		wantErr  bool
	}{
		{
			name: "Test: No Insights key",
			newrelic: &NewRelic{
				MetricPrefix: "prefix",
			},
			wantErr: true,
		},
		{
			name: "Test: Insights key",
			newrelic: &NewRelic{
				InsightsKey:  "12312133",
				MetricPrefix: "prefix",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nr := &NewRelic{
				harvestor:    tt.newrelic.harvestor,
				InsightsKey:  tt.newrelic.InsightsKey,
				MetricPrefix: tt.newrelic.MetricPrefix,
			}
			if err := nr.Connect(); (err != nil) != tt.wantErr {
				t.Errorf("NewRelic.Connect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
