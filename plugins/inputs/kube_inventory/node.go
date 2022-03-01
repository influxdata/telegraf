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

	ki.gatherNodeCount(len(list.Items), acc)

	for _, n := range list.Items {
		ki.gatherNode(n, acc)
	}
}

func (ki *KubernetesInventory) gatherNodeCount(count int, acc telegraf.Accumulator) {
	fields := map[string]interface{}{}
	tags := map[string]string{}
	fields["count"] = count

	acc.AddFields(nodeMeasurement, fields, tags)
}

func (ki *KubernetesInventory) gatherNode(n corev1.Node, acc telegraf.Accumulator) {
	fields := map[string]interface{}{}
	tags := map[string]string{
		"node_name":         n.Name,
		"cluster_namespace": n.Annotations["cluster.x-k8s.io/cluster-namespace"],
	}

	for resourceName, val := range n.Status.Capacity {
		switch resourceName {
		case "cpu":
			fields["capacity_cpu_cores"] = ki.convertQuantity(val.String(), 1)
			fields["capacity_millicpu_cores"] = ki.convertQuantity(val.String(), 1000)
		case "memory":
			fields["capacity_memory_bytes"] = ki.convertQuantity(val.String(), 1)
		case "pods":
			fields["capacity_pods"] = atoi(val.String())
		}
	}

	for resourceName, val := range n.Status.Allocatable {
		switch resourceName {
		case "cpu":
			fields["allocatable_cpu_cores"] = ki.convertQuantity(val.String(), 1)
			fields["allocatable_millicpu_cores"] = ki.convertQuantity(val.String(), 1000)
		case "memory":
			fields["allocatable_memory_bytes"] = ki.convertQuantity(val.String(), 1)
		case "pods":
			fields["allocatable_pods"] = atoi(val.String())
		}
	}

	for _, val := range n.Status.Conditions {
		conditionfields := map[string]interface{}{}
		conditiontags := map[string]string{}
		conditiontags["status"] = string(val.Status)
		conditiontags["condition"] = string(val.Type)
		conditiontags["node_name"] = n.Name
		running := 0
		nodeready := 0
		if val.Status == "True" {
			if val.Type == "Ready" {
				nodeready = 1
			}
			running = 1
		} else if val.Status == "Unknown" {
			if val.Type == "Ready" {
				nodeready = 0
			}
			running = 2
		}
		conditionfields["status_condition"] = running
		conditionfields["ready"] = nodeready
		acc.AddFields(nodeMeasurement, conditionfields, conditiontags)
	}

	unschedulable := 0
	if n.Spec.Unschedulable {
		unschedulable = 1
	}
	fields["spec_unschedulable"] = unschedulable

	acc.AddFields(nodeMeasurement, fields, tags)
}
