package kube_inventory

import (
	"context"
	"strings"

	"github.com/ericchiang/k8s/apis/core/v1"

	"github.com/influxdata/telegraf"
)

func collectPersistentVolumes(ctx context.Context, acc telegraf.Accumulator, ki *KubernetesInventory) {
	list, err := ki.client.getPersistentVolumes(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, pv := range list.Items {
		if err = ki.gatherPersistentVolume(*pv, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ki *KubernetesInventory) gatherPersistentVolume(pv v1.PersistentVolume, acc telegraf.Accumulator) error {
	phaseType := 5
	switch strings.ToLower(pv.Status.GetPhase()) {
	case "bound":
		phaseType = 0
	case "failed":
		phaseType = 1
	case "pending":
		phaseType = 2
	case "released":
		phaseType = 3
	case "available":
		phaseType = 4
	}
	fields := map[string]interface{}{
		"phase_type": phaseType,
	}
	tags := map[string]string{
		"pv_name":      pv.Metadata.GetName(),
		"phase":        pv.Status.GetPhase(),
		"storageclass": pv.Spec.GetStorageClassName(),
	}

	acc.AddFields(persistentVolumeMeasurement, fields, tags)

	return nil
}
