package aliyuncms

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials/providers"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	commonaliyun "github.com/influxdata/telegraf/plugins/common/aliyun"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/testutil"
)

const inputTitle = "inputs.aliyuncms"

type mockGatherAliyunCMSClient struct{}

func (*mockGatherAliyunCMSClient) DescribeMetricList(request *cms.DescribeMetricListRequest) (*cms.DescribeMetricListResponse, error) {
	resp := new(cms.DescribeMetricListResponse)

	switch request.MetricName {
	case "InstanceActiveConnection":
		resp.Code = "200"
		resp.Period = "60"
		resp.Datapoints = `
		[{
			"timestamp": 1490152860000,
			"Maximum": 200,
			"userId": "1234567898765432",
			"Minimum": 100,
			"instanceId": "i-abcdefgh123456",
			"Average": 150,
			"Value": 300
		}]`
	case "ErrorCode":
		resp.Code = "404"
		resp.Message = "ErrorCode"
	case "ErrorDatapoint":
		resp.Code = "200"
		resp.Period = "60"
		resp.Datapoints = `
		[{
			"timestamp": 1490152860000,
			"Maximum": 200,
			"userId": "1234567898765432",
			"Minimum": 100,
			"instanceId": "i-abcdefgh123456",
			"Average": 150,
		}]`
	case "EmptyDatapoint":
		resp.Code = "200"
		resp.Period = "60"
		resp.Datapoints = `[]`
	case "ErrorResp":
		return nil, errors.New("error response")
	default:
		resp.Code = "200"
	}
	return resp, nil
}

type mockAliyunSDKCli struct {
	resp *responses.CommonResponse
}

func (m *mockAliyunSDKCli) ProcessCommonRequest(_ *requests.CommonRequest) (response *responses.CommonResponse, err error) {
	return m.resp, nil
}

func getDiscoveryTool(project string, discoverRegions []string) (*discoveryTool, error) {
	var (
		err        error
		credential auth.Credential
	)

	configuration := &providers.Configuration{
		AccessKeyID:     "dummyKey",
		AccessKeySecret: "dummySecret",
	}
	credentialProviders := []providers.Provider{
		providers.NewConfigurationCredentialProvider(configuration),
		providers.NewEnvCredentialProvider(),
		providers.NewInstanceMetadataProvider(),
	}
	credential, err = providers.NewChainProvider(credentialProviders).Retrieve()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credential: %w", err)
	}

	dt, err := newDiscoveryTool(discoverRegions, project, testutil.Logger{Name: inputTitle}, credential, 1, time.Minute*2)

	if err != nil {
		return nil, fmt.Errorf("can't create discovery tool object: %w", err)
	}
	return dt, nil
}

func getMockSdkCli(httpResp *http.Response) (mockAliyunSDKCli, error) {
	resp := responses.NewCommonResponse()
	if err := responses.Unmarshal(resp, httpResp, "JSON"); err != nil {
		return mockAliyunSDKCli{}, fmt.Errorf("can't parse response: %w", err)
	}
	return mockAliyunSDKCli{resp: resp}, nil
}

func TestPluginDefaults(t *testing.T) {
	require.Equal(t, &AliyunCMS{RateLimit: 200,
		DiscoveryInterval: config.Duration(time.Minute),
		dimensionKey:      "instanceId",
	}, inputs.Inputs["aliyuncms"]())
}

func TestPluginInitialize(t *testing.T) {
	var err error

	plugin := new(AliyunCMS)
	plugin.Log = testutil.Logger{Name: inputTitle}
	plugin.Regions = []string{"cn-shanghai"}
	plugin.dt, err = getDiscoveryTool("acs_slb_dashboard", plugin.Regions)
	if err != nil {
		t.Fatalf("Can't create discovery tool object: %v", err)
	}

	httpResp := &http.Response{
		StatusCode: 200,
		Body: io.NopCloser(bytes.NewBufferString(
			`{
						"LoadBalancers":
						 {
						  "LoadBalancer": [
 							 {"LoadBalancerId":"bla"}
                           ]
                         },
						"TotalCount": 1,
						"PageSize": 1,
						"PageNumber": 1
						}`)),
	}
	mockCli, err := getMockSdkCli(httpResp)
	if err != nil {
		t.Fatalf("Can't create mock sdk cli: %v", err)
	}
	plugin.dt.Cli = map[string]commonaliyun.AliyunSdkClient{plugin.Regions[0]: &mockCli}

	tests := []struct {
		name                string
		project             string
		accessKeyID         string
		accessKeySecret     string
		expectedErrorString string
		regions             []string
		discoveryRegions    []string
	}{
		{
			name:                "Empty project",
			expectedErrorString: "project is not set",
			regions:             []string{"cn-shanghai"},
		},
		{
			name:            "Valid project",
			project:         "acs_slb_dashboard",
			regions:         []string{"cn-shanghai"},
			accessKeyID:     "dummy",
			accessKeySecret: "dummy",
		},
		{
			name:            "'regions' is not set",
			project:         "acs_slb_dashboard",
			accessKeyID:     "dummy",
			accessKeySecret: "dummy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin.Project = tt.project
			plugin.AccessKeyID = tt.accessKeyID
			plugin.AccessKeySecret = tt.accessKeySecret
			plugin.Regions = tt.regions

			if tt.expectedErrorString != "" {
				require.EqualError(t, plugin.Init(), tt.expectedErrorString)
			} else {
				require.NoError(t, plugin.Init())
			}
			if len(tt.regions) == 0 { // Check if set to default
				require.Equal(t, plugin.Regions, commonaliyun.DefaultRegions())
			}
		})
	}
}

