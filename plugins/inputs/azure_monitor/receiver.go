package azure_monitor

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"

	"github.com/influxdata/telegraf"
)

var nameReplacer = strings.NewReplacer(".", "_", "/", "_", " ", "_", "(", "_", ")", "_")

type metricReceiver struct {
	client         client
	subscriptionID string
	resources      []*resourceTarget
}

func newReceiver(
	ctx context.Context,
	client client,
	subscriptionID string,
	resources []*resourceTarget,
	groups []*resourceGroupTarget,
	subscriptions []*resource,
) (*metricReceiver, error) {
	r := &metricReceiver{
		client:         client,
		subscriptionID: subscriptionID,
	}

	// Create all resource targets from the different configurations
	r.resources = slices.Clone(resources)

	for _, target := range groups {
		if err := r.createResourceTargetFromResourceGroupTarget(ctx, target); err != nil {
			return nil, fmt.Errorf("creating resource targets from resource group target %q failed: %w", target.ResourceGroup, err)
		}
	}

	if err := r.createResourceTargetsFromSubscriptionTargets(ctx, subscriptions); err != nil {
		return nil, fmt.Errorf("error creating resource targets from subscription targets: %w", err)
	}

	// Get valid aggregation types
	pagt := armmonitor.PossibleAggregationTypeEnumValues()
	validAggregationTypes := make([]string, 0, len(pagt))
	for _, v := range pagt {
		validAggregationTypes = append(validAggregationTypes, string(v))
	}

	// Validate the metrics of the final result
	result := make([]*resourceTarget, 0, len(r.resources))
	for _, target := range r.resources {
		// Get all valid metric definitions for the given resource
		response, err := r.client.MetricDefinitionsList(ctx, target.ResourceID, nil)
		if err != nil {
			return nil, fmt.Errorf("listing metric definitions for resource target %q failed: %w", target.ResourceID, err)
		}
		if len(response.Value) == 0 {
			return nil, errors.New("no value for metric definitions response")
		}

		// Validate the received metric definitions
		for _, definition := range response.Value {
			if definition == nil {
				return nil, fmt.Errorf("bad metric definition %v for resource target %q", definition, target.ResourceID)
			}

			if definition.Name == nil || definition.Name.Value == nil {
				return nil, fmt.Errorf("bad name-value in metric definition %v for resource target %q", definition, target.ResourceID)
			}

			if len(definition.MetricAvailabilities) == 0 || definition.MetricAvailabilities[0] == nil {
				return nil, fmt.Errorf("bad availability in metric definition %v for resource target %q", definition, target.ResourceID)
			}

			if definition.MetricAvailabilities[0].TimeGrain == nil {
				return nil, fmt.Errorf("bad time grain in metric definition %v for resource target %q", definition, target.ResourceID)
			}
		}

		// Validate the target aggregations or set defaults
		if len(target.Aggregations) == 0 {
			target.Aggregations = validAggregationTypes
		} else {
			for _, a := range target.Aggregations {
				if !slices.Contains(validAggregationTypes, a) {
					return nil, fmt.Errorf("invalid aggregation %q in resource target %q", a, target.ResourceID)
				}
			}
		}

		if len(target.Metrics) > 0 {
			// Validate metrics assigned to the resource target if any
			for _, metric := range target.Metrics {
				exists := slices.ContainsFunc(response.Value, func(definition *armmonitor.MetricDefinition) bool {
					return metric == *definition.Name.Value
				})

				if !exists {
					return nil, fmt.Errorf("invalid metric %q for resource target %q", metric, target.ResourceID)
				}
			}
		} else {
			// Create metrics for all definitions
			for _, definition := range response.Value {
				target.Metrics = append(target.Metrics, *definition.Name.Value)
			}
		}

		timeGrains := make(map[string][]string)
		for _, metric := range target.Metrics {
			for _, definition := range response.Value {
				if metric != *definition.Name.Value {
					continue
				}

				timeGrain := *definition.MetricAvailabilities[0].TimeGrain
				if _, found := timeGrains[timeGrain]; !found {
					timeGrains[timeGrain] = make([]string, 0, 1)
				}
				timeGrains[timeGrain] = append(timeGrains[timeGrain], metric)
			}
		}

		// Group the resource targets by minimum time gain and split large batches
		for _, metrics := range timeGrains {
			// Split the metrics into batches which do not exceed the maximum size
			for start := 0; start < len(metrics); start += maxMetricsBatch {
				end := min(start+maxMetricsBatch, len(metrics))

				result = append(result, &resourceTarget{
					ResourceID:   target.ResourceID,
					Metrics:      metrics[start:end],
					Aggregations: target.Aggregations,
				})
			}
		}
	}
	r.resources = result

	return r, nil
}

