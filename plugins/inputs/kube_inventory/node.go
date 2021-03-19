package kube_inventory

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"github.com/influxdata/telegraf"
)

func collectNodes(ctx context.Context, acc telegraf.Accumulator, ki *KubernetesInventory) {
	list, err := ki.client.getNodes(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, n := range list.Items {
		ki.gatherNode(n, acc)
	}
}

func (ki *KubernetesInventory) gatherNode(n corev1.Node, acc telegraf.Accumulator) {
	fields := map[string]interface{}{}
	tags := map[string]string{
		"node_name": n.Name,
	}

	for resourceName, val := range n.Status.Capacity {
		switch resourceName {
		case "cpu":
			fields["capacity_cpu_cores"] = convertQuantity(string(val.Format), 1)
			fields["capacity_millicpu_cores"] = convertQuantity(string(val.Format), 1000)
		case "memory":
			fields["capacity_memory_bytes"] = convertQuantity(string(val.Format), 1)
		case "pods":
			fields["capacity_pods"] = atoi(string(val.Format))
		}
	}

	for resourceName, val := range n.Status.Allocatable {
		switch resourceName {
		case "cpu":
			fields["allocatable_cpu_cores"] = convertQuantity(string(val.Format), 1)
			fields["allocatable_millicpu_cores"] = convertQuantity(string(val.Format), 1000)
		case "memory":
			fields["allocatable_memory_bytes"] = convertQuantity(string(val.Format), 1)
		case "pods":
			fields["allocatable_pods"] = atoi(string(val.Format))
		}
	}

	acc.AddFields(nodeMeasurement, fields, tags)
}
