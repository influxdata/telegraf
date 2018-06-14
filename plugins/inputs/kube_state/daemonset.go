package kube_state

import (
	"context"

	"github.com/influxdata/telegraf"
	"k8s.io/api/apps/v1beta2"
)

var (
	daemonSetMeasurement = "kube_daemonset"
)

func registerDaemonSetCollector(ctx context.Context, acc telegraf.Accumulator, ks *KubenetesState) {
	list, err := ks.client.getDaemonSets(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, d := range list.Items {
		if err = ks.gatherDaemonSet(d, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ks *KubenetesState) gatherDaemonSet(d v1beta2.DaemonSet, acc telegraf.Accumulator) error {
	status := d.Status
	fields := map[string]interface{}{
		"metadata_generation":             d.ObjectMeta.Generation,
		"status_current_number_scheduled": status.CurrentNumberScheduled,
		"status_desired_number_scheduled": status.DesiredNumberScheduled,
		"status_number_available":         status.NumberAvailable,
		"status_number_misscheduled":      status.NumberMisscheduled,
		"status_number_ready":             status.NumberReady,
		"status_number_unavailable":       status.NumberUnavailable,
		"status_updated_number_scheduled": status.UpdatedNumberScheduled,
	}
	tags := map[string]string{
		"namespace": d.Namespace,
		"daemonset": d.Name,
	}
	for k, v := range d.Labels {
		tags["label_"+sanitizeLabelName(k)] = v
	}
	if !d.CreationTimestamp.IsZero() {
		fields["created"] = d.CreationTimestamp.Unix()
	}

	acc.AddFields(daemonSetMeasurement, fields, tags)
	return nil
}
