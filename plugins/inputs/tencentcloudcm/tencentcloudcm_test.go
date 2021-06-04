package tencentcloudcm

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

type mockGatherCloudMonitorClient struct{}

func (m *mockGatherCloudMonitorClient) GetMetricObjects(t TencentCloudCM) []MetricObject {
	return []MetricObject{
		{
			Metric:    "CPUUsage",
			Region:    "ap-hongkong",
			Namespace: "QCE/CVM",
			Account: &Account{
				Name: "name",
				Crs:  common.NewCredential("secret_id", "secret_key"),
			},
			Instances: []*Instance{{
				Dimensions: []*Dimension{{
					Name:  "InstanceId",
					Value: "ins-xxxxxxx1",
				}},
			}},
		},
	}
}

func (m *mockGatherCloudMonitorClient) NewClient(region string, crs *common.Credential, t TencentCloudCM) Client {
	return Client{}
}

func (m *mockGatherCloudMonitorClient) NewGetMonitorDataRequest(namespace, metric string, instances []*Instance, t TencentCloudCM) *GetMonitorDataRequest {
	return NewGetMonitorDataRequest()
}

func (m *mockGatherCloudMonitorClient) GatherMetrics(client Client, request *GetMonitorDataRequest, t TencentCloudCM) (*GetMonitorDataResponse, error) {
	response := &GetMonitorDataResponse{
		Response: &struct {
			Period     *uint64             `json:"Period,omitempty" name:"Period"`
			MetricName *string             `json:"MetricName,omitempty" name:"MetricName"`
			DataPoints []*MonitorDataPoint `json:"DataPoints,omitempty" name:"DataPoints"`
			StartTime  *string             `json:"StartTime,omitempty" name:"StartTime"`
			EndTime    *string             `json:"EndTime,omitempty" name:"EndTime"`
			RequestId  *string             `json:"RequestId,omitempty" name:"RequestId"`
		}{
			RequestId:  common.StringPtr("request_id"),
			Period:     common.Uint64Ptr(300),
			MetricName: common.StringPtr("CPUUsage"),
			DataPoints: []*MonitorDataPoint{
				{
					Dimensions: []*MonitorDimension{
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

	client := cm.client.NewClient("ap-hongkong", common.NewCredential(
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

	request := cm.client.NewGetMonitorDataRequest("QCE/CVM", "CPUUsage", []*Instance{
		{
			Dimensions: []*Dimension{
				{
					Name:  "InstanceId",
					Value: "ins-xxxxxxxx",
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
						Dimensions: []*Dimension{{
							Name:  "InstanceId",
							Value: "ins-xxxxxxx1",
						}},
					}},
				}}}, {
				Name:    "QCE/CDB",
				Metrics: []string{"CPUUseRate", "MemoryUseRate", "RealCapacity"},
				Regions: []*Region{{
					RegionName: "ap-hongkong",
					Instances: []*Instance{{
						Dimensions: []*Dimension{{
							Name:  "InstanceId",
							Value: "cdb-xxxxxxx1",
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
		"value": common.Float64Ptr(0.1),
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

	assert.True(t, acc.HasMeasurement("QCE/CVM"))
	acc.AssertContainsTaggedFields(t, "QCE/CVM", fields, tags)
}
