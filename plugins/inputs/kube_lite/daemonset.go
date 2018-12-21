package kube_lite

import (
	"context"

	"github.com/ericchiang/k8s/apis/apps/v1beta2"

	"github.com/influxdata/telegraf"
)

var (
	daemonSetMeasurement = "kube_daemonset"
)

func registerDaemonSetCollector(ctx context.Context, acc telegraf.Accumulator, ks *KubernetesState) {
	list, err := ks.client.getDaemonSets(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, d := range list.Items {
		if err = ks.gatherDaemonSet(*d, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ks *KubernetesState) gatherDaemonSet(d v1beta2.DaemonSet, acc telegraf.Accumulator) error {
	fields := map[string]interface{}{
		"metadata_generation":             d.Metadata.GetGeneration(),
		"status_current_number_scheduled": d.Status.GetCurrentNumberScheduled(),
		"status_desired_number_scheduled": d.Status.GetDesiredNumberScheduled(),
		"status_number_available":         d.Status.GetNumberAvailable(),
		"status_number_misscheduled":      d.Status.GetNumberMisscheduled(),
		"status_number_ready":             d.Status.GetNumberReady(),
		"status_number_unavailable":       d.Status.GetNumberUnavailable(),
		"status_updated_number_scheduled": d.Status.GetUpdatedNumberScheduled(),
	}
	tags := map[string]string{
		"namespace": d.Metadata.GetNamespace(),
		"name":      d.Metadata.GetName(),
	}
	if d.Metadata.CreationTimestamp.GetSeconds() != 0 {
		fields["created"] = d.Metadata.CreationTimestamp.GetSeconds()
		// fields["created"] = time.Unix(d.Metadata.CreationTimestamp.GetSeconds(), int64(d.Metadata.CreationTimestamp.GetNanos()))
	}

	acc.AddFields(daemonSetMeasurement, fields, tags)
	return nil
}