func (r *metricReceiver) collectMetrics(ctx context.Context, acc telegraf.Accumulator, target *resourceTarget, log telegraf.Logger) {
	names := strings.Join(target.Metrics, ",")
	aggregations := strings.Join(target.Aggregations, ",")

	response, err := r.client.MetricsList(ctx, target.ResourceID, &armmonitor.MetricsClientListOptions{
		Metricnames: &names,
		Aggregation: &aggregations,
	})
	if err != nil {
		acc.AddError(fmt.Errorf("listing metrics for resource target %q failed: %w", target.ResourceID, err))
		return
	}

	// Check the reponse
	if response.Namespace == nil {
		acc.AddError(fmt.Errorf("bad namespace in response for resource target %q", target.ResourceID))
		return
	}
	if response.Resourceregion == nil {
		acc.AddError(fmt.Errorf("bad resource region in response for resource target %q", target.ResourceID))
		return
	}

	for _, metric := range response.Value {
		// Check the returned metric
		if metric == nil {
			acc.AddError(fmt.Errorf("bad metric for resource target %q", target.ResourceID))
			continue
		}
		if metric.ErrorCode != nil && *metric.ErrorCode != "Success" {
			var err error
			if metric.ErrorMessage != nil {
				err = fmt.Errorf("metric error for resource target %q: %s: %s", target.ResourceID, *metric.ErrorCode, *metric.ErrorMessage)
			} else {
				err = fmt.Errorf("metric error for resource target %q: %s", target.ResourceID, *metric.ErrorCode)
			}
			acc.AddError(err)
			continue
		}
		if metric.ID == nil {
			acc.AddError(fmt.Errorf("missing metric ID for resource target %q", target.ResourceID))
			continue
		}

		if metric.Name == nil || metric.Name.LocalizedValue == nil {
			acc.AddError(fmt.Errorf("bad metric name for metric %q of resource target %q", *metric.ID, target.ResourceID))
			continue
		}

		if metric.Unit == nil {
			acc.AddError(fmt.Errorf("missing unit for metric %q of resource target %q", *metric.ID, target.ResourceID))
			continue
		}

		if len(metric.Timeseries) == 0 || metric.Timeseries[0] == nil || len(metric.Timeseries[0].Data) == 0 {
			log.Debugf("no timeseries data for metric %q of resource target %q", *metric.ID, target.ResourceID)
			continue
		}
		// This is from https://github.com/logzio/azure-monitor-metrics-receiver/blob/master/collector.go but why do we
		// only get the first index?
		timeseries := metric.Timeseries[0].Data

		// Construct the metric name
		name := "azure_monitor_" + *response.Namespace + "_" + *metric.Name.LocalizedValue
		name = nameReplacer.Replace(strings.ToLower(name))

		// Construct the metric tags
		idSplit := strings.Split(*metric.ID, "/providers/")
		parts := make([][]string, 0, len(idSplit))
		for _, s := range idSplit {
			parts = append(parts, strings.Split(s, "/"))
		}
		if len(parts) < 2 {
			acc.AddError(fmt.Errorf("not enough top parts for metric %q of resource target %q", *metric.ID, target.ResourceID))
			continue
		}
		if len(parts[0]) < 5 {
			err := fmt.Errorf("not enough parts (%d/5) for section 0 for metric %q of resource target %q", len(parts[0]), *metric.ID, target.ResourceID)
			acc.AddError(err)
			continue
		}
		if len(parts[1]) < 3 {
			err := fmt.Errorf("not enough parts (%d/3) for section 1 for metric %q of resource target %q", len(parts[1]), *metric.ID, target.ResourceID)
			acc.AddError(err)
			continue
		}

		tags := map[string]string{
			"subscription_id": parts[0][2],
			"resource_group":  parts[0][4],
			"resource_name":   strings.Join(parts[1][2:], "/"),
			"namespace":       *response.Namespace,
			"resource_region": *response.Resourceregion,
			"unit":            string(*metric.Unit),
		}

		// Construct the metric fields. Iterate the timeseries in reverse order
		// and only accept the latest valid field set
		for i := len(timeseries) - 1; i >= 0; i-- {
			data := timeseries[i]
			if data == nil {
				acc.AddError(fmt.Errorf("bad timeseries data for metric %q of resource target %q", *metric.ID, target.ResourceID))
				continue
			}

			if data.TimeStamp == nil {
				acc.AddError(fmt.Errorf("bad timestamp for data in metric %q of resource target %q", *metric.ID, target.ResourceID))
				continue
			}

			fields := make(map[string]interface{}, 6)
			fields["timeStamp"] = data.TimeStamp.Format("2006-01-02T15:04:05Z07:00")
			if data.Total != nil {
				fields["total"] = *data.Total
			}
			if data.Average != nil {
				fields["average"] = *data.Average
			}
			if data.Count != nil {
				fields["count"] = *data.Count
			}
			if data.Minimum != nil {
				fields["minimum"] = *data.Minimum
			}
			if data.Maximum != nil {
				fields["maximum"] = *data.Maximum
			}

			// Add metric if we do have at least the timestamp and another field
			// Exit on success to only keep the latest metric
			if len(fields) > 1 {
				acc.AddFields(name, fields, tags)
				break
			}
		}
	}
}