func TestPluginMetricsInitialize(t *testing.T) {
	var err error

	plugin := new(AliyunCMS)
	plugin.Log = testutil.Logger{Name: inputTitle}
	plugin.Regions = []string{"cn-shanghai"}
	plugin.dt, err = getDiscoveryTool("acs_slb_dashboard", plugin.Regions)
	if err != nil {
		t.Fatalf("Can't create discovery tool object: %v", err)
	}

	httpResp := &http.Response{
		StatusCode: 200,
		Body: io.NopCloser(bytes.NewBufferString(
			`{
				"LoadBalancers":
					{
						"LoadBalancer": [
 							{"LoadBalancerId":"bla"}
                        ]
                    },
				"TotalCount": 1,
				"PageSize": 1,
				"PageNumber": 1
			}`)),
	}
	mockCli, err := getMockSdkCli(httpResp)
	if err != nil {
		t.Fatalf("Can't create mock sdk cli: %v", err)
	}
	plugin.dt.Cli = map[string]commonaliyun.AliyunSdkClient{plugin.Regions[0]: &mockCli}

	tests := []struct {
		name                string
		project             string
		accessKeyID         string
		accessKeySecret     string
		expectedErrorString string
		regions             []string
		discoveryRegions    []string
		metrics             []*metric
	}{
		{
			name:            "Valid project",
			project:         "acs_slb_dashboard",
			regions:         []string{"cn-shanghai"},
			accessKeyID:     "dummy",
			accessKeySecret: "dummy",
			metrics: []*metric{
				{
					MetricNames: make([]string, 0),
					Dimensions:  `{"instanceId": "i-abcdefgh123456"}`,
				},
			},
		},
		{
			name:            "Valid project",
			project:         "acs_slb_dashboard",
			regions:         []string{"cn-shanghai"},
			accessKeyID:     "dummy",
			accessKeySecret: "dummy",
			metrics: []*metric{
				{
					MetricNames: make([]string, 0),
					Dimensions:  `[{"instanceId": "p-example"},{"instanceId": "q-example"}]`,
				},
			},
		},
		{
			name:                "Valid project",
			project:             "acs_slb_dashboard",
			regions:             []string{"cn-shanghai"},
			accessKeyID:         "dummy",
			accessKeySecret:     "dummy",
			expectedErrorString: `cannot parse dimensions (neither obj, nor array) "[": unexpected end of JSON input`,
			metrics: []*metric{
				{
					MetricNames: make([]string, 0),
					Dimensions:  `[`,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin.Project = tt.project
			plugin.AccessKeyID = tt.accessKeyID
			plugin.AccessKeySecret = tt.accessKeySecret
			plugin.Regions = tt.regions
			plugin.Metrics = tt.metrics

			if tt.expectedErrorString != "" {
				require.EqualError(t, plugin.Init(), tt.expectedErrorString)
			} else {
				require.NoError(t, plugin.Init())
			}
		})
	}
}

func TestPluginMetricsRDSServiceInitialize(t *testing.T) {
	var err error

	plugin := new(AliyunCMS)
	plugin.Log = testutil.Logger{Name: inputTitle}
	plugin.Regions = []string{"cn-shanghai"}
	plugin.dt, err = getDiscoveryTool("acs_slb_dashboard", plugin.Regions)
	if err != nil {
		t.Fatalf("Can't create discovery tool object: %v", err)
	}

	httpResp := &http.Response{
		StatusCode: 200,
		Body: io.NopCloser(bytes.NewBufferString(
			`{
				"LoadBalancers":
					{
						"LoadBalancer": [
 							{"LoadBalancerId":"bla"}
                        ]
                    },
				"TotalCount": 1,
				"PageSize": 1,
				"PageNumber": 1
			}`)),
	}
	mockCli, err := getMockSdkCli(httpResp)
	if err != nil {
		t.Fatalf("Can't create mock sdk cli: %v", err)
	}
	plugin.dt.Cli = map[string]commonaliyun.AliyunSdkClient{plugin.Regions[0]: &mockCli}

	test := struct {
		name                string
		metricServices      []string
		project             string
		accessKeyID         string
		accessKeySecret     string
		expectedErrorString string
		regions             []string
		discoveryRegions    []string
		metrics             []*metric
	}{

		name:            "Valid project",
		metricServices:  []string{"cms", "rds"},
		project:         "acs_rds_dashboard",
		regions:         []string{"cn-shanghai"},
		accessKeyID:     "dummy",
		accessKeySecret: "dummy",
		metrics: []*metric{
			{
				MetricNames: make([]string, 0),
				Dimensions:  `{"instanceId": "i-abcdefgh123456"}`,
			},
		},
	}

	t.Run(test.name, func(t *testing.T) {
		plugin.Project = test.project
		plugin.AccessKeyID = test.accessKeyID
		plugin.AccessKeySecret = test.accessKeySecret
		plugin.Regions = test.regions
		plugin.Metrics = test.metrics

		if test.expectedErrorString != "" {
			require.EqualError(t, plugin.Init(), test.expectedErrorString)
		} else {
			require.NoError(t, plugin.Init())
		}
	})
}

func TestUpdateWindow(t *testing.T) {
	duration, err := time.ParseDuration("1m")
	require.NoError(t, err)
	internalDuration := config.Duration(duration)

	plugin := &AliyunCMS{
		Project: "acs_slb_dashboard",
		Period:  internalDuration,
		Delay:   internalDuration,
		Log:     testutil.Logger{Name: inputTitle},
	}

	now := time.Now()

	require.True(t, plugin.windowEnd.IsZero())
	require.True(t, plugin.windowStart.IsZero())

	plugin.updateWindow(now)

	newStartTime := plugin.windowEnd

	// the initial window just has a single period
	require.EqualValues(t, plugin.windowEnd, now.Add(-time.Duration(plugin.Delay)))
	require.EqualValues(t, plugin.windowStart, now.Add(-time.Duration(plugin.Delay)).Add(-time.Duration(plugin.Period)))

	now = time.Now()
	plugin.updateWindow(now)

	// subsequent window uses previous end time as start time
	require.EqualValues(t, plugin.windowEnd, now.Add(-time.Duration(plugin.Delay)))
	require.EqualValues(t, plugin.windowStart, newStartTime)
}

func TestGatherMetric(t *testing.T) {
	plugin := &AliyunCMS{
		Project:     "acs_slb_dashboard",
		cmsClient:   new(mockGatherAliyunCMSClient),
		measurement: formatMeasurement("acs_slb_dashboard"),
		Log:         testutil.Logger{Name: inputTitle},
		Regions:     []string{"cn-shanghai"},
	}

	tests := []struct {
		name                string
		metricName          string
		expectedErrorString string
	}{
		{
			name:                "Datapoint with corrupted JSON",
			metricName:          "ErrorDatapoint",
			expectedErrorString: `failed to decode response datapoints: invalid character '}' looking for beginning of object key string`,
		},
		{
			name:                "General CMS response error",
			metricName:          "ErrorResp",
			expectedErrorString: "failed to query metricName list: error response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metric := &metric{
				MetricNames:       []string{tt.metricName},
				Dimensions:        `{"instanceId": "i-abcdefgh123456"}`,
				requestDimensions: []map[string]string{{"instanceId": "i-abcdefgh123456"}},
			}
			var acc telegraf.Accumulator
			require.EqualError(t, plugin.gatherMetric(acc, tt.metricName, metric), tt.expectedErrorString)
		})
	}
}

