package kube_lite

import (
	"context"
	"time"

	"github.com/ericchiang/k8s/apis/apps/v1beta1"

	"github.com/influxdata/telegraf"
)

var (
	deploymentMeasurement = "kube_deployment"
)

func collectDeployments(ctx context.Context, acc telegraf.Accumulator, ks *KubernetesState) {
	list, err := ks.client.getDeployments(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, d := range list.Items {
		if err = ks.gatherDeployment(*d, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ks *KubernetesState) gatherDeployment(d v1beta1.Deployment, acc telegraf.Accumulator) error {
	fields := map[string]interface{}{
		"status_replicas_available":   d.Status.GetAvailableReplicas(),
		"status_replicas_unavailable": d.Status.GetUnavailableReplicas(),
		"created":                     time.Unix(d.Metadata.CreationTimestamp.GetSeconds(), int64(d.Metadata.CreationTimestamp.GetNanos())).UnixNano(),
	}
	tags := map[string]string{
		"name":      d.Metadata.GetName(),
		"namespace": d.Metadata.GetNamespace(),
	}

	acc.AddFields(deploymentMeasurement, fields, tags)

	return nil
}
