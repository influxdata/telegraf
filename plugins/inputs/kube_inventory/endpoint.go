package kube_inventory

import (
	"context"
	"strings"
	"time"

	"github.com/ericchiang/k8s/apis/core/v1"

	"github.com/influxdata/telegraf"
)

func collectEndpoints(ctx context.Context, acc telegraf.Accumulator, ki *KubernetesInventory) {
	list, err := ki.client.getEndpoints(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, i := range list.Items {
		if err = ki.gatherEndpointWithHosts(*i, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ki *KubernetesInventory) gatherEndpoint(e v1.Endpoints, acc telegraf.Accumulator) error {
	if e.Metadata.CreationTimestamp.GetSeconds() == 0 && e.Metadata.CreationTimestamp.GetNanos() == 0 {
		return nil
	}

	fields := map[string]interface{}{
		"created":    time.Unix(e.Metadata.CreationTimestamp.GetSeconds(), int64(e.Metadata.CreationTimestamp.GetNanos())).UnixNano(),
		"generation": e.Metadata.GetGeneration(),
	}

	tags := map[string]string{
		"name":      e.Metadata.GetName(),
		"namespace": e.Metadata.GetNamespace(),
	}

	for _, endpoint := range e.GetSubsets() {
		for _, port := range endpoint.GetPorts() {
			fields["port"] = port.GetPort()

			tags["port_name"] = port.GetName()
			tags["port_protocol"] = port.GetProtocol()

			acc.AddFields(endpointMeasurement, fields, tags)
		}
	}

	return nil
}

// todo: do we want to add cardinality and collect hostnames/ready?
func (ki *KubernetesInventory) gatherEndpointWithHosts(e v1.Endpoints, acc telegraf.Accumulator) error {
	if e.Metadata.CreationTimestamp.GetSeconds() == 0 && e.Metadata.CreationTimestamp.GetNanos() == 0 {
		return nil
	}

	fields := map[string]interface{}{
		"created":    time.Unix(e.Metadata.CreationTimestamp.GetSeconds(), int64(e.Metadata.CreationTimestamp.GetNanos())).UnixNano(),
		"generation": e.Metadata.GetGeneration(),
	}

	tags := map[string]string{
		"name":      e.Metadata.GetName(),
		"namespace": e.Metadata.GetNamespace(),
	}

	for _, endpoint := range e.GetSubsets() {
		for _, readyAddr := range endpoint.GetAddresses() {
			fields["ready"] = true

			// todo: do we want tags["hostname"] = readyAddr.GetHostname()
			tags["ip"] = readyAddr.GetIp()
			tags["node_name"] = readyAddr.GetNodeName()
			if readyAddr.TargetRef != nil {
				tags[strings.ToLower(readyAddr.GetTargetRef().GetKind())] = readyAddr.GetTargetRef().GetName()
			}

			for _, port := range endpoint.GetPorts() {
				fields["port"] = port.GetPort()

				tags["port_name"] = port.GetName()
				tags["port_protocol"] = port.GetProtocol()

				acc.AddFields(endpointMeasurement, fields, tags)
			}
		}
		for _, notReadyAddr := range endpoint.GetNotReadyAddresses() {
			fields["ready"] = false

			// todo: do we want tags["hostname"] = readyAddr.GetHostname()
			tags["ip"] = notReadyAddr.GetIp()
			tags["node_name"] = notReadyAddr.GetNodeName()
			if notReadyAddr.TargetRef != nil {
				tags[strings.ToLower(notReadyAddr.GetTargetRef().GetKind())] = notReadyAddr.GetTargetRef().GetName()
			}

			for _, port := range endpoint.GetPorts() {
				fields["port"] = port.GetPort()

				tags["port_name"] = port.GetName()
				tags["port_protocol"] = port.GetProtocol()

				acc.AddFields(endpointMeasurement, fields, tags)
			}
		}
	}

	return nil
}
