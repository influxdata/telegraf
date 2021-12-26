package azure_monitor

import (
	"fmt"
	"strings"
)

const (
	maxMetricsPerRequest   = 20
	totalAggregationName   = "Total"
	countAggregationName   = "Count"
	averageAggregationName = "Average"
	minAggregationName     = "Minimum"
	maxAggregationName     = "Maximum"
)

func newResourceTarget(resourceID string, metrics []string, aggregations []string) *ResourceTarget {
	return &ResourceTarget{
		ResourceID:   resourceID,
		Metrics:      metrics,
		Aggregations: aggregations,
	}
}

func newResourceGroupTarget(resourceGroup string, resources []*Resource) *ResourceGroupTarget {
	return &ResourceGroupTarget{
		ResourceGroup: resourceGroup,
		Resources:     resources,
	}
}

func (am *AzureMonitor) buildMetricDefinitionsAPIURL(resourceTargetResourceID string) string {
	apiURL := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/%s/providers/microsoft.insights/metricDefinitions?api-version=2018-01-01",
		am.SubscriptionID, resourceTargetResourceID)

	return apiURL
}

func (am *AzureMonitor) buildResourceGroupResourcesAPIURL(resourceGroup string) string {
	apiURL := fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourceGroups/%s/resources?api-version=2018-02-01",
		am.SubscriptionID, resourceGroup)

	return apiURL
}

func (am *AzureMonitor) buildSubscriptionResourceGroupsAPIURL() string {
	apiURL := fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourceGroups?api-version=2018-02-01",
		am.SubscriptionID)

	return apiURL
}

func (am *AzureMonitor) checkConfigValidation() error {
	if am.SubscriptionID == "" {
		return fmt.Errorf("subscription_id is empty or missing. Please check your configuration")
	}

	if am.ClientID == "" {
		return fmt.Errorf("client_id is empty or missing. Please check your configuration")
	}

	if am.ClientSecret == "" {
		return fmt.Errorf("client_secret is empty or missing. Please check your configuration")
	}

	if am.TenantID == "" {
		return fmt.Errorf("tenant_id is empty or missing. Please check your configuration")
	}

	if len(am.ResourceTargets) == 0 && len(am.ResourceGroupTargets) == 0 && len(am.SubscriptionTargets) == 0 {
		return fmt.Errorf("no target to collect metrics from in your configuration. Please check your configuration")
	}

	if err := am.checkResourceTargetsValidation(); err != nil {
		return err
	}

	if err := am.checkResourceGroupTargetsValidation(); err != nil {
		return err
	}

	if err := am.checkSubscriptionTargetValidation(); err != nil {
		return err
	}

	return nil
}

func (am *AzureMonitor) checkResourceTargetsValidation() error {
	for index, target := range am.ResourceTargets {
		if target.ResourceID == "" {
			return fmt.Errorf(
				"resource target #%d resource_id is empty or missing. Please check your configuration", index+1)
		}

		if len(target.Aggregations) > 0 {
			if !areTargetAggregationsValid(target.Aggregations) {
				return fmt.Errorf("resource target #%d aggregations contain invalid aggregation/s. "+
					"Please check your configuration. The valid aggregations are: %s, %s, %s, %s, %s", index,
					totalAggregationName, countAggregationName, averageAggregationName, minAggregationName, maxAggregationName)
			}
		}
	}

	return nil
}

func (am *AzureMonitor) checkResourceGroupTargetsValidation() error {
	for resourceGroupIndex, target := range am.ResourceGroupTargets {
		if target.ResourceGroup == "" {
			return fmt.Errorf(
				"resource group target #%d resource_group is empty or missing. Please check your configuration",
				resourceGroupIndex+1)
		}

		if len(target.Resources) == 0 {
			return fmt.Errorf("resource group target #%d has no resources. Please check your configuration", resourceGroupIndex+1)
		}

		for resourceIndex, resource := range target.Resources {
			if resource.ResourceType == "" {
				return fmt.Errorf(
					"resource group target #%d resource #%d resource_type is empty or missing. Please check your configuration",
					resourceGroupIndex+1, resourceIndex+1)
			}

			if len(resource.Aggregations) > 0 {
				if !areTargetAggregationsValid(resource.Aggregations) {
					return fmt.Errorf("resource group target #%d resource #%d aggregations contain invalid aggregation/s. "+
						"Please check your configuration. The valid aggregations are: %s, %s, %s, %s, %s", resourceGroupIndex, resourceIndex,
						totalAggregationName, countAggregationName, averageAggregationName, minAggregationName, maxAggregationName)
				}
			}
		}
	}

	return nil
}

