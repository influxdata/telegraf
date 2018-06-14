package kube_state

import (
	"context"

	"github.com/influxdata/telegraf"
	"k8s.io/api/apps/v1beta2"
)

var (
	replicaSetMeasurement = "kube_replicasets"
)

func registerReplicaSetCollector(ctx context.Context, acc telegraf.Accumulator, ks *KubenetesState) {
	list, err := ks.client.getReplicaSets(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, d := range list.Items {
		if err = ks.gatherReplicaSet(d, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ks *KubenetesState) gatherReplicaSet(d v1beta2.ReplicaSet, acc telegraf.Accumulator) error {
	status := d.Status
	fields := map[string]interface{}{
		"metadata_generation":           d.ObjectMeta.Generation,
		"status_replicas":               status.Replicas,
		"status_fully_labeled_replicas": status.FullyLabeledReplicas,
		"status_ready_replicas":         status.ReadyReplicas,
		"status_observed_generation":    status.ObservedGeneration,
	}
	if !d.CreationTimestamp.IsZero() {
		fields["created"] = d.CreationTimestamp.Unix()
	}
	tags := map[string]string{
		"namespace":  d.Namespace,
		"replicaset": d.Name,
	}
	if d.Spec.Replicas != nil {
		fields["spec_replicas"] = *d.Spec.Replicas

	}
	acc.AddFields(replicaSetMeasurement, fields, tags)
	return nil
}
