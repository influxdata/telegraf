package kube_state

import (
	"context"

	"github.com/ericchiang/k8s/apis/core/v1"

	"github.com/influxdata/telegraf"
)

func collectNodes(ctx context.Context, acc telegraf.Accumulator, ks *KubernetesState) {
	list, err := ks.client.getNodes(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, n := range list.Items {
		if err = ks.gatherNode(*n, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ks *KubernetesState) gatherNode(n v1.Node, acc telegraf.Accumulator) error {
	fields := map[string]interface{}{}
	tags := map[string]string{
		"node_name": *n.Metadata.Name,
	}

	for resourceName, val := range n.Status.Capacity {
		switch resourceName {
		case "cpu":
			fields["capacity_cpu_cores"] = atoi(val.GetString_())
		case "memory":
			fields["capacity_memory_bytes"] = val.GetString_()
		case "pods":
			fields["capacity_pods"] = atoi(val.GetString_())
		}
	}

	for resourceName, val := range n.Status.Allocatable {
		switch resourceName {
		case "cpu":
			fields["allocatable_cpu_cores"] = atoi(val.GetString_())
		case "memory":
			fields["allocatable_memory_bytes"] = val.GetString_()
		case "pods":
			fields["allocatable_pods"] = atoi(val.GetString_())
		}
	}

	acc.AddFields(nodeMeasurement, fields, tags)

	return nil
}
