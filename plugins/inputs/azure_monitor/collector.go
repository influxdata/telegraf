package azure_monitor

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/influxdata/telegraf"
)

const (
	minMetricsFields = 2

	metricTagSubscriptionID = "subscription_id"
	metricTagResourceGroup  = "resource_group"
	metricTagResourceName   = "resource_name"
	metricTagNamespace      = "namespace"
	metricTagResourceRegion = "resource_region"
	metricTagUnit           = "unit"

	timeseriesIsEmpty     = "timeseries is empty"
	dataIsEmpty           = "data is empty"
	noDataInMetricsFields = "no data in metric fields"
)

func (am *AzureMonitor) collectResourceTargetsMetrics(acc telegraf.Accumulator) {
	var waitGroup sync.WaitGroup

	for _, target := range am.ResourceTargets {
		am.Log.Debug("Collecting metrics for resource target ", target.ResourceID)
		waitGroup.Add(1)

		go func(target *ResourceTarget) {
			defer waitGroup.Done()

			metricNames := strings.Join(target.Metrics, ",")
			aggregations := strings.Join(target.Aggregations, ",")
			response, err := am.azureClients.metricsClient.List(am.azureClients.ctx, target.ResourceID, &armmonitor.MetricsClientListOptions{
				Metricnames: &metricNames,
				Aggregation: &aggregations,
			})
			if err != nil {
				acc.AddError(fmt.Errorf("error listing metrics for the resource target %s: %v", target.ResourceID, err))
			}

			if err = am.collectResourceTargetMetrics(&response, acc); err != nil {
				acc.AddError(fmt.Errorf("error collecting resource target %s metrics: %v", target.ResourceID, err))
				return
			}
		}(target)

		waitGroup.Wait()
	}
}

func (am *AzureMonitor) collectResourceTargetMetrics(response *armmonitor.MetricsClientListResponse, acc telegraf.Accumulator) error {
	for _, metric := range response.Value {
		errorMessage, err := getMetricsClientMetricErrorMessage(metric)
		if err != nil {
			return err
		}

		if errorMessage != nil {
			return fmt.Errorf("response error: %s", *errorMessage)
		}

		if len(metric.Timeseries) == 0 {
			if err = am.writeNoMetricDataLog(metric, timeseriesIsEmpty); err != nil {
				return fmt.Errorf("error writing no metric data log: %v", err)
			}
			continue
		}

		timeseries := metric.Timeseries[0]
		if timeseries == nil {
			return fmt.Errorf("metrics client response is bad formatted: metric timeseries is missing")
		}

		if len(timeseries.Data) == 0 {
			if err = am.writeNoMetricDataLog(metric, dataIsEmpty); err != nil {
				return fmt.Errorf("error writing no metric data log: %v", err)
			}
			continue
		}

		metricName, err := createMetricName(metric, response)
		if err != nil {
			return fmt.Errorf("error creating metric name: %v", err)
		}

		metricFields := getMetricFields(timeseries.Data)
		if metricFields == nil {
			if err = am.writeNoMetricDataLog(metric, noDataInMetricsFields); err != nil {
				return fmt.Errorf("error writing no metric data log: %v", err)
			}
			continue
		}

		metricTags, err := getMetricTags(metric, response)
		if err != nil {
			return fmt.Errorf("error getting metric tags: %v", err)
		}

		acc.AddFields(*metricName, metricFields, metricTags, time.Now())
	}

	return nil
}

func (am *AzureMonitor) writeNoMetricDataLog(metric *armmonitor.Metric, reason string) error {
	metricID, err := getMetricsClientMetricID(metric)
	if err != nil {
		return err
	}

	am.Log.Info("No data from Azure Monitor API about metric ", *metricID, ": ", reason)

	return nil
}

func getMetricsClientMetricErrorMessage(metric *armmonitor.Metric) (*string, error) {
	if metric == nil {
		return nil, fmt.Errorf("metrics client response is bad formatted: metric is missing")
	}

	if metric.ErrorCode == nil {
		return nil, fmt.Errorf("metrics client response is bad formatted: metric ErrorCode is missing")
	}

	if *metric.ErrorCode == "Success" {
		return nil, nil
	}

	errorMessage := fmt.Sprintf("error code %s: %s", *metric.ErrorCode, *metric.ErrorMessage)
	return &errorMessage, nil
}

func getMetricsClientMetricID(metric *armmonitor.Metric) (*string, error) {
	if metric == nil {
		return nil, fmt.Errorf("metrics client response is bad formatted: metric is missing")
	}

	if metric.ID == nil {
		return nil, fmt.Errorf("metrics client response is bad formatted: metric ID is missing")
	}

	return metric.ID, nil
}

func getMetricsClientMetricNameLocalizedValue(metric *armmonitor.Metric) (*string, error) {
	if metric == nil {
		return nil, fmt.Errorf("metrics client response is bad formatted: metric is missing")
	}

	metricName := metric.Name
	if metricName == nil {
		return nil, fmt.Errorf("metrics client response is bad formatted: metric Name is missing")
	}

	metricNameLocalizedValue := metricName.LocalizedValue
	if metricNameLocalizedValue == nil {
		return nil, fmt.Errorf("metrics client response is bad formatted: metric Name.LocalizedValue is missing")
	}

	return metricNameLocalizedValue, nil
}

func getMetricsClientMetricUnit(metric *armmonitor.Metric) (*string, error) {
	if metric == nil {
		return nil, fmt.Errorf("metrics client response is bad formatted: metric is missing")
	}

	if metric.Unit == nil {
		return nil, fmt.Errorf("metrics client response is bad formatted: metric Unit is missing")
	}

	metricUnit := string(*metric.Unit)
	return &metricUnit, nil
}

