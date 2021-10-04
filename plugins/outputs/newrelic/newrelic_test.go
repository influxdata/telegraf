package newrelic

import (
	"math"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"github.com/newrelic/newrelic-telemetry-sdk-go/telemetry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	nr := &NewRelic{
		MetricPrefix: "Test",
		InsightsKey:  "12345",
		Timeout:      config.Duration(time.Second * 5),
	}
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	err := nr.Connect()
	require.NoError(t, err)

	err = nr.Write(testutil.MockMetrics())
	assert.Contains(t, err.Error(), "unable to harvest metrics")
}

func TestNewRelic_Write(t *testing.T) {
	tests := []struct {
		name         string
		metrics      []telegraf.Metric
		auditMessage string
		wantErr      bool
	}{
		{
			name:         "Test: Basic mock metric write",
			metrics:      testutil.MockMetrics(),
			wantErr:      false,
			auditMessage: `"metrics":[{"name":"test1.value","type":"gauge","value":1,"timestamp":1257894000000,"attributes":{"tag1":"value1"}}]`,
		},
		{
			name: "Test: Test string ",
			metrics: []telegraf.Metric{
				testutil.TestMetric("value1", "test_String"),
			},
			wantErr:      false,
			auditMessage: "",
		},
		{
			name: "Test: Test int64 ",
			metrics: []telegraf.Metric{
				testutil.TestMetric(int64(15), "test_int64"),
			},
			wantErr:      false,
			auditMessage: `"metrics":[{"name":"test_int64.value","type":"gauge","value":15,"timestamp":1257894000000,"attributes":{"tag1":"value1"}}]`,
		},
		{
			name: "Test: Test  uint64 ",
			metrics: []telegraf.Metric{
				testutil.TestMetric(uint64(20), "test_uint64"),
			},
			wantErr:      false,
			auditMessage: `"metrics":[{"name":"test_uint64.value","type":"gauge","value":20,"timestamp":1257894000000,"attributes":{"tag1":"value1"}}]`,
		},
		{
			name: "Test: Test bool true ",
			metrics: []telegraf.Metric{
				testutil.TestMetric(bool(true), "test_bool_true"),
			},
			wantErr:      false,
			auditMessage: `"metrics":[{"name":"test_bool_true.value","type":"gauge","value":1,"timestamp":1257894000000,"attributes":{"tag1":"value1"}}]`,
		},
		{
			name: "Test: Test bool false ",
			metrics: []telegraf.Metric{
				testutil.TestMetric(bool(false), "test_bool_false"),
			},
			wantErr:      false,
			auditMessage: `"metrics":[{"name":"test_bool_false.value","type":"gauge","value":0,"timestamp":1257894000000,"attributes":{"tag1":"value1"}}]`,
		},
		{
			name: "Test: Test max float64 ",
			metrics: []telegraf.Metric{
				testutil.TestMetric(math.MaxFloat64, "test_maxfloat64"),
			},
			wantErr:      false,
			auditMessage: `"metrics":[{"name":"test_maxfloat64.value","type":"gauge","value":1.7976931348623157e+308,"timestamp":1257894000000,"attributes":{"tag1":"value1"}}]`,
		},
		{
			name: "Test: Test NAN ",
			metrics: []telegraf.Metric{
				testutil.TestMetric(math.NaN, "test_NaN"),
			},
			wantErr:      false,
			auditMessage: ``,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var auditLog map[string]interface{}
			nr := &NewRelic{}
			nr.harvestor, _ = telemetry.NewHarvester(
				telemetry.ConfigHarvestPeriod(0),
				func(cfg *telemetry.Config) {
					cfg.APIKey = "dummyTestKey"
					cfg.HarvestPeriod = 0
					cfg.HarvestTimeout = 0
					cfg.AuditLogger = func(e map[string]interface{}) {
						auditLog = e
					}
				})
			err := nr.Write(tt.metrics)
			assert.NoError(t, err)
			if auditLog["data"] != nil {
				assert.Contains(t, auditLog["data"], tt.auditMessage)
			} else {
				assert.Contains(t, "", tt.auditMessage)
			}

			if (err != nil) != tt.wantErr {
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
		{
			name: "Test: Only Insights key",
			newrelic: &NewRelic{
				InsightsKey: "12312133",
			},
			wantErr: false,
		},
		{
			name: "Test: Insights key and Timeout",
			newrelic: &NewRelic{
				InsightsKey: "12312133",
				Timeout:     config.Duration(time.Second * 5),
			},
			wantErr: false,
		},
		{
			name: "Test: HTTP Proxy",
			newrelic: &NewRelic{
				InsightsKey: "12121212",
				HTTPProxy:   "https://my.proxy",
			},
			wantErr: false,
		},
		{
			name: "Test: Metric URL ",
			newrelic: &NewRelic{
				InsightsKey: "12121212",
				MetricURL:   "https://test.nr.com",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nr := tt.newrelic
			if err := nr.Connect(); (err != nil) != tt.wantErr {
				t.Errorf("NewRelic.Connect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
