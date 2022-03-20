package azure_monitor

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

const (
	maxMetricsPerRequest = 20
)

func newResourceTarget(resourceID string, metrics []string, aggregations []string) *ResourceTarget {
	return &ResourceTarget{
		ResourceID:   resourceID,
		Metrics:      metrics,
		Aggregations: aggregations,
	}
}

func newAzureResourcesClient(subscriptionID string, credential *azidentity.ClientSecretCredential) *azureResourcesClient {
	return &azureResourcesClient{
		client: armresources.NewClient(subscriptionID, credential, nil),
	}
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

	return am.checkSubscriptionTargetValidation()
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
					"Please check your configuration. The valid aggregations are: %s", index, strings.Join(getPossibleAggregations(), ", "))
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
						"Please check your configuration. The valid aggregations are: %s", resourceGroupIndex, resourceIndex,
						strings.Join(getPossibleAggregations(), ", "))
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
					"Please check your configuration. The valid aggregations are: %s", index, strings.Join(getPossibleAggregations(), ", "))
			}
		}
	}

	return nil
}

func (am *AzureMonitor) addPrefixToResourceTargetsResourceID() {
	for _, target := range am.ResourceTargets {
		target.ResourceID = "/subscriptions/" + am.SubscriptionID + "/" + target.ResourceID
	}
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
	resourceTargetsCreatedNum := 0
	filter := createClientResourcesFilter(target.Resources)
	responses, err := am.azureClients.resourcesClient.ListByResourceGroup(am.azureClients.ctx, target.ResourceGroup,
		&armresources.ClientListByResourceGroupOptions{Filter: &filter})
	if err != nil {
		return err
	}

	for _, response := range responses {
		currentResourceTargetsCreatedNum, err := am.createResourceTargetFromTargetResources(response.Value, target.Resources)
		if err != nil {
			return fmt.Errorf("error creating resource target from resource group target resources: %v", err)
		}

		resourceTargetsCreatedNum += currentResourceTargetsCreatedNum
	}

	am.Log.Debug("Total resource targets created from resource group target ", target.ResourceGroup, ": ", resourceTargetsCreatedNum)
	return nil
}

func (am *AzureMonitor) createResourceTargetsFromSubscriptionTargets() error {
	if len(am.SubscriptionTargets) == 0 {
		am.Log.Debug("No subscription targets in configuration")
		return nil
	}

	am.Log.Debug("Creating resource targets from subscription targets")

	resourceTargetsCreatedNum := 0
	filter := createClientResourcesFilter(am.SubscriptionTargets)
	responses, err := am.azureClients.resourcesClient.List(am.azureClients.ctx, &armresources.ClientListOptions{Filter: &filter})
	if err != nil {
		return err
	}

	for _, response := range responses {
		currentResourceTargetsCreatedNum, err := am.createResourceTargetFromTargetResources(response.Value, am.SubscriptionTargets)
		if err != nil {
			return fmt.Errorf("error creating resource target from subscription targets: %v", err)
		}

		resourceTargetsCreatedNum += currentResourceTargetsCreatedNum
	}

	am.Log.Debug("Total resource targets created from subscription targets: ", resourceTargetsCreatedNum)
	return nil
}

func (am *AzureMonitor) createResourceTargetFromTargetResources(resources []*armresources.GenericResourceExpanded, targetResources []*Resource) (int, error) {
	resourceTargetsCreatedNum := 0

	for _, targetResource := range targetResources {
		isResourceTargetCreated := false

		for _, resource := range resources {
			resourceID, err := getResourcesClientResourceID(resource)
			if err != nil {
				return resourceTargetsCreatedNum, err
			}

			resourceType, err := getResourcesClientResourceType(resource)
			if err != nil {
				return resourceTargetsCreatedNum, err
			}

			if *resourceType != targetResource.ResourceType {
				continue
			}

			am.ResourceTargets = append(am.ResourceTargets, newResourceTarget(*resourceID, targetResource.Metrics, targetResource.Aggregations))
			isResourceTargetCreated = true
			resourceTargetsCreatedNum++
		}

		if !isResourceTargetCreated {
			return resourceTargetsCreatedNum, fmt.Errorf("could not find resources with resource type %s", targetResource.ResourceType)
		}
	}

	return resourceTargetsCreatedNum, nil
}

