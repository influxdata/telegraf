package azure_monitor

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
)

const (
	minMetricsFields      = 2
	timeSeriesIsEmpty     = "timeseries is empty"
	dataIsEmpty           = "data is empty"
	noDataInMetricsFields = "no data in metric fields"
)

func (am *AzureMonitor) buildMetricValuesAPIURL(target *ResourceTarget) string {
	metrics := strings.Join(target.Metrics, ",")
	metrics = strings.Replace(metrics, " ", "+", -1)
	apiURL := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/%s/providers/microsoft.insights/metrics?metricnames=%s&"+
			"aggregation=%s&api-version=2019-07-01",
		am.SubscriptionID, target.ResourceID, metrics, strings.Join(target.Aggregations, ","))

	return apiURL
}

func (am *AzureMonitor) collectResourceTargetsMetrics(acc telegraf.Accumulator) {
	var waitGroup sync.WaitGroup

	for _, target := range am.ResourceTargets {
		am.Log.Debug("Collecting metrics for resource target ", target.ResourceID)
		waitGroup.Add(1)

		go func(target *ResourceTarget) {
			defer waitGroup.Done()

			apiURL := am.buildMetricValuesAPIURL(target)
			body, err := am.getAPIResponseBody(apiURL)
			if err != nil {
				acc.AddError(fmt.Errorf("error getting metric values API response body for resource target %s: %v",
					target.ResourceID, err))
				return
			}

			if err = am.collectResourceTargetMetrics(body, acc); err != nil {
				acc.AddError(fmt.Errorf("error collecting resource target %s metrics: %v", target.ResourceID, err))
				return
			}
		}(target)

		waitGroup.Wait()
	}
}

func (am *AzureMonitor) collectResourceTargetMetrics(body map[string]interface{}, acc telegraf.Accumulator) error {
	values, ok := body["value"].([]interface{})
	if !ok {
		return fmt.Errorf("metric values API response body bad format: value is missing")
	}

	for _, value := range values {
		timesSeries, ok := value.(map[string]interface{})["timeseries"].([]interface{})
		if !ok {
			return fmt.Errorf("metric values API response body bad format: timeseries of value is missing")
		}

		if len(timesSeries) == 0 {
			if err := am.writeNoMetricDataLog(body, value.(map[string]interface{}), timeSeriesIsEmpty); err != nil {
				return fmt.Errorf("error getting metric details: %v", err)
			}
			continue
		}

		timeSeries := timesSeries[0].(map[string]interface{})
		data, ok := timeSeries["data"].([]interface{})
		if !ok {
			return fmt.Errorf("metric values API response body bad format: data of timeseries is missing")
		}

		if len(data) == 0 {
			if err := am.writeNoMetricDataLog(body, value.(map[string]interface{}), dataIsEmpty); err != nil {
				return fmt.Errorf("error getting metric details: %v", err)
			}
			continue
		}

		metricName, err := getMetricName(body, value.(map[string]interface{}))
		if err != nil {
			return fmt.Errorf("error getting metric name: %v", err)
		}

		metricFields := getMetricFields(data)
		if metricFields == nil {
			if err = am.writeNoMetricDataLog(body, value.(map[string]interface{}), noDataInMetricsFields); err != nil {
				return fmt.Errorf("error getting metric details: %v", err)
			}
			continue
		}

		metricTags, err := getMetricTags(body, value.(map[string]interface{}))
		if err != nil {
			return fmt.Errorf("error getting metric tags: %v", err)
		}

		acc.AddFields(*metricName, metricFields, metricTags, time.Now())
	}

	return nil
}

func (am *AzureMonitor) writeNoMetricDataLog(body map[string]interface{}, value map[string]interface{}, reason string) error {
	name, err := getMetricValuesValueName(value)
	if err != nil {
		return err
	}

	metricName, ok := name["value"].(string)
	if !ok {
		return fmt.Errorf("metric values API response body bad format: value of name is missing")
	}

	resourceGroupName, err := getResourceGroupName(value)
	if err != nil {
		return err
	}

	resourceName, err := getResourceName(value)
	if err != nil {
		return err
	}

	namespace, err := getMetricValuesNamespace(body)
	if err != nil {
		return err
	}

	am.Log.Info("No data from Azure Monitor API about metric: ", metricName, " resource group: ",
		*resourceGroupName, " resource: ", *resourceName, " type: ", *namespace, " (", reason, ")")

	return nil
}

