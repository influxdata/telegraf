package kube_inventory

import (
	"context"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/influxdata/telegraf"
)

func collectPods(ctx context.Context, acc telegraf.Accumulator, ki *KubernetesInventory) {
	var list corev1.PodList
	listRef := &list
	var err error

	if ki.KubeletURL != "" {
		err = ki.queryPodsFromKubelet(ki.KubeletURL+"/pods", listRef)
	} else {
		listRef, err = ki.client.getPods(ctx, ki.NodeName)
	}

	if err != nil {
		acc.AddError(err)
		return
	}

	var deployments map[string]string
	if ki.PodDeploymentTag {
		replicaSets, err := ki.client.getReplicaSets(ctx)
		if err != nil {
			acc.AddError(fmt.Errorf("listing replicasets failed: %w", err))
		} else {
			deployments = replicaSetDeployments(replicaSets)
		}
	}

	for i := range listRef.Items {
		ki.gatherPod(&listRef.Items[i], deployments, acc)
	}
}

func (ki *KubernetesInventory) gatherPod(p *corev1.Pod, deployments map[string]string, acc telegraf.Accumulator) {
	creationTS := p.GetCreationTimestamp()
	if creationTS.IsZero() {
		return
	}

	initContainers := make([]corev1.Container, 0, len(p.Spec.InitContainers))
	for _, c := range p.Spec.InitContainers {
		if c.RestartPolicy != nil && *c.RestartPolicy == corev1.ContainerRestartPolicyAlways {
			initContainers = append(initContainers, c)
		}
	}
	deployment := podDeployment(p, deployments)
	ki.gatherPodContainers(p, deployment, initContainers, p.Status.InitContainerStatuses, acc)
	ki.gatherPodContainers(p, deployment, p.Spec.Containers, p.Status.ContainerStatuses, acc)
}

func (ki *KubernetesInventory) gatherPodContainers(
	p *corev1.Pod,
	deployment string,
	containers []corev1.Container,
	statuses []corev1.ContainerStatus,
	acc telegraf.Accumulator,
) {
	containerList := make(map[string]*corev1.ContainerStatus, len(statuses))
	for i := range statuses {
		containerList[statuses[i].Name] = &statuses[i]
	}

	for _, c := range containers {
		cs, ok := containerList[c.Name]
		if !ok {
			cs = &corev1.ContainerStatus{}
		}
		ki.gatherPodContainer(p, deployment, *cs, c, acc)
	}
}

func (ki *KubernetesInventory) gatherPodContainer(
	p *corev1.Pod,
	deployment string,
	cs corev1.ContainerStatus,
	c corev1.Container,
	acc telegraf.Accumulator,
) {
	stateCode := 3
	stateReason := ""
	state := "unknown"
	readiness := "unready"

	switch {
	case cs.State.Running != nil:
		stateCode = 0
		state = "running"
	case cs.State.Terminated != nil:
		stateCode = 1
		state = "terminated"
		stateReason = cs.State.Terminated.Reason
	case cs.State.Waiting != nil:
		stateCode = 2
		state = "waiting"
		stateReason = cs.State.Waiting.Reason
	}

	if cs.Ready {
		readiness = "ready"
	}

	fields := map[string]any{
		"restarts_total": cs.RestartCount,
		"state_code":     stateCode,
	}

	// deprecated in 1.15: use `state_reason` instead
	if state == "terminated" {
		fields["terminated_reason"] = stateReason
	}

	if stateReason != "" {
		fields["state_reason"] = stateReason
	}

	phaseReason := p.Status.Reason
	if phaseReason != "" {
		fields["phase_reason"] = phaseReason
	}

	tags := map[string]string{
		"container_name": c.Name,
		"namespace":      p.Namespace,
		"node_name":      p.Spec.NodeName,
		"pod_name":       p.Name,
		"phase":          string(p.Status.Phase),
		"state":          state,
		"readiness":      readiness,
	}
	splitImage := strings.Split(c.Image, ":")
	if len(splitImage) == 2 {
		tags["version"] = splitImage[1]
	}
	tags["image"] = splitImage[0]
	if deployment != "" {
		tags["deployment_name"] = deployment
	}
	for key, val := range p.Spec.NodeSelector {
		if ki.selectorFilter.Match(key) {
			tags["node_selector_"+key] = val
		}
	}

	req := c.Resources.Requests
	lim := c.Resources.Limits

	for resourceName, val := range req {
		switch resourceName {
		case "cpu":
			fields["resource_requests_millicpu_units"] = ki.convertQuantity(val.String(), 1000)
		case "memory":
			fields["resource_requests_memory_bytes"] = ki.convertQuantity(val.String(), 1)
		}
	}
	for resourceName, val := range lim {
		switch resourceName {
		case "cpu":
			fields["resource_limits_millicpu_units"] = ki.convertQuantity(val.String(), 1000)
		case "memory":
			fields["resource_limits_memory_bytes"] = ki.convertQuantity(val.String(), 1)
		}
	}

	for _, val := range p.Status.Conditions {
		conditiontags := map[string]string{
			"container_name": c.Name,
			"image":          splitImage[0],
			"status":         string(val.Status),
			"namespace":      p.Namespace,
			"node_name":      p.Spec.NodeName,
			"pod_name":       p.Name,
			"condition":      string(val.Type),
		}
		if len(splitImage) == 2 {
			conditiontags["version"] = splitImage[1]
		}
		if deployment != "" {
			conditiontags["deployment_name"] = deployment
		}
		running := 0
		podready := 0
		if val.Status == "True" {
			if val.Type == "Ready" {
				podready = 1
			}
			running = 1
		} else if val.Status == "Unknown" {
			if val.Type == "Ready" {
				podready = 0
			}
			running = 2
		}
		conditionfields := map[string]any{
			"status_condition": running,
			"ready":            podready,
		}
		acc.AddFields(podContainerMeasurement, conditionfields, conditiontags)
	}

	acc.AddFields(podContainerMeasurement, fields, tags)
}

// replicaSetDeployments maps each Deployment-owned ReplicaSet, keyed by "namespace/name", to its Deployment.
func replicaSetDeployments(list *appsv1.ReplicaSetList) map[string]string {
	deployments := make(map[string]string, len(list.Items))
	for i := range list.Items {
		rs := &list.Items[i]
		if owner := metav1.GetControllerOf(rs); owner != nil && owner.Kind == "Deployment" {
			deployments[rs.Namespace+"/"+rs.Name] = owner.Name
		}
	}
	return deployments
}

func podDeployment(p *corev1.Pod, deployments map[string]string) string {
	owner := metav1.GetControllerOf(p)
	if owner == nil || owner.Kind != "ReplicaSet" {
		return ""
	}
	return deployments[p.Namespace+"/"+owner.Name]
}
