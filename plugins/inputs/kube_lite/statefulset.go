package kube_lite

import (
	"context"

	"github.com/ericchiang/k8s/apis/apps/v1beta1"

	"github.com/influxdata/telegraf"
)

var (
	statefulSetMeasurement = "kube_statefulset"
)

func registerStatefulSetCollector(ctx context.Context, acc telegraf.Accumulator, ks *KubernetesState) {
	list, err := ks.client.getStatefulSets(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, s := range list.Items {
		if err = ks.gatherStatefulSet(*s, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ks *KubernetesState) gatherStatefulSet(statefulSet v1beta1.StatefulSet, acc telegraf.Accumulator) error {
	status := statefulSet.Status
	fields := map[string]interface{}{
		"metadata_generation":     *statefulSet.Metadata.Generation,
		"status_replicas":         *status.Replicas,
		"status_replicas_current": *status.CurrentReplicas,
		"status_replicas_ready":   *status.ReadyReplicas,
		"status_replicas_updated": *status.UpdatedReplicas,
	}
	tags := map[string]string{
		"name":      *statefulSet.Metadata.Name,
		"namespace": *statefulSet.Metadata.Namespace,
	}
	if statefulSet.Spec.Replicas != nil {
		fields["replicas"] = *statefulSet.Spec.Replicas
	}
	if statefulSet.Status.ObservedGeneration != nil {
		fields["status_observed_generation"] = *statefulSet.Status.ObservedGeneration
	}

	acc.AddFields(statefulSetMeasurement, fields, tags)
	return nil
}
