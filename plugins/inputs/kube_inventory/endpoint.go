package kube_inventory

import (
	"context"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	corev1 "k8s.io/api/core/v1"
)

func collectEndpoints(ctx context.Context, acc telegraf.Accumulator, ki *KubernetesInventory) {
	list, err := ki.client.getEndpoints(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, i := range list.Items {
		if err = ki.gatherEndpoint(i, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ki *KubernetesInventory) gatherEndpoint(e corev1.Endpoints, acc telegraf.Accumulator) error {
	if e.GetCreationTimestamp().Second() == 0 && e.GetCreationTimestamp().Nanosecond() == 0 {
		return nil
	}

	fields := map[string]interface{}{
		"created":    time.Unix(int64(e.GetCreationTimestamp().Second()), int64(e.GetCreationTimestamp().Nanosecond())).UnixNano(),
		"generation": e.Generation,
	}

	tags := map[string]string{
		"endpoint_name": e.Name,
		"namespace":     e.Namespace,
	}

	for _, endpoint := range e.Subsets {
		for _, readyAddr := range endpoint.Addresses {
			fields["ready"] = true

			tags["hostname"] = readyAddr.Hostname
			tags["node_name"] = *readyAddr.NodeName
			if readyAddr.TargetRef != nil {
				tags[strings.ToLower(readyAddr.TargetRef.Kind)] = readyAddr.TargetRef.Name
			}

			for _, port := range endpoint.Ports {
				fields["port"] = port.Port

				tags["port_name"] = port.Name
				tags["port_protocol"] = string(port.Protocol)

				acc.AddFields(endpointMeasurement, fields, tags)
			}
		}
		for _, notReadyAddr := range endpoint.NotReadyAddresses {
			fields["ready"] = false

			tags["hostname"] = notReadyAddr.Hostname
			tags["node_name"] = *notReadyAddr.NodeName
			if notReadyAddr.TargetRef != nil {
				tags[strings.ToLower(notReadyAddr.TargetRef.Kind)] = notReadyAddr.TargetRef.Name
			}

			for _, port := range endpoint.Ports {
				fields["port"] = port.Port

				tags["port_name"] = port.Name
				tags["port_protocol"] = string(port.Protocol)

				acc.AddFields(endpointMeasurement, fields, tags)
			}
		}
	}

	return nil
}
