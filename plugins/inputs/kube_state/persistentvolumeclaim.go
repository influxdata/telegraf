package kube_state

import (
	"context"
	"strings"

	"github.com/influxdata/telegraf"
	"k8s.io/api/core/v1"
)

var persistentVolumeClaimMeasurement = "kube_persistentvolumeclaim"

func registerPersistentVolumeClaimCollector(ctx context.Context, acc telegraf.Accumulator, ks *KubenetesState) {
	list, err := ks.client.getPersistentVolumeClaims(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, pvc := range list.Items {
		if err = ks.collectPersistentVolumeClaim(pvc, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ks *KubenetesState) collectPersistentVolumeClaim(pvc v1.PersistentVolumeClaim, acc telegraf.Accumulator) error {
	fields := map[string]interface{}{}
	tags := map[string]string{
		"namespace":             pvc.Namespace,
		"persistentvolumeclaim": pvc.Name,
		"storageclass":          getPersistentVolumeClaimClass(&pvc),
		"volumename":            pvc.Spec.VolumeName,
	}
	for k, v := range pvc.Labels {
		tags["label_"+sanitizeLabelName(k)] = v
	}

	// Set current phase to 1, others to 0 if it is set.
	if p := pvc.Status.Phase; p != "" {
		fields["status_"+strings.ToLower(string(v1.ClaimLost))] = boolInt(p == v1.ClaimLost)
		fields["status_"+strings.ToLower(string(v1.ClaimBound))] = boolInt(p == v1.ClaimBound)
		fields["status_"+strings.ToLower(string(v1.ClaimPending))] = boolInt(p == v1.ClaimPending)
		tags["status"] = strings.ToLower(string(p))
	}

	if storage, ok := pvc.Spec.Resources.Requests[v1.ResourceStorage]; ok {
		fields["resource_requests_storage_bytes"] = storage.Value()
	}

	acc.AddFields(persistentVolumeClaimMeasurement, fields, tags)
	return nil
}

// getPersistentVolumeClaimClass returns StorageClassName. If no storage class was
// requested, it returns "".
func getPersistentVolumeClaimClass(claim *v1.PersistentVolumeClaim) string {
	// Use beta annotation first
	if class, found := claim.Annotations[v1.BetaStorageClassAnnotation]; found {
		return class
	}

	if claim.Spec.StorageClassName != nil {
		return *claim.Spec.StorageClassName
	}

	return ""
}
