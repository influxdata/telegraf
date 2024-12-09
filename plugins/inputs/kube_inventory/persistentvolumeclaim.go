package kube_inventory

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/influxdata/telegraf"
)

func collectPersistentVolumeClaims(ctx context.Context, acc telegraf.Accumulator, ki *KubernetesInventory) {
	list, err := ki.client.getPersistentVolumeClaims(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, pvc := range list.Items {
		ki.gatherPersistentVolumeClaim(pvc, acc)
	}
}

func (ki *KubernetesInventory) gatherPersistentVolumeClaim(pvc corev1.PersistentVolumeClaim, acc telegraf.Accumulator) {
	phaseType := 3
	switch strings.ToLower(string(pvc.Status.Phase)) {
	case "bound":
		phaseType = 0
	case "lost":
		phaseType = 1
	case "pending":
		phaseType = 2
	}
	fields := map[string]interface{}{
		"phase_type": phaseType,
	}
	tags := map[string]string{
		"pvc_name":  pvc.Name,
		"namespace": pvc.Namespace,
		"phase":     string(pvc.Status.Phase),
	}
	if pvc.Spec.StorageClassName != nil {
		tags["storageclass"] = *pvc.Spec.StorageClassName
	}
	if pvc.Spec.Selector != nil {
		for key, val := range pvc.Spec.Selector.MatchLabels {
			if ki.selectorFilter.Match(key) {
				tags["selector_"+key] = val
			}
		}
	}

	acc.AddFields(persistentVolumeClaimMeasurement, fields, tags)
}
