package kube_state

import (
	"context"

	"github.com/influxdata/telegraf"
	"k8s.io/api/core/v1"
)

var secretMeasurement = "kube_secret"

func registerSecretCollector(ctx context.Context, acc telegraf.Accumulator, ks *KubenetesState) {
	list, err := ks.client.getSecrets(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, s := range list.Items {
		if err = ks.gatherSecret(s, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ks *KubenetesState) gatherSecret(s v1.Secret, acc telegraf.Accumulator) error {
	if s.CreationTimestamp.IsZero() {
		return nil
	}
	fields := map[string]interface{}{
		"gauge": 1,
	}
	tags := map[string]string{
		"namespace":        s.Namespace,
		"secret":           s.Name,
		"resource_version": s.ResourceVersion,
	}
	for k, v := range s.Labels {
		tags["label_"+sanitizeLabelName(k)] = v
	}
	acc.AddFields(secretMeasurement, fields, tags, s.CreationTimestamp.Time)

	return nil
}
