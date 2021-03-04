package kube_inventory

import (
	"context"
	"time"

	v1 "k8s.io/api/apps/v1"

	"github.com/influxdata/telegraf"
)

func collectStatefulSets(ctx context.Context, acc telegraf.Accumulator, ki *KubernetesInventory) {
	list, err := ki.client.getStatefulSets(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, s := range list.Items {
		if err = ki.gatherStatefulSet(s, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ki *KubernetesInventory) gatherStatefulSet(s v1.StatefulSet, acc telegraf.Accumulator) error {
	status := s.Status
	fields := map[string]interface{}{
		"created":             time.Unix(int64(s.GetCreationTimestamp().Second()), int64(s.GetCreationTimestamp().Nanosecond())).UnixNano(),
		"generation":          s.Generation,
		"replicas":            status.Replicas,
		"replicas_current":    status.CurrentReplicas,
		"replicas_ready":      status.ReadyReplicas,
		"replicas_updated":    status.UpdatedReplicas,
		"spec_replicas":       *s.Spec.Replicas,
		"observed_generation": s.Status.ObservedGeneration,
	}
	tags := map[string]string{
		"statefulset_name": s.Name,
		"namespace":        s.Namespace,
	}
	for key, val := range s.Spec.Selector.MatchLabels {
		if ki.selectorFilter.Match(key) {
			tags["selector_"+key] = val
		}
	}

	acc.AddFields(statefulSetMeasurement, fields, tags)

	return nil
}
