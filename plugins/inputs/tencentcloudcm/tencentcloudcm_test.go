package tencentcloudcm

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
)

type mockGatherCloudMonitorClient struct{}

func (m *mockGatherCloudMonitorClient) GetMetricObjects(t TencentCloudCM) []metricObject {
	return []metricObject{
		{
			Metric:    "CPUUsage",
			Region:    "ap-hongkong",
			Namespace: "QCE/CVM",
			Account: &Account{
				Name: "name",
				crs:  common.NewCredential("secret_id", "secret_key"),
			},
			MonitorInstances: []*monitor.Instance{{
				Dimensions: []*monitor.Dimension{{
					Name:  common.StringPtr("InstanceId"),
					Value: common.StringPtr("ins-xxxxxxx1"),
				}},
			}},
		},
	}
}

func (m *mockGatherCloudMonitorClient) NewClient(region string, crs *common.Credential, t TencentCloudCM) (monitor.Client, error) {
	return monitor.Client{}, nil
}

func (m *mockGatherCloudMonitorClient) NewGetMonitorDataRequest(namespace, metric string, instances []*monitor.Instance, t TencentCloudCM) *monitor.GetMonitorDataRequest {
	return monitor.NewGetMonitorDataRequest()
}

func (m *mockGatherCloudMonitorClient) GatherMetrics(client monitor.Client, request *monitor.GetMonitorDataRequest, t TencentCloudCM) (*monitor.GetMonitorDataResponse, error) {
	response := &monitor.GetMonitorDataResponse{
		Response: &struct {
			Period     *uint64              `json:"Period,omitempty" name:"Period"`
			MetricName *string              `json:"MetricName,omitempty" name:"MetricName"`
			DataPoints []*monitor.DataPoint `json:"DataPoints,omitempty" name:"DataPoints"`
			StartTime  *string              `json:"StartTime,omitempty" name:"StartTime"`
			EndTime    *string              `json:"EndTime,omitempty" name:"EndTime"`
			RequestId  *string              `json:"RequestId,omitempty" name:"RequestId"`
		}{
			RequestId:  common.StringPtr("request_id"),
			Period:     common.Uint64Ptr(300),
			MetricName: common.StringPtr("CPUUsage"),
			DataPoints: []*monitor.DataPoint{
				{
					Dimensions: []*monitor.Dimension{
						{
							Name:  common.StringPtr("InstanceId"),
							Value: common.StringPtr("ins-xxxxxxx1"),
						},
					},
					Timestamps: []*float64{
						common.Float64Ptr(1618588800),
					},
					Values: []*float64{
						common.Float64Ptr(0.1),
					},
				},
			},
		},
	}
	return response, nil
}

func TestUpdateWindow(t *testing.T) {
	cm := &TencentCloudCM{
		Period: config.Duration(1 * time.Minute),
		Delay:  config.Duration(5 * time.Minute),
	}

	now := time.Now()

	assert.True(t, cm.windowEnd.IsZero())
	assert.True(t, cm.windowStart.IsZero())

	cm.updateWindow(now)

	assert.EqualValues(t, cm.windowEnd, now.Add(-time.Duration(cm.Delay)))
	assert.EqualValues(t, cm.windowStart, now.Add(-time.Duration(cm.Delay)).Add(-time.Duration(cm.Period)*2))
}

func TestNewClient(t *testing.T) {
	cm := &TencentCloudCM{
		Period: config.Duration(1 * time.Minute),
		Delay:  config.Duration(5 * time.Minute),
		client: &cloudmonitorClient{},
	}

	client, _ := cm.client.NewClient("ap-hongkong", common.NewCredential(
		"secret_id",
		"secret_key",
	), *cm)

	assert.NotZero(t, client)
	assert.EqualValues(t, "ap-hongkong", client.GetRegion())
}

func TestNewGetMonitorDataRequest(t *testing.T) {
	cm := &TencentCloudCM{
		Period: config.Duration(1 * time.Minute),
		Delay:  config.Duration(5 * time.Minute),
		client: &cloudmonitorClient{},
	}

	request := cm.client.NewGetMonitorDataRequest("QCE/CVM", "CPUUsage", []*monitor.Instance{
		{
			Dimensions: []*monitor.Dimension{
				{
					Name:  common.StringPtr("InstanceId"),
					Value: common.StringPtr("ins-xxxxxxxx"),
				},
			},
		},
	}, *cm)

	assert.EqualValues(t, "QCE/CVM", *request.Namespace)
	assert.EqualValues(t, "CPUUsage", *request.MetricName)

	assert.Len(t, request.Instances, 1)
	assert.Len(t, request.Instances[len(request.Instances)-1].Dimensions, 1)

	assert.EqualValues(t, "InstanceId", *request.Instances[len(request.Instances)-1].Dimensions[0].Name)
	assert.EqualValues(t, "ins-xxxxxxxx", *request.Instances[len(request.Instances)-1].Dimensions[0].Value)
}

func TestGetMetricObjects(t *testing.T) {
	cm := &TencentCloudCM{
		Period: config.Duration(1 * time.Minute),
		Delay:  config.Duration(5 * time.Minute),
		client: &cloudmonitorClient{},
		Accounts: []*Account{{
			Name:      "name",
			SecretID:  "secret_id",
			SecretKey: "secret_key",
			Namespaces: []*Namespace{{
				Name:    "QCE/CVM",
				Metrics: []string{"CPUUsage", "MemUsage", "MemUsed"},
				Regions: []*Region{{
					RegionName: "ap-hongkong",
					Instances: []*Instance{{
						Dimensions: []map[string]string{{
							"InstanceId": "ins-xxxxxxx1",
						}},
					}},
				}}}, {
				Name:    "QCE/CDB",
				Metrics: []string{"CPUUseRate", "MemoryUseRate", "RealCapacity"},
				Regions: []*Region{{
					RegionName: "ap-hongkong",
					Instances: []*Instance{{
						Dimensions: []map[string]string{{
							"InstanceId": "cdb-xxxxxxx1",
						}},
					}},
				}},
			}},
		}},
	}

	metricObjects := cm.client.GetMetricObjects(*cm)

	assert.Len(t, metricObjects, 6)
}

func TestGather(t *testing.T) {
	cm := &TencentCloudCM{
		Period:    config.Duration(1 * time.Minute),
		Delay:     config.Duration(5 * time.Minute),
		RateLimit: 20,
	}

	var acc testutil.Accumulator
	cm.client = &mockGatherCloudMonitorClient{}

	assert.NoError(t, acc.GatherError(cm.Gather))

	fields := map[string]interface{}{
		"CPUUsage": common.Float64Ptr(0.1),
	}

	tags := map[string]string{
		"InstanceId": "ins-xxxxxxx1",
		"account":    "name",
		"metric":     "CPUUsage",
		"namespace":  "QCE/CVM",
		"period":     "300",
		"region":     "ap-hongkong",
		"request_id": "request_id",
	}

	assert.True(t, acc.HasMeasurement("tencentcloudcm_QCE/CVM"))
	acc.AssertContainsTaggedFields(t, "tencentcloudcm_QCE/CVM", fields, tags)
}
