package kube_state

import (
	"context"
	"time"

	"github.com/influxdata/telegraf"
	"k8s.io/api/core/v1"
)

var configMapMeasurement = "kube_configmap"

func registerConfigMapCollector(ctx context.Context, acc telegraf.Accumulator, ks *KubenetesState) {
	list, err := ks.client.getConfigMaps(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, s := range list.Items {
		if err = ks.gatherConfigMap(s, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ks *KubenetesState) gatherConfigMap(s v1.ConfigMap, acc telegraf.Accumulator) error {
	if s.CreationTimestamp.IsZero() {
		return nil
	} else if !ks.firstTimeGather &&
		ks.MaxConfigMapAge.Duration < time.Now().Sub(s.CreationTimestamp.Time) {
		return nil
	}

	creationTime := s.CreationTimestamp.Time
	fields := map[string]interface{}{
		"gauge": 1,
	}
	tags := map[string]string{
		"namespace":        s.Namespace,
		"configmap":        s.Name,
		"resource_version": s.ResourceVersion,
	}
	acc.AddFields(configMapMeasurement, fields, tags, creationTime)
	return nil
}