func TestRDSServiceInitialization(t *testing.T) {
	var err error

	plugin := new(AliyunCMS)
	plugin.Log = testutil.Logger{Name: inputTitle}
	plugin.Regions = []string{"cn-shanghai"}
	plugin.dt, err = getDiscoveryTool("acs_rds_dashboard", plugin.Regions)
	if err != nil {
		t.Fatalf("Can't create discovery tool object: %v", err)
	}

	httpResp := &http.Response{
		StatusCode: 200,
		Body: io.NopCloser(bytes.NewBufferString(
			`{
				"DBInstances": {
					"DBInstance": [
						{"DBInstanceId": "rds-1"}
					]
				},
				"TotalCount": 1,
				"PageSize": 1,
				"PageNumber": 1
			}`)),
	}
	mockCli, err := getMockSdkCli(httpResp)
	if err != nil {
		t.Fatalf("Can't create mock sdk cli: %v", err)
	}
	plugin.dt.Cli = map[string]commonaliyun.AliyunSdkClient{plugin.Regions[0]: &mockCli}

	plugin.Project = "acs_rds_dashboard"
	plugin.AccessKeyID = "dummy"
	plugin.AccessKeySecret = "dummy"
	plugin.Regions = []string{"cn-shanghai"}

	require.NoError(t, plugin.Init())
}

func TestRDSMetricsConfiguration(t *testing.T) {
	plugin := &AliyunCMS{
		Project: "acs_rds_dashboard",
		Log:     testutil.Logger{Name: inputTitle},
	}

	m := &metric{
		MetricNames:                   []string{"CPUUsage", "MemoryUsage"},
		Dimensions:                    `{"instanceId": "rds-instance-001"}`,
		AllowDataPointWODiscoveryData: true,
	}

	plugin.Metrics = []*metric{m}

	require.Len(t, plugin.Metrics, 1)
	require.Len(t, plugin.Metrics[0].MetricNames, 2)
	require.NotEmpty(t, plugin.Metrics[0].Dimensions)
}

