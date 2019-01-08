package kube_lite

import (
	"context"
	"strings"

	"github.com/ericchiang/k8s/apis/core/v1"

	"github.com/influxdata/telegraf"
)

var persistentVolumeMeasurement = "kube_persistentvolume"

func collectPersistentVolumes(ctx context.Context, acc telegraf.Accumulator, ks *KubernetesState) {
	list, err := ks.client.getPersistentVolumes(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, pv := range list.Items {
		if err = ks.gatherPersistentVolume(*pv, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ks *KubernetesState) gatherPersistentVolume(pv v1.PersistentVolume, acc telegraf.Accumulator) error {
	fields := map[string]interface{}{}
	tags := map[string]string{
		"name":         pv.Metadata.GetName(),
		"status":       pv.Status.GetPhase(),
		"storageclass": pv.Spec.GetStorageClassName(),
	}

	// Set current phase to 1, others to 0 if it is set.
	if p := pv.Status.GetPhase(); p != "" {
		fields["status_available"] = boolInt(strings.ToLower(p) == "available")
		fields["status_bound"] = boolInt(strings.ToLower(p) == "bound")
		fields["status_failed"] = boolInt(strings.ToLower(p) == "failed")
		fields["status_pending"] = boolInt(strings.ToLower(p) == "pending")
		fields["status_released"] = boolInt(strings.ToLower(p) == "released")
	}

	acc.AddFields(persistentVolumeMeasurement, fields, tags)

	return nil
}
