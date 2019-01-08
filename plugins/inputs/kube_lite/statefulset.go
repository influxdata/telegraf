package kube_lite

import (
	"context"
	"time"

	"github.com/ericchiang/k8s/apis/apps/v1beta1"

	"github.com/influxdata/telegraf"
)

var (
	statefulSetMeasurement = "kube_statefulset"
)

func collectStatefulSets(ctx context.Context, acc telegraf.Accumulator, ks *KubernetesState) {
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

func (ks *KubernetesState) gatherStatefulSet(s v1beta1.StatefulSet, acc telegraf.Accumulator) error {
	status := s.Status
	fields := map[string]interface{}{
		"metadata_generation":     *s.Metadata.Generation,
		"status_replicas":         *status.Replicas,
		"status_replicas_current": *status.CurrentReplicas,
		"status_replicas_ready":   *status.ReadyReplicas,
		"status_replicas_updated": *status.UpdatedReplicas,
	}
	tags := map[string]string{
		"name":      *s.Metadata.Name,
		"namespace": *s.Metadata.Namespace,
	}

	if s.Spec.Replicas != nil {
		fields["replicas"] = *s.Spec.Replicas
	}
	if s.Status.ObservedGeneration != nil {
		fields["status_observed_generation"] = *s.Status.ObservedGeneration
	}

	acc.AddFields(statefulSetMeasurement, fields, tags, time.Unix(s.Metadata.CreationTimestamp.GetSeconds(), int64(s.Metadata.CreationTimestamp.GetNanos())))

	return nil
}