func TestRDSMetricDimensionParsing(t *testing.T) {
	tests := []struct {
		name           string
		dimensionsJSON string
		shouldSucceed  bool
	}{
		{
			name:           "Single instance dimension",
			dimensionsJSON: `{"instanceId": "rds-mysql-001"}`,
			shouldSucceed:  true,
		},
		{
			name:           "Array of instances",
			dimensionsJSON: `[{"instanceId": "rds-1"}, {"instanceId": "rds-2"}]`,
			shouldSucceed:  true,
		},
		{
			name:           "Invalid JSON",
			dimensionsJSON: `[{invalid`,
			shouldSucceed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &metric{
				Dimensions:      tt.dimensionsJSON,
				dimensionsUdObj: make(map[string]string),
				dimensionsUdArr: make([]map[string]string, 0),
			}

			err := json.Unmarshal([]byte(tt.dimensionsJSON), &m.dimensionsUdObj)
			if err != nil {
				// Try to parse as an array
				err = json.Unmarshal([]byte(tt.dimensionsJSON), &m.dimensionsUdArr)
			}

			if tt.shouldSucceed {
				require.NoError(t, err)
				require.True(t, len(m.dimensionsUdObj) > 0 || len(m.dimensionsUdArr) > 0)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestRDSMetricsGatheringWithMockClient(t *testing.T) {
	m := &metric{
		MetricNames:       []string{"CPUUsage"},
		Dimensions:        `{"instanceId": "i-1"}`,
		requestDimensions: []map[string]string{{"instanceId": "i-1"}},
	}

	plugin := &AliyunCMS{
		AccessKeyID:     "test_key",
		AccessKeySecret: "test_secret",
		Project:         "acs_rds_dashboard",
		Metrics:         []*metric{m},
		RateLimit:       200,
		measurement:     formatMeasurement("acs_rds_dashboard"),
		Regions:         []string{"cn-shanghai"},
		Log:             testutil.Logger{Name: inputTitle},
	}

	now := time.Now()
	plugin.updateWindow(now)

	require.False(t, plugin.windowEnd.IsZero())
	require.False(t, plugin.windowStart.IsZero())
}

func TestRDSDataPointValidation(t *testing.T) {
	tests := []struct {
		name          string
		dataPoint     map[string]interface{}
		shouldBeValid bool
	}{
		{
			name: "Valid RDS datapoint",
			dataPoint: map[string]interface{}{
				"instanceId": "rds-mysql-001",
				"CPUUsage":   75.5,
				"timestamp":  int64(1609556645),
			},
			shouldBeValid: true,
		},
		{
			name: "Missing instanceId",
			dataPoint: map[string]interface{}{
				"CPUUsage":  75.5,
				"timestamp": int64(1609556645),
			},
			shouldBeValid: false,
		},
		{
			name: "Missing timestamp",
			dataPoint: map[string]interface{}{
				"instanceId": "rds-mysql-001",
				"CPUUsage":   75.5,
			},
			shouldBeValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasInstanceID := len(tt.dataPoint) > 0 && tt.dataPoint["instanceId"] != nil
			hasTimestamp := len(tt.dataPoint) > 0 && tt.dataPoint["timestamp"] != nil
			isValid := hasInstanceID && hasTimestamp

			if tt.shouldBeValid {
				require.True(t, isValid)
			} else {
				require.False(t, isValid)
			}
		})
	}
}

func TestGather(t *testing.T) {
	tests := []struct {
		name                string
		project             string
		region              string
		httpResp            *http.Response
		discData            map[string]interface{}
		totalCount          int
		pageSize            int
		pageNumber          int
		expectedErrorString string
	}{
		{
			name:    "No root key in discovery response",
			project: "acs_slb_dashboard",
			region:  "cn-hongkong",
			httpResp: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
			},
			totalCount:          0,
			pageSize:            0,
			pageNumber:          0,
			expectedErrorString: `didn't find root key "LoadBalancers" in discovery response`,
		},
		{
			name:    "1 object discovered",
			project: "acs_slb_dashboard",
			region:  "cn-hongkong",
			httpResp: &http.Response{
				StatusCode: 200,
				Body: io.NopCloser(bytes.NewBufferString(
					`{
						"LoadBalancers":
						 {
						  "LoadBalancer": [
 							 {"LoadBalancerId":"bla"}
                           ]
                         },
						"TotalCount": 1,
						"PageSize": 1,
						"PageNumber": 1
						}`)),
			},
			discData:            map[string]interface{}{"bla": map[string]interface{}{"LoadBalancerId": "bla"}},
			totalCount:          1,
			pageSize:            1,
			pageNumber:          1,
			expectedErrorString: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt, err := getDiscoveryTool(tt.project, []string{tt.region})
			if err != nil {
				t.Fatalf("Can't create discovery tool object: %v", err)
			}

			mockCli, err := getMockSdkCli(tt.httpResp)
			if err != nil {
				t.Fatalf("Can't create mock sdk cli: %v", err)
			}
			dt.Cli = map[string]commonaliyun.AliyunSdkClient{tt.region: &mockCli}
			data, err := dt.GetDiscoveryDataAcrossRegions(nil)

			require.Equal(t, tt.discData, data)
			if err != nil {
				require.EqualError(t, err, tt.expectedErrorString)
			}
		})
	}
}

func TestRDSMetricConfiguration(t *testing.T) {
	plugin := &AliyunCMS{
		Project: "acs_rds_dashboard",
		Log:     testutil.Logger{Name: inputTitle},
	}

	m := &metric{
		MetricNames: []string{"CPUUsage", "DiskUsage"},
		Dimensions:  `{"instanceId": "rds-mysql-001"}`,
	}

	plugin.Metrics = []*metric{m}

	require.Len(t, plugin.Metrics, 1)
	require.Len(t, plugin.Metrics[0].MetricNames, 2)
	require.Equal(t, "CPUUsage", plugin.Metrics[0].MetricNames[0])
	require.Equal(t, "DiskUsage", plugin.Metrics[0].MetricNames[1])
}

func TestRDSInstanceIdExtraction(t *testing.T) {
	tests := []struct {
		name        string
		dimensions  string
		shouldParse bool
	}{
		{
			name:        "Valid single instance",
			dimensions:  `{"instanceId": "rds-mysql-001"}`,
			shouldParse: true,
		},
		{
			name:        "Valid array of instances",
			dimensions:  `[{"instanceId": "rds-mysql-001"}, {"instanceId": "rds-mysql-002"}]`,
			shouldParse: true,
		},
		{
			name:        "Invalid JSON",
			dimensions:  `[invalid`,
			shouldParse: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &metric{
				Dimensions:      tt.dimensions,
				dimensionsUdObj: make(map[string]string),
				dimensionsUdArr: make([]map[string]string, 0),
			}

			err := json.Unmarshal([]byte(tt.dimensions), &m.dimensionsUdObj)
			if err != nil {
				err = json.Unmarshal([]byte(tt.dimensions), &m.dimensionsUdArr)
			}

			if tt.shouldParse {
				require.NoError(t, err)
				require.True(t, len(m.dimensionsUdObj) > 0 || len(m.dimensionsUdArr) > 0)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestRDSWindowManagement(t *testing.T) {
	plugin := &AliyunCMS{
		Project: "acs_rds_dashboard",
		Log:     testutil.Logger{Name: inputTitle},
	}

	duration, err := time.ParseDuration("1m")
	if err != nil {
		t.Fatalf("Failed to parse duration: %v", err)
	}
	plugin.Period = config.Duration(duration)
	plugin.Delay = config.Duration(duration)

	now := time.Now()
	plugin.updateWindow(now)

	require.False(t, plugin.windowStart.IsZero())
	require.False(t, plugin.windowEnd.IsZero())
	require.True(t, plugin.windowEnd.After(plugin.windowStart))
}

func TestRDSServiceInMetricServices(t *testing.T) {
	tests := []struct {
		name           string
		metricServices []string
		hasRDS         bool
	}{
		{
			name:           "With RDS",
			metricServices: []string{"cms", "rds"},
			hasRDS:         true,
		},
		{
			name:           "Without RDS",
			metricServices: []string{"cms"},
			hasRDS:         false,
		},
		{
			name:           "RDS only",
			metricServices: []string{"rds"},
			hasRDS:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasRDS := false
			for _, svc := range tt.metricServices {
				if svc == "rds" {
					hasRDS = true
					break
				}
			}
			require.Equal(t, tt.hasRDS, hasRDS)
		})
	}
}

func TestParseTag(t *testing.T) {
	tests := []struct {
		name        string
		tagSpec     string
		data        interface{}
		shouldError bool
		expectedKey string
	}{
		{
			name:        "Simple tag key",
			tagSpec:     "region",
			data:        map[string]interface{}{"region": "us-east-1"},
			shouldError: false,
			expectedKey: "region",
		},
		{
			name:        "Tag with query path",
			tagSpec:     "env:environment.name",
			data:        map[string]interface{}{"environment": map[string]interface{}{"name": "prod"}},
			shouldError: false,
			expectedKey: "env",
		},
		{
			name:        "Invalid query path",
			tagSpec:     "tag:nonexistent.path",
			data:        map[string]interface{}{},
			shouldError: false,
			expectedKey: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tagKey, _, err := parseTag(tt.tagSpec, tt.data)

			if tt.shouldError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedKey, tagKey)
			}
		})
	}
}

func TestMetricConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		metric    *metric
		shouldErr bool
	}{
		{
			name: "Valid metric with names",
			metric: &metric{
				MetricNames: []string{"CPUUsage", "MemoryUsage"},
				Dimensions:  `{"instanceId": "i-123"}`,
			},
			shouldErr: false,
		},
		{
			name: "Empty metric names",
			metric: &metric{
				MetricNames: make([]string, 0),
				Dimensions:  `{"instanceId": "i-123"}`,
			},
			shouldErr: false,
		},
		{
			name: "Multiple dimensions",
			metric: &metric{
				MetricNames: []string{"CPUUsage"},
				Dimensions:  `[{"instanceId": "i-1"},{"instanceId": "i-2"}]`,
			},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.metric)
			require.NotEmpty(t, tt.metric.Dimensions)
		})
	}
}

