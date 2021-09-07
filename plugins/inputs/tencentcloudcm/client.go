package tencentcloudcm

import (
	"fmt"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
)

type cloudmonitorClient struct {
	Accounts []*Account
}

func (c *cloudmonitorClient) GetMetricObjects(t TencentCloudCM) []metricObject {
	// holds all metrics with it's corresponding region, namespace, credential and instances(dimensions) information.
	metricObjects := []metricObject{}

	// construct metric object
	for _, account := range t.Accounts {
		for _, namespace := range account.Namespaces {
			for _, region := range namespace.Regions {
				for _, metric := range namespace.Metrics {
					monitorInstances := region.Instances
					isDiscovered := false
					if len(monitorInstances) == 0 {
						// if instances are not specified. look them up in the discoverTool
						monitorInstances = t.discoverTool.GetMonitorInstances(account.Name, namespace.Name, region.RegionName)
						isDiscovered = true
					}
					if len(monitorInstances) == 0 {
						continue
					}
					metricObjects = append(metricObjects, metricObject{
						Metric:    metric,
						Region:    region.RegionName,
						Namespace: namespace.Name,
						Account:   account,

						isDiscovered: isDiscovered,

						MonitorInstances: monitorInstances,
					})
				}
			}
		}
	}
	return metricObjects
}

func (c *cloudmonitorClient) NewClient(region string, crs *common.Credential, t TencentCloudCM) (monitor.Client, error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = fmt.Sprintf("monitor.%s", t.Endpoint)
	cpf.HttpProfile.ReqTimeout = int(time.Duration(t.Timeout).Milliseconds()) / 1000
	client, err := monitor.NewClient(crs, region, cpf)
	return *client, err
}

func (c *cloudmonitorClient) NewGetMonitorDataRequest(namespace, metric string, instances []*monitor.Instance, t TencentCloudCM) *monitor.GetMonitorDataRequest {
	request := monitor.NewGetMonitorDataRequest()
	request.Namespace = common.StringPtr(namespace)
	request.MetricName = common.StringPtr(metric)
	period := uint64(time.Duration(t.Period).Seconds())
	request.Period = &period
	request.StartTime = common.StringPtr(t.windowStart.Format(time.RFC3339))
	request.EndTime = common.StringPtr(t.windowEnd.Format(time.RFC3339))
	request.Instances = []*monitor.Instance{}
	// Transform instances and dimensions from config to monitor struct
	request.Instances = instances
	return request
}

func (c *cloudmonitorClient) GatherMetrics(client monitor.Client, request *monitor.GetMonitorDataRequest, t TencentCloudCM) (*monitor.GetMonitorDataResponse, error) {
	response, err := client.GetMonitorData(request)
	if err != nil {
		return nil, fmt.Errorf("getting monitoring data for namespace %q failed: %v", *request.Namespace, err)
	}
	return response, nil
}
