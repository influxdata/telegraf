package kube_inventory

import (
	"context"
	"time"

	"github.com/ericchiang/k8s/apis/core/v1"

	"github.com/influxdata/telegraf"
)

func collectServices(ctx context.Context, acc telegraf.Accumulator, ki *KubernetesInventory) {
	list, err := ki.client.getServices(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, i := range list.Items {
		if err = ki.gatherService(*i, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ki *KubernetesInventory) gatherService(s v1.Service, acc telegraf.Accumulator) error {
	if s.Metadata.CreationTimestamp.GetSeconds() == 0 && s.Metadata.CreationTimestamp.GetNanos() == 0 {
		return nil
	}

	fields := map[string]interface{}{
		"created":    time.Unix(s.Metadata.CreationTimestamp.GetSeconds(), int64(s.Metadata.CreationTimestamp.GetNanos())).UnixNano(),
		"generation": s.Metadata.GetGeneration(),
	}

	tags := map[string]string{
		"service_name": s.Metadata.GetName(),
		"namespace":    s.Metadata.GetNamespace(),
	}

	var getPorts = func() {
		for _, port := range s.GetSpec().GetPorts() {
			fields["port"] = port.GetPort()
			fields["target_port"] = port.GetTargetPort().GetIntVal()

			tags["port_name"] = port.GetName()
			tags["port_protocol"] = port.GetProtocol()

			if s.GetSpec().GetType() == "ExternalName" {
				tags["external_name"] = s.GetSpec().GetExternalName()
			} else {
				tags["cluster_ip"] = s.GetSpec().GetClusterIP()
			}

			acc.AddFields(serviceMeasurement, fields, tags)
		}
	}

	if externIPs := s.GetSpec().GetExternalIPs(); externIPs != nil {
		for _, ip := range externIPs {
			tags["ip"] = ip

			getPorts()
		}
	} else {
		getPorts()
	}

	return nil
}