func getMetricName(body map[string]interface{}, value map[string]interface{}) (*string, error) {
	namespace, err := getMetricValuesNamespace(body)
	if err != nil {
		return nil, err
	}

	replacer := strings.NewReplacer(".", "_", "/", "_", " ", "_", "(", "_", ")", "_")
	name, err := getMetricValuesValueName(value)
	if err != nil {
		return nil, err
	}

	localizedValue, ok := name["localizedValue"].(string)
	if !ok {
		return nil, fmt.Errorf("localizedValue key in name is missing in metric values API response body")
	}

	metricName := fmt.Sprintf("azure_monitor_%s_%s",
		replacer.Replace(strings.ToLower(*namespace)),
		replacer.Replace(strings.ToLower(localizedValue)))

	return &metricName, nil
}

func getMetricFields(data []interface{}) map[string]interface{} {
	for index := len(data) - 1; index >= 0; index-- {
		if len(data[index].(map[string]interface{})) < minMetricsFields {
			continue
		}

		return data[index].(map[string]interface{})
	}

	return nil
}

func getMetricTags(body map[string]interface{}, value map[string]interface{}) (map[string]string, error) {
	tags := make(map[string]string)
	subscriptionID, err := getSubscriptionID(value)
	if err != nil {
		return nil, err
	}

	tags["subscription_id"] = *subscriptionID

	resourceGroupName, err := getResourceGroupName(value)
	if err != nil {
		return nil, err
	}

	tags["resource_group"] = *resourceGroupName

	resourceName, err := getResourceName(value)
	if err != nil {
		return nil, err
	}

	tags["resource_name"] = *resourceName

	namespace, err := getMetricValuesNamespace(body)
	if err != nil {
		return nil, err
	}

	tags["namespace"] = *namespace

	resourceRegion, ok := body["resourceregion"].(string)
	if !ok {
		return nil, fmt.Errorf("metric values API response body bad format: resourceregion is missing")
	}

	tags["resource_region"] = resourceRegion

	unit, ok := value["unit"].(string)
	if !ok {
		return nil, fmt.Errorf("metric values API response body bad format: unit of value is missing")
	}

	tags["unit"] = unit

	return tags, nil
}

func getMetricValuesNamespace(body map[string]interface{}) (*string, error) {
	namespace, ok := body["namespace"].(string)
	if !ok {
		return nil, fmt.Errorf("metric values API response body bad format: namespace is missing")
	}

	return &namespace, nil
}
func getMetricValuesValueID(value map[string]interface{}) (*string, error) {
	resourceID, ok := value["id"].(string)
	if !ok {
		return nil, fmt.Errorf("metric values API response body bad format: id of value is missing")
	}

	return &resourceID, nil
}

func getMetricValuesValueName(value map[string]interface{}) (map[string]interface{}, error) {
	name, ok := value["name"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("name key in value is missing in metric values API response body")
	}

	return name, nil
}

func getPartOfResourceID(resourceID string, partIndex int, partSubPartIndex int, getPartToEnd bool) (*string, error) {
	resourceIDParts := strings.Split(resourceID, "/providers/")
	if len(resourceIDParts) <= partIndex {
		return nil, fmt.Errorf("metric values API value id bad format")
	}

	resourceIDPartSubParts := strings.Split(resourceIDParts[partIndex], "/")
	if len(resourceIDPartSubParts) <= partSubPartIndex {
		return nil, fmt.Errorf("metric values API value id bad format")
	}

	var part string
	if getPartToEnd {
		part = strings.Join(resourceIDPartSubParts[partSubPartIndex:], "/")
	} else {
		part = resourceIDPartSubParts[partSubPartIndex]
	}

	return &part, nil
}

func getSubscriptionID(value map[string]interface{}) (*string, error) {
	resourceID, err := getMetricValuesValueID(value)
	if err != nil {
		return nil, err
	}

	subscriptionID, err := getPartOfResourceID(*resourceID, 0, 2, false)
	if err != nil {
		return nil, err
	}

	return subscriptionID, nil
}

func getResourceGroupName(value map[string]interface{}) (*string, error) {
	resourceID, err := getMetricValuesValueID(value)
	if err != nil {
		return nil, err
	}

	resourceGroupName, err := getPartOfResourceID(*resourceID, 0, 4, false)
	if err != nil {
		return nil, err
	}

	return resourceGroupName, nil
}

func getResourceName(value map[string]interface{}) (*string, error) {
	resourceID, err := getMetricValuesValueID(value)
	if err != nil {
		return nil, err
	}

	resourceGroupName, err := getPartOfResourceID(*resourceID, 1, 2, true)
	if err != nil {
		return nil, err
	}

	return resourceGroupName, nil
}
