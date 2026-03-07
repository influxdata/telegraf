package aliyunrds

import (
	"errors"
	"testing"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	common_aliyun "github.com/influxdata/telegraf/plugins/common/aliyun"
	"github.com/influxdata/telegraf/testutil"
)

const inputTitle = "inputs.aliyunrds"

type mockRDSClient struct {
	returnError      bool
	emptyResponse    bool
	invalidDate      bool
	invalidValue     bool
	multiValueMetric bool
}

func (m *mockRDSClient) DescribeDBInstancePerformance(request *rds.DescribeDBInstancePerformanceRequest) (
	*rds.DescribeDBInstancePerformanceResponse, error) {
	if m.returnError {
		return nil, errors.New("mock error from RDS API")
	}

	resp := rds.CreateDescribeDBInstancePerformanceResponse()

	if m.emptyResponse {
		resp.PerformanceKeys.PerformanceKey = make([]rds.PerformanceKey, 0)
		return resp, nil
	}

	if m.invalidDate {
		resp.PerformanceKeys.PerformanceKey = []rds.PerformanceKey{
			{
				ValueFormat: "Sessions",
				Values: rds.ValuesInDescribeDBInstancePerformance{
					PerformanceValue: []rds.PerformanceValue{
						{
							Value: "10",
							Date:  "invalid-date-format",
						},
					},
				},
			},
		}
		return resp, nil
	}

	if m.invalidValue {
		resp.PerformanceKeys.PerformanceKey = []rds.PerformanceKey{
			{
				ValueFormat: "Sessions",
				Values: rds.ValuesInDescribeDBInstancePerformance{
					PerformanceValue: []rds.PerformanceValue{
						{
							Value: "not-a-number",
							Date:  "2024-01-01T00:00:00Z",
						},
					},
				},
			},
		}
		return resp, nil
	}

	if m.multiValueMetric {
		resp.PerformanceKeys.PerformanceKey = []rds.PerformanceKey{
			{
				ValueFormat: "read_iops&write_iops",
				Values: rds.ValuesInDescribeDBInstancePerformance{
					PerformanceValue: []rds.PerformanceValue{
						{
							Value: "100&200",
							Date:  "2024-01-01T00:00:00Z",
						},
					},
				},
			},
		}
		return resp, nil
	}

	switch request.Key {
	case "MySQL_Sessions":
		resp.PerformanceKeys.PerformanceKey = []rds.PerformanceKey{
			{
				ValueFormat: "Sessions",
				Values: rds.ValuesInDescribeDBInstancePerformance{
					PerformanceValue: []rds.PerformanceValue{
						{
							Value: "10",
							Date:  "2024-01-01T00:00:00Z",
						},
					},
				},
			},
		}
	case "MySQL_QPS":
		resp.PerformanceKeys.PerformanceKey = []rds.PerformanceKey{
			{
				ValueFormat: "QPS",
				Values: rds.ValuesInDescribeDBInstancePerformance{
					PerformanceValue: []rds.PerformanceValue{
						{
							Value: "100.5",
							Date:  "2024-01-01T00:00:00Z",
						},
					},
				},
			},
		}
	case "MySQL_IOPS":
		resp.PerformanceKeys.PerformanceKey = []rds.PerformanceKey{
			{
				ValueFormat: "read_iops&write_iops",
				Values: rds.ValuesInDescribeDBInstancePerformance{
					PerformanceValue: []rds.PerformanceValue{
						{
							Value: "100&200",
							Date:  "2024-01-01T00:00:00Z",
						},
					},
				},
			},
		}
	case "CpuUsage":
		resp.PerformanceKeys.PerformanceKey = []rds.PerformanceKey{
			{
				ValueFormat: "cpu_usage",
				Values: rds.ValuesInDescribeDBInstancePerformance{
					PerformanceValue: []rds.PerformanceValue{
						{
							Value: "45.2",
							Date:  "2024-01-01T00:00:00Z",
						},
					},
				},
			},
		}
	case "MySQL_NetworkInNew&MySQL_NetworkOutNew":
		resp.PerformanceKeys.PerformanceKey = []rds.PerformanceKey{
			{
				ValueFormat: "in&out",
				Values: rds.ValuesInDescribeDBInstancePerformance{
					PerformanceValue: []rds.PerformanceValue{
						{
							Value: "1024&2048",
							Date:  "2024-01-01T00:00:00Z",
						},
					},
				},
			},
		}
	default:
		resp.PerformanceKeys.PerformanceKey = make([]rds.PerformanceKey, 0)
	}

	return resp, nil
}

