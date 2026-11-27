package kube_inventory

import (
	"context"
	"maps"
	"strings"

	discoveryv1 "k8s.io/api/discovery/v1"

	"github.com/influxdata/telegraf"
)

func collectEndpoints(ctx context.Context, acc telegraf.Accumulator, ki *KubernetesInventory) {
	list, err := ki.client.getEndpoints(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, i := range list.Items {
		gatherEndpoint(i, acc)
	}
}

func gatherEndpoint(e discoveryv1.EndpointSlice, acc telegraf.Accumulator) {
	creationTS := e.GetCreationTimestamp()
	if creationTS.IsZero() {
		return
	}

	sliceFields := map[string]any{
		"created":    e.GetCreationTimestamp().UnixNano(),
		"generation": e.Generation,
	}

	sliceTags := map[string]string{
		"endpoint_name": e.Name,
		"namespace":     e.Namespace,
	}

	for _, endpoint := range e.Endpoints {
		// Each endpoint sets only the optional values it has, so start from
		// the slice-wide values instead of carrying over the previous endpoint.
		endpointFields := maps.Clone(sliceFields)
		endpointTags := maps.Clone(sliceTags)

		if endpoint.Conditions.Ready == nil {
			endpointFields["ready"] = true
		} else {
			endpointFields["ready"] = *endpoint.Conditions.Ready
		}

		if endpoint.Hostname != nil {
			endpointTags["hostname"] = *endpoint.Hostname
		}
		if endpoint.NodeName != nil {
			endpointTags["node_name"] = *endpoint.NodeName
		}
		if endpoint.TargetRef != nil {
			endpointTags[strings.ToLower(endpoint.TargetRef.Kind)] = endpoint.TargetRef.Name
		}

		for _, port := range e.Ports {
			fields := maps.Clone(endpointFields)
			tags := maps.Clone(endpointTags)

			if port.Port != nil {
				fields["port"] = *port.Port
			}

			tags["port_name"] = *port.Name
			tags["port_protocol"] = string(*port.Protocol)

			acc.AddFields(endpointMeasurement, fields, tags)
		}
	}
}
