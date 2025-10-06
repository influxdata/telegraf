package kube_inventory

import (
	"context"
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

	fields := map[string]interface{}{
		"created":    e.GetCreationTimestamp().UnixNano(),
		"generation": e.Generation,
	}

	tags := map[string]string{
		"endpoint_name": e.Name,
		"namespace":     e.Namespace,
	}

	for _, endpoint := range e.Endpoints {
		if endpoint.Conditions.Ready == nil {
			fields["ready"] = true
		} else {
			fields["ready"] = *endpoint.Conditions.Ready
		}

		if endpoint.Hostname != nil {
			tags["hostname"] = *endpoint.Hostname
		}
		if endpoint.NodeName != nil {
			tags["node_name"] = *endpoint.NodeName
		}
		if endpoint.TargetRef != nil {
			tags[strings.ToLower(endpoint.TargetRef.Kind)] = endpoint.TargetRef.Name
		}

		for _, port := range e.Ports {
			if port.Port != nil {
				fields["port"] = *port.Port
			}

			tags["port_name"] = *port.Name
			tags["port_protocol"] = string(*port.Protocol)

			acc.AddFields(endpointMeasurement, fields, tags)
		}
	}
}