func TestAliyunRDSGather(t *testing.T) {
	plugin := &AliyunRDS{
		Regions:   []string{"cn-hangzhou"},
		Period:    config.Duration(5 * time.Minute),
		Delay:     config.Duration(1 * time.Minute),
		RateLimit: 200,
		Metrics: []*metric{
			{
				MetricNames:  []string{"MySQL_Sessions"},
				Instances:    []string{"rm-test123"},
				instanceList: []string{"rm-test123"},
			},
		},
		rdsClient: &mockRDSClient{},
		Log:       testutil.Logger{},
	}

	plugin.updateWindow(time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC))

	acc := &testutil.Accumulator{}
	err := plugin.Gather(acc)
	require.NoError(t, err)

	require.True(t, acc.HasMeasurement("aliyunrds"))
}

func TestAliyunRDSFetchPerformanceData(t *testing.T) {
	plugin := &AliyunRDS{
		rdsClient: &mockRDSClient{},
		Log:       testutil.Logger{},
	}

	plugin.updateWindow(time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC))

	tests := []struct {
		name           string
		metricName     string
		instanceID     string
		expectedFields int
	}{
		{
			name:           "single value metric",
			metricName:     "MySQL_Sessions",
			instanceID:     "rm-test123",
			expectedFields: 1,
		},
		{
			name:           "multiple value metric",
			metricName:     "MySQL_IOPS",
			instanceID:     "rm-test123",
			expectedFields: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataPoints, err := plugin.fetchRDSPerformanceDatapoints("cn-hangzhou", tt.metricName, tt.instanceID)
			require.NoError(t, err)
			require.NotEmpty(t, dataPoints)
			require.GreaterOrEqual(t, len(dataPoints), tt.expectedFields)
		})
	}
}

func TestUpdateWindow(t *testing.T) {
	plugin := &AliyunRDS{
		Period: config.Duration(5 * time.Minute),
		Delay:  config.Duration(1 * time.Minute),
	}

	now := time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC)

	plugin.updateWindow(now)
	require.False(t, plugin.windowStart.IsZero())
	require.False(t, plugin.windowEnd.IsZero())
	require.Equal(t, 5*time.Minute, plugin.windowEnd.Sub(plugin.windowStart))

	firstEnd := plugin.windowEnd
	plugin.updateWindow(now.Add(5 * time.Minute))
	require.Equal(t, firstEnd, plugin.windowStart)
	require.True(t, plugin.windowEnd.After(firstEnd))
}

