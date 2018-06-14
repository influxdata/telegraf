package kube_state

import (
	"context"

	"github.com/influxdata/telegraf"
	"k8s.io/api/core/v1"
)

var (
	replicationControllerMeasurement = "kube_replicationcontroller"
)

func registerReplicationControllerCollector(ctx context.Context, acc telegraf.Accumulator, ks *KubenetesState) {
	list, err := ks.client.getReplicationControllers(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, d := range list.Items {
		if err = ks.gatherReplicationController(d, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ks *KubenetesState) gatherReplicationController(d v1.ReplicationController, acc telegraf.Accumulator) error {
	status := d.Status

	fields := map[string]interface{}{
		"metadata_generation":           d.ObjectMeta.Generation,
		"status_replicas":               status.Replicas,
		"status_fully_labeled_replicas": status.FullyLabeledReplicas,
		"status_ready_replicas":         status.ReadyReplicas,
		"status_available_replicas":     status.AvailableReplicas,
		"status_observed_generation":    status.ObservedGeneration,
	}
	if !d.CreationTimestamp.IsZero() {
		fields["created"] = d.CreationTimestamp.Unix()
	}
	tags := map[string]string{
		"namespace":             d.Namespace,
		"replicationcontroller": d.Name,
	}
	if d.Spec.Replicas != nil {
		fields["spec_replicas"] = *d.Spec.Replicas

	}
	acc.AddFields(replicationControllerMeasurement, fields, tags)
	return nil
}
