package kube_state

import (
	"context"
	"strings"

	"github.com/influxdata/telegraf"
	"k8s.io/api/core/v1"
)

var persistentVolumeMeasurement = "kube_persistentvolume"

func registerPersistentVolumeCollector(ctx context.Context, acc telegraf.Accumulator, ks *KubenetesState) {
	list, err := ks.client.getPersistentVolumes(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, pv := range list.Items {
		if err = ks.collectPersistentVolume(pv, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ks *KubenetesState) collectPersistentVolume(pv v1.PersistentVolume, acc telegraf.Accumulator) error {
	fields := map[string]interface{}{}
	tags := map[string]string{
		"persistentvolume": pv.Name,
		"storageclass":     pv.Spec.StorageClassName,
	}
	for k, v := range pv.Labels {
		tags["label_"+sanitizeLabelName(k)] = v
	}

	// Set current phase to 1, others to 0 if it is set.
	if p := pv.Status.Phase; p != "" {
		fields["status_"+strings.ToLower(string(v1.VolumePending))] = boolInt(p == v1.VolumePending)
		fields["status_"+strings.ToLower(string(v1.VolumeAvailable))] = boolInt(p == v1.VolumeAvailable)
		fields["status_"+strings.ToLower(string(v1.VolumeBound))] = boolInt(p == v1.VolumeBound)
		fields["status_"+strings.ToLower(string(v1.VolumeReleased))] = boolInt(p == v1.VolumeReleased)
		fields["status_"+strings.ToLower(string(v1.VolumeFailed))] = boolInt(p == v1.VolumeFailed)
		tags["status"] = strings.ToLower(string(p))
	}

	acc.AddFields(persistentVolumeMeasurement, fields, tags)
	return nil
}
