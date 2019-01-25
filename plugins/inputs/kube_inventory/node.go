package kube_inventory

import (
	"context"

	"github.com/ericchiang/k8s/apis/core/v1"

	"github.com/influxdata/telegraf"
)

func collectNodes(ctx context.Context, acc telegraf.Accumulator, ki *KubernetesInventory) {
	list, err := ki.client.getNodes(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, n := range list.Items {
		if err = ki.gatherNode(*n, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ki *KubernetesInventory) gatherNode(n v1.Node, acc telegraf.Accumulator) error {
	fields := map[string]interface{}{}
	tags := map[string]string{
		"node_name": *n.Metadata.Name,
	}

	for resourceName, val := range n.Status.Capacity {
		switch resourceName {
		case "cpu":
			fields["capacity_cpu_cores"] = atoi(val.GetString_())
		case "memory":
			fields["capacity_memory_bytes"] = convertQuantity(val.GetString_(), 1)
		case "pods":
			fields["capacity_pods"] = atoi(val.GetString_())
		}
	}

	for resourceName, val := range n.Status.Allocatable {
		switch resourceName {
		case "cpu":
			fields["allocatable_cpu_cores"] = atoi(val.GetString_())
		case "memory":
			fields["allocatable_memory_bytes"] = convertQuantity(val.GetString_(), 1)
		case "pods":
			fields["allocatable_pods"] = atoi(val.GetString_())
		}
	}

	acc.AddFields(nodeMeasurement, fields, tags)

	return nil
}
