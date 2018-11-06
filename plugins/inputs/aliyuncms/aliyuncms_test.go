package aliyuncms

import (
	"errors"
	"testing"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

type mockGatherAliyunCMSClient struct{}

func (m *mockGatherAliyunCMSClient) QueryMetricList(request *cms.QueryMetricListRequest) (*cms.QueryMetricListResponse, error) {
	resp := new(cms.QueryMetricListResponse)

	switch request.Metric {
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
	case "ErrorResp":
		return nil, errors.New("error response")
	}
	return resp, nil
}

func TestInit(t *testing.T) {
	require.Equal(t, &AliyunCMS{RateLimit: 200}, inputs.Inputs["aliyuncms"]())
}

func TestGatherMetric(t *testing.T) {
	duration, _ := time.ParseDuration("1m")
	internalDuration := internal.Duration{
		Duration: duration,
	}

	var acc telegraf.Accumulator
	s := &AliyunCMS{
		Period:  internalDuration,
		Delay:   internalDuration,
		Project: "acs_slb_dashboard",
		client:  new(mockGatherAliyunCMSClient),
	}

	dimension := &Dimension{
		Value: `"instanceId": "p-example"`,
	}
	require.EqualError(t, s.gatherMetric(acc, "DecodeError", dimension), `failed to decode "instanceId": "p-example": invalid character ':' after top-level value`)

	dimension = &Dimension{
		Value: `{"instanceId": "p-example"}`,
	}
	require.EqualError(t, s.gatherMetric(acc, "ErrorCode", dimension), "failed to query metric list: ErrorCode")
	require.EqualError(t, s.gatherMetric(acc, "ErrorDatapoint", dimension),
		`failed to decode response datapoints: invalid character '}' looking for beginning of object key string`)
	require.EqualError(t, s.gatherMetric(acc, "ErrorResp", dimension), "failed to query metric list: error response")
}

func TestGather(t *testing.T) {
	metric := &Metric{
		MetricNames: []string{},
		Dimensions: []*Dimension{
			{
				Value: `{"instanceId": "p-example"}`,
			},
		},
	}

	s := &AliyunCMS{
		AccessKeyID:     "my_access_key_id",
		AccessKeySecret: "my_access_key_secret",
		Project:         "acs_slb_dashboard",
		Metrics:         []*Metric{metric},
		RateLimit:       200,
	}

	// initialize error
	var acc testutil.Accumulator
	require.EqualError(t, acc.GatherError(s.Gather), "region id is not set")

	s.RegionID = "cn-shanghai"
	s.client = new(mockGatherAliyunCMSClient)
	// empty datapoint test
	s.Metrics[0].MetricNames = []string{"EmptyDatapoint"}
	require.Empty(t, acc.GatherError(s.Gather))
	require.False(t, acc.HasMeasurement("aliyuncms_acs_slb_dashboard"))

	// correct data test
	fields := map[string]interface{}{
		"instance_active_connection_minimum": float64(100),
		"instance_active_connection_maximum": float64(200),
		"instance_active_connection_average": float64(150),
		"instance_active_connection_value":   float64(300),
	}

	tags := map[string]string{
		"regionId":   "cn-shanghai",
		"instanceId": "p-example",
		"userId":     "1234567898765432",
	}

	s.Metrics[0].MetricNames = []string{"InstanceActiveConnection"}
	require.Empty(t, acc.GatherError(s.Gather))
	require.True(t, acc.HasMeasurement("aliyuncms_acs_slb_dashboard"))
	acc.AssertContainsTaggedFields(t, "aliyuncms_acs_slb_dashboard", fields, tags)
}

func TestUpdateWindow(t *testing.T) {
	duration, _ := time.ParseDuration("1m")
	internalDuration := internal.Duration{
		Duration: duration,
	}

	s := &AliyunCMS{
		Project: "acs_slb_dashboard",
		Period:  internalDuration,
		Delay:   internalDuration,
	}

	now := time.Now()

	require.True(t, s.windowEnd.IsZero())
	require.True(t, s.windowStart.IsZero())

	s.updateWindow(now)

	newStartTime := s.windowEnd

	// initial window just has a single period
	require.EqualValues(t, s.windowEnd, now.Add(-s.Delay.Duration))
	require.EqualValues(t, s.windowStart, now.Add(-s.Delay.Duration).Add(-s.Period.Duration))

	now = time.Now()
	s.updateWindow(now)

	// subsequent window uses previous end time as start time
	require.EqualValues(t, s.windowEnd, now.Add(-s.Delay.Duration))
	require.EqualValues(t, s.windowStart, newStartTime)
}

func TestInitializeAliyunCMS(t *testing.T) {
	s := new(AliyunCMS)
	require.EqualError(t, s.initializeAliyunCMS(), "region id is not set")

	s.RegionID = "cn-shanghai"
	require.EqualError(t, s.initializeAliyunCMS(), "project is not set")

	s.Project = "acs_slb_dashboard"
	require.EqualError(t, s.initializeAliyunCMS(), "failed to retrieve credential")

	s.AccessKeyID = "my_access_key_id"
	s.AccessKeySecret = "my_access_key_secret"
	require.Equal(t, nil, s.initializeAliyunCMS())
}

func TestSampleConfig(t *testing.T) {
	s := new(AliyunCMS)
	require.Equal(t, sampleConfig, s.SampleConfig())
}

func TestDescription(t *testing.T) {
	s := new(AliyunCMS)
	require.Equal(t, description, s.Description())
}
