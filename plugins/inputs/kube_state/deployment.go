package kube_state

import (
	"context"
	"fmt"
	"strconv"

	"github.com/influxdata/telegraf"
	"k8s.io/api/apps/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	deploymentMeasurement = "kube_deployment"
)

func registerDeploymentCollector(ctx context.Context, acc telegraf.Accumulator, ks *KubenetesState) {
	list, err := ks.client.getDeployments(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, d := range list.Items {
		if err = ks.gatherDeployment(d, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ks *KubenetesState) gatherDeployment(d v1beta1.Deployment, acc telegraf.Accumulator) error {
	fields := map[string]interface{}{
		"spec_replicas":               *d.Spec.Replicas,
		"metadata_generation":         d.ObjectMeta.Generation,
		"status_replicas":             d.Status.Replicas,
		"status_replicas_available":   d.Status.AvailableReplicas,
		"status_replicas_unavailable": d.Status.UnavailableReplicas,
		"status_replicas_updated":     d.Status.UpdatedReplicas,
		"status_observed_generation":  d.Status.ObservedGeneration,
	}
	if !d.CreationTimestamp.IsZero() {
		fields["created"] = d.CreationTimestamp.Unix()
	}
	tags := map[string]string{
		"namespace":   d.Namespace,
		"deployment":  d.Name,
		"spec_paused": strconv.FormatBool(d.Spec.Paused),
	}
	for k, v := range d.Labels {
		tags["label_"+sanitizeLabelName(k)] = v
	}
	var maxUnavailable, maxSurge int
	var err error
	if d.Spec.Strategy.RollingUpdate == nil {
		goto collectDeployment
	}

	maxUnavailable, err = intstr.GetValueFromIntOrPercent(d.Spec.Strategy.RollingUpdate.MaxUnavailable, int(*d.Spec.Replicas), true)
	if err != nil {
		acc.AddError(fmt.Errorf("Error converting RollingUpdate MaxUnavailable to int: %s", err))
	} else {
		fields["spec_strategy_rollingupdate_max_unavailable"] = maxUnavailable
	}

	maxSurge, err = intstr.GetValueFromIntOrPercent(d.Spec.Strategy.RollingUpdate.MaxSurge, int(*d.Spec.Replicas), true)
	if err != nil {
		acc.AddError(fmt.Errorf("Error converting RollingUpdate MaxSurge to int: %s", err))
	} else {
		fields["spec_strategy_rollingupdate_max_surge"] = maxSurge
	}
collectDeployment:
	acc.AddFields(deploymentMeasurement, fields, tags)
	return nil
}