func createClientResourcesFilter(resources []*resource) string {
	filter := make([]string, 0, len(resources))
	for _, r := range resources {
		filter = append(filter, "resourceType eq "+"'"+r.ResourceType+"'")
	}
	return strings.Join(filter, " or ")
}

func (r *metricReceiver) createResourceTargetFromResourceGroupTarget(ctx context.Context, target *resourceGroupTarget) error {
	filter := createClientResourcesFilter(target.Resources)
	option := &armresources.ClientListByResourceGroupOptions{Filter: &filter}
	responses, err := r.client.ResourcesListByResourceGroup(ctx, target.ResourceGroup, option)
	if err != nil {
		return fmt.Errorf("listing by resource group failed: %w", err)
	}

	for _, response := range responses {
		if err := r.createResourceTargetFromTargetResources(response.Value, target.Resources); err != nil {
			return fmt.Errorf("error creating resource target from resource group target resources: %w", err)
		}
	}

	return nil
}

func (r *metricReceiver) createResourceTargetsFromSubscriptionTargets(ctx context.Context, targets []*resource) error {
	if len(targets) == 0 {
		return nil
	}

	filter := createClientResourcesFilter(targets)
	responses, err := r.client.ResourcesList(ctx, &armresources.ClientListOptions{Filter: &filter})
	if err != nil {
		return fmt.Errorf("listing resources failed: %w", err)
	}

	for _, response := range responses {
		if err := r.createResourceTargetFromTargetResources(response.Value, targets); err != nil {
			return fmt.Errorf("error creating resource target from subscription targets: %w", err)
		}
	}

	return nil
}

func (r *metricReceiver) createResourceTargetFromTargetResources(resources []*armresources.GenericResourceExpanded, targetResources []*resource) error {
	for _, targetResource := range targetResources {
		var isResourceTargetCreated bool

		for _, resource := range resources {
			if resource == nil {
				return errors.New("invalid resource")
			}
			if resource.ID == nil {
				return errors.New("invalid resource ID")
			}
			if resource.Type == nil {
				return errors.New("invalid resource type")
			}
			if *resource.Type != targetResource.ResourceType {
				continue
			}

			r.resources = append(r.resources, &resourceTarget{*resource.ID, targetResource.Metrics, targetResource.Aggregations})
			isResourceTargetCreated = true
		}

		if !isResourceTargetCreated {
			return fmt.Errorf("could not find resources with resource type %q", targetResource.ResourceType)
		}
	}

	return nil
}
