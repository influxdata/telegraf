package kube_inventory

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"github.com/influxdata/telegraf"
)

func collectResourceQuotas(ctx context.Context, acc telegraf.Accumulator, ki *KubernetesInventory) {
	list, err := ki.client.getResourceQuotas(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, i := range list.Items {
		ki.gatherResourceQuota(i, acc)
	}
}

func (ki *KubernetesInventory) gatherResourceQuota(r corev1.ResourceQuota, acc telegraf.Accumulator) {
	fields := map[string]interface{}{}
	tags := map[string]string{
		"resource":  r.Name,
		"namespace": r.Namespace,
	}

	for resourceName, val := range r.Status.Hard {
		switch resourceName {
		case "cpu":
			fields["hard_cpu_cores_limit"] = ki.convertQuantity(val.String(), 1)
		case "memory":
			fields["hard_memory_bytes_limit"] = ki.convertQuantity(val.String(), 1)
		case "pods":
			fields["hard_pods_limit"] = atoi(val.String())
		}
	}

	for resourceName, val := range r.Status.Used {
		switch resourceName {
		case "cpu":
			fields["used_cpu_cores"] = ki.convertQuantity(val.String(), 1)
		case "memory":
			fields["used_memory_bytes"] = ki.convertQuantity(val.String(), 1)
		case "pods":
			fields["used_pods"] = atoi(val.String())
		}
	}

	acc.AddFields(resourcequotaMeasurement, fields, tags)
}