func (am *AzureMonitor) checkResourceTargetsMetricsValidation() error {
	for _, target := range am.ResourceTargets {
		if len(target.Metrics) > 0 {
			response, err := am.getMetricDefinitionsResponse(target.ResourceID)
			if err != nil {
				return fmt.Errorf("error getting metric definitions response for resource target %s: %v", target.ResourceID, err)
			}

			if err = target.checkMetricsValidation(response.Value); err != nil {
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

		am.Log.Debug("Setting metrics for resource target ", target.ResourceID)

		response, err := am.getMetricDefinitionsResponse(target.ResourceID)
		if err != nil {
			return fmt.Errorf("error getting metric definitions response for resource target %s: %v", target.ResourceID, err)
		}

		if err = target.setMetrics(response.Value); err != nil {
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
	response, err := am.getMetricDefinitionsResponse(target.ResourceID)
	if err != nil {
		return fmt.Errorf("error getting metric definitions response for resource target %s: %v", target.ResourceID, err)
	}

	timeGrainsMetricsMap, err := target.createResourceTargetTimeGrainsMetricsMap(response.Value)
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

func (am *AzureMonitor) getMetricDefinitionsResponse(resourceID string) (*armmonitor.MetricDefinitionsClientListResponse, error) {
	response, err := am.azureClients.metricDefinitionsClient.List(am.azureClients.ctx, resourceID, nil)
	if err != nil {
		return nil, fmt.Errorf("error listing metric definitions for the resource target %s: %v", resourceID, err)
	}

	if len(response.Value) == 0 {
		return nil, fmt.Errorf("metric definitions response is bad formatted: Value is empty")
	}

	return &response, nil
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

func (am *AzureMonitor) changeResourceTargetsMetricsWithComma() {
	for _, target := range am.ResourceTargets {
		target.changeMetricsWithComma()
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

func (arc *azureResourcesClient) List(ctx context.Context, options *armresources.ClientListOptions) ([]*armresources.ClientListResponse, error) {
	responses := make([]*armresources.ClientListResponse, 0)
	pager := arc.client.List(options)

	for pager.NextPage(ctx) {
		response := pager.PageResponse()
		responses = append(responses, &response)
	}

	if err := pager.Err(); err != nil {
		return nil, err
	}

	return responses, nil
}

func (arc *azureResourcesClient) ListByResourceGroup(
	ctx context.Context,
	resourceGroup string,
	options *armresources.ClientListByResourceGroupOptions,
) ([]*armresources.ClientListByResourceGroupResponse, error) {
	responses := make([]*armresources.ClientListByResourceGroupResponse, 0)
	pager := arc.client.ListByResourceGroup(resourceGroup, options)

	for pager.NextPage(ctx) {
		response := pager.PageResponse()
		responses = append(responses, &response)
	}

	if err := pager.Err(); err != nil {
		return nil, err
	}

	return responses, nil
}

func (rt *ResourceTarget) setMetrics(metricDefinitions []*armmonitor.MetricDefinition) error {
	for _, metricDefinition := range metricDefinitions {
		metricNameValue, err := getMetricDefinitionsClientMetricNameValue(metricDefinition)
		if err != nil {
			return err
		}

		rt.Metrics = append(rt.Metrics, *metricNameValue)
	}

	return nil
}

func (rt *ResourceTarget) setAggregations() {
	rt.Aggregations = append(rt.Aggregations, getPossibleAggregations()...)
}

func (rt *ResourceTarget) checkMetricsValidation(metricDefinitions []*armmonitor.MetricDefinition) error {
	for _, metric := range rt.Metrics {
		isMetricExist := false

		for _, metricDefinition := range metricDefinitions {
			metricNameValue, err := getMetricDefinitionsClientMetricNameValue(metricDefinition)
			if err != nil {
				return err
			}

			if metric == *metricNameValue {
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

func (rt *ResourceTarget) createResourceTargetTimeGrainsMetricsMap(metricDefinitions []*armmonitor.MetricDefinition) (map[string][]string, error) {
	timeGrainsMetrics := make(map[string][]string)

	for _, metric := range rt.Metrics {
		for _, metricDefinition := range metricDefinitions {
			metricNameValue, err := getMetricDefinitionsClientMetricNameValue(metricDefinition)
			if err != nil {
				return nil, err
			}

			if metric == *metricNameValue {
				metricMinTimeGrain, err := getMetricDefinitionsMetricMinTimeGrain(metricDefinition)
				if err != nil {
					return nil, err
				}

				if _, found := timeGrainsMetrics[*metricMinTimeGrain]; !found {
					timeGrainsMetrics[*metricMinTimeGrain] = []string{metric}
				} else {
					timeGrainsMetrics[*metricMinTimeGrain] = append(timeGrainsMetrics[*metricMinTimeGrain], metric)
				}
			}
		}
	}

	return timeGrainsMetrics, nil
}

func (rt *ResourceTarget) changeMetricsWithComma() {
	for index := 0; index < len(rt.Metrics); index++ {
		rt.Metrics[index] = strings.Replace(rt.Metrics[index], ",", "%2", -1)
	}
}

func getPossibleAggregations() []string {
	possibleAggregations := make([]string, 0)

	for _, aggregation := range armmonitor.PossibleAggregationTypeEnumValues() {
		possibleAggregations = append(possibleAggregations, string(aggregation))
	}

	return possibleAggregations
}

func areTargetAggregationsValid(targetAggregations []string) bool {
	for _, targetAggregation := range targetAggregations {
		isTargetAggregationValid := false

		for _, aggregation := range getPossibleAggregations() {
			if targetAggregation == aggregation {
				isTargetAggregationValid = true
				break
			}
		}

		if !isTargetAggregationValid {
			return false
		}
	}

	return true
}

func createClientResourcesFilter(resources []*Resource) string {
	var filter string
	resourcesSize := len(resources)

	for index, resource := range resources {
		if index+1 == resourcesSize {
			filter += "resourceType eq " + "'" + resource.ResourceType + "'"
		} else {
			filter += "resourceType eq " + "'" + resource.ResourceType + "'" + " or "
		}
	}

	return filter
}

func getResourcesClientResourceID(resource *armresources.GenericResourceExpanded) (*string, error) {
	if resource == nil {
		return nil, fmt.Errorf("resources client response is bad formatted: resource is missing")
	}

	if resource.ID == nil {
		return nil, fmt.Errorf("resources client response is bad formatted: resource ID is missing")
	}

	return resource.ID, nil
}

func getResourcesClientResourceType(resource *armresources.GenericResourceExpanded) (*string, error) {
	if resource == nil {
		return nil, fmt.Errorf("resources client response is bad formatted: resource is missing")
	}

	if resource.Type == nil {
		return nil, fmt.Errorf("resources client response is bad formatted: resource Type is missing")
	}

	return resource.Type, nil
}

func getMetricDefinitionsClientMetricNameValue(metricDefinition *armmonitor.MetricDefinition) (*string, error) {
	if metricDefinition == nil {
		return nil, fmt.Errorf("metric definitions client response is bad formatted: metric definition is missing")
	}

	metricName := metricDefinition.Name
	if metricName == nil {
		return nil, fmt.Errorf("metric definitions client response is bad formatted: metric definition Name is missing")
	}

	metricNameValue := metricName.Value
	if metricNameValue == nil {
		return nil, fmt.Errorf("metric definitions client response is bad formatted: metric definition Name.Value is missing")
	}

	return metricNameValue, nil
}

func getMetricDefinitionsMetricMinTimeGrain(metricDefinition *armmonitor.MetricDefinition) (*string, error) {
	if metricDefinition == nil {
		return nil, fmt.Errorf("metric definitions client response is bad formatted: metric definition is missing")
	}

	if len(metricDefinition.MetricAvailabilities) == 0 {
		return nil, fmt.Errorf("metric definitions client response is bad formatted: metric definition MetricAvailabilities is empty")
	}

	metricAvailability := metricDefinition.MetricAvailabilities[0]
	if metricDefinition.MetricAvailabilities[0] == nil {
		return nil, fmt.Errorf("metric definitions client response is bad formatted: metric definition MetricAvailabilities[0] is missing")
	}

	timeGrain := metricAvailability.TimeGrain
	if timeGrain == nil {
		return nil, fmt.Errorf("metric definitions client response is bad formatted: metric definition MetricAvailabilities[0].TimeGrain is missing")
	}

	return timeGrain, nil
}