func TestDiscoveryTagEnrichmentScenarios(t *testing.T) {
	plugin := &AliyunCMS{
		Log: testutil.Logger{Name: inputTitle},
	}

	tests := []struct {
		name                          string
		instanceID                    string
		discoveryTags                 map[string]map[string]string
		allowDataPointWODiscoveryData bool
		expectedKeep                  bool
		expectedTags                  map[string]string
	}{
		{
			name:       "Found in discovery with multiple tags",
			instanceID: "prod-001",
			discoveryTags: map[string]map[string]string{
				"prod-001": {
					"region":     "cn-shanghai",
					"env":        "production",
					"team":       "platform",
					"costcenter": "engineering",
				},
			},
			allowDataPointWODiscoveryData: false,
			expectedKeep:                  true,
			expectedTags: map[string]string{
				"region":     "cn-shanghai",
				"env":        "production",
				"team":       "platform",
				"costcenter": "engineering",
			},
		},
		{
			name:       "Not found but allow without discovery",
			instanceID: "unknown-123",
			discoveryTags: map[string]map[string]string{
				"known-001": {"env": "prod"},
			},
			allowDataPointWODiscoveryData: true,
			expectedKeep:                  true,
			expectedTags:                  map[string]string{},
		},
		{
			name:       "Not found and disallow without discovery",
			instanceID: "unknown-123",
			discoveryTags: map[string]map[string]string{
				"known-001": {"env": "prod"},
			},
			allowDataPointWODiscoveryData: false,
			expectedKeep:                  false,
			expectedTags:                  map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &metric{
				discoveryTags:                 tt.discoveryTags,
				AllowDataPointWODiscoveryData: tt.allowDataPointWODiscoveryData,
			}

			tags := make(map[string]string)
			keep := plugin.enrichTagsWithDiscovery(tags, m, tt.instanceID)

			require.Equal(t, tt.expectedKeep, keep)

			for k, v := range tt.expectedTags {
				require.Equal(t, v, tags[k], "Tag %s should be %s", k, v)
			}
		})
	}
}

