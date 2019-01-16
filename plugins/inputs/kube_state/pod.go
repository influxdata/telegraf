package kube_state

import (
	"context"
	"strings"
	"time"

	"github.com/ericchiang/k8s/apis/core/v1"

	"github.com/influxdata/telegraf"
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
	if p.Metadata.CreationTimestamp.GetSeconds() == 0 && p.Metadata.CreationTimestamp.GetNanos() == 0 {
		return nil
	}

	for i, cs := range p.Status.ContainerStatuses {
		c := p.Spec.Containers[i]
		gatherPodContainer(*p.Spec.NodeName, p, *cs, *c, acc)
	}

	for _, c := range p.Status.Conditions {
		if c.LastTransitionTime.GetSeconds() == 0 && c.LastTransitionTime.GetNanos() == 0 {
			continue
		}
		gatherPodStatus(p, *c, acc)
	}

	return nil
}

func gatherPodContainer(nodeName string, p v1.Pod, cs v1.ContainerStatus, c v1.Container, acc telegraf.Accumulator) {
	fields := map[string]interface{}{
		"restarts_total":    cs.GetRestartCount(),
		"running":           boolInt(cs.State.Running != nil),
		"terminated":        boolInt(cs.State.Terminated != nil),
		"terminated_reason": cs.State.Terminated.GetReason(),
	}
	tags := map[string]string{
		"container_name": *c.Name,
		"namespace":      *p.Metadata.Namespace,
		"node_name":      *p.Spec.NodeName,
		"pod_name":       *p.Metadata.Name,
	}

	req := c.Resources.Requests
	lim := c.Resources.Limits

	for resourceName, val := range req {
		switch resourceName {
		case "cpu":
			fields["resource_requests_cpu_units"] = *val.String_
		case "memory":
			fields["resource_requests_memory_bytes"] = *val.String_
		}
	}
	for resourceName, val := range lim {
		switch resourceName {
		case "cpu":
			fields["resource_limits_cpu_units"] = *val.String_
		case "memory":
			fields["resource_limits_memory_bytes"] = *val.String_
		}
	}

	acc.AddFields(podContainerMeasurement, fields, tags)
}

func gatherPodStatus(p v1.Pod, c v1.PodCondition, acc telegraf.Accumulator) {
	tags := map[string]string{
		"namespace": p.Metadata.GetNamespace(),
		"pod_name":  p.Metadata.GetName(),
		"node_name": p.Spec.GetNodeName(),
		"reason":    p.Status.GetReason(),
	}
	fields := map[string]interface{}{
		"last_transition_time": time.Unix(c.LastTransitionTime.GetSeconds(), int64(c.LastTransitionTime.GetNanos())).UnixNano(),
	}

	switch strings.ToLower(*c.Type) {
	case "ready":
		fields["ready"] = "true"
	default:
		fields["ready"] = "false"
	}

	acc.AddFields(podStatusMeasurement, fields, tags)
}
