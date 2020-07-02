package kube_inventory

import (
	"context"
	"time"

	v1 "github.com/ericchiang/k8s/apis/batch/v1"
	"github.com/influxdata/telegraf"
)

func collectJobs(ctx context.Context, acc telegraf.Accumulator, ki *KubernetesInventory) {
	list, err := ki.client.getJobs(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, d := range list.Items {
		if err = ki.gatherJob(*d, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ki *KubernetesInventory) gatherJob(d v1.Job, acc telegraf.Accumulator) error {
	fields := map[string]interface{}{
		"active":    d.Status.GetActive(),
		"completed": time.Unix(d.Status.CompletionTime.GetSeconds(), int64(d.Status.CompletionTime.GetNanos())).UnixNano(),
		"failed":    d.Status.GetFailed(),
		"started":   time.Unix(d.Status.StartTime.GetSeconds(), int64(d.Status.StartTime.GetNanos())).UnixNano(),
		"succeeded": d.Status.GetSucceeded(),
	}
	tags := map[string]string{
		"job_name":  d.Metadata.GetName(),
		"namespace": d.Metadata.GetNamespace(),
	}
	for key, val := range d.GetSpec().GetSelector().GetMatchLabels() {
		if ki.selectorFilter.Match(key) {
			tags["selector_"+key] = val
		}
	}

	acc.AddFields(jobMeasurement, fields, tags)

	return nil
}