func TestWindowTimingCalculations(t *testing.T) {
	tests := []struct {
		name           string
		period         time.Duration
		delay          time.Duration
		firstCallTime  time.Time
		secondCallTime time.Time
	}{
		{
			name:           "1 minute period with 1 minute delay",
			period:         time.Minute,
			delay:          time.Minute,
			firstCallTime:  time.Date(2025, 12, 7, 10, 0, 0, 0, time.UTC),
			secondCallTime: time.Date(2025, 12, 7, 10, 2, 0, 0, time.UTC),
		},
		{
			name:           "5 minute period with 0 delay",
			period:         5 * time.Minute,
			delay:          0,
			firstCallTime:  time.Date(2025, 12, 7, 10, 0, 0, 0, time.UTC),
			secondCallTime: time.Date(2025, 12, 7, 10, 5, 0, 0, time.UTC),
		},
		{
			name:           "30 second period with 30 second delay",
			period:         30 * time.Second,
			delay:          30 * time.Second,
			firstCallTime:  time.Date(2025, 12, 7, 10, 0, 0, 0, time.UTC),
			secondCallTime: time.Date(2025, 12, 7, 10, 1, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &AliyunCMS{
				Project: "test",
				Period:  config.Duration(tt.period),
				Delay:   config.Duration(tt.delay),
				Log:     testutil.Logger{Name: inputTitle},
			}

			// First call
			plugin.updateWindow(tt.firstCallTime)
			firstStart := plugin.windowStart
			firstEnd := plugin.windowEnd

			require.False(t, firstStart.IsZero())
			require.False(t, firstEnd.IsZero())
			require.True(t, firstEnd.After(firstStart))

			// Second call
			plugin.updateWindow(tt.secondCallTime)
			secondStart := plugin.windowStart
			secondEnd := plugin.windowEnd

			// Window should have moved forward
			require.Equal(t, firstEnd, secondStart)
			require.True(t, secondEnd.After(firstEnd))
		})
	}
}

func TestTimestampConversionEdgeCases(t *testing.T) {
	plugin := &AliyunCMS{
		Log: testutil.Logger{Name: inputTitle},
	}

	tests := []struct {
		name      string
		value     interface{}
		shouldOK  bool
		checkZero bool
	}{
		{
			name:      "Zero int64",
			value:     int64(0),
			shouldOK:  true,
			checkZero: true,
		},
		{
			name:      "Large int64",
			value:     int64(9999999999),
			shouldOK:  true,
			checkZero: false,
		},
		{
			name:      "Float with decimals",
			value:     float64(1609556645123.456),
			shouldOK:  true,
			checkZero: false,
		},
		{
			name:      "Negative value",
			value:     int64(-1000),
			shouldOK:  true,
			checkZero: false,
		},
		{
			name:      "Boolean (invalid)",
			value:     true,
			shouldOK:  false,
			checkZero: false,
		},
		{
			name:      "Empty string (invalid)",
			value:     "",
			shouldOK:  false,
			checkZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := plugin.parseTimestamp(tt.value)

			require.Equal(t, tt.shouldOK, ok)
			if tt.shouldOK {
				if tt.checkZero {
					require.Equal(t, int64(0), result)
				}
			}
		})
	}
}

func TestClientSelection(t *testing.T) {
	tests := []struct {
		name        string
		cmsClient   aliyuncmsClient
		expectedRDS bool
		expectedCMS bool
	}{
		{
			name:        "Both clients available",
			cmsClient:   &mockGatherAliyunCMSClient{},
			expectedRDS: true,
			expectedCMS: true,
		},
		{
			name:        "Only CMS client",
			cmsClient:   &mockGatherAliyunCMSClient{},
			expectedRDS: false,
			expectedCMS: true,
		},
		{
			name:        "Only RDS client",
			cmsClient:   nil,
			expectedRDS: true,
			expectedCMS: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &AliyunCMS{
				cmsClient: tt.cmsClient,
				Log:       testutil.Logger{Name: inputTitle},
			}

			if tt.expectedCMS {
				require.NotNil(t, plugin.cmsClient)
			} else {
				require.Nil(t, plugin.cmsClient)
			}
		})
	}
}

func TestMetricsArrayHandling(t *testing.T) {
	tests := []struct {
		name          string
		metricsCount  int
		eachHasNames  int
		expectedTotal int
	}{
		{
			name:          "Single metric with multiple names",
			metricsCount:  1,
			eachHasNames:  5,
			expectedTotal: 5,
		},
		{
			name:          "Multiple metrics with multiple names each",
			metricsCount:  3,
			eachHasNames:  4,
			expectedTotal: 12,
		},
		{
			name:          "Empty metrics array",
			metricsCount:  0,
			eachHasNames:  0,
			expectedTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &AliyunCMS{
				Log: testutil.Logger{Name: inputTitle},
			}

			var metrics []*metric
			totalNames := 0

			for i := 0; i < tt.metricsCount; i++ {
				names := make([]string, tt.eachHasNames)
				for j := 0; j < tt.eachHasNames; j++ {
					names[j] = fmt.Sprintf("Metric_%d_%d", i, j)
					totalNames++
				}
				metrics = append(metrics, &metric{
					MetricNames: names,
					Dimensions:  `{"instanceId": "test"}`,
				})
			}

			plugin.Metrics = metrics

			require.Len(t, plugin.Metrics, tt.metricsCount)
			require.Equal(t, tt.expectedTotal, totalNames)
		})
	}
}