func (am *AzureMonitor) checkSubscriptionTargetValidation() error {
	for index, target := range am.SubscriptionTargets {
		if target.ResourceType == "" {
			return fmt.Errorf(
				"subscription target #%d resource_type is empty or missing. Please check your configuration", index+1)
		}

		if len(target.Aggregations) > 0 {
			if !areTargetAggregationsValid(target.Aggregations) {
				return fmt.Errorf("subscription target #%d aggregations contain invalid aggregation/s. "+
					"Please check your configuration. The valid aggregations are: %s, %s, %s, %s, %s", index,
					totalAggregationName, countAggregationName, averageAggregationName, minAggregationName, maxAggregationName)
			}
		}
	}

	return nil
}

func (am *AzureMonitor) createResourceGroupTargetsFromSubscriptionTargets() error {
	if len(am.SubscriptionTargets) == 0 {
		am.Log.Debug("No subscription targets in configuration")
		return nil
	}

	am.Log.Debug("Creating resource group targets from subscription targets")

	apiURL := am.buildSubscriptionResourceGroupsAPIURL()
	body, err := am.getAPIResponseBody(apiURL)
	if err != nil {
		return fmt.Errorf("error getting subscription resource groups API response body: %v", err)
	}

	values, ok := body["value"].([]interface{})
	if !ok {
		return fmt.Errorf("subscription resource groups API response body bad format: value is missing")
	}

	for _, value := range values {
		resourceGroup, ok := value.(map[string]interface{})["name"].(string)
		if !ok {
			return fmt.Errorf("subscription resource groups API response body bad format: name of value is missing")
		}

		resourceGroupTarget := newResourceGroupTarget(resourceGroup, am.SubscriptionTargets)
		am.ResourceGroupTargets = append(am.ResourceGroupTargets, resourceGroupTarget)
	}

	am.Log.Debug("Total resource group targets: ", len(am.ResourceGroupTargets))
	return nil
}

func (am *AzureMonitor) createResourceTargetsFromResourceGroupTargets() error {
	if len(am.ResourceGroupTargets) == 0 {
		am.Log.Debug("No resource group targets in configuration")
		return nil
	}

	for _, target := range am.ResourceGroupTargets {
		am.Log.Debug("Creating resource targets from resource group target ", target.ResourceGroup)

		if err := am.createResourceTargetFromResourceGroupTarget(target); err != nil {
			return fmt.Errorf("error creating resource targets from resource group target %s: %v", target.ResourceGroup, err)
		}
	}

	return nil
}

func (am *AzureMonitor) createResourceTargetFromResourceGroupTarget(target *ResourceGroupTarget) error {
	apiURL := am.buildResourceGroupResourcesAPIURL(target.ResourceGroup)
	body, err := am.getAPIResponseBody(apiURL)
	if err != nil {
		return fmt.Errorf("error getting resource group resources API response body: %v", err)
	}

	values, ok := body["value"].([]interface{})
	if !ok {
		return fmt.Errorf("resource group resources API response body bad format: value is missing")
	}

	for _, value := range values {
		fullResourceID, ok := value.(map[string]interface{})["id"].(string)
		if !ok {
			return fmt.Errorf("resource group resources API response body bad format: id of value is missing")
		}

		ResourceIDParts := strings.Split(fullResourceID, "/")
		resourceID := strings.Join(ResourceIDParts[3:], "/")

		resourceType, ok := value.(map[string]interface{})["type"].(string)
		if !ok {
			return fmt.Errorf("resource group resources API response body bad format: type of value is missing")
		}

		resourcesWithResourceType := target.getResourcesWithResourceType(resourceType)
		if resourcesWithResourceType == nil {
			continue
		}

		for _, resourceWithResourceType := range resourcesWithResourceType {
			resourceTarget := newResourceTarget(resourceID, resourceWithResourceType.Metrics, resourceWithResourceType.Aggregations)
			am.ResourceTargets = append(am.ResourceTargets, resourceTarget)
		}
	}

	return nil
}

func (am *AzureMonitor) getResourceTargetMetricDefinitions(target *ResourceTarget) (map[string]interface{}, error) {
	apiURL := am.buildMetricDefinitionsAPIURL(target.ResourceID)
	body, err := am.getAPIResponseBody(apiURL)
	if err != nil {
		return nil, fmt.Errorf("error getting metric definitions API response body: %v", err)
	}

	return body, nil
}

