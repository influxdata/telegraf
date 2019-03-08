package kube_inventory

import (
	"context"
	"time"

	"github.com/ericchiang/k8s/apis/apps/v1beta2"

	"github.com/influxdata/telegraf"
)

func collectDaemonSets(ctx context.Context, acc telegraf.Accumulator, ki *KubernetesInventory) {
	list, err := ki.client.getDaemonSets(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, d := range list.Items {
		if err = ki.gatherDaemonSet(*d, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ki *KubernetesInventory) gatherDaemonSet(d v1beta2.DaemonSet, acc telegraf.Accumulator) error {
	fields := map[string]interface{}{
		"generation":               d.Metadata.GetGeneration(),
		"current_number_scheduled": d.Status.GetCurrentNumberScheduled(),
		"desired_number_scheduled": d.Status.GetDesiredNumberScheduled(),
		"number_available":         d.Status.GetNumberAvailable(),
		"number_misscheduled":      d.Status.GetNumberMisscheduled(),
		"number_ready":             d.Status.GetNumberReady(),
		"number_unavailable":       d.Status.GetNumberUnavailable(),
		"updated_number_scheduled": d.Status.GetUpdatedNumberScheduled(),
	}
	tags := map[string]string{
		"daemonset_name": d.Metadata.GetName(),
		"namespace":      d.Metadata.GetNamespace(),
	}

	if d.Metadata.CreationTimestamp.GetSeconds() != 0 {
		fields["created"] = time.Unix(d.Metadata.CreationTimestamp.GetSeconds(), int64(d.Metadata.CreationTimestamp.GetNanos())).UnixNano()
	}

	acc.AddFields(daemonSetMeasurement, fields, tags)

	return nil
}
