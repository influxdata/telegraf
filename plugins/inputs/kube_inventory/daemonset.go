package kube_inventory

import (
	"context"

	v1 "k8s.io/api/apps/v1"

	"github.com/influxdata/telegraf"
)

func collectDaemonSets(ctx context.Context, acc telegraf.Accumulator, ki *KubernetesInventory) {
	list, err := ki.client.getDaemonSets(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, d := range list.Items {
		ki.gatherDaemonSet(d, acc)
	}
}

func (ki *KubernetesInventory) gatherDaemonSet(d v1.DaemonSet, acc telegraf.Accumulator) {
	fields := map[string]interface{}{
		"generation":               d.Generation,
		"current_number_scheduled": d.Status.CurrentNumberScheduled,
		"desired_number_scheduled": d.Status.DesiredNumberScheduled,
		"number_available":         d.Status.NumberAvailable,
		"number_misscheduled":      d.Status.NumberMisscheduled,
		"number_ready":             d.Status.NumberReady,
		"number_unavailable":       d.Status.NumberUnavailable,
		"updated_number_scheduled": d.Status.UpdatedNumberScheduled,
	}
	tags := map[string]string{
		"daemonset_name": d.Name,
		"namespace":      d.Namespace,
	}
	for key, val := range d.Spec.Selector.MatchLabels {
		if ki.selectorFilter.Match(key) {
			tags["selector_"+key] = val
		}
	}

	creationTs := d.GetCreationTimestamp()
	if !creationTs.IsZero() {
		fields["created"] = d.GetCreationTimestamp().UnixNano()
	}

	acc.AddFields(daemonSetMeasurement, fields, tags)
}