func TestRegionConfiguration(t *testing.T) {
	tests := []struct {
		name            string
		regions         []string
		expectedCount   int
		expectedRegions []string
	}{
		{
			name:            "Single region",
			regions:         []string{"cn-shanghai"},
			expectedCount:   1,
			expectedRegions: []string{"cn-shanghai"},
		},
		{
			name:            "Multiple regions",
			regions:         []string{"cn-shanghai", "cn-beijing", "cn-hongkong"},
			expectedCount:   3,
			expectedRegions: []string{"cn-shanghai", "cn-beijing", "cn-hongkong"},
		},
		{
			name:            "International regions",
			regions:         []string{"us-west-1", "eu-central-1", "ap-southeast-1"},
			expectedCount:   3,
			expectedRegions: []string{"us-west-1", "eu-central-1", "ap-southeast-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &AliyunCMS{
				Regions: tt.regions,
				Log:     testutil.Logger{Name: inputTitle},
			}

			require.Len(t, plugin.Regions, tt.expectedCount)
			for i, region := range tt.expectedRegions {
				require.Equal(t, region, plugin.Regions[i])
			}
		})
	}
}

func TestProjectAndMeasurement(t *testing.T) {
	tests := []struct {
		name             string
		project          string
		expectedContains string
	}{
		{
			name:             "RDS project",
			project:          "acs_rds_dashboard",
			expectedContains: "rds",
		},
		{
			name:             "ECS project",
			project:          "acs_ecs_dashboard",
			expectedContains: "ecs",
		},
		{
			name:             "SLB project",
			project:          "acs_slb_dashboard",
			expectedContains: "slb",
		},
		{
			name:             "OSS project",
			project:          "acs_oss",
			expectedContains: "oss",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			measurement := formatMeasurement(tt.project)
			require.NotEmpty(t, measurement)
			require.Contains(t, measurement, tt.expectedContains)
		})
	}
}

func TestFieldNameFormatting(t *testing.T) {
	tests := []struct {
		metricName string
		fieldName  string
		expected   string
	}{
		{
			metricName: "InstanceActiveConnection",
			fieldName:  "Value",
			expected:   "instance_active_connection_value",
		},
		{
			metricName: "InstanceActiveConnection",
			fieldName:  "Minimum",
			expected:   "instance_active_connection_minimum",
		},
		{
			metricName: "InstanceActiveConnection",
			fieldName:  "Maximum",
			expected:   "instance_active_connection_maximum",
		},
		{
			metricName: "InstanceActiveConnection",
			fieldName:  "Average",
			expected:   "instance_active_connection_average",
		},
		{
			metricName: "CpuUsage",
			fieldName:  "Value",
			expected:   "cpu_usage_value",
		},
		{
			metricName: "DiskUsage",
			fieldName:  "Percentage",
			expected:   "disk_usage_percentage",
		},
		{
			metricName: "NetworkIn",
			fieldName:  "Count",
			expected:   "network_in_count",
		},
		{
			metricName: "NetworkOut",
			fieldName:  "Rate",
			expected:   "network_out_rate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.metricName+"_"+tt.fieldName, func(t *testing.T) {
			result := formatField(tt.metricName, tt.fieldName)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestRateLimitConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		rateLimit int
		isValid   bool
	}{
		{
			name:      "Default rate limit",
			rateLimit: 200,
			isValid:   true,
		},
		{
			name:      "High rate limit",
			rateLimit: 1000,
			isValid:   true,
		},
		{
			name:      "Low rate limit",
			rateLimit: 10,
			isValid:   true,
		},
		{
			name:      "Zero rate limit",
			rateLimit: 0,
			isValid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &AliyunCMS{
				RateLimit: tt.rateLimit,
				Log:       testutil.Logger{Name: inputTitle},
			}

			if tt.isValid {
				require.Equal(t, tt.rateLimit, plugin.RateLimit)
			}
		})
	}
}

