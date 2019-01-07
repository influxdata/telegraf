package kube_lite

// - resource_requests_cpu_cores
// - resource_limits_cpu_cores
// - resource_requests_memory_bytes
// - resource_limits_memory_bytes
// - status_ready

import (
	"context"
	"strings"
	"time"

	"github.com/ericchiang/k8s/apis/core/v1"

	"github.com/influxdata/telegraf"
)

var (
	podStatusMeasurement    = "kube_pod_status"
	podContainerMeasurement = "kube_pod_container"
)

func collectPods(ctx context.Context, acc telegraf.Accumulator, ks *KubernetesState) {
	list, err := ks.client.getPods(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, p := range list.Items {
		if err = ks.gatherPod(*p, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ks *KubernetesState) gatherPod(p v1.Pod, acc telegraf.Accumulator) error {
	// todo: size is likely wrong
	if p.Metadata.CreationTimestamp.Size() == 0 {
		return nil
	}

	for i, cs := range p.Status.ContainerStatuses {
		c := p.Spec.Containers[i]
		gatherPodContainer(*p.Spec.NodeName, p, *cs, *c, acc)
	}

	for _, c := range p.Status.Conditions {
		if c.LastTransitionTime.Size() == 0 {
			continue
		}
		gatherPodStatus(p, *c, acc)
	}

	return nil
}

func gatherPodContainer(nodeName string, p v1.Pod, cs v1.ContainerStatus, c v1.Container, acc telegraf.Accumulator) {
	fields := map[string]interface{}{
		"status_restarts_total":    cs.GetRestartCount(),
		"status_running":           boolInt(cs.State.Running != nil),
		"status_terminated":        boolInt(cs.State.Terminated != nil),
		"status_terminated_reasom": cs.State.Terminated.GetReason(),
	}
	tags := map[string]string{
		"namespace": *p.Metadata.Namespace,
		"name":      *c.Name,
		"node":      *p.Spec.NodeName,
		"pod":       *p.Metadata.Name,
		// "reason":      ,
	}

	req := c.Resources.Requests
	lim := c.Resources.Limits

	for resourceName, val := range req {
		switch resourceName {
		case "cpu":
			// todo: better way to get value
			fields["resource_requests_cpu_cores"] = atoi(*val.String_)
		default:
			// todo: better way to get value
			fields["resource_requests_"+sanitizeLabelName(string(resourceName))+"_bytes"] = atoi(*val.String_)
		}
	}
	for resourceName, val := range lim {
		switch resourceName {
		case "cpu":
			// todo: better way to get value
			fields["resource_limits_cpu_cores"] = atoi(*val.String_)
		default:
			// todo: better way to get value
			fields["resource_limits_"+sanitizeLabelName(string(resourceName))+"_bytes"] = atoi(*val.String_)
		}
	}

	acc.AddFields(podContainerMeasurement, fields, tags)
}

func gatherPodStatus(p v1.Pod, c v1.PodCondition, acc telegraf.Accumulator) {
	tags := map[string]string{
		"namespace": p.Metadata.GetNamespace(),
		"name":      p.Metadata.GetName(),
		"node":      p.Spec.GetNodeName(),
		"reason":    p.Status.GetReason(),
	}

	fields := make(map[string]interface{})

	switch strings.ToLower(*c.Type) {
	case "ready":
		fields["ready"] = "true"
	default:
		fields["ready"] = "false"
	}

	// todo: ensure time works properly
	acc.AddFields(podStatusMeasurement, fields, tags, time.Unix(c.LastTransitionTime.GetSeconds(), int64(c.LastTransitionTime.GetNanos())))
}
