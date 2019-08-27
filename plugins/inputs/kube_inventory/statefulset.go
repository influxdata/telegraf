package kube_inventory

import (
	"context"
	"time"

	"github.com/ericchiang/k8s/apis/apps/v1beta1"

	"github.com/influxdata/telegraf"
)

func collectStatefulSets(ctx context.Context, acc telegraf.Accumulator, ki *KubernetesInventory) {
	list, err := ki.client.getStatefulSets(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, s := range list.Items {
		if err = ki.gatherStatefulSet(*s, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ki *KubernetesInventory) gatherStatefulSet(s v1beta1.StatefulSet, acc telegraf.Accumulator) error {
	status := s.Status
	fields := map[string]interface{}{
		"created":             time.Unix(s.Metadata.CreationTimestamp.GetSeconds(), int64(s.Metadata.CreationTimestamp.GetNanos())).UnixNano(),
		"generation":          *s.Metadata.Generation,
		"replicas":            *status.Replicas,
		"replicas_current":    *status.CurrentReplicas,
		"replicas_ready":      *status.ReadyReplicas,
		"replicas_updated":    *status.UpdatedReplicas,
		"spec_replicas":       *s.Spec.Replicas,
		"observed_generation": *s.Status.ObservedGeneration,
	}
	tags := map[string]string{
		"statefulset_name": *s.Metadata.Name,
		"namespace":        *s.Metadata.Namespace,
	}

	acc.AddFields(statefulSetMeasurement, fields, tags)

	return nil
}
