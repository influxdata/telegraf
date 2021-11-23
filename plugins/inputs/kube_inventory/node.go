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

	convertConditionStatusToNumber := func(status corev1.ConditionStatus) int64 {
		switch status {
		case corev1.ConditionTrue:
			return 1
		case corev1.ConditionFalse:
			return 0
		case corev1.ConditionUnknown:
			return 2
		default:
			return 2
		}
	}

	for _, condition := range n.Status.Conditions {
		status := convertConditionStatusToNumber(condition.Status)
		var conditionTypeName string
		switch condition.Type {
		case corev1.NodeReady:
			conditionTypeName = "condition_ready"
		case corev1.NodeDiskPressure:
			conditionTypeName = "condition_disk_pressure"
		case corev1.NodeMemoryPressure:
			conditionTypeName = "condition_memory_pressure"
		case corev1.NodeNetworkUnavailable:
			conditionTypeName = "condition_network_unavailable"
		case corev1.NodePIDPressure:
			conditionTypeName = "condition_pid_pressure"
		}
		if conditionTypeName != "" {
			fields[conditionTypeName] = status
		}
	}

	acc.AddFields(nodeMeasurement, fields, tags)
}