func getMetricsClientResponseNamespace(response *armmonitor.MetricsClientListResponse) (*string, error) {
	if response.Namespace == nil {
		return nil, fmt.Errorf("metrics client response is bad formatted: reponse Namespace is missing")
	}

	return response.Namespace, nil
}

func getMetricsClientResponseResourceRegion(response *armmonitor.MetricsClientListResponse) (*string, error) {
	if response.Resourceregion == nil {
		return nil, fmt.Errorf("metrics client response is bad formatted: reponse Resourceregion is missing")
	}

	return response.Resourceregion, nil
}

func getMetricsClientMetricValueFields(metricValue *armmonitor.MetricValue) map[string]interface{} {
	if metricValue == nil {
		return nil
	}

	if metricValue.TimeStamp == nil {
		return nil
	}

	metricFields := make(map[string]interface{})
	metricValueFieldsNum := 1

	metricFields["timeStamp"] = metricValue.TimeStamp.Format("2006-01-02T15:04:05Z07:00")

	if metricValue.Total != nil {
		metricFields["total"] = *metricValue.Total
		metricValueFieldsNum++
	}

	if metricValue.Average != nil {
		metricFields["average"] = *metricValue.Average
		metricValueFieldsNum++
	}

	if metricValue.Count != nil {
		metricFields["count"] = *metricValue.Count
		metricValueFieldsNum++
	}

	if metricValue.Minimum != nil {
		metricFields["minimum"] = *metricValue.Minimum
		metricValueFieldsNum++
	}

	if metricValue.Maximum != nil {
		metricFields["maximum"] = *metricValue.Maximum
		metricValueFieldsNum++
	}

	if metricValueFieldsNum < minMetricsFields {
		return nil
	}

	return metricFields
}

func createMetricName(metric *armmonitor.Metric, response *armmonitor.MetricsClientListResponse) (*string, error) {
	namespace, err := getMetricsClientResponseNamespace(response)
	if err != nil {
		return nil, err
	}

	name, err := getMetricsClientMetricNameLocalizedValue(metric)
	if err != nil {
		return nil, err
	}

	replacer := strings.NewReplacer(".", "_", "/", "_", " ", "_", "(", "_", ")", "_")
	metricName := fmt.Sprintf("azure_monitor_%s_%s",
		replacer.Replace(strings.ToLower(*namespace)),
		replacer.Replace(strings.ToLower(*name)))

	return &metricName, nil
}

func getMetricFields(metricValues []*armmonitor.MetricValue) map[string]interface{} {
	for index := len(metricValues) - 1; index >= 0; index-- {
		metricFields := getMetricsClientMetricValueFields(metricValues[index])
		if metricFields == nil {
			continue
		}

		return metricFields
	}

	return nil
}

func getMetricTags(metric *armmonitor.Metric, response *armmonitor.MetricsClientListResponse) (map[string]string, error) {
	tags := make(map[string]string)
	subscriptionID, err := getMetricSubscriptionID(metric)
	if err != nil {
		return nil, err
	}

	tags[metricTagSubscriptionID] = *subscriptionID

	resourceGroupName, err := getMetricResourceGroupName(metric)
	if err != nil {
		return nil, err
	}

	tags[metricTagResourceGroup] = *resourceGroupName

	resourceName, err := getMetricResourceName(metric)
	if err != nil {
		return nil, err
	}

	tags[metricTagResourceName] = *resourceName

	namespace, err := getMetricsClientResponseNamespace(response)
	if err != nil {
		return nil, err
	}

	tags[metricTagNamespace] = *namespace

	resourceRegion, err := getMetricsClientResponseResourceRegion(response)
	if err != nil {
		return nil, err
	}

	tags[metricTagResourceRegion] = *resourceRegion

	unit, err := getMetricsClientMetricUnit(metric)
	if err != nil {
		return nil, err
	}

	tags[metricTagUnit] = *unit

	return tags, nil
}

func getMetricSubscriptionID(metric *armmonitor.Metric) (*string, error) {
	metricID, err := getMetricsClientMetricID(metric)
	if err != nil {
		return nil, err
	}

	subscriptionID, err := getPartOfMetricID(*metricID, 0, 2, false)
	if err != nil {
		return nil, err
	}

	return subscriptionID, nil
}

func getMetricResourceGroupName(metric *armmonitor.Metric) (*string, error) {
	metricID, err := getMetricsClientMetricID(metric)
	if err != nil {
		return nil, err
	}

	resourceGroupName, err := getPartOfMetricID(*metricID, 0, 4, false)
	if err != nil {
		return nil, err
	}

	return resourceGroupName, nil
}

func getMetricResourceName(metric *armmonitor.Metric) (*string, error) {
	metricID, err := getMetricsClientMetricID(metric)
	if err != nil {
		return nil, err
	}

	resourceGroupName, err := getPartOfMetricID(*metricID, 1, 2, true)
	if err != nil {
		return nil, err
	}

	return resourceGroupName, nil
}

func getPartOfMetricID(metricID string, partIndex int, partSubPartIndex int, getPartToEnd bool) (*string, error) {
	metricIDParts := strings.Split(metricID, "/providers/")
	if len(metricIDParts) <= partIndex {
		return nil, fmt.Errorf("metrics client response is bad formatted: metric ID is bad formatted")
	}

	resourceIDPartSubParts := strings.Split(metricIDParts[partIndex], "/")
	if len(resourceIDPartSubParts) <= partSubPartIndex {
		return nil, fmt.Errorf("metrics client response is bad formatted: metric ID is bad formatted")
	}

	var part string

	if getPartToEnd {
		part = strings.Join(resourceIDPartSubParts[partSubPartIndex:], "/")
	} else {
		part = resourceIDPartSubParts[partSubPartIndex]
	}

	return &part, nil
}