func TestFormatField(t *testing.T) {
	tests := []struct {
		metric   string
		stat     string
		expected string
	}{
		{
			metric:   "MySQL_Sessions",
			stat:     "Sessions",
			expected: "my_sql_sessions_sessions",
		},
		{
			metric:   "MySQL_QPS",
			stat:     "MySQL_QPS",
			expected: "my_sql_qps_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.metric+"_"+tt.stat, func(t *testing.T) {
			result := formatField(tt.metric, tt.stat)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestPrepareInstances(t *testing.T) {
	plugin := &AliyunRDS{
		dt: nil,
	}

	m := &metric{
		Instances: []string{"rm-test1", "rm-test2"},
	}

	plugin.prepareInstances(m)

	require.Len(t, m.instanceList, 2)
	require.Contains(t, m.instanceList, "rm-test1")
	require.Contains(t, m.instanceList, "rm-test2")
}

func TestGetStringValue(t *testing.T) {
	testMap := map[string]interface{}{
		"string_key": "value",
		"int_key":    123,
		"nil_key":    nil,
	}

	require.Equal(t, "value", getStringValue(testMap, "string_key"))
	require.Empty(t, getStringValue(testMap, "int_key"))
	require.Empty(t, getStringValue(testMap, "nil_key"))
	require.Empty(t, getStringValue(testMap, "nonexistent_key"))
}

func TestAliyunRDSConfiguration(t *testing.T) {
	t.Run("region list defaults exist", func(t *testing.T) {
		require.NotEmpty(t, aliyunRegionList)
		require.Len(t, aliyunRegionList, 21)
		require.Contains(t, aliyunRegionList, "cn-hangzhou")
		require.Contains(t, aliyunRegionList, "cn-beijing")
	})

	t.Run("sample config exists", func(t *testing.T) {
		plugin := &AliyunRDS{}
		sampleConf := plugin.SampleConfig()
		require.NotEmpty(t, sampleConf)
		require.Contains(t, sampleConf, "aliyunrds")
		require.Contains(t, sampleConf, "regions")
		require.Contains(t, sampleConf, "period")
	})
}

func TestAliyunRDSGatherMultipleMetrics(t *testing.T) {
	plugin := &AliyunRDS{
		Regions:   []string{"cn-hangzhou"},
		Period:    config.Duration(5 * time.Minute),
		Delay:     config.Duration(1 * time.Minute),
		RateLimit: 200,
		Metrics: []*metric{
			{
				MetricNames:  []string{"MySQL_Sessions", "MySQL_QPS"},
				Instances:    []string{"rm-test123"},
				instanceList: []string{"rm-test123"},
			},
		},
		rdsClient: &mockRDSClient{},
		Log:       testutil.Logger{Name: inputTitle},
	}

	plugin.updateWindow(time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC))

	acc := &testutil.Accumulator{}
	err := plugin.Gather(acc)
	require.NoError(t, err)

	require.True(t, acc.HasMeasurement("aliyunrds"))

	metrics := acc.GetTelegrafMetrics()
	require.NotEmpty(t, metrics)
}

func TestAliyunRDSGatherMultipleInstances(t *testing.T) {
	plugin := &AliyunRDS{
		Regions:   []string{"cn-hangzhou"},
		Period:    config.Duration(5 * time.Minute),
		Delay:     config.Duration(1 * time.Minute),
		RateLimit: 200,
		Metrics: []*metric{
			{
				MetricNames:  []string{"MySQL_Sessions"},
				Instances:    []string{"rm-test1", "rm-test2"},
				instanceList: []string{"rm-test1", "rm-test2"},
			},
		},
		rdsClient: &mockRDSClient{},
		Log:       testutil.Logger{Name: inputTitle},
	}

	plugin.updateWindow(time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC))

	acc := &testutil.Accumulator{}
	err := plugin.Gather(acc)
	require.NoError(t, err)

	require.True(t, acc.HasMeasurement("aliyunrds"))
}

func TestAliyunRDSGatherWithErrors(t *testing.T) {
	tests := []struct {
		name        string
		metricNames []string
		expectError bool
		mockClient  *mockRDSClient
	}{
		{
			name:        "RDS API error",
			metricNames: []string{"MySQL_Sessions"},
			expectError: true,
			mockClient:  &mockRDSClient{returnError: true},
		},
		{
			name:        "Bad date format",
			metricNames: []string{"MySQL_Sessions"},
			expectError: true,
			mockClient:  &mockRDSClient{invalidDate: true},
		},
		{
			name:        "Bad value format",
			metricNames: []string{"MySQL_Sessions"},
			expectError: true,
			mockClient:  &mockRDSClient{invalidValue: true},
		},
		{
			name:        "Empty response",
			metricNames: []string{"MySQL_Sessions"},
			expectError: false,
			mockClient:  &mockRDSClient{emptyResponse: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &AliyunRDS{
				Regions:   []string{"cn-hangzhou"},
				Period:    config.Duration(5 * time.Minute),
				Delay:     config.Duration(1 * time.Minute),
				RateLimit: 200,
				Metrics: []*metric{
					{
						MetricNames:  tt.metricNames,
						Instances:    []string{"rm-test123"},
						instanceList: []string{"rm-test123"},
					},
				},
				rdsClient: tt.mockClient,
				Log:       testutil.Logger{Name: inputTitle},
			}

			plugin.updateWindow(time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC))

			acc := &testutil.Accumulator{}
			err := plugin.Gather(acc)

			if tt.expectError {
				hasError := err != nil || len(acc.Errors) > 0
				require.True(t, hasError, "Expected an error but got none")
			} else {
				require.NoError(t, err)
				require.Empty(t, acc.Errors)
			}
		})
	}
}

func TestFetchRDSPerformanceDatapointsEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		mockClient     *mockRDSClient
		metricName     string
		instanceID     string
		expectedError  bool
		expectedPoints int
	}{
		{
			name:           "single value metric",
			mockClient:     &mockRDSClient{},
			metricName:     "MySQL_Sessions",
			instanceID:     "rm-test123",
			expectedError:  false,
			expectedPoints: 1,
		},
		{
			name:           "multiple values metric",
			mockClient:     &mockRDSClient{multiValueMetric: true},
			metricName:     "MySQL_IOPS",
			instanceID:     "rm-test123",
			expectedError:  false,
			expectedPoints: 2,
		},
		{
			name:           "compound metric",
			mockClient:     &mockRDSClient{},
			metricName:     "MySQL_NetworkInNew&MySQL_NetworkOutNew",
			instanceID:     "rm-test123",
			expectedError:  false,
			expectedPoints: 2,
		},
		{
			name:           "bad date format",
			mockClient:     &mockRDSClient{invalidDate: true},
			metricName:     "MySQL_Sessions",
			instanceID:     "rm-test123",
			expectedError:  true,
			expectedPoints: 0,
		},
		{
			name:           "bad value format",
			mockClient:     &mockRDSClient{invalidValue: true},
			metricName:     "MySQL_Sessions",
			instanceID:     "rm-test123",
			expectedError:  true,
			expectedPoints: 0,
		},
		{
			name:           "empty response",
			mockClient:     &mockRDSClient{emptyResponse: true},
			metricName:     "MySQL_Sessions",
			instanceID:     "rm-test123",
			expectedError:  false,
			expectedPoints: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &AliyunRDS{
				rdsClient: tt.mockClient,
				Log:       testutil.Logger{Name: inputTitle},
			}

			plugin.updateWindow(time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC))

			dataPoints, err := plugin.fetchRDSPerformanceDatapoints("cn-hangzhou", tt.metricName, tt.instanceID)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, dataPoints, tt.expectedPoints)
			}
		})
	}
}

func TestUpdateWindowSequence(t *testing.T) {
	plugin := &AliyunRDS{
		Period: config.Duration(5 * time.Minute),
		Delay:  config.Duration(1 * time.Minute),
	}

	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	plugin.updateWindow(baseTime)
	firstStart := plugin.windowStart
	firstEnd := plugin.windowEnd

	require.Equal(t, 5*time.Minute, firstEnd.Sub(firstStart))
	require.Equal(t, baseTime.Add(-1*time.Minute), firstEnd)

	plugin.updateWindow(baseTime.Add(5 * time.Minute))
	require.Equal(t, firstEnd, plugin.windowStart)
	require.True(t, plugin.windowEnd.After(firstEnd))

	secondEnd := plugin.windowEnd
	plugin.updateWindow(baseTime.Add(10 * time.Minute))
	require.Equal(t, secondEnd, plugin.windowStart)
	require.True(t, plugin.windowEnd.After(secondEnd))
}

func TestSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "MySQL_QPS",
			expected: "my_sql_qps",
		},
		{
			input:    "CpuUsage",
			expected: "cpu_usage",
		},
		{
			input:    "NetworkInNew",
			expected: "network_in_new",
		},
		{
			input:    "simple",
			expected: "simple",
		},
		{
			input:    "ALLCAPS",
			expected: "allcaps",
		},
		{
			input:    "already_snake_case",
			expected: "already_snake_case",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := snakeCase(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatFieldVariations(t *testing.T) {
	tests := []struct {
		metric   string
		stat     string
		expected string
	}{
		{
			metric:   "MySQL_Sessions",
			stat:     "Sessions",
			expected: "my_sql_sessions_sessions",
		},
		{
			metric:   "MySQL_QPS",
			stat:     "MySQL_QPS",
			expected: "my_sql_qps_value",
		},
		{
			metric:   "CpuUsage",
			stat:     "cpu_usage",
			expected: "cpu_usage_cpu_usage",
		},
		{
			metric:   "IOPS",
			stat:     "read",
			expected: "iops_read",
		},
		{
			metric:   "Network",
			stat:     "in",
			expected: "network_in",
		},
	}

	for _, tt := range tests {
		t.Run(tt.metric+"_"+tt.stat, func(t *testing.T) {
			result := formatField(tt.metric, tt.stat)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestPrepareInstancesWithDiscovery(t *testing.T) {
	plugin := &AliyunRDS{
		dt: nil,
	}

	m := &metric{
		Instances: []string{"rm-configured1", "rm-configured2"},
	}

	plugin.prepareInstances(m)
	require.Len(t, m.instanceList, 2)
	require.Contains(t, m.instanceList, "rm-configured1")
	require.Contains(t, m.instanceList, "rm-configured2")

	m2 := &metric{
		Instances:     []string{"rm-configured"},
		discoveryTags: make(map[string]map[string]string),
	}
	m2.discoveryTags["rm-discovered"] = map[string]string{"RegionId": "cn-hangzhou"}

	plugin.prepareInstances(m2)
	require.Contains(t, m2.instanceList, "rm-configured")
}

func TestGatherMetricWithTags(t *testing.T) {
	plugin := &AliyunRDS{
		Regions:   []string{"cn-hangzhou"},
		Period:    config.Duration(5 * time.Minute),
		Delay:     config.Duration(1 * time.Minute),
		RateLimit: 200,
		Metrics: []*metric{
			{
				MetricNames:  []string{"MySQL_QPS"},
				Instances:    []string{"rm-test123"},
				instanceList: []string{"rm-test123"},
				discoveryTags: map[string]map[string]string{
					"rm-test123": {
						"RegionId":              "cn-hangzhou",
						"Engine":                "MySQL",
						"EngineVersion":         "8.0",
						"DBInstanceType":        "Primary",
						"DBInstanceDescription": "Test Instance",
					},
				},
			},
		},
		rdsClient: &mockRDSClient{},
		Log:       testutil.Logger{Name: inputTitle},
	}

	plugin.updateWindow(time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC))

	acc := &testutil.Accumulator{}
	err := plugin.Gather(acc)
	require.NoError(t, err)

	require.True(t, acc.HasMeasurement("aliyunrds"))

	metrics := acc.GetTelegrafMetrics()
	require.NotEmpty(t, metrics)

	for _, m := range metrics {
		tags := m.Tags()
		require.Contains(t, tags, "region")
		require.Contains(t, tags, "instanceId")
		require.Equal(t, "cn-hangzhou", tags["region"])
	}
}

func TestSampleConfig(t *testing.T) {
	plugin := &AliyunRDS{}
	sampleConf := plugin.SampleConfig()
	require.NotEmpty(t, sampleConf)
	require.Contains(t, sampleConf, "aliyunrds")
	require.Contains(t, sampleConf, "regions")
}

func TestStartStop(t *testing.T) {
	plugin := &AliyunRDS{
		dt:  nil,
		Log: testutil.Logger{Name: inputTitle},
	}

	acc := &testutil.Accumulator{}

	require.NoError(t, plugin.Start(acc))
	plugin.Stop()
}

func TestMultipleRegions(t *testing.T) {
	plugin := &AliyunRDS{
		Regions:   []string{"cn-hangzhou", "cn-beijing", "cn-shanghai"},
		Period:    config.Duration(5 * time.Minute),
		Delay:     config.Duration(1 * time.Minute),
		RateLimit: 200,
		Metrics: []*metric{
			{
				MetricNames:  []string{"MySQL_Sessions"},
				Instances:    []string{"rm-test123"},
				instanceList: []string{"rm-test123"},
			},
		},
		rdsClient: &mockRDSClient{},
		Log:       testutil.Logger{Name: inputTitle},
	}

	plugin.updateWindow(time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC))

	acc := &testutil.Accumulator{}
	err := plugin.Gather(acc)
	require.NoError(t, err)

	require.True(t, acc.HasMeasurement("aliyunrds"))
}

func TestRDSClientInterface(_ *testing.T) {
	var _ aliyunrdsClient = &mockRDSClient{}
}

func TestMetricDataPointParsing(t *testing.T) {
	plugin1 := &AliyunRDS{
		rdsClient: &mockRDSClient{},
		Log:       testutil.Logger{Name: inputTitle},
	}
	plugin1.updateWindow(time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC))

	dataPoints, err := plugin1.fetchRDSPerformanceDatapoints("cn-hangzhou", "MySQL_Sessions", "rm-test")
	require.NoError(t, err)
	require.Len(t, dataPoints, 1)
	require.Contains(t, dataPoints[0], "Sessions")
	require.Contains(t, dataPoints[0], "instanceId")
	require.Contains(t, dataPoints[0], "timestamp")

	plugin2 := &AliyunRDS{
		rdsClient: &mockRDSClient{multiValueMetric: true},
		Log:       testutil.Logger{Name: inputTitle},
	}
	plugin2.updateWindow(time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC))

	dataPoints, err = plugin2.fetchRDSPerformanceDatapoints("cn-hangzhou", "MySQL_IOPS", "rm-test")
	require.NoError(t, err)
	require.Len(t, dataPoints, 2)

	foundRead := false
	foundWrite := false
	for _, dp := range dataPoints {
		if _, ok := dp["read_iops"]; ok {
			foundRead = true
		}
		if _, ok := dp["write_iops"]; ok {
			foundWrite = true
		}
	}
	require.True(t, foundRead, "Should have read_iops datapoint")
	require.True(t, foundWrite, "Should have write_iops datapoint")
}

func TestEmptyMetricsConfiguration(t *testing.T) {
	plugin := &AliyunRDS{
		Regions:   []string{"cn-hangzhou"},
		Period:    config.Duration(5 * time.Minute),
		Delay:     config.Duration(1 * time.Minute),
		RateLimit: 200,
		Metrics: []*metric{
			{
				MetricNames:  make([]string, 0),
				Instances:    []string{"rm-test123"},
				instanceList: []string{"rm-test123"},
			},
		},
		rdsClient: &mockRDSClient{},
		Log:       testutil.Logger{Name: inputTitle},
	}

	plugin.updateWindow(time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC))

	acc := &testutil.Accumulator{}
	err := plugin.Gather(acc)
	require.NoError(t, err)

	require.False(t, acc.HasMeasurement("aliyunrds"))
}

func TestInitialDiscoveryDataIsUsedInFirstGather(t *testing.T) {
	// This test verifies that discovery data retrieved in Init() is properly used in the first Gather() call
	plugin := &AliyunRDS{
		AccessKeyID:     "dummy",
		AccessKeySecret: "dummy",
		Regions:         []string{"cn-shanghai"},
		Period:          config.Duration(time.Minute),
		Delay:           config.Duration(time.Minute),
		RateLimit:       200,
		Log:             testutil.Logger{Name: inputTitle},
		Metrics: []*metric{
			{
				MetricNames: []string{"MySQL_Sessions"},
			},
		},
	}

	// Mock RDS client
	plugin.rdsClient = &mockRDSClient{}

	// Mock discovery tool with initial data
	discoveryData := map[string]interface{}{
		"rm-test-123": map[string]interface{}{
			"DBInstanceId":          "rm-test-123",
			"RegionId":              "cn-shanghai",
			"DBInstanceType":        "Primary",
			"Engine":                "MySQL",
			"EngineVersion":         "8.0",
			"DBInstanceDescription": "Test Instance",
		},
	}

	// Create a mock discovery tool
	mockDT := &discoveryTool{
		DiscoveryTool: &common_aliyun.DiscoveryTool{
			DataChan: make(chan map[string]interface{}, 1),
		},
	}

	// Put initial discovery data into the channel so it's available during first prepareInstances
	mockDT.DataChan <- discoveryData

	plugin.dt = mockDT

	// Update window for gathering
	plugin.updateWindow(time.Now())

	// First gather - should use initial discovery data
	acc := &testutil.Accumulator{}
	err := plugin.Gather(acc)
	require.NoError(t, err)

	// Verify that the metric has instance list populated from discovery
	m := plugin.Metrics[0]
	require.NotNil(t, m.instanceList)
	require.Contains(t, m.instanceList, "rm-test-123")

	// Verify discovery tags were populated
	require.NotNil(t, m.discoveryTags)
	require.Contains(t, m.discoveryTags, "rm-test-123")
	require.Equal(t, "cn-shanghai", m.discoveryTags["rm-test-123"]["RegionId"])
	require.Equal(t, "MySQL", m.discoveryTags["rm-test-123"]["Engine"])
}
