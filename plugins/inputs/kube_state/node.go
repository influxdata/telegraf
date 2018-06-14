package kube_state

import (
	"context"
	"strconv"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/apis/core/v1/helper"
)

var (
	nodeMeasurement                 = "kube_node"
	nodeTaintMeasurement            = "kube_node_spec_taint"
	nodeStatusConditionsMeasurement = "kube_node_status_conditions"
)

func registerNodeCollector(ctx context.Context, acc telegraf.Accumulator, ks *KubenetesState) {
	list, err := ks.client.getNodes(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, n := range list.Items {
		if err = ks.gatherNode(n, acc); err != nil {
			acc.AddError(err)
			return
		}
	}

}
func (ks *KubenetesState) gatherNode(n v1.Node, acc telegraf.Accumulator) error {
	fields := map[string]interface{}{}
	tags := map[string]string{
		"node":                      n.Name,
		"kernel_version":            n.Status.NodeInfo.KernelVersion,
		"os_image":                  n.Status.NodeInfo.OSImage,
		"container_runtime_version": n.Status.NodeInfo.ContainerRuntimeVersion,
		"kubelet_version":           n.Status.NodeInfo.KubeletVersion,
		"kubeproxy_version":         n.Status.NodeInfo.KubeProxyVersion,
		"status_phase":              strings.ToLower(string(n.Status.Phase)),
		"provider_id":               n.Spec.ProviderID,
		"spec_unschedulable":        strconv.FormatBool(n.Spec.Unschedulable),
	}

	if !n.CreationTimestamp.IsZero() {
		fields["created"] = n.CreationTimestamp.Unix()
	}

	for k, v := range n.Labels {
		tags["label_"+sanitizeLabelName(k)] = v
	}

	var wg sync.WaitGroup

	// Collect node taints
	for _, taint := range n.Spec.Taints {
		wg.Add(1)
		go func(n v1.Node, taint v1.Taint) {
			defer wg.Done()
			gatherNodeTaint(n, taint, acc)
		}(n, taint)
	}

	for _, condition := range n.Status.Conditions {
		wg.Add(1)
		go func(n v1.Node, condition v1.NodeCondition) {
			defer wg.Done()
			gatherNodeStatusCondition(n, condition, acc)
		}(n, condition)
	}

	capacity := n.Status.Capacity
	allocatable := n.Status.Allocatable
	for resourceName, val := range capacity {
		switch resourceName {
		case v1.ResourceCPU:
			fields["status_capacity_cpu_cores"] = val.MilliValue() / 1000
		case v1.ResourceStorage:
			fallthrough
		case v1.ResourceEphemeralStorage:
			fallthrough
		case v1.ResourceMemory:
			fields["status_capacity_"+sanitizeLabelName(string(resourceName))+"_bytes"] = val.MilliValue() / 1000
		case v1.ResourcePods:
			fields["status_capacity_pods"] = val.MilliValue() / 1000
		default:
			if helper.IsHugePageResourceName(resourceName) || helper.IsAttachableVolumeResourceName(resourceName) {
				fields["status_capacity_"+sanitizeLabelName(string(resourceName))+"_bytes"] = val.MilliValue() / 1000
			}
			if helper.IsExtendedResourceName(resourceName) {
				fields["status_capacity_"+sanitizeLabelName(string(resourceName))] = val.MilliValue() / 1000
			}
		}
	}

	for resourceName, val := range allocatable {
		switch resourceName {
		case v1.ResourceCPU:
			fields["status_allocatable_cpu_cores"] = val.MilliValue() / 1000
		case v1.ResourceStorage:
			fallthrough
		case v1.ResourceEphemeralStorage:
			fallthrough
		case v1.ResourceMemory:
			fields["status_allocatable_"+sanitizeLabelName(string(resourceName))+"_bytes"] = val.MilliValue() / 1000
		case v1.ResourcePods:
			fields["status_allocatable_pods"] = val.MilliValue() / 1000
		default:
			if helper.IsHugePageResourceName(resourceName) || helper.IsAttachableVolumeResourceName(resourceName) {
				fields["status_allocatable_"+sanitizeLabelName(string(resourceName))+"_bytes"] = val.MilliValue() / 1000
			}
			if helper.IsExtendedResourceName(resourceName) {
				fields["status_allocatable_"+sanitizeLabelName(string(resourceName))] = val.MilliValue() / 1000
			}
		}
	}

	wg.Wait()
	acc.AddFields(nodeMeasurement, fields, tags)
	return nil
}

func gatherNodeStatusCondition(n v1.Node, condition v1.NodeCondition, acc telegraf.Accumulator) {
	fields := map[string]interface{}{
		"gauge": 1,
	}
	tags := map[string]string{
		"node":      n.Name,
		"condition": strings.ToLower(string(condition.Type)),
		"status":    strings.ToLower(string(condition.Status)),
	}

	acc.AddFields(nodeStatusConditionsMeasurement, fields, tags, condition.LastTransitionTime.Time)
}

func gatherNodeTaint(n v1.Node, taint v1.Taint, acc telegraf.Accumulator) {
	fields := map[string]interface{}{
		"gauge": 1,
	}
	tags := map[string]string{
		"node":   n.Name,
		"key":    taint.Key,
		"value":  taint.Value,
		"effect": string(taint.Effect),
	}

	acc.AddFields(nodeTaintMeasurement, fields, tags)

}
