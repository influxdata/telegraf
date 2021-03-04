package kube_inventory

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/influxdata/telegraf"
)

func collectServices(ctx context.Context, acc telegraf.Accumulator, ki *KubernetesInventory) {
	list, err := ki.client.getServices(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, i := range list.Items {
		if err = ki.gatherService(i, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ki *KubernetesInventory) gatherService(s corev1.Service, acc telegraf.Accumulator) error {
	if s.GetCreationTimestamp().Second() == 0 && s.GetCreationTimestamp().Nanosecond() == 0 {
		return nil
	}

	fields := map[string]interface{}{
		"created":    time.Unix(int64(s.GetCreationTimestamp().Second()), int64(s.GetCreationTimestamp().Nanosecond())).UnixNano(),
		"generation": s.Generation,
	}

	tags := map[string]string{
		"service_name": s.Name,
		"namespace":    s.Namespace,
	}

	for key, val := range s.Spec.Selector {
		if ki.selectorFilter.Match(key) {
			tags["selector_"+key] = val
		}
	}

	var getPorts = func() {
		for _, port := range s.Spec.Ports {
			fields["port"] = port.Port
			fields["target_port"] = port.TargetPort.IntVal

			tags["port_name"] = port.Name
			tags["port_protocol"] = string(port.Protocol)

			if s.Spec.Type == "ExternalName" {
				tags["external_name"] = s.Spec.ExternalName
			} else {
				tags["cluster_ip"] = s.Spec.ClusterIP
			}

			acc.AddFields(serviceMeasurement, fields, tags)
		}
	}

	if externIPs := s.Spec.ExternalIPs; externIPs != nil {
		for _, ip := range externIPs {
			tags["ip"] = ip

			getPorts()
		}
	} else {
		getPorts()
	}

	return nil
}
