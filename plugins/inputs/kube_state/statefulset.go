package kube_state

import (
	"context"

	"github.com/influxdata/telegraf"
	"k8s.io/api/apps/v1beta1"
)

var (
	statefulSetMeasurement = "kube_statefulset"
)

func registerStatefulSetCollector(ctx context.Context, acc telegraf.Accumulator, ks *KubenetesState) {
	list, err := ks.client.getStatefulSets(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, s := range list.Items {
		if err = ks.gatherStatefulSet(s, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ks *KubenetesState) gatherStatefulSet(statefulSet v1beta1.StatefulSet, acc telegraf.Accumulator) error {
	status := statefulSet.Status

	fields := map[string]interface{}{
		"metadata_generation":     statefulSet.ObjectMeta.Generation,
		"status_replicas":         status.Replicas,
		"status_replicas_current": status.CurrentReplicas,
		"status_replicas_ready":   status.ReadyReplicas,
		"status_replicas_updated": status.UpdatedReplicas,
	}
	if !statefulSet.CreationTimestamp.IsZero() {
		fields["created"] = statefulSet.CreationTimestamp.Time.Unix()
	}
	tags := map[string]string{
		"namespace":   statefulSet.Namespace,
		"statefulset": statefulSet.Name,
	}
	if statefulSet.Spec.Replicas != nil {
		fields["replicas"] = *statefulSet.Spec.Replicas
	}
	if statefulSet.Status.ObservedGeneration != nil {
		fields["status_observed_generation"] = *statefulSet.Status.ObservedGeneration
	}
	for k, v := range statefulSet.Labels {
		tags["label_"+sanitizeLabelName(k)] = v
	}
	acc.AddFields(statefulSetMeasurement, fields, tags)
	return nil
}
