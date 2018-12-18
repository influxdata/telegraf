package kube_lite

// - resource_requests_cpu_cores
// - resource_limits_cpu_cores
// - resource_requests_memory_bytes
// - resource_limits_memory_bytes
// - status_ready

import (
	"context"
	"time"

	"github.com/ericchiang/k8s/apis/core/v1"

	"github.com/influxdata/telegraf"
)

var (
	podStatusMeasurement    = "kube_pod_status"
	podContainerMeasurement = "kube_pod_container"
)

func registerPodCollector(ctx context.Context, acc telegraf.Accumulator, ks *KubernetesState) {
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
	if p.Metadata.CreationTimestamp.Size() == 0 {
		return nil
	}

	fields := map[string]interface{}{}
	tags := map[string]string{
		"namespace": *p.Metadata.Namespace,
	}

	for i, cs := range p.Status.ContainerStatuses {
		c := p.Spec.Containers[i]
		gatherPodContainer(*p.Spec.NodeName, p, *cs, *c, acc)
	}

	for _, c := range p.Status.Conditions {
		if c.LastTransitionTime.Size() == 0 {
			continue
		}
		gatherPodStatus(tags, fields, *c, acc)
	}

	return nil
}

func gatherPodContainer(nodeName string, p v1.Pod, cs v1.ContainerStatus, c v1.Container, acc telegraf.Accumulator) {
	fields := map[string]interface{}{}
	tags := map[string]string{
		"namespace": *p.Metadata.Namespace,
	}

	req := c.Resources.Requests
	lim := c.Resources.Limits

	for resourceName, val := range req {
		switch resourceName {
		case "cpu":
			// todo: use terrible atoi?
			fields["resource_requests_cpu_cores"] = *val.String_
			// default:
			// 	// todo: ensure `Size` is what we expect
			// 	fields["resource_requests_"+sanitizeLabelName(string(resourceName))+"_bytes"] = val.Size()
		}
	}
	for resourceName, val := range lim {
		switch resourceName {
		case "cpu":
			// todo: use terrible atoi?
			fields["resource_limits_cpu_cores"] = *val.String_
			// default:
			// 	// todo: ensure `Size` is what we expect
			// 	fields["resource_limits_"+sanitizeLabelName(string(resourceName))+"_bytes"] = val.Size()
		}
	}

	acc.AddFields(podContainerMeasurement, fields, tags)
}

func gatherPodStatus(t map[string]string, f map[string]interface{}, c v1.PodCondition, acc telegraf.Accumulator) {
	tags := make(map[string]string)
	for k, v := range t {
		tags[k] = v
	}
	fields := make(map[string]interface{})
	for k, v := range f {
		fields[k] = v
	}

	switch *c.Type {
	case "ready":
		tags["ready"] = "true"
	default:
		tags["ready"] = "false"
	}

	// todo: ensure time works properly
	acc.AddFields(podStatusMeasurement, fields, tags, time.Unix(c.LastTransitionTime.GetSeconds(), int64(c.LastTransitionTime.GetNanos())))
}
