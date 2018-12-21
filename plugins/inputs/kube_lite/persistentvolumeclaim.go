package kube_lite

import (
	"context"
	"strings"

	"github.com/ericchiang/k8s/apis/core/v1"

	"github.com/influxdata/telegraf"
)

var persistentVolumeClaimMeasurement = "kube_persistentvolumeclaim"

func registerPersistentVolumeClaimCollector(ctx context.Context, acc telegraf.Accumulator, ks *KubernetesState) {
	list, err := ks.client.getPersistentVolumeClaims(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, pvc := range list.Items {
		if err = ks.collectPersistentVolumeClaim(*pvc, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ks *KubernetesState) collectPersistentVolumeClaim(pvc v1.PersistentVolumeClaim, acc telegraf.Accumulator) error {
	fields := map[string]interface{}{}
	tags := map[string]string{
		"name":         pvc.Metadata.GetName(),
		"namespace":    pvc.Metadata.GetNamespace(),
		"status":       pvc.Status.GetPhase(),
		"storageclass": pvc.Spec.GetStorageClassName(),
		// "volumename":   pvc.Spec.GetVolumeName(),
	}

	// Set current phase to 1, others to 0 if it is set.
	if p := pvc.Status.GetPhase(); p != "" {
		fields["status_lost"] = boolInt(strings.ToLower(p) == "lost")
		fields["status_bound"] = boolInt(strings.ToLower(p) == "bound")
		fields["status_pending"] = boolInt(strings.ToLower(p) == "pending")
	}

	acc.AddFields(persistentVolumeClaimMeasurement, fields, tags)
	return nil
}