func (am *AzureMonitor) checkResourceTargetsMetricsValidation() error {
	for _, target := range am.ResourceTargets {
		if len(target.Metrics) > 0 {
			body, err := am.getResourceTargetMetricDefinitions(target)
			if err != nil {
				return fmt.Errorf("error getting resource target %s metric definitions: %v", target.ResourceID, err)
			}

			if err = target.checkMetricsValidation(body); err != nil {
				return fmt.Errorf("error checking resource target %s metrics: %v", target.ResourceID, err)
			}
		}
	}

	return nil
}

func (am *AzureMonitor) setResourceTargetsMetrics() error {
	for _, target := range am.ResourceTargets {
		if len(target.Metrics) > 0 {
			continue
		}

		am.Log.Debug("Getting metrics for resource target ", target.ResourceID)

		body, err := am.getResourceTargetMetricDefinitions(target)
		if err != nil {
			return fmt.Errorf("error getting resource target %s metric definitions: %v", target.ResourceID, err)
		}

		if err = target.setMetrics(body); err != nil {
			return fmt.Errorf("error setting resource target %s metrics: %v", target.ResourceID, err)
		}
	}

	return nil
}

func (am *AzureMonitor) checkResourceTargetsMetricsMinTimeGrain() error {
	am.Log.Debug("Checking resource targets metrics min time grain")

	for _, target := range am.ResourceTargets {
		if err := am.checkResourceTargetMetricsMinTimeGrain(target); err != nil {
			return fmt.Errorf("error checking resource target %s metrics min time grain: %v", target.ResourceID, err)
		}
	}

	return nil
}

func (am *AzureMonitor) checkResourceTargetMetricsMinTimeGrain(target *ResourceTarget) error {
	body, err := am.getResourceTargetMetricDefinitions(target)
	if err != nil {
		return fmt.Errorf("error getting resource target %s metric definitions: %v", target.ResourceID, err)
	}

	values, err := getMetricDefinitionsValues(body)
	if err != nil {
		return err
	}

	timeGrainsMetricsMap, err := am.createResourceTargetTimeGrainsMetricsMap(target, values)
	if err != nil {
		return fmt.Errorf("error creating resource target time grains metrics map: %v", err)
	}

	if len(timeGrainsMetricsMap) == 1 {
		am.Log.Debug("All metrics of resource target ", target.ResourceID, " share the same min time grain")
		return nil
	}

	am.Log.Debug("Not all metrics of resource target ", target.ResourceID, " share the same min time grain")

	var firstTimeGrain string

	for timeGrain := range timeGrainsMetricsMap {
		firstTimeGrain = timeGrain
		break
	}

	for timeGrain, metrics := range timeGrainsMetricsMap {
		if timeGrain == firstTimeGrain {
			target.Metrics = metrics
			continue
		}

		newTargetAggregations := make([]string, 0)
		newTargetAggregations = append(newTargetAggregations, target.Aggregations...)
		am.ResourceTargets = append(am.ResourceTargets, newResourceTarget(target.ResourceID, metrics, newTargetAggregations))
	}

	return nil
}

func (am *AzureMonitor) createResourceTargetTimeGrainsMetricsMap(target *ResourceTarget, values []interface{}) (map[string][]string, error) {
	timeGrainsMetrics := make(map[string][]string, 0)

	for _, metric := range target.Metrics {
		for _, value := range values {
			metricName, err := getMetricDefinitionsMetricName(value.(map[string]interface{}))
			if err != nil {
				return nil, err
			}

			if metric == *metricName {
				metricMinTimeGrain, err := getMetricDefinitionsMetricMinTimeGrain(value.(map[string]interface{}))
				if err != nil {
					return nil, err
				}

				if _, found := timeGrainsMetrics[*metricMinTimeGrain]; !found {
					timeGrainsMetrics[*metricMinTimeGrain] = []string{*metricName}
				} else {
					timeGrainsMetrics[*metricMinTimeGrain] = append(timeGrainsMetrics[*metricMinTimeGrain], *metricName)
				}
			}
		}
	}

	return timeGrainsMetrics, nil
}

