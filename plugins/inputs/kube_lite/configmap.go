package kube_lite

import (
	"context"
	"time"

	"github.com/ericchiang/k8s/apis/core/v1"

	"github.com/influxdata/telegraf"
)

var configMapMeasurement = "kube_configmap"

func collectConfigMaps(ctx context.Context, acc telegraf.Accumulator, ks *KubernetesState) {
	list, err := ks.client.getConfigMaps(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, s := range list.Items {
		if s != nil {
			if err = ks.gatherConfigMap(*s, acc); err != nil {
				acc.AddError(err)
				return
			}
		}
	}
}

func (ks *KubernetesState) gatherConfigMap(s v1.ConfigMap, acc telegraf.Accumulator) error {
	if s.Metadata.CreationTimestamp.GetSeconds() == 0 {
		return nil
	} else if !ks.firstTimeGather &&
		ks.MaxConfigMapAge.Duration < time.Now().Sub(time.Unix(s.Metadata.CreationTimestamp.GetSeconds(), int64(s.Metadata.CreationTimestamp.GetNanos()))) {
		return nil
	}

	fields := map[string]interface{}{
		"gauge": 1,
	}
	tags := map[string]string{
		"name":             s.Metadata.GetName(),
		"namespace":        s.Metadata.GetNamespace(),
		"resource_version": s.Metadata.GetResourceVersion(),
	}

	acc.AddFields(configMapMeasurement, fields, tags, time.Unix(s.Metadata.CreationTimestamp.GetSeconds(), int64(s.Metadata.CreationTimestamp.GetNanos())))

	return nil
}
