package kube_state

import (
	"context"
	"time"

	"github.com/ericchiang/k8s/apis/apps/v1beta1"

	"github.com/influxdata/telegraf"
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
		"replicas_available":   d.Status.GetAvailableReplicas(),
		"replicas_unavailable": d.Status.GetUnavailableReplicas(),
		"created":              time.Unix(d.Metadata.CreationTimestamp.GetSeconds(), int64(d.Metadata.CreationTimestamp.GetNanos())).UnixNano(),
	}
	tags := map[string]string{
		"deployment_name": d.Metadata.GetName(),
		"namespace":       d.Metadata.GetNamespace(),
	}

	acc.AddFields(deploymentMeasurement, fields, tags)

	return nil
}