func (am *AzureMonitor) checkResourceTargetsMaxMetrics() {
	am.Log.Debug("Checking resource targets max metrics")

	for _, target := range am.ResourceTargets {
		if len(target.Metrics) <= maxMetricsPerRequest {
			am.Log.Debug("Resource target, ", target.ResourceID, " has less or equal to ", maxMetricsPerRequest, " (max) metrics")
			continue
		}

		am.Log.Debug("Resource target ", target.ResourceID, " has more than ", maxMetricsPerRequest, " (max) metrics")

		for start := maxMetricsPerRequest; start < len(target.Metrics); start += maxMetricsPerRequest {
			end := start + maxMetricsPerRequest

			if end > len(target.Metrics) {
				end = len(target.Metrics)
			}

			newTargetMetrics := target.Metrics[start:end]
			newTargetAggregations := make([]string, 0)
			newTargetAggregations = append(newTargetAggregations, target.Aggregations...)
			newTarget := newResourceTarget(target.ResourceID, newTargetMetrics, newTargetAggregations)
			am.ResourceTargets = append(am.ResourceTargets, newTarget)
		}

		target.Metrics = target.Metrics[:maxMetricsPerRequest]
	}
}

func (am *AzureMonitor) setResourceTargetsAggregations() {
	for _, target := range am.ResourceTargets {
		if len(target.Aggregations) == 0 {
			am.Log.Debug("Setting aggregations to resource target ", target.ResourceID)
			target.setAggregations()
		}
	}
}

func (rt *ResourceTarget) setMetrics(body map[string]interface{}) error {
	values, ok := body["value"].([]interface{})
	if !ok {
		return fmt.Errorf("metric definitions API response body bad format: value is missing")
	}

	for _, value := range values {
		metricName, err := getMetricDefinitionsMetricName(value.(map[string]interface{}))
		if err != nil {
			return err
		}

		rt.Metrics = append(rt.Metrics, *metricName)
	}

	return nil
}

func (rt *ResourceTarget) setAggregations() {
	rt.Aggregations = append(rt.Aggregations, totalAggregationName, countAggregationName, averageAggregationName,
		minAggregationName, maxAggregationName)
}

func (rt *ResourceTarget) checkMetricsValidation(metricDefinitionsBody map[string]interface{}) error {
	values, err := getMetricDefinitionsValues(metricDefinitionsBody)
	if err != nil {
		return err
	}

	for _, metric := range rt.Metrics {
		isMetricExist := false

		for _, value := range values {
			metricName, err := getMetricDefinitionsMetricName(value.(map[string]interface{}))
			if err != nil {
				return err
			}

			if metric == *metricName {
				isMetricExist = true
				break
			}
		}

		if !isMetricExist {
			return fmt.Errorf("resource target has invalid metric %s. Please check your resource targets, "+
				"resource group targets and subscription targets in your configuration", metric)
		}
	}

	return nil
}

func (rgt *ResourceGroupTarget) getResourcesWithResourceType(resourceType string) []*Resource {
	resourcesWithResourceType := make([]*Resource, 0)

	for _, resource := range rgt.Resources {
		if resource.ResourceType == resourceType {
			resourcesWithResourceType = append(resourcesWithResourceType, resource)
		}
	}

	if len(resourcesWithResourceType) == 0 {
		return nil
	}

	return resourcesWithResourceType
}

func areTargetAggregationsValid(aggregations []string) bool {
	for _, aggregation := range aggregations {
		if aggregation == totalAggregationName {
			continue
		} else if aggregation == countAggregationName {
			continue
		} else if aggregation == averageAggregationName {
			continue
		} else if aggregation == minAggregationName {
			continue
		} else if aggregation == maxAggregationName {
			continue
		} else {
			return false
		}
	}

	return true
}

func getMetricDefinitionsValues(body map[string]interface{}) ([]interface{}, error) {
	values, ok := body["value"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("metric definitions API response body bad format: value is missing")
	}

	return values, nil
}

func getMetricDefinitionsMetricName(value map[string]interface{}) (*string, error) {
	metricName, ok := value["name"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("metric definitions API response body bad format: name of value is missing")
	}

	metricNameValue, ok := metricName["value"].(string)
	if !ok {
		return nil, fmt.Errorf("metric definitions API response body bad format: value of name is missing")
	}

	return &metricNameValue, nil
}

func getMetricDefinitionsMetricMinTimeGrain(value map[string]interface{}) (*string, error) {
	metricAvailabilities, ok := value["metricAvailabilities"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("metric definitions API response body bad format: metricAvailabilities of value is missing")
	}

	if len(metricAvailabilities) == 0 {
		return nil, fmt.Errorf("metric definitions API response body bad format: metricAvailabilities of value is empty")
	}

	metricMinTimeGrain, ok := metricAvailabilities[0].(map[string]interface{})["timeGrain"].(string)
	if !ok {
		return nil, fmt.Errorf("metric definitions API response body bad format: timeGrain of metricAvailabilities is missing")
	}

	return &metricMinTimeGrain, nil
}