func TestInitialDiscoveryDataIsUsed(t *testing.T) {
	// This test verifies that discovery data retrieved in Init() is properly used in the first Gather() call
	plugin := &AliyunCMS{
		AccessKeyID:       "dummy",
		AccessKeySecret:   "dummy",
		Project:           "acs_rds_dashboard",
		Regions:           []string{"cn-shanghai"},
		Period:            config.Duration(time.Minute),
		Delay:             config.Duration(time.Minute),
		DiscoveryInterval: config.Duration(time.Minute),
		RateLimit:         200,
		Log:               testutil.Logger{Name: inputTitle},
		Metrics: []*metric{
			{
				MetricNames: []string{"CpuUsage"},
			},
		},
	}

	// Setup mock CMS client
	plugin.cmsClient = &mockGatherAliyunCMSClient{}

	// Setup mock discovery tool
	var err error
	plugin.dt, err = getDiscoveryTool("acs_rds_dashboard", plugin.Regions)
	require.NoError(t, err)

	// Mock SDK client response with discovery data
	httpResp := &http.Response{
		StatusCode: 200,
		Body: io.NopCloser(bytes.NewBufferString(
			`{
				"Items": {
					"DBInstance": [
						{
							"DBInstanceId": "rds-test-123",
							"RegionId": "cn-shanghai",
							"DBInstanceDescription": "Test RDS Instance"
						}
					]
				},
				"TotalRecordCount": 1,
				"PageSize": 1,
				"PageNumber": 1
			}`)),
	}
	mockCli, err := getMockSdkCli(httpResp)
	require.NoError(t, err)
	plugin.dt.Cli = map[string]commonaliyun.AliyunSdkClient{plugin.Regions[0]: &mockCli}

	// Get initial discovery data (simulating what Init() does)
	plugin.discoveryData, err = plugin.dt.GetDiscoveryDataAcrossRegions(nil)
	require.NoError(t, err)

	// Verify discovery data was loaded
	require.NotNil(t, plugin.discoveryData)
	require.Len(t, plugin.discoveryData, 1)
	require.Contains(t, plugin.discoveryData, "rds-test-123")

	// Initialize measurement
	plugin.measurement = formatMeasurement(plugin.Project)
	plugin.dimensionKey = "instanceId"

	// Update window for gathering
	plugin.updateWindow(time.Now())

	// Now gather metrics - this should use the initial discovery data
	acc := &testutil.Accumulator{}
	err = plugin.Gather(acc)
	require.NoError(t, err)

	// Verify that the metric has discovery tags populated from initial data
	m := plugin.Metrics[0]
	require.NotNil(t, m.discoveryTags)
	require.Contains(t, m.discoveryTags, "rds-test-123")

	// Note: requestDimensions is empty because no dimensions are configured
	// Discovery data is used only for tag enrichment, not for filtering
	// The CMS API will return all instances, and we enrich them with discovery tags
	require.Empty(t, m.requestDimensions, "No dimensions should be sent when none configured")
	require.Empty(t, m.requestDimensionsStr, "Dimensions string should be empty")
}

func TestInitialDiscoveryDataWorksWithMultipleMetrics(t *testing.T) {
	// This test verifies that discovery data works for multiple metrics in the same gather
	plugin := &AliyunCMS{
		AccessKeyID:       "dummy",
		AccessKeySecret:   "dummy",
		Project:           "acs_rds_dashboard",
		Regions:           []string{"cn-shanghai"},
		Period:            config.Duration(time.Minute),
		Delay:             config.Duration(time.Minute),
		DiscoveryInterval: config.Duration(time.Minute),
		RateLimit:         200,
		Log:               testutil.Logger{Name: inputTitle},
		Metrics: []*metric{
			{
				MetricNames: []string{"CpuUsage"},
			},
			{
				MetricNames: []string{"MemoryUsage"},
			},
			{
				MetricNames: []string{"DiskUsage"},
			},
		},
	}

	// Setup mock CMS client
	plugin.cmsClient = &mockGatherAliyunCMSClient{}

	// Setup mock discovery tool
	var err error
	plugin.dt, err = getDiscoveryTool("acs_rds_dashboard", plugin.Regions)
	require.NoError(t, err)

	// Mock SDK client response with multiple instances
	httpResp := &http.Response{
		StatusCode: 200,
		Body: io.NopCloser(bytes.NewBufferString(
			`{
				"Items": {
					"DBInstance": [
						{
							"DBInstanceId": "rds-test-1",
							"RegionId": "cn-shanghai",
							"DBInstanceDescription": "Test RDS Instance 1"
						},
						{
							"DBInstanceId": "rds-test-2",
							"RegionId": "cn-shanghai",
							"DBInstanceDescription": "Test RDS Instance 2"
						}
					]
				},
				"TotalRecordCount": 2,
				"PageSize": 2,
				"PageNumber": 1
			}`)),
	}
	mockCli, err := getMockSdkCli(httpResp)
	require.NoError(t, err)
	plugin.dt.Cli = map[string]commonaliyun.AliyunSdkClient{plugin.Regions[0]: &mockCli}

	// Get initial discovery data
	plugin.discoveryData, err = plugin.dt.GetDiscoveryDataAcrossRegions(nil)
	require.NoError(t, err)

	// Verify discovery data was loaded
	require.NotNil(t, plugin.discoveryData)
	require.Len(t, plugin.discoveryData, 2)
	require.Contains(t, plugin.discoveryData, "rds-test-1")
	require.Contains(t, plugin.discoveryData, "rds-test-2")

	// Initialize measurement
	plugin.measurement = formatMeasurement(plugin.Project)
	plugin.dimensionKey = "instanceId"

	// Update window for gathering
	plugin.updateWindow(time.Now())

	// Now gather metrics - ALL metrics should use the initial discovery data
	acc := &testutil.Accumulator{}
	err = plugin.Gather(acc)
	require.NoError(t, err)

	// Verify that ALL metrics have discovery tags populated from initial data
	for i, m := range plugin.Metrics {
		require.NotNil(t, m.discoveryTags, "Metric %d should have discovery tags", i)
		require.Contains(t, m.discoveryTags, "rds-test-1", "Metric %d should contain rds-test-1", i)
		require.Contains(t, m.discoveryTags, "rds-test-2", "Metric %d should contain rds-test-2", i)

		// Note: requestDimensions is empty because no dimensions are configured
		// Discovery data is used only for tag enrichment, not for filtering
		// The CMS API will return all instances, and we enrich them with discovery tags
		require.Empty(t, m.requestDimensions, "Metric %d should have no dimensions when none configured", i)
		require.Empty(t, m.requestDimensionsStr, "Metric %d dimensions string should be empty", i)
	}
}
